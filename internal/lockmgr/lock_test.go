package lockmgr

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func TestConflictPaths(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"src/api", "src/api", true},            // identical
		{"src/api", "src/api/handler.go", true}, // parent/child
		{"src/api/handler.go", "src/api", true}, // child/parent (symmetric)
		{"src/api", "src/apidocs", false},       // shared string prefix, different segment
		{"src/api", "src/web", false},           // disjoint
		{".", "anything/here", true},            // root covers all
	}
	for _, c := range cases {
		if got := conflictPaths(c.a, c.b); got != c.want {
			t.Errorf("conflictPaths(%q,%q)=%v want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestExpired(t *testing.T) {
	fresh := Lock{AcquiredAt: nowRFC(), HeartbeatAt: nowRFC(), TTLSeconds: 900}
	if expired(fresh) {
		t.Error("fresh lock should not be expired")
	}
	old := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	stale := Lock{AcquiredAt: old, HeartbeatAt: old, TTLSeconds: 60}
	if !expired(stale) {
		t.Error("hour-old lock with 60s ttl should be expired")
	}
}

func newTestManager(t *testing.T) (*Manager, string) {
	t.Helper()
	root := t.TempDir()
	p := &project.Project{Root: root}
	return New(p, config.Default()), root
}

func abs(root, rel string) string { return filepath.Join(root, rel) }

func TestAcquireDenyReleaseCycle(t *testing.T) {
	m, root := newTestManager(t)

	// claude takes src/api
	if res, err := m.Acquire("claude", abs(root, "src/api")); err != nil || !res.Acquired {
		t.Fatalf("claude acquire src/api: res=%v err=%v", res, err)
	}

	// codex is denied on the child path
	res, err := m.Acquire("codex", abs(root, "src/api/handler.go"))
	if err != nil {
		t.Fatalf("codex acquire child: %v", err)
	}
	if res.Acquired {
		t.Fatal("codex should be denied on a child of a held path")
	}
	if res.Conflict == nil || res.Conflict.Owner != "claude" {
		t.Fatalf("expected conflict owned by claude, got %+v", res.Conflict)
	}

	// codex succeeds on a disjoint path
	if res, err := m.Acquire("codex", abs(root, "src/web")); err != nil || !res.Acquired {
		t.Fatalf("codex acquire src/web: res=%v err=%v", res, err)
	}

	// claude releases; codex can now take the child
	if err := m.Release("claude", abs(root, "src/api")); err != nil {
		t.Fatalf("claude release: %v", err)
	}
	if res, err := m.Acquire("codex", abs(root, "src/api/handler.go")); err != nil || !res.Acquired {
		t.Fatalf("codex acquire after release: res=%v err=%v", res, err)
	}
}

func TestReleaseWrongOwnerRejected(t *testing.T) {
	m, root := newTestManager(t)
	if _, err := m.Acquire("claude", abs(root, "src")); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := m.Release("codex", abs(root, "src")); err == nil {
		t.Fatal("releasing a lock owned by another agent should fail")
	}
}

func TestReclaimExpiredDeadOwner(t *testing.T) {
	m, root := newTestManager(t)

	// Plant an expired lock owned by a participant with no live presence.
	// (Production always creates the locks dir via the registry first.)
	if err := os.MkdirAll(filepath.Dir(m.lockFile("legacy/module")), 0o755); err != nil {
		t.Fatalf("mkdir locks: %v", err)
	}
	old := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if err := m.writeLock(Lock{
		Path:        "legacy/module",
		Mode:        "exclusive",
		Owner:       "ghost",
		PID:         999999, // not a live process
		AcquiredAt:  old,
		HeartbeatAt: old,
		TTLSeconds:  60,
	}); err != nil {
		t.Fatalf("plant lock: %v", err)
	}

	// A live agent should be able to steal it.
	res, err := m.Acquire("claude", abs(root, "legacy/module"))
	if err != nil {
		t.Fatalf("acquire reclaimable: %v", err)
	}
	if !res.Acquired {
		t.Fatal("expired lock with dead owner should be reclaimable")
	}
}
