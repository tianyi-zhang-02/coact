package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

// cmdDoctor reports whether coact is set up correctly and — crucially — runs a
// self-test that confirms the enforcement engine works on this machine, without
// needing a second live agent.
func cmdDoctor(args []string) int {
	healthy := true
	ok := func(m string) { fmt.Printf("  ok    %s\n", m) }
	warn := func(m string) { fmt.Printf("  warn  %s\n", m) }
	bad := func(m string) { fmt.Printf("  FAIL  %s\n", m); healthy = false }

	if _, err := exec.LookPath("coact"); err == nil {
		ok("coact is on your PATH")
	} else {
		warn("coact is not on your PATH (the hook uses an absolute path, so this is cosmetic)")
	}

	p, err := project.Find()
	if err != nil {
		bad("not initialized in this directory — run `coact init`")
		return 1
	}
	ok("workspace: " + p.Root)

	cfg, err := config.Load(p.ConfigPath())
	if err != nil {
		bad("config invalid: " + err.Error())
	} else {
		ok(fmt.Sprintf("config ok (mode %s)", cfg.Mode))
	}

	if cmd, cargs, found := wiredHookCommand(p.Root); found {
		ok("Claude hook wired: " + strings.TrimSpace(cmd+" "+strings.Join(cargs, " ")))
		if _, err := os.Stat(cmd); err != nil {
			bad("the wired hook binary is missing at that path — re-run `coact init`")
		}
	} else {
		bad("Claude hook is not wired in .claude/settings.json — run `coact init`")
	}

	if hasCoactBlock(filepath.Join(p.Root, "CLAUDE.md")) {
		ok("CLAUDE.md has the coact contract")
	} else {
		warn("CLAUDE.md is missing the coact contract — run `coact init`")
	}
	if hasCoactBlock(filepath.Join(p.Root, "AGENTS.md")) {
		ok("AGENTS.md has the coact contract")
	} else {
		warn("AGENTS.md is missing the coact contract — run `coact init`")
	}

	sessions, _ := presence.List(p.SessionDir())
	live := 0
	for _, s := range sessions {
		if presence.IsLive(p.SessionDir(), s.Agent, cfg.Presence.TTLSeconds) {
			live++
		}
	}
	ok(fmt.Sprintf("%d live participant(s)", live))

	if pass, detail := selfTest(); pass {
		ok("enforcement self-test passed — " + detail)
	} else {
		bad("enforcement self-test FAILED — " + detail)
	}

	fmt.Println()
	if healthy {
		fmt.Println("coact is set up correctly.")
		return 0
	}
	fmt.Println("coact found problems above; fix the FAIL lines (usually: re-run `coact init`).")
	return 1
}

func hasCoactBlock(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), coactBegin)
}

// selfTest builds a throwaway workspace, plants a lock as one agent, and checks
// that the gate blocks a conflicting edit, allows a free path, and gates a
// protected path — exercising the real lock + policy + hook decision code.
func selfTest() (bool, string) {
	root, err := os.MkdirTemp("", "coact-selftest-*")
	if err != nil {
		return false, "could not create a temp workspace"
	}
	defer os.RemoveAll(root)

	p := &project.Project{Root: root}
	for _, d := range []string{p.CoactDir(), p.LocksDir(), p.SessionDir(), p.JournalDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return false, "temp workspace setup failed"
		}
	}
	if err := config.Default().Save(p.ConfigPath()); err != nil {
		return false, "temp config save failed"
	}
	cfg, _ := config.Load(p.ConfigPath())

	m := lockmgr.New(p, cfg)
	if _, err := m.Acquire("agentA", filepath.Join(root, "src")); err != nil {
		return false, "lock acquire failed: " + err.Error()
	}

	pay := func(rel string) hookPayload {
		var pl hookPayload
		pl.Cwd = root
		pl.ToolName = "Edit"
		pl.ToolInput.FilePath = filepath.Join(root, rel)
		return pl
	}
	if deny, _ := hookDecision("agentB", pay("src/file.go")); !deny {
		return false, "a conflicting edit was NOT blocked"
	}
	if deny, _ := hookDecision("agentB", pay("elsewhere/thing.go")); deny {
		return false, "a free path was wrongly blocked"
	}
	if deny, _ := hookDecision("agentB", pay(".coact/config.json")); !deny {
		return false, "a protected path was NOT gated"
	}
	return true, "blocks conflicts, allows free paths, gates protected paths"
}
