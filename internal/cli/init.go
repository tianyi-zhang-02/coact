package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/platform"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func cmdInit(args []string) int {
	root, err := project.Locate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	p := &project.Project{Root: root}

	for _, d := range []string{
		p.CoactDir(), p.LocksDir(), p.SessionDir(), p.InboxDir(),
		p.JournalDir(), p.AdaptersDir(),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "coact: creating %s: %v\n", d, err)
			return 1
		}
	}

	created := []string{}

	if !exists(p.ConfigPath()) {
		if err := config.Default().Save(p.ConfigPath()); err != nil {
			fmt.Fprintf(os.Stderr, "coact: writing config: %v\n", err)
			return 1
		}
		created = append(created, rel(root, p.ConfigPath()))
	}

	if !exists(p.BoardPath()) {
		if err := os.WriteFile(p.BoardPath(), []byte(boardTemplate), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "coact: writing board: %v\n", err)
			return 1
		}
		created = append(created, rel(root, p.BoardPath()))
	}

	adapters := map[string]string{
		filepath.Join(p.AdaptersDir(), "claude.md"): claudeFragment,
		filepath.Join(p.AdaptersDir(), "codex.md"):  codexFragment,
	}
	for path, body := range adapters {
		if !exists(path) {
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "coact: writing %s: %v\n", path, err)
				return 1
			}
			created = append(created, rel(root, path))
		}
	}

	if wired, err := ensureClaudeHook(root); err != nil {
		fmt.Fprintf(os.Stderr, "coact: could not wire Claude hook: %v\n", err)
	} else if wired {
		created = append(created, ".claude/settings.json (PreToolUse hook)")
	}

	ensureGitignore(root)
	_ = journal.Append(p.JournalDir(), agentID(""), "session.start", map[string]string{"action": "init"})

	fmt.Printf("coact initialized at %s\n", root)
	if len(created) > 0 {
		fmt.Println("created:")
		for _, c := range created {
			fmt.Printf("  %s\n", c)
		}
	} else {
		fmt.Println("(already initialized — nothing to create)")
	}
	fmt.Print(`
next steps:
  1. Claude Code's enforcement hook is wired in .claude/settings.json.
     For Codex, include .coact/adapters/codex.md in your AGENTS.md.
  2. Set a per-session id before launching each agent:
       export COACT_AGENT=claude      (or codex)
  3. Optionally keep the session live for accurate presence:
       coact sidecar &
  4. Coordinate and observe:
       coact board / coact claim <id>   # divide the work
       coact status                     # who holds what
       coact log                        # audit trail
`)
	return 0
}

// ensureClaudeHook idempotently wires coact's PreToolUse gate into the repo's
// .claude/settings.json, merging with any existing settings. Returns true if it
// added the hook (false if already present).
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

	for _, entry := range pre {
		em, _ := entry.(map[string]any)
		hs, _ := em["hooks"].([]any)
		for _, h := range hs {
			if hm, ok := h.(map[string]any); ok && isCoactHook(hm) {
				return false, nil // already wired
			}
		}
	}

	entry := map[string]any{
		"matcher": "Edit|Write|MultiEdit|NotebookEdit",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": coactBinary(),
				"args":    []any{"hook", "claude"},
			},
		},
	}
	hooks["PreToolUse"] = append(pre, entry)
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
	// Match the command string form (e.g. "coact hook claude")...
	if c, _ := hm["command"].(string); strings.Contains(c, "coact") && strings.Contains(c, "hook") {
		return true
	}
	// ...or the exec-form args signature, independent of the binary path.
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

// coactBinary returns an absolute path to the running coact binary so the wired
// hook works whether or not coact is on PATH. Falls back to "coact".
func coactBinary() string {
	if exe, err := os.Executable(); err == nil {
		if abs, err := filepath.Abs(exe); err == nil {
			return abs
		}
		return exe
	}
	return "coact"
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

func ensureGitignore(root string) {
	path := filepath.Join(root, ".gitignore")
	needed := []string{
		".coact/locks/", ".coact/session/", ".coact/journal/", ".coact/inbox/",
	}
	var content string
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}
	var add []string
	for _, n := range needed {
		if !strings.Contains(content, n) {
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
