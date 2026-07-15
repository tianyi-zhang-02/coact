# Feature status

Reviewed: 2026-07-14

Release baseline: `v1.0.0`

“Ready” means the feature has an implemented CLI path, automated tests, a
documented safety boundary, and a graceful failure mode. It does not mean every
agent adapter has identical enforcement strength.

| Feature | Status | Ready to use? | Important limit |
|---|---|---:|---|
| `init` / `doctor` / contract migration | Ready | Yes | Re-run `coact init` after upgrading to refresh generated contracts and config migrations. |
| Claude/Codex/Antigravity native launchers | Ready | Yes | The underlying agent CLI must already be installed and authenticated. |
| Shared team policy and project memory | Ready | Yes | Human-controlled local context; never store secrets. |
| Planning runs and final distributor | Ready | Yes | `plan finalize` validates distributor/readiness/locks and creates assigned board tasks; agents still need a turn to read inbox and CoAct does not autonomously wake them. |
| Board add/assign/claim/done, reopen, and handoff | Ready | Yes | Strict `todo → claimed → doing → done` transitions are coordinated locally, not synced to an external issue tracker. |
| Inbox and `@agent` / `@all` | Ready (turn-based) | Yes | `@all` targets live workspace agents; offline direct messages remain available with `@agent`. |
| Locks and policy | Ready, asymmetric | Yes | Claude is L2 hook-enforced; Codex/Antigravity are L1 contract-enforced in shared mode. |
| Worktree isolation and merge gate | Ready (basic) | Yes | Merge conflicts stop for human resolution; no automatic conflict solver. |
| Presence and audit journal | Ready | Yes | Presence is process/heartbeat based; journal is local and gitignored. |
| Usage windows and 20% alerts | Ready (provider-independent) | Yes | CoAct does not scrape provider accounts; a human/adapter supplies snapshots. |
| Collaboration ratings and reports | Ready (decision support) | Yes | Audit metrics are factual; quality scores are subjective and labeled as such. |
| Chinese expression adapter | Ready as foundation | Diagnostics/API | Not automatically inserted into provider output pipelines. |
| Managed update/switch | Experimental | Test before relying | HTTPS + SHA-256, but release metadata is not cryptographically signed. |
| Claude↔Codex real-time bridge | Experimental | Dogfood only | Provider versions/protocols can change; turn-based inbox is the stable fallback. |
| Local web UI | Beta/optional | Yes, for oversight | Includes task filters, readable output-only terminal snapshots, task-completion effects, and human-approved help/handoff suggestions; terminal-native CLI remains primary. |
| Embedded terminals/model switching/autopilot | Not included | No | Deliberately outside the current release. |

## Recommended first-use check

```sh
coact init
coact doctor
coact status
coact plan --with codex,claude --distributor human "Small dogfood task"
# after proposals are ready and final-plan.md lists execution tasks:
coact plan finalize --agent human
```

Then verify one agent can claim a task, a second agent sees it, a conflicting
Claude edit is blocked, and `coact log` records the sequence.
