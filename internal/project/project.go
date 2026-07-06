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
	Root string
}

func (p *Project) CoactDir() string    { return filepath.Join(p.Root, ".coact") }
func (p *Project) LocksDir() string    { return filepath.Join(p.CoactDir(), "locks") }
func (p *Project) SessionDir() string  { return filepath.Join(p.CoactDir(), "session") }
func (p *Project) InboxDir() string    { return filepath.Join(p.CoactDir(), "inbox") }
func (p *Project) JournalDir() string  { return filepath.Join(p.CoactDir(), "journal") }
func (p *Project) AdaptersDir() string { return filepath.Join(p.CoactDir(), "adapters") }
func (p *Project) ConfigPath() string  { return filepath.Join(p.CoactDir(), "config.json") }
func (p *Project) BoardPath() string   { return filepath.Join(p.CoactDir(), "board.md") }

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
			return &Project{Root: dir}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNotInitialized
		}
		dir = parent
	}
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
