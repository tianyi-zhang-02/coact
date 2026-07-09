package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func setupProject(t *testing.T) *project.Project {
	t.Helper()
	p := &project.Project{Root: t.TempDir()}
	for _, d := range []string{p.CoactDir(), p.LocksDir(), p.SessionDir(), p.JournalDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := config.Default().Save(p.ConfigPath()); err != nil {
		t.Fatal(err)
	}
	return p
}

func editPayload(root, rel string) hookPayload {
	var pl hookPayload
	pl.Cwd = root
	pl.ToolName = "Edit"
	pl.ToolInput.FilePath = filepath.Join(root, rel)
	return pl
}

func TestHookDecisionBlocksForeignLock(t *testing.T) {
	p := setupProject(t)
	cfg, _ := config.Load(p.ConfigPath())
	m := lockmgr.New(p, cfg)
	if _, err := m.Acquire("codex", filepath.Join(p.Root, "src/api")); err != nil {
		t.Fatal(err)
	}
	deny, reason := hookDecision("claude", editPayload(p.Root, "src/api/handler.go"))
	if !deny {
		t.Fatal("claude edit into codex's locked path should be denied")
	}
	if !strings.Contains(reason, "codex") {
		t.Errorf("reason should name codex: %s", reason)
	}
}

func TestHookDecisionAllowsFreePath(t *testing.T) {
	p := setupProject(t)
	if deny, _ := hookDecision("claude", editPayload(p.Root, "src/web/page.go")); deny {
		t.Fatal("editing a free path should be allowed")
	}
}

func TestHookDecisionBlocksProtectedPath(t *testing.T) {
	p := setupProject(t) // default protects machine-mutated coact state
	deny, reason := hookDecision("claude", editPayload(p.Root, ".coact/config.json"))
	if !deny {
		t.Fatal("editing a protected path should be denied")
	}
	if !strings.Contains(reason, "protected") {
		t.Errorf("reason should mention protected: %s", reason)
	}
}

func TestHookDecisionAllowsPlanningRunPath(t *testing.T) {
	p := setupProject(t)
	deny, reason := hookDecision("claude", editPayload(p.Root, ".coact/runs/run-1/proposals/claude.md"))
	if deny {
		t.Fatalf("planning proposal path should be writable through normal locks: %s", reason)
	}
}

func TestHookDecisionFailsOpenOutsideCoact(t *testing.T) {
	dir := t.TempDir() // no .coact anywhere above
	var pl hookPayload
	pl.Cwd = dir
	pl.ToolName = "Edit"
	pl.ToolInput.FilePath = filepath.Join(dir, "x.go")
	if deny, _ := hookDecision("claude", pl); deny {
		t.Fatal("a repo without coact must fail open (allow)")
	}
}

func TestHookDecisionIgnoresNonEditTools(t *testing.T) {
	p := setupProject(t)
	var pl hookPayload
	pl.Cwd = p.Root
	pl.ToolName = "Bash"
	if deny, _ := hookDecision("claude", pl); deny {
		t.Fatal("non-edit tools should never be gated")
	}
}
