# Roadmap — v0.1.0 MVP

Target ship: **2026-07-09** (3 days from 2026-07-06).

> **Progress (2026-07-06):** ahead of schedule. Day 1 (enforcement hook +
> init auto-wiring + `coact log`) and Day 2 (release pipeline validated
> end-to-end, agent contracts, 30-way concurrency stress test) are done. The
> full two-agent workflow passes in simulation, including the hook blocking a
> stray edit. The one remaining gate before tagging `v0.1.0` is a **live
> two-agent dogfood** — running real Claude Code + Codex sessions through coact.

## What "usable MVP" means

The demo that must work end to end: **one agent is editing a path; another
agent's attempt to edit under it is actually _blocked_ by a hook — not merely
asked** — and a human can see who holds what. If that moment works and coact
installs in one line, it's an MVP.

### Acceptance criteria

1. One-line install on macOS/Linux (GitHub Release binaries + `install.sh`;
   `go install` as backup).
2. `coact init` **auto-wires** Claude Code's `PreToolUse` hook into
   `.claude/settings.json` and drops the Codex `AGENTS.md` contract — no manual
   copy-paste.
3. A real lock conflict **blocks** a Claude edit via the hook, with a clear
   "held by <agent> since …" message.
4. `board` / `claim` / `done` work; `status` truthfully shows live agents,
   locks, and tasks.
5. Journal records every action; the human has an oversight command
   (`coact log` / `status`).
6. README quickstart a stranger can follow in < 5 minutes, tested clean.
7. Tagged `v0.1.0` release, CI-built, cross-platform.

## Scope

**In:** Claude Code + Codex, shared tree, advisory locks with Claude L2 hook
enforcement, board, journal, status, real release.

**Out (v0.2+):** adapter registry / Gemini, capability-policy engine, worktree
mode + merge gates, messaging plane, live TUI dashboard, visual/binary diffs.
Listed in the README as "coming" so the MVP reads as _scoped_, not thin.

## Day-by-day

### Day 1 — 2026-07-07 · make enforcement real (highest risk, so first)
- Verify the current Claude Code `PreToolUse` hook contract (stdin schema,
  block-decision format, exit codes).
- Implement `coact hook claude`: parse the payload, extract `file_path` for
  Edit/Write/MultiEdit/NotebookEdit, map to a repo path, and for `$COACT_AGENT`:
  block on a conflicting foreign lock, else acquire/refresh and allow. The hook
  also beats presence (liveness works with no separate sidecar).
- Extend `coact init` to write the hook into `.claude/settings.json`
  (idempotent) + contract includes.
- Test with synthetic hook payloads.
- **Milestone:** a simulated Claude edit is blocked by a held lock.

### Day 2 — 2026-07-08 · ship pipeline + Codex side + hardening
- goreleaser release workflow: tag → build 6 targets → GitHub Release with
  archives + checksums. Validate with a `v0.1.0-rc1` tag.
- Validate `install.sh` against a real release.
- Tune the Codex `AGENTS.md` contract so Codex reliably runs
  `coact lock --check` before edits (L1).
- Windows/path edge pass on the hook; concurrent-acquire stress test.

### Day 3 — 2026-07-09 · dogfood, polish, ship
- Real acceptance test: run actual Claude Code + Codex through coact, force a
  collision, confirm the block fires and board division works. Reserve most of
  the day for fixes.
- README quickstart rewrite + a short terminal gif.
- Tag `v0.1.0`, publish, verify install on a clean machine.

## Risks & pre-decided cut lines

- **#1 — the hook doesn't hard-block in time.** Scheduled Day-1 morning to find
  out early. Fallback: ship both agents at L1 (advisory) for v0.1.0 and document
  it. Do not let this slip past Day 1.
- **Release pipeline overruns →** cut to `go install` + manually-attached
  binaries; goreleaser lands in v0.1.1.
- **Homebrew tap →** out of scope; `curl | sh` + `go install` suffice.
- **No Codex quota for live testing →** ship the validated `coact lock --check`
  path + documented contract; mark Codex L1 honestly.
- Keep ~30% of each day as bug buffer.

The critical path is Day 1: if the hook blocks a real edit by end of Day 1, the
MVP ships; if not, invoke the L1 fallback and still ship.
