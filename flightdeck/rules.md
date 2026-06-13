---
version: 3.0
disabled_folders: []
---

# flightdeck rules — aep-parser

## House rules

### Project conventions

- **完成即 commit**：每做完一个可交付单元（feature slice / bugfix / 一批文档同步）就 commit，不要攒在工作树里。**本项目的任务进度靠 commit 历史追溯**——未提交的工作不算 landed，会话结束即丢失上下文锚点。
- **直推 main**：本项目无 PR 流，所有工作直接 commit 到 main（见 git log 既有惯例）。不要为常规改动单开分支。
- commit 拆分按逻辑单元：代码 + 其配套测试/fixture/文档同步算一个 commit；无关的 meta/文档改动单独 commit。commit body 可英文按既有惯例。
- **导出 API doc comment = 文档源（docgen）**：`docs/*.md` 由 `cmd/docgen` 从 `internal/aep` 导出符号的 doc comment 自动生成（Swagger 式，唯一源=注释）。**doc comment 用英文为源**（中文可后续一键翻译）；prose 写描述、`Example*` 测试函数写示例、概念表走 `docs/_includes/*.head.md`/`*.tail.md`。改导出符号的 doc comment 后须重生成（`go generate ./cmd/docgen` 或 `go run ./cmd/docgen -manifest docs/docgen.json`）。`docs/*.gen.md` 头有 `DO NOT EDIT` —— 不手改生成物。区分点：**内部实现行内注释仍禁**（CLAUDE.md 铁律），导出符号上方的 doc comment 是文档载体不算违反。

### Showcase（大阶段产出审核）

- **大阶段必出 showcase 供用户审核**（2026-06-13 用户立规）。**大阶段** = 有独立 plan/spec arc 的可交付 feature（如「矢量滤镜家族」整体、新 layer 类型组、expression 激活）；落地后须在 `flightdeck/showcase/<方向>/` 产出一个**纯 Go 从零生成 + AE 实渲**的示例工程并通知用户审核。**小阶段**（单个滤镜 / 单个 `Set*` 字段 / 单 slice）**不单独出**，攒批到所属方向的 showcase 一起更新。
- **目录形态**：`flightdeck/showcase/<方向>/`（按**能力方向**分区：shape-filters / shape-primitives / keyframes-ease / expressions / precomp-nesting / gradient / text / layers …）。每个方向文件夹含：`INDEX.md`（frontmatter 格式段写明测哪个方向 + 正文列测试文件/产出 aep/类型/布局）、`gen.go`（package main 纯 Go 生成器）、`render.jsx`（AE 打开+saveFrameToPng 出 png）。格式细则 + 新增方向流程见 `checklists/showcase.md`。
- **gitignore 策略**：只 ignore 重产物 `*.aep` / `*.png`（已在根 `.gitignore`）；`INDEX.md` + `gen.go` + `render.jsx` **tracked** —— 干净 clone 后 `go run ./flightdeck/showcase/<方向>` + 跑 render.jsx 即可一键重生成全部产物。**生成器必须保持 `go build ./...` 绿**（它是 tracked 代码）。
- **交付准则对齐**（CLAUDE.md #7 / `checklists/delivery-contract.md`）：showcase 里每个能力须是已过双版本 ship-gate 的；showcase 的「组合工程」本身是独立交付项，**产出后必须 AE 实渲眼验**（红线4：先看图），不靠值 round-trip 假绿。

### Autonomy overrides
<!-- migrated from commit_mode:confirm — 旧 toggle 取默认值 confirm，但上方 House rule「完成即 commit·无需用户每次点名」是既定的真实意图，故按 auto 编码 -->
commit without asking

- **文档可自由更新（含破坏性重写、含 CLAUDE.md）**（2026-06-11 用户授权「允许破坏性更新 你可以更新任何文档 包括claude.md」）：任何项目文档——CLAUDE.md / flightdeck deck / docs / doc comment——无需逐次确认即可更新、重组、删改；分级（Alpha↔Stable）等契约表述变更照常在 commit body 说明。代码 API 契约本身仍按 CLAUDE.md #2 执行。

- **大计划完成即自动 landing**：当一个代表已 ship 功能 arc 的 plan 翻到 `status: done`，主动跑 `/flightdeck:landing` 把它归档到 `landed/plans/` 并同步 INDEX + cockpit，无需用户点名。（landing log `landed/HISTORY.md` 已废弃——commit 级历史以 `git log` 为准。）判定「大计划」= 有独立 plan 文件、对应一个可交付 feature/结构性 arc（如 P3 §3G）；滚动 reference 矩阵（coverage*.md）等长期文档不适用，保持 active。
