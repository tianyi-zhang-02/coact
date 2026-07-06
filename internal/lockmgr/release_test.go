package lockmgr

import (
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func TestReleaseAll(t *testing.T) {
	root := t.TempDir()
	m := New(&project.Project{Root: root}, config.Default())

	for _, path := range []string{"a", "b"} {
		if _, err := m.Acquire("claude", abs(root, path)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.Acquire("codex", abs(root, "c")); err != nil {
		t.Fatal(err)
	}

	n, err := m.ReleaseAll("claude")
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("want 2 released, got %d", n)
	}

	locks, _ := m.List()
	codexRemains := false
	for _, lk := range locks {
		if lk.Owner == "claude" {
			t.Fatal("a claude lock survived ReleaseAll")
		}
		if lk.Owner == "codex" {
			codexRemains = true
		}
	}
	if !codexRemains {
		t.Fatal("the codex lock should remain")
	}
}
