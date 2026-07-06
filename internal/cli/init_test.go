package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readSettings(t *testing.T, root string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var s map[string]any
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("settings not valid JSON: %v", err)
	}
	return s
}

func preToolUse(t *testing.T, s map[string]any) []any {
	t.Helper()
	hooks, _ := s["hooks"].(map[string]any)
	pre, _ := hooks["PreToolUse"].([]any)
	return pre
}

func TestEnsureClaudeHookIdempotent(t *testing.T) {
	root := t.TempDir()

	added, err := ensureClaudeHook(root)
	if err != nil || !added {
		t.Fatalf("first wire: added=%v err=%v", added, err)
	}
	if n := len(preToolUse(t, readSettings(t, root))); n != 1 {
		t.Fatalf("after first wire want 1 PreToolUse entry, got %d", n)
	}

	// Re-running init must not duplicate the hook.
	added, err = ensureClaudeHook(root)
	if err != nil {
		t.Fatal(err)
	}
	if added {
		t.Fatal("second wire should be a no-op")
	}
	if n := len(preToolUse(t, readSettings(t, root))); n != 1 {
		t.Fatalf("after second wire want 1 PreToolUse entry, got %d", n)
	}
}

func TestEnsureClaudeHookMergesExistingSettings(t *testing.T) {
	root := t.TempDir()
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Pre-existing settings with an unrelated key that must survive.
	pre := `{"model":"opus","hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"echo hi"}]}]}}`
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(pre), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ensureClaudeHook(root); err != nil {
		t.Fatal(err)
	}
	s := readSettings(t, root)
	if s["model"] != "opus" {
		t.Errorf("unrelated key 'model' was clobbered: %v", s["model"])
	}
	if n := len(preToolUse(t, s)); n != 2 {
		t.Errorf("want 2 PreToolUse entries (existing + coact), got %d", n)
	}
}
