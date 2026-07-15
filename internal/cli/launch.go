package cli

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tianyi-zhang-02/coact/internal/adapter"
	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/journal"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

func cmdClaude(args []string) int      { return launchAdapter("claude", args) }
func cmdCodex(args []string) int       { return launchAdapter("codex", args) }
func cmdAntigravity(args []string) int { return launchAdapter("antigravity", args) }

func launchAdapter(id string, args []string) int {
	ad, ok := adapter.Get(id)
	if !ok {
		fmt.Fprintf(os.Stderr, "coact: no adapter %q (see `coact adapters`)\n", id)
		return 2
	}
	return launchAgent(ad.ID, ad.Binary, args)
}

func launchAgent(agent, binary string, args []string) int {
	useWorktree := false
	var pass []string
	for _, a := range args {
		if a == "--worktree" {
			useWorktree = true
			continue
		}
		pass = append(pass, a)
	}

	p, cfg, ok := loadProject()
	if !ok {
		return 1
	}
	bin, err := exec.LookPath(binary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %q is not installed or not on your PATH\n", binary)
		return 1
	}

	workdir := ""
	if useWorktree || cfg.Mode == "worktree" {
		wt, created, err := worktreeAdd(p.Root, agent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "coact: %v\n", err)
			return 1
		}
		workdir = wt
		verb := "using"
		if created {
			verb = "created"
		}
		fmt.Fprintf(os.Stderr, "coact: %s worktree for %s at %s (branch %s)\n", verb, agent, wt, branchName(agent))
	}
	return runWrapped(p, cfg, agent, bin, pass, workdir)
}

// runWrapped owns the agent's coact session in a single process: it registers
// presence, heartbeats while the agent runs, forwards signals to the agent, and
// on exit marks the session stopped and releases the agent's locks. This
// replaces the manual `export COACT_AGENT=...; coact sidecar &; <agent>` dance.
func runWrapped(p *project.Project, cfg *config.Config, agent, bin string, args []string, workdir string) int {
	iv := cfg.Presence.IntervalSeconds
	if iv <= 0 {
		iv = 20
	}
	_ = presence.Register(p.SessionDir(), agent, "working")
	_ = journal.Append(p.JournalDir(), agent, "session.start", map[string]string{"mode": "launcher"})

	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(time.Duration(iv) * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_ = presence.Register(p.SessionDir(), agent, "working")
			case <-stop:
				return
			}
		}
	}()

	cmd := exec.Command(bin, args...)
	cmd.Env = coactAgentEnv(os.Environ(), agent)
	if workdir != "" {
		cmd.Dir = workdir
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	if err := cmd.Start(); err != nil {
		close(stop)
		fmt.Fprintf(os.Stderr, "coact: could not start %s: %v\n", bin, err)
		return 1
	}
	go func() {
		for s := range sigCh {
			_ = cmd.Process.Signal(s)
		}
	}()

	runErr := cmd.Wait()

	close(stop)
	_ = presence.Register(p.SessionDir(), agent, "stopped")
	_ = journal.Append(p.JournalDir(), agent, "session.stop", nil)
	m := lockmgr.New(p, cfg)
	if n, _ := m.ReleaseAll(agent); n > 0 {
		fmt.Fprintf(os.Stderr, "coact: released %d lock(s) held by %s\n", n, agent)
	}

	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "coact: %s exited abnormally: %v\n", agent, runErr)
		return 1
	}
	return 0
}

func coactAgentEnv(base []string, agent string) []string {
	env := append([]string{}, base...)
	env = setEnv(env, "COACT_AGENT", agent)
	if exe, err := os.Executable(); err == nil && exe != "" {
		if abs, err := filepath.Abs(exe); err == nil {
			exe = abs
		}
		env = setEnv(env, "COACT_BIN", exe)
		if dir := filepath.Dir(exe); dir != "." && dir != "" {
			env = prependPath(env, dir)
		}
	}
	return env
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if len(item) >= len(prefix) && item[:len(prefix)] == prefix {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func prependPath(env []string, dir string) []string {
	for i, item := range env {
		if len(item) >= len("PATH=") && item[:len("PATH=")] == "PATH=" {
			if item == "PATH=" {
				env[i] = "PATH=" + dir
			} else {
				env[i] = "PATH=" + dir + string(os.PathListSeparator) + item[len("PATH="):]
			}
			return env
		}
	}
	return append(env, "PATH="+dir)
}
