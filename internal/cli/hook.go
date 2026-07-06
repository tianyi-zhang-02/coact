package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tianyi-zhang-02/coact/internal/config"
	"github.com/tianyi-zhang-02/coact/internal/lockmgr"
	"github.com/tianyi-zhang-02/coact/internal/presence"
	"github.com/tianyi-zhang-02/coact/internal/project"
)

// editingTools are the tool names whose writes coact gates.
var editingTools = map[string]bool{
	"Edit": true, "Write": true, "MultiEdit": true, "NotebookEdit": true,
}

// hookPayload is the subset of the Claude Code PreToolUse stdin JSON we use.
type hookPayload struct {
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath     string `json:"file_path"`
		NotebookPath string `json:"notebook_path"`
	} `json:"tool_input"`
}

func cmdHook(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: coact hook <adapter>   (adapters: claude)")
		return 2
	}
	switch args[0] {
	case "claude":
		return hookClaude(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "coact: unknown hook adapter %q\n", args[0])
		return 2
	}
}

// hookClaude implements the Claude Code PreToolUse gate.
//
// It fails OPEN: any error, or a repo without coact, results in exit 0 with no
// decision so a coact problem can never brick the user's editing. It only ever
// emits a "deny" decision for a genuine foreign lock conflict; on allow it stays
// silent (deferring to Claude's normal permission flow) but records the lock.
func hookClaude(args []string) int {
	fs := flag.NewFlagSet("hook claude", flag.ContinueOnError)
	agentFlag := fs.String("agent", "claude", "participant id")
	_ = fs.Parse(args)

	agent := *agentFlag
	if e := os.Getenv("COACT_AGENT"); e != "" {
		agent = e
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return 0
	}
	var pl hookPayload
	if json.Unmarshal(data, &pl) != nil {
		return 0
	}
	if !editingTools[pl.ToolName] {
		return 0
	}
	path := pl.ToolInput.FilePath
	if path == "" {
		path = pl.ToolInput.NotebookPath
	}
	if path == "" {
		return 0
	}

	dir := pl.Cwd
	if dir == "" {
		dir, _ = os.Getwd()
	}
	p, err := project.FindFrom(dir)
	if err != nil {
		return 0 // coact not in use here → fail open
	}
	cfg, err := config.Load(p.ConfigPath())
	if err != nil {
		return 0
	}

	// Hook-only liveness: each gated edit refreshes presence.
	_ = presence.Beat(p.SessionDir(), agent, "working", "")

	m := lockmgr.New(p, cfg)
	res, err := m.Acquire(agent, path)
	if err != nil || res == nil {
		return 0 // fail open on internal error or path outside repo
	}
	if res.Acquired {
		return 0 // allow (silent) — the lock now records claude's intent
	}

	reason := fmt.Sprintf(
		"coact: %s is locked by %q since %s. Another agent is working there — coordinate via `coact status` or pick different files.",
		res.Path, res.Conflict.Owner, res.Conflict.AcquiredAt)
	out := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
	return 0
}
