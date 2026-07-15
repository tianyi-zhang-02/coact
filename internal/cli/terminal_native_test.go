package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
	"github.com/tianyi-zhang-02/coact/internal/setup"
	"github.com/tianyi-zhang-02/coact/internal/taskprompt"
)

func chdirInitializedProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if _, err := setup.Initialize(dir, "human"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestMentionSendsToAgentInbox(t *testing.T) {
	dir := chdirInitializedProject(t)

	if code := Run([]string{"@claude", "--agent", "codex", "please review"}); code != 0 {
		t.Fatalf("Run(@claude) = %d", code)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", "claude.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "from codex") || !strings.Contains(content, "please review") {
		t.Fatalf("unexpected inbox content:\n%s", content)
	}
}

func TestMentionAllSkipsSender(t *testing.T) {
	dir := chdirInitializedProject(t)
	sessionDir := filepath.Join(dir, ".coact", "session")
	for _, agent := range []string{"claude", "codex", "antigravity"} {
		if err := presence.Register(sessionDir, agent, "working"); err != nil {
			t.Fatal(err)
		}
	}

	if code := Run([]string{"@all", "--agent", "codex", "planning starts now"}); code != 0 {
		t.Fatalf("Run(@all) = %d", code)
	}

	for _, agent := range []string{"claude", "antigravity"} {
		data, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", agent+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "planning starts now") {
			t.Fatalf("%s inbox missing broadcast:\n%s", agent, string(data))
		}
	}
	if data, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", "codex.md")); err == nil && strings.TrimSpace(string(data)) != "" {
		t.Fatalf("sender should not receive @all broadcast:\n%s", string(data))
	}
}

func TestMentionAllOnlyTargetsLiveWorkspaceAgents(t *testing.T) {
	dir := chdirInitializedProject(t)
	if err := presence.Register(filepath.Join(dir, ".coact", "session"), "claude", "working"); err != nil {
		t.Fatal(err)
	}

	if code := Run([]string{"@all", "--agent", "human", "live only"}); code != 0 {
		t.Fatalf("Run(@all) = %d", code)
	}

	claudeInbox, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", "claude.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(claudeInbox), "live only") {
		t.Fatalf("live claude should receive broadcast:\n%s", string(claudeInbox))
	}
	for _, agent := range []string{"codex", "antigravity"} {
		if data, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", agent+".md")); err == nil && strings.TrimSpace(string(data)) != "" {
			t.Fatalf("offline %s should not receive @all broadcast:\n%s", agent, string(data))
		}
	}
}

func TestPlanCreatesRunAndNotifiesParticipants(t *testing.T) {
	dir := chdirInitializedProject(t)

	code := Run([]string{"plan", "--id", "r-001", "--with", "codex,claude", "--distributor", "claude", "Build auth safely"})
	if code != 0 {
		t.Fatalf("Run(plan) = %d", code)
	}

	for _, rel := range []string{
		".coact/runs/r-001/brief.md",
		".coact/runs/r-001/final-plan.md",
		".coact/runs/r-001/proposals/codex.md",
		".coact/runs/r-001/proposals/claude.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
	brief, err := os.ReadFile(filepath.Join(dir, ".coact", "runs", "r-001", "brief.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(brief), "Build auth safely") || !strings.Contains(string(brief), "final_task_distributor: claude") {
		t.Fatalf("unexpected brief:\n%s", string(brief))
	}
	proposal, err := os.ReadFile(filepath.Join(dir, ".coact", "runs", "r-001", "proposals", "claude.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(proposal), "Status: draft") {
		t.Fatalf("proposal should start as draft:\n%s", string(proposal))
	}
	for _, agent := range []string{"codex", "claude"} {
		data, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", agent+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "Planning phase r-001 started") {
			t.Fatalf("%s inbox missing plan notice:\n%s", agent, string(data))
		}
		if !strings.Contains(string(data), "coact plan status r-001") {
			t.Fatalf("%s inbox missing plan status readiness instruction:\n%s", agent, string(data))
		}
	}
}

func TestProposalStatusParsesCaseInsensitively(t *testing.T) {
	path := filepath.Join(t.TempDir(), "proposal.md")
	if err := os.WriteFile(path, []byte("STATUS: Ready\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := proposalStatus(path); got != "ready" {
		t.Fatalf("proposalStatus = %q, want ready", got)
	}
}

func TestParseAgentListSkipsEmptyAndInvalidIDs(t *testing.T) {
	got := parseAgentList("codex, ,../../,CLAUDE,codex")
	if strings.Join(got, ",") != "codex,claude" {
		t.Fatalf("parseAgentList = %v", got)
	}
}

func TestPlanReadyMarksOwnProposalAndNotifiesDistributor(t *testing.T) {
	dir := chdirInitializedProject(t)
	if code := Run([]string{"plan", "--id", "r-002", "--with", "codex,claude", "--distributor", "claude", "Plan safely"}); code != 0 {
		t.Fatalf("Run(plan) = %d", code)
	}
	if code := Run([]string{"plan", "ready", "--agent", "codex", "r-002"}); code != 0 {
		t.Fatalf("Run(plan ready) = %d", code)
	}
	proposal, err := os.ReadFile(filepath.Join(dir, ".coact", "runs", "r-002", "proposals", "codex.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(proposal), "Status: ready") {
		t.Fatalf("proposal not ready:\n%s", proposal)
	}
	inbox, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", "claude.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(inbox), "Proposal ready") {
		t.Fatalf("distributor was not notified:\n%s", inbox)
	}
}

func TestPlanFinalizeCreatesAssignedTasksAndIsIdempotent(t *testing.T) {
	dir := chdirInitializedProject(t)
	if code := Run([]string{"plan", "--id", "r-003", "--with", "codex,claude", "--distributor", "claude", "--approval", "auto", "Ship safely"}); code != 0 {
		t.Fatalf("Run(plan) = %d", code)
	}
	for _, agent := range []string{"codex", "claude"} {
		if code := Run([]string{"plan", "ready", "--agent", agent, "r-003"}); code != 0 {
			t.Fatalf("Run(plan ready %s) = %d", agent, code)
		}
	}
	finalPath := filepath.Join(dir, ".coact", "runs", "r-003", "final-plan.md")
	final := `# Final plan for r-003

Status: pending
Distributor: claude

## Decision

Proceed with two independently owned tasks.

## Execution tasks

- [codex] Implement the coordination endpoint
- [claude] Review safety and documentation
- [unassigned] Run release smoke tests
`
	if err := os.WriteFile(finalPath, []byte(final), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-003"}); code != 0 {
		t.Fatalf("Run(plan finalize) = %d", code)
	}

	sharedBoard, err := board.Load(filepath.Join(dir, ".coact", "board.md"))
	if err != nil {
		t.Fatal(err)
	}
	tasks := sharedBoard.Tasks()
	if len(tasks) != 4 {
		t.Fatalf("board has %d tasks, want init example + 3 finalized tasks", len(tasks))
	}
	assertTask := func(owner, state, title string) {
		t.Helper()
		for _, task := range tasks {
			if task.Title != title {
				continue
			}
			if task.Owner != owner || task.State != state {
				t.Fatalf("task = %+v, want owner=%q state=%q title=%q", task, owner, state, title)
			}
			return
		}
		t.Fatalf("task %q not found", title)
	}
	assertTask("", "todo", "Example: describe a task here")
	assertTask("", "todo", "Run release smoke tests")
	assertTask("claude", "claimed", "Review safety and documentation")
	assertTask("codex", "claimed", "Implement the coordination endpoint")

	finalData, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(finalData), "Status: finalized") || !strings.Contains(string(finalData), "## Created board tasks") {
		t.Fatalf("final plan was not recorded:\n%s", finalData)
	}
	codexInbox, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", "codex.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(codexInbox), "Your assigned tasks") || !strings.Contains(string(codexInbox), "Implement the coordination endpoint") {
		t.Fatalf("codex did not receive assignment:\n%s", codexInbox)
	}
	records, err := journal.ReadRecent(filepath.Join(dir, ".coact", "journal"), 20)
	if err != nil {
		t.Fatal(err)
	}
	foundFinish := false
	for _, record := range records {
		if record["event"] == "plan.finish" && record["id"] == "r-003" && record["tasks"] == "T-002,T-003,T-004" {
			foundFinish = true
			break
		}
	}
	if !foundFinish {
		t.Fatalf("plan.finish audit record missing: %+v", records)
	}

	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-003"}); code == 0 {
		t.Fatal("second finalize should be rejected")
	}
	sharedBoard, err = board.Load(filepath.Join(dir, ".coact", "board.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(sharedBoard.Tasks()); got != 4 {
		t.Fatalf("duplicate finalize created tasks: got %d", got)
	}
}

func TestPlanFinalizeRequiresDistributorAndReadyUnlockedProposals(t *testing.T) {
	dir := chdirInitializedProject(t)
	if code := Run([]string{"plan", "--id", "r-004", "--with", "codex,claude", "--distributor", "claude", "--approval", "auto", "Coordinate safely"}); code != 0 {
		t.Fatalf("Run(plan) = %d", code)
	}
	finalPath := filepath.Join(dir, ".coact", "runs", "r-004", "final-plan.md")
	final := `# Final plan for r-004

Status: pending
Distributor: claude

## Execution tasks

- [codex] Implement it
`
	if err := os.WriteFile(finalPath, []byte(final), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{"plan", "finalize", "--agent", "codex", "r-004"}); code == 0 {
		t.Fatal("non-distributor should not finalize")
	}
	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-004"}); code == 0 {
		t.Fatal("draft proposals should block finalize")
	}
	for _, agent := range []string{"codex", "claude"} {
		if code := Run([]string{"plan", "ready", "--agent", agent, "r-004"}); code != 0 {
			t.Fatalf("Run(plan ready %s) = %d", agent, code)
		}
	}

	p := &project.Project{Root: dir, CheckoutRoot: dir}
	cfg, err := config.Load(p.ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	manager := lockmgr.New(p, cfg)
	proposalPath := filepath.Join(dir, ".coact", "runs", "r-004", "proposals", "codex.md")
	result, err := manager.Acquire("codex", proposalPath)
	if err != nil || !result.Acquired {
		t.Fatalf("acquire proposal lock: result=%+v err=%v", result, err)
	}
	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-004"}); code == 0 {
		t.Fatal("locked proposal should block finalize")
	}
	if err := manager.Release("codex", proposalPath); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-004"}); code != 0 {
		t.Fatalf("ready unlocked plan should finalize, code=%d", code)
	}
}

func TestPlanFinalizeRejectsUnknownTaskOwner(t *testing.T) {
	dir := chdirInitializedProject(t)
	if code := Run([]string{"plan", "--id", "r-005", "--with", "codex,claude", "--distributor", "claude", "--approval", "auto", "Keep ownership explicit"}); code != 0 {
		t.Fatalf("Run(plan) = %d", code)
	}
	for _, agent := range []string{"codex", "claude"} {
		if code := Run([]string{"plan", "ready", "--agent", agent, "r-005"}); code != 0 {
			t.Fatalf("Run(plan ready %s) = %d", agent, code)
		}
	}
	finalPath := filepath.Join(dir, ".coact", "runs", "r-005", "final-plan.md")
	final := "# Final plan\n\nStatus: pending\n\n## Execution tasks\n\n- [retired-agent] Surprise work\n"
	if err := os.WriteFile(finalPath, []byte(final), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-005"}); code == 0 {
		t.Fatal("owner outside planning participants should be rejected")
	}
	sharedBoard, err := board.Load(filepath.Join(dir, ".coact", "board.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(sharedBoard.Tasks()); got != 1 {
		t.Fatalf("rejected finalize mutated board: %d tasks", got)
	}
}

func TestPlanReviewGateAndFullTaskPrompt(t *testing.T) {
	dir := chdirInitializedProject(t)
	if code := Run([]string{"plan", "--id", "r-006", "--with", "codex,claude", "--lead", "claude", "Ship with review"}); code != 0 {
		t.Fatalf("Run(plan) = %d", code)
	}
	for _, agent := range []string{"codex", "claude"} {
		if code := Run([]string{"plan", "ready", "--agent", agent, "r-006"}); code != 0 {
			t.Fatalf("Run(plan ready %s) = %d", agent, code)
		}
	}
	finalPath := filepath.Join(dir, ".coact", "runs", "r-006", "final-plan.md")
	final := `# Final plan

Status: pending
Distributor: claude

## Execution tasks

- [codex] Add reviewed task workflow
  Prompt: Implement the approved workflow, add focused tests, and preserve compatibility.
`
	if err := os.WriteFile(finalPath, []byte(final), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-006"}); code == 0 {
		t.Fatal("review-gated plan should reject direct finalize")
	}
	if code := Run([]string{"plan", "submit", "--agent", "claude", "r-006"}); code != 0 {
		t.Fatalf("Run(plan submit) = %d", code)
	}
	if got := documentStatus(finalPath); got != "review" {
		t.Fatalf("status after submit = %q, want review", got)
	}
	if code := Run([]string{"plan", "approve", "--agent", "human", "r-006"}); code != 0 {
		t.Fatalf("Run(plan approve) = %d", code)
	}
	if code := Run([]string{"plan", "finalize", "--agent", "claude", "r-006"}); code != 0 {
		t.Fatalf("Run(plan finalize) = %d", code)
	}
	detail, err := taskprompt.Read(filepath.Join(dir, ".coact", "tasks"), "T-002")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Description != "Add reviewed task workflow" || !strings.Contains(detail.Prompt, "focused tests") {
		t.Fatalf("task detail = %#v", detail)
	}
}

func TestTaskAddSeparatesDescriptionFromPrompt(t *testing.T) {
	dir := chdirInitializedProject(t)
	if code := Run([]string{"task", "add", "--owner", "codex", "--prompt", "Implement the endpoint and run the focused tests.", "Add endpoint"}); code != 0 {
		t.Fatalf("Run(task add) = %d", code)
	}
	sharedBoard, err := board.Load(filepath.Join(dir, ".coact", "board.md"))
	if err != nil {
		t.Fatal(err)
	}
	tasks := sharedBoard.Tasks()
	if len(tasks) != 2 || tasks[0].Title != "Add endpoint" || tasks[0].Owner != "codex" {
		t.Fatalf("tasks = %#v", tasks)
	}
	detail, err := taskprompt.Read(filepath.Join(dir, ".coact", "tasks"), tasks[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Description != "Add endpoint" || !strings.Contains(detail.Prompt, "focused tests") {
		t.Fatalf("detail = %#v", detail)
	}
	inboxData, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", "codex.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(inboxData), detail.Prompt) {
		t.Fatalf("assigned prompt missing from inbox: %s", inboxData)
	}
}
