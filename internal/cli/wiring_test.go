package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkedBlockRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CLAUDE.md")
	if err := os.WriteFile(path, []byte("# Mine\n\nkeep this\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	added, err := ensureMarkedBlock(path, "coact contract body")
	if err != nil || !added {
		t.Fatalf("add: err=%v added=%v", err, added)
	}
	if added, _ := ensureMarkedBlock(path, "coact contract body"); added {
		t.Fatal("second ensure should be a no-op")
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "coact contract body") {
		t.Fatal("block missing after add")
	}
	if !strings.Contains(string(data), "keep this") {
		t.Fatal("original content lost")
	}

	removed, err := removeMarkedBlock(path)
	if err != nil || !removed {
		t.Fatalf("remove: err=%v removed=%v", err, removed)
	}
	data, _ = os.ReadFile(path)
	if strings.Contains(string(data), coactBegin) || strings.Contains(string(data), "coact contract body") {
		t.Fatal("block or marker left behind after remove")
	}
	if !strings.Contains(string(data), "keep this") {
		t.Fatal("original content lost after remove")
	}
}

func TestRemoveClaudeHookRoundTrip(t *testing.T) {
	root := t.TempDir()
	if _, err := ensureClaudeHook(root); err != nil {
		t.Fatal(err)
	}
	if _, _, found := wiredHookCommand(root); !found {
		t.Fatal("hook should be wired after ensure")
	}
	removed, err := removeClaudeHook(root)
	if err != nil || !removed {
		t.Fatalf("remove hook: err=%v removed=%v", err, removed)
	}
	if _, _, found := wiredHookCommand(root); found {
		t.Fatal("hook should be gone after remove")
	}
}
