# CoAct

[English](README.md) · **中文**

<img src="assets/mascot/icon.png" alt="CoAct 机器人宇航员 CoBot" width="140">

**让 Claude Code、Codex 和 Gemini 在同一个仓库协作，同时保留各自原生 terminal。**

CoAct 是本地的多 agent 协作与安全层：共享项目记忆、共同规划、任务归属、`@agent`
消息、写入意图锁、用量提醒、合作报告和审计记录。它不会替代 agent CLI，也不是模型
提供方。

## 两分钟开始

先把 CoAct 安装到 `PATH`，然后在每个项目初始化一次：

```sh
go install github.com/tianyi-zhang-02/coact/cmd/coact@main
cd your-project
coact init
coact doctor
```

每个 agent 使用一个原生 terminal：

```sh
# terminal 1
coact claude

# terminal 2
coact codex

# 可选 terminal 3
coact gemini
```

不需要额外开启第三个“管理 terminal”。你可以在任意 agent 对话里让它执行 CoAct
命令，也可以在普通 shell 里执行：

```sh
coact @codex "实现前先 review Claude 的 proposal。"
coact @all "读取新 brief，并分别提出方案。"
coact inbox
coact status
```

launcher 会设置 `COACT_AGENT`、`COACT_BIN` 和 `PATH`。即使你用
`/some/path/coact codex` 启动，agent 仍然可以直接运行 `coact inbox`。

`@main` 安装的是本文对应的已审查开发版本；发布对应 GitHub Release 后，再优先使用
tagged version。

## 日常工作流

### 1. 设置共享偏好

`coact init` 会创建两个由 human 控制的文件：

- `.coact/team.md`：agent 分工、planning 参与者和最终任务分配者
- `.coact/memory/project.md`：长期项目事实和偏好

不要在这些文件里放密钥或私人数据。

### 2. 一起规划

```sh
coact plan --with codex,claude --distributor codex "安全地重构 authentication"
coact plan status
```

每个 agent 会收到本地 inbox 消息，并在 `.coact/runs/<run>/` 下独立写 proposal。
distributor 等所有 proposal 都变成 `Status: ready` 且已解锁，再写 `final-plan.md`
并创建 board tasks。默认消息是 turn-based：空闲 agent 会在下一次 turn 读取；实验性的
real-time bridge 可以提供 mid-turn push。

proposal 写完后，agent 不需要手动改 metadata：

```sh
coact plan ready <run-id>
```

### 3. 不抢任务、不互相覆盖

```sh
coact board
coact claim T-001
coact lock internal/auth
# 修改和测试
coact unlock internal/auth
coact done T-001
```

board claim 会串行化。Claude Code 遇到冲突路径时会被 hook 硬拦截；Codex 和 Gemini
通过注入的 contract 自律，因此共享目录里的保护属于 advisory。需要更强物理隔离时，
使用 `coact <agent> --worktree`。

### 4. 消息、交接与审计

```sh
coact @claude "请 review T-001。"
coact handoff codex "parser 已完成；integration tests 还没做。"
coact log -n 50
```

消息只会写本地 inbox 文件，不会执行 shell 命令。

## 用量与配额提醒

CoAct 不抓取你的 provider 私有账户。human、adapter 或 agent 可以把已经知道的配额
数据写入 CoAct；系统会立即计算，并默认每 20% 提醒一次：

```sh
coact usage set --agent claude --model "Opus" --percent 42 --refresh-in 7d
coact usage set --agent codex --used 250000 --limit 1000000 --refresh "2026-07-17T00:00:00Z"
coact usage report
coact usage alerts
```

`coact usage report --watch` 会持续刷新本地状态。刷新时间到达前 CoAct 不轮询；到期后
report 会提示录入新 snapshot。跨过阈值时，会写 journal，并通知本地 human/workmate
inbox。

## 合作质量报告

一个 run 结束后，agent 可以互相评分。审计事实和主观评分会明确分开：

```sh
coact eval rate --peer claude --model "Opus" --score 4 \
  --code-quality 5 --responsiveness 3 --note "review 很扎实，但响应偏慢。"
coact eval report run-20260710-120000
```

报告汇总任务完成、消息、锁冲突、merge conflict、观测响应时间、discrepancy 处理和
互评 code quality。`--watch` 可持续刷新。这个报告用于 human 动态调整分工和模型，不是
客观 benchmark。

## 中文表达诊断

默认开启、模型无关的中文表达基础层可检测中文和中英混合文本，保护 code、URL、路径
和表格；校验不通过就回退原文：

```sh
echo '这个 feature 的 goal 是共享 memory，同时运行 `coact inbox`。' | coact zh check --diagnostics
echo '这是一个测试。' | coact zh check --off
```

当前版本提供 detection/protection 诊断和 Go adapter，但不会自动接管 Claude、Codex
或 Gemini 的输出，也不会自行调用润色模型。

## 哪些功能可以直接用？

完整审查结果见 [功能状态](docs/FEATURES.md)。简要结论：

- **Ready：**初始化、原生 launcher、共享记忆、planning files、board ownership、
  inbox、locks/policy、audit log、worktrees、本地用量提醒、合作报告和中文诊断。
- **Experimental：**Claude↔Codex real-time bridge、本地 UI、managed updates。
- **尚未包含：**自动唤醒空闲 agent、抓取 provider 私有账户、嵌入式 terminal、自动
  模型切换或 full autopilot。

## 安全模型

- 协作数据位于 `.coact/`；敏感 runtime 数据默认 gitignore。
- agent/run ID 有严格路径校验；状态使用 atomic write，并在并发 mutation 处加锁。
- config、board 内部状态、locks、inbox、journal、terminal logs、usage 和 evaluations
  都禁止 agent 直接改写。
- `coact doctor` 会检查接线，并运行 enforcement self-test。
- hook 失败开放以保证可用性；CoAct 是 guardrail，不是 process sandbox。
- `coact update` 只在用户主动调用时联网，使用 HTTPS + SHA-256；release 尚未签名。

高安全需求请先阅读 [SECURITY.md](SECURITY.md)。

## 命令地图

| 需求 | 命令 |
|---|---|
| 初始化/检查 | `coact init`, `coact doctor`, `coact deinit` |
| 启动 agent | `coact claude`, `coact codex`, `coact gemini` |
| 一起规划 | `coact plan`, `coact plan ready`, `coact plan status` |
| 管理任务 | `coact board`, `task add`, `claim`, `done` |
| 协作沟通 | `coact @agent`, `@all`, `inbox`, `handoff` |
| 防止覆盖 | `coact lock`, `unlock`, `policy`, `worktree`, `merge` |
| 查看状态 | `coact`, `status`, `log` |
| 配额提醒 | `coact usage set`, `report`, `alerts` |
| 合作复盘 | `coact eval rate`, `report` |
| 中文诊断 | `coact zh check` |
| 版本管理 | `coact versions`, `update`, `switch` |

完整参数见 `coact help`。`coact ui`、`channel` 和 `bridge` 保留为可选实验功能。

## 从源码安装

```sh
git clone https://github.com/tianyi-zhang-02/coact
cd coact
go build -o coact ./cmd/coact
```

MIT —— 见 [LICENSE](LICENSE)。
