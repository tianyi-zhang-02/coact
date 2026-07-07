package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/adapter"
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

	// Enforcement hook + agent contracts, driven by the adapter registry (all
	// idempotent). Each configured agent gets its contract file; hook-enforced
	// agents (Claude) also get the PreToolUse hook.
	cfg, _ := config.Load(p.ConfigPath())
	for _, ac := range cfg.Agents {
		ad, ok := adapter.Get(ac.ID)
		if !ok {
			continue
		}
		if ad.HardHook {
			if added, err := ensureClaudeHook(root); err != nil {
				fmt.Fprintf(os.Stderr, "coact: wiring %s hook: %v\n", ad.ID, err)
			} else if added {
				note(".claude/settings.json (" + ad.ID + " PreToolUse hook)")
			}
		}
		if added, err := ensureMarkedBlock(filepath.Join(root, ad.ContractFile), ad.Contract()); err != nil {
			fmt.Fprintf(os.Stderr, "coact: writing %s: %v\n", ad.ContractFile, err)
		} else if added {
			note(ad.ContractFile + " (" + ad.ID + " contract)")
		}
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

then launch each agent in its own terminal — one command each:
  coact claude          # or:  coact codex   /   coact gemini

coact adds a gate — it does not disable your permissions, and the hook fails
open, so if coact ever errors your editing still works. Run "coact adapters" to
see the agents it can coordinate, and "coact deinit" to undo everything.
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
