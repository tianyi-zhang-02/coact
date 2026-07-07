package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The hook command MUST be a single string ("<coact> hook claude"); older Claude
// Code ignores the args-array form and runs the bare binary (exit 1, non-blocking).
func TestEnsureClaudeHookStringFormAndMigration(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Plant the OLD args-array form that 1.0.x can't run.
	old := `{"hooks":{"PreToolUse":[{"matcher":"Edit|Write","hooks":[{"type":"command","command":"/x/coact","args":["hook","claude"]}]}]}}`
	if err := os.WriteFile(filepath.Join(root, ".claude", "settings.json"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := ensureClaudeHook(root)
	if err != nil || !changed {
		t.Fatalf("migration should rewrite the hook: changed=%v err=%v", changed, err)
	}
	cmd, args, found := wiredHookCommand(root)
	if !found {
		t.Fatal("hook not found after migration")
	}
	if len(args) != 0 {
		t.Errorf("string form must have no args, got %v", args)
	}
	if !strings.Contains(cmd, "hook claude") {
		t.Errorf("command should be the string form, got %q", cmd)
	}

	// Now idempotent — the correct form is a no-op.
	if again, _ := ensureClaudeHook(root); again {
		t.Error("second ensure on the correct form should be a no-op")
	}
}
