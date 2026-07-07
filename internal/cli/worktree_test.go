package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".coact"), 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "coact-test")
	if err := config.Default().Save(filepath.Join(repo, ".coact", "config.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-qm", "init")
	return repo
}

func TestWorktreeAddAndSharedResolution(t *testing.T) {
	repo := initGitRepo(t)

	wt, created, err := worktreeAdd(repo, "claude")
	if err != nil || !created {
		t.Fatalf("worktreeAdd: created=%v err=%v", created, err)
	}
	if _, err := os.Stat(wt); err != nil {
		t.Fatalf("worktree path missing: %v", err)
	}

	// Idempotent: a second add reuses the existing worktree.
	if _, created2, _ := worktreeAdd(repo, "claude"); created2 {
		t.Fatal("second worktreeAdd should reuse (created=false)")
	}

	// From inside the worktree, coact state resolves to the MAIN repo.
	p, err := project.ResolveFrom(wt)
	if err != nil {
		t.Fatalf("ResolveFrom: %v", err)
	}
	got, _ := filepath.EvalSymlinks(p.Root) // macOS /var vs /private/var
	want, _ := filepath.EvalSymlinks(repo)
	if got != want {
		t.Fatalf("ResolveFrom(worktree).Root = %q, want main repo %q", got, want)
	}
}
