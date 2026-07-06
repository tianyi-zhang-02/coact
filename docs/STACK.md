# Technology decisions

Last updated: 2026-07-06

## Core: Go, single static binary

The core — CLI, coordination substrate, presence sidecar, dashboard — is written
in Go and ships as **one static binary with zero runtime dependencies**.

Rationale, ranked by the project's priorities:

1. **Easy install is a feature, and a competitive wedge.** The nearest prior
   tool (AgentBridge) requires the Bun runtime, a global npm install, and a
   plugin-marketplace registration, with a documented "installed but won't run"
   failure mode. A Go binary installs in one line (`brew`, `curl | sh`,
   `go install`) with no runtime to manage. The tools developers actually keep
   — ripgrep, `gh`, fzf, lazygit — are single binaries for this reason.
2. **Right shape for the workload.** A CLI plus a long-lived presence sidecar
   and daemon maps cleanly onto goroutines and signal handling in one process.
3. **The substrate is filesystem primitives.** Atomic create (`O_EXCL`), atomic
   rename, and directory watching are native and fast; nothing exotic is needed.
4. **Cross-platform from one build.** `GOOS`/`GOARCH` cross-compilation covers
   macOS, Linux, and Windows on amd64 and arm64 via goreleaser.

## Cross-platform strategy

Coverage target: **macOS, Linux, Windows** on **amd64 + arm64**.

- Coordination primitives use only portable operations: `os.OpenFile` with
  `O_CREATE|O_EXCL` for atomic lock creation, and `os.Rename` for atomic
  replace. Go's `os.Rename` maps to `MoveFileEx(..., REPLACE_EXISTING)` on
  Windows, so replace-over-existing is atomic on all three platforms.
- OS-specific behavior is isolated behind build tags in `internal/platform`:
  - process-liveness (`ProcessAlive`) uses signal 0 on Unix and a process-handle
    probe on Windows.
- Line endings: the repo sets `core.autocrlf false`; state files are written
  with `\n` explicitly.
- No dependency on POSIX-only advisory locks (`flock`) — locking is the coact
  protocol's own lock files, which are portable.

## Dependencies

**Zero external Go modules in the core** (standard library only) for as long as
practical. Consequences of that choice in v0:

- Machine config is `.coact/config.json` (JSON via `encoding/json`) rather than
  YAML, to avoid a YAML dependency. The SPEC shows YAML for readability; the
  implementation reads JSON. A YAML front-end is a later, optional convenience.
- Subcommand routing is hand-rolled on `flag` rather than a CLI framework.

## Messaging plane (optional, later phase)

The one place another ecosystem is genuinely stronger is the messaging plane
(MCP channel + Codex app-server protocol). Plan: keep the core pure Go; if the
Go MCP/Codex integration proves costly, ship messaging as a **separate opt-in
sidecar** (possibly reusing AgentBridge's bus) rather than pulling that runtime
into the binary everyone installs. Polyglot only at the optional edge.

## Distribution

- Homebrew tap (macOS/Linux).
- `curl -fsSL .../install.sh | sh` pulling a prebuilt binary from GitHub
  releases (see `install.sh`).
- `go install` for developers.
- Prebuilt release binaries via goreleaser (see `.goreleaser.yaml`).

No npm, no Bun, no Python, no plugin marketplace for the core. Agent wiring is
written directly by `coact init` (a hook entry for Claude Code, an `AGENTS.md`
fragment for Codex) and shells out to the single binary.
