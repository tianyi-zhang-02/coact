package inbox

import (
	"strings"
	"testing"
)

func TestSendAndRead(t *testing.T) {
	dir := t.TempDir()

	if err := Send(dir, "claude", "codex", "please build the gateway"); err != nil {
		t.Fatal(err)
	}
	if err := Send(dir, "claude", "codex", "and add tests"); err != nil {
		t.Fatal(err)
	}

	// peek shows without consuming
	peeked, err := Read(dir, "codex", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(peeked, "gateway") || !strings.Contains(peeked, "add tests") {
		t.Fatalf("peek missing messages: %q", peeked)
	}

	// real read consumes
	got, err := Read(dir, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "from claude") || !strings.Contains(got, "gateway") {
		t.Fatalf("read missing content: %q", got)
	}

	// second read is empty (consumed)
	again, err := Read(dir, "codex", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(again) != "" {
		t.Fatalf("messages should have been consumed, got: %q", again)
	}

	// a different recipient has nothing
	if other, _ := Read(dir, "claude", false); strings.TrimSpace(other) != "" {
		t.Fatalf("claude should have no messages, got: %q", other)
	}
}
