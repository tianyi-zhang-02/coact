// Package setup initializes a repository for coact. It is used by both the CLI
// and the local UI so initialization never shells out to the coact binary.
package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/adapter"
	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

const (
	coactBegin = "<!-- coact:begin -->"
	coactEnd   = "<!-- coact:end -->"
)

// Result describes what initialization changed.
type Result struct {
	Root  string   `json:"root"`
	Wired []string `json:"wired"`
}

// EnsureMarkedBlock appends a coact-marked contract block if absent.
func EnsureMarkedBlock(path, body string) (bool, error) {
	return ensureMarkedBlock(path, body)
}

// EnsureClaudeHook wires the Claude Code PreToolUse hook if absent.
func EnsureClaudeHook(root string) (bool, error) {
	return ensureClaudeHook(root)
}

// Initialize creates .coact state, wires contracts/hooks, and appends runtime
// ignore entries. It is idempotent.
func Initialize(root, agent string) (*Result, error) {
	p := &project.Project{Root: root, CheckoutRoot: root}
	for _, d := range []string{
		p.CoactDir(), p.LocksDir(), p.SessionDir(), p.InboxDir(), p.JournalDir(),
		p.MemoryDir(), p.RunsDir(), p.TasksDir(), p.UsageDir(), p.EvaluationDir(),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("creating %s: %w", d, err)
		}
	}

	res := &Result{Root: root}
	note := func(what string) { res.Wired = append(res.Wired, what) }

	if !exists(p.ConfigPath()) {
		if err := config.Default().Save(p.ConfigPath()); err != nil {
			return nil, fmt.Errorf("writing config: %w", err)
		}
		note(rel(root, p.ConfigPath()))
	}
	if !exists(p.BoardPath()) {
		if err := os.WriteFile(p.BoardPath(), []byte(boardTemplate), 0o644); err != nil {
			return nil, fmt.Errorf("writing board: %w", err)
		}
		note(rel(root, p.BoardPath()))
	}
	if !exists(p.TeamPath()) {
		if err := os.WriteFile(p.TeamPath(), []byte(teamTemplate), 0o644); err != nil {
			return nil, fmt.Errorf("writing team: %w", err)
		}
		note(rel(root, p.TeamPath()))
	}
	memoryPath := filepath.Join(p.MemoryDir(), "project.md")
	if !exists(memoryPath) {
		if err := os.WriteFile(memoryPath, []byte(memoryTemplate), 0o644); err != nil {
			return nil, fmt.Errorf("writing project memory: %w", err)
		}
		note(rel(root, memoryPath))
	}

	cfg, _ := config.Load(p.ConfigPath())
	policyMigrated := migrateProtectedPaths(cfg)
	agentsMigrated := migrateRetiredGeminiAgent(cfg)
	if policyMigrated || agentsMigrated {
		if err := cfg.Save(p.ConfigPath()); err != nil {
			return nil, fmt.Errorf("updating config: %w", err)
		}
		migration := "policy migration"
		if agentsMigrated {
			migration = "Antigravity agent migration"
		}
		if policyMigrated && agentsMigrated {
			migration = "policy + Antigravity agent migration"
		}
		note(rel(root, p.ConfigPath()) + " (" + migration + ")")
	}
	for _, ac := range cfg.Agents {
		ad, ok := adapter.Get(ac.ID)
		if !ok {
			continue
		}
		if ad.HardHook {
			added, err := ensureClaudeHook(root)
			if err != nil {
				return nil, fmt.Errorf("wiring %s hook: %w", ad.ID, err)
			}
			if added {
				note(".claude/settings.json (" + ad.ID + " PreToolUse hook)")
			}
		}
		added, err := ensureMarkedBlock(filepath.Join(root, ad.ContractFile), ad.Contract())
		if err != nil {
			return nil, fmt.Errorf("writing %s: %w", ad.ContractFile, err)
		}
		if added {
			note(ad.ContractFile + " (" + ad.ID + " contract)")
		}
	}
	legacyContract := filepath.Join(root, "GEMINI.md")
	removed, err := removeMarkedBlockFile(legacyContract)
	if err != nil {
		return nil, fmt.Errorf("removing retired agent contract: %w", err)
	}
	if removed {
		note("GEMINI.md (retired contract removed)")
	}

	ensureGitignore(root)
	_ = journal.Append(p.JournalDir(), agent, "session.start", map[string]string{"action": "init"})
	return res, nil
}

// migrateRetiredGeminiAgent replaces the retired third-party adapter while
// preserving any write restrictions the workspace assigned to that role.
func migrateRetiredGeminiAgent(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	hasAntigravity := false
	for _, agent := range cfg.Agents {
		if agent.ID == "antigravity" {
			hasAntigravity = true
			break
		}
	}
	changed := false
	updated := make([]config.AgentConfig, 0, len(cfg.Agents))
	for _, agent := range cfg.Agents {
		if agent.ID != "gemini" {
			updated = append(updated, agent)
			continue
		}
		changed = true
		if hasAntigravity {
			continue
		}
		agent.ID = "antigravity"
		agent.Adapter = "antigravity-cli"
		updated = append(updated, agent)
		hasAntigravity = true
	}
	if changed {
		cfg.Agents = updated
	}
	return changed
}

func migrateProtectedPaths(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	changed := false
	if len(cfg.Policy.ProtectedPaths) == 1 && cfg.Policy.ProtectedPaths[0] == ".coact/**" {
		cfg.Policy.ProtectedPaths = append([]string(nil), config.Default().Policy.ProtectedPaths...)
		changed = true
	}
	for _, required := range config.Default().Policy.ProtectedPaths {
		if !contains(cfg.Policy.ProtectedPaths, required) {
			cfg.Policy.ProtectedPaths = append(cfg.Policy.ProtectedPaths, required)
			changed = true
		}
	}
	if cfg.Version != config.Default().Version {
		cfg.Version = config.Default().Version
		changed = true
	}
	return changed
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

func ensureMarkedBlock(path, body string) (bool, error) {
	var content string
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	} else if !os.IsNotExist(err) {
		return false, err
	}
	block := coactBegin + "\n" + strings.TrimRight(body, "\n") + "\n" + coactEnd
	if start := strings.Index(content, coactBegin); start >= 0 {
		endRel := strings.Index(content[start:], coactEnd)
		if endRel < 0 {
			return false, fmt.Errorf("coact contract block in %s is missing %s", path, coactEnd)
		}
		end := start + endRel + len(coactEnd)
		updated := content[:start] + block + content[end:]
		if updated == content {
			return false, nil
		}
		return true, platform.AtomicWrite(path, []byte(updated), 0o644)
	}
	var b strings.Builder
	b.WriteString(content)
	if content != "" {
		if !strings.HasSuffix(content, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(block + "\n")
	return true, platform.AtomicWrite(path, []byte(b.String()), 0o644)
}

func removeMarkedBlockFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	content := string(data)
	start := strings.Index(content, coactBegin)
	if start < 0 {
		return false, nil
	}
	endRel := strings.Index(content[start:], coactEnd)
	if endRel < 0 {
		return false, fmt.Errorf("coact contract block in %s is missing %s", path, coactEnd)
	}
	end := start + endRel + len(coactEnd)
	before := strings.TrimSpace(content[:start])
	after := strings.TrimSpace(content[end:])
	if before == "" && after == "" {
		return true, os.Remove(path)
	}
	updated := before
	if updated != "" && after != "" {
		updated += "\n\n"
	}
	updated += after
	if updated != "" {
		updated += "\n"
	}
	return true, platform.AtomicWrite(path, []byte(updated), 0o644)
}

func ensureClaudeHook(root string) (bool, error) {
	claudeDir := filepath.Join(root, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return false, err
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")

	settings := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return false, fmt.Errorf(".claude/settings.json is not valid JSON: %w", err)
		}
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	pre, _ := hooks["PreToolUse"].([]any)
	desired := coactBinary() + " hook claude"

	present := false
	dirty := false
	var kept []any
	for _, entry := range pre {
		em, _ := entry.(map[string]any)
		hs, _ := em["hooks"].([]any)
		var keptHooks []any
		for _, h := range hs {
			hm, ok := h.(map[string]any)
			if ok && isCoactHook(hm) {
				c, _ := hm["command"].(string)
				_, hasArgs := hm["args"]
				if c == desired && !hasArgs && !present {
					present = true
					keptHooks = append(keptHooks, h)
				} else {
					dirty = true
				}
				continue
			}
			keptHooks = append(keptHooks, h)
		}
		if len(keptHooks) > 0 {
			em["hooks"] = keptHooks
			kept = append(kept, em)
		} else {
			dirty = true
		}
	}

	if present && !dirty {
		return false, nil
	}
	if !present {
		kept = append(kept, map[string]any{
			"matcher": "Edit|Write|MultiEdit|NotebookEdit",
			"hooks":   []any{map[string]any{"type": "command", "command": desired}},
		})
	}
	hooks["PreToolUse"] = kept
	settings["hooks"] = hooks

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return false, err
	}
	data = append(data, '\n')
	if err := platform.AtomicWrite(settingsPath, data, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func isCoactHook(hm map[string]any) bool {
	if c, _ := hm["command"].(string); strings.Contains(c, "hook claude") {
		return true
	}
	if args, ok := hm["args"].([]any); ok {
		var hasHook, hasClaude bool
		for _, a := range args {
			switch s, _ := a.(string); s {
			case "hook":
				hasHook = true
			case "claude":
				hasClaude = true
			}
		}
		if hasHook && hasClaude {
			return true
		}
	}
	return false
}

func coactBinary() string {
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(exe); err == nil {
			return abs
		}
		return exe
	}
	return "coact"
}

func ensureGitignore(root string) {
	path := filepath.Join(root, ".gitignore")
	needed := []string{
		".coact/locks/", ".coact/session/", ".coact/journal/", ".coact/inbox/", ".coact/terminal/", ".coact/runs/", ".coact/tasks/", ".coact/memory/", ".coact/usage/", ".coact/evaluations/",
	}
	var content string
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}
	var add []string
	for _, n := range needed {
		if !gitignoreCovers(content, n) {
			add = append(add, n)
		}
	}
	if len(add) == 0 {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	if content != "" && !strings.HasSuffix(content, "\n") {
		f.WriteString("\n")
	}
	f.WriteString("\n# coact runtime state\n")
	for _, a := range add {
		f.WriteString(a + "\n")
	}
}

func gitignoreCovers(content, pattern string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == pattern || line == "/"+pattern {
			return true
		}
		if line == ".coact/" || line == "/.coact/" {
			return strings.HasPrefix(pattern, ".coact/")
		}
	}
	return false
}

const boardTemplate = `# Task board

Tasks below carry machine-readable metadata in an HTML comment. The checkbox
mirrors state for humans: [ ] todo, [~] doing, [x] done, [!] blocked.

## Backlog

- [ ] Example: describe a task here <!-- coact: id=T-001 state=todo owner= -->

## Done
`

const teamTemplate = `# CoAct team policy

This file defines how coding agents coordinate in this repository. Agents should
read it at the start of every session, before planning, and before claiming work.

## Roles

- final_task_distributor: human
- planning_agents: codex, claude

## Agent preferences

- codex: implementation, tests, refactors, validation, security review.
- claude: product copy, UX review, docs, planning critique, second-pass review.
- antigravity: optional research, alternate implementation ideas, broad review.

## Protocol

1. Read .coact/team.md and .coact/memory/project.md.
2. Read your inbox with coact inbox.
3. During planning, write your proposal under .coact/runs/<run>/proposals/<agent>.md.
4. Read peer proposals before finalizing execution tasks.
5. The lead writes .coact/runs/<run>/final-plan.md using a short description and full Prompt for each task.
6. By default, the lead submits the plan and waits for explicit human approval before distributing tasks.
7. Claim tasks with coact claim <id> before editing.
8. Lock files or directories before editing unless your adapter has a hard CoAct hook.

`

const memoryTemplate = `# CoAct project memory

Shared local context for agents working in this checkout.

- Keep this file concise.
- Do not store secrets, API keys, credentials, or private user data.
- Prefer durable project facts over transient task notes.

`
