package board

import "testing"

func TestReassign(t *testing.T) {
	b := writeBoard(t, sample) // T-014 todo/unowned, T-011 doing/claude, T-009 done/codex

	// give claude one more active task by claiming T-014
	if _, err := b.Claim("T-014", "claude", 0); err != nil {
		t.Fatal(err)
	}

	moved := b.Reassign("claude", "codex")
	if len(moved) != 2 {
		t.Fatalf("want 2 tasks reassigned (T-011, T-014), got %v", moved)
	}
	for _, tk := range b.Tasks() {
		if tk.Owner == "claude" {
			t.Fatalf("claude still owns %s after handoff", tk.ID)
		}
		if (tk.ID == "T-011" || tk.ID == "T-014") && (tk.Owner != "codex" || tk.State != "claimed") {
			t.Fatalf("task %s not handed off correctly: %+v", tk.ID, tk)
		}
	}
	// the done task stays with codex, untouched
	if got := taskByID(b, "T-009"); got.State != "done" {
		t.Fatalf("done task should be untouched: %+v", got)
	}
}
