# aep-parser

从头实现的 Adobe After Effects `.aep` 项目文件解析器 + length-preserving 写回库，纯 Go，无需 AE 运行实例。

> ⚠️ **开发中（WIP）** —— 功能仍在扩展，详细能力清单待项目成熟后再补。
> 当前 API 以自动生成的 [docs/](docs/) 为准。

**兼容下限：After Effects 2020（CC 17.0）**。新版本写的 .aep 也能读；AE 24+ 才引入的字段本库不主动解码，对调用方返回 nil 而非报错。

## 文档

- **API 参考**：[docs/](docs/) —— 每个核心类型一个 markdown，由 `cmd/docgen` 从导出符号的 doc comment **自动生成**（`go generate ./cmd/docgen`）。
- **能力覆盖矩阵 / 暂搁 / 不可达 / negative findings**：[docs/capabilities.md](docs/capabilities.md)。

## 跨平台构建验证

解析、profile、diff、search、recipe、technique report 和自托管编排的 Go 入口应保持跨平台可构建；AE 自动化是可选 worker 能力，Linux/macOS 服务节点不应因为没有 AE 或 PowerShell 而阻断纯解析工作。

```powershell
go run ./cmd/aepverify cross-platform
```

Recipe 示例的 expected profile 覆盖也走同一个 Go 验证入口：

```powershell
go run ./cmd/aepverify recipe-profiles
```

## 纯解析服务

`cmd/aepserver` 提供不依赖 AE 的 HTTP 服务入口，适合放在 Windows、macOS 或 Linux 节点上做解析、profile 和能力探测：

```powershell
go run ./cmd/aepserver -addr 127.0.0.1:8080
```

当前 endpoint：

- `GET /health`
- `GET /capabilities`
- `POST /parse`
- `POST /profile`

`/parse` 和 `/profile` 默认支持 `application/octet-stream` 直接上传 `.aep` 字节；上传时可用 `X-AEP-Path` 标记来源路径。本地批处理需要直接读取服务端路径时，必须同时指定 `-allow-path-input -allowed-path-roots <root1,root2>`，再发送 `application/json` 的 `{"path":"data/samples/.../file.aep"}`；服务会拒绝目录逃逸和 symlink 逃逸。服务默认限制 32 个并发请求并配置 HTTP 读写超时，不会启动 AE，`render` / `ae_readback` 在 capabilities 中会报告为 unavailable。

## 自托管技法报告

对本地 `.aep` 语料生成可浏览的 technique learning report：

```powershell
go run ./cmd/aepselfhost technique-report -input data\samples -out tmp\technique_samples_report -verify
```

输出包括 `manifest.json`、`learning.md`、`report.html`、`projects.csv`、`project_playbooks.csv`、`compositions.csv`、`layers.csv`、`recreation_steps.csv`、`patterns.csv`、`study_queue.csv`、`study_tasks.csv`、`recreation_blockers.csv`、`signal_layers.csv`、`effect_stacks.csv`、`shape_operators.csv`、`text_animators.csv`、`dependency_edges.csv`、`learning_actions.csv`、`mechanisms.csv`、`mechanism_examples.csv`、`coverage_scorecard.csv`、`reconstruction_blueprints.jsonl`、`recipe_drafts.jsonl`、`errors.csv`、`digest.json`、`summary.json`、`corpus.jsonl` 和 `report.md`。`report.html` 是自包含入口，支持按路径、readiness、pattern、effect、plugin、mechanism、study task 搜索，并展示每个项目的确定性复刻步骤、artifact 覆盖状态、复刻蓝图和安全 recipe 草稿；`manifest.json` 记录输入、git 版本、耗时和 artifact 清单；`learning.md` 是更短的人工学习索引；CSV/JSONL 文件适合直接用表格或流式脚本筛项目、项目级复刻 playbook、逐项复刻步骤、合成/图层结构、逐层 effect stack、shape/text 操作项、依赖边、技法模式、学习顺序、学习任务、复刻阻塞项、关键层、学习动作、机制目录、机制代表项目、覆盖计分卡、代码生成蓝图、recipe 骨架草稿、错误项目和 pattern 级复刻步骤分布。

对比两次报告：

```powershell
go run ./cmd/aepselfhost compare-reports -base tmp\old_report -new tmp\technique_samples_report
```

一键跑完整自托管验收：

```powershell
go run ./cmd/aepselfhost verify -out-root tmp\technique_selfhost_gate
go run ./cmd/aepselfhost outcome -out-root tmp\technique_selfhost_gate
go run ./cmd/aepselfhost status -out-root tmp\technique_selfhost_gate
go run ./cmd/aepselfhost watch -out-root tmp\technique_selfhost_gate -duration-minutes 60 -interval-seconds 300
go run ./cmd/aepselfhost start-watch -out-root tmp\technique_selfhost_gate -duration-minutes 60 -interval-seconds 300
```

想跑完后立刻查看最短成效页，可以加 `-open`，它会打开 `latest_outcome.html`。Go CLI 是唯一维护入口：`aepselfhost verify` 跑验收，`aepselfhost status` 看进程、watch 状态、最新 outcome 和日志路径，`aepselfhost outcome` 看最短结论，`aepselfhost watch` 前台守着，`aepselfhost start-watch` 后台启动。后台入口会写 `watch_process.json`，stdout/stderr 会落在 `watch_logs`。watch 会按间隔重复刷新同一组 `latest_outcome.*` 并写 `watch_status.json`。验收完成后也可直接打开 `tmp\technique_selfhost_gate\latest_outcome.html`、`tmp\technique_selfhost_gate\latest_outcome.md`、`tmp\technique_selfhost_gate\latest_outcome.json`、`tmp\technique_selfhost_gate\latest_effectiveness.md`、`tmp\technique_selfhost_gate\latest_effectiveness.json`、`tmp\technique_selfhost_gate\latest_index.html` 或 `tmp\technique_selfhost_gate\latest_acceptance.md` 查看最近一次结果；`latest_outcome.html` 是最短浏览器结论页，直接展示状态、Effectiveness Headline、数字、Next Actions、复刻 readiness、当前 blocker、Top Plugin Blockers、最近运行趋势和下一步学习信号，`latest_outcome.md` 是同一批关键信号的最短文本结论，包含 Effectiveness Headline、Next Actions、复刻 readiness 和当前 blocker，`latest_outcome.json` 是同一批关键信号的紧凑机器结论，适合前端、状态面板和批处理直接读，`latest_effectiveness.md` 是更完整的人工成效摘要，`latest_effectiveness.json` 是给脚本、前端和批处理消费的完整成效快照，包含 `outcome_status`、`outcome_summary.headline`、`action_plan.next_actions`、`reconstruction_status`、语料统计、闭环复刻结果、关键 artifact 路径、下一步学习信号、相对上次运行的 `history_delta` 和最近运行 `history_recent`；其中 `reconstruction_status.plugin_blockers_top` 汇总当前最影响复刻 readiness 的第三方效果。每次验收还会追加 `history.jsonl` 和 `history.csv`，用于观察多次运行之间的项目数、pattern 数、闭环复刻和 batch smoke 趋势。HTML 入口会直接显示 Outcome Status、History Delta、Effectiveness History、Study Queue、Pattern Playbook、Coverage Scorecard、Reconstruction Blueprint 和 Recipe Draft 预览，并链接由 recipe draft 自动批量 validate/compile/reparse 出来的 smoke artifacts。

## 原理与分层

`.aep` 是 **RIFX**（Big-Endian RIFF）格式（魔数 `RIFX` + `Egg!`），内部为嵌套 Chunk 树；本库通过逆向工程已知偏移量提取数据。核心代码仍按单向 DAG 分层，但现在已经扩展出迁移、配方、治理和技法内化几条工作面：

```text
internal/rifx          通用 RIFX/Chunk 二进制读写器（无 AEP 语义）
internal/codec         纯值/字节编解码与 layout helper

internal/scene         运行时模型（Project/Composition/Layer/…）+ writer 接口
internal/serializer    chunk ⇄ scene（parse / lower / write / back / mutate）
internal/aep           公开 API facade（Open / FromReader / New* / 类型别名）

internal/aep_test      public API + AE ship-gate 测试面
internal/aehost        AE host 调用与发现
internal/aeoracle      AE render / frame / pixel compare 证据
internal/profile       解析工程的稳定 profile，用于 diff / migration / recipe

internal/aepmigrate    AE 版本迁移：profile → rebuild → verify / matrix
internal/recipe        JSON recipe 校验与编译，走 internal/aep facade 生成工程
internal/recipedoc     recipe schema / capability 文档生成

internal/apidoc        public API 注释词表真相源
internal/capindex      capability 索引与查询
internal/registry      位置 / 证据 / coverage / ownership / cleanup 治理

internal/technique     从 profile 抽取技法事实、画像、解释
internal/selfhost      技法内化的报告与验收面
internal/server        HTTP parse/profile 服务
```

新增 public 能力时通常沿这条线落地：`internal/aep` facade doc comment + cap tag → `internal/serializer` 实现 → `internal/scene` 状态 / writer contract → `internal/aep_test` gate → docgen/capindex。字节布局知识沉淀到 `flightdeck/knowledge/<domain>/`；长期引用的机器证据进入 `registry/` 或 `registry/evidence/<topic>/`，不把 `tmp/` 当真相源。

## 参考

- RIFX 规范：[RIFF/RIFX on Wikipedia](https://en.wikipedia.org/wiki/Resource_Interchange_File_Format)
- Go 实现参考：[boltframe/aftereffects-aep-parser](https://github.com/boltframe/aftereffects-aep-parser)
- Python 实现参考：[forticheprod/py-aep](https://github.com/forticheprod/py-aep)
- API 文档风格参考：[docsforadobe/after-effects-scripting-guide](https://github.com/docsforadobe/after-effects-scripting-guide)
