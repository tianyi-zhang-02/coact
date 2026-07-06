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

// hookClaude implements the Claude Code PreToolUse gate. It reads the payload,
// asks hookDecision whether to block, and emits a deny decision if so.
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

	deny, reason := hookDecision(agent, pl)
	if !deny {
		return 0
	}
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

// hookDecision decides whether to block an edit and why. It fails OPEN
// (deny=false) on any error or a repo without coact, so a coact problem can
// never brick editing. A "deny" is emitted only for a genuine foreign lock
// conflict or a policy violation; on allow it records the lock and stays silent.
func hookDecision(agent string, pl hookPayload) (deny bool, reason string) {
	if !editingTools[pl.ToolName] {
		return false, ""
	}
	path := pl.ToolInput.FilePath
	if path == "" {
		path = pl.ToolInput.NotebookPath
	}
	if path == "" {
		return false, ""
	}

	dir := pl.Cwd
	if dir == "" {
		dir, _ = os.Getwd()
	}
	p, err := project.FindFrom(dir)
	if err != nil {
		return false, "" // coact not in use here → fail open
	}
	cfg, err := config.Load(p.ConfigPath())
	if err != nil {
		return false, ""
	}

	// Hook-only liveness: each gated edit refreshes presence.
	_ = presence.Beat(p.SessionDir(), agent, "working", "")

	m := lockmgr.New(p, cfg)
	res, err := m.Acquire(agent, path)
	if err != nil || res == nil {
		return false, "" // fail open on internal error or path outside repo
	}
	if res.Acquired {
		return false, "" // allow — the lock now records the agent's intent
	}
	if res.Conflict != nil {
		return true, fmt.Sprintf(
			"coact: %s is locked by %q since %s. Another agent is working there — coordinate via `coact status` or pick different files.",
			res.Path, res.Conflict.Owner, res.Conflict.AcquiredAt)
	}
	// Policy denial (no conflicting lock).
	return true, "coact: " + res.Detail
}
