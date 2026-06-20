---
version: 3.0
runtime: uv
agents_md: off
---

# flightdeck rules — aep-parser

## House rules

### Project conventions

- **完成即 commit**：每做完一个可交付单元（feature slice / bugfix / 一批文档同步）就 commit，不要攒在工作树里。**本项目的任务进度靠 commit 历史追溯**——未提交的工作不算 landed，会话结束即丢失上下文锚点。
- **直推 main**：本项目无 PR 流，所有工作直接 commit 到 main（见 git log 既有惯例）。不要为常规改动单开分支。
- commit 拆分按逻辑单元：代码 + 其配套测试/fixture/文档同步算一个 commit；无关的 meta/文档改动单独 commit。commit body 可英文按既有惯例。
- **`data/` 是合法测试数据,勿提议清理**：`data/项目/`（含中文名目录 `001_项目A/` `#ep资产/` 等）已由用户确认非敏感（2026-05-17 closed decision）；勿在代码审计中再提 git-purge / gitignore / 换合成 fixture。新出现的路径（如有人丢 `Downloads/xxx.aep`）是另案。
- **导出 API doc comment = 文档源（docgen）**：`docs/*.md` 由 `cmd/docgen` 从 `internal/aep` 导出符号的 doc comment 自动生成（Swagger 式，唯一源=注释）。**doc comment 用英文为源**（中文可后续一键翻译）；prose 写描述、`Example*` 测试函数写示例、概念表走 `docs/_includes/*.head.md`/`*.tail.md`。改导出符号的 doc comment 后须重生成（`go generate ./cmd/docgen` 或 `go run ./cmd/docgen -manifest docs/docgen.json`）。`docs/*.gen.md` 头有 `DO NOT EDIT` —— 不手改生成物。区分点：**内部实现行内注释仍禁**（CLAUDE.md 铁律），导出符号上方的 doc comment 是文档载体不算违反。

### Showcase（视觉能力眼验 — 每渲染类域必覆盖）

- **大阶段必出 showcase 供用户审核**（2026-06-13 用户立规）。**大阶段** = 有独立 plan/spec arc 的可交付 feature（如「矢量滤镜家族」整体、新 layer 类型组、expression 激活）；落地后须在 `flightdeck/showcase/<方向>/` 产出一个**纯 Go 从零生成 + AE 实渲**的示例工程并通知用户审核。**小阶段**（单个滤镜 / 单个 `Set*` 字段 / 单 slice）**不单独出**，攒批到所属方向的 showcase 一起更新。
- **视觉能力域全覆盖 + 占位 + 批验**（2026-06-16 强化，从「大阶段」升级为「每个渲染类能力域」）：每个**会改变渲染像素**的能力方向（对齐 capindex 渲染类 domain：shape / gradient / effect / text / mask / layer-create / keyframe / expr + layer-set·structural·comp 的视觉子集）**必须有一个 showcase 分类**——任何渲染类写能力都需眼验面（**红线4：值对 ≠ 渲染对**，shape「蓝存红显」栽在这；ship-gate 像素门禁 + showcase 人眼 = 双保险）。未做完/受阻 → 建**占位条目**（`status: 待建`/`blocked` + stub/空 gen），使覆盖在 `showcase/INDEX.md` 可见、用户可一处批量校验。**非视觉域豁免看图**（project / render-queue / io / eg-面板 / meta，无渲染面）：ship-gate 值验足够，或出 📋 **读值档**（gen + verify.jsx dump DOM 值核对）。
- **批验入口 + capindex 对齐**：`showcase/INDEX.md` = 全方向 → 产物 → status 总表（A 类看图 / B 类读值 / blocked / from-scratch 不可表达）。**check**：capindex tag `verify=render-pixel` 的能力应被某 showcase 方向覆盖；render-pixel 能力无 showcase = 缺眼验面（红线4 风险）→ 补占位或实档。
- **状态有 review-gate**（2026-06-13 用户立规）：showcase 的 `status` 分两档——**`待review`** = agent 已建+AE 实渲+自己眼验，但**用户尚未在真机打开 .aep 复核**；**`complete`** = **用户真机验过后**才可标。**agent 不得自行把 showcase 标 `complete`**（agent 眼验 ≠ 用户真机验收，同交付准则「Go round-trip ≠ AE 接受」的精神）。新建/更新 showcase 默认落 `待review`，并通知用户复核；用户确认后才翻 `complete`。
- **目录形态**：`flightdeck/showcase/<方向>/`（按**能力方向**分区：shape-filters / shape-primitives / keyframes-ease / expressions / precomp-nesting / gradient / text / layers …）。每个方向文件夹含：`INDEX.md`（frontmatter 格式段写明测哪个方向 + 正文列测试文件/产出 aep/类型/布局）、`gen.go`（package main 纯 Go 生成器）、`render.jsx`（AE 打开+saveFrameToPng 出 png）。格式细则 + 新增方向流程见 `checklists/showcase.md`。
- **gitignore 策略**：只 ignore 重产物 `*.aep` / `*.png`（已在根 `.gitignore`）；`INDEX.md` + `gen.go` + `render.jsx` **tracked** —— 干净 clone 后 `go run ./flightdeck/showcase/<方向>` + 跑 render.jsx 即可一键重生成全部产物。**生成器必须保持 `go build ./...` 绿**（它是 tracked 代码）。
- **交付准则对齐**（CLAUDE.md #7 / `checklists/delivery-contract.md`）：showcase 里每个能力须是已过双版本 ship-gate 的；showcase 的「组合工程」本身是独立交付项，**产出后必须 AE 实渲眼验**（红线4：先看图），不靠值 round-trip 假绿。

### Rules

<!-- AI-authored behavior rules from natural-language requests (source + date). -->

- commit 直推本地 main，无需逐次点名（既定项目惯例：完成一个可交付单元即 commit；无 PR 流）。
- 永不 push 远端：所有工作 commit 到本地 main 即可，landing/任何流程都不提议或执行 `git push`，本地 commit 历史就是交付记录（you, 2026-06-16）。
- 文档可自由更新（含破坏性重写、含 CLAUDE.md）：CLAUDE.md / flightdeck deck / docs / doc comment 无需逐次确认即可更新、重组、删改；分级（Alpha↔Stable）等契约表述变更照常在 commit body 说明；代码 API 契约本身仍按 CLAUDE.md #2 执行（you, 2026-06-11）。
- 大计划完成即自动 landing：代表已 ship 功能 arc 的 plan 翻到 `status: done` 时主动跑 `/flightdeck:landing` 归档 + 同步 INDEX/cockpit，无需点名。判定「大计划」= 有独立 plan 文件、对应一个可交付 feature/结构性 arc；滚动 reference 矩阵不适用，保持 active。
