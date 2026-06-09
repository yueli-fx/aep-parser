---
status: active
summary: GradientStroke R/W（最小·对称已 ship 的 GradientFill）：reader 补 G-Stroke 节点解 gradient（局部降级，对齐 G-Fill），writer 加 GradientStrokeNode 只 model gradient color/alpha stops；fill 的 lower/hydrate 本体抽共享 helper、stroke 复用；双版本 ship-gate（color+alpha，解 stops 比值，对齐 GradientFill）。stroke 几何 + ramp geometry deferred
last_updated: 2026-06-10
---

# GradientStroke R/W 写读路径设计

## 背景

GradientFill 已 ship（V2.2.1 子项⑭，read 先行 + write 后加，双版本 ship-gate PASS，详 `../incidents/gradient-fill-write-re.md`）。GradientStroke（match-name `ADBE Vector Graphic - G-Stroke`）是其 follow-on —— incident 结尾明示「走同样 GCst 路径，source fixture 已含 G-Stroke body」。**基于现有已验证 fixture 的结构性实现**，无需新增 RE 工作（go/no-go 已 dump 实证，见下）。

经两轮外部 review 校正：从初稿「W-only」修正为「R/W」（reader 当前不认 G-Stroke、跳过节点，ship-gate readback 依赖读路径）；ship-gate 验证方式对齐 GradientFill 的实际做法（解 stops 比值，非整文件 byte-compare）；reader 错误策略对齐 GradientFill 的局部降级。

## 范围（最小·对称 GradientFill，R/W）

- ✅ **read**：parse G-Stroke 节点 → `GradientStrokeNode`，解出 gradient color/alpha stops
- ✅ **write**：`AddGradientStroke` + gradient stops 经 prop.map XML 覆写
- 🗑️ **deferred**：stroke 几何（width / cap / join / miter / dashes / taper / wave）+ ramp geometry（Grad Type / Start Pt / End Pt）—— 停 AE default，同 GradientFill。**`GradientStrokeNode` 不暴露这些 setter**（避免误以为设了会生效）；保持模板提取的原值。

理由：stroke 几何能否 model 取决于 G-Stroke 模板里 AE 留了哪些 slot（fixture 默认值被 elide）。最小版只依赖 gradient colors slot（已 dump 实证存在）。

## 模板提取（已 dump 实证可行）

`test_data/v2_2_gradient_src.aep`（AE 25.6-saved，唯一 stops-bearing fixture）的 dump 确认 G-Stroke 节点含 GCst、与 G-Fill 结构同构：

```
tdmn = ADBE Vector Graphic - G-Stroke
  tdmn = ADBE Vector Grad Colors          ← 与 G-Fill 同一挂载点
  [LIST GCst] → [LIST GCky] → Utf8 (1707 B)   ← 非默认 gradient XML，含 GCst slot
tdmn = ADBE Vector Graphic - G-Fill
  tdmn = ADBE Vector Grad Colors → GCst → GCky → Utf8 (1789 B)
```

→ 提取 G-Stroke body subtree 存 `internal/serializer/templates/v2_2_shape_gradstroke_body.bin`。挂载点与 G-Fill 一致，共享 helper 可直接复用。

**no-go 兜底（流程处置）**：前提已实证，预期不触发。若实现期模板不可用（提取后 AE 拒收），则：本 spec 转 **blocked**、登记 incident、**在 coverage.md 的 GradientStroke 行加一行 `blocked: 见 incident <link>` 注记**（保持可见信号，不让矩阵与实际脱节）、等待 UI 重制 fixture（JSX 不能 author gradient stops，详 incident「Two elision traps」节）。

## 设计

### 数据流（R/W）

```
read:  parse → collectShapeKids 命中 G-Stroke → hydrateGradientStrokeNode
          → hydrateGradientStops(body) → GradientStrokeNode{gradient}
write: VectorGroup.AddGradientStroke() → GradientStrokeNode{gradient}
          → SetColorStops / SetAlphaStops（mutator：直接改内部 gradient 字段）
          → WriteAEP → lowerGradientStrokeNode → lowerGradientStops(clone(模板), gradient)
              → rifx.Chunk.Write 自动 reflow LIST sizes（length-variable，免手工 fixup）
```

### 1. scene 层（`internal/scene/scene_shape_graph.go`）

- 新增 enum 值 `ShapeKindGradientStroke`（加入 `ShapeNodeKind`）
- `GradientStrokeNode struct { gradient *codec.Gradient }` —— 对称 `GradientFillNode`，只持 gradient
- `NewGradientStrokeNode()` → 默认 2-stop 黑→白
- 方法 `Kind() / Gradient() / SetColorStops() / SetAlphaStops()` —— 与 `GradientFillNode` 相同验证逻辑（已核实 G-Fill：≥2 stops、color/alpha/offset/midpoint 全 ∈[0,1]，对齐正确）。mutator 语义：setter 改内部 `gradient` 字段，lower 时一次性 encode
- `VectorGroup.AddGradientStroke() (*GradientStrokeNode, error)` —— 对称 `AddGradientFill`：追加到 Children 末尾（= AE render/stacking order）

**设计论证（为何独立节点，不扩展 StrokeNode）**：AE 里 Fill/Stroke/G-Fill/G-Stroke 是 4 个独立 match-name 节点。`StrokeNode.color` 是纯色 `PropertyStream`，G-Stroke 的「颜色」是 gradient（GCst XML），数据模型本质不同。独立节点对称已有的 `GradientFillNode`。

**与 AddStroke（纯色）共存**：均为独立节点、按 Children 顺序渲染、可叠加（非互斥）；渲染结果由 AE 的 render order 决定，本库只保证 Children 顺序在 lower 时保留（替换单节点 body，顺序由 slice 决定，对称 GradientFill）。

### 2. reader 层（`internal/serializer/parse_shape_hydrate.go`）

- `collectShapeKids` 分派 switch 加 `case "ADBE Vector Graphic - G-Stroke"` → `hydrateGradientStrokeNode`
- **错误策略对齐 G-Fill（局部降级，已核实）**：`hydrateGradientFillNode` 找不到 `ADBE Vector Grad Colors` 时返回带默认 gradient 的 node、不报错，`collectShapeKids` 跳过 nil 保留其余、整个 parse 不失败。`hydrateGradientStrokeNode` 照此 —— **不返 error 中断 parse**（修正初稿的过激进策略）。多个 grad colors（异常）取第一个，同 G-Fill。

### 3. serializer/lower 层（`internal/serializer/lower_shape_node.go`）

- `shapeMatchNames` 表加 `ShapeKindGradientStroke: "ADBE Vector Graphic - G-Stroke"`
- embed `templates/v2_2_shape_gradstroke_body.bin` + `cloneShapeGradStrokeBody()`（独立 `sync.Once`）
- `lowerShapeNode` switch 加 `case *GradientStrokeNode`
- **抽共享 helper（YAGNI 阈值已到——第二个 gradient 节点）**：把现有 `lowerGradientFillNode` / `hydrateGradientFillNode` 的本体抽成 `lowerGradientStops(body, *codec.Gradient)` / `hydrateGradientStops(body) *codec.Gradient`（覆写 / 解析 `ADBE Vector Grad Colors` XML）。fill 重构去调它、stroke 复用。`lowerGradientFillNode` / `lowerGradientStrokeNode` 各自 clone 自己的模板后调 `lowerGradientStops`；**本体逻辑唯一差异 = clone 哪个模板**（match-name 差异由 `shapeMatchNames` 表承担）。fill 既有 ship-gate test 守护重构无回归。

### 4. codec 层

**零改动**。`codec.Gradient` / `EncodeGradientXML` / `ParseGradientXML` fill/stroke 通用。

### 5. facade（`internal/aep`）

按 `AddGradientFill` 同样方式暴露 `AddGradientStroke`。

## 测试 / 验收

### ship-gate（铁律 #6，对齐 `runV2_2GradientFillShipGate`）

- `TestV2_2_GradientStroke_AEShipGate_AE2020` / `_AE2025`。同一测试函数内完整流程：构造 NewProject → NewComposition → NewShapeLayer → AddRect → AddGradientStroke → WriteAEP → `runAeRunShipGate` 自驱 `scripts/ae_run.ps1`（AE 打开 + JSX resave）→ `aep.Open(resaved)` 二次解析。resaved 路径经 JSX args JSON 约定（同 GradientFill）。
- **必须同时设非默认 color stops（红/绿/蓝 3 stop）+ 非默认 alpha stops**（如 [0→1.0, 1→0.3] ramp）—— 否则 alpha 写路径未验证。
- **验证方式 = 解 stops 比值（对齐 GradientFill，非整文件 byte-compare）**：`aep.Open(resaved)` 解出 `GradientStrokeNode` 的 color + alpha stops，与写入值比（RGB/alpha 误差 < 0.02）。**不做整文件 byte-identical 断言** —— AE resave 会规范化 chunk，整文件比较不可靠；opaque preservation 由 lower 层「克隆模板只覆写 gradient slot」在实现层保证（铁律 #5），非 ship-gate 断言对象。
- JSX `test_data/verify_v2_2_gradstroke.jsx`（对称 `test_data/verify_v2_2_gradient.jsx`，复用其 args JSON 参数化）：打开 .aep 找 `ADBE Vector Graphic - G-Stroke` 确认 AE 未丢弃 + **顺手验 color stop 数 = 3**（更早暴露 AE 打开时的静默回退），resave。
- agent 自助跑 `ae_run.ps1`（AE 2020 + 2025 双开）。

**复用已验证机制（不重验）**：length-variable XML overwrite（stop 数变化时 LIST 自动 reflow）+ 非默认 stops 规避 AE elision trap —— 已在 GradientFill ship-gate 验证，G-Stroke 经共享 `lowerGradientStops` 复用，本 spec 只验 G-Stroke 特定的 match-name / 模板 / reader。

### 验收清单（commit 前逐项过）

1. `go vet ./... && go test ./...` 全绿，无新增 warning
2. 双版本 ship-gate PASS（AE 2020 + 2025）
3. round-trip：parse 含 G-Stroke 的 .aep → 解出 stops；WriteAEP → byte 结构正确
4. `cmd/aepdemo/main.go` 加对称 `AddGradientStroke` 演示图层（对齐既有 AddGradientFill 演示，第 78-84 行附近）
5. docs（docgen 重生成）+ coverage.md 加「子项⑮ — Gradient stroke（R/W done，alpha）」行（**对齐既有「子项⑭」格式，不生造 R/W/RW 枚举**）+ coverage-detail.md + cockpit 同步
6. commit（完成即 commit；新增 alpha API，按 `../checklists/commits.md` § API 表惯例，非 BREAKING）

## 复用 vs 新增

**100% 复用（零改）**：`codec.Gradient` / `EncodeGradientXML` / `ParseGradientXML`。
**重构为共享**：`lowerGradientStops` / `hydrateGradientStops`（从 fill 本体抽出，fill + stroke 共用）。
**新增**：`ShapeKindGradientStroke` enum · match-name 映射 · `GradientStrokeNode` + 4 方法 + 构造 · `AddGradientStroke` · reader `hydrateGradientStrokeNode` + switch case · G-Stroke 模板 `.bin` · `cloneShapeGradStrokeBody` + `lowerGradientStrokeNode` · ship-gate test + JSX · aepdemo 演示。

## 硬约束遵循（详 `../../CLAUDE.md` § 硬约束）

- **#1**：结构性写（模板克隆 + length-variable XML），非 length-preserving；`rifx.Chunk.Write` 自动 reflow。
- **#5（opaque preservation）**：模板克隆保留 fixture 的全部 opaque chunk，只覆写 gradient slot —— lower 层保证（非 ship-gate byte-compare）。
- **#6**：双版本 AE 2020 + 2025 ship-gate 通过才算 ship。
- **public API 分级**：`AddGradientStroke` 是新 **Alpha** API。`GradientStrokeNode` 经 `AddGradientStroke` 构造（不暴露 unkeyed 字面量），后续加 stroke 几何 setter 是新增方法（API 兼容）。

## 实现步骤（交 writing-plans）

1. **[go，已实证]** 从 `v2_2_gradient_src.aep` 提取 G-Stroke body → `templates/v2_2_shape_gradstroke_body.bin`。
2. scene：enum `ShapeKindGradientStroke` + `GradientStrokeNode` + 4 方法 + 构造 + `AddGradientStroke`。**enum 改完立即全仓检查 `ShapeNodeKind` 的 switch、补齐 case**（防 exhaustive 漏；不留到最后）。
3. reader：`collectShapeKids` 加 G-Stroke case + `hydrateGradientStrokeNode`（局部降级，对齐 G-Fill）。
4. serializer/lower：先把 fill 的 lower/hydrate 本体抽成共享 `lowerGradientStops` / `hydrateGradientStops`、fill 改调它；再加 match-name + 模板 embed/clone（独立 once）+ lower switch + `lowerGradientStrokeNode`。
5. facade：暴露 `AddGradientStroke`。
6. 单测（lower byte 结构 + parse round-trip）+ `go vet ./... && go test ./...`；fill 既有测试须仍绿（守护重构）。
7. ship-gate test（color + alpha + 解 stops 比值 + JSX stop-count）+ JSX；自助跑 AE 2020 + 2025。
8. 验收清单逐项过（含 aepdemo + coverage 子项⑮）；commit。

## Deferred（后续 follow-on）

- stroke 几何（width / cap / join / miter / dashes / taper / wave）—— 需 G-Stroke 模板有对应 slot；可复用 StrokeNode 现成 lower。
- ramp geometry（Grad Type / Start Pt / End Pt）—— 需带非默认 ramp 的 fixture。
- animated gradient stroke —— deferred。

## 参考

- `../incidents/gradient-fill-write-re.md` —— gradient 写 RE + ship 教训（GCst→GCky→Utf8 / 两个 elision trap / cross-version AE25→AE2020 portable）
- `../plans/coverage.md` § 子项⑭ —— GradientFill 行（GradientStroke 子项⑮ 对齐其格式）
