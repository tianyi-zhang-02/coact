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
  1. Wire each agent's contract (once):
       Claude Code -> include .coact/adapters/claude.md in CLAUDE.md
       Codex       -> include .coact/adapters/codex.md in AGENTS.md
  2. Set a per-session id: export COACT_AGENT=claude   (or codex)
  3. Coordinate:
       coact lock <path>     # before editing
       coact status          # see who holds what
       coact unlock <path>   # when done
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
