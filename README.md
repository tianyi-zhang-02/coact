# coact

**Govern multiple coding agents working in one repository.**

coact is the governance core for multi-agent coding. It turns a working
directory into a coordinated, auditable workspace that two or more agents
(e.g. Claude Code and Codex) can share without corrupting each other's work,
and ships as a single static binary (Go, zero runtime dependencies).

Getting agents to talk to each other is the easy part. coact is built for the
harder problem underneath: making concurrent agents **safe, controllable, and
cheap** to run against the same files.

## Why

Running several agents in one directory raises three problems coact is
organized around:

- **Security** — each agent works in its own `git` worktree by default;
  writes are scoped by a capability policy; protected paths need a human gate;
  every action lands in an append-only journal. A prompt-injected or wrong agent
  is contained and reviewable, not already in your files.
- **Controllability** — the plan is an explicit task board you own and edit, not
  an emergent chat between agents. Integration happens through merge gates you
  approve. All state is plain, inspectable files.
- **Cost** — coordination lives in the filesystem (locks, board, journal —
  zero tokens), not in the agents' context windows. Concurrency and real-time
  messaging are opt-in, not the always-on baseline.

Real-time messaging between agents (cross-review, hand-offs) is an **optional
plane on top** — and every message that crosses is policy-gated and journaled,
so it never bypasses the governance core.

## Status

Early. The coordination substrate is being built first. See
[docs/SPEC.md](docs/SPEC.md) for the protocol and [docs/STACK.md](docs/STACK.md)
for the technology decisions.

Working today:

```
coact init                 # scaffold .coact/ in the current repo
coact sidecar              # per-session presence heartbeat (run in background)
coact status               # live participants, current tasks, active locks
coact lock <path>          # advisory write-intent lock (prefix-aware conflicts)
coact unlock <path>        # release a lock you hold
coact task add "<title>"   # add a task to the shared board
coact board                # list tasks and owners
coact claim <id>           # claim a task (serialized; no double-claims)
coact done <id>            # mark your task done
```

Every one of these lands an event in the append-only journal, and locks are
only stolen from a participant that is both past its TTL and not live per
presence — so a long build or a long reasoning turn never loses its lock.

## Install

Prebuilt single binary, no runtime needed (macOS, Linux, Windows):

```sh
# from source (requires Go 1.22+)
go install github.com/tianyi-zhang-02/coact/cmd/coact@latest

# or build locally
git clone https://github.com/tianyi-zhang-02/coact && cd coact
go build -o coact ./cmd/coact
```

Release binaries and a one-line install script land with the first tagged
release.

## Platforms

`darwin`, `linux`, `windows` — `amd64` and `arm64`. The coordination primitives
use only portable filesystem operations (atomic create, atomic rename); the few
OS-specific pieces (process-liveness checks) are isolated behind build tags in
`internal/platform`.

## License

MIT — see [LICENSE](LICENSE).
