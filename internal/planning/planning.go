// Package planning creates and inspects collaborative planning runs shared by
// the CLI and local Dashboard.
package planning

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/inbox"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

var identifierPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type StartOptions struct {
	RunID        string
	Brief        string
	Initiator    string
	Lead         string
	ApprovalMode string
	Participants []string
}

type Info struct {
	ID           string   `json:"id"`
	Brief        string   `json:"brief"`
	Lead         string   `json:"lead"`
	ApprovalMode string   `json:"approval_mode"`
	Status       string   `json:"status"`
	Participants []string `json:"participants"`
}

func NormalizeApprovalMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "review", "manual", "human":
		return "review"
	case "auto", "auto-distribute", "autodistribute":
		return "auto"
	default:
		return ""
	}
}

func Start(p *project.Project, opts StartOptions) (*Info, error) {
	opts.RunID = strings.TrimSpace(opts.RunID)
	opts.Brief = strings.TrimSpace(opts.Brief)
	opts.Initiator = strings.ToLower(strings.TrimSpace(opts.Initiator))
	opts.Lead = strings.ToLower(strings.TrimSpace(opts.Lead))
	opts.ApprovalMode = NormalizeApprovalMode(opts.ApprovalMode)
	if !identifierPattern.MatchString(opts.RunID) {
		return nil, fmt.Errorf("invalid planning run id %q", opts.RunID)
	}
	if opts.Brief == "" {
		return nil, fmt.Errorf("planning brief is required")
	}
	if opts.ApprovalMode == "" {
		return nil, fmt.Errorf("approval mode must be review or auto")
	}
	if !validActor(opts.Initiator) || !validActor(opts.Lead) {
		return nil, fmt.Errorf("invalid initiator or lead")
	}
	participants := uniqueParticipants(opts.Participants)
	if len(participants) == 0 {
		return nil, fmt.Errorf("at least one planning participant is required")
	}
	if opts.Lead != "human" && opts.Lead != "vote" && !contains(participants, opts.Lead) {
		return nil, fmt.Errorf("lead %q must be a planning participant", opts.Lead)
	}
	if opts.ApprovalMode == "auto" && (opts.Lead == "human" || opts.Lead == "vote") {
		return nil, fmt.Errorf("auto approval requires an agent lead")
	}

	runDir := filepath.Join(p.RunsDir(), opts.RunID)
	if _, err := os.Stat(runDir); err == nil {
		return nil, fmt.Errorf("planning run %s already exists", opts.RunID)
	}
	if err := os.MkdirAll(filepath.Join(runDir, "proposals"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(runDir, "notes"), 0o755); err != nil {
		return nil, err
	}
	if err := writeFiles(runDir, opts, participants); err != nil {
		return nil, err
	}
	for _, agent := range participants {
		if err := inbox.Send(p.InboxDir(), opts.Initiator, agent, message(opts.RunID, opts.Lead, opts.ApprovalMode, agent)); err != nil {
			return nil, err
		}
	}
	_ = journal.Append(p.JournalDir(), opts.Initiator, "plan.start", map[string]string{
		"id": opts.RunID, "with": strings.Join(participants, ","), "distributor": opts.Lead, "approval": opts.ApprovalMode,
	})
	return &Info{ID: opts.RunID, Brief: opts.Brief, Lead: opts.Lead, ApprovalMode: opts.ApprovalMode, Status: "pending", Participants: participants}, nil
}

func Latest(p *project.Project) *Info {
	entries, err := os.ReadDir(p.RunsDir())
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && identifierPattern.MatchString(entry.Name()) {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	return Read(p, names[len(names)-1])
}

func Read(p *project.Project, runID string) *Info {
	if !identifierPattern.MatchString(runID) {
		return nil
	}
	runDir := filepath.Join(p.RunsDir(), runID)
	briefData, err := os.ReadFile(filepath.Join(runDir, "brief.md"))
	if err != nil {
		return nil
	}
	finalData, _ := os.ReadFile(filepath.Join(runDir, "final-plan.md"))
	info := &Info{ID: runID, ApprovalMode: "review", Status: field(string(finalData), "Status")}
	content := string(briefData)
	info.Brief = section(content, "Brief")
	info.Lead = field(content, "final_task_distributor")
	if mode := NormalizeApprovalMode(field(content, "approval_mode")); mode != "" {
		info.ApprovalMode = mode
	}
	info.Participants = splitParticipants(field(content, "participants"))
	return info
}

// Approve opens the human safety gate for a lead-submitted review plan.
func Approve(p *project.Project, runID string) (*Info, error) {
	info := Read(p, runID)
	if info == nil {
		return nil, fmt.Errorf("planning run %q was not found", runID)
	}
	if info.ApprovalMode != "review" {
		return nil, fmt.Errorf("planning run %s does not use human review", runID)
	}
	humanLedPending := (info.Lead == "human" || info.Lead == "vote") && info.Status == "pending"
	if info.Status != "review" && !humanLedPending {
		return nil, fmt.Errorf("final plan status is %s, not review", info.Status)
	}
	finalPath := filepath.Join(p.RunsDir(), runID, "final-plan.md")
	if err := setStatus(finalPath, "approved"); err != nil {
		return nil, err
	}
	info.Status = "approved"
	_ = journal.Append(p.JournalDir(), "human", "plan.approve", map[string]string{"id": runID, "distributor": info.Lead})
	_ = inbox.Send(p.InboxDir(), "human", info.Lead, fmt.Sprintf("Human approved planning run %s. Finalize and distribute with `coact plan finalize %s`.", runID, runID))
	return info, nil
}

func writeFiles(runDir string, opts StartOptions, participants []string) error {
	finalStep := fmt.Sprintf("The lead submits with coact plan submit %s. A human reviews and runs coact plan approve %s. The lead then runs coact plan finalize %s.", opts.RunID, opts.RunID, opts.RunID)
	if opts.Lead == "human" || opts.Lead == "vote" {
		finalStep = fmt.Sprintf("The human reviews final-plan.md, runs coact plan approve %s, then runs coact plan finalize %s.", opts.RunID, opts.RunID)
	}
	if opts.ApprovalMode == "auto" {
		finalStep = fmt.Sprintf("DANGEROUS AUTO MODE: the lead may run coact plan finalize %s without human approval.", opts.RunID)
	}
	brief := fmt.Sprintf(`# CoAct planning run %s

## Brief

%s

## Protocol

1. Read .coact/team.md, .coact/memory/project.md, and this file.
2. Write one proposal to .coact/runs/%s/proposals/<agent>.md.
3. Run coact plan ready %s when the proposal is complete.
4. Read peer proposals and add review notes when useful.
5. The lead waits until all required proposals are ready and unlocked.
6. The lead writes final-plan.md using a short task description plus a full Prompt line.
7. %s
8. Assigned agents claim tasks before editing.

## Metadata

- initiator: %s
- participants: %s
- final_task_distributor: %s
- approval_mode: %s
- created_at: %s
`, opts.RunID, opts.Brief, opts.RunID, opts.RunID, finalStep, opts.Initiator, strings.Join(participants, ", "), opts.Lead, opts.ApprovalMode, time.Now().UTC().Format(time.RFC3339))
	if err := platform.AtomicWrite(filepath.Join(runDir, "brief.md"), []byte(brief), 0o644); err != nil {
		return err
	}
	finalPlan := fmt.Sprintf(`# Final plan for %s

Status: pending
Distributor: %s

Wait for every required proposal to be ready and unlocked.

## Decision

## Execution tasks

<!-- - [codex] Short Dashboard description -->
<!--   Prompt: Full instructions sent to the assigned agent. -->
<!-- Use [unassigned] when an agent should claim the task later. -->

Next: %s
`, opts.RunID, opts.Lead, finalStep)
	if err := platform.AtomicWrite(filepath.Join(runDir, "final-plan.md"), []byte(finalPlan), 0o644); err != nil {
		return err
	}
	for _, agent := range participants {
		template := fmt.Sprintf(`# Proposal: %s

Run: %s
Status: draft

When complete, run: coact plan ready %s

## Proposed approach

## Risks

## Suggested tasks
`, agent, opts.RunID, opts.RunID)
		if err := platform.AtomicWrite(filepath.Join(runDir, "proposals", agent+".md"), []byte(template), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func message(runID, lead, approval, recipient string) string {
	role := "planning participant"
	if recipient == lead {
		role = "lead planner and final task distributor"
	}
	next := fmt.Sprintf("Submit for human review with `coact plan submit %s`; after approval, finalize with `coact plan finalize %s`.", runID, runID)
	if lead == "human" || lead == "vote" {
		next = fmt.Sprintf("The human operator will review, approve, and finalize run %s after every proposal is ready.", runID)
	}
	if approval == "auto" {
		next = fmt.Sprintf("AUTO DISTRIBUTION IS ENABLED. After all proposals are ready, you may finalize with `coact plan finalize %s` without human approval.", runID)
	}
	return fmt.Sprintf("Planning phase %s started.\n\nYour role: %s.\n\nRead `.coact/runs/%s/brief.md`, write `.coact/runs/%s/proposals/%s.md`, then run `coact plan ready %s`. Read peer proposals before the lead writes final-plan.md. Check readiness with `coact plan status %s`. Each final task needs a short Dashboard description and a full Prompt.\n\n%s", runID, role, runID, runID, recipient, runID, runID, next)
}

func uniqueParticipants(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if !identifierPattern.MatchString(value) || seen[value] || value == "human" || value == "vote" {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func validActor(value string) bool {
	return value == "human" || value == "vote" || identifierPattern.MatchString(value)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func field(content, name string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
		if key, value, ok := strings.Cut(line, ":"); ok && strings.EqualFold(strings.TrimSpace(key), name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func section(content, name string) string {
	inSection := false
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			if inSection {
				break
			}
			inSection = strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(line, "## ")), name)
			continue
		}
		if inSection {
			lines = append(lines, line)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func splitParticipants(raw string) []string {
	return uniqueParticipants(strings.Split(raw, ","))
}

func setStatus(path, status string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	for index, line := range lines {
		if key, _, ok := strings.Cut(strings.TrimSpace(line), ":"); ok && strings.EqualFold(strings.TrimSpace(key), "status") {
			lines[index] = "Status: " + status
			return platform.AtomicWrite(path, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}
	return fmt.Errorf("final plan is missing a Status field")
}
