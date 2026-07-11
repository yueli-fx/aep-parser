# Flightdeck 简报 — aep-parser

## 语言规则

- 对话、任务说明、Flightdeck 工作记录、计划、知识沉淀、交付总结一律使用中文。
- 之前已经存在的英文内容不强制追溯清理；后续新增和更新优先中文。
- 例外：导出的公共 API 注释、代码标识符、命令、错误信息、外部原文引用可保留英文。`internal/aep` 导出符号的公共 API 注释是 `cmd/docgen` 的英文源文本，不要为了中文化而改写。

## 启动入口

- 先读 `flightdeck/cockpit.md`，了解当前活跃工作和下一步。
- 探索或编辑仓库目录时，先读最近且相关的 `README.md`。不要启动时批量读取所有 README。
- 不要批量读取 `flightdeck/knowledge/**`。先看 `flightdeck/knowledge/INDEX.md`，再只加载 `READ WHEN` 与当前任务匹配的笔记。
- 持久知识必须自洽完整：未来行动规则写入 `flightdeck/knowledge/<domain>/`，不要只留在临时规格或工作笔记里。

## 仓库规则

- 直接在当前 branch/mainline 上工作。除非用户要求，不要例行创建 feature branch。
- 永远不要 push 到远端。
- `data/samples/` 是获准使用的本地语料库，不是清理目标。
- `learning/` 是私有学习区，已忽略，不要提交其中内容。
- `tmp/` 是一次性目录。可持久复用的生成证据应放到受跟踪的真相位置，例如 `registry/evidence/<topic>/`、`data/reference/<topic>/`、已提交 fixture 或嵌入模板。
- 大型计划使用 total -> slices -> total：先定义控制台账，再执行切片，最后把状态和缺口汇总回台账。
- 已完成的大型工作从 `flightdeck/work/` 归档到 `~/.flightdeck/projects/<slug>/archive/`，并从 cockpit 的进行中列表移除。

## 提交策略

- 提交代码、测试、fixture、生成文档，或有意义的已完成检查点。
- 不要提交探索性 Markdown 抖动。草稿规格、计划、知识草图、cockpit/index 措辞在讨论或塑形阶段可以保持未提交。
- 纯文档工作只有在用户明确要求提交、明确表示文档已定稿，或它随同一个已落地的代码/测试/fixture 变更一起交付时才提交。
- 不要因为用户说“记录一下”“先放着”“探索一下”“开始设计后续任务”就提交 cockpit/work/spec/plan/knowledge 草稿。
- 提交按已落地单元保持原子性。代码及其必要测试、fixture、文档、Flightdeck 同步可以放在一个提交；无关清理应分开。
- stage 或写 commit message 前，使用 `knowledge/git/commits.md` 中的全局提交检查清单。

## 公共 API 与文档

- `internal/aep` 导出符号的公共 API 注释是文档源。它们是 `cmd/docgen` 使用的英文源文本；生成文档不要手改。
- 公共 API 变更必须同步相关 docs/docgen 注册，并与 capindex 对齐。
- 能力真相源：`go run ./cmd/capindex -q "<term>"`。
- API 稳定性、包边界、写入语义、注释和交付规则见 `flightdeck/knowledge/workflow/project-operating-rules.md`。

## 验证路径

- 通用验证和产物放置：`flightdeck/knowledge/workflow/verify.md`。
- AE ship-gate、fixture 再生成、JSX 逆向：`flightdeck/knowledge/workflow/re-fixture.md`。
- 交付/可发布声明：`flightdeck/knowledge/workflow/delivery-contract.md`。
- Showcase 生成和视觉 review：`flightdeck/knowledge/showcase/showcase.md`。

## 项目提交类型

沿用全局提交约定，并补充以下项目特定规则：

- `re`：纯逆向发现或 fixture 证据，尚未形成已交付 setter/API。
- 常见 scope：`layer`、`comp`、`text`、`mask`、`keyframe`、`shape`、`property`、`marker`、`footage`、`project`、`aep`、`flightdeck`。
- `feat`：新的已交付字段读写、setter、结构 API，或 fixture 支撑的能力。
- `fix`：错误读写字节、setter 行为错误、round-trip / AE-acceptance 回归。
- `docs`：公共 API 文档或 Flightdeck 文档。

## 订阅

<!-- Global knowledge subscriptions, one ~/.flightdeck-relative path per line. -->
knowledge/coding/comments.md
knowledge/git/commits.md
knowledge/agents/subagent-guide.md
