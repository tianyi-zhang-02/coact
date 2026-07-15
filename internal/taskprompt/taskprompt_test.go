package taskprompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Detail{ID: "T-012", Description: "Short dashboard label", Prompt: "Implement the full change.\n\nInclude tests."}
	if err := Write(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := Read(dir, want.ID)
	if err != nil {
		t.Fatal(err)
	}
	if *got != want {
		t.Fatalf("Read() = %#v, want %#v", got, want)
	}
	info, err := os.Stat(filepath.Join(dir, "T-012.md"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("prompt mode = %o, want 600", info.Mode().Perm())
	}
}

func TestRejectsUnsafeOrOversizedPrompt(t *testing.T) {
	for _, prompt := range []string{"", "bad\x00prompt", strings.Repeat("x", maxPromptBytes+1)} {
		if err := ValidatePrompt(prompt); err == nil {
			t.Fatalf("ValidatePrompt(%q) unexpectedly succeeded", prompt[:min(len(prompt), 20)])
		}
	}
	if err := Write(t.TempDir(), Detail{ID: "../../escape", Description: "x", Prompt: "y"}); err == nil {
		t.Fatal("path-like task id should be rejected")
	}
}
