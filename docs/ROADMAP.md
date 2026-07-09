# Roadmap — v0.1.0 MVP

Target ship: **2026-07-09** (3 days from 2026-07-06).

> **Progress (2026-07-06):** the MVP is functionally complete and went well
> beyond the original 3-day scope in one session. Shipped: the Claude Code
> enforcement hook (exit-2 block) with `init` auto-wiring; the capability +
> policy engine; one-command UX (`coact doctor` self-test, `coact claude` /
> `coact codex` launchers, `coact deinit`); lock cleanup (`unlock --all` +
> release on session end); `coact status --watch`; a validated release pipeline;
> all CI-green cross-platform. The only remaining gate before tagging `v0.1.0`
> is a **live two-agent dogfood** — real Claude Code + Codex sessions, which
> needs an authenticated terminal. See [ARCHITECTURE.md](ARCHITECTURE.md) for
> the current design.

> **Progress (2026-07-08):** pivoted the default product back to a
> terminal-native coordination layer. `coact` now shows a workspace summary,
> `coact @agent` / `coact @all` provide direct inbox messaging, and
> `coact plan` creates planning runs under `.coact/runs/<run>/`. The optional
> local UI remains experimental; the release is centered on native terminals,
> shared memory, task ownership, locks, policy, and journaled coordination.

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

**In:** Claude Code + Codex + Gemini adapters, native-terminal workflow,
`@agent` inbox messaging, planning runs, shared local memory, shared tree,
advisory locks with Claude L2 hook enforcement, capability policy, board,
journal, status, worktree mode + merge gates, and real release.

**Out (v0.2+):** embedded live terminals, UI model switching, deeper real-time
mid-turn control, autopilot, release signing, visual/binary diffs, and automatic
Chinese polishing inside provider pipelines. The v0.1 line includes the
default-on model-agnostic Chinese expression adapter foundation and CLI diagnostics only; callers can still disable it explicitly.
Listed in the README as "coming" so the first release reads as scoped, not thin.

## v0.1.0 — Initial terminal-native coordination

The acceptance bar for the first public release:

1. `coact` shows a terminal workspace summary and never opens a browser by
   default.
2. `coact @agent` and `coact @all` send local, journaled inbox messages without
   executing shell commands.
3. `coact plan` creates `.coact/runs/<run>/` with a brief, proposal files,
   final-plan file, and inbox notifications for participating agents.
4. `.coact/team.md` defines coordination preferences; `.coact/memory/project.md`
   carries local shared project memory.
5. `coact update` downloads releases into `~/.coact/bin`, verifies SHA-256, and
   switches only the managed `~/.coact/coact` shim.
6. Version descriptions show channel, stability, feature support, and notes so a
   user can choose stable/beta/experimental intentionally.
7. `coact zh check` diagnoses the default-on, model-agnostic Chinese expression
   adapter trigger/protection behavior without calling an external model, and
   `--off` verifies the explicit disable path.

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
