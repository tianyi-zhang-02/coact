package cli

const boardTemplate = `# Task board

Tasks below carry machine-readable metadata in an HTML comment. The checkbox
mirrors state for humans: [ ] todo, [~] doing, [x] done, [!] blocked.

## Backlog

- [ ] Example: describe a task here <!-- coact: id=T-001 state=todo owner= -->

## Done
`

const claudeFragment = `# coact collaboration contract (Claude Code)

You share this repository with another agent. Coordinate through coact.

- Before editing a file or directory, acquire a write-intent lock:
  ` + "`coact lock <path> --agent claude`" + `
  If it prints "denied", do NOT edit — the other agent holds it. Pick other work
  or wait, and re-check with ` + "`coact lock <path> --check --agent claude`" + `.
- Release when done: ` + "`coact unlock <path> --agent claude`" + `.
- Check the shared picture any time with ` + "`coact status`" + `.
- Claim work from ` + "`.coact/board.md`" + ` rather than inventing overlapping tasks.

Set COACT_AGENT=claude in this session so the id is implicit.
`

const codexFragment = `# coact collaboration contract (Codex)

You share this repository with another agent. Coordinate through coact.

- Before editing a file or directory, acquire a write-intent lock:
  ` + "`coact lock <path> --agent codex`" + `
  If it prints "denied", do NOT edit — the other agent holds it. Pick other work
  or wait, and re-check with ` + "`coact lock <path> --check --agent codex`" + `.
- Release when done: ` + "`coact unlock <path> --agent codex`" + `.
- Check the shared picture any time with ` + "`coact status`" + `.
- Claim work from ` + "`.coact/board.md`" + ` rather than inventing overlapping tasks.

Set COACT_AGENT=codex in this session so the id is implicit.
`
