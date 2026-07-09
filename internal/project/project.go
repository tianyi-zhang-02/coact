// Package project locates a coact workspace and resolves the paths under
// .coact/ that hold coordination state.
package project

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNotInitialized is returned when no .coact directory is found.
var ErrNotInitialized = errors.New("coact: not initialized in this directory tree (run `coact init` in your repo root)")

// Project is a workspace rooted at the directory containing .coact.
type Project struct {
	Root         string
	CheckoutRoot string
}

func (p *Project) CoactDir() string    { return filepath.Join(p.Root, ".coact") }
func (p *Project) LocksDir() string    { return filepath.Join(p.CoactDir(), "locks") }
func (p *Project) SessionDir() string  { return filepath.Join(p.CoactDir(), "session") }
func (p *Project) InboxDir() string    { return filepath.Join(p.CoactDir(), "inbox") }
func (p *Project) JournalDir() string  { return filepath.Join(p.CoactDir(), "journal") }
func (p *Project) AdaptersDir() string { return filepath.Join(p.CoactDir(), "adapters") }
func (p *Project) MemoryDir() string   { return filepath.Join(p.CoactDir(), "memory") }
func (p *Project) RunsDir() string     { return filepath.Join(p.CoactDir(), "runs") }
func (p *Project) ConfigPath() string  { return filepath.Join(p.CoactDir(), "config.json") }
func (p *Project) BoardPath() string   { return filepath.Join(p.CoactDir(), "board.md") }
func (p *Project) BriefPath() string   { return filepath.Join(p.CoactDir(), "brief.md") }
func (p *Project) TeamPath() string    { return filepath.Join(p.CoactDir(), "team.md") }

// WorkRoot is the checkout where user paths should be interpreted. For the main
// worktree this is Root; for linked worktrees it is the linked checkout while
// Root remains the main worktree that owns .coact/.
func (p *Project) WorkRoot() string {
	if p.CheckoutRoot != "" {
		return p.CheckoutRoot
	}
	return p.Root
}

// Find locates the project by walking up from the current directory looking for
// an existing .coact directory.
func Find() (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return FindFrom(cwd)
}

// FindFrom locates the project by walking up from start.
func FindFrom(start string) (*Project, error) {
	dir := start
	for {
		info, err := os.Stat(filepath.Join(dir, ".coact"))
		if err == nil && info.IsDir() {
			return &Project{Root: dir, CheckoutRoot: dir}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNotInitialized
		}
		dir = parent
	}
}

// Resolve is like Find but worktree-aware: when run inside a linked git
// worktree, coact's shared state lives in the MAIN worktree's .coact, so the
// board, journal, and locks stay global across an agent's isolated worktrees.
func Resolve() (*Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ResolveFrom(cwd)
}

// ResolveFrom resolves the project from start, mapping a linked worktree to its
// main worktree's .coact and otherwise walking up for .coact.
func ResolveFrom(start string) (*Project, error) {
	if main := mainWorktreeRoot(start); main != "" {
		if st, err := os.Stat(filepath.Join(main, ".coact")); err == nil && st.IsDir() {
			checkout := gitTopLevel(start)
			if checkout == "" {
				checkout = start
			}
			return &Project{Root: main, CheckoutRoot: checkout}, nil
		}
	}
	return FindFrom(start)
}

// mainWorktreeRoot returns the main worktree's root if start is inside a linked
// git worktree, else "". A linked worktree's git dir looks like
// <main>/.git/worktrees/<name>.
func mainWorktreeRoot(start string) string {
	out, err := exec.Command("git", "-C", start, "rev-parse", "--absolute-git-dir").Output()
	if err != nil {
		return ""
	}
	gitDir := strings.TrimSpace(string(out))
	parent := filepath.Dir(gitDir) // .../.git/worktrees
	if filepath.Base(parent) == "worktrees" {
		return filepath.Dir(filepath.Dir(parent)) // .../.git -> <main>
	}
	return ""
}

// Locate returns the directory where a new project should be initialized: the
// git top-level if the current directory is inside a git repo, otherwise the
// current directory.
func Locate() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if root := gitTopLevel(cwd); root != "" {
		return root, nil
	}
	return cwd, nil
}

func gitTopLevel(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
