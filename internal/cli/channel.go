package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/tianyi-zhang-02/coact/internal/channel"
)

// cmdChannel runs the Claude Code channel MCP server for an agent. Claude Code
// spawns it over stdio (registered in .mcp.json, launched with
// --dangerously-load-development-channels). Requires Claude Code v2.1.80+.
func cmdChannel(args []string) int {
	fs := flag.NewFlagSet("channel", flag.ContinueOnError)
	peerFlag := fs.String("peer", "", "the other agent (default: the opposite of <agent>)")
	pos, err := parseInterspersed(fs, args)
	if err != nil || len(pos) != 1 {
		fmt.Fprintln(os.Stderr, "usage: coact channel <agent> [--peer <agent>]")
		return 2
	}
	agent := sanitizeAgent(pos[0])
	peer := sanitizeAgent(*peerFlag)
	if peer == "" || peer == "agent" {
		peer = otherAgent(agent)
	}

	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	// stdout is the MCP transport — never print to it; diagnostics go to stderr.
	srv := channel.New(agent, peer, p.InboxDir(), p.JournalDir(), os.Stdin, os.Stdout)
	if err := srv.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "coact channel: %v\n", err)
		return 1
	}
	return 0
}

func otherAgent(a string) string {
	switch a {
	case "claude":
		return "codex"
	case "codex":
		return "claude"
	default:
		return "peer"
	}
}
