# 自托管技法报告

[中文首页](../README.md) · [English overview](../README.en.md)

以下命令从仓库根目录运行。报告分析不需要 AE；完整验收中的宿主读取与渲染步骤需要可用的 AE worker。`data/samples/` 是用户自行提供的本地语料，不随仓库分发。


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
