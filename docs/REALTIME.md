# Real-time bridge (experimental)

By default agents coordinate turn-based (they read their inbox at the start of a
turn). The real-time bridge pushes messages between them **mid-turn**, CLI to
CLI, with no daemon — the two sides coordinate through the shared inbox.

```
Claude ──reply tool──▶ .coact/inbox/codex.md ──▶ coact bridge codex ──turn/start·steer──▶ codex app-server
Claude ◀─push mid-turn─ .coact/inbox/claude.md ◀── coact bridge codex ◀──agentMessage/delta── codex app-server
        (coact channel claude)
```

## Prerequisites

- **Claude Code ≥ 2.1.80** — channels are a research preview added in 2.1.80.
  Older versions can't receive mid-turn pushes (you'll still get turn-based).
- Anthropic auth (claude.ai / Pro / Max / Console key) — channels aren't
  available on Bedrock/Vertex/Foundry.
- **codex** on your PATH (the bridge drives `codex app-server`).
- coact installed, and `coact init` run in the repo.

## Setup (one-step + two commands)

```sh
coact channel install            # registers coact-claude in .mcp.json
```

Then, in two terminals:

```sh
# terminal 1 — drive Codex, relaying to/from the inbox
coact bridge codex

# terminal 2 — launch Claude with the channel enabled
claude --dangerously-load-development-channels server:coact-claude
```

Now message Codex from Claude (or `coact msg codex "…"`); Codex works and its
reply appears in Claude **mid-turn** as a `<channel source="coact-codex">` event.
`coact doctor` reports whether the channel is registered and codex is reachable.
Remove all wiring with `coact deinit`.

## Two things confirmed against real codex

The bridge is tested against a fake codex; two details are provider-specific and
verified on first real run (both behind clean seams):

- **stdio framing** — `codex app-server` may frame JSON-RPC newline-delimited or
  LSP-style `Content-Length`. `internal/codex` uses a swappable `Codec` (newline
  by default). If nothing flows, that's the first thing to switch.
- **approval result shape** — the bridge auto-approves Codex's `*/requestApproval`
  requests; the exact result payload is confirmed live.

Override the app-server command if needed: `COACT_CODEX_CMD="codex app-server" coact bridge codex`.

## How it degrades

Without the upgrade / without the channel, nothing breaks — you fall back to the
turn-based path (`coact inbox` / `coact msg`), which works on any Claude Code
version. Real-time is strictly additive.
