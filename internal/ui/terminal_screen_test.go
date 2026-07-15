package ui

import (
	"strings"
	"testing"
)

func TestTerminalScreenSnapshotAppliesCursorRedraw(t *testing.T) {
	input := "\x1b[2J\x1b[HOld screen\x1b[H\x1b[2KNew\x1b[10Gscreen"
	if got, want := terminalScreenSnapshot([]byte(input)), "New      screen"; got != want {
		t.Fatalf("snapshot = %q, want %q", got, want)
	}
}

func TestTerminalScreenSnapshotKeepsLatestFullscreenFrame(t *testing.T) {
	input := "first frame\x1b[?1049h\x1b[2J\x1b[HClaude Code\x1b[2;1HOpus 4.8\x1b[3;1Hworking"
	got := terminalScreenSnapshot([]byte(input))
	if got != "Claude Code\nOpus 4.8\nworking" {
		t.Fatalf("unexpected fullscreen snapshot:\n%s", got)
	}
	if strings.Contains(got, "first frame") {
		t.Fatal("snapshot retained content cleared by alternate screen")
	}
}

func TestTerminalScreenSnapshotHandlesCarriageReturnAndWideRunes(t *testing.T) {
	input := "stale text\r当前状态\x1b[K\nready"
	got := terminalScreenSnapshot([]byte(input))
	if got != "当前状态\n        ready" {
		t.Fatalf("unexpected wide-rune snapshot: %q", got)
	}
}

func TestTerminalScreenSnapshotStripsColorWithoutRemovingSpaces(t *testing.T) {
	input := "\x1b[38;2;215;119;87mClaude Code\x1b[39m\x1b[2;5HOpus 4.8"
	got := terminalScreenSnapshot([]byte(input))
	if got != "Claude Code\n    Opus 4.8" {
		t.Fatalf("unexpected styled snapshot: %q", got)
	}
}
