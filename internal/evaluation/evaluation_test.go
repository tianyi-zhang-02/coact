package evaluation

import (
	"strings"
	"testing"
	"time"
)

func TestSaveLoadAndBuildReport(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rating := Rating{
		Run: "r-001", Reviewer: "codex", Peer: "claude", Model: "opus",
		Scores: Scores{Overall: 4, Cooperation: 5, CodeQuality: 4, Responsiveness: 3, Alignment: 5},
		Note:   "clear review",
	}
	if err := SaveRating(dir, rating, now); err != nil {
		t.Fatal(err)
	}
	ratings, err := LoadRatings(dir, "r-001")
	if err != nil || len(ratings) != 1 {
		t.Fatalf("ratings=%#v err=%v", ratings, err)
	}
	records := []map[string]string{
		{"ts": now.Add(-time.Minute).Format(time.RFC3339), "agent": "human", "event": "plan.start", "id": "r-001"},
		{"ts": now.Format(time.RFC3339), "agent": "codex", "event": "msg.send", "to": "claude"},
		{"ts": now.Add(2 * time.Minute).Format(time.RFC3339), "agent": "claude", "event": "task.claim"},
		{"ts": now.Add(3 * time.Minute).Format(time.RFC3339), "agent": "claude", "event": "task.finish"},
	}
	report := BuildReport("r-001", records, ratings, now.Add(4*time.Minute))
	markdown := Markdown(report)
	for _, want := range []string{"claude", "1/1", "2m (1)", "4.0/5", "clear review"} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("report missing %q:\n%s", want, markdown)
		}
	}
}

func TestSaveRatingRejectsUnsafeOrInvalidInput(t *testing.T) {
	base := Rating{Run: "run", Reviewer: "codex", Peer: "claude", Scores: Scores{1, 1, 1, 1, 1}}
	bad := base
	bad.Run = "../run"
	if err := SaveRating(t.TempDir(), bad, time.Now()); err == nil {
		t.Fatal("unsafe run should fail")
	}
	bad = base
	bad.Scores.Overall = 6
	if err := SaveRating(t.TempDir(), bad, time.Now()); err == nil {
		t.Fatal("invalid score should fail")
	}
}
