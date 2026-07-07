package cli

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/journal"
)

// gitC runs git in root and returns combined output.
func gitC(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// worktreePath is where an agent's isolated worktree lives: outside the repo (so
// its edits fall outside the main root and the shared-tree hook no-ops),
// grouped under a hidden sibling directory.
func worktreePath(root, agent string) string {
	return filepath.Join(filepath.Dir(root), ".coact-worktrees", filepath.Base(root), agent)
}

func branchName(agent string) string { return "coact/" + agent }

func branchExists(root, branch string) bool {
	_, err := gitC(root, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

func isWorktree(root, path string) bool {
	out, _ := gitC(root, "worktree", "list", "--porcelain")
	abs, _ := filepath.Abs(path)
	for _, line := range strings.Split(out, "\n") {
		if wt, ok := strings.CutPrefix(strings.TrimSpace(line), "worktree "); ok {
			if wtabs, _ := filepath.Abs(wt); wtabs == abs {
				return true
			}
		}
	}
	return false
}

// worktreeAdd creates (or reuses) the agent's worktree on branch coact/<agent>.
// Returns the path and whether it was newly created.
func worktreeAdd(root, agent string) (string, bool, error) {
	wtPath := worktreePath(root, agent)
	if isWorktree(root, wtPath) {
		return wtPath, false, nil
	}
	if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
		return "", false, err
	}
	branch := branchName(agent)
	var out string
	var err error
	if branchExists(root, branch) {
		out, err = gitC(root, "worktree", "add", wtPath, branch)
	} else {
		out, err = gitC(root, "worktree", "add", "-b", branch, wtPath)
	}
	if err != nil {
		return "", false, fmt.Errorf("git worktree add: %s", strings.TrimSpace(out))
	}
	return wtPath, true, nil
}

func cmdWorktree(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: coact worktree <add|list|rm> [agent]")
		return 2
	}
	switch args[0] {
	case "add":
		return worktreeAddCmd(args[1:])
	case "list", "ls":
		return worktreeListCmd(args[1:])
	case "rm", "remove":
		return worktreeRmCmd(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "coact: unknown worktree subcommand %q\n", args[0])
		return 2
	}
}

func worktreeAddCmd(args []string) int {
	fs := flag.NewFlagSet("worktree add", flag.ContinueOnError)
	pos, err := parseInterspersed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact worktree add <agent>")
		return 2
	}
	agent := sanitizeAgent(pos[0])
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	wtPath, created, err := worktreeAdd(p.Root, agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	if created {
		_ = journal.Append(p.JournalDir(), agent, "worktree.add", map[string]string{"branch": branchName(agent)})
		fmt.Printf("created worktree for %s on branch %s:\n  %s\n", agent, branchName(agent), wtPath)
	} else {
		fmt.Printf("worktree for %s already exists:\n  %s\n", agent, wtPath)
	}
	fmt.Printf("\nwork there with:\n  cd %s && COACT_AGENT=%s coact %s\n", wtPath, agent, agent)
	return 0
}

func worktreeListCmd(args []string) int {
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	out, _ := gitC(p.Root, "worktree", "list", "--porcelain")
	var path, branch string
	found := false
	flush := func() {
		if path != "" && strings.HasPrefix(branch, "refs/heads/coact/") {
			fmt.Printf("  %-10s %s\n", strings.TrimPrefix(branch, "refs/heads/coact/"), path)
			found = true
		}
		path, branch = "", ""
	}
	for _, line := range strings.Split(out, "\n") {
		if v, ok := strings.CutPrefix(line, "worktree "); ok {
			flush()
			path = v
		} else if v, ok := strings.CutPrefix(line, "branch "); ok {
			branch = v
		}
	}
	flush()
	if !found {
		fmt.Println("no coact worktrees (create one with `coact worktree add <agent>`)")
	}
	return 0
}

func worktreeRmCmd(args []string) int {
	fs := flag.NewFlagSet("worktree rm", flag.ContinueOnError)
	delBranch := fs.Bool("branch", false, "also delete the coact/<agent> branch")
	force := fs.Bool("force", false, "remove even with uncommitted changes")
	pos, err := parseInterspersed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact worktree rm [--branch] [--force] <agent>")
		return 2
	}
	agent := sanitizeAgent(pos[0])
	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	rmArgs := []string{"worktree", "remove", worktreePath(p.Root, agent)}
	if *force {
		rmArgs = append(rmArgs, "--force")
	}
	if out, err := gitC(p.Root, rmArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "coact: %s\n", strings.TrimSpace(out))
		return 1
	}
	fmt.Printf("removed worktree for %s\n", agent)
	if *delBranch {
		if out, err := gitC(p.Root, "branch", "-D", branchName(agent)); err != nil {
			fmt.Fprintf(os.Stderr, "coact: %s\n", strings.TrimSpace(out))
		} else {
			fmt.Printf("deleted branch %s\n", branchName(agent))
		}
	}
	_ = journal.Append(p.JournalDir(), agent, "worktree.remove", nil)
	return 0
}
