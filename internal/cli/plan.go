package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func cmdPlan(args []string) int {
	if len(args) > 0 && args[0] == "status" {
		return cmdPlanStatus(args[1:])
	}

	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "initiator id")
	withFlag := fs.String("with", "codex,claude", "comma-separated planning agents")
	distributorFlag := fs.String("distributor", "", "final task distributor (human, codex, claude, gemini, vote)")
	idFlag := fs.String("id", "", "run id (default: timestamp)")
	pos, err := parseInterspersed(fs, args)
	if err != nil || len(pos) == 0 {
		fmt.Fprintln(os.Stderr, `usage: coact plan [--with codex,claude] [--distributor codex] "brief..."`)
		return 2
	}
	brief := strings.TrimSpace(strings.Join(pos, " "))
	if brief == "" {
		fmt.Fprintln(os.Stderr, `usage: coact plan [--with codex,claude] [--distributor codex] "brief..."`)
		return 2
	}

	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	initiator := agentID(*agentFlag)
	participants := parseAgentList(*withFlag)
	if len(participants) == 0 {
		fmt.Fprintln(os.Stderr, "coact plan: --with must name at least one agent")
		return 2
	}
	distributor := sanitizeAgent(*distributorFlag)
	if *distributorFlag == "" {
		distributor = defaultDistributor(p)
	}
	runID := safeRunID(*idFlag)
	if runID == "" {
		runID = "run-" + time.Now().UTC().Format("20060102-150405")
	}

	runDir := filepath.Join(p.RunsDir(), runID)
	if err := os.MkdirAll(filepath.Join(runDir, "proposals"), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "coact plan: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(filepath.Join(runDir, "notes"), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "coact plan: %v\n", err)
		return 1
	}
	if err := writePlanFiles(p, runDir, runID, brief, initiator, distributor, participants); err != nil {
		fmt.Fprintf(os.Stderr, "coact plan: %v\n", err)
		return 1
	}
	for _, agent := range participants {
		if err := inbox.Send(p.InboxDir(), initiator, agent, planMessage(runID, distributor, agent)); err != nil {
			fmt.Fprintf(os.Stderr, "coact plan: %v\n", err)
			return 1
		}
	}
	_ = journal.Append(p.JournalDir(), initiator, "plan.start", map[string]string{
		"id":          runID,
		"with":        strings.Join(participants, ","),
		"distributor": distributor,
	})
	fmt.Printf("planning run %s created\n", runID)
	fmt.Printf("brief: .coact/runs/%s/brief.md\n", runID)
	fmt.Printf("participants: %s\n", strings.Join(participants, ", "))
	fmt.Printf("final distributor: %s\n", distributor)
	return 0
}

func cmdPlanStatus(args []string) int {
	fs := flag.NewFlagSet("plan status", flag.ContinueOnError)
	pos, err := parseInterspersed(fs, args)
	if err != nil {
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	runID := ""
	if len(pos) > 0 {
		runID = safeRunID(pos[0])
	} else {
		runID = latestRunID(p)
	}
	if runID == "" {
		fmt.Println("no planning runs yet")
		return 0
	}
	runDir := filepath.Join(p.RunsDir(), runID)
	fmt.Printf("planning run: %s\n", runID)
	for _, rel := range []string{"brief.md", "final-plan.md"} {
		path := filepath.Join(runDir, rel)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  exists  .coact/runs/%s/%s\n", runID, rel)
		} else {
			fmt.Printf("  missing .coact/runs/%s/%s\n", runID, rel)
		}
	}
	for _, dir := range []string{"proposals", "notes"} {
		matches, _ := filepath.Glob(filepath.Join(runDir, dir, "*.md"))
		fmt.Printf("  %s: %d file(s)\n", dir, len(matches))
	}
	fmt.Println()
	fmt.Println("proposal readiness:")
	m := lockmgr.New(p, cfg)
	locks, _ := m.List()
	proposals, _ := filepath.Glob(filepath.Join(runDir, "proposals", "*.md"))
	if len(proposals) == 0 {
		fmt.Println("  (none)")
	}
	for _, proposal := range proposals {
		agent := strings.TrimSuffix(filepath.Base(proposal), ".md")
		status := proposalStatus(proposal)
		lock := lockOwnerFor(locks, planRelPath(p, proposal))
		if lock == "" {
			lock = "unlocked"
		} else {
			lock = "locked by " + lock
		}
		fmt.Printf("  %-10s %-8s %s\n", agent, status, lock)
	}
	fmt.Println()
	fmt.Println("final distributor should wait until all required proposals are ready and unlocked.")
	return 0
}

func writePlanFiles(p *project.Project, runDir, runID, brief, initiator, distributor string, participants []string) error {
	doc := fmt.Sprintf(`# CoAct planning run %s

## Brief

%s

## Protocol

1. Every planning agent reads .coact/team.md, .coact/memory/project.md, and this file.
2. Every planning agent writes one proposal to .coact/runs/%s/proposals/<agent>.md.
3. Every planning agent changes proposal Status from draft to ready and releases any lock on the proposal file.
4. Every planning agent reads peer proposals and may add second-pass notes under .coact/runs/%s/notes/<agent>.md.
5. The final distributor runs coact plan status %s and waits until all required proposals are ready and unlocked.
6. The final distributor writes .coact/runs/%s/final-plan.md.
7. The final distributor creates execution tasks with coact task add "..."
8. Agents claim tasks before implementation with coact claim <id>.

## Metadata

- initiator: %s
- participants: %s
- final_task_distributor: %s
- created_at: %s

`, runID, brief, runID, runID, runID, runID, initiator, strings.Join(participants, ", "), distributor, time.Now().UTC().Format(time.RFC3339))
	if err := platform.AtomicWrite(filepath.Join(runDir, "brief.md"), []byte(doc), 0o644); err != nil {
		return err
	}
	final := fmt.Sprintf(`# Final plan for %s

Status: pending
Distributor: %s

Do not finalize until coact plan status %s shows all required proposals are ready and unlocked.
When proposals are ready, summarize the decision here and create board tasks.

`, runID, distributor, runID)
	if err := platform.AtomicWrite(filepath.Join(runDir, "final-plan.md"), []byte(final), 0o644); err != nil {
		return err
	}
	for _, agent := range participants {
		path := filepath.Join(runDir, "proposals", agent+".md")
		if _, err := os.Stat(path); err == nil {
			continue
		}
		template := fmt.Sprintf(`# Proposal: %s

Run: %s
Status: draft

Change Status to ready when this proposal is complete, then release any lock on this file.

## Proposed approach

## Risks

## Suggested tasks

`, agent, runID)
		if err := platform.AtomicWrite(path, []byte(template), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func planMessage(runID, distributor, recipient string) string {
	role := "planning participant"
	if recipient == distributor {
		role = "final task distributor"
	}
	return fmt.Sprintf(`Planning phase %s started.

Your role: %s.

Read:
- .coact/team.md
- .coact/memory/project.md
- .coact/runs/%s/brief.md

Write your proposal:
- .coact/runs/%s/proposals/%s.md

Then read peer proposals. If you are the final distributor, write:
- .coact/runs/%s/final-plan.md

Before finalizing, run:
- coact plan status %s

Only finalize after all required proposals are ready and unlocked.

After final planning, create/claim execution tasks through the board.
`, runID, role, runID, runID, recipient, runID, runID)
}

func proposalStatus(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "missing"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		key, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "status") {
			status := strings.TrimSpace(value)
			status = strings.ToLower(status)
			if status == "" {
				return "unknown"
			}
			return status
		}
	}
	return "unknown"
}

func lockOwnerFor(locks []lockmgr.Lock, relPath string) string {
	for _, lock := range locks {
		if lock.Path == relPath {
			return lock.Owner
		}
	}
	return ""
}

func planRelPath(p *project.Project, path string) string {
	if rel, err := filepath.Rel(p.Root, path); err == nil {
		return rel
	}
	return path
}

func parseAgentList(raw string) []string {
	seen := map[string]bool{}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		agent := sanitizeAgent(part)
		if agent == "" || seen[agent] {
			continue
		}
		seen[agent] = true
		out = append(out, agent)
	}
	return out
}

func defaultDistributor(p *project.Project) string {
	data, err := os.ReadFile(p.TeamPath())
	if err != nil {
		return "human"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- final_task_distributor:") {
			return sanitizeAgent(strings.TrimSpace(strings.TrimPrefix(line, "- final_task_distributor:")))
		}
	}
	return "human"
}

func safeRunID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

func latestRunID(p *project.Project) string {
	entries, err := os.ReadDir(p.RunsDir())
	if err != nil {
		return ""
	}
	latest := ""
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() > latest {
			latest = entry.Name()
		}
	}
	return latest
}
