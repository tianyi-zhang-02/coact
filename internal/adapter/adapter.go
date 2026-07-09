// Package adapter is CoAct's agent registry: each supported coding agent is a
// declarative descriptor (binary, contract file, enforcement level, contract
// text). Adding an agent is adding an entry here — the rest of CoAct (init,
// launchers, doctor) iterates the registry rather than hardcoding claude/codex.
package adapter

import "fmt"

// Adapter describes how one coding agent participates in CoAct.
type Adapter struct {
	ID           string // participant id, e.g. "claude"
	Binary       string // CLI to launch, e.g. "claude"
	ContractFile string // where its contract is injected, e.g. "CLAUDE.md"
	HardHook     bool   // true if it gets a hard pre-write hook (L2); else L1
}

var builtin = []Adapter{
	{ID: "claude", Binary: "claude", ContractFile: "CLAUDE.md", HardHook: true},
	{ID: "codex", Binary: "codex", ContractFile: "AGENTS.md", HardHook: false},
	{ID: "gemini", Binary: "gemini", ContractFile: "GEMINI.md", HardHook: false},
}

// All returns every built-in adapter.
func All() []Adapter { return builtin }

// Get returns the adapter for id.
func Get(id string) (Adapter, bool) {
	for _, a := range builtin {
		if a.ID == id {
			return a, true
		}
	}
	return Adapter{}, false
}

// Enforcement is a human-readable tier.
func (a Adapter) Enforcement() string {
	if a.HardHook {
		return "L2 (hook hard-blocks)"
	}
	return "L1 (self-enforced via contract)"
}

const coordinate = `Coordinate with the other agents in this repo:
- Read ".coact/team.md" and ".coact/memory/project.md" at the start of each session.
- Divide work on the board: "coact board", "coact claim <id>", "coact done <id>".
- Talk directly through the local inbox: run "coact inbox" at the start of each
  turn; send with "coact @codex \"...\"", "coact @claude \"...\"", or
  "coact @all \"...\""; and use "coact handoff <agent> \"context\"" if you stop.
- Planning phases live under ".coact/runs/<run>/"; write proposals there, read
  peer proposals, then let the configured final_task_distributor create tasks.
- "coact status" and "coact log" show the shared picture.`

// Contract returns the collaboration contract injected into the agent's context
// file. Hook-enforced agents are told their edits are gated automatically;
// self-enforced agents are told to lock before editing.
func (a Adapter) Contract() string {
	if a.HardHook {
		return fmt.Sprintf(`# coact collaboration contract (%[1]s)

You share this repository with other agents, coordinated by coact.

Your file edits are gated automatically by a coact hook:
- If an edit is denied with "coact: <path> is locked by <agent>", another agent
  is working there. Do NOT force it — switch to other files or wait, and run
  "coact status" to see who holds what.
- Allowed edits record your lock automatically; you never run "coact lock" by hand.

%[2]s

This session should have COACT_AGENT=%[1]s set.
`, a.ID, coordinate)
	}
	return fmt.Sprintf(`# coact collaboration contract (%[1]s)

You share this repository with other agents, coordinated by coact. Your edits are
NOT gated automatically, so you MUST follow this protocol yourself.

Before editing ANY file or directory:
  1. Run: coact lock <path>
  2. If it prints "denied", STOP — another agent holds it. Choose other work or
     wait, and re-check with: coact lock <path> --check
  3. Edit only after the lock is granted. Run "coact unlock <path>" when done.

%[2]s

This session should have COACT_AGENT=%[1]s set.
`, a.ID, coordinate)
}
