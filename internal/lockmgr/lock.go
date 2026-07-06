// Package lockmgr implements coact's advisory write-intent locks.
//
// All acquisitions are serialized under a single registry meta-lock so that the
// overlap (prefix) check and the lock write happen atomically with respect to
// other participants — closing the TOCTOU race that a plain create-then-scan
// would have (SPEC §2.3). A lock is stolen only when it is both expired and its
// owner is not live per presence (SPEC §2.4).
package lockmgr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/metalock"
	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/policy"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

const registryLockName = ".registry.lock"

// Lock is the on-disk representation of a held write-intent.
type Lock struct {
	Path        string `json:"path"`
	Mode        string `json:"mode"`
	Owner       string `json:"owner"`
	PID         int    `json:"pid"`
	AcquiredAt  string `json:"acquired_at"`
	TTLSeconds  int    `json:"ttl_seconds"`
	HeartbeatAt string `json:"heartbeat_at"`
}

// Result describes the outcome of an Acquire or Check.
type Result struct {
	Acquired bool
	Reason   string // acquired | reentrant | stolen | denied | policy
	Conflict *Lock  // set on a lock conflict (Reason=denied)
	Path     string // normalized repo-relative path
	Detail   string // human-readable reason (set on Reason=policy)
}

// Manager operates the lock files under a project's .coact/locks directory.
type Manager struct {
	root        string
	locksDir    string
	sessionDir  string
	journalDir  string
	ttl         int
	presenceTTL int
	registryTTL time.Duration
	policy      *policy.Engine
}

// New builds a Manager from a project and its config.
func New(p *project.Project, cfg *config.Config) *Manager {
	regTTL := time.Duration(cfg.Locks.RegistryLockTTLSeconds) * time.Second
	if regTTL <= 0 {
		regTTL = 5 * time.Second
	}
	ttl := cfg.Locks.DefaultTTLSeconds
	if ttl <= 0 {
		ttl = 900
	}
	return &Manager{
		root:        p.Root,
		locksDir:    p.LocksDir(),
		sessionDir:  p.SessionDir(),
		journalDir:  p.JournalDir(),
		ttl:         ttl,
		presenceTTL: cfg.Presence.TTLSeconds,
		registryTTL: regTTL,
		policy:      policy.New(cfg),
	}
}

func nowRFC() string { return time.Now().UTC().Format(time.RFC3339) }

// rel normalizes a user-supplied path to a clean, slash-separated path relative
// to the repo root, rejecting anything outside the root.
func (m *Manager) rel(raw string) (string, error) {
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	r, err := filepath.Rel(m.root, abs)
	if err != nil {
		return "", err
	}
	r = filepath.ToSlash(filepath.Clean(r))
	if r == ".." || strings.HasPrefix(r, "../") {
		return "", fmt.Errorf("path %q is outside the repo root", raw)
	}
	return r, nil
}

func hashPath(rel string) string {
	sum := sha256.Sum256([]byte(rel))
	return hex.EncodeToString(sum[:6]) // 12 hex chars
}

func (m *Manager) lockFile(rel string) string {
	return filepath.Join(m.locksDir, hashPath(rel)+".lock")
}

// --- registry meta-lock ---------------------------------------------------

func (m *Manager) acquireRegistry() error {
	if err := os.MkdirAll(m.locksDir, 0o755); err != nil {
		return err
	}
	return metalock.Acquire(filepath.Join(m.locksDir, registryLockName), m.registryTTL, 10*time.Second)
}

func (m *Manager) releaseRegistry() {
	metalock.Release(filepath.Join(m.locksDir, registryLockName))
}

// --- scanning -------------------------------------------------------------

type storedLock struct {
	file string
	lock Lock
}

func (m *Manager) liveScan() ([]storedLock, error) {
	entries, err := os.ReadDir(m.locksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []storedLock
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || name == registryLockName || !strings.HasSuffix(name, ".lock") {
			continue
		}
		full := filepath.Join(m.locksDir, name)
		data, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		var lk Lock
		if json.Unmarshal(data, &lk) != nil {
			continue
		}
		out = append(out, storedLock{file: full, lock: lk})
	}
	return out, nil
}

func (m *Manager) writeLock(lk Lock) error {
	data, err := json.Marshal(lk)
	if err != nil {
		return err
	}
	return platform.AtomicWrite(m.lockFile(lk.Path), data, 0o644)
}

// --- conflict / expiry logic ---------------------------------------------

// conflictPaths reports whether two normalized paths overlap: equal, or one is
// a path-segment prefix of the other.
func conflictPaths(a, b string) bool {
	if a == b {
		return true
	}
	return isAncestor(a, b) || isAncestor(b, a)
}

func isAncestor(parent, child string) bool {
	if parent == "." || parent == "" {
		return true // the repo root covers everything
	}
	return strings.HasPrefix(child, parent+"/")
}

func expired(lk Lock) bool {
	ref := maxTime(lk.AcquiredAt, lk.HeartbeatAt)
	if ref.IsZero() {
		return true
	}
	ttl := lk.TTLSeconds
	if ttl <= 0 {
		ttl = 900
	}
	return time.Since(ref) > time.Duration(ttl)*time.Second
}

func maxTime(a, b string) time.Time {
	ta, ea := time.Parse(time.RFC3339, a)
	tb, eb := time.Parse(time.RFC3339, b)
	switch {
	case ea != nil && eb != nil:
		return time.Time{}
	case ea != nil:
		return tb
	case eb != nil:
		return ta
	case tb.After(ta):
		return tb
	default:
		return ta
	}
}

// reclaimable reports whether a conflicting lock may be stolen: it must be both
// expired and owned by a participant that is not live per presence.
func (m *Manager) reclaimable(lk Lock) bool {
	return expired(lk) && !presence.IsLive(m.sessionDir, lk.Owner, m.presenceTTL)
}

func (m *Manager) journal(agent, event string, fields map[string]string) {
	_ = journal.Append(m.journalDir, agent, event, fields)
}

// policyDeny returns a denial Result if policy forbids agent writing rel, else
// nil. It journals the denial only when doJournal is set (Acquire, not Check).
func (m *Manager) policyDeny(agent, rel string, doJournal bool) *Result {
	if m.policy == nil {
		return nil
	}
	d := m.policy.Check(agent, rel)
	if d.Allowed {
		return nil
	}
	if doJournal {
		m.journal(agent, "policy.deny", map[string]string{"path": rel, "reason": d.Reason})
	}
	return &Result{Acquired: false, Reason: "policy", Path: rel, Detail: d.Reason}
}

// --- public API -----------------------------------------------------------

// Acquire attempts to take a write-intent lock on rawPath for agent.
func (m *Manager) Acquire(agent, rawPath string) (*Result, error) {
	rel, err := m.rel(rawPath)
	if err != nil {
		return nil, err
	}
	if r := m.policyDeny(agent, rel, true); r != nil {
		return r, nil
	}
	if err := m.acquireRegistry(); err != nil {
		return nil, err
	}
	defer m.releaseRegistry()

	locks, err := m.liveScan()
	if err != nil {
		return nil, err
	}

	for _, sl := range locks {
		if !conflictPaths(sl.lock.Path, rel) {
			continue
		}
		if sl.lock.Owner == agent {
			if sl.lock.Path == rel {
				sl.lock.HeartbeatAt = nowRFC()
				if err := m.writeLock(sl.lock); err != nil {
					return nil, err
				}
				return &Result{Acquired: true, Reason: "reentrant", Path: rel}, nil
			}
			// Same owner already holds an overlapping path — no conflict.
			continue
		}
		if m.reclaimable(sl.lock) {
			os.Remove(sl.file)
			m.journal(agent, "lock.stolen", map[string]string{
				"path": rel, "from": sl.lock.Owner, "reason": "ttl_expired_owner_dead",
			})
			continue
		}
		// Live conflict owned by another participant.
		conflict := sl.lock
		m.journal(agent, "lock.denied", map[string]string{
			"path": rel, "held_by": conflict.Owner,
		})
		return &Result{Acquired: false, Reason: "denied", Conflict: &conflict, Path: rel}, nil
	}

	lk := Lock{
		Path:        rel,
		Mode:        "exclusive",
		Owner:       agent,
		PID:         os.Getpid(),
		AcquiredAt:  nowRFC(),
		TTLSeconds:  m.ttl,
		HeartbeatAt: nowRFC(),
	}
	if err := m.writeLock(lk); err != nil {
		return nil, err
	}
	m.journal(agent, "lock.acquire", map[string]string{"path": rel})
	return &Result{Acquired: true, Reason: "acquired", Path: rel}, nil
}

// Check evaluates whether agent could acquire rawPath, without modifying state.
func (m *Manager) Check(agent, rawPath string) (*Result, error) {
	rel, err := m.rel(rawPath)
	if err != nil {
		return nil, err
	}
	if r := m.policyDeny(agent, rel, false); r != nil {
		return r, nil
	}
	if err := m.acquireRegistry(); err != nil {
		return nil, err
	}
	defer m.releaseRegistry()

	locks, err := m.liveScan()
	if err != nil {
		return nil, err
	}
	for _, sl := range locks {
		if !conflictPaths(sl.lock.Path, rel) {
			continue
		}
		if sl.lock.Owner == agent {
			continue
		}
		if m.reclaimable(sl.lock) {
			continue
		}
		conflict := sl.lock
		return &Result{Acquired: false, Reason: "denied", Conflict: &conflict, Path: rel}, nil
	}
	return &Result{Acquired: true, Reason: "acquired", Path: rel}, nil
}

// Release drops agent's lock on rawPath.
func (m *Manager) Release(agent, rawPath string) error {
	rel, err := m.rel(rawPath)
	if err != nil {
		return err
	}
	if err := m.acquireRegistry(); err != nil {
		return err
	}
	defer m.releaseRegistry()

	file := m.lockFile(rel)
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no lock held on %q", rel)
		}
		return err
	}
	var lk Lock
	if err := json.Unmarshal(data, &lk); err != nil {
		return err
	}
	if lk.Owner != agent {
		m.journal(agent, "lock.violation", map[string]string{
			"path": rel, "held_by": lk.Owner, "action": "release",
		})
		return fmt.Errorf("lock on %q is held by %q, not %q", rel, lk.Owner, agent)
	}
	if err := os.Remove(file); err != nil {
		return err
	}
	m.journal(agent, "lock.release", map[string]string{"path": rel})
	return nil
}

// ReleaseAll drops every lock held by agent (e.g. at session end) and returns
// how many were released.
func (m *Manager) ReleaseAll(agent string) (int, error) {
	if err := m.acquireRegistry(); err != nil {
		return 0, err
	}
	defer m.releaseRegistry()

	locks, err := m.liveScan()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, sl := range locks {
		if sl.lock.Owner != agent {
			continue
		}
		if os.Remove(sl.file) == nil {
			n++
			m.journal(agent, "lock.release", map[string]string{"path": sl.lock.Path, "reason": "session_end"})
		}
	}
	return n, nil
}

// List returns all live locks (excluding the registry meta-lock).
func (m *Manager) List() ([]Lock, error) {
	stored, err := m.liveScan()
	if err != nil {
		return nil, err
	}
	out := make([]Lock, 0, len(stored))
	for _, sl := range stored {
		out = append(out, sl.lock)
	}
	return out, nil
}

// Expired reports whether a lock has passed its TTL (ignoring presence).
func Expired(lk Lock) bool { return expired(lk) }
