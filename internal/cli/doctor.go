package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tianyi-zhang-02/coact/internal/adapter"
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
		cfg = config.Default()
	} else {
		ok(fmt.Sprintf("config ok (mode %s)", cfg.Mode))
		if missing := missingProtectedPaths(cfg); len(missing) > 0 {
			bad("config is missing required protected paths: " + strings.Join(missing, ", ") + " — run `coact init`")
		} else {
			ok("machine-managed coordination state is policy-protected")
		}
	}

	if cmd, cargs, found := wiredHookCommand(p.Root); found {
		ok("Claude hook wired: " + strings.TrimSpace(cmd+" "+strings.Join(cargs, " ")))
		bin := cmd
		if fields := strings.Fields(cmd); len(fields) > 0 {
			bin = fields[0] // command is "<binary> hook claude"
		}
		if _, err := os.Stat(bin); err != nil {
			bad("the wired hook binary is missing at that path — re-run `coact init`")
		}
	} else {
		bad("Claude hook is not wired in .claude/settings.json — run `coact init`")
	}

	for _, ac := range cfg.Agents {
		ad, found := adapter.Get(ac.ID)
		if !found {
			continue
		}
		path := filepath.Join(p.Root, ad.ContractFile)
		if contractCurrent(path, ad.Contract()) {
			ok(ad.ContractFile + " has the " + ad.ID + " contract")
		} else if hasCoactBlock(path) {
			warn(ad.ContractFile + " has a stale " + ad.ID + " contract — run `coact init` to refresh it")
		} else {
			warn(ad.ContractFile + " is missing the " + ad.ID + " contract — run `coact init`")
		}
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

	// Real-time bridge readiness (optional; not required for coordination).
	if channelMCPInstalled(p.Root, "claude") {
		ok("real-time channel registered in .mcp.json (needs Claude Code >= 2.1.80)")
	} else {
		warn("real-time channel not installed (optional) — run `coact channel install`")
	}
	if _, err := exec.LookPath("codex"); err == nil {
		ok("codex on PATH (for `coact bridge codex`)")
	} else {
		warn("codex not on PATH — `coact bridge codex` needs it for real-time")
	}

	fmt.Println()
	if healthy {
		fmt.Println("coact is set up correctly.")
		return 0
	}
	fmt.Println("coact found problems above; fix the FAIL lines (usually: re-run `coact init`).")
	return 1
}

func missingProtectedPaths(cfg *config.Config) []string {
	if cfg == nil {
		return append([]string(nil), config.Default().Policy.ProtectedPaths...)
	}
	present := map[string]bool{}
	for _, path := range cfg.Policy.ProtectedPaths {
		present[path] = true
	}
	var missing []string
	for _, path := range config.Default().Policy.ProtectedPaths {
		if !present[path] {
			missing = append(missing, path)
		}
	}
	return missing
}

func contractCurrent(path, body string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	want := coactBegin + "\n" + strings.TrimRight(body, "\n") + "\n" + coactEnd
	return strings.Contains(string(data), want)
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
