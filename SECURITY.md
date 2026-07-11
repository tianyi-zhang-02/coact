# Security model

CoAct is itself a security-adjacent tool: it wires a hook into your Claude Code
settings and runs on every edit. This document states what it does, what it
guarantees, and what it deliberately does not.

## What CoAct is (and isn't)

CoAct coordinates **semi-trusted** agents — it defends against agents making
mistakes, wandering into each other's files, or acting on a prompt injection. It
is **not a sandbox** and does not defend against a determined local attacker who
already controls your account, your `$PATH`, or the coact binary itself. Agents
still run with whatever permissions you give them; CoAct adds a coordination and
audit layer, it does not confine the processes.

## Trust boundary — what CoAct touches

Running `coact init` writes only inside the repository:

- `.coact/` — all coordination state (locks, board, session, journal, config).
- `.claude/settings.json` — merged (existing keys preserved) to add one
  `PreToolUse` hook.
- `CLAUDE.md`, `AGENTS.md` — a marked contract block appended.
- `.gitignore` — coact runtime-state entries appended.

Nothing else is modified. `coact deinit` removes all of it; `--purge` also
removes `.coact/` (and refuses to `RemoveAll` any directory not named `.coact`).

## The enforcement hook

- The hook command is the **coact binary itself**, pinned by absolute path, with
  fixed args `hook claude`. `coact doctor` prints the exact wired command so you
  can verify it.
- It **fails open**: on any error, or in a repo without coact, it returns exit 0
  and the edit proceeds. A coact bug can never brick your editing. The tradeoff
  is that enforcement is not *guaranteed* under coact failure — CoAct chooses
  availability over hard enforcement.
- It does **not** require `--dangerously-skip-permissions`. It runs alongside
  Claude Code's normal permission flow and only ever *adds* a denial.

## Enforcement guarantees and limits

- **Claude Code (L2)** — edits are hard-blocked via the hook (exit 2).
- **Codex (L1)** — no pre-write hook exists, so it self-enforces from the
  contract in `AGENTS.md`. This is advisory: a non-cooperating Codex session is
  not mechanically stopped.
- A pool is only as safe as its weakest **live** participant. Mixing an L1 agent
  with an L2 agent on a shared tree lowers the shared-resource guarantee to L1;
  worktree isolation (roadmap) is the answer as lower-tier agents join.

## Optional local control center (`coact ui`)

The default `coact` command prints a terminal workspace summary. The optional
`coact ui` command starts a **local-only** web control center. Because a browser
can reach any localhost server, binding to loopback is
necessary but not sufficient; the server adds two defenses:

- **Loopback bind + Host-header allowlist.** It listens only on
  `127.0.0.1`/`localhost` and rejects any request whose `Host` header is not a
  loopback name. This defeats DNS-rebinding, where a malicious page resolves its
  own hostname to `127.0.0.1` — the rebound request still carries the attacker's
  Host and is refused.
- **Per-run CSRF token.** A random token is minted at startup, embedded in the
  served page, and required on every mutating request. A cross-origin page cannot
  read the token out of the same-origin HTML, so it cannot forge the header.
  Read-only GETs are exempt: the same-origin policy already protects their
  bodies, and they change nothing.

The UI accepts **no arbitrary shell input**. Mutations reuse governed CLI
primitives and are journaled; the macOS-only launch action can start only a
built-in, installed agent adapter using fixed argument construction. The server
refreshes by polling and opens no outbound connections.

Terminal transcripts can contain prompts, source code, or accidentally printed
credentials. CoAct stores them only under the gitignored `.coact/terminal/`
directory. Treat that directory as sensitive local data and delete it before
sharing a workspace archive.

## Usage and collaboration reports

Quota snapshots and peer ratings are local decision-support data under
`.coact/usage/` and `.coact/evaluations/`. Both directories are gitignored,
written with owner-only file permissions where supported, path-safe, and
protected from direct agent edits by policy.

- CoAct does not log in to providers or scrape private account pages. A human or
  adapter explicitly supplies usage values and refresh times.
- Threshold notifications contain percentages, model labels, and refresh times,
  not credentials or raw prompts.
- Collaboration reports label journal-derived facts separately from subjective
  peer ratings. They must not be treated as an objective model benchmark.
- Rating notes are length-limited but human-authored; do not put secrets in them.

## Managed updates (`coact update`)

`coact update` / `switch` / `versions` are **experimental** and manage
side-by-side binaries under `~/.coact` — the only place CoAct writes outside the
repository.

- Releases are fetched from the pinned GitHub repository over **HTTPS** and
  verified against a **SHA-256** checksum before install.
- Archive extraction is **not** controlled by the archive's own paths: it matches
  the binary by base name (`coact`/`coact.exe`) and writes to a fixed
  `~/.coact/bin/coact-<version>` path, so a hostile archive cannot escape via
  `../` entries ("zip-slip").
- Manifest, archive, and extracted-binary sizes are bounded to reduce memory,
  disk-exhaustion, and decompression-bomb risk; redirects must remain HTTPS.
- It **never overwrites a system install** — it only re-points the managed
  `~/.coact/coact` shim, and only when the running binary is itself managed.
- Checksums verify **integrity, not authenticity**: there is not yet a
  cryptographic signature, so `coact update` trusts GitHub + TLS. Treat it as a
  convenience until release signing lands; prefer your package manager or a
  verified download for high-assurance installs.

## Design safeguards

- **No shell execution.** CoAct never invokes `sh -c`/`bash -c`; subprocesses
  (`git`, the agent binary) use `exec.Command` with fixed or pass-through args —
  no command-injection surface.
- **No network in the coordination core.** The core makes no network calls; the
  sole exception is the opt-in `coact update` above, which is never invoked
  automatically.
- **Path containment.** Lockable paths are normalized and rejected if they
  escape the repo root; lock files are content-hashed, not path-named; agent
  identifiers are restricted to `[a-z0-9_-]` so they cannot traverse out of
  `.coact/session/`; every `Remove`/`RemoveAll` is bounded to `.coact/`.
- **Atomic, auditable state.** All state is written atomically (temp + rename)
  and is plain text; every significant action is appended to the journal.
- **Concurrent local reports.** Usage snapshots and peer ratings use local
  meta-locks; shared threshold history is separately serialized.

## Auditing and removing

- `coact doctor` — shows the wired hook command and self-tests the engine.
- `coact log` / `coact status` — the full audit trail and live state.
- `coact deinit` — removes every trace of CoAct's wiring.

## Reporting a vulnerability

Please open a private security advisory on the GitHub repository (Security →
Report a vulnerability) rather than a public issue. Include a minimal
reproduction and the coact version (`coact version`).
