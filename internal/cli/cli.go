// Package cli implements the coact command-line interface.
package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Run dispatches a subcommand and returns a process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		// Bare `coact` opens the control center only in an interactive terminal.
		// In a pipe, script, or CI, stdout is not a TTY, so it prints usage and
		// exits instead of hanging on a blocking server. `coact ui` always
		// launches regardless of TTY.
		if isInteractive() {
			return cmdUI(nil)
		}
		printUsage()
		return 1
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "ui":
		return cmdUI(rest)
	case "init":
		return cmdInit(rest)
	case "deinit":
		return cmdDeinit(rest)
	case "doctor":
		return cmdDoctor(rest)
	case "claude":
		return cmdClaude(rest)
	case "codex":
		return cmdCodex(rest)
	case "gemini":
		return cmdGemini(rest)
	case "adapters":
		return cmdAdapters(rest)
	case "worktree":
		return cmdWorktree(rest)
	case "merge":
		return cmdMerge(rest)
	case "msg":
		return cmdMsg(rest)
	case "inbox":
		return cmdInbox(rest)
	case "handoff":
		return cmdHandoff(rest)
	case "channel":
		return cmdChannel(rest)
	case "bridge":
		return cmdBridge(rest)
	case "status":
		return cmdStatus(rest)
	case "lock":
		return cmdLock(rest)
	case "unlock":
		return cmdUnlock(rest)
	case "board":
		return cmdBoard(rest)
	case "claim":
		return cmdClaim(rest)
	case "done":
		return cmdDone(rest)
	case "task":
		return cmdTask(rest)
	case "sidecar":
		return cmdSidecar(rest)
	case "log":
		return cmdLog(rest)
	case "policy":
		return cmdPolicy(rest)
	case "hook":
		return cmdHook(rest)
	case "version", "--version", "-v":
		return cmdVersion()
	case "versions":
		return cmdVersions(rest)
	case "update":
		return cmdUpdate(rest)
	case "switch":
		return cmdSwitch(rest)
	case "help", "--help", "-h":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "coact: unknown command %q\n\n", cmd)
		printUsage()
		return 1
	}
}

func isInteractive() bool {
	st, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

// parseInterspersed parses flags that may appear before or after positional
// arguments (Go's flag package otherwise stops at the first positional), and
// returns the collected positional arguments.
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			break
		}
		positionals = append(positionals, fs.Arg(0))
		rest = fs.Args()[1:]
	}
	return positionals, nil
}

func agentID(flagVal string) string {
	if flagVal != "" {
		return sanitizeAgent(flagVal)
	}
	if e := os.Getenv("COACT_AGENT"); e != "" {
		return sanitizeAgent(e)
	}
	return "human"
}

// sanitizeAgent restricts an agent id to the SPEC §1.1 charset [a-z0-9_-]. The
// id is used in session filenames (session/<agent>.json), so this prevents a
// stray or hostile value (e.g. "../../etc") from escaping the .coact directory.
func sanitizeAgent(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	var b strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "agent"
	}
	return b.String()
}

func printUsage() {
	fmt.Print(`coact — govern two coding agents in one repository.

Usage:
  coact [command] [flags]

Commands:
  ui               Open the local CoAct control center (--addr, --port, --no-open, --lang)
  init             Scaffold .coact/ and wire the agents in this repository
  doctor           Check the setup and self-test enforcement (no agent needed)
  deinit           Remove coact's wiring (--purge also deletes .coact/)
  claude [args]    Launch Claude Code wired into coact (--worktree to isolate)
  codex [args]     Launch Codex wired into coact (--worktree to isolate)
  gemini [args]    Launch Gemini CLI wired into coact
  adapters         List the agents coact can coordinate
  worktree         Manage per-agent git worktrees (add | list | rm)
  merge <agent>    Merge an agent's coact/<agent> branch (stops on conflict)
  status           Show live participants and active locks
  lock <path>      Acquire an advisory write-intent lock
  unlock <path>    Release a lock you hold
  board            List tasks on the shared board
  claim <id>       Claim a task from the board
  done <id>        Mark a claimed task done
  msg <to> <text>  Send a message to another agent
  inbox            Read your messages from other agents
  handoff <to>     Hand your tasks + context to another agent
  channel <agent>  Run the Claude Code channel MCP server (real-time push)
  bridge codex     Drive Codex's app-server, relaying live to/from Claude
  task add "<t>"   Add a task to the board
  sidecar          Run the presence heartbeat for this session
  log              Show recent journal events (oversight view)
  policy           Show or check write policy (check <path> | show)
  hook claude      PreToolUse gate for Claude Code (wired by init)
  versions         List locally managed coact versions (experimental)
  update           Download a release into ~/.coact (experimental; checksum-verified, unsigned)
  switch <version> Switch the ~/.coact/coact shim to a managed version (experimental)
  version          Print version
  help             Show this help

Common flags:
  --agent <id>     Participant id (default: $COACT_AGENT, else "human")

Examples:
  coact                  # open the local UI
  coact ui --no-open --port 7331
  coact init
  export COACT_AGENT=claude
  coact sidecar &          # keep this session live
  coact task add "Add rate limiting"
  coact claim T-001
  coact lock src/api
  coact update --channel stable
  coact status
`)
}
