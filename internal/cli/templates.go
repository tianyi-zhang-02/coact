package cli

const boardTemplate = `# Task board

Tasks below carry machine-readable metadata in an HTML comment. The checkbox
mirrors state for humans: [ ] todo, [~] doing, [x] done, [!] blocked.

## Backlog

- [ ] Example: describe a task here <!-- coact: id=T-001 state=todo owner= -->

## Done
`

// claudeFragment is injected into CLAUDE.md. Claude Code is gated automatically
// by the PreToolUse hook, so it does not lock by hand — the contract explains
// how to react to denials and how to divide work.
const claudeFragment = `# coact collaboration contract (Claude Code)

You share this repository with another agent, coordinated by coact.

Your file edits are gated automatically by a coact PreToolUse hook:
- If an edit is denied with "coact: <path> is locked by <agent>", another agent
  is working there. Do NOT try to force it — switch to different files or wait.
  Run "coact status" to see who holds what.
- Allowed edits automatically record your lock; you never run "coact lock" by hand.

Divide the work explicitly instead of overlapping:
- Run "coact board" to see tasks, "coact claim <id>" before starting one, and
  "coact done <id>" when finished.
- Run "coact status" for live participants and locks, "coact log" for the audit trail.

This session should have COACT_AGENT=claude set.
`

// codexFragment is injected into AGENTS.md. Codex has no hard pre-write gate
// (L1), so the contract is an explicit protocol it must self-enforce.
const codexFragment = `# coact collaboration contract (Codex)

You share this repository with another agent, coordinated by coact. Your edits
are NOT gated automatically, so you MUST follow this protocol yourself.

Before editing ANY file or directory:
  1. Run: coact lock <path>
  2. If it prints "denied", STOP — another agent holds it. Do not edit. Choose
     other work or wait, and re-check with: coact lock <path> --check
  3. Edit only after the lock is granted.
When you are done with a path, run: coact unlock <path>

Divide the work explicitly:
- Run "coact board" to see tasks, "coact claim <id>" before starting, and
  "coact done <id>" after.
- Run "coact status" for who holds what, and "coact log" for the audit trail.

This session should have COACT_AGENT=codex set so the id is implicit.
`
