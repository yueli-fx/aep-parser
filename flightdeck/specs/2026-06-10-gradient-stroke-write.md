---
status: active
summary: GradientStroke R/W（最小·对称已 ship 的 GradientFill）：reader 补 G-Stroke 节点解 gradient（collectShapeKids case + hydrateGradientStrokeNode），writer 加 GradientStrokeNode 只 model gradient color/alpha stops（克隆 G-Stroke 模板 + 覆写 prop.map XML，复用 codec + overwriteGradientStopsXML），双版本 ship-gate（color+alpha）。stroke 几何 + ramp geometry deferred
last_updated: 2026-06-10
---

# GradientStroke R/W 写读路径设计

## 背景

GradientFill 已 ship（V2.2.1 子项⑭，read 先行 + write 后加，双版本 ship-gate PASS，详 `../incidents/gradient-fill-write-re.md`）。GradientStroke（match-name `ADBE Vector Graphic - G-Stroke`）是其 follow-on —— incident 结尾明示「走同样 GCst 路径，source fixture 已含 G-Stroke body」。本 spec 是 V3 剩余前沿里唯一纯代码可推 + ship-gate 可自助跑、无需用户额外提供 fixture/RE 的项。

**外部三家 review（2026-06-10）暴露并已验证两点**，本 spec 据此从初稿的「W-only」修订为「R/W」：
- reader 当前**不认 G-Stroke**（节点分派 switch 无 case，遇到即跳过）—— ship-gate 的 readback 验证依赖 reader，故必须补读路径（对称 GradientFill 的 read-先行）。
- go/no-go 前提**已 dump 实证通过**（见 §模板提取），不再是乐观假设。

## 范围（最小·对称 GradientFill，R/W）

- ✅ **read**：parse G-Stroke 节点 → `GradientStrokeNode`，解出 gradient color/alpha stops
- ✅ **write**：`AddGradientStroke` + gradient stops 经 prop.map XML 覆写
- 🗑️ **deferred**：stroke 几何（width / cap / join / miter / dashes / taper / wave）+ ramp geometry（Grad Type / Start Pt / End Pt）—— 停 AE default，同 GradientFill。**`GradientStrokeNode` 不暴露这些 setter**（避免用户误以为设了会生效）；这些字段保持模板提取的原值。

理由：stroke 几何能否 model 取决于 G-Stroke 模板里 AE 留了哪些 slot（fixture 里是默认值的属性被 AE elide，没 slot 就改不了）。最小版只依赖 gradient colors slot（已 dump 实证存在），可行性最确定、风险最低。

## 模板提取（已 dump 实证可行）

`test_data/v2_2_gradient_src.aep`（AE 25.6-saved，唯一 stops-bearing fixture）的 dump 确认 G-Stroke 节点含 GCst、且与 G-Fill 结构同构：

```
tdmn = ADBE Vector Graphic - G-Stroke
  tdmn = ADBE Vector Grad Colors          ← 与 G-Fill 同一挂载点
  [LIST GCst] → [LIST GCky] → Utf8 (1707 B)   ← 非默认 gradient XML，含 GCst slot
tdmn = ADBE Vector Graphic - G-Fill
  tdmn = ADBE Vector Grad Colors → GCst → GCky → Utf8 (1789 B)
```

→ 提取 G-Stroke body subtree 存 `internal/serializer/templates/v2_2_shape_gradstroke_body.bin`。因挂载点 `ADBE Vector Grad Colors` 与 G-Fill 一致，`overwriteGradientStopsXML` 可直接复用。

**no-go 兜底（流程处置）**：本前提已实证，预期不触发。但若实现期发现模板不可用（如提取后 AE 拒收），则：本 spec 转 **blocked**、登记 incident、**coverage.md 状态保持不变（不留半完成）**、等待 fixture 重制（UI 授权，JSX 不能 author gradient stops，详 incident「Two elision traps」节）。

## 设计

### 数据流（R/W）

```
read:  parse → collectShapeKids 命中 G-Stroke → hydrateGradientStrokeNode
          → findGradientStopsXML(body,"ADBE Vector Grad Colors") → codec.ParseGradientXML
          → GradientStrokeNode{gradient}
write: VectorGroup.AddGradientStroke() → GradientStrokeNode{gradient}
          → SetColorStops / SetAlphaStops（mutator：直接改内部 gradient 字段）
          → WriteAEP → lowerGradientStrokeNode
              → 克隆 v2_2_shape_gradstroke_body.bin
              → overwriteGradientStopsXML(body,"ADBE Vector Grad Colors",EncodeGradientXML(gradient))
              → rifx.Chunk.Write 自动 reflow LIST sizes（length-variable，免手工 fixup）
```

### 1. scene 层（`internal/scene/scene_shape_graph.go`）

- 新 enum 值 `ShapeKindGradientStroke`（追加到 `ShapeNodeKind`，紧邻 `ShapeKindGradientFill`；当前注释已把 `GradientStroke` 列为 V2.3+ candidate）
- `GradientStrokeNode struct { gradient *codec.Gradient }` —— 对称 `GradientFillNode`，只持 gradient
- `NewGradientStrokeNode()` → 默认 2-stop 黑→白（对称 `NewGradientFillNode`）
- 方法 `Kind() / Gradient() / SetColorStops() / SetAlphaStops()` —— 与 `GradientFillNode` 相同验证逻辑（≥2 stops，offset/midpoint/color/alpha ∈ [0,1]）。**mutator 语义**：setter 直接改内部 `gradient` 字段，lower 时一次性 encode（同 GradientFillNode）。首版直接对称实现，不强行抽共享 helper（YAGNI；第三处 gradient 节点出现再重构）
- `VectorGroup.AddGradientStroke() (*GradientStrokeNode, error)` —— 对称 `AddGradientFill`：追加到 Children 末尾（= AE render/stacking order）；可与既有 Fill/Stroke/G-Fill 节点共存（独立节点）

**设计论证（为何独立节点，不扩展 StrokeNode）**：AE 里 Fill/Stroke/G-Fill/G-Stroke 是 4 个独立 match-name 节点，不是「Stroke + gradient flag」。`StrokeNode.color` 是纯色 `PropertyStream`，而 G-Stroke 的「颜色」是 gradient（GCst XML），数据模型本质不同。独立 `GradientStrokeNode` 对称已有的 `GradientFillNode`（也独立于 FillNode），最简且与现有体系一致。

### 2. reader 层（`internal/serializer/parse_shape_hydrate.go`）

- `collectShapeKids` 的 match-name 分派 switch 加 `case "ADBE Vector Graphic - G-Stroke"` → `hydrateGradientStrokeNode`
- `hydrateGradientStrokeNode` —— 对称 `hydrateGradientFillNode`：`findGradientStopsXML(body,"ADBE Vector Grad Colors")` → `codec.ParseGradientXML` → 构造 `GradientStrokeNode{gradient}`
- 防御：找不到 `ADBE Vector Grad Colors`（异常模板）返回 error，不 panic

### 3. serializer/lower 层（`internal/serializer/lower_shape_node.go`）

- `shapeMatchNames` 表加 `ShapeKindGradientStroke: "ADBE Vector Graphic - G-Stroke"`
- embed `templates/v2_2_shape_gradstroke_body.bin` + `cloneShapeGradStrokeBody()`（**独立 `sync.Once`，不复用 gradFill 的 once 变量**）
- `lowerShapeNode` switch 加 `case *GradientStrokeNode: return lowerGradientStrokeNode(n, ctx)`
- `lowerGradientStrokeNode`：**本体逻辑与 `lowerGradientFillNode` 逐字相同**（克隆模板 + `overwriteGradientStopsXML(body,"ADBE Vector Grad Colors",codec.EncodeGradientXML(n.Gradient()))`）；**唯一差异 = 克隆的模板不同**（gradstroke vs gradfill `.bin`）。match-name 差异由 `shapeMatchNames` 表承担，lower 本体不直接引用 match-name。

### 4. codec 层

**零改动**。`codec.Gradient` / `EncodeGradientXML` / `ParseGradientXML` fill/stroke 通用（gradient XML 不分 fill/stroke）。

### 5. facade（`internal/aep`）

按 `AddGradientFill` 同样方式暴露 `AddGradientStroke`（shape builder API 暴露路径）。

## 测试 / 验收

### ship-gate（铁律 #6）

- `TestV2_2_GradientStroke_AEShipGate_AE2020` / `_AE2025`（对称 `runV2_2GradientFillShipGate`，`internal/aep/shape_gradient_shipgate_test.go`）。动态构造 NewProject → NewComposition → NewShapeLayer → AddRect → AddGradientStroke。
- **必须同时设非默认 color stops（红/绿/蓝 3 stop）+ 非默认 alpha stops**（如 [0→1.0, 1→0.3] ramp）—— 否则 alpha 写路径完全未被验证。
- **opaque round-trip 断言**：resaved 后，模板里非 gradient 的 chunk（stroke 几何等）byte-identical（铁律 #5；模板克隆只覆写 gradient slot）。
- **readback**：reader 已补 G-Stroke，Go 端 parse resaved → `GradientStrokeNode` 解回写入的 color + alpha stops（值匹配）。
- JSX `test_data/verify_v2_2_gradstroke.jsx`（对称 `verify_v2_2_gradient.jsx`，**复用现成 args JSON 参数化模式**）：打开 .aep 找 `ADBE Vector Graphic - G-Stroke` 确认 AE 未丢弃，resave。
- ship-gate agent 自助跑 `scripts/ae_run.ps1`（AE 2020 + 2025 双开）。

**复用已验证机制（无需再验）**：length-variable XML overwrite（stop 数变化时 LIST 自动 reflow）+ 用非默认 stops 规避 AE elision trap —— 二者已在 GradientFill ship-gate（`runV2_2GradientFillShipGate`）验证，G-Stroke 复用同一 `overwriteGradientStopsXML` / `EncodeGradientXML`，本 spec 不重验机制，只验 G-Stroke 特定的 match-name / 模板 / reader。

### 验收清单（step 8 commit 前逐项过）

1. `go vet ./... && go test ./...` 全绿，无新增 warning
2. 双版本 ship-gate PASS（AE 2020 + 2025）
3. round-trip：parse 含 G-Stroke 的 .aep → 解出 stops；WriteAEP → byte 结构正确
4. docs（docgen 重生成）+ coverage.md / coverage-detail.md 把 GradientStroke 行从「W 待 fixture / V2.3 候选」改为 **R/W done（alpha）** + cockpit 同步
5. commit（完成即 commit；新增 alpha API，按 `../checklists/commits.md` § API 表惯例，非 BREAKING）

## 复用 vs 新增

**100% 复用（零改）**：`codec.Gradient` / `EncodeGradientXML` / `ParseGradientXML` / `overwriteGradientStopsXML` / `findGradientStopsXML`。

**新增**：`ShapeKindGradientStroke` enum · match-name 映射 · `GradientStrokeNode` + 4 方法 + 构造 · `AddGradientStroke` · reader `hydrateGradientStrokeNode` + switch case · G-Stroke 模板 `.bin` · `cloneShapeGradStrokeBody` + `lowerGradientStrokeNode` · ship-gate test + JSX。

## 硬约束遵循（详 `../../CLAUDE.md` § 硬约束）

- **#1**：结构性写（模板克隆 + length-variable XML），非 length-preserving；`rifx.Chunk.Write` 自动 reflow。
- **#5（opaque preservation）**：模板克隆保留 fixture 提取的全部 opaque chunk，只覆写 gradient XML slot —— ship-gate 显式断言（见验收）。
- **#6**：双版本 AE 2020 + 2025 ship-gate 通过才算 ship。
- **public API 分级**：`AddGradientStroke` 是新 **Alpha** API。

## 实现步骤（交 writing-plans）

1. **[go，已实证]** 从 `v2_2_gradient_src.aep` 提取 G-Stroke body → `templates/v2_2_shape_gradstroke_body.bin`。
2. scene：enum + `GradientStrokeNode` + 4 方法 + 构造 + `AddGradientStroke`。
3. **reader**：`collectShapeKids` 加 G-Stroke case + `hydrateGradientStrokeNode`。
4. serializer/lower：match-name 映射 + 模板 embed/clone（独立 once）+ lower switch case + `lowerGradientStrokeNode`。
5. facade：暴露 `AddGradientStroke`。
6. 单测（lower byte 结构 + parse round-trip）+ `go vet ./... && go test ./...`。**新增 enum 后全仓检查 `ShapeNodeKind` 的 switch，补齐 case（防 exhaustive 漏）**。
7. ship-gate test（color + alpha + opaque 断言 + readback）+ JSX；自助跑 AE 2020 + 2025。
8. 验收清单逐项过；docs + coverage + cockpit 同步；commit。

## Deferred（后续 follow-on）

- stroke 几何（width / cap / join / miter / dashes / taper / wave）—— 需 G-Stroke 模板有对应 slot（提取后评估）；可复用 StrokeNode 现成 lower（`lowerStrokeTaper / Wave / Dashes`）。
- ramp geometry（Grad Type / Start Pt / End Pt）—— 同 GradientFill，需带非默认 ramp 的 fixture。
- animated gradient stroke —— 同 GradientFill，deferred。

## 参考

- `../incidents/gradient-fill-write-re.md` —— gradient 写 RE + ship 教训（GCst→GCky→Utf8 结构 / 两个 elision trap / cross-version AE25→AE2020 portable）
- `../plans/coverage.md` —— shape 字段覆盖矩阵（GradientStroke R/W 状态登记处）
