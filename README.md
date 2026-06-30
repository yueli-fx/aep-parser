# aep-parser

从头实现的 Adobe After Effects `.aep` 项目文件解析器 + length-preserving 写回库，纯 Go，无需 AE 运行实例。

> ⚠️ **开发中（WIP）** —— 功能仍在扩展，详细能力清单待项目成熟后再补。
> 当前 API 以自动生成的 [docs/](docs/) 为准。

**兼容下限：After Effects 2020（CC 17.0）**。新版本写的 .aep 也能读；AE 24+ 才引入的字段本库不主动解码，对调用方返回 nil 而非报错。

## 文档

- **API 参考**：[docs/](docs/) —— 每个核心类型一个 markdown，由 `cmd/docgen` 从导出符号的 doc comment **自动生成**（`go generate ./cmd/docgen`）。
- **能力覆盖矩阵 / 暂搁 / 不可达 / negative findings**：[docs/capabilities.md](docs/capabilities.md)。

## 自托管技法报告

对本地 `.aep` 语料生成可浏览的 technique learning report：

```powershell
pwsh -NoProfile -File scripts\technique_showcase_report.ps1 -InputPath data\samples -OutDir tmp\technique_samples_report -Verify
```

输出包括 `manifest.json`、`learning.md`、`report.html`、`projects.csv`、`project_playbooks.csv`、`compositions.csv`、`layers.csv`、`recreation_steps.csv`、`patterns.csv`、`study_queue.csv`、`study_tasks.csv`、`recreation_blockers.csv`、`signal_layers.csv`、`effect_stacks.csv`、`shape_operators.csv`、`text_animators.csv`、`dependency_edges.csv`、`learning_actions.csv`、`mechanisms.csv`、`mechanism_examples.csv`、`coverage_scorecard.csv`、`reconstruction_blueprints.jsonl`、`recipe_drafts.jsonl`、`errors.csv`、`digest.json`、`summary.json`、`corpus.jsonl` 和 `report.md`。`report.html` 是自包含入口，支持按路径、readiness、pattern、effect、plugin、mechanism、study task 搜索，并展示每个项目的确定性复刻步骤、artifact 覆盖状态、复刻蓝图和安全 recipe 草稿；`manifest.json` 记录输入、git 版本、耗时和 artifact 清单；`learning.md` 是更短的人工学习索引；CSV/JSONL 文件适合直接用表格或流式脚本筛项目、项目级复刻 playbook、逐项复刻步骤、合成/图层结构、逐层 effect stack、shape/text 操作项、依赖边、技法模式、学习顺序、学习任务、复刻阻塞项、关键层、学习动作、机制目录、机制代表项目、覆盖计分卡、代码生成蓝图、recipe 骨架草稿、错误项目和 pattern 级复刻步骤分布。

对比两次报告：

```powershell
pwsh -NoProfile -File scripts\compare_technique_reports.ps1 -BaseDir tmp\old_report -NewDir tmp\technique_samples_report
```

一键跑完整自托管验收：

```powershell
pwsh -NoProfile -File scripts\verify_technique_selfhost.ps1 -OutRoot tmp\technique_selfhost_gate
```

想跑完后立刻查看浏览器入口，可以加 `-Open`。验收完成后也可直接打开 `tmp\technique_selfhost_gate\latest_outcome.html`、`tmp\technique_selfhost_gate\latest_outcome.md`、`tmp\technique_selfhost_gate\latest_effectiveness.md`、`tmp\technique_selfhost_gate\latest_effectiveness.json`、`tmp\technique_selfhost_gate\latest_index.html` 或 `tmp\technique_selfhost_gate\latest_acceptance.md` 查看最近一次结果；`latest_outcome.html` 是最短浏览器结论页，`latest_outcome.md` 是最短人工结论，`latest_effectiveness.md` 是更完整的人工成效摘要，`latest_effectiveness.json` 是给脚本、前端和批处理消费的同一份成效快照，包含 `outcome_status`、语料统计、闭环复刻结果、关键 artifact 路径、下一步学习信号和相对上次运行的 `history_delta`。每次验收还会追加 `history.jsonl` 和 `history.csv`，用于观察多次运行之间的项目数、pattern 数、闭环复刻和 batch smoke 趋势。HTML 入口会直接显示 Outcome Status、History Delta、Effectiveness History、Study Queue、Pattern Playbook、Coverage Scorecard、Reconstruction Blueprint 和 Recipe Draft 预览，并链接由 recipe draft 自动批量 validate/compile/reparse 出来的 smoke artifacts。

## 原理与分层

`.aep` 是 **RIFX**（Big-Endian RIFF）格式（魔数 `RIFX` + `Egg!`），内部为嵌套 Chunk 树；本库通过逆向工程已知偏移量提取数据。代码按单向 DAG 分层（M8 物理分包）：

```text
internal/rifx        通用 RIFX/Chunk 二进制读写器（无 AEP 语义）
        ↑
internal/codec       纯值/字节编解码
internal/scene       运行时模型（Project/Composition/Layer/…）+ writer 接口
internal/serializer  chunk ⇄ scene（parse / lower / write / back / mutate）
        ↑
internal/aep         薄 facade：公开 API（Open / FromReader / New* / 类型别名）
```

## 参考

- RIFX 规范：[RIFF/RIFX on Wikipedia](https://en.wikipedia.org/wiki/Resource_Interchange_File_Format)
- Go 实现参考：[boltframe/aftereffects-aep-parser](https://github.com/boltframe/aftereffects-aep-parser)
- Python 实现参考：[forticheprod/py-aep](https://github.com/forticheprod/py-aep)
- API 文档风格参考：[docsforadobe/after-effects-scripting-guide](https://github.com/docsforadobe/after-effects-scripting-guide)
