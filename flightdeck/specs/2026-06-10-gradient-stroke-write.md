---
status: active
summary: GradientStroke W（最小·对称已 ship 的 GradientFill）：加 GradientStrokeNode 只 model gradient color/alpha stops，克隆从 fixture 提取的 G-Stroke 模板 + 覆写 prop.map XML（复用 codec + overwriteGradientStopsXML），新增 enum/match-name/lower + 双版本 ship-gate。stroke 几何 + ramp geometry deferred
last_updated: 2026-06-10
---

# GradientStroke W 写路径设计

## 背景

GradientFill 写已 ship（V2.2.1 子项⑭，双版本 ship-gate PASS，详 `../incidents/gradient-fill-write-re.md`）。GradientStroke（match-name `ADBE Vector Graphic - G-Stroke`）是其 straightforward follow-on —— incident 结尾明示「走同样 GCst 路径，source fixture 已含 G-Stroke body」。本 spec 是 V3 剩余前沿里**唯一纯代码可推 + ship-gate 可自助跑、无需用户提供 fixture/RE** 的项。

## 范围（最小·对称 GradientFill）

首版**只 model gradient color/alpha stops**，与 GradientFill 完全对称：

- ✅ gradient stops（color + alpha）经 prop.map XML 覆写
- 🗑️ **deferred**：stroke 几何（width / cap / join / miter / dashes / taper / wave）+ ramp geometry（Grad Type / Start Pt / End Pt）—— 停 AE default，同 GradientFill。

理由：stroke 几何能否 model 取决于提取的 G-Stroke 模板里 AE 留了哪些 slot（fixture 里是默认值的属性被 AE elide，没 slot 就改不了）。最小版只依赖 gradient colors slot（G-Stroke 的核心，fixture 必带非默认 stops），可行性最确定、风险最低。

## 设计

### 数据流

```
VectorGroup.AddGradientStroke() → GradientStrokeNode{gradient}
  → SetColorStops / SetAlphaStops
  → WriteAEP → lowerGradientStrokeNode
      → 克隆 v2_2_shape_gradstroke_body.bin 模板
      → overwriteGradientStopsXML(body, "ADBE Vector Grad Colors", EncodeGradientXML(gradient))
      → rifx.Chunk.Write 自动 reflow LIST sizes（length-variable，免手工 size fixup）
```

### 1. scene 层（`internal/scene/scene_shape_graph.go`）

- 新 enum `ShapeKindGradientStroke`（ShapeNodeKind 第 8 值，现 enum 在 line 16-26）
- `GradientStrokeNode struct { gradient *codec.Gradient }` —— 对称 `GradientFillNode`（line 372），只持 gradient，不建模 stroke 几何
- `NewGradientStrokeNode()` → 默认 2-stop 黑→白（对称 `NewGradientFillNode` line 386）
- 方法 `Kind() / Gradient() / SetColorStops() / SetAlphaStops()` —— 与 `GradientFillNode`（line 408-456）相同验证逻辑（≥2 stops，offset/midpoint/color/alpha ∈ [0,1]）。首版直接对称实现，**不强行抽共享 helper**（YAGNI；若日后第三处 gradient 节点出现再重构）
- `VectorGroup.AddGradientStroke() (*GradientStrokeNode, error)` —— 对称 `AddGradientFill`（line 97）

### 2. serializer/lower 层（`internal/serializer/lower_shape_node.go`）

- `shapeMatchNames` 表（line 235-243）加 `ShapeKindGradientStroke: "ADBE Vector Graphic - G-Stroke"`
- embed `templates/v2_2_shape_gradstroke_body.bin` + `cloneShapeGradStrokeBody()`（sync.Once cache，模仿 `cloneShapeGradFillBody` line 172）
- `lowerShapeNode` switch（line ~257）加 `case *GradientStrokeNode: return lowerGradientStrokeNode(n, ctx)`
- `lowerGradientStrokeNode(n, _ *lowerCtx)` —— 几乎照抄 `lowerGradientFillNode`（line 709）：克隆模板 + `overwriteGradientStopsXML(body, "ADBE Vector Grad Colors", codec.EncodeGradientXML(n.Gradient()))`（复用现成辅助 line 731）

### 3. codec 层

**零改动**。`codec.Gradient` / `EncodeGradientXML`（line 128）/ `ParseGradientXML` fill/stroke 通用（gradient XML 不分 fill/stroke，差异仅在 lower 的 match-name）。

### 4. facade（`internal/aep`）

按 `AddGradientFill` 同样方式暴露 `AddGradientStroke`（shape builder API 暴露路径）。

### 5. 模板提取（实现第一步 = go/no-go gate）

从 `test_data/v2_2_gradient_src.aep`（AE 25.6-saved，唯一 stops-bearing fixture，← py-aep 的 gradient.aep）dump G-Stroke 节点，**验证它含 GCst（即 stroke 带非默认 gradient stops）**：

- **含 GCst** → 提取 G-Stroke body subtree 存 `internal/serializer/templates/v2_2_shape_gradstroke_body.bin`，继续。
- **无 GCst**（fixture 的 stroke 是默认 gradient，被 AE elide）→ 方案受阻。回退：UI 授权重制一个带非默认 gradient stroke 的 fixture（JSX 不能 author gradient stops —— 同 GradientFill 当初的 fixture 困境，详 incident「Two elision traps」节）。

incident 暗示 fixture 含 G-Stroke body，但「含 GCst」必须 dump 实证 —— **这是本 spec 唯一可行性风险点**。

### 6. 测试

- 双版本 ship-gate：`TestV2_2_GradientStroke_AEShipGate_AE2020` / `_AE2025`（对称 `runV2_2GradientFillShipGate`，`internal/aep/shape_gradient_shipgate_test.go:27`）。动态构造 NewProject → NewComposition → NewShapeLayer → AddRect → AddGradientStroke → SetColorStops（红/绿/蓝 3 stops，异于模板默认 2-stop 黑→白）。
- JSX 验证 `test_data/verify_v2_2_gradstroke.jsx`（对称 `verify_v2_2_gradient.jsx`）：打开 .aep，找 `ADBE Vector Graphic - G-Stroke` 节点确认 AE 未丢弃，resave；Go 端打开 resaved 解回写入的 stops。
- ship-gate agent 自助跑 `scripts/ae_run.ps1`（AE 2020 + 2025 双开），不需用户开 AE。

## 复用 vs 新增

**100% 复用（零改）**：`codec.Gradient` / `EncodeGradientXML` / `ParseGradientXML` / `overwriteGradientStopsXML`。

**新增**：`ShapeKindGradientStroke` enum · match-name 映射 · `GradientStrokeNode` 类型 + 4 方法 + 构造 · `AddGradientStroke` · G-Stroke 模板 `.bin` · `cloneShapeGradStrokeBody` + `lowerGradientStrokeNode` · ship-gate test + JSX。

## 硬约束遵循（详 `../../CLAUDE.md` § 硬约束）

- **#1**：走结构性写（模板克隆 + length-variable XML），非 length-preserving；`rifx.Chunk.Write` 自动 reflow LIST + 内嵌 size。
- **#5（opaque preservation）**：模板克隆保留 fixture 提取的全部 opaque chunk，只覆写 gradient XML slot —— 任何 regenerate 路径不得丢它。
- **#6**：双版本 AE 2020 + 2025 ship-gate 通过才算 ship。
- **public API 分级**：`AddGradientStroke` 是新 **Alpha** API（commit 标 BREAKING-free 新增；coverage.md 标 alpha）。

## 实现步骤（交 writing-plans）

1. **[go/no-go]** dump `v2_2_gradient_src.aep` G-Stroke 节点，确认含 GCst；提取 → `templates/v2_2_shape_gradstroke_body.bin`。
2. scene：enum + `GradientStrokeNode` + 4 方法 + 构造 + `AddGradientStroke`。
3. serializer：match-name 映射 + 模板 embed/clone + lower switch case + `lowerGradientStrokeNode`。
4. facade：暴露 `AddGradientStroke`。
5. lower 单测（byte 结构 / round-trip）+ `go vet ./... && go test ./...`。
6. ship-gate test + JSX；自助跑 AE 2020 + 2025 双版本。
7. 同步 docs（docgen）+ coverage.md / coverage-detail.md + cockpit；commit（完成即 commit）。

## Deferred（后续 follow-on）

- stroke 几何（width / cap / join / miter / dashes / taper / wave）—— 需 G-Stroke 模板有对应 slot（dump 后评估）；可复用 StrokeNode 现成 lower（`lowerStrokeTaper / Wave / Dashes`）。
- ramp geometry（Grad Type / Start Pt / End Pt）—— 同 GradientFill，需带非默认 ramp 的 fixture。
- animated gradient stroke —— 同 GradientFill，deferred。

## 参考

- `../incidents/gradient-fill-write-re.md` —— gradient 写 RE + ship 教训（GCst→GCky→Utf8 结构 / 两个 elision trap / cross-version AE25→AE2020 portable）
- `../plans/coverage.md` —— shape 字段覆盖矩阵（GradientStroke W 状态登记处）
