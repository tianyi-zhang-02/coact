package board

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = `# Task board

## Backlog

- [ ] Add rate limiting <!-- coact: id=T-014 state=todo owner= -->
- [~] Refactor auth <!-- coact: id=T-011 state=doing owner=claude ttl=1800 -->

## Done

- [x] Write schema <!-- coact: id=T-009 state=done owner=codex -->
`

func writeBoard(t *testing.T, body string) *Board {
	t.Helper()
	path := filepath.Join(t.TempDir(), "board.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParse(t *testing.T) {
	b := writeBoard(t, sample)
	tasks := b.Tasks()
	if len(tasks) != 3 {
		t.Fatalf("want 3 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != "T-014" || tasks[0].Title != "Add rate limiting" || tasks[0].State != "todo" {
		t.Errorf("bad parse of task 0: %+v", tasks[0])
	}
	if tasks[1].Owner != "claude" || tasks[1].Extra["ttl"] != "1800" {
		t.Errorf("bad parse of task 1: %+v", tasks[1])
	}
}

func TestClaimAndFinishRoundTrip(t *testing.T) {
	b := writeBoard(t, sample)

	if _, err := b.Claim("T-014", "codex", 900); err != nil {
		t.Fatalf("claim: %v", err)
	}
	// Re-parse from the mutated lines.
	got := taskByID(b, "T-014")
	if got.Owner != "codex" || got.State != "doing" {
		t.Fatalf("after claim: %+v", got)
	}
	if got.Extra["ttl"] != "900" {
		t.Errorf("ttl not recorded: %+v", got.Extra)
	}

	// Claiming an already-owned doing task by another agent fails.
	if _, err := b.Claim("T-011", "codex", 0); err == nil {
		t.Error("claiming another agent's doing task should fail")
	}

	if _, err := b.Finish("T-014", "codex"); err != nil {
		t.Fatalf("finish: %v", err)
	}
	if taskByID(b, "T-014").State != "done" {
		t.Error("task should be done after finish")
	}
}

func TestAddAssignsNextID(t *testing.T) {
	b := writeBoard(t, sample)
	nt := b.Add("New thing")
	if nt.ID != "T-015" {
		t.Errorf("want next id T-015, got %s", nt.ID)
	}
	if got := taskByID(b, "T-015"); got == nil || got.Title != "New thing" {
		t.Errorf("added task not found by re-parse: %+v", got)
	}
	// The new line must sit under the Backlog header, not in Done.
	joined := strings.Join(b.lines, "\n")
	backlogIdx := strings.Index(joined, "## Backlog")
	doneIdx := strings.Index(joined, "## Done")
	newIdx := strings.Index(joined, "New thing")
	if !(backlogIdx < newIdx && newIdx < doneIdx) {
		t.Errorf("new task not placed under Backlog (backlog=%d new=%d done=%d)", backlogIdx, newIdx, doneIdx)
	}
}

func TestAssignReservesTaskUntilClaim(t *testing.T) {
	b := writeBoard(t, sample)
	assigned, err := b.Assign("T-014", "codex")
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	if assigned.Owner != "codex" || assigned.State != "claimed" {
		t.Fatalf("after assign: %+v", assigned)
	}
	if _, err := b.Claim("T-014", "claude", 900); err == nil {
		t.Fatal("another agent should not claim a reserved task")
	}
	claimed, err := b.Claim("T-014", "codex", 900)
	if err != nil {
		t.Fatalf("owner claim: %v", err)
	}
	if claimed.State != "doing" || claimed.Extra["ttl"] != "900" {
		t.Fatalf("after owner claim: %+v", claimed)
	}
}

func TestFinishRequiresStartedOwnedTask(t *testing.T) {
	b := writeBoard(t, sample)
	if _, err := b.Finish("T-014", "codex"); err == nil {
		t.Fatal("unstarted task should not be finishable")
	}
	if _, err := b.Assign("T-014", "codex"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Finish("T-014", "codex"); err == nil {
		t.Fatal("assigned task should still require claim before finish")
	}
	if _, err := b.Claim("T-014", "codex", 900); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Finish("T-014", "claude"); err == nil {
		t.Fatal("non-owner should not finish active work")
	}
}

func TestUnassignAndReopen(t *testing.T) {
	b := writeBoard(t, sample)
	if _, err := b.Assign("T-014", "codex"); err != nil {
		t.Fatal(err)
	}
	unassigned, err := b.Unassign("T-014")
	if err != nil {
		t.Fatal(err)
	}
	if unassigned.State != "todo" || unassigned.Owner != "" {
		t.Fatalf("unexpected unassigned task: %+v", unassigned)
	}
	if _, err := b.Reopen("T-014"); err == nil {
		t.Fatal("todo task should not reopen")
	}
	reopened, err := b.Reopen("T-009")
	if err != nil {
		t.Fatal(err)
	}
	if reopened.State != "todo" || reopened.Owner != "" {
		t.Fatalf("unexpected reopened task: %+v", reopened)
	}
}

func TestValidateTitleRejectsMetadataInjection(t *testing.T) {
	for _, title := range []string{"", "line one\nline two", "safe <!-- coact: id=T-999 -->"} {
		if err := ValidateTitle(title); err == nil {
			t.Fatalf("ValidateTitle(%q) should fail", title)
		}
	}
	if err := ValidateTitle("Review authentication safety"); err != nil {
		t.Fatalf("safe title rejected: %v", err)
	}
}

func taskByID(b *Board, id string) *Task {
	for _, t := range b.Tasks() {
		if t.ID == id {
			return t
		}
	}
	return nil
}
