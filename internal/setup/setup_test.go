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
	} {
		if !strings.Contains(content, want) {
			t.Fatalf(".gitignore missing %q in:\n%s", want, content)
		}
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

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
