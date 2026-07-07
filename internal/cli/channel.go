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
	if len(args) >= 1 && args[0] == "install" {
		return channelInstall(args[1:])
	}

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

// channelInstall registers the coact channel in .mcp.json and prints the
// real-time setup — the one-step prep for the live bridge.
func channelInstall(args []string) int {
	fs := flag.NewFlagSet("channel install", flag.ContinueOnError)
	agentFlag := fs.String("agent", "claude", "agent to register a channel for")
	if _, err := parseInterspersed(fs, args); err != nil {
		return 2
	}
	agent := sanitizeAgent(*agentFlag)

	p, _, ok := loadProject()
	if !ok {
		return 1
	}
	added, err := ensureChannelMCP(p.Root, agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "coact: %v\n", err)
		return 1
	}
	name := channelServerName(agent)
	if added {
		fmt.Printf("registered %s in .mcp.json\n", name)
	} else {
		fmt.Printf("%s is already registered in .mcp.json\n", name)
	}
	fmt.Printf(`
real-time setup (needs Claude Code >= 2.1.80):
  1. terminal 1 — drive Codex, relaying to/from the inbox:
       coact bridge codex
  2. terminal 2 — launch Claude with the channel enabled:
       claude --dangerously-load-development-channels server:%s

Messages between the agents then arrive mid-turn. Undo with "coact deinit".
`, name)
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
