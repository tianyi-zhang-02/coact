package planning

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/project"
)

func TestStartCreatesReviewRunAndNotifiesParticipants(t *testing.T) {
	root := t.TempDir()
	p := &project.Project{Root: root, CheckoutRoot: root}
	for _, dir := range []string{p.RunsDir(), p.InboxDir(), p.JournalDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	info, err := Start(p, StartOptions{
		RunID: "run-001", Brief: "Build the safe workflow", Initiator: "human",
		Lead: "codex", ApprovalMode: "review", Participants: []string{"codex", "claude"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if info.Status != "pending" || info.Lead != "codex" || info.ApprovalMode != "review" {
		t.Fatalf("info = %#v", info)
	}
	brief, err := os.ReadFile(filepath.Join(p.RunsDir(), "run-001", "brief.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"approval_mode: review", "short task description plus a full Prompt"} {
		if !strings.Contains(string(brief), want) {
			t.Fatalf("brief missing %q:\n%s", want, brief)
		}
	}
	for _, agent := range []string{"codex", "claude"} {
		message, err := os.ReadFile(filepath.Join(p.InboxDir(), agent+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(message), "Planning phase run-001 started") {
			t.Fatalf("%s was not notified: %s", agent, message)
		}
	}
}

func TestApproveRequiresSubmittedReview(t *testing.T) {
	root := t.TempDir()
	p := &project.Project{Root: root, CheckoutRoot: root}
	for _, dir := range []string{p.RunsDir(), p.InboxDir(), p.JournalDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Start(p, StartOptions{RunID: "run-002", Brief: "Review me", Initiator: "human", Lead: "claude", ApprovalMode: "review", Participants: []string{"claude"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(p, "run-002"); err == nil {
		t.Fatal("pending plan should not be approved")
	}
	finalPath := filepath.Join(p.RunsDir(), "run-002", "final-plan.md")
	if err := setStatus(finalPath, "review"); err != nil {
		t.Fatal(err)
	}
	info, err := Approve(p, "run-002")
	if err != nil {
		t.Fatal(err)
	}
	if info.Status != "approved" {
		t.Fatalf("status = %q", info.Status)
	}
}

func TestHumanLedPlanCanApprovePendingDraft(t *testing.T) {
	root := t.TempDir()
	p := &project.Project{Root: root, CheckoutRoot: root}
	for _, dir := range []string{p.RunsDir(), p.InboxDir(), p.JournalDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Start(p, StartOptions{RunID: "run-human", Brief: "Human distributes", Initiator: "human", Lead: "human", ApprovalMode: "review", Participants: []string{"codex"}}); err != nil {
		t.Fatal(err)
	}
	info, err := Approve(p, "run-human")
	if err != nil {
		t.Fatal(err)
	}
	if info.Status != "approved" {
		t.Fatalf("status = %q", info.Status)
	}
}
