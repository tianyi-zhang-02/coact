// Package presence tracks per-session liveness. Presence is written frequently
// (by a sidecar or by adapter hooks) and is the sole authority for whether a
// participant is live — which in turn decides whether its locks may be stolen.
package presence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coactdev/coact/internal/platform"
)

// DefaultTTLSeconds is the liveness window if none is configured.
const DefaultTTLSeconds = 60

type Session struct {
	Agent       string   `json:"agent"`
	PID         int      `json:"pid"`
	Cwd         string   `json:"cwd"`
	CurrentTask string   `json:"current_task,omitempty"`
	HeldLocks   []string `json:"held_locks,omitempty"`
	StartedAt   string   `json:"started_at,omitempty"`
	HeartbeatAt string   `json:"heartbeat_at"`
	Status      string   `json:"status"`
}

func sessionPath(sessionDir, agent string) string {
	return filepath.Join(sessionDir, agent+".json")
}

func nowRFC() string { return time.Now().UTC().Format(time.RFC3339) }

// Register marks the current process as the owner of agent's session and writes
// a fresh heartbeat. Used by the long-lived presence sidecar; the owner PID it
// records is what liveness checks probe.
func Register(sessionDir, agent, status string) error {
	return write(sessionDir, agent, status, "", true)
}

// Beat refreshes agent's heartbeat (and optionally status / current task)
// without claiming ownership: it preserves whatever owner PID is already on
// file. Used by one-shot commands and by hook-only mode, so an ephemeral
// `coact lock` process never overwrites the sidecar's PID with its own.
func Beat(sessionDir, agent, status, task string) error {
	return write(sessionDir, agent, status, task, false)
}

func write(sessionDir, agent, status, task string, own bool) error {
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return err
	}
	s := &Session{}
	if data, err := os.ReadFile(sessionPath(sessionDir, agent)); err == nil {
		_ = json.Unmarshal(data, s)
	}
	s.Agent = agent
	if own {
		s.PID = os.Getpid()
		if cwd, err := os.Getwd(); err == nil {
			s.Cwd = cwd
		}
	}
	if s.StartedAt == "" {
		s.StartedAt = nowRFC()
	}
	s.HeartbeatAt = nowRFC()
	if status != "" {
		s.Status = status
	}
	if task != "" {
		s.CurrentTask = task
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return platform.AtomicWrite(sessionPath(sessionDir, agent), data, 0o644)
}

// Read returns the session for agent, or an error if none exists.
func Read(sessionDir, agent string) (*Session, error) {
	data, err := os.ReadFile(sessionPath(sessionDir, agent))
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// List returns all known sessions.
func List(sessionDir string) ([]*Session, error) {
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*Session
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sessionDir, e.Name()))
		if err != nil {
			continue
		}
		var s Session
		if json.Unmarshal(data, &s) == nil {
			out = append(out, &s)
		}
	}
	return out, nil
}

// IsLive reports whether agent's session is fresh (within ttlSeconds) and its
// process is still running.
func IsLive(sessionDir, agent string, ttlSeconds int) bool {
	s, err := Read(sessionDir, agent)
	if err != nil {
		return false
	}
	if ttlSeconds <= 0 {
		ttlSeconds = DefaultTTLSeconds
	}
	t, err := time.Parse(time.RFC3339, s.HeartbeatAt)
	if err != nil {
		return false
	}
	if time.Since(t) > time.Duration(ttlSeconds)*time.Second {
		return false
	}
	if s.PID > 0 && !platform.ProcessAlive(s.PID) {
		return false
	}
	return true
}

// Age returns how long ago the session last beat, and whether it parsed.
func (s *Session) Age() (time.Duration, bool) {
	t, err := time.Parse(time.RFC3339, s.HeartbeatAt)
	if err != nil {
		return 0, false
	}
	return time.Since(t), true
}
