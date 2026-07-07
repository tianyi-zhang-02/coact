# Coact Protocol Specification (v0.1 draft)

Status: draft · Last updated: 2026-07-06

Coact is a **filesystem-based coordination protocol** that lets two or more
autonomous coding agents (e.g. Claude Code, Codex) operate concurrently in the
**same working directory** without corrupting each other's work.

The protocol is deliberately implementation-agnostic: it defines only files,
their formats, and the state machines that govern them. Any process — an agent,
a git hook, a human running the CLI — that follows these rules is a conformant
participant.

---

## Implementation status (v0.1)

This spec is the design target. What ships today:

| Spec area | Status | Code |
|---|---|---|
| §1 Directory layout | built | `internal/project` |
| §2 Lock protocol (registry-serialized acquire, TTL, presence-gated steal) | built | `internal/lockmgr`, `internal/metalock` |
| §3 Task board | built | `internal/board` |
| §4 Messaging (inbox) | planned | `.coact/inbox/` reserved, not wired |
| §5 Presence & heartbeat | built | `internal/presence` (+ sidecar and launcher) |
| §6 Journal | built | `internal/journal` |
| §7 Config | built | `internal/config` |
| §8 Enforcement L0/L1/L2 | built | Claude L2 hook, Codex L1 contract, `internal/policy` capability gating |
| §9 Worktree mode | built (basic) | `coact worktree` + `coact merge`, shared-state resolution |

Where the implementation currently diverges from the text below:

- **Config is `.coact/config.json` (JSON), not the YAML shown in §7** — to keep
  the binary dependency-free (see [STACK.md](STACK.md)).
- **The L2 block (§8) is delivered as exit code 2 + stderr**, which every Claude
  Code version honors, rather than the newer `permissionDecision` JSON.
- The optional broker (§0/§7) is not built.

For the component map and diagrams, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## 0. Design principles

1. **The filesystem is the only shared medium.** No *central* daemon or broker
   is required for correctness — every coordination primitive is a file whose
   creation/rename is atomic on POSIX. A per-session presence sidecar (§2.9)
   may run for fast crash detection, but it holds no shared authority and is
   optional (`presence.mode: hook-only` removes it). An optional central broker
   (§7) can layer on real-time push, but nothing in this spec depends on it.
2. **Advisory, not mandatory.** Locks are cooperative. Enforcement strength is a
   property of the *adapter* (Claude's hooks can hard-block; Codex mostly
   self-reports). The protocol assumes participants are mostly honest but may
   crash.
3. **Crash-safe by default.** Every claim has a TTL and a heartbeat. A dead
   participant's locks and tasks are reclaimable without manual cleanup.
4. **Human-auditable.** Every state file is plain text/JSON. The journal is an
   append-only replayable log — the substrate for human oversight.

---

## 1. Directory layout

All coordination state lives under `.coact/` at the repository root.

```
.coact/
  config.yaml            # static configuration (committed to git)
  board.md               # shared task board (committed)
  locks/                 # advisory locks (gitignored)
    <path-hash>.lock
  session/               # presence + heartbeat (gitignored)
    <agent>.json
  inbox/                 # inter-agent messages (gitignored or committed)
    <agent>.md
  journal/               # append-only activity log (gitignored)
    YYYY-MM-DD.jsonl
```

`config.yaml` and `board.md` are committed (shared intent). `locks/`,
`session/`, `journal/` are runtime state and go in `.gitignore`.

### 1.1 Identity

Each participant has a stable `agent_id` (lowercase, `[a-z0-9_-]+`), e.g.
`claude`, `codex`. It is passed via `--agent` or the `COACT_AGENT` env var.
Two live processes MUST NOT share an `agent_id`.

### 1.2 Path hashing

A lock filename is `sha256(normalized_path)[:12] + ".lock"`, where
`normalized_path` is the repo-relative POSIX path with no trailing slash. The
full path is also stored *inside* the lock file (the hash is only for a flat,
collision-resistant filename).

---

## 2. Lock protocol

### 2.1 Purpose

A lock is an **intent to write** a file or subtree. Reads never require a lock.

### 2.2 Lock file format

`.coact/locks/<path-hash>.lock` — JSON, single line:

```json
{
  "path": "src/api/handler.ts",
  "mode": "exclusive",
  "owner": "claude",
  "pid": 48213,
  "acquired_at": "2026-07-06T14:03:21Z",
  "ttl_seconds": 900,
  "heartbeat_at": "2026-07-06T14:07:50Z"
}
```

- `mode`: `exclusive` (default) or `shared`. `shared` locks are advisory
  read-intent hints; multiple `shared` locks on one path may coexist but a
  `shared` and an `exclusive` on the same path conflict. v0.1 implementations
  MAY treat every lock as `exclusive`.
- A lock covers the path **and everything beneath it** if the path is a
  directory. Conflict check is prefix-based (§2.5).

### 2.3 Acquire (serialized under the registry meta-lock)

`O_EXCL` guarantees atomicity only for an *exact filename*, not for the
subtree (prefix) relationship between paths. A naive "create-then-scan"
approach has a symmetric TOCTOU race: two agents locking overlapping paths
(e.g. `src/` and `src/api/handler.ts`) each create their own distinct lock
file, then each scan sees the other and both roll back — a livelock.

To close this, **all acquisitions are serialized under a single registry
meta-lock**, `.coact/locks/.registry.lock`. It is a leaf path with no subtree
semantics, so acquiring *it* is race-free via plain `O_EXCL`.

To acquire a lock on `P`:

1. Acquire the registry meta-lock `.coact/locks/.registry.lock`
   (`O_CREAT | O_EXCL`, own short TTL ~5s, spin-retry with backoff). This
   serializes the critical section below.
2. Run the **overlap scan** (§2.5) over all live locks:
   - If a **live** conflicting lock exists and `owner != self` → release the
     meta-lock, return `DENIED` with the owner's info.
   - If a conflicting lock exists but is **reclaimable** (§2.4) → steal it
     (§2.6) and continue.
   - If a conflicting lock is `owner == self` → this is reentrant; refresh and
     continue.
3. Compute `H = hash(P)`, write `.coact/locks/<H>.lock` (temp + atomic rename),
   fsync.
4. Release the registry meta-lock. Return `ACQUIRED`.

Because steps 2–3 run inside the meta-lock, no other participant can observe or
create a conflicting lock mid-check. The meta-lock's own short TTL bounds the
blast radius if a holder crashes inside the critical section.

> Contention note: this serializes *lock acquisition*, not file editing. For 2
> agents the critical section is sub-millisecond; the meta-lock is not a
> throughput concern. Revisit only for large N (see §10).

### 2.4 Expiry and reclaimability

A lock is **expired** at read time if:

```
now > max(acquired_at, heartbeat_at) + ttl_seconds
```

Expiry alone does **not** authorize a steal. Lock TTL and liveness are
decoupled (§2.9): a lock is **reclaimable** only if it is *both* expired *and*
its owner is **not live** per presence (§5):

```
reclaimable(lock) = expired(lock) AND NOT presence_live(lock.owner)
```

This prevents stealing a lock from an owner that is alive but momentarily
silent — e.g. mid-way through a long build or a long model-reasoning turn where
no lock heartbeat fired. An expired-but-live lock is left untouched; the
acquirer returns `DENIED` (or queues, per policy).

### 2.5 Overlap (conflict) scan

Two paths conflict if one is a prefix of the other (path-segment aware):
`src/api` conflicts with `src/api/handler.ts` but not with `src/apidocs`.
Because filenames are hashes, the scan reads every live lock's `path` field and
compares. For repos with many locks, an implementation MAY maintain an index
file `.coact/locks/index.jsonl`; the flat scan is the normative fallback.

### 2.6 Steal

A steal is only permitted on a **reclaimable** lock (§2.4) and only while
holding the registry meta-lock (§2.3). Stealing = overwrite the lock file's
contents with the new owner's record via **write-to-temp + atomic rename**
(`.coact/locks/<H>.lock.<rand>` → rename over target). The steal MUST be
journaled with event `lock.stolen` including the previous owner and the reason
(`ttl_expired_owner_dead`). Adapters SHOULD notify the victim via its inbox
(§4) in case it later reconnects.

### 2.7 Heartbeat & release

- While holding a lock, the owner MUST rewrite `heartbeat_at` at least every
  `ttl_seconds / 3`. The CLI `coact heartbeat` (or a background refresher)
  handles this.
- Release = delete the lock file. Releasing a lock you don't own is a protocol
  violation (journaled as `lock.violation`), except when stealing an expired
  lock.

### 2.8 Lock state machine

```
          acquire (meta-lock, no conflict)      release
   FREE ──────────────────────────────────► HELD ────────► FREE
     ▲        acquire on reclaimable lock      │
     │        = steal (expired AND owner dead)  │ heartbeat
     └──────────────────◄──────────────────────┘ (self-loop, refresh)
```

### 2.9 Liveness vs lock TTL (the no-daemon constraint)

Agents like Claude Code are request/response processes: **no participant code
runs between tool calls** while the model reasons, and a single long-running
tool call (e.g. a 10-minute build) fires no lock heartbeat for its duration.
Coupling crash-detection to lock TTL would therefore force an impossible
tradeoff — short TTL steals live agents' locks mid-work, long TTL leaves dead
agents' locks stuck.

The protocol splits the two concerns:

- **Lock TTL** (default 900s, §2.4) is a coarse safety net, deliberately long.
- **Presence heartbeat** (§5, default every 20s, `presence_ttl` 60s) is a
  cheap, frequent liveness signal written by a **per-session presence
  sidecar** — a lightweight process the adapter starts alongside the agent and
  that dies with the session. It only rewrites `session/<agent>.json`; it never
  touches locks. This is *not* the central broker deliberated in §0 — it is
  per-agent and holds no shared authority.

A lock is stolen only when it is expired **and** its owner's presence is stale
(§2.4). Result: a live-but-silent owner (long build, long reasoning turn)
keeps its locks; a truly crashed owner is detected within `presence_ttl` and
its locks become reclaimable immediately, without waiting out the full lock
TTL.

> Fallback for environments that forbid any resident process: run in
> **hook-only mode** — refresh both lock and presence heartbeats on every
> `PreToolUse`/`PostToolUse`, and raise lock TTL to 15–30 min. This tolerates
> reasoning gaps at the cost of slower crash detection. Selectable via
> `presence.mode: sidecar | hook-only` in config.

---

## 3. Task board protocol

### 3.1 Format

`.coact/board.md` is human-first Markdown with machine-parseable task lines. A
task is a list item with an HTML-comment metadata tag:

```markdown
## Backlog
- [ ] Add rate limiting to the API gateway <!-- coact: id=T-014 state=todo owner= -->
- [~] Refactor auth middleware <!-- coact: id=T-011 state=doing owner=claude ttl=1800 hb=2026-07-06T14:07Z -->
- [x] Write OpenAPI schema <!-- coact: id=T-009 state=done owner=codex -->
```

Checkbox glyph mirrors state for humans: `[ ]` todo, `[~]` claimed/doing,
`[x]` done, `[!]` blocked/review.

### 3.2 Task states

```
todo ──claim──► claimed ──start──► doing ──finish──► review ──accept──► done
  ▲                │                  │                                  
  └────release/────┴───── TTL expiry ─┘                                  
     TTL expiry
```

- `claimed` vs `doing`: `claimed` = owner reserved it; `doing` = actively
  working (lock(s) held). Implementations MAY collapse the two.
- A task in `claimed`/`doing`/`review` carries `owner`, `ttl`, `hb` (heartbeat).
  On TTL expiry with no heartbeat, it reverts to `todo` (owner cleared) so the
  other agent can pick it up.

### 3.3 Atomic claim

`board.md` is a single file, so claim is done via **read-modify-write under a
meta-lock**: acquire the reserved lock path `.coact/board` (§2) before editing
the board, then release. This serializes all board mutations. Claiming a task
whose `owner` is already set (and not expired) fails.

---

## 4. Messaging protocol

`.coact/inbox/<recipient>.md` — append-only. A message is a fenced block:

```markdown
--- <!-- coact-msg: from=claude to=codex ts=2026-07-06T14:09:03Z id=M-31 -->
Reworked `AuthContext`; the `token` field is now `accessToken`. If you're
touching `session/`, pull first. Lock on `src/auth/` released.
```

- Append is done by opening the file `O_APPEND` and writing the whole block in
  one `write()` (atomic for small writes on POSIX). For safety across large
  messages, use the `.coact/inbox/<recipient>` meta-lock.
- A recipient marks messages read by moving them to
  `.coact/inbox/<recipient>.read.md`, or an adapter injects unread messages into
  the agent's context at turn start and truncates.
- `coact msg <to> <text>` is the producer; `coact inbox [--unread]` the consumer.

---

## 5. Presence & heartbeat

`.coact/session/<agent>.json`, rewritten atomically (temp+rename) on each beat:

```json
{
  "agent": "codex",
  "pid": 51002,
  "cwd": "/repo",
  "current_task": "T-014",
  "held_locks": ["src/gateway/"],
  "started_at": "2026-07-06T13:40:00Z",
  "heartbeat_at": "2026-07-06T14:08:12Z",
  "status": "working"
}
```

`status` ∈ `working | idle | waiting | stopped`. The file is written by the
per-session **presence sidecar** (§2.9) every `presence.interval_seconds`
(default 20s), or by the adapter's hooks in `hook-only` mode.

A participant is **live** (`presence_live`) if
`now - heartbeat_at < presence_ttl` (default 60s). Implementations SHOULD also
treat a participant as dead if `pid` is present and no longer running on the
same host, tightening detection below `presence_ttl`. `presence_live` is the
sole authority for whether a lock may be stolen (§2.4). `coact status` renders
board + live participants + active locks from these files alone.

---

## 6. Journal

`.coact/journal/YYYY-MM-DD.jsonl` — append-only, one JSON event per line. Every
protocol-significant action emits an event:

```json
{"ts":"2026-07-06T14:03:21Z","agent":"claude","event":"lock.acquire","path":"src/api/handler.ts"}
{"ts":"2026-07-06T14:03:59Z","agent":"codex","event":"lock.denied","path":"src/api/handler.ts","held_by":"claude"}
{"ts":"2026-07-06T14:07:50Z","agent":"claude","event":"task.finish","id":"T-011"}
{"ts":"2026-07-06T14:08:40Z","agent":"claude","event":"lock.stolen","path":"src/gateway/","from":"codex","reason":"ttl_expired"}
```

Event vocabulary (v0.1): `lock.acquire`, `lock.release`, `lock.denied`,
`lock.stolen`, `lock.violation`, `task.claim`, `task.start`, `task.finish`,
`task.accept`, `task.revert`, `msg.send`, `session.start`, `session.stop`.

The journal is the audit substrate: it is sufficient to reconstruct who held
what, who was blocked by whom, and where a conflict originated.

---

## 7. config.yaml

```yaml
version: 0.1
mode: shared            # shared | worktree
agents:
  - id: claude
    adapter: claude-code
  - id: codex
    adapter: codex
locks:
  default_ttl_seconds: 900      # coarse safety net, not liveness (see §2.9)
  heartbeat_divisor: 3          # refresh every ttl/divisor
  registry_lock_ttl_seconds: 5  # meta-lock guarding acquisition (§2.3)
presence:
  mode: sidecar                 # sidecar | hook-only (§2.9)
  ttl_seconds: 60               # liveness window; sole steal authority
  interval_seconds: 20          # sidecar beat cadence
broker:
  enabled: false                # optional central mediator; protocol stays
                                # daemon-agnostic — nothing above requires it
policy:
  on_conflict: block            # block | queue | warn
  allow_steal_expired: true
  protected_paths:              # never auto-lockable; require human
    - ".coact/config.yaml"
    - ".github/"
```

---

## 8. Conformance & enforcement levels

The protocol distinguishes what a participant *claims* to do from what its
adapter can *enforce*:

| Level | Guarantee | Example |
|-------|-----------|---------|
| **L0 report** | Participant journals actions after the fact | git pre-commit hook |
| **L1 self-check** | Participant queries lock before writing, obeys result voluntarily | Codex following `AGENTS.md` |
| **L2 gated** | A hook blocks the write syscall/tool-call on `DENIED` | Claude Code `PreToolUse` hook |

A pair is only as safe as its weakest live participant. The dashboard MUST
surface each participant's enforcement level so the human knows the real
guarantee.

---

## 9. Worktree mode (summary, spec'd in v0.2)

In `mode: worktree`, each agent operates in its own `git worktree` on its own
branch. `.coact/` lives in the shared common dir (`git rev-parse --git-common-dir`)
so the board/journal/inbox remain global, while file edits are physically
isolated per branch. Coordination shifts from live locks to:
- board-level task ownership (still via §3), and
- `coact merge` orchestrating integration (rebase/merge + conflict surfacing).

Full worktree-mode spec is deferred to v0.2.

---

## 10. Open questions

- **Sub-file locking.** v0.1 locks whole files. Function/range-level locks would
  cut false conflicts but need language awareness. Deferred.
- **Codex hard enforcement.** Without a native pre-write hook, can we interpose
  at the filesystem (FUSE / `LD_PRELOAD`) portably? Probably not cross-platform;
  L1 is the realistic ceiling today.
- **Board as bottleneck.** The single-file meta-lock serializes all board
  writes. Fine for 2 agents; revisit for N.
```