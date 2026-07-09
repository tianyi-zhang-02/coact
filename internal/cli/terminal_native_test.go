package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/setup"
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
	for _, agent := range []string{"claude", "codex", "gemini"} {
		if err := presence.Register(sessionDir, agent, "working"); err != nil {
			t.Fatal(err)
		}
	}

	if code := Run([]string{"@all", "--agent", "codex", "planning starts now"}); code != 0 {
		t.Fatalf("Run(@all) = %d", code)
	}

	for _, agent := range []string{"claude", "gemini"} {
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
	for _, agent := range []string{"codex", "gemini"} {
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
	for _, agent := range []string{"codex", "claude"} {
		data, err := os.ReadFile(filepath.Join(dir, ".coact", "inbox", agent+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "Planning phase r-001 started") {
			t.Fatalf("%s inbox missing plan notice:\n%s", agent, string(data))
		}
	}
}
