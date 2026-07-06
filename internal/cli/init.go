package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/journal"
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
		p.CoactDir(), p.LocksDir(), p.SessionDir(), p.InboxDir(), p.JournalDir(),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "coact: creating %s: %v\n", d, err)
			return 1
		}
	}

	var wired []string
	note := func(what string) { wired = append(wired, what) }

	if !exists(p.ConfigPath()) {
		if err := config.Default().Save(p.ConfigPath()); err != nil {
			fmt.Fprintf(os.Stderr, "coact: writing config: %v\n", err)
			return 1
		}
		note(rel(root, p.ConfigPath()))
	}
	if !exists(p.BoardPath()) {
		if err := os.WriteFile(p.BoardPath(), []byte(boardTemplate), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "coact: writing board: %v\n", err)
			return 1
		}
		note(rel(root, p.BoardPath()))
	}

	// Enforcement hook + agent contracts (all idempotent).
	if added, err := ensureClaudeHook(root); err != nil {
		fmt.Fprintf(os.Stderr, "coact: wiring Claude hook: %v\n", err)
	} else if added {
		note(".claude/settings.json (Claude PreToolUse hook)")
	}
	if added, err := ensureMarkedBlock(filepath.Join(root, "CLAUDE.md"), claudeFragment); err != nil {
		fmt.Fprintf(os.Stderr, "coact: writing CLAUDE.md: %v\n", err)
	} else if added {
		note("CLAUDE.md (coact contract)")
	}
	if added, err := ensureMarkedBlock(filepath.Join(root, "AGENTS.md"), codexFragment); err != nil {
		fmt.Fprintf(os.Stderr, "coact: writing AGENTS.md: %v\n", err)
	} else if added {
		note("AGENTS.md (coact contract)")
	}

	ensureGitignore(root)
	_ = journal.Append(p.JournalDir(), agentID(""), "session.start", map[string]string{"action": "init"})

	fmt.Printf("coact initialized at %s\n", root)
	if len(wired) > 0 {
		fmt.Println("wired:")
		for _, c := range wired {
			fmt.Printf("  %s\n", c)
		}
	} else {
		fmt.Println("(already initialized — nothing to change)")
	}
	fmt.Print(`
verify it works (no second agent needed):
  coact doctor           checks the wiring and runs an enforcement self-test

then run each agent in its own terminal:
  export COACT_AGENT=claude    # or codex
  coact sidecar &              # keeps this session's presence live
  claude                       # or codex

coact adds a gate — it does not disable your permissions, and the hook fails
open, so if coact ever errors your editing still works. Undo everything with:
  coact deinit
`)
	return 0
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
