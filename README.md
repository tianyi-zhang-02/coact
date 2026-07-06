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

Running several agents in one directory raises three problems coact is built
around:

- **Security** — writes are scoped and gated: an agent's edit is blocked when
  another holds that file (enforced by a hook for Claude Code) or when policy
  forbids it — protected paths need a human, and each agent can be confined to a
  write scope. Every action lands in an append-only journal, so a wrong or
  prompt-injected agent is contained and auditable. (Git-worktree isolation is
  on the roadmap.)
- **Controllability** — the plan is an explicit task board you own and edit, not
  an emergent chat between agents. All state is plain, inspectable files.
- **Cost** — coordination lives in the filesystem (locks, board, journal —
  zero tokens), not in the agents' context windows. Concurrency and real-time
  messaging are opt-in, not the always-on baseline.

Real-time messaging between agents (cross-review, hand-offs) is an **optional
plane on top** — and every message that crosses is policy-gated and journaled,
so it never bypasses the governance core.

## Quickstart

Install coact (see below), then in your repository:

```sh
coact init        # wires the Claude Code hook + writes the agent contracts
coact doctor      # verify: checks the wiring and self-tests enforcement
```

`coact doctor` confirms coact works on your machine **without needing a second
agent** — it plants a lock and checks that the gate blocks a conflict, allows a
free path, and gates protected paths.

Then launch each agent in its own terminal — one command each:

```sh
coact claude      # terminal 1 — Claude Code, session managed by coact
coact codex       # terminal 2 — Codex
```

`coact claude` sets the identity, keeps presence live while the agent runs, and
releases the session's locks when it exits — no background process to manage.
(You can still do it by hand: `export COACT_AGENT=claude; coact sidecar &; claude`.)

coact adds a gate; it does **not** require `--dangerously-skip-permissions`, and
the hook **fails open** — if coact ever errors, your editing still works. Remove
all wiring any time with `coact deinit`.

Divide the work on the shared board:

```sh
coact task add "Build auth module"
coact task add "Build API gateway"
coact claim T-002     # claude takes auth
coact claim T-003     # codex takes the gateway
```

Now they work in parallel. If one strays into files the other holds, coact stops
it — for Claude the hook blocks the edit outright:

```
coact: src/gateway/router.go is locked by "codex" since 2026-07-06T21:09:32Z.
Another agent is working there — coordinate via `coact status`.
```

Watch it at any time:

```sh
coact status      # live agents, their current task, and held locks
coact log         # the full audit trail
```

## Commands

| Command | Purpose |
|---|---|
| `coact init` | Wire the hook + contracts in this repo |
| `coact doctor` | Check setup and self-test enforcement (no agent needed) |
| `coact deinit` | Remove coact's wiring (`--purge` also removes `.coact/`) |
| `coact claude` / `coact codex` | Launch an agent with its coact session managed |
| `coact status` | Live participants, current tasks, active locks |
| `coact log` | Recent journal events (oversight view) |
| `coact board` | List tasks and owners |
| `coact task add "<t>"` | Add a task to the board |
| `coact claim <id>` / `done <id>` | Claim / complete a task |
| `coact lock <path>` / `unlock <path>` | Advisory write-intent lock (`unlock --all` frees all yours) |
| `coact policy check <path>` / `show` | Test or view the write policy |
| `coact sidecar` | Per-session presence heartbeat |

Locks are stolen only from a participant that is both past its TTL **and** not
live per presence, so a long build or a long reasoning turn never loses its lock.

## Status

Works today: two-agent coordination (Claude Code + Codex), advisory locks with
Claude Code hook enforcement, a capability policy (protected paths + per-agent
write scopes), the task board, presence, and the journal — as a single
cross-platform binary. On the roadmap: git-worktree isolation with merge gates,
more agent adapters, and the optional messaging plane. See
[docs/ROADMAP.md](docs/ROADMAP.md), [docs/SPEC.md](docs/SPEC.md), and
[docs/STACK.md](docs/STACK.md).

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

## Troubleshooting

- **First, run `coact doctor`.** It pinpoints most problems and self-tests the
  engine.
- **An edit wasn't blocked.** Claude Code reads `.claude/settings.json` at
  startup — restart it after `coact init`. Confirm the hook is wired with
  `coact doctor`. Note enforcement is Claude-side; Codex is L1 (self-enforced via
  `AGENTS.md`).
- **"coact: command not found" from the hook, or the wired binary moved.** The
  hook stores an absolute path to the binary. If you moved or reinstalled coact,
  re-run `coact init`.
- **An agent seems stuck, blocked on files it should own.** During a session an
  agent accumulates locks on the files it edits. Free them with
  `coact unlock --all` (the `coact claude`/`coact codex` launchers do this
  automatically on exit).
- **Watch what's happening live.** `coact status --watch`.
- **Remove everything.** `coact deinit` (add `--purge` to also delete `.coact/`).

## License

MIT — see [LICENSE](LICENSE).
