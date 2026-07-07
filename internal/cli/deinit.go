package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tianyi-zhang-02/coact/internal/project"
)

// cmdDeinit removes coact's wiring — the Claude hook and the contract blocks —
// so a user can cleanly back out. Reversibility is a trust feature.
func cmdDeinit(args []string) int {
	fs := flag.NewFlagSet("deinit", flag.ContinueOnError)
	purge := fs.Bool("purge", false, "also delete the .coact directory (board, journal, config)")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}

	p, err := project.Find()
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}

	var removed []string
	if changed, err := removeClaudeHook(p.Root); err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
	} else if changed {
		removed = append(removed, ".claude/settings.json (PreToolUse hook)")
	}
	for _, f := range []string{"CLAUDE.md", "AGENTS.md"} {
		if changed, err := removeMarkedBlock(filepath.Join(p.Root, f)); err != nil {
			fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		} else if changed {
			removed = append(removed, f+" (coact contract)")
		}
	}
	if *purge {
		dir := p.CoactDir()
		if filepath.Base(dir) != ".coact" {
			// Belt-and-suspenders: never RemoveAll anything not named .coact.
			fmt.Fprintln(os.Stderr, "coact: refusing to purge a directory not named .coact")
		} else if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(os.Stderr, "coact: removing .coact: %v\n", err)
		} else {
			removed = append(removed, ".coact/ (purged)")
		}
	}

	if len(removed) == 0 {
		fmt.Println("nothing to remove — coact was not wired here")
		return 0
	}
	fmt.Println("removed:")
	for _, r := range removed {
		fmt.Printf("  %s\n", r)
	}
	if !*purge {
		fmt.Println("\n.coact/ (board, journal, config) was kept — use `coact deinit --purge` to remove it too.")
	}
	return 0
}
