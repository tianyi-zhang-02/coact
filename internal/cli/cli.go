// Package cli implements the coact command-line interface.
package cli

import (
	"flag"
	"fmt"
	"os"
)

// Run dispatches a subcommand and returns a process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 1
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "init":
		return cmdInit(rest)
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
	case "help", "--help", "-h":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "coact: unknown command %q\n\n", cmd)
		printUsage()
		return 1
	}
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
		return flagVal
	}
	if e := os.Getenv("COACT_AGENT"); e != "" {
		return e
	}
	return "human"
}

func printUsage() {
	fmt.Print(`coact — govern two coding agents in one repository.

Usage:
  coact <command> [flags]

Commands:
  init             Scaffold .coact/ in the current repository
  status           Show live participants and active locks
  lock <path>      Acquire an advisory write-intent lock
  unlock <path>    Release a lock you hold
  board            List tasks on the shared board
  claim <id>       Claim a task from the board
  done <id>        Mark a claimed task done
  task add "<t>"   Add a task to the board
  sidecar          Run the presence heartbeat for this session
  log              Show recent journal events (oversight view)
  policy           Show or check write policy (check <path> | show)
  hook claude      PreToolUse gate for Claude Code (wired by init)
  version          Print version
  help             Show this help

Common flags:
  --agent <id>     Participant id (default: $COACT_AGENT, else "human")

Examples:
  coact init
  export COACT_AGENT=claude
  coact sidecar &          # keep this session live
  coact task add "Add rate limiting"
  coact claim T-001
  coact lock src/api
  coact status
`)
}
