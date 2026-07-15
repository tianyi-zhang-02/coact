package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/tianyi-zhang-02/coact/internal/board"
	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func cmdPlan(args []string) int {
	if len(args) > 0 && args[0] == "finalize" {
		return cmdPlanFinalize(args[1:])
	}
	if len(args) > 0 && args[0] == "status" {
		return cmdPlanStatus(args[1:])
	}
	if len(args) > 0 && args[0] == "ready" {
		return cmdPlanReady(args[1:])
	}

	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "initiator id")
	withFlag := fs.String("with", "codex,claude", "comma-separated planning agents")
	distributorFlag := fs.String("distributor", "", "final task distributor (human, codex, claude, antigravity, vote)")
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

type executionTask struct {
	Owner string
	Title string
}

type createdExecutionTask struct {
	ID    string
	Owner string
	Title string
}

func cmdPlanFinalize(args []string) int {
	fs := flag.NewFlagSet("plan finalize", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "final distributor id")
	positionals, err := parseInterspersed(fs, args)
	if err != nil || len(positionals) > 1 {
		fmt.Fprintln(os.Stderr, "usage: coact plan finalize [--agent id] [run-id]")
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	runID := ""
	if len(positionals) == 1 {
		runID = safeRunID(positionals[0])
	} else {
		runID = latestRunID(p)
	}
	if runID == "" {
		fmt.Fprintln(os.Stderr, "coact plan finalize: no planning run found")
		return 1
	}

	runDir := filepath.Join(p.RunsDir(), runID)
	briefPath := filepath.Join(runDir, "brief.md")
	finalPath := filepath.Join(runDir, "final-plan.md")
	distributor := planDistributor(briefPath)
	actor := agentID(*agentFlag)
	if !canFinalizePlan(actor, distributor) {
		fmt.Fprintf(os.Stderr, "coact plan finalize: %s is not the final distributor (%s)\n", actor, distributor)
		return 1
	}

	manager := lockmgr.New(p, cfg)
	result, err := manager.Acquire(actor, runDir)
	if err != nil || !result.Acquired {
		detail := "planning run lock denied"
		if err != nil {
			detail = err.Error()
		} else if result.Detail != "" {
			detail = result.Detail
		} else if result.Conflict != nil {
			detail = "locked by " + result.Conflict.Owner
		}
		fmt.Fprintf(os.Stderr, "coact plan finalize: %s\n", detail)
		return 1
	}
	defer func() { _ = manager.Release(actor, runDir) }()

	status := documentStatus(finalPath)
	if status == "finalized" {
		fmt.Fprintf(os.Stderr, "coact plan finalize: %s is already finalized\n", runID)
		return 1
	}
	if status != "pending" {
		fmt.Fprintf(os.Stderr, "coact plan finalize: final plan status is %s, not pending\n", status)
		return 1
	}
	participants := planParticipants(briefPath)
	if len(participants) == 0 {
		fmt.Fprintln(os.Stderr, "coact plan finalize: planning brief has no participants")
		return 1
	}
	locks, err := manager.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact plan finalize: reading locks: %v\n", err)
		return 1
	}
	for _, participant := range participants {
		proposal := filepath.Join(runDir, "proposals", participant+".md")
		if status := proposalStatus(proposal); status != "ready" {
			fmt.Fprintf(os.Stderr, "coact plan finalize: proposal for %s is %s, not ready\n", participant, status)
			return 1
		}
		if owner := lockOwnerFor(locks, planRelPath(p, proposal)); owner != "" {
			fmt.Fprintf(os.Stderr, "coact plan finalize: proposal for %s is still locked by %s\n", participant, owner)
			return 1
		}
	}

	data, err := os.ReadFile(finalPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact plan finalize: %v\n", err)
		return 1
	}
	tasks, err := parseExecutionTasks(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact plan finalize: %v\n", err)
		return 1
	}
	allowedOwners := make(map[string]bool, len(participants))
	for _, participant := range participants {
		allowedOwners[participant] = true
	}
	for _, task := range tasks {
		if task.Owner != "" && !allowedOwners[task.Owner] {
			fmt.Fprintf(os.Stderr, "coact plan finalize: task owner %q is not a planning participant\n", task.Owner)
			return 1
		}
	}

	created := make([]createdExecutionTask, 0, len(tasks))
	err = withBoardLock(p, func() error {
		originalBoard, err := os.ReadFile(p.BoardPath())
		if err != nil {
			return err
		}
		sharedBoard, err := board.Load(p.BoardPath())
		if err != nil {
			return err
		}
		for _, task := range tasks {
			added := sharedBoard.Add(task.Title)
			createdTask := createdExecutionTask{ID: added.ID, Owner: task.Owner, Title: task.Title}
			if task.Owner != "" {
				assigned, err := sharedBoard.Assign(added.ID, task.Owner)
				if err != nil {
					return err
				}
				createdTask.ID = assigned.ID
			}
			created = append(created, createdTask)
		}
		if err := sharedBoard.Save(); err != nil {
			return err
		}
		if err := markPlanFinalized(finalPath, created); err != nil {
			if rollbackErr := platform.AtomicWrite(p.BoardPath(), originalBoard, 0o644); rollbackErr != nil {
				return fmt.Errorf("updating final plan: %v; rolling back board: %v", err, rollbackErr)
			}
			return fmt.Errorf("updating final plan: %w (board rolled back)", err)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact plan finalize: committing plan: %v\n", err)
		return 1
	}

	taskIDs := make([]string, 0, len(created))
	for _, task := range created {
		taskIDs = append(taskIDs, task.ID)
		_ = journal.Append(p.JournalDir(), actor, "task.add", map[string]string{
			"id": task.ID, "owner": task.Owner, "source": "plan:" + runID,
		})
	}
	_ = journal.Append(p.JournalDir(), actor, "plan.finish", map[string]string{
		"id": runID, "count": strconv.Itoa(len(created)), "tasks": strings.Join(taskIDs, ","),
	})
	notifyPlanFinalized(p, runID, actor, participants, created)

	fmt.Printf("planning run %s finalized by %s\n", runID, actor)
	for _, task := range created {
		owner := task.Owner
		if owner == "" {
			owner = "unassigned"
		}
		fmt.Printf("  %s  %-10s %s\n", task.ID, owner, task.Title)
	}
	return 0
}

func cmdPlanReady(args []string) int {
	fs := flag.NewFlagSet("plan ready", flag.ContinueOnError)
	agentFlag := fs.String("agent", "", "planning agent id")
	positionals, err := parseInterspersed(fs, args)
	if err != nil || len(positionals) > 1 {
		fmt.Fprintln(os.Stderr, "usage: coact plan ready [--agent id] [run-id]")
		return 2
	}
	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	runID := ""
	if len(positionals) == 1 {
		runID = safeRunID(positionals[0])
	} else {
		runID = latestRunID(p)
	}
	if runID == "" {
		fmt.Fprintln(os.Stderr, "coact plan ready: no planning run found")
		return 1
	}
	agent := agentID(*agentFlag)
	proposal := filepath.Join(p.RunsDir(), runID, "proposals", agent+".md")
	if _, err := os.Stat(proposal); err != nil {
		fmt.Fprintf(os.Stderr, "coact plan ready: no proposal for %s in %s\n", agent, runID)
		return 1
	}
	m := lockmgr.New(p, cfg)
	result, err := m.Acquire(agent, proposal)
	if err != nil || !result.Acquired {
		detail := "proposal lock denied"
		if err != nil {
			detail = err.Error()
		} else if result.Detail != "" {
			detail = result.Detail
		}
		fmt.Fprintf(os.Stderr, "coact plan ready: %s\n", detail)
		return 1
	}
	if err := setProposalStatus(proposal, "ready"); err != nil {
		_ = m.Release(agent, proposal)
		fmt.Fprintf(os.Stderr, "coact plan ready: %v\n", err)
		return 1
	}
	if err := m.Release(agent, proposal); err != nil {
		fmt.Fprintf(os.Stderr, "coact plan ready: marked ready but could not release lock: %v\n", err)
		return 1
	}
	distributor := planDistributor(filepath.Join(p.RunsDir(), runID, "brief.md"))
	if distributor != "" && distributor != agent {
		_ = inbox.Send(p.InboxDir(), agent, distributor, fmt.Sprintf("Proposal ready: .coact/runs/%s/proposals/%s.md. Run `coact plan status %s` before finalizing.", runID, agent, runID))
	}
	_ = journal.Append(p.JournalDir(), agent, "plan.ready", map[string]string{"id": runID})
	fmt.Printf("proposal ready: %s (%s)\n", runID, agent)
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
3. Every planning agent runs coact plan ready %s after finishing its proposal; this marks it ready and releases the command's lock.
4. Every planning agent reads peer proposals and may add second-pass notes under .coact/runs/%s/notes/<agent>.md.
5. The final distributor runs coact plan status %s and waits until all required proposals are ready and unlocked.
6. The final distributor writes .coact/runs/%s/final-plan.md and lists tasks under the Execution tasks heading.
7. The final distributor runs coact plan finalize %s to create and assign board tasks.
8. Agents claim tasks before implementation with coact claim <id>.

## Metadata

- initiator: %s
- participants: %s
- final_task_distributor: %s
- created_at: %s

`, runID, brief, runID, runID, runID, runID, runID, runID, initiator, strings.Join(participants, ", "), distributor, time.Now().UTC().Format(time.RFC3339))
	if err := platform.AtomicWrite(filepath.Join(runDir, "brief.md"), []byte(doc), 0o644); err != nil {
		return err
	}
	final := fmt.Sprintf(`# Final plan for %s

Status: pending
Distributor: %s

Do not finalize until coact plan status %s shows all required proposals are ready and unlocked.
When proposals are ready, summarize the decision and replace the examples below.

## Decision

## Execution tasks

<!-- Add one task per line. Example: - [codex] Implement the agreed change -->
<!-- Use [unassigned] when the distributor wants an agent to claim it later. -->

Run: coact plan finalize %s

`, runID, distributor, runID, runID)
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

When complete, run: coact plan ready %s

## Proposed approach

## Risks

## Suggested tasks

`, agent, runID, runID)
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

When complete:
- coact plan ready %s

Then read peer proposals. If you are the final distributor, write:
- .coact/runs/%s/final-plan.md

Before finalizing, run:
- coact plan status %s

Only finalize after all required proposals are ready and unlocked.

After writing the Execution tasks section, run:
- coact plan finalize %s

Assigned agents then claim their task through the board before editing.
`, runID, role, runID, runID, recipient, runID, runID, runID, runID)
}

func canFinalizePlan(actor, distributor string) bool {
	if distributor == "human" || distributor == "vote" {
		return actor == "human"
	}
	return distributor != "" && actor == distributor
}

func planParticipants(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
		if key, value, ok := strings.Cut(line, ":"); ok && strings.EqualFold(strings.TrimSpace(key), "participants") {
			return parseAgentList(value)
		}
	}
	return nil
}

func documentStatus(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "missing"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if key, value, ok := strings.Cut(strings.TrimSpace(line), ":"); ok && strings.EqualFold(strings.TrimSpace(key), "status") {
			status := strings.ToLower(strings.TrimSpace(value))
			if status == "" {
				return "unknown"
			}
			return status
		}
	}
	return "unknown"
}

func parseExecutionTasks(content string) ([]executionTask, error) {
	inSection := false
	var tasks []executionTask
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if strings.HasPrefix(line, "## ") {
			if inSection {
				break
			}
			inSection = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "## ")), "Execution tasks")
			continue
		}
		if !inSection || !strings.HasPrefix(line, "- ") {
			continue
		}
		body := strings.TrimSpace(strings.TrimPrefix(line, "- "))
		if !strings.HasPrefix(body, "[") {
			return nil, fmt.Errorf("execution task must use `- [owner] title`: %s", line)
		}
		end := strings.Index(body, "]")
		if end < 0 {
			return nil, fmt.Errorf("execution task has an unclosed owner: %s", line)
		}
		rawOwner := strings.ToLower(strings.TrimSpace(body[1:end]))
		title := strings.TrimSpace(body[end+1:])
		if err := board.ValidateTitle(title); err != nil {
			return nil, fmt.Errorf("invalid execution task: %w", err)
		}
		owner := ""
		if rawOwner != "" && rawOwner != "-" && rawOwner != "none" && rawOwner != "unassigned" {
			owner = sanitizeAgent(rawOwner)
			if owner != rawOwner {
				return nil, fmt.Errorf("invalid execution task owner %q", rawOwner)
			}
		}
		tasks = append(tasks, executionTask{Owner: owner, Title: title})
	}
	if !inSection {
		return nil, fmt.Errorf("final-plan.md is missing `## Execution tasks`")
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("final-plan.md has no execution tasks")
	}
	return tasks, nil
}

func markPlanFinalized(path string, tasks []createdExecutionTask) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	found := false
	for index, line := range lines {
		if key, _, ok := strings.Cut(strings.TrimSpace(line), ":"); ok && strings.EqualFold(strings.TrimSpace(key), "status") {
			lines[index] = "Status: finalized"
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("final plan is missing a Status field")
	}
	lines = append(lines, "", "## Created board tasks", "")
	for _, task := range tasks {
		owner := task.Owner
		if owner == "" {
			owner = "unassigned"
		}
		lines = append(lines, fmt.Sprintf("- %s [%s] %s", task.ID, owner, task.Title))
	}
	lines = append(lines, "", "Finalized-At: "+time.Now().UTC().Format(time.RFC3339), "")
	return platform.AtomicWrite(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func notifyPlanFinalized(p *project.Project, runID, from string, participants []string, tasks []createdExecutionTask) {
	for _, participant := range participants {
		var assigned []string
		for _, task := range tasks {
			if task.Owner == participant {
				assigned = append(assigned, fmt.Sprintf("- %s: %s", task.ID, task.Title))
			}
		}
		message := fmt.Sprintf("Planning run %s is finalized. Read `.coact/runs/%s/final-plan.md` and `coact board`.", runID, runID)
		if len(assigned) > 0 {
			message += "\n\nYour assigned tasks:\n" + strings.Join(assigned, "\n") + "\n\nClaim one with `coact claim <task-id>` before editing."
		}
		_ = inbox.Send(p.InboxDir(), from, participant, message)
	}
}

func setProposalStatus(path, status string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	found := false
	for index, line := range lines {
		key, _, ok := strings.Cut(strings.TrimSpace(line), ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "status") {
			lines[index] = "Status: " + status
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("proposal is missing a Status field")
	}
	return platform.AtomicWrite(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func planDistributor(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
		if key, value, ok := strings.Cut(line, ":"); ok && strings.EqualFold(strings.TrimSpace(key), "final_task_distributor") {
			raw := strings.ToLower(strings.TrimSpace(value))
			if raw == "" || sanitizeAgent(raw) != raw {
				return ""
			}
			return raw
		}
	}
	return ""
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
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		agent := sanitizeAgent(part)
		if agent != part || seen[agent] {
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
			raw := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "- final_task_distributor:")))
			if raw != "" && sanitizeAgent(raw) == raw {
				return raw
			}
			return "human"
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
