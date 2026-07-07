package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/journal"
)

// cmdMerge integrates one or more agents' coact/<agent> branches into the branch
// currently checked out in the main worktree. On conflict it stops and surfaces
// the conflicted files — the human is the integration gate; coact does not
// auto-resolve.
func cmdMerge(args []string) int {
	fs := flag.NewFlagSet("merge", flag.ContinueOnError)
	agents, err := parseInterspersed(fs, args)
	if err != nil || len(agents) == 0 {
		fmt.Fprintln(os.Stderr, "usage: coact merge <agent> [agent...]")
		return 2
	}
	p, _, ok := loadProject()
	if !ok {
		return 1
	}

	for _, raw := range agents {
		agent := sanitizeAgent(raw)
		branch := branchName(agent)
		if !branchExists(p.Root, branch) {
			fmt.Fprintf(os.Stderr, "coact: no branch %s (has %s worked yet?)\n", branch, agent)
			return 1
		}
		out, err := gitC(p.Root, "merge", "--no-edit", branch)
		if err != nil {
			conflicts, _ := gitC(p.Root, "diff", "--name-only", "--diff-filter=U")
			_ = journal.Append(p.JournalDir(), agent, "merge.conflict", map[string]string{"branch": branch})
			fmt.Fprintf(os.Stderr, "merge of %s hit conflicts:\n", branch)
			for _, f := range strings.Fields(strings.TrimSpace(conflicts)) {
				fmt.Fprintf(os.Stderr, "  %s\n", f)
			}
			fmt.Fprintln(os.Stderr, "\nresolve them, then `git add` + `git commit` — or `git merge --abort` to back out.")
			if strings.TrimSpace(conflicts) == "" {
				fmt.Fprintf(os.Stderr, "git: %s\n", strings.TrimSpace(out))
			}
			return 3
		}
		_ = journal.Append(p.JournalDir(), agent, "merge.ok", map[string]string{"branch": branch})
		fmt.Printf("merged %s\n", branch)
	}
	return 0
}
