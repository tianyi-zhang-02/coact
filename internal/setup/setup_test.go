package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
)

func TestEnsureGitignoreIncludesRuntimeState(t *testing.T) {
	root := t.TempDir()

	ensureGitignore(root)

	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		".coact/locks/",
		".coact/session/",
		".coact/journal/",
		".coact/inbox/",
		".coact/terminal/",
		".coact/runs/",
		".coact/memory/",
		".coact/usage/",
		".coact/evaluations/",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf(".gitignore missing %q in:\n%s", want, content)
		}
	}
}

func TestEnsureMarkedBlockUpdatesExistingContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "AGENTS.md")
	initial := "user text\n\n<!-- coact:begin -->\nold contract\n<!-- coact:end -->\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := ensureMarkedBlock(path, "new contract")
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "user text") || !strings.Contains(content, "new contract") || strings.Contains(content, "old contract") {
		t.Fatalf("unexpected updated contract:\n%s", content)
	}
}

func TestEnsureGitignoreDoesNotDuplicateWhenCoactDirIgnored(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gitignore")
	initial := "/.coact/\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	ensureGitignore(root)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != initial {
		t.Fatalf("whole .coact ignore should already cover runtime paths; got:\n%s", string(data))
	}
}

func TestMigrateLegacyProtectedCoactGlob(t *testing.T) {
	cfg := config.Default()
	cfg.Policy.ProtectedPaths = []string{".coact/**"}

	if !migrateProtectedPaths(cfg) {
		t.Fatal("expected legacy .coact/** policy to migrate")
	}
	for _, protected := range cfg.Policy.ProtectedPaths {
		if protected == ".coact/**" {
			t.Fatalf("legacy broad protection should be removed: %#v", cfg.Policy.ProtectedPaths)
		}
	}
	if !containsString(cfg.Policy.ProtectedPaths, ".coact/config.json") {
		t.Fatalf("config should remain protected: %#v", cfg.Policy.ProtectedPaths)
	}
}

func TestMigrateVersionAddsNewProtectedStateWithoutDroppingCustomPaths(t *testing.T) {
	cfg := config.Default()
	cfg.Version = "0.1"
	cfg.Policy.ProtectedPaths = []string{"secrets/**"}

	if !migrateProtectedPaths(cfg) {
		t.Fatal("expected old config version to migrate")
	}
	for _, want := range []string{"secrets/**", ".coact/usage/**", ".coact/evaluations/**"} {
		if !containsString(cfg.Policy.ProtectedPaths, want) {
			t.Fatalf("migration missing %q: %#v", want, cfg.Policy.ProtectedPaths)
		}
	}
	if cfg.Version != config.Default().Version {
		t.Fatalf("version = %q", cfg.Version)
	}
}

func TestMigrateRepairsMissingRequiredPathAtCurrentVersion(t *testing.T) {
	cfg := config.Default()
	cfg.Policy.ProtectedPaths = []string{".coact/config.json"}
	if !migrateProtectedPaths(cfg) {
		t.Fatal("expected missing mandatory protection to be repaired")
	}
	if !containsString(cfg.Policy.ProtectedPaths, ".coact/evaluations/**") {
		t.Fatalf("evaluation protection was not restored: %#v", cfg.Policy.ProtectedPaths)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
