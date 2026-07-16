# Contributing to CoAct

Thank you for helping make multi-agent coding safer and easier to use. Small,
focused pull requests are easiest to review. For a large feature or behavior
change, please open a feature request before investing significant time.

[中文说明](#中文贡献说明)

## Before you start

- Search existing issues and pull requests to avoid duplicate work.
- Use a fork and create a branch such as `fix/windows-paths` or
  `feat/agent-adapter`.
- Never include credentials, raw agent transcripts, `.coact/` runtime state, or
  proprietary project content in a report or commit.
- Report vulnerabilities through a private GitHub security advisory as
  described in [SECURITY.md](SECURITY.md), not through a public issue.
- New visual assets must be original or have a license compatible with MIT.

## Development setup

CoAct is a dependency-free Go project. Go 1.22 or newer is recommended.

```sh
git clone https://github.com/YOUR-USERNAME/coact.git
cd coact
go build ./cmd/coact
go test ./...
go vet ./...
```

To exercise the local coordination wiring without a second agent:

```sh
go build -o /tmp/coact ./cmd/coact
/tmp/coact init
/tmp/coact doctor
```

Use a disposable repository for initialization tests. `coact init` intentionally
writes project contracts, local coordination state, and a Claude hook.

## Pull request expectations

- Keep the change scoped to one problem.
- Add or update tests for behavior changes.
- Run `gofmt` on modified Go files.
- Ensure `go test ./...`, `go vet ./...`, and `go build ./cmd/coact` pass.
- Update both `README.md` and `README.zh-CN.md` when user-facing behavior changes.
- Describe security, compatibility, migration, and rollback implications.
- Disclose meaningful AI assistance and verify generated code and assets before
  submission. Contributors remain responsible for correctness and provenance.

The `main` branch requires all Linux, macOS, Windows, and cross-compilation CI
checks to pass. External pull requests also require a code-owner review and all
review conversations to be resolved. Force pushes and branch deletion are
disabled on `main`.

## Project principles

Changes should preserve these invariants:

1. **Local first:** coordination state remains local unless the user explicitly
   requests a network operation.
2. **No arbitrary shell API:** user or browser input must not become shell code.
3. **Safe coexistence:** ownership, locks, and audit records prevent agents from
   silently overwriting one another.
4. **Honest capability labels:** experimental features must not be documented as
   fully autonomous or production hardened.
5. **Native-agent independence:** Codex, Claude Code, and other agents retain
   their own authentication, permissions, model controls, and terminals.

## Adding an agent adapter

Keep adapters declarative where possible. Document the executable name,
supported platforms, contract file, hook/enforcement level, and failure mode.
An unavailable visualizer or coordination channel must never block the native
agent from starting.

## 中文贡献说明

感谢你参与 CoAct。为了方便 review，请优先提交范围明确、可以独立验证的小 PR；如果是
大型功能或行为调整，建议先开 Feature Request 对齐方向。

### 开始之前

- 先搜索现有 Issues 和 Pull Requests，避免重复开发。
- Fork 仓库后创建独立分支，例如 `fix/windows-paths`。
- 不要提交 API key、登录信息、原始 agent transcript、`.coact/` runtime state 或
  私有项目内容。
- 安全漏洞请按照 [SECURITY.md](SECURITY.md) 提交 private security advisory。
- 新增视觉素材必须是原创，或拥有与 MIT 兼容的授权。

### 提交前检查

```sh
gofmt -w <修改过的 Go 文件>
go test ./...
go vet ./...
go build ./cmd/coact
```

- 行为变化需要添加或更新测试。
- 面向用户的变化需要同步更新中英文 README。
- PR 中请说明安全、兼容性、迁移和回滚影响。
- 如果大量使用 AI 生成代码或素材，请在 PR 中说明，并自行确认正确性与来源授权。

`main` 必须通过 Linux、macOS、Windows 和 cross-compile 检查；外部 PR 需要
CODEOWNER review，并解决所有 review conversation。请勿在 PR 中加入与目标无关的
重构。
