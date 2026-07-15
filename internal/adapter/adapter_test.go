package adapter

import (
	"strings"
	"testing"
)

func TestRegistry(t *testing.T) {
	if len(All()) < 3 {
		t.Fatalf("want at least 3 adapters, got %d", len(All()))
	}
	for _, id := range []string{"claude", "codex", "antigravity"} {
		a, ok := Get(id)
		if !ok {
			t.Fatalf("adapter %q missing", id)
		}
		if a.Binary == "" || a.ContractFile == "" {
			t.Fatalf("adapter %q incomplete: %+v", id, a)
		}
		c := a.Contract()
		if !strings.Contains(c, id) {
			t.Errorf("%q contract missing its id", id)
		}
		if !strings.Contains(c, "coact inbox") {
			t.Errorf("%q contract missing messaging instructions", id)
		}
	}
	if _, ok := Get("nope"); ok {
		t.Fatal("Get should fail for an unknown id")
	}
	if len(Defaults()) != 3 {
		t.Fatalf("want 3 default adapters, got %d", len(Defaults()))
	}
	if _, ok := Get("gemini"); ok {
		t.Fatal("retired Gemini adapter should not remain registered")
	}
}

func TestEnforcementTiers(t *testing.T) {
	claude, _ := Get("claude")
	codex, _ := Get("codex")
	if !claude.HardHook {
		t.Error("claude should be hook-enforced (L2)")
	}
	if codex.HardHook {
		t.Error("codex should be self-enforced (L1)")
	}
	if !strings.Contains(claude.Contract(), "gated automatically") {
		t.Error("claude (L2) contract should say edits are auto-gated")
	}
	if !strings.Contains(codex.Contract(), "coact lock <path>") {
		t.Error("codex (L1) contract should tell it to lock before editing")
	}
}
