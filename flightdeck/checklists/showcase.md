---
status: active
last_updated: 2026-06-13
when_to_read: 一个大阶段（独立 plan/spec arc）落地后要产出审核示例；新增一个 showcase 方向文件夹；写或更新 showcase/<方向>/INDEX.md；纠结某次成果算大阶段（必出）还是小阶段（攒批）；想知道 showcase 产物怎么重生成
applies_to: [showcase, example, user-review, deliverable, big-stage, gen-go, render-jsx, index-format, gitignore, regenerate, from-scratch, ae-render]
---

# showcase 产出规约 — 大阶段示例 + 用户审核

> 2026-06-13 用户立规：每个大阶段须从零产出一个示例工程让用户审核；小阶段攒批。规则正文见 `rules.md` § Showcase；本文是格式细则 + 操作流程。

## 何时出（大阶段 vs 小阶段）

- **大阶段（必出 / 必通知用户审核）** = 有独立 plan/spec arc 的可交付 feature。例：矢量滤镜家族整体、新 layer 类型组（Solid/Null/Camera/Light…）、expression 激活、precomp 嵌套、gradient 体系。
- **小阶段（不单独出，攒批）** = 单 slice：单个滤镜、单个 `Set*` 字段、单个 enum。落地时只更新所属方向 showcase 的 `gen.go` + `INDEX.md`，不新建文件夹、不单独通知。

## 目录形态（按能力方向分区）

```
flightdeck/showcase/
  INDEX.md                      ← 总览：列所有方向 + 一句话 + 重生成命令
  <方向>/                        ← shape-filters / shape-primitives / keyframes-ease /
    INDEX.md                    ←   expressions / precomp-nesting / gradient / text / layers …
    gen.go                      ← package main，纯 Go 从零构建 <方向>.aep（tracked）
    render.jsx                  ← AE 打开 + saveFrameToPng 出 <方向>.png（tracked）
    <方向>.aep                   ← 产出工程（gitignored）
    <方向>.png                   ← 渲染帧（gitignored）
```

- **tracked**：`INDEX.md` · `gen.go` · `render.jsx`。**gitignored**：`*.aep` · `*.png`（根 `.gitignore` 已加 `flightdeck/showcase/**/*.aep|*.png`）。
- `gen.go` 是 tracked 代码 → **必须保持 `go build ./...` / `go vet ./...` 绿**；它 `package main`，import `internal/aep` facade（与 tmp_debug 同，模块内允许）。

## INDEX.md 格式（frontmatter 格式段 + 正文）

frontmatter 写明「测哪个方向」+ 元信息；正文列产物 + 类型 + 布局。模板：

```markdown
---
showcase: <方向 slug>
direction: <一句话：这个方向从零测什么>
capabilities: [<能力1>, <能力2>, …]        # 覆盖的能力点
gates: [<对应 ship-gate 测试名>, …]          # 每个能力的双版本 gate（溯源）
status: complete | partial
last_updated: YYYY-MM-DD
regenerate: "go run ./flightdeck/showcase/<方向>  +  AE render.jsx"
---

# <方向> — showcase

## 这个方向测什么
<2-3 句：从零生成了什么、覆盖哪些能力、为什么这样布局>

## 产物
| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go) | 纯 Go 构建 <方向>.aep |
| render.jsx | 渲染脚本 | AE saveFrameToPng → <方向>.png |
| <方向>.aep | 产出工程 (gitignored) | <层数/结构> |
| <方向>.png | 渲染帧 (gitignored) | frame 0 视觉 |

## 布局
<网格/层次表：每个 cell/layer = 哪个能力 + 期望效果，供用户逐项眼验>
```

## 状态 review-gate（`status` 两档）

| 值 | 含义 | 谁能标 |
|---|---|---|
| **`待review`** | agent 已建 + AE 实渲 + 自己眼验，**用户尚未真机复核** | agent（默认落这档） |
| **`complete`** | **用户在真机打开 .aep 验收过** | **仅用户确认后** agent 才翻 |

**agent 自己渲染眼验 ≠ 用户真机验收**（同交付准则「Go round-trip ≠ AE 接受」）。**agent 禁止自行标 `complete`**；新建/更新一律落 `待review` 并通知用户。顶层 `showcase/INDEX.md` 用 `🔍 待review` / `✅ complete` 对应。

## 操作流程（出一个大阶段 showcase）

1. 新方向 → 建 `flightdeck/showcase/<方向>/`，写 `gen.go`（参考既有方向；纯 Go facade 调用，输出到本目录的 `<方向>.aep`）。
2. `go run ./flightdeck/showcase/<方向>` 构建 .aep。
3. 写 `render.jsx`（打开 .aep → `saveFrameToPng(0, …)`），用 `scripts/ae_run.ps1` 跑出 `<方向>.png`。
4. **AE 实渲眼验**（红线4）：Read 渲染 png，确认每个能力的视觉对了——不靠值 round-trip 假绿。
5. 写/更新 `<方向>/INDEX.md`（上面模板，`status: 待review`）+ 顶层 `showcase/INDEX.md` 加一行（`🔍 待review`）。
6. commit tracked 三件（INDEX.md/gen.go/render.jsx）；**通知用户真机复核**。
7. **用户真机验收通过后** → 才把该方向 `status` 翻 `complete`（顶层表同步 ✅）。

## 交付准则对齐

showcase 里每个能力须已过双版本 ship-gate（`delivery-contract.md`）；showcase「组合工程」本身是独立交付项，**产出必经 AE 实渲眼验**，禁止只验值 round-trip 就宣称示例「能看」。
