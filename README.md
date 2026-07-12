# CoAct

**English** · [中文](README.zh-CN.md)

<img src="assets/mascot/icon.png" alt="CoBot, the CoAct robot astronaut" width="140">

**Let Claude Code, Codex, and Gemini collaborate in the same repository—while
each keeps its native terminal.**

CoAct is a local coordination and safety layer for coding agents. It gives them
shared project memory, structured planning, task ownership, direct `@agent`
messages, write-intent locks, usage alerts, collaboration reports, and an audit
trail. It does not replace an agent CLI or model provider.

## Start in two minutes

Install CoAct on your `PATH`, then initialize a repository once:

```sh
go install github.com/tianyi-zhang-02/coact/cmd/coact@v1.0.0
cd your-project
coact init
coact doctor
```

Open one native terminal per agent:

```sh
# terminal 1
coact claude

# terminal 2
coact codex

# optional terminal 3
coact gemini
```

You do **not** need a separate management terminal. In any agent prompt, ask it
to run a CoAct command, or run commands in any normal shell:

```sh
coact @codex "Review Claude's proposal before implementation."
coact @all "Read the new brief and propose a plan."
coact inbox
coact status
```

The launcher sets `COACT_AGENT`, `COACT_BIN`, and `PATH`, so an agent launched
with `/some/path/coact codex` can still run bare `coact inbox`.

`v1.0.0` is the first stable terminal-native coordination release.

## The normal workflow

### 1. Set shared preferences

`coact init` creates two human-controlled files:

- `.coact/team.md` — agent roles, planning participants, and final distributor
- `.coact/memory/project.md` — durable project facts and preferences

Do not put secrets in either file.

### 2. Plan together

```sh
coact plan --with codex,claude --distributor codex "Refactor authentication safely"
coact plan status
```

Each agent receives a local inbox message and writes an independent proposal
under `.coact/runs/<run>/`. The distributor waits until proposals say
`Status: ready` and are unlocked, then writes `final-plan.md` and creates board
tasks. Delivery is turn-based by default: an idle agent reads the message on its
next turn. The optional real-time bridge is experimental.

After finishing a proposal, an agent can mark it safely without hand-editing
metadata:

```sh
coact plan ready <run-id>
```

### 3. Execute without stealing work

```sh
coact board
coact claim T-001
coact lock internal/auth
# edit and test
coact unlock internal/auth
coact done T-001
```

Board claims are serialized. Claude Code edits are hard-blocked by its hook when
a path conflicts; Codex and Gemini follow an injected contract, so their shared-
tree enforcement is advisory. Use `coact <agent> --worktree` when stronger
physical isolation is important.

### 4. Message, hand off, and audit

```sh
coact @claude "Please review T-001."
coact handoff codex "Parser is complete; integration tests remain."
coact log -n 50
```

Messages only write local inbox files. They never execute shell commands.

## Usage alerts

CoAct does not scrape private provider accounts. A human, adapter, or agent can
record the quota data it already knows; CoAct evaluates it immediately and
alerts at 20% steps by default:

```sh
coact usage set --agent claude --model "Opus" --percent 42 --refresh-in 7d
coact usage set --agent codex --used 250000 --limit 1000000 --refresh "2026-07-17T00:00:00Z"
coact usage report
coact usage alerts
```

`coact usage report --watch` refreshes the local view. When a window is due,
the report asks for a new snapshot; CoAct does not poll a provider before then.
Crossed thresholds are journaled and sent to local workmate/human inboxes.

## Collaboration reports

Agents can rate each other after a run. Audit-derived facts and subjective peer
scores remain clearly separated:

```sh
coact eval rate --peer claude --model "Opus" --score 4 \
  --code-quality 5 --responsiveness 3 --note "Strong review; response was slow."
coact eval report run-20260710-120000
```

The report summarizes task completion, messages, lock issues, merge conflicts,
observed response delay, discrepancy handling, and peer-rated code quality.
`--watch` provides a continuously refreshed terminal report. Reports are local
decision support—not an objective model benchmark.

## Chinese expression diagnostics

The default-on, model-independent Chinese expression foundation detects Chinese
and mixed Chinese/English output, protects code, URLs, paths, and tables, and
falls back to raw text if validation fails:

```sh
echo '这个 feature 的 goal 是共享 memory，同时运行 `coact inbox`。' | coact zh check --diagnostics
echo '这是一个测试。' | coact zh check --off
```

The current release exposes detection/protection diagnostics and a Go adapter;
it does not automatically call a polishing model inside Claude/Codex/Gemini.

## What is ready?

See [Feature status](docs/FEATURES.md) for the reviewed matrix. In short:

- **Ready:** initialization, native launchers, shared memory, planning files,
  board ownership, inbox, locks/policy, audit log, worktrees, local usage alerts,
  collaboration reports, and Chinese diagnostics.
- **Experimental:** real-time Claude↔Codex bridge, local UI, and managed updates.
- **Not included:** autonomous agent wake-up, provider-account scraping, embedded
  terminals, automatic model switching, or a full autopilot.

## Safety model

- Coordination data stays under `.coact/`; sensitive runtime data is gitignored.
- Agent/run identifiers are path-safe; state writes are atomic and serialized
  where concurrent mutation matters.
- `.coact/config.json`, board internals, locks, inboxes, journals, terminal logs,
  usage snapshots, and evaluations are protected from direct agent rewrites.
- `coact doctor` checks wiring and runs an enforcement self-test.
- Hooks fail open for availability: CoAct is a guardrail, not a process sandbox.
- `coact update` is opt-in, HTTPS-only, and SHA-256 verified, but releases are
  not cryptographically signed yet.

Read [SECURITY.md](SECURITY.md) before relying on CoAct for high-assurance work.

## Command map

| Need | Command |
|---|---|
| Set up or verify | `coact init`, `coact doctor`, `coact deinit` |
| Launch agents | `coact claude`, `coact codex`, `coact gemini` |
| Plan | `coact plan`, `coact plan ready`, `coact plan status` |
| Own work | `coact board`, `task add`, `claim`, `done` |
| Coordinate | `coact @agent`, `@all`, `inbox`, `handoff` |
| Prevent overlap | `coact lock`, `unlock`, `policy`, `worktree`, `merge` |
| Observe | `coact`, `status`, `log` |
| Track quota | `coact usage set`, `report`, `alerts` |
| Review teamwork | `coact eval rate`, `report` |
| Diagnose Chinese output | `coact zh check` |
| Manage versions | `coact versions`, `update`, `switch` |

Run `coact help` for all flags. `coact ui`, `channel`, and `bridge` remain
optional experimental commands.

## Install from source

```sh
git clone https://github.com/tianyi-zhang-02/coact
cd coact
go build -o coact ./cmd/coact
```

MIT — see [LICENSE](LICENSE).
