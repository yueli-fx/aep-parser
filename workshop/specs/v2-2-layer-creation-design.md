# V2.2 — ShapeLayer creation design

**Status**: spec, ready for plan phase
**Created**: 2026-05-22
**Brainstorm**: chat 转录已整合；3 路 LLM 反馈（deepseek / gpt / claude）已 absorb 进各 section
**Predecessor**: V2.1 (NewProject + NewComposition) — `workshop/plans/v2-1-foundation-plan.md` + `workshop/scars/ae25-acceptance-gate.md`
**Successor planning input**: `workshop/specs/v3-direction.md`

## 0. Context & 范围决策

V2.2 = V2 第二个 sub-project：在 V2.1 NewComposition 基础上加 **ShapeLayer end-to-end**（含节点合成 + PropertyStream 动画 + AE 2020/2025 ship gate）。

**Dual-track 战略**（per `workshop/feedback/gpt`）:
- **V2 track**: 继续 ship 实际 RE 进度 + 字段覆盖
- **V3 track**: 从 V2 work 中增量提取 runtime IR + serializer split + capability matrix

V2.2 既 ship Layer creation 又服务 V3 brainstorm 阶段所需的所有 runtime/serializer 边界 + classification 产出。

### V2.2 范围决策（brainstorm 已收齐）

| Q | 选定 |
|---|---|
| 1. Layer kind 范围 | ShapeLayer end-to-end，其它 layer kinds 推 V2.3+ |
| 2. PropertyStream 深度 | 静态 + keyframe（状态机：Static ↔ Animated 互斥）|
| 3. Shape 节点集 | Rect + Ellipse + Path + Fill + Stroke (5) |
| 4. Capability matrix | 增量加（实施期发现 + admission rule 闸门）|
| 5. 测试策略 | Tier 1 Go roundtrip + Tier 2 JSX ship gate + Tier 3 preservation tolerance |
| 6. API 形态 | Typed setters（α）为主体；escape hatch β 通向 generic property tree |
| 7. 主 fixture | **(ii) + (C)** — 3 ShapeLayer multi-root, Layer Position + Rect Size keyframes |
| 8. 次 fixture | (iii) nested group preservation (parse / write / reopen 稳定，不开 synthesis API) |

### V2.2 范围**外**（V2.3+）

- 删 layer / 删节点 / reorder API
- 嵌套 group **合成** API（preservation territory：parse/write 稳，不开 synthesis 入口）
- 3D ShapeLayer + 3D Layer Position
- Group-level Transform typed setter（用户走 escape hatch β）
- LineCap / LineJoin / Dashes / Gradient / Trim / Merge / Repeater
- Effect / Mask / Text / Camera / Light / AV 等 layer 创建
- Separated dimensions PropertyStream 模式
- Expression typed API（V1 已支持 string setter；V2.2 默认 empty）
- Golden byte fixture 测试
- Path 带 tangent 的 typed setter（V2.2 默认 linear，tangent 全 0）
- 性能 / 并发 / 大文件压力测

---

## 1. Architecture

### 1.1 核心声明

> **Chunk trees are serialization artifacts, not canonical semantic structures. The runtime graph is the authoritative representation of project semantics.**

V2.2 serializer 实现内部可能仍走 patch-oriented mechanisms（length-preserving splice / archetype seeding），但**对外暴露的 runtime semantics 已经稳定**。不是 "V2 临时 / V3 才真"；是 "runtime API 已对，serializer 内部允许过渡"。

Parser 解出的 chunk tree = **intermediate reconstruction substrate**，最终 lowering 反向回 canonical runtime graph 才算 hydration 完成。

### 1.2 架构 Invariants

V2.1 加进 12 条 Project Invariants（mutation 正确性）。V2.2 加 10 条 Architecture Invariant（分层正确性）。

| # | Invariant |
|---|---|
| **Inv-1** | **Runtime never references chunks**. runtime types 不带 `*rifx.Chunk` / chunk offsets / serializer IDs。既有 `Layer.ldta *rifx.Chunk` 等 V1 字段属 serializer-side cache，必须 unexported；新 runtime types 一开始就不带 |
| **Inv-2** | **Serializer owns all binary concerns**. RIFX / DLay / tdgp / ldta / padding / chunk ordering / counter layout / version-specific encoding 全部局限在 serialization path。**Serializer primitives are permitted to mutate chunk topology. Runtime primitives are not.** Topology ownership 是 boundary 的核心 |
| **Inv-3** | **Runtime semantics are version-independent**. Version divergence 走 capability-based serializer lowering，feature code 不允许 `if AE2020 / if AE2025` |
| **Inv-4** | **PropertyStream is canonical**. Layer Transform / Shape props / Effect params / Mask 路径 keyframe —— 同一个 `PropertyStream<T>` shape（dimensions / keyframes / interpolation / expression）|
| **Inv-5** | **Runtime semantics independent of serializer artifacts**. padding / ordering / counter / sibling chunk topology 不影响 runtime API 行为 |
| **Inv-6** | **Archetypes are bootstrap seeds only**. `EmptyShapeLayer` 等是 minimum valid **serializer substrate**，**不是** runtime semantic template。runtime 不通过 "继承 archetype" 获得 default 值；archetype 只服务 lowering 时找一份合法字节起点 |
| **Inv-7** | **Tick-based timing is serializer-only**. Runtime timing 唯一表达 = `float64 seconds`。ticks / masterTicks / fps quirks / frame rate canonical encoding 全部 serializer 内部，不进 runtime API |
| **Inv-8** | **PropertyStream is a state machine**: Static ↔ Animated, mutually exclusive. Adding the first keyframe transitions stream from static to animated mode and **invalidates the previous static value**. `Clear()` transitions back to static mode |
| **Inv-9** | **Opaque preservation**: Unknown child chunks are preserved byte-identically. Serializer only rewrites owned subgraphs (per Section 4.0 ownership map). Future AE versions / 3rd-party effects / plugin blobs 跨 mutation 不变 |
| **Inv-10** | **Runtime semantic MUST NOT depend on**: chunk ordering / chunk IDs / byte offsets / match-name strings / AE version-specific serialization behavior / chunk padding / hidden counters |

### 1.3 Serializer Invariants

| # | Invariant |
|---|---|
| **S-Inv-1** | **Lowering is deterministic**. 等价 runtime graph 在同 capability set 下产 byte-equivalent serializer 输出 |
| **S-Inv-2** | **Serializer may reorder artifacts internally**. chunk ordering 是 implementation-defined，除非 AE compatibility 显式要求 |
| **S-Inv-3** | **Runtime semantics must survive hydration**. hydration 必须从 serializer 输出重建语义等价的 runtime graph |
| **S-Inv-4** | **Lowering must be capability-driven**. version-specific behavior 必须走 `capabilities`，不允许 ad-hoc version check |

### 1.4 Capability admission rule

不是任何版本差异都该进 `AECapabilities`。增 trait **必须同时满足**：

1. **多个 stable serializer lowering 存在**（至少 2 个 AE 版本写出可观测不同字节，且都通过 AE accept）
2. **divergence 不可表达为 runtime semantics**（不是 "AE 24 新增 X 字段" → runtime API 加字段；而是 "AE 25 把 X 编成 4 字节 / AE 2020 编成 0 字节，runtime 看不到区别"）
3. **serializer behavior 必须 branch**（不分支就出错，不是 cosmetic difference）

V2.2 实施期发现的差异 → 走 admission rule 决定要不要进 matrix。**V2.2 ship 时 `AECapabilities` 大概率仍空 struct**。

**Capability resolution direction** (Open question — V2.2 不实现): Deepseek 提议改 auto-derive（打开旧文件时从 chunk 特征推；新建时显式最低版本；写入 forward-only feature 时自动升）。V2.2 接口预留 —— `Capabilities` 是按 `AETarget` 索引的纯函数 lookup，**不直接挂在 Project 上**，避免 V3 改 auto-derive 时还要拆 Project struct。

### 1.5 概念分层 + Vocabulary

```
┌─ Runtime (authoritative truth) ──────────────────────────┐
│ Project / Composition / Layer (interface)                │
│   ↳ ShapeLayer / TextLayer / AVLayer / Camera / Light    │
│ VectorGroup / ShapeNode (Geometry / Render / Modifier)   │
│ PropertyStream<T> / Keyframe / Ease                      │
└──────────────────────────────────────────────────────────┘
       ↓ lowering           ↑ recovery (hydration)
┌─ Serializer ─────────────────────────────────────────────┐
│ chunk synthesis (ldta / tdgp / tdmn / tdb4 / cdat ...)  │
│ capability matrix lookup → AETarget-specific decisions  │
│ archetype seeds (bootstrap substrate)                   │
└──────────────────────────────────────────────────────────┘
       ↓                     ↑
┌─ rifx (existing, unchanged) ─────────────────────────────┐
│ RIFX framing / chunk tree datastructure                 │
└──────────────────────────────────────────────────────────┘
```

V2.2 不动 Go package 结构（仍在 `internal/aep/`），但在文件命名 + 注释里标层。V3 brainstorm 阶段再决定要不要正式拆 `internal/scene/` + `internal/serializer/`。

**Vocabulary 严格**:
- 用 `lowering` （runtime graph lowering / PropertyStream lowering / ShapeNode lowering）
- 不用 `inject / patch / rewrite`（这些是 V1 chunk-patch 时代的词）
- **ShapeNode hierarchy 是 runtime-facing**，**match-name hierarchy 是 serializer-facing** —— 防 ADBE naming leakage 进 runtime API（用户看不到 `"ADBE Vector Shape - Rect"` 字符串；只看到 `aep.ShapeKindRect` 等 Go enum）

### 1.6 文件布局

| 文件 | 层 | 备注 |
|---|---|---|
| `new_layer.go` (新) | runtime | `comp.NewShapeLayer(name) (*ShapeLayer, error)` |
| `shape_graph.go` (新) | runtime | `VectorGroup` / `ShapeNode` interface / `RectNode` / `EllipseNode` / `PathNode` / `FillNode` / `StrokeNode` typed structs |
| `property_stream.go` (新) | runtime | `PropertyStream[T]` typed wrappers + `StreamMode` + `Keyframe[T]` |
| `capability_matrix.go` (新；空骨架) | runtime → serializer 边界 | `AECapabilities` + `Capabilities(target)` 纯函数 lookup |
| `ldta_layout.go` (新) | serializer | ldta byte offset 常量（mirror `cdta_layout.go`）|
| `lower_layer.go` (新) | serializer | runtime ShapeLayer → ldta + tdgp + Slin chunks |
| `lower_shape_node.go` (新) | serializer | ShapeNode → tdmn + tdgp + tdbs + tdb4 chunks |
| `lower_property_stream.go` (新) | serializer | PropertyStream → tdb4 + cdat (static) 或 list[lhd3+ldat] (keyframes) |
| `lower_item_siblings.go` (rename from V2.1) | serializer | Fold-level 8 sibling chunks |
| `parse_layer.go` / `parse_shape.go` / `parse_property.go` (既有) | serializer + hydration | 复用 V1 反向 lowering |
| `shape_layer_test.go` (新) / `shape_graph_roundtrip_test.go` (新) / `shape_preservation_test.go` (新) | test | Tier 1 Go roundtrip + 单元 |
| `new_composition_test.go` (扩展) | test | + `TestV2_2_AEShipGate_AE2025/2020` |
| `test_data/canonical_shape_graph.aep` (新) | test fixture | Tier 2 primary ship gate fixture (Go 端 build 生成) |
| `test_data/v2_2_shape_tolerance.aep` (新) | test fixture | Tier 3 nested group preservation (AE-saved) |
| `test_data/verify_v2_2.jsx` (新) | test driver | Tier 2 AE-side JSX |
| `tmp_debug/gen_shape_dummy.jsx` (新) | RE tool | Phase 1 RE fixture generator |
| `tmp_debug/gen_shape_tolerance.jsx` (新) | RE tool | Tier 3 tolerance fixture generator |

`lower_*.go` 命名是给 V2.2 内部 reusable serializer primitive 划区，未来 V3 可整体迁出 `internal/aep/`。

---

## 2. Runtime types

### 2.1 Layer hierarchy（typed wrapper 模式）

```go
// V1 既有 type，runtime base，所有 V1 setter 保留兼容
type Layer struct {
    Type LayerType
    ID, ParentID, SourceID uint32
    Name string
    BlendingMode BlendingMode
    TrackMatte TrackMatteType
    TrackMatteLayerID uint32
    // ... ~30 既有 fields ...

    ldta *rifx.Chunk  // V2.2 维持 unexported (Inv-1 合规)
}

// V2.2 引入 typed wrapper（只为新建路径，不影响 V1 callers）
type ShapeLayer struct {
    *Layer                  // embed: 所有 V1 setter / getter 自动可用
    rootGroup *VectorGroup  // V2.2 新增 runtime model
}

// V2.3+ 同样形态：
// type TextLayer struct { *Layer; doc *TextDocument }
// type AVLayer struct { *Layer; source *Footage }
// type CameraLayer struct { *Layer }
// type LightLayer struct { *Layer }
```

**为什么不彻底接口化**: V1 已 ship `Composition.Layers []*Layer`。改成 `[]Layer` interface 会 break V1 callers。V2.2 加 typed wrapper 不破坏 V1；V3 brainstorm 决定要不要全面接口化。

**Inv-1 合规**: `Layer.ldta *rifx.Chunk` 字段已 unexported（V1 既有）。新 typed wrappers 不新增 chunk-ref 字段。

### 2.2 VectorGroup + ShapeNode tree

```go
// runtime-facing；match-name 字符串严格 serializer-only
type ShapeNodeKind int

const (
    // Geometry
    ShapeKindRect    ShapeNodeKind = iota // V2.2 ship
    ShapeKindEllipse                       // V2.2 ship
    ShapeKindPath                          // V2.2 ship
    ShapeKindPolyStar                      // V2.3+
    // Render
    ShapeKindFill                          // V2.2 ship
    ShapeKindStroke                        // V2.2 ship
    ShapeKindGradientFill                  // V2.3+
    ShapeKindGradientStroke                // V2.3+
    // Modifier
    ShapeKindTrim                          // V2.3+
    ShapeKindMerge                         // V2.3+
    ShapeKindRepeater                      // V2.3+
    // Structural
    ShapeKindGroup                         // V2.2 internal only（嵌套不开放 synthesis API）
    ShapeKindTransform                     // 见 2.3
)

type ShapeNode interface {
    Kind() ShapeNodeKind
    Properties() *PropertyGroup  // escape hatch β
}

type VectorGroup struct {
    Children  []ShapeNode    // 节点顺序 = AE 渲染顺序（顶到底）
    Transform *PropertyGroup // group-level Transform；V2.2 默认 identity，不暴露 typed setter
}

// V2.2 ship 的 typed structs
type RectNode struct {
    size      *PropertyStream[[2]float64]  // ADBE Vector Rect Size
    position  *PropertyStream[[2]float64]  // ADBE Vector Rect Position
    roundness *PropertyStream[float64]     // ADBE Vector Rect Roundness
}

type EllipseNode struct {
    size, position *PropertyStream[[2]float64]
}

type PathNode struct {
    path *PropertyStream[BezierPath]  // ADBE Vector Shape - Group
}

type FillNode struct {
    color   *PropertyStream[[4]float64]  // RGBA 0..1
    opacity *PropertyStream[float64]     // percent 0..100
}

type StrokeNode struct {
    color   *PropertyStream[[4]float64]
    opacity *PropertyStream[float64]
    width   *PropertyStream[float64]
    // LineCap / LineJoin / Miter / Dashes 等 V2.3+ 走 escape hatch β
}
```

### 2.3 Transform：layer-level + group-level

- **Layer-level Transform** (ldta 内): 所有 Layer 共有 — anchor / position / scale / rotation / opacity
- **Group-level Transform** (tdgp 内 "ADBE Vector Transform Group"): 每个 VectorGroup 自带 — 同 5 字段
- **Shape-node-level Transform**: V2.2 不开放（preservation territory 时 parse / write 不破坏）

V2.2 hot path 走 layer-level：`shapeLayer.SetPosition(...)`（V1 既有，通过 embedded `*Layer`）+ typed `shapeLayer.Transform()` (Section 3.3a)。Group-level Transform 内部保留（runtime 默认 identity；serializer 必须写 chunk 否则 AE 拒），**不暴露 typed setter** —— 走 escape hatch β。

### 2.4 PropertyStream（含 separated dimensions 预留）

```go
type StreamMode int

const (
    StreamModeStatic   StreamMode = iota
    StreamModeAnimated
)

type PropertyStream[T any] struct {
    mode      StreamMode    // Inv-8 state machine
    static    T             // Static mode 时有效；Animated mode 时无效
    keyframes []Keyframe[T] // Animated mode 时非空
    expression string       // empty = 无表达式
}

type Keyframe[T any] struct {
    Time    float64 // Inv-7: seconds only
    Value   T
    InEase, OutEase TemporalEase
    // SpatialIn/OutTangent 仅 multidim 用，V2.2 不开放（默认零）
}

type TemporalEase struct {
    Speed    float64 // value/sec
    Influence float64 // 0..1
}

// 多维 dimension：每维一条 stream + parent PropertyGroup（Deepseek 推荐方案）
type PropertyGroup struct {
    Name     string                       // runtime-facing name
    Children map[string]*PropertyGroup    // 嵌套（Effect param 组复用）
    streams  map[string]any               // 叶子 PropertyStream[T]; T = float64 / [2]float64 / ... — 内部 keyed by runtime name
    // Separated/Unified mode（AE 支持把 Position 拆 X/Y 独立 stream）
    Separated bool
}

// V2.2 typed aliases
type PropertyStream1D    = *PropertyStream[float64]
type PropertyStream2D    = *PropertyStream[[2]float64]
type PropertyStream3D    = *PropertyStream[[3]float64]
type PropertyStreamColor = *PropertyStream[[4]float64]
type PropertyStreamPath  = *PropertyStream[BezierPath]

type BezierPath struct {
    Vertices    [][2]float64
    InTangents  [][2]float64  // V2.2 默认全 0（linear segments）
    OutTangents [][2]float64
    Closed      bool
}
```

**Inv-7 合规**: `Keyframe.Time` 单位**只是秒**。Serializer 通过 capability + comp.TickRate 转 ticks (`lowerCtx.tickRate`)。

**Inv-8 state machine**:
- 初始: Static mode (`static` 字段持值，默认零)
- 第一次 `AddKeyframeLinear` / `AddKeyframeWithEase` → 转 Animated mode，static 字段失效
- `Clear()` → 删 keyframes，回 Static mode（值 = 上次 `SetStaticValue` 或零默认）
- Static mode 下 `SetStaticValue` 改 `static`；Animated mode 下 `SetStaticValue` 报 error

**Separated dimensions**: V2.2 默认 unified（`PropertyGroup.Separated = false`），所有 Position 都是单条 2D/3D stream。Separated 模式留接口，V2.3+ ship。

### 2.5 V2.2 不动的现有 runtime types

- `Composition` / `Project` / `Footage` / `Marker` / `Folder` / `Effect` / `Mask`
- `*Property` （V1 既有 generic property type）：保留；V2.2 新 `PropertyStream[T]` 是它的 typed wrapper / facade
- `ShapePrimitive` / `ShapePath` （V1 parse-side types）：V2.2 可与新 `RectNode` / `PathNode` 等并存一段时间，V2.3 评估统一

### 2.6 Open questions

| # | Question | 影响 | 解决路径 |
|---|---|---|---|
| OQ-1 | Capability auto-derive vs explicit target | V2.2 不做，但 `Capabilities` API 接口已预留 | V3 brainstorm 期决议 |
| OQ-2 | Group-level Transform 默认值 | Phase RE 时 dump 1 个 AE-saved 空 ShapeLayer 看 | RE-S3 |
| OQ-3 | Mutation atomicity 粒度 | V2.2 决议: 沿用 V2.1 per-op atomic | 无需新机制 |
| OQ-4 | Separated PropertyStream 模式 V2.2 实现深度 | V2.2 决议: 留 `PropertyGroup.Separated` 字段，不暴露 setter；V2.3 加 | 无需 |

---

## 3. Public API

> **Public APIs expose runtime semantics only. Chunk structure, match-names, binary offsets, and serializer topology are intentionally hidden.**

### 3.1 Creation entry

```go
func (c *Composition) NewShapeLayer(name string) (*ShapeLayer, error)
// Failure: name empty / lowering 失败 / parse-after-build 产 warning
// On failure, no partial runtime mutation is committed.
```

### 3.2 RootGroup + node attachment

```go
// V2.2 ShapeLayer 默认带 1 个 RootGroup（创建时 lowering 同时初始化空 group）
func (s *ShapeLayer) RootGroup() *VectorGroup

// 5 个节点 typed Add
func (g *VectorGroup) AddRect() (*RectNode, error)
func (g *VectorGroup) AddEllipse() (*EllipseNode, error)
func (g *VectorGroup) AddPath() (*PathNode, error)
func (g *VectorGroup) AddFill() (*FillNode, error)
func (g *VectorGroup) AddStroke() (*StrokeNode, error)
```

**Render order**: **Newly attached nodes are appended to the top of the render stack.** AE UI / `VectorGroup.Children` slice 中，index 越大越靠后 append、render 越靠上（覆盖 lower index）。Children[0] = 最底层。

```go
group := shape.RootGroup()
rect := group.AddRect()       // Children[0]: bottom geometry
fill := group.AddFill()       // Children[1]: fills rect
// → AE: rect 被 fill 染色（fill 在 rect 之上覆盖）
```

### 3.3 Per-node typed setters

```go
// Rect
func (r *RectNode) SetSize(size [2]float64) error
func (r *RectNode) SetPosition(pos [2]float64) error
func (r *RectNode) SetRoundness(v float64) error
func (r *RectNode) Size() *PropertyStream[[2]float64]
func (r *RectNode) Position() *PropertyStream[[2]float64]
func (r *RectNode) Roundness() *PropertyStream[float64]

// Ellipse: SetSize / SetPosition + Size() / Position()
func (e *EllipseNode) SetSize(size [2]float64) error
func (e *EllipseNode) SetPosition(pos [2]float64) error
func (e *EllipseNode) Size() *PropertyStream[[2]float64]
func (e *EllipseNode) Position() *PropertyStream[[2]float64]

// Path
// SetVertices: 仅替换顶点坐标，清切线为 0（V2.2 默认 linear segments），不改 Closed
// 要求 len(verts) >= 2 (RE-S8 确认；可能 Closed=true/false 差异)；lowering 时校
// Closed 默认 true（NewPathNode 时）
func (p *PathNode) SetVertices(verts [][2]float64) error
func (p *PathNode) SetClosed(closed bool) error
func (p *PathNode) Path() *PropertyStream[BezierPath]

// BezierPath is a runtime geometry object, not a serializer encoding mirror.
// Closed flag / Vertices / Tangents semantics are defined by runtime geometry;
// lowering decides how to express that as AE bytes.

// Fill
func (f *FillNode) SetColor(rgba [4]float64) error  // RGBA 0..1
func (f *FillNode) SetOpacity(v float64) error      // percent 0..100, runtime-canonical
func (f *FillNode) Color() *PropertyStream[[4]float64]
func (f *FillNode) Opacity() *PropertyStream[float64]

// Stroke
func (s *StrokeNode) SetColor(rgba [4]float64) error  // RGBA 0..1
func (s *StrokeNode) SetWidth(v float64) error
func (s *StrokeNode) SetOpacity(v float64) error      // percent 0..100, runtime-canonical
func (s *StrokeNode) Color() *PropertyStream[[4]float64]
func (s *StrokeNode) Width() *PropertyStream[float64]
// Opacity() *PropertyStream[float64]
```

### 3.3a ShapeLayer Transform typed surface

```go
// Layer-level Transform (hot path)
func (s *ShapeLayer) Transform() *LayerTransform

type LayerTransform struct{ /* embed PropertyGroup; runtime-facing */ }
func (t *LayerTransform) AnchorPoint() *PropertyStream[[2]float64]
func (t *LayerTransform) Position()    *PropertyStream[[2]float64]  // 2D for V2.2 ShapeLayer
func (t *LayerTransform) Scale()       *PropertyStream[[2]float64]
func (t *LayerTransform) Rotation()    *PropertyStream[float64]
func (t *LayerTransform) Opacity()     *PropertyStream[float64]

// Shorthand
func (s *ShapeLayer) Position() *PropertyStream[[2]float64]
func (s *ShapeLayer) Scale()    *PropertyStream[[2]float64]
func (s *ShapeLayer) Rotation() *PropertyStream[float64]
func (s *ShapeLayer) Opacity()  *PropertyStream[float64]
```

**3D ShapeLayer 推到 V2.3+**: V2.2 `Position()` 永远返回 2D stream。

### 3.4 PropertyStream API

```go
// Static entry
func (ps *PropertyStream[T]) SetStaticValue(v T) error
// Static mode 下改 static 字段；Animated mode 下报 error（callers should Clear() first）

// Keyframe entry (flat functions; KeyframeBuilder 不 ship V2.2)
func (ps *PropertyStream[T]) AddKeyframeLinear(time float64, value T) error
func (ps *PropertyStream[T]) AddKeyframeWithEase(time float64, value T, in, out TemporalEase) error
// 第一次调用时 mode Static → Animated，invalidate static value (Inv-8)
// 时间约束: time >= 0 (Inv runtime layer 不 enforce time <= comp.Duration)
// 重复 time → error

// Mode introspection
func (ps *PropertyStream[T]) Mode() StreamMode
func (ps *PropertyStream[T]) StaticValue() (T, bool)  // (value, isStatic); Animated 时 ok=false
func (ps *PropertyStream[T]) Keyframes() []Keyframe[T]
func (ps *PropertyStream[T]) HasKeyframes() bool

// Bulk
func (ps *PropertyStream[T]) Clear() error  // 删全部 keyframe，回 Static mode
```

**PropertyStream ownership**: Returned `*PropertyStream[T]` handles are stable references owned by the node/group/layer that created them. Holding a reference across mutations is safe (handle stays bound to the same logical stream); deletion of the owning node invalidates the reference (V2.3+ once deletion API ships).

### 3.5 Escape hatch β（generic）

```go
// 通过 PropertyGroup 访问任何 stream（即使 V2.2 typed wrapper 没覆盖）
func (n ShapeNode) Properties() *PropertyGroup
func (g *VectorGroup) Properties() *PropertyGroup
func (s *ShapeLayer) Properties() *PropertyGroup

// PropertyGroup 通过 runtime-facing name 索引子项 / stream
// runtime name 例: "Anchor Point" / "Position" / "Scale" / "Rotation" / "Opacity"
//                  "Size" / "Position" / "Roundness" (Rect)
// match-name (e.g. "ADBE Vector Rect Roundness") 严格 serializer-only，永不暴露
func (pg *PropertyGroup) Child(name string) *PropertyGroup
func (pg *PropertyGroup) Float64Stream(name string) (*PropertyStream[float64], error)
func (pg *PropertyGroup) Vec2Stream(name string)    (*PropertyStream[[2]float64], error)
func (pg *PropertyGroup) Vec3Stream(name string)    (*PropertyStream[[3]float64], error)
func (pg *PropertyGroup) ColorStream(name string)   (*PropertyStream[[4]float64], error)
func (pg *PropertyGroup) PathStream(name string)    (*PropertyStream[BezierPath], error)
// Errors: name 不存在 / name 对应 stream 类型不匹配
```

V2.3+ 加新 T 时增方法。

### 3.6 Defaults（基于 Phase 1 RE 校准）

| 对象 | 字段 | 默认 (provisional, RE confirm 前) |
|---|---|---|
| ShapeLayer | InPoint=0 / OutPoint=comp.Duration / 3D=false / Visible=true / BlendingMode=Normal | V1 NewComposition 既有 + RE-S1 |
| RootGroup | Transform=identity（anchor [0,0], position [0,0], scale [100,100], rotation 0, opacity 100）| RE-S3 |
| RectNode | Size=[100,100] / Position=[0,0] / Roundness=0 | RE-S4 校准前 provisional |
| EllipseNode | Size=[100,100] / Position=[0,0] | RE-S5a 校准前 provisional |
| PathNode | Vertices=empty / Closed=true / Tangents=all 0 | API 要求用户 SetVertices 才能写盘 |
| FillNode | Color=[1,1,1,1] white / Opacity=100 | RE-S5c |
| StrokeNode | Color=[0,0,0,1] black / Width=2 / Opacity=100 | RE-S5d |

Phase 1 RE 校准前 RectNode/EllipseNode default 不进 user-facing docs。

### 3.7 Mutation atomicity

**每个 Add\* / Set\* 是独立 atomic op**（沿用 V2.1 NewComposition pattern）:
- 失败 → no partial runtime mutation committed
- 用户管 multi-step 失败处理（无库内 transaction concept）

不引入 `func (proj *Project) WithMutation(fn func() error) error` 类 transaction API —— V2.2 scope 不需要。

*V3 重新评估 transaction API*: 见 `workshop/specs/v3-direction.md` §M6 Mutation API。

---

## 4. Serializer primitives

### 4.0 Serializer ownership map

| Chunk class | Ownership | 含义 |
|---|---|---|
| ldta (Layer header) | **synthesized** | 每次 lowering 重写；defaults from RE |
| tdgp Transform (layer-level) | **synthesized** | layer Transform group |
| tdgp Vector Transform (group-level) | **synthesized** | RootGroup Transform group |
| tdgp Shape contents | **synthesized** | Rect/Ellipse/Path/Fill/Stroke 的 tdgp |
| tdb4 / cdat / lhd3 / ldat | **synthesized** | PropertyStream lowering 产物 |
| FEE / fvdv / fiop / ftts / foac / fiac / fipc / fifl | **cloned substrate** | 来自 2020_dummy_comp.aep |
| 未识别 tdgp child（plugin / future field） | **preserved** | byte-identical 透传 |
| 未识别 Layr child | **preserved** | byte-identical 透传 |
| svap / nhed / nnhd / head | **patched** | V2.1 既有路径；syncHeadCounters 等 |
| pcms / PwCs / pdvc Utf8 (project metadata) | **preserved** | V1 既有 |
| 全 root-level 其它 chunks (Pefl / gpuG / sfnm / ...) | **preserved** | V1 既有 |

### 4.1 Serializer 责任 + 边界

V2.2 引入 5 个 reusable lowering primitives（全 serializer-side）:

| Primitive | 文件 | 职责 |
|---|---|---|
| Layer lowering | `lower_layer.go` | runtime `*ShapeLayer` → `LIST(Layr)` (ldta + Slin + Transform tdgp + Contents tdgp) |
| Shape node lowering | `lower_shape_node.go` | runtime `ShapeNode` → `LIST(tdgp)` 含 tdmn + sub-streams |
| PropertyStream lowering | `lower_property_stream.go` | runtime `*PropertyStream[T]` → `LIST(tdgp)` 含 tdmn + tdbs + tdb4 + cdat (static) 或 list[lhd3+ldat] (keyframes) |
| Item sibling synthesizer | `lower_item_siblings.go` (rename V2.1) | Fold-level 8 sibling chunks |
| Capability lookup | `capability_matrix.go` | `Capabilities(target AETarget) AECapabilities` 纯函数 lookup |

**Inv-2 合规**: 所有 chunk knowledge 只出现在 `lower_*.go` + 既有 `parse_*.go` + `write*.go` + `rifx/`。

### 4.2 Layer lowering (`lower_layer.go`)

```go
func lowerShapeLayer(s *ShapeLayer, ctx *lowerCtx) (*rifx.Chunk, error)

type lowerCtx struct {
    tickRate     float64
    capabilities AECapabilities
    nextLayerID  func() uint32
    // lowerCtx carries lowering state only. It is not a runtime graph handle.
}

// 输出 LIST(Layr) children 顺序 (per V1 parse_layer + RE-S1 实测 freeze):
//   ldta (160 B for AE 2020 canonical)
//   Utf8 (layer name; length-variable)
//   LIST(tdgp, "ADBE Transform Group")
//   LIST(tdgp, "ADBE Vector Materials Group")  ← shape contents root
//   ... 其它 chunks per RE-S1 dump
```

ldta byte layout：复用 V1 `parse_layer.go` offset 注释 + 新建 `ldta_layout.go`（mirror `cdta_layout.go`）。RE-S1 任务包括 dump empty ShapeLayer ldta 默认字节。

### 4.3 Shape node lowering (`lower_shape_node.go`)

```go
func lowerShapeNode(n ShapeNode, ctx *lowerCtx) (*rifx.Chunk, error)
// 返回完整 LIST(tdgp) chunk，children: tdmn + N × lowerPropertyStream(...)

// match-name 表（serializer-only）
var shapeMatchNames = map[ShapeNodeKind]string{
    ShapeKindRect:    "ADBE Vector Shape - Rect",
    ShapeKindEllipse: "ADBE Vector Shape - Ellipse",
    ShapeKindPath:    "ADBE Vector Shape - Group",
    ShapeKindFill:    "ADBE Vector Graphic - Fill",
    ShapeKindStroke:  "ADBE Vector Graphic - Stroke",
    ShapeKindGroup:   "ADBE Vector Group",
    // V2.3+: ...
}

// 每个 kind 对应 lowering function (switch dispatch)
func lowerRectNode(r *RectNode, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerEllipseNode(e *EllipseNode, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerPathNode(p *PathNode, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerFillNode(f *FillNode, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerStrokeNode(s *StrokeNode, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerVectorGroup(g *VectorGroup, ctx *lowerCtx) (*rifx.Chunk, error)
```

**RootGroup Transform 归属**: `lowerVectorGroup` 内部:
1. lower Transform PropertyGroup → tdgp("ADBE Vector Transform Group")
2. lower each ShapeNode in Children → tdgp(...)
3. wrap into outer tdgp("ADBE Vector Group")

Escape hatch β 通过 `group.Properties().Child("Transform").Vec2Stream("Position")` 改的 Transform 修改通过 group.Transform PropertyGroup 反映。

**lowerVectorGroup 递归深度**: V2.2 只支持 depth = 1 (RootGroup → ShapeNode 直接子项)。嵌套 group 是 V2.3 关注点（届时加 maxDepth 防护 + 循环引用检测）。V2.2 hydration 路径可能从 AE 文件遇到嵌套 group (preservation territory) → 标 unsupported nesting 但**不破坏 byte-for-byte preservation**。

**Inv-4 合规**: 每个节点的子属性（Size / Position / Color / ...）都走 `lowerPropertyStream`，不写自己的 chunk 编码逻辑。

### 4.4 PropertyStream lowering (`lower_property_stream.go`) — V3 核心

```go
// 5 个具体 lowering function（explicit switch，不引 reflect 也不引 codegen）
func lowerFloat64Stream(ps *PropertyStream[float64], match, display string, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerVec2Stream(ps *PropertyStream[[2]float64], match, display string, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerVec3Stream(ps *PropertyStream[[3]float64], match, display string, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerColorStream(ps *PropertyStream[[4]float64], match, display string, ctx *lowerCtx) (*rifx.Chunk, error)
func lowerPathStream(ps *PropertyStream[BezierPath], match, display string, ctx *lowerCtx) (*rifx.Chunk, error)

// PropertyStream lowering is node-agnostic. No `if RectNode` allowed.

// 输出 LIST(tdgp) chunk，children:
//   tdmn (match-name)
//   LIST(tdbs)
//     tdsb (flag bits)
//     tdsn (display name; embeds Utf8)
//     tdb4 (property header: dimensions / numKeyframes / interp mode / ...)
//     IF static:
//       cdat (raw value bytes; per T encoding)
//     IF keyframes:
//       LIST(list)
//         lhd3 (52 B keyframe header: tickRate + numKeyframes + ...)
//         ldat (variable: numKeyframes × per-T keyframe record encoding)
//     tdum / tduM (V1 既有 min/max bound chunks; V2.2 默认全 0)

// T encoding table (serializer-only; Inv-1 不暴露给 runtime)
//   T = float64           → 8 bytes BE
//   T = [2]float64        → 16 bytes BE (x, y)
//   T = [3]float64        → 24 bytes BE (x, y, z)
//   T = [4]float64 (RGBA) → 32 bytes BE (r, g, b, a in 0..1)
//   T = BezierPath        → variable: header(closed flag + vertex count) +
//                            N × 24 bytes (in tangent + vertex + out tangent each [2]f64)
```

**Inv-7 合规**: `Keyframe.Time` (秒) → ticks 转换在 `lhd3 / ldat` 编码处发生（用 `ctx.tickRate`）。

### 4.5 Item sibling synthesizer (formalize V2.1)

```go
func lowerItemSiblings(ctx *lowerCtx) []*rifx.Chunk

// 实现：deep-clone 自 `compTmpl.siblingChunks` （2020_dummy_comp.aep 来源），按 capability 调整：
//   - cap.FEEHasPpSn == false → 清空 FEE LIST children (AE 2020 path)
//   - cap.FEEHasPpSn == true  → 保留 ppSn (AE 2022/25 path)
// V2.1 escape hatch 选 AE 2020 minimum → 当前固定走 no-ppSn 路径
```

**Inv-6 合规**: `compTmpl.siblingChunks` = bootstrap substrate，lowering primitive 只 clone + 字段 patch。

Cross-ref: §4.6 `TargetAE2020` minimum 策略 ↔ 本节 `2020_dummy_comp.aep` substrate 来源 一致。

### 4.6 Capability matrix lookup (`capability_matrix.go`)

```go
type AECapabilities struct {
    // V2.2 开篇 empty；实施期发现就加（受 §1.4 admission rule 限制）
    // First candidates (V3 planning note M7 已知差异):
    //   LdtaSize       int  // 160 (AE 2020/2022) / 164 (AE 25)
    //   FEEHasPpSn     bool // 0 (AE 2020) / 1 (AE 22/25)
    // 当前 escape hatch (AE 2020 canonical minimum) 路径不需 branch，
    // 故 V2.2 ship 时此 struct 可能仍空。
}

// 纯函数 lookup（不挂在 Project 上 —— 为 V3 auto-derive 留接口）
func Capabilities(target AETarget) AECapabilities
```

Cross-ref: §4.5 `2020_dummy_comp.aep` substrate ↔ 本节 `TargetAE2020` minimum 策略 一致。

### 4.7 Hydration (reverse lowering)

V2.2 复用 V1 既有 `parse_layer.go` / `parse_shape.go` / `parse_property.go`，**parse-after-build** pattern (V2.1) 保持：

> **Hydration reconstructs canonical runtime semantics from serializer artifacts.** 不是 rebuild patch structs。

```
NewShapeLayer / AddRect / SetSize ...
        ↓ (mutation)
lower_*.go emit chunks
        ↓
append to comp.itemList.Children
        ↓ reparse closed loop
parseLayer(layr) → *Layer (V1 type)
        ↓ typed wrap + hydrate ShapeNode tree
*ShapeLayer ← wraps *Layer + builds runtime ShapeNode tree from tdgp
```

新增 hydration helper：`hydrateShapeNodes(tdgp *rifx.Chunk) *VectorGroup` —— 从 chunk tree 重建 runtime ShapeNode 树。复用 V1 既有 `parse_shape.go` 的 ShapePrimitive 解码，typed-wrap 进新 RectNode/EllipseNode/PathNode/FillNode/StrokeNode。

**Failure 处理 (parse-after-build closed loop)**:
1. reparse 报 error → rollback (chunk tree restore + typed index + warnings) → 返 internal error
2. reparse 成功但产 warnings → rollback (warnings-as-failure invariant from V2.1) → 返 internal error
3. 两路都从 `oldChildLen` / `oldWarningsLen` snapshot 恢复

跟 V2.1 `NewComposition` 一致。

### 4.8 Phase 1 RE prerequisites

V2.2 实施前必须用 `tmp_debug/gen_shape_dummy.jsx` 跑 AE 2020 + 2022 + 2025，dump 以下：

| RE 任务 | 目标 | 失败影响 |
|---|---|---|
| RE-S1 | empty ShapeLayer 完整 ldta 字节 + Layr LIST 全 children 顺序 + sibling 8 chunks。输出填入 §4.2 chunk 顺序表 | lowering 不知道写啥；**blocking** |
| RE-S2 | empty ShapeLayer Transform tdgp 完整字节（anchor/position/scale/rotation/opacity 默认值 + tdb4 字段）| LayerTransform typed setter defaults；**blocking** |
| RE-S3 | empty VectorGroup tdgp + 内部 Transform group 默认字节 | RootGroup defaults; **blocking** |
| RE-S4 | AE 创建 1 Rect 后 dump: tdmn + tdgp + 子 stream (Size/Position/Roundness) 默认字节 | RectNode defaults / lowering struct; **blocking** |
| RE-S5a | 同上：1 Ellipse | EllipseNode defaults |
| RE-S5b | 同上：1 Path (4 vertex closed) | PathNode defaults |
| RE-S5c | 同上：1 Fill | FillNode defaults |
| RE-S5d | 同上：1 Stroke | StrokeNode defaults |
| RE-S6 | AE 写 1 个 Position keyframe 后 dump：tdb4 + lhd3 + ldat 完整字节 + ease 编码 | PropertyStream keyframe encoding；**blocking** |
| RE-S7 | AE 写 2 个 keyframe 后同上：lhd3 header / ldat record stride | numKeyframes / stride 公式；**blocking** |
| RE-S8 | AE 写 BezierPath（4 顶点 closed）→ dump PathStream tdb4 + ldat。**含 tangent vs no-tangent 对比**：4 顶点 closed 全 0 tangent vs 4 顶点 closed 非 0 tangent，确认 encoding 是否单一格式 | BezierPath encoding；**blocking** |
| RE-S9 | 跨 AE 2020 + AE 25 同 fixture diff | capability admission rule 输入 |

每个 RE task 产 finding 记录（追加到本文档 §8 RE Findings 段）。Findings 推动 lowering 字节级精确。

---

## 5. Test plan

### 5.1 Test pyramid

```
┌─ Tier 1: Go roundtrip (fast, default go test) ─────────────────┐
│ build runtime → lower → write → parse → hydrate → assert       │
│ ~ms per test, no AE dependency                                 │
│ 穷举字段断言（每个 Setter / typed PropertyStream）              │
└────────────────────────────────────────────────────────────────┘
                          ↓ subset 校验
┌─ Tier 2: JSX ship gate (slow, AE_SHIP_GATE=1) ─────────────────┐
│ Go build + WriteAEP → AfterFX -r verify_v2_2.jsx → JSX assert  │
│ AE 2020 + AE 2025 各一次                                       │
│ 验证 AE 真接受 + ScriptingAPI 读回符合 runtime semantic         │
└────────────────────────────────────────────────────────────────┘

┌─ Tier 3: Preservation tolerance (parse/write reopen) ──────────┐
│ AE-saved nested-shape-group fixture → parse → write → reopen   │
│ byte-for-byte 不要求；语义 unchanged + 未识别 chunks byte-同  │
│ 不通过 V2.2 synthesis API (preservation territory)              │
└────────────────────────────────────────────────────────────────┘
```

**失败处理差异**:
- Tier 1 失败 → block merge
- Tier 2 失败 (AE_SHIP_GATE=1) → block merge
- Tier 3 失败:
  - 全版本失败 → block merge
  - 仅某 AE 版本失败 → trigger capability investigation (admission rule 评估)；非 block

### 5.2 Tier 1 — Go roundtrip (`shape_layer_test.go` + 同伴 test files)

**Canonical fixture: (ii) + (C)** — 3 ShapeLayer multi-root。

GPT 关键修正: Layer A 全 animated streams / Layer B/C 全 static streams (不混 Static + Animated 在同 stream)。

```go
func TestV2_2_CanonicalShapeGraph_Roundtrip(t *testing.T) {
    p := aep.NewProject() // TargetAE2020 default
    comp, _ := p.NewComposition("Main", 1920, 1080, 30, 5)
    
    // ShapeLayer A: 全 animated streams
    a, _ := comp.NewShapeLayer("A_RectFill_Animated")
    rectA, _ := a.RootGroup().AddRect()
    rectA.Size().AddKeyframeLinear(0, [2]float64{50, 50})        // Static → Animated mode
    rectA.Size().AddKeyframeLinear(2, [2]float64{300, 200})
    fillA, _ := a.RootGroup().AddFill()
    fillA.Color().AddKeyframeLinear(0, [4]float64{1, 0, 0, 1})  // red → blue 动画
    fillA.Color().AddKeyframeLinear(2, [4]float64{0, 0, 1, 1})
    a.Position().AddKeyframeLinear(0, [2]float64{0, 0})          // Layer Position 2D
    a.Position().AddKeyframeLinear(2, [2]float64{500, 300})
    
    // ShapeLayer B: 全 static streams
    b, _ := comp.NewShapeLayer("B_EllipseStroke_Static")
    ellB, _ := b.RootGroup().AddEllipse()
    ellB.SetSize([2]float64{150, 150})
    strokeB, _ := b.RootGroup().AddStroke()
    strokeB.SetColor([4]float64{0, 0, 1, 1})
    strokeB.SetWidth(5)
    
    // ShapeLayer C: 全 static
    c, _ := comp.NewShapeLayer("C_PathFillStroke_Static")
    pathC, _ := c.RootGroup().AddPath()
    pathC.SetVertices([][2]float64{{0,0}, {100,0}, {100,100}, {0,100}})
    pathC.SetClosed(true)
    fillC, _ := c.RootGroup().AddFill()
    fillC.SetColor([4]float64{0, 1, 0, 1})
    strokeC, _ := c.RootGroup().AddStroke()
    strokeC.SetWidth(2)
    
    // Roundtrip + 穷举断言
    var buf bytes.Buffer
    p.WriteAEP(&buf)
    re, _ := aep.FromReader(bytes.NewReader(buf.Bytes()))
    
    // Layer A: 验证 Animated 字段（用 Keyframes() / Mode()，不调 StaticValue()）
    // Layer B/C: 验证 Static 字段（用 StaticValue() / Mode()）
    // 不混用
}

// 子测试 — 每个 setter 独立 unit
func TestRectNode_SetSize(t *testing.T) { ... }
func TestPathNode_SetVertices_RejectsEmptyOrSingle(t *testing.T) { ... }
func TestPropertyStream_AddKeyframe_TransitionsToAnimatedMode(t *testing.T) { ... }
func TestPropertyStream_SetStaticValue_InAnimatedMode_Error(t *testing.T) { ... }
func TestPropertyStream_Clear_RestoresStaticMode(t *testing.T) { ... }
func TestPropertyStream_NegativeTime_Rejected(t *testing.T) { ... }
func TestPropertyStream_DuplicateTime_Rejected(t *testing.T) { ... }

// Atomicity / rollback
func TestShapeLayer_LoweringFailure_Rollback(t *testing.T) { ... }
func TestPropertyStream_AddKeyframe_InvalidArgs_NoSideEffect(t *testing.T) { ... }
func TestAddRect_ParseWarningAfterLowering_Rollback(t *testing.T) { ... }

// Mixed authored/preserved test (GPT 推荐)
func TestV2_2_MutateExistingShape(t *testing.T) {
    // parse 既有 fixture → mutate runtime field → write → reparse
    // 校验 mutate 字段更新 + 其它字段不变 (preservation)
}
```

期望规模：~30-35 Go test functions。V2.1 122 tests 不动。

### 5.3 Tier 2 — JSX ship gate (`verify_v2_2.jsx`)

**Builder canonical target**: `TargetAE2020` (single canonical serializer subset).
**AE 2020 gate**: validates AE-2020-native acceptance.
**AE 2025 gate**: validates **forward readability** of AE-2020 canonical output (NOT AE-25-native serialization).

JSX per-check logging (mirror V2.1 `verify_v2_1.jsx` pattern):

```javascript
function check(name, cond) {
    if (!cond) { log.push("FAIL: " + name); return false; }
    log.push("OK:   " + name);
    return true;
}

var checks = [
    // Structural
    check("layer count 3", comp.layers.length === 3),
    check("layer A name", layerByName("A_RectFill_Animated") !== null),
    check("layer B name", layerByName("B_EllipseStroke_Static") !== null),
    check("layer C name", layerByName("C_PathFillStroke_Static") !== null),

    // ShapeLayer A (animated)
    check("A contents count 2", countShapeNodes(A) === 2),
    check("A has Rect", hasShapeNode(A, "Rect")),
    check("A has Fill", hasShapeNode(A, "Fill")),
    check("A Rect Size kf count", rectSizeKfCount(A) === 2),
    check("A Rect Size kf[0]", approxEq(rectSizeKfValue(A, 0), [50, 50], 1e-3)),
    check("A Rect Size kf[1]", approxEq(rectSizeKfValue(A, 1), [300, 200], 1e-3)),
    check("A Fill Color kf[0] red", approxEq(fillColorKf(A, 0), [1, 0, 0], 1e-3)),
    check("A Fill Color kf[1] blue", approxEq(fillColorKf(A, 1), [0, 0, 1], 1e-3)),
    check("A Position kf[1]", approxEq(positionKfValue(A, 1), [500, 300], 1e-3)),

    // ShapeLayer B (static)
    check("B Ellipse Size", approxEq(ellipseSize(B), [150, 150], 1e-3)),
    check("B Stroke color blue", approxEq(strokeColor(B), [0, 0, 1], 1e-3)),
    check("B Stroke width 5", strokeWidth(B) === 5),

    // ShapeLayer C (static)
    check("C Path vertex count 4", pathVertexCount(C) === 4),
    check("C Path closed", pathClosed(C) === true),
    check("C Fill color green", approxEq(fillColor(C), [0, 1, 0], 1e-3)),
    check("C Stroke width 2", strokeWidth(C) === 2),
];
ok = checks.every(function(c){ return c; });
// .done 文件含每行 OK/FAIL，调试时一眼定位失败项
```

**不验证项** (per GPT):
- serializer ordering
- hidden default values
- padding artifacts
- internal IDs
- path tangents (V2.2 默认 linear；不 brittle assert)

Go side:

```go
func TestV2_2_AEShipGate_AE2025(t *testing.T) { runV2_2ShipGate(t, aep.TargetAE2025, ae2025exe) }
func TestV2_2_AEShipGate_AE2020(t *testing.T) { runV2_2ShipGate(t, aep.TargetAE2020, ae2020exe) }
// 沿用 V2.1 runAEShipGate 形态；AE_SHIP_GATE=1 才跑
```

### 5.4 Tier 3 — Preservation tolerance fixture (`v2_2_shape_tolerance.aep`)

```
来源: tmp_debug/gen_shape_tolerance.jsx 跑 AE 2025
  创建 1 ShapeLayer，root group 内嵌 1 子 group，子 group 包含 1 Rect + 1 Fill
  保存为 test_data/v2_2_shape_tolerance.aep
```

```go
func TestV2_2_NestedGroup_Preservation(t *testing.T) {
    p, _ := aep.Open("test_data/v2_2_shape_tolerance.aep")
    var buf bytes.Buffer
    p.WriteAEP(&buf)
    reP, _ := aep.FromReader(bytes.NewReader(buf.Bytes()))
    
    // V2.2 minimum preserved semantic set:
    //   - Layer count / names
    //   - ShapeLayer Contents tree topology (含嵌套深度)
    //   - Each leaf ShapeNode kind (Rect / Ellipse / Path / Fill / Stroke / Group)
    //   - Each leaf primitive's hot-path static values (Rect Size / Fill Color / Stroke Width)
    //   - Layer Transform static values (Position / Scale / Rotation / Opacity)
    // 不要求 byte-for-byte; 扩展在 V2.3 nested group API ship 时同步加
}

func TestV2_2_NestedGroup_AEReopens(t *testing.T) {
    // AE_SHIP_GATE=1
    // Go: parse fixture → write tmp.aep → AE 2025 + AE 2020 open it → JSX 校验同 5.3 风格
}

// GPT 强推 (Inv-9 opaque preservation)
func TestV2_2_OpaquePreservation_UnknownChunks(t *testing.T) {
    // 用 fixture 含未来 AE 版本 chunk 或 plugin blob
    // parse → write → reparse
    // 校验：所有 ownership map 标 "preserved" 的 chunks 字节一致
    // 通过 chunk tree walk 对比 raw bytes
}
```

### 5.5 Phase 1 RE fixtures

`tmp_debug/gen_shape_dummy.jsx` 在 AE 2020 / 2022 / 2025 各跑一次，输出：

```
dummy_shape_layer_empty.aep        ← RE-S1/S2/S3 (empty ShapeLayer ldta / Transform group / RootGroup)
dummy_shape_layer_1rect.aep        ← RE-S4 (Rect defaults)
dummy_shape_layer_1ellipse.aep     ← RE-S5a (Ellipse defaults)
dummy_shape_layer_1path_4vtx.aep   ← RE-S5b/S8 (Path defaults + BezierPath encoding)
dummy_shape_layer_1fill.aep        ← RE-S5c (Fill defaults)
dummy_shape_layer_1stroke.aep      ← RE-S5d (Stroke defaults)
dummy_shape_layer_kf_2.aep         ← RE-S6/S7 (PropertyStream keyframe encoding)
dummy_shape_layer_path_tangent.aep ← RE-S8 expand (tangent vs no-tangent encoding compare)
```

每个 fixture 配套 RE finding 段填本文档 §8。

### 5.6 Test data 清单

| 路径 | 类型 | 用途 |
|---|---|---|
| `test_data/canonical_shape_graph.aep` | Go builder 输出 | Tier 2 JSX ship gate primary fixture |
| `test_data/v2_2_shape_tolerance.aep` | AE 2025-saved | Tier 3 nested group preservation |
| `test_data/verify_v2_2.jsx` | JSX | Tier 2 driver |
| `test_data/v2_2_args.json` (runtime-generated) | args | JSX 通信 |
| `tmp_debug/gen_shape_dummy.jsx` | JSX | Phase 1 RE 生成 dummy 系列 |
| `tmp_debug/gen_shape_tolerance.jsx` | JSX | 生成 tolerance fixture |
| (Phase 1 RE fixtures, NOT in git) | byte dumps | RE findings 输入 |

### 5.7 Test coverage 目标

- 5 个新 ShapeNode (Rect / Ellipse / Path / Fill / Stroke) × N typed setter ≈ 20 setter unit
- PropertyStream API (SetStatic / AddKeyframeLinear / AddKeyframeWithEase / Clear / Mode transitions / err paths) ≈ 10 unit
- Atomicity / rollback ≈ 3 unit
- Roundtrip canonical_shape_graph (Tier 1) = 1 integration
- Mixed authored/preserved (Tier 1) = 1-2 integration
- AE ship gate × 2 versions (Tier 2) = 2 tests
- Preservation tolerance (Tier 3) = 2-3 tests

**总 V2.2 加测**: ~35 Go test functions。

**Phase 6 hard ship criterion**: total PASS ≥ **155**（122 baseline + 33 minimum；具体值 Phase 6 commit 前实际计数 freeze）。低于此数视为 V2.2 不完。

### 5.8 不在 V2.2 测试范围

- Golden byte fixture（V2.x RE 期不稳定）
- 性能 / 大文件压力测
- Concurrent mutation
- Path with tangents 完整测（V2.2 lowering 默认全 0 tangent）
- Separated dimensions PropertyStream 测

---

## 6. Classification deliverable (V2.2 留给 V3)

### 6.1 Classification framework

每个 V2.2 涉及的 concept / field / encoding 必归一类：

| 类别 | 判定 |
|---|---|
| **Runtime semantic** | 用户在 runtime API 看得到、操作得到的概念。跨 AE 版本含义一致。Time domain 必须秒。值不依赖 chunk 编码 (per Inv-10) |
| **Serialization artifact** | 仅 serializer 内部存在。runtime API 看不到。值 / 字节 / ordering / encoding 都是 AE 兼容专属 |
| **Bootstrap substrate** | dummy_comp.aep 等提供的最小合法字节起点（Inv-6）。不是 semantic template 也不是 serialization 决策点 |
| **Capability trait candidate** | 跨 AE 版本有差异、且当前由 escape hatch 单一 canonical 收 —— 未被 admission rule 触发，但记录为 V3 候选 |
| **Negative finding / ScriptingAPI quirk** | RE 期间发现的 AE 行为反常，runtime API 不暴露，文档记录 |

### 6.2 Runtime semantic catalog

| Concept | 类型 / API | V2.2 来源 |
|---|---|---|
| `ShapeLayer` | typed wrapper struct embedding `*Layer` | §2.1 / §3.1 |
| `VectorGroup` | runtime tree node + embedded Transform PropertyGroup | §2.2 / §4.3 |
| `ShapeNode` (interface) | runtime node abstraction | §2.2 |
| `RectNode` / `EllipseNode` / `PathNode` / `FillNode` / `StrokeNode` | typed concrete node structs | §3.3 |
| `BezierPath` | runtime geometry object（不是 ldat mirror）| §3.3 |
| `LayerTransform` | typed PropertyGroup wrapper (AnchorPoint/Position/Scale/Rotation/Opacity) | §3.3a |
| `PropertyStream[T]` | typed animation primitive；状态机 (Static ↔ Animated) | §3.4 / Inv-8 |
| `Keyframe[T]` | (Time:seconds, Value:T, In/OutEase) | §3.4 |
| `TemporalEase` | runtime ease abstraction | §3.4 |
| `PropertyGroup` | runtime-name keyed tree | §3.5 |
| `StreamMode` enum (Static / Animated) | mode state | Inv-8 |
| `ShapeNodeKind` enum | Geometry / Render / Modifier / Structural classification | §2.2 |
| Time domain = seconds | Inv-7 | runtime invariant |
| Render order = "append-top" semantic | §3.2 | runtime invariant |
| Mutation atomicity = per-op | §3.7 | V2.1 pattern continuation |
| Layer Position 维度 (V2.2 = 2D) | `*PropertyStream[[2]float64]` | §3.3a 决议（3D 推 V2.3+）|
| `KeyframeBuilder` | **DROPPED** — not in V2.2 | per Section 3 修订 |

### 6.3 Serialization artifact catalog

| Artifact | Owner (lowering primitive) | RE 来源 |
|---|---|---|
| ldta layout (160 B AE 2020 canonical) | `lower_layer.go` + `ldta_layout.go` | RE-S1 |
| LIST(Layr) children 顺序 | `lower_layer.go` | RE-S1 |
| Layer-level Transform tdgp 字节 | `lower_layer.go` | RE-S2 |
| VectorGroup tdgp + RootGroup Transform 默认 | `lower_shape_node.go` | RE-S3 |
| ShapeNode tdgp children 顺序 (tdmn + sub-streams) | `lower_shape_node.go` | RE-S4/S5 |
| Shape match-name 字符串表 | `lower_shape_node.go` 内部 map | serializer-only |
| PropertyStream tdgp children 顺序 (tdmn / tdbs / tdb4 / cdat OR list[lhd3+ldat]) | `lower_property_stream.go` | RE-S6/S7 |
| Per-T cdat 编码 (8/16/24/32 B) | `lower_property_stream.go` switch | RE-S4/S5 |
| BezierPath ldat 编码 (24 B/vertex, in/vertex/out tangents) | `lower_property_stream.go` | RE-S8 |
| lhd3 header (52 B：TickRate / numKeyframes / ...) | `lower_property_stream.go` | RE-S6/S7 |
| ldat record stride (per-T keyframe encoding) | `lower_property_stream.go` | RE-S6/S7 |
| Time domain 秒→ticks 转换 | `lower_property_stream.go` (用 lowerCtx.tickRate) | Inv-7 边界 |
| TemporalEase encoding | `lower_property_stream.go` | RE-S6 |
| Item Fold-level siblings (FEE / fvdv / fiop / ftts / foac / fiac / fipc / fifl) | `lower_item_siblings.go` | V2.1 既知 |
| iide / idpc / idta | V2.1 `new_composition.go` | V2.1 既知 |
| cdta @0x06/0x08/0x10/0x18/0x2C/0x30/0xB8 timing fields | V2.1 既知 + `framerate_canonical.go` | V2.1 scar |
| head[12..15] / [16..19] counter | `write.go syncHeadCounters` | V2.1 既知 |
| svap + nhed + nnhd patching | V2.1 既有 | V2.1 既知 |
| Chunk ordering rules (Inv-2 / S-Inv-2: implementation-defined except AE 兼容要求) | 所有 lower_*.go | meta |

### 6.4 Bootstrap substrate

| Substrate | 文件 | 用途 |
|---|---|---|
| Empty project (AE 2020) | `internal/aep/templates/2020.aep` | V2.1 既有 |
| Empty project (AE 2025) | `internal/aep/templates/2025.aep` | V2.1 既有；svap/nhed 不同 |
| Single-comp dummy (AE 2020 minimum) | `internal/aep/templates/2020_dummy_comp.aep` | V2.1 collapsed |
| Empty ShapeLayer canonical bytes | V2.2 实施期 embed from RE-S1 | `lower_layer.go` 内部 const |
| Empty VectorGroup Transform 字节 | V2.2 实施期 embed from RE-S3 | `lower_shape_node.go` 内部 const |

### 6.4a Unclassified / Pending (RE 期临时区)

RE 期间发现但归类未决的 concept / artifact。允许临时驻留，防止 premature semanticization 或 accidental serializer leakage。每条记录：

```
[pending]
  - 名字 + 简述
  - 发现来源 (Phase 1 RE / ship gate / implementation)
  - pending 原因 (still RE-ing / cross-version data 缺 / 语义 unclear)
  - 决议 deadline (Phase 6 commit 前必清空)
```

**V2.2 ship 前此段必须空。** 所有 pending item 全归入 §6.2 / §6.3 / §6.4 / §6.5 / §6.6 之一。

### 6.5 Capability trait candidates

*Admission rule (定义在 §1.4)*: (1) 跨版本有实测差异；(2) escape hatch 单一 canonical 不能覆盖；(3) lowering 必须 branch。三条同时满足才入 capability matrix。

| Candidate | AE 2020 / 2022 / 2025 | Admission decision (V2.2) |
|---|---|---|
| `LdtaSize` | 160 / 160 / 164 | V2.2 **不入 matrix** (escape hatch：单一 160 canonical；条件 3 不满足) |
| `FEEHasPpSn` | false / true / true | V2.2 **不入 matrix** (同上) |
| `MaterialLightingGroups` | none / none / 10 groups | V2.2 **不入 matrix** (V2.2 不出 Material groups) |
| `TdgpDefaultChildren` | 19 / 19 / 37 | V2.2 **不入 matrix** |
| `ShapeMatchNameVariants` | RE-S9 实测 | V2.2 待 RE-S9 决议 |
| `PathBezierEncoding` | RE-S8 实测 | tangent 编码格式跨版本是否一致；若不一致 → admission |
| `KeyframeEaseEncoding` | RE-S6/S7 实测 | 若跨版本不同 → admission |

**V2.2 ship 时 `AECapabilities` 大概率仍空 struct**。

### 6.6 Negative findings / ScriptingAPI quirks

| Finding | 来源 | 处理 |
|---|---|---|
| NTSC `compItem.shutterAngle` ScriptingAPI 返回 stored × 1.2 | V2.1 scar | runtime API 不修补；JSX gate 不校 NTSC shutter |
| AE drops 0-vertex / 1-vertex Path (推测，待 RE-S8 确认) | RE-S8 任务 | runtime API 拒 (PathNode.SetVertices 校 len >= 2)；待 RE 确认 |
| 同上 V2.2 可能新发现的 quirks | V2.2 RE / ship gate | 单条 finding 一条；scar 段更新 |

### 6.6a Decision log

每个 RE 发现归类时写一行 record（Phase 6 落 commit）：

```
[runtime|serialization|substrate|capability-cand|negative]
  - 名字: X
  - 字节/概念: ...
  - 决策依据: 是否满足 "runtime API 可见 + 跨 AE 含义一致 + 不依赖 chunk encoding"
  - 影响: V3 inherit / V2.x serializer-only / ...
```

### 6.7 V3 inheritance map

V2.2 ship 后，V3 brainstorm 阶段可直接 inherit：

```
Runtime concepts (§6.2): 17 个 typed primitives → V3 scene-graph IR seed
Serialization artifacts (§6.3): 19 个 lowering primitive owner mapping
  → V3 可整批迁出 internal/aep/ 到 internal/serializer/
Bootstrap substrates (§6.4): 5 个 substrate
  → V3 可换格式：V2.x 现状 byte const 嵌 Go 源码（部署一份二进制不依赖外部文件，
    但 RE 不友好——hex dump 找不到，得 grep Go source）；V3 候选外部 .aep 文件
    (RE 友好：直接 dump_root；部署仍 //go:embed)。V2.2 ship 后量化 RE 期手动操作成本，再 V3 决
Capability candidates (§6.5): admission rule template + 7 候选
  → V3 capability matrix 起步集 + auto-derive 接管设计 input
Negative findings (§6.6): 2+ scar records
  → V3 知道哪些 AE 行为不能假设
```

### 6.8 Catalog 维护责任

- **Phase 1 RE 完成后** → Phase 1 subagent 填 §6.3 / §6.4 dump 字段
- **Phase 2+ 实施期** → 实施 subagent 在 commit 前更新 §6.4a (Unclassified)；新发现自动入对应 catalog 段
- **Phase 6 commit 前** → ship subagent finalize：§6.5 admission decision / §6.4a 清空 / §6.6 living section 当前快照
- **V3 brainstorm 启动** → V3 owner 接管 §6.7 inheritance map 推进

---

## 7. Phase 落地路径

V2.2 实施按 phase 拆，每 phase 有 RE / impl / test 三种 task：

```
Phase 0  RE prerequisites blocking
         RE-S1 / S2 / S3 / S4 / S5 / S6 / S7 / S8 / S9
         产出: Phase 1 RE findings 填本文档 §8

Phase 1  Runtime types
         shape_graph.go + property_stream.go + capability_matrix.go (空骨架)
         types_core.go LayerTransform 接口扩展
         Go unit tests for type plumbing

Phase 2  Serializer primitives
         lower_property_stream.go (含 5 个 T 类型 lowering)
         lower_shape_node.go (5 ShapeNode + VectorGroup)
         lower_layer.go (ShapeLayer)
         ldta_layout.go (offset 常量)
         Item sibling synthesizer rename
         
Phase 3  Public API
         new_layer.go (NewShapeLayer entry)
         shape_graph.go AddRect/AddEllipse/AddPath/AddFill/AddStroke
         per-node typed setters
         ShapeLayer.Transform() + shorthand
         Escape hatch (Properties / *Stream)
         Tier 1 unit tests 穷举

Phase 4  Roundtrip + hydration
         hydrateShapeNodes (parse → runtime tree)
         parse-after-build closed loop
         TestV2_2_CanonicalShapeGraph_Roundtrip
         TestV2_2_MutateExistingShape
         Atomicity / rollback unit tests

Phase 5  AE ship gate (Tier 2)
         test_data/verify_v2_2.jsx
         TestV2_2_AEShipGate_AE2020/2025 (AE_SHIP_GATE)
         tmp_debug/gen_shape_tolerance.jsx → v2_2_shape_tolerance.aep
         Tier 3 preservation tests

Phase 6  Docs sync + ship gate
         docs/composition.md + docs/layer.md NewShapeLayer ref
         docs/shape.md ShapeNode hierarchy
         workshop/board.md V2.2 archive entry
         workshop/plans/coverage.md / coverage-detail.md V2.2 段
         §8 RE Findings 段 finalize
         §6.4a Unclassified 清空
         §6.5 admission decision finalize
         scar update (if quirks 新发现)
         Phase 6 hard ship criterion: PASS ≥ 155
```

具体 task 拆分由 writing-plans skill 产出 `workshop/plans/v2-2-layer-creation-plan.md` 接管。

---

## 8. RE Findings (待 Phase 0 / Phase 1 填)

```
本段在 Phase 0 RE 完成后 + Phase 2+ 实施期发现新字段时追加。

模板:
### RE-Sx finding: <title>
- Date: YYYY-MM-DD
- Source: <fixture path>
- Method: <how dumped/measured>
- Bytes: <hex / offset>
- Classification: [runtime|serialization|substrate|capability-cand|negative]
- 影响: ...
```

### RE-S1 finding: empty ShapeLayer Layr LIST 结构 (跨 AE 2020/2022/2025)

- Date: 2026-05-23
- Source: `tmp_debug/re_v22/empty_ae{2020,2022,2025}.aep`
- Method: `tmp_debug/gen_shape_dummy.jsx` (kind=empty) on AE 17.7x45 / 22.6x64 / 25.1x68; structure dumped via `go run ./tmp_debug/dump_root <aep>` → `empty_ae<year>.dump`
- AE versions captured: `app.version` = 17.7x45 (AE 2020) / 22.6x64 (AE 2022) / 25.1x68 (AE 2025)
- **Layr LIST children 顺序 (canonical, 全部三个版本一致)**:
  1. `ldta` — 见 ldta 尺寸表
  2. `Utf8` (2 bytes) — layer name `"S1"` (hex `5331`)
  3. `LIST(tdgp)` — 15 children, 外层 layer property group, 内含 tdsb + tdsn + 6 个 named sub-tdgp + 终止 tdmn
  4. `LIST(Gide)` — 2 children: `gdta` (8B) + `LIST(list)` 内嵌单 `lhd3` (52B)
- 外层 `LIST(tdgp, 15)` 的命名 sub-tdgp 顺序 (tdmn → LIST(tdgp)):
  1. `ADBE Transform Group` → `LIST(tdgp, 15)` — Position_0 / Position_1 / Orientation / RotateX / RotateY / Envir Appear + 各自 tdbs
  2. `ADBE Layer Styles` → `LIST(tdgp, 25)` — Blend Options + dropShadow / innerShadow / outerGlow / innerGlow / bevelEmboss / chromeFX / solidFill / gradientFill / patternFill / frameFX 全部 disabled placeholder
  3. `ADBE Extrsn Options` → `LIST(tdgp, 5)` — Bevel Direction
  4. `ADBE Material Options` → `LIST(tdgp, 17/37)` — 见跨版本 diff 下方
  5. `ADBE Audio Group` → `LIST(tdgp, 3)` — 空
  6. `ADBE Layer Sets` → `LIST(tdgp, 3)` — 空
  7. `ADBE Group End` (terminator tdmn)
- **关键 negative finding**: 当 `shape.addShape()` 被调用但 Root Vectors Group 未追加任何子节点时，**Layr 子树完全不出现 `ADBE Root Vectors Group` 或 `ADBE Vector Materials Group` 的 tdmn**。验证: 在三个 fixture 的整棵 Layr 子树中 grep `566563746f72` (= "Vector") 零命中。空 ShapeLayer 的 binary structure **与一个普通 light/camera 占位层无法区分**，唯一辨识来自 `ldta` 的 layer type 字段。
- **ldta 尺寸跨版本 diff**:
  - AE 2020: 160 bytes
  - AE 2022: 160 bytes (**与 AE 2020 byte-identical**, 整 160B 全等)
  - AE 2025: 164 bytes (160B 前缀与 2020/2022 全等 + 4 bytes 零填充 @ offset `0xA0..0xA3` = `00 00 00 00`)
- ldta 关键字段（基于 head 16B + 整段对照已有 `cdta_layout.go` / `aep_test.go` 注释）：
  - `@0x00` u32 = `0x0000000D` (index = 13, AE 内部为 ShapeLayer 分配的 layer id)
  - `@0x04` u16 = `0x0002` (layer kind bits — Shape Layer 标识；与 Camera/Light/Null 不同)
  - `@0x88` u32 LightKind 字段 (Shape 层为 0x00000004 — 但与 LightKind 同 offset 复用，语义视 layer type 而定)
  - `@0x40` 8B = `53 31 00 00 00 00 00 00` ("S1" + zero pad — ldta-internal 7-bit 名字缓冲，与 Utf8 chunk 重复)
- **V2.2 escape hatch**: 选 **AE 2020 的 160-byte ldta** 为 minimum canonical baseline。AE 2025 多出的尾 4 字节为常量零填充，length-preserving 写入路径无需感知此差异——只要 `addShape` runtime 路径在生成 ShapeLayer 时按"读到的 ldta size 原样保留"即可向上兼容；AE 2025 字段差**不入 capability matrix**。
- Classification: [serialization, negative]
- 影响:
  - `lower_layer.go` ldta layout: 三版本 160B 前缀完全一致, AE 2025 的 4B 尾部按 ldta size 透传即可
  - `LayerCreate` 的 Layr LIST 子节点顺序 **freeze 为 4-tuple**: `ldta + Utf8 + LIST(tdgp,15) + LIST(Gide,2)`
  - 空 ShapeLayer fixture **不需要** 同步预先创建 `ADBE Root Vectors Group` — 只有第一次 `addProperty("ADBE Vector Shape - Rect")` 等 shape contents 操作触发后，AE 才会生成 Vector Materials/Root Vectors Group tdmn 节点（待 RE-S4 实测）
  - V2.2 layer creation runtime 推 ShapeLayer 时, 可借用 dummy_comp template 的 Camera/Light layer 模板 + 修改 ldta @0x04 layer kind 字段即可（绝大多数 property tree 复用）

### RE-S2 finding: Layer Transform group default values

- Date: 2026-05-23
- Source: `tmp_debug/re_v22/empty_ae2020.aep` (RE-S1 produced)
- Method: `go run ./tmp_debug/dump_root tmp_debug/re_v22/empty_ae2020.aep | sed -n '42,90p' > tmp_debug/re_v22/empty_ae2020.transform.dump`（plan 原 awk `/ADBE Vector Materials Group/` 不存在; 改用行号 — RE-S1 已记录 Layr 子树不含 Vector*）。tdb4 / cdat 字节用 `tmp_debug/extract_ldta/main.go`（一次性脚本，已删除）解码：tdb4 @0x03 dim byte 决定 cdat 取多少个 float64 BE。
- **关键发现**: 空 ShapeLayer 的 Layr `ADBE Transform Group` **不是**用户视角的 2D Transform (Anchor / Position / Scale / Rotation / Opacity)，而是 AE 内部为所有 3D 兼容 layer 通用的 6-axis 形式（与 Light / Camera 同 schema）。这与 RE-S1 "空 ShapeLayer binary structure 与 light/camera 占位层无法区分" 结论吻合。用户面 2D 属性大概率在 **ShapeLayer contents tree**（Root Vectors Group 等，待 RE-S4）通过 transform-on-shape 而非 Layr-level 暴露。
- **Transform group 子节点顺序 (`LIST tdgp, 15 children` — tdsb + tdsn + 6 个 named property + Group End)**:

  | idx (在 LIST 中) | tdmn name | sibling LIST | tdb4 dim (@0x03) | cdat size | observed default value(s) |
  |---:|---|---|---:|---:|---|
  | 2  | `ADBE Position_0`              | `tdbs` (6 children: tdsb / tdsn / tdb4 / cdat / tdum / tduM) | 1 | 40 B | `0.0` |
  | 4  | `ADBE Position_1`              | `tdbs` (6 children) | 1 | 40 B | `0.0` |
  | 6  | `ADBE Orientation`             | `otst` (2 children: tdbs + otky) — tdbs 内 tdb4@0x03=0x01 | 1 | 24 B | `0.0` (cdat 24B 实际 = 3 个 float64 BE 全 0；只有第一个属 dim=1 语义值) |
  | 8  | `ADBE Rotate X`                | `tdbs` (4 children: tdsb / tdsn / tdb4 / cdat) | 1 | 40 B | `0.0` |
  | 10 | `ADBE Rotate Y`                | `tdbs` (4 children) | 1 | 40 B | `0.0` |
  | 12 | `ADBE Envir Appear in Reflect` | `tdbs` (4 children) | 1 | 40 B | `1.0` (hex `3ff0000000000000`) |
  | 14 | `ADBE Group End`               | (无 sibling, terminator)               | - | -    | - |

- **缺位的用户面 2D 属性**: `ADBE Anchor Point` / `ADBE Position` (无 _0/_1 后缀的合并版) / `ADBE Scale` / `ADBE Rotate Z` / `ADBE Opacity` 在 Layr Transform group **均未出现**。Position_0 / Position_1 可能是 AE 把 split-Position 写入了 schema (而合并 Position 走另一路径，或当无 keyframe / split disabled 时被 elide)。Rotate Z 不在此 group — 大概率与 AE 的"3D layer always stores X/Y/Z rotation, Z 在 2D 层时通过 user-facing 'Rotation' 别名暴露"有关。
- **cdat 占地 vs 实际值**: 即使 dim=1，AE 把 cdat padding 到 40 B (5 × float64) 或 24 B。剩余 bytes 全 0 — 是 channel storage 上限的预留 (RGBA-style 5 通道？)，不是值。仅前 `dim * 8` 字节有语义。
- **tdb4 字节速读** (本 RE 涉及到的 6 个 property tdb4 head[0..0x10])：
  - Position_0/_1, Rotate X/Y: `db99000100010000 0001ffff00007800` — @0x03=0x01 (1D), @0x05=0x01, @0x07=0x00
  - Orientation: `db99000100070000 00060007 00007800` — @0x03=0x01 但 @0x07=0x07，@0x0E..0x0F = `0007` (与其他不同, 可能编码 'spatial 3D + axis count')
  - Envir Appear in Reflect: `db99000100010000 ffff000400007800` — @0x05=0x00, @0x08..0x09=`ffff`, @0x0A..0x0B=`0004` (不同 mask)
- Classification: [serialization defaults, negative — 缺 user-facing 2D Transform props]
- 影响:
  - `lower_layer.go` 实现 Layer Transform group 时, **不能** 假设它含 Anchor/Position/Scale/Rotate Z/Opacity — 空 ShapeLayer 的 Layr Transform group 是 6-axis 内部 schema。
  - V2.2 ShapeLayer creation **必须**完整 splice 这 6 个 named property + tdsb/tdsn header + Group End；其值用本表观察到的 defaults：5 个 0.0 + Envir 1.0。
  - `ShapeLayer.Transform()` typed setter 的 user-facing API 设计需推迟到 **RE-S4**（Root Vectors Group 实测）— 用户视角的 2D Transform 可能不在 Layr 而在 contents subtree。
  - V1 `Layer.SetPosition/SetScale/SetRotation/SetOpacity` 等 API **不能直接复用**在新生成的空 ShapeLayer 上 — 这些 API 默认 Layr-level Transform 含合并的 2D Position 等，但 RE-S2 证伪。需要 (a) 文档警告 或 (b) typed setter 拒绝在空 contents 状态下写入并要求先创建一个 shape。
  - `cdat` 写入需保留 40B / 24B padding 以维持 length-preserving — 仅前 `dim * 8` 字节是有效 value。

### RE-S3 finding: Root Vectors Group (root) defaults (uses 1rect fixture per RE-S1 negative)

- Date: 2026-05-23
- Source: `tmp_debug/re_v22/1rect_ae2020.aep` (kind=1rect, AE 17.7x45)
- Method: full dump 文件 `1rect_ae2020.full.dump` 行 42-51 切片到 `tmp_debug/re_v22/1rect_ae2020.contents.dump`。fixture 是 RE-S4 同源生成，单 Rect 在 contents tree。
- **Note**: RE-S1 已证 empty ShapeLayer 完全无 Vector* tdmn 出现 (negative)；本 RE 改用 1rect fixture 观察 AE 在有 1 个 Rect 时，Root Vectors Group 实际写的子结构。
- **关键 negative finding**: AE 在添加单 Rect 后**不**新增 "ADBE Vector Materials Group" 或 "ADBE Vector Transform Group" tdmn。整 fixture grep 对应 hex 序列 (`4144424520566563746f72204d6174657269616c73` / `4144424520566563746f72205472616e73666f726d`) 零命中。Plan task 0.4 原假设的 "Materials Group + 子 Transform Group + Anchor/Position/Scale/Rotation/Opacity/Skew/SkewAxis" 完全**未在 binary 出现**。
- **Layr LIST tdgp 17 children** (vs RE-S1 empty 的 15 children) — 多出的 2 children 是 `tdmn(ADBE Root Vectors Group) + LIST(tdgp, 5)`，插入位置在原 15-child sequence 的最前 (`tdsb + tdsn` 之后，`tdmn(ADBE Transform Group)` 之前)。
- **Root Vectors Group binary 实测结构**:
  ```
  chunk tdmn (40 B) head=4144424520526f6f7420566563746f72   # "ADBE Root Vectors Group"
  LIST tdgp (5 children)
    chunk tdsb hex=00000401                                  # subprop flags (与 Layer Transform 的 0x00000001 不同;
                                                              # 高 byte 0x04 疑为 "shape contents container" flag)
    chunk tdsn hex=55746638000000062d5f305f2f2d              # Utf8 len=6 "-_0_/-" (沿用 RE-S2 sibling tdsn 同模式)
    chunk tdmn (40 B) head=4144424520566563746f722053686170   # "ADBE Vector Shape - Rect" (单子节点 Rect)
    LIST tdgp (3 children)                                    # Rect body (见 RE-S4)
    chunk tdmn (40 B) head=414442452047726f757020456e640000   # "ADBE Group End" terminator
  ```
- **5-child 命名 sub-structure** (含 header + terminator): `tdsb + tdsn + tdmn(子项) + LIST(子项 body) + tdmn(Group End)` —— 与一般 named-property tdgp 同 schema (与 RE-S1 Layer Transform LIST tdgp,15 相比，本处只有 1 个 named property=Rect，所以 child count = 2 header + 1 tdmn + 1 LIST + 1 terminator = 5)。
- **缺位 children** (Plan task 0.4 原期望但未观察到):
  - 无 `tdmn "ADBE Vector Group"` (Plan task 0.4 step 1 期望的 root group identifier) — 实际只有 `tdmn "ADBE Root Vectors Group"` 作为 Layr-level sibling，contents container LIST tdgp 内**不再有**任何 group identifier tdmn
  - 无 `LIST tdgp "ADBE Vectors Group"` (children container — 期望嵌套 LIST 实际不存在；Root Vectors Group LIST tdgp 直接装 shape primitives)
  - 无 `LIST tdgp "ADBE Vector Transform Group"` (group-level Transform — 完全不在 binary 中)
  - 无 Anchor Point / Position / Scale / Rotation / Opacity / Skew / Skew Axis 7 个 default-valued sub-property 的 tdmn/tdb4/cdat
- Classification: [serialization defaults, negative — 缺多个 plan 假定的中间 tdgp 层]
- 影响:
  - `lower_shape_node.go` lowerVectorGroup defaults: 不能生成 "ADBE Vector Materials Group" / "ADBE Vector Transform Group" 这两层 — 它们在 AE 默认序列化中**不存在**
  - V2.2 spec §3 / §4 中 "Vector Materials Group" / "Vector Transform Group" 命名需要 reframe — runtime API 概念可能保留 (用户视角的 "group-level transform")，但 serializer 路径必须按 AE 实测 schema: 单层 Root Vectors Group → 直接装 shape primitives
  - 若 user runtime 显式给 VectorGroup 加 sub-Transform (e.g. 用户调 `group.Transform().SetScale(...)`)，serializer 才需要追加 "ADBE Vector Transform Group" subtree (length-variable 路径，与 RE-S4 同向：缺位 = default)

### RE-S4 finding: RectNode defaults (AE 完全 elides default-valued Rect 子属性)

- Date: 2026-05-23
- Source: `tmp_debug/re_v22/1rect_ae2020.aep` (kind=1rect, AE 17.7x45)
- Method: `tmp_debug/gen_shape_dummy.jsx` (kind=1rect) → AE 用 ExtendScript `addProperty("ADBE Vector Shape - Rect")` 添加一个 Rect 到 Root Vectors Group；`go run ./tmp_debug/dump_root` 全量 dump → `1rect_ae2020.full.dump`；Rect 子树切片到 `tmp_debug/re_v22/1rect.dump` (full dump 行 42-51)。
- **关键 negative finding**: 在用户脚本只调用 `addProperty("ADBE Vector Shape - Rect")` 没显式 setValue 的状态下，**AE 不为 Rect 节点写任何 Size / Position / Roundness tdmn / tdb4 / cdat**。整个 fixture 中 grep `4144424520566563746f7220526563` ("ADBE Vector Rec") 零命中——Size/Position/Roundness 三个属性的 match-name 字节序列**完全不出现**。
- **Rect 节点 binary 实测结构** (Root Vectors Group LIST tdgp,5 children → Rect LIST tdgp,3 children):
  ```
  chunk tdmn (40 B) head=4144424520566563746f722053686170   # "ADBE Vector Shape - Rect"
  LIST tdgp (3 children)
    chunk tdsb hex=00000001                                  # subprop flags
    chunk tdsn hex=557466380000000ee79fa9e5bda2e8b7afe5be842031
                                                              # Utf8 len=0x0e localized name "矩形路径 1"
    chunk tdmn (40 B) head=414442452047726f757020456e640000   # "ADBE Group End"
  ```
- **没有任何子属性 tdmn (Size/Position/Roundness)**: Rect LIST tdgp 只含 header (tdsb + tdsn) + terminator (Group End)，**没有 tdmn "ADBE Vector Rect Size" / "ADBE Vector Rect Position" / "ADBE Vector Rect Roundness"** 出现。
- **观察到的 defaults** (从 binary 缺位 + match-name table 反推): Size / Position / Roundness 无任何 binary 表达 — 这些值的 default 必须**由 parser 在缺位时填入**。
  - **Default Size**: `[?, ?]` — fixture 不解 (AE 用 ExtendScript 默认 100×100，可在 host code 校准时通过 boltframe `defaults.go` 或 AE schema 反查；本 RE 不能从 binary 决定)
  - **Default Position**: `[?, ?]` — fixture 不解 (默认为 comp center [0,0]，与 RE-S2 Layer Transform Position_0/_1 一致;非合并 Position)
  - **Default Roundness**: `?` — fixture 不解 (默认 0.0)
  - tdb4 dimensions (Size=2, Position=2, Roundness=1) 与 RE 无关——本 fixture 无 tdb4/cdat 出现
- **重要 implication**: V2.2 ShapeLayer creation 路径**不能**靠 length-preserving splice 来插入 Rect 默认属性子树——AE serialization 本身就不写这些字节。要么 (a) 在新建 Rect 时显式 splice 缺位的 tdmn + LIST tdbs + tdb4 + cdat 节点 (length-variable 路径)，要么 (b) 与 AE 同向：runtime 创建时 Rect node 子属性 tree 全空，由 parser default-fill。
- 仅当用户对某个子属性显式 `setValue` 或加 keyframe 时，AE 才会在 Rect LIST tdgp 内追加该子属性的 tdmn + LIST tdbs 节点 (待 RE-S5 / RE-S6 实测确认追加点 vs 替换点)。
- Classification: [serialization defaults, negative]
- 影响:
  - `lower_shape_node.go` lowerRectNode runtime → serializer 路径必须能输出**两种**形态: 全 default 时输出 3-child empty LIST tdgp; 任一子属性被 explicit set 时按需追加对应 tdmn + tdbs subtree
  - V2.2 spec §3.6 defaults table 校准: Rect 子属性 defaults 不能从 RE 取，需查 boltframe 源或 AE schema（运行时 default 表）
  - Parser default 填充逻辑（如果有）需对齐 AE 缺位语义；当前 `internal/aep/parse_shape.go` 的行为待 cross-check (本 RE 不动 parser)

### RE-S5a finding: EllipseNode defaults + Size/Position encoding

- Date: 2026-05-23
- Source:
  - `tmp_debug/re_v22/1ellipse_ae2020.aep` (kind=1ellipse, AE 17.7x45) — base (无 setValue)
  - `tmp_debug/re_v22/1ellipse_set_ae2020.aep` (kind=1ellipse_set) — Size=[120,80], Position=[10,20] explicit setValue
- Method: `gen_shape_dummy.jsx` _set 分支 + `tmp_debug/dump_cdat_seq` 工具 (一次性 Go scaffold, 已 add) 定位 Ellipse 子树 + 解码 cdat float64 BE
- **matchNames 实测** (from `addProperty()` 后 child enumeration; AE 自动创建)：
  - `ADBE Vector Shape Direction` (child[1], 1D, scalar — direction flag)
  - `ADBE Vector Ellipse Size` (child[2], dim@03=2)
  - `ADBE Vector Ellipse Position` (child[3], dim@03=2)
- **Base 默认值情况** (1ellipse): Ellipse LIST tdgp 仅 **3 children** (`tdsb + tdsn + tdmn(Group End)`)，完全无任何子属性 tdmn。**elision pattern 与 RE-S4 一致** —— AE 在 default 状态下不写 Direction / Size / Position 任何字节。
- **Set 时实测** (1ellipse_set): Ellipse LIST tdgp **7 children** = header(tdsb + tdsn) + 2 × (tdmn + tdbs LIST) + Group End。**Direction 仍被 elide**（仅 Size 和 Position 出现，因为只对这两个 setValue）。
- **Size encoding** (`ADBE Vector Ellipse Size` setValue=[120,80]):
  - tdb4 (124B) `db99000200010000ffffffff00007800`，@0x03=`02`(dim=2)
  - cdat **80B** = 10 × float64 BE = `[120.0, 80.0, 0, 0, 0, 0, 0, 0, 0, 0]`
  - hex 前 16B: `405e000000000000 4054000000000000`
  - 仅前 `dim*8 = 16B` 是 value，后续 64B = 8 个 padding float64 0 (channel storage 上限预留)
  - tdbs 还含 `tdum`/`tduM` siblings (min/max=`±0xc0df_400000000000`= ±32000.0 等？ — 用户范围 hint, 与 cdat encoding 无关)
- **Position encoding** (`ADBE Vector Ellipse Position` setValue=[10,20]):
  - tdb4 (124B) `db990002000f0003ffffffff00007800`，@0x03=`02`(dim=2)；@0x05=`0f`、@0x06-0x07=`00 03` 与 Size 不同 (`00 01 0000`) — 可能是 'spatial 2D position' marker (与 Layer Transform Position_0/_1 dim=1 但 flag 不同 fashion)
  - cdat **48B** = 6 × float64 BE = `[10.0, 20.0, 0, 0, 0, 0]` (padding 32B = 4 个 0)
  - hex 前 16B: `4024000000000000 4034000000000000`
  - 注意：Position 的 tdbs **4 children** (无 tdum/tduM)，Size 的 tdbs **6 children** (含 tdum/tduM) —— 标量 hint range chunks 是 per-property 决策
- **跨子属性 cdat padding 不一致**: Size 80B (10 float64) vs Position 48B (6 float64) — padding 大小是 per-property 决策, 不是固定 5-channel
- Classification: [serialization encoding]
- 影响: `lower_shape_node.go` lowerEllipseNode 编码路径 (length-variable 当任一 sub-property 显式被 set 时)

### RE-S5b finding: PathNode + BezierPath linear encoding

- Date: 2026-05-23
- Source: `tmp_debug/re_v22/1path_ae2020.aep` (kind=1path_4vtx, AE 17.7x45)
- Method: ExtendScript `new Shape()` 4 顶点 closed 全 0 tangent → AE 保存；dump 工具 `tmp_debug/dump_cdat_seq` + 一次性 `tmp_debug/dump_path_bytes` + `tmp_debug/dump_ldat_f32` (本 RE 内 ad-hoc, 已 add)
- **Path 节点 binary 实测结构** (`ADBE Vector Shape - Group` LIST tdgp 5 children):
  ```
  chunk tdsb hex=00000001
  chunk tdsn hex=5574663800000008e8b7afe5be842031   # Utf8 len=8 "路径 1"
  chunk tdmn name="ADBE Vector Shape"
  LIST om-s (2 children)
    LIST tdbs (4 children: tdsb + tdsn + tdb4 + cdat)
      tdb4 (124B) dim@03=1 head=db990001000700010002000700007800
      cdat (4B) hex=00000000              # 仅 4 字节, 非 float64
    LIST omks (1 child)
      LIST shap (3 children)
        chunk shph (24B)                  # path header + bbox
        LIST list (2 children: lhd3 + ldat)
          lhd3 (52B)
          ldat (96B)
        chunk omtn (0B)
  chunk tdmn name="ADBE Group End"
  ```
- **关键 negative finding**: `ADBE Vector Shape` 不走 tdbs/tdb4/cdat (float64) 路径，而是 `om-s` LIST 包裹 — 与 Mask path encoding 同 schema (`parse_mask.go` 注释)。tdbs sibling 内的 `cdat` 仅 **4B = `00000000`**，看似是 enable/flag byte 而非数值，与 RE-S2 cdat 40B/24B 完全不同 layout。
- **shph (24B header)** `b3de0201_00000000_00000000_42c80000_42c80000_01000000`:
  - bytes 0-1: `b3de` magic
  - bytes 2-3: `0201` flags (lowest bit = closed?；high byte 02 待 RE-S8)
  - bytes 4-11: 0 (likely bbox min = [0,0] in f32 BE)
  - bytes 12-15: `42c80000` = float32 BE = 100.0 (bbox max x)
  - bytes 16-19: `42c80000` = 100.0 (bbox max y)
  - bytes 20-23: `01000000` (trailer flag)
- **lhd3 (52B header)** `00d00bee_00000000_0000000c_00000004_00000008_00000004_00000001_00000010_00000000 ...`:
  - bytes 0-3: `00d00bee` magic
  - bytes 8-11: u32 BE = 12 (待解；可能是 "stride bytes" 或 "value count")
  - bytes 12-15: u32 BE = 4 = **vertex count**
  - bytes 16-19: u32 BE = 8
  - bytes 20-23: u32 BE = 4
  - bytes 24-27: u32 BE = 1 (closed?)
- **ldat (96B)** = 24 × float32 BE = **4 vertices × 6 f32 each** (与 `parse_mask.go` "3 consecutive float32 X/Y pairs: anchor, in-tangent, out-tangent" 一致)。**关键 negative**: 解码后值非 raw `[0,0],[100,0],[100,100],[0,100]`，而是 **bbox-normalized** 0/1 pattern：
  - vert 0: f32[0..5] = `(0, 0, 0, 0, 1, 0)` — anchor=(0,0) in=(0,0) out=(1,0)?
  - vert 1: f32[6..11] = `(1, 0, 1, 0, 1, 1)` — anchor=(1,0) out shifts
  - vert 2: f32[12..17] = `(1, 1, 1, 1, 0, 1)`
  - vert 3: f32[18..23] = `(0, 1, 0, 1, 0, 0)`
  - 注意：tangent 全 0 input 但 f32 解码出现 1.0 — 待 RE-S8 (tangent vs no-tangent) 联合复核才能精准 attribute 每个 f32 slot 的语义；本 RE 仅记观察。
- **Per-vertex stride**: 24B (6 × f32)。对 zero-tangent case：tangent bytes **不被 omit**, 占用完整 stride (但写的 1.0 而非 0.0 — 待 RE-S8 决议是否 normalization 副产物)
- Classification: [serialization encoding]
- 影响: `lower_property_stream.go` lowerPathStream encoding (linear case) 需走 om-s + omks + shap + list(lhd3 + ldat f32) 路径，不复用 tdbs/cdat float64 path
- 待 RE-S8 决议: tangent vs no-tangent 是否同一 encoding format / ldat f32 值是否 bbox-normalized 还是 raw coords / lhd3 头部各 u32 字段语义

### RE-S5c finding: FillNode defaults + Color/Opacity encoding (Color setValue silent fail)

- Date: 2026-05-23
- Source:
  - `tmp_debug/re_v22/1fill_ae2020.aep` (kind=1fill) — base
  - `tmp_debug/re_v22/1fill_set_ae2020.aep` (kind=1fill_set) — Color=[1,0,0,1], Opacity=75 setValue
- Method: `gen_shape_dummy.jsx` _set 分支 + `tmp_debug/dump_cdat_seq`
- **matchNames 实测** (Fill addProperty 后 5 children)：
  - `ADBE Vector Blend Mode` (child[1])
  - `ADBE Vector Composite Order` (child[2])
  - `ADBE Vector Fill Rule` (child[3])
  - `ADBE Vector Fill Color` (child[4])
  - `ADBE Vector Fill Opacity` (child[5])
- **Base 默认值情况** (1fill): Fill LIST tdgp **3 children** — elision 与 RE-S4 / RE-S5a 一致
- **Set 时实测** (1fill_set): Fill LIST tdgp **5 children**：
  - 只有 `ADBE Vector Fill Opacity` 子属性出现 (Opacity=75 持久化成功)
  - **`ADBE Vector Fill Color` 设置静默失败** — `setValue([1,0,0,1])` 调用成功 (script log 显示 "set Color [1,0,0,1]")，但 binary fixture 中**无 `ADBE Vector Fill Color` tdmn**。
  - **possible causes**:
    1. ExtendScript Color setValue 在 AE 17.7 期望 0-1 范围但格式不对（无 error throw）
    2. AE 当 setValue 等于 default 值时 elide（但 [1,0,0,1] != default white [1,1,1,1]）
    3. Fill default Color 实际是 red [1,0,0]，setValue 与 default 一致触发 elision
  - 优先猜测 (3) — 见 Stroke (RE-S5d) Color 持久化成功且 setValue [0,0,1] 与 default black 不同
- **Opacity encoding** (`ADBE Vector Fill Opacity` setValue=75):
  - tdb4 (124B) `db99000100010000ffffffff00007800`，@0x03=`01`(dim=1)
  - cdat **40B** = 5 × float64 BE = `[75.0, 0, 0, 0, 0]`
  - hex 前 8B: `4052c00000000000`
  - tdbs sibling 含 tdum=`00000000_00000000` (0.0) / tduM=`4059000000000000` (100.0) hint range
- **Color encoding**: 本 RE 未能观察 (setValue 未持久化)；推迟到下一轮 RE w/ setValue=非 default 颜色 (e.g. [0.5, 0.5, 0.5, 1]) 或参考 RE-S5d Stroke Color
- Classification: [serialization encoding]
- 影响:
  - `lower_shape_node.go` lowerFillNode + Opacity 走标准 tdbs/tdb4/cdat float64 path
  - Color 编码同 RE-S5d Stroke Color (12 float64 cdat, dim=4) — 推迟实证
  - V2.2 spec §3.6 defaults table 提案的 Fill default `Color=[1,1,1,1] white` **可能错误** — 待与 boltframe / AE schema 复核 (若实际默认是 [1,0,0,1] 红色则修正)

### RE-S5d finding: StrokeNode defaults + Color/Width/Opacity encoding

- Date: 2026-05-23
- Source:
  - `tmp_debug/re_v22/1stroke_ae2020.aep` (kind=1stroke) — base
  - `tmp_debug/re_v22/1stroke_set_ae2020.aep` (kind=1stroke_set) — Color=[0,0,1,1], Width=5, Opacity=80 setValue
- Method: `gen_shape_dummy.jsx` _set 分支 + `tmp_debug/dump_cdat_seq`
- **matchNames 实测** (Stroke addProperty 后 11 children)：
  - `ADBE Vector Blend Mode` (child[1])
  - `ADBE Vector Composite Order` (child[2])
  - `ADBE Vector Stroke Color` (child[3])
  - `ADBE Vector Stroke Opacity` (child[4])
  - `ADBE Vector Stroke Width` (child[5])
  - `ADBE Vector Stroke Line Cap` (child[6])
  - `ADBE Vector Stroke Line Join` (child[7])
  - `ADBE Vector Stroke Miter Limit` (child[8])
  - `ADBE Vector Stroke Dashes` (child[9])
  - `ADBE Vector Stroke Taper` (child[10])
  - `ADBE Vector Stroke Wave` (child[11])
- **Base 默认值情况** (1stroke): Stroke LIST tdgp **9 children** — **不是** 3 child! 与 RE-S4 (Rect) / RE-S5a (Ellipse) 的纯 elision 不同。base 子树为：
  ```
  tdsb + tdsn + (3 个 named group tdmn + 3-child empty LIST tdgp) × 3 + Group End
  ```
  即 `Dashes` / `Taper` / `Wave` **三个 nested-group sub-properties 即使默认状态也保留 tdmn + 空 LIST tdgp** (3-child header-only group)。这与 scalar/vector default elision **行为不同**。其余 8 个 scalar/vector default sub-properties (Blend Mode / Composite Order / Color / Opacity / Width / Line Cap / Line Join / Miter Limit) 全 elide。
- **Set 时实测** (1stroke_set): Stroke LIST tdgp **15 children** = base 9 + 6 (3 个新增 named property = Color/Opacity/Width 各 2 children)。Color/Opacity/Width 三个全部持久化成功。
- **Color encoding** (`ADBE Vector Stroke Color` setValue=[0,0,1,1]):
  - tdb4 (124B) `db990004000700000002ffff00007800`，@0x03=`04`(dim=4)
  - cdat **96B** = 12 × float64 BE
  - 解码: `[255.0, 0, 0, 0, 0, 0, 255.0, 0, 0, 0, 0, 0]` (offsets 0、48 处的 255.0)
  - hex: `406fe00000000000 0000000000000000 0000000000000000 0000000000000000 0000000000000000 0000000000000000 406fe00000000000 0000000000000000 ...rest zeros`
  - **关键观察**: setValue input `[R=0,G=0,B=1,A=1]` (0-1 range)。saved 12-float cdat 中前 4 个 (dim=4 part) = `[255, 0, 0, 0]` — 这看似不直接是 RGBA 缩放，但前 4 + 第 7 位 = 255。
  - **可能 encoding**: 0-255 scaled + component order ≠ RGBA。Hypothesis: AE 实际写 `[A*255, R*255, G*255, B*255]` (ARGB) 对 dim=4 head, 第 7 个 float (offset 48 = 6×8) = 第 2 行起点 = 重复或备用通道。
  - 验证 ARGB hypothesis: `[A=1, R=0, G=0, B=1] × 255 = [255, 0, 0, 255]` — 但 saved 是 `[255, 0, 0, 0, 0, 0, 255, 0, ...]`, 第 4 位是 0 不是 255 — 不完全匹配
  - **更精确 layout 待 RE 跟进** (e.g. 加 1stroke_set_red 设 [1,0,0,0.5] 半透明红色对比)
- **Width encoding** (`ADBE Vector Stroke Width` setValue=5):
  - tdb4 (124B) `db99000100010000ffffffff00007800`，@0x03=`01`(dim=1)
  - cdat **40B** = 5 × float64 BE = `[5.0, 0, 0, 0, 0]`
  - hex 前 8B: `4014000000000000`
  - tdbs 6 children 含 tdum=0.0 / tduM=100.0
- **Opacity encoding** (`ADBE Vector Stroke Opacity` setValue=80):
  - tdb4 (124B) `db99000100010000ffffffff00007800` — 与 Width 同 layout
  - cdat **40B** = 5 × float64 BE = `[80.0, 0, 0, 0, 0]`
  - hex 前 8B: `4054000000000000`
- **其它 Stroke sub-properties 状态** (base 与 _set 同):
  - Line Cap / Line Join / Miter Limit: 默认 elide (无 tdmn)
  - Dashes / Taper / Wave: 默认即保留为 3-child empty LIST tdgp (header-only, non-elided container)
- Classification: [serialization encoding]
- 影响:
  - `lower_shape_node.go` lowerStrokeNode：**两种 elision 模式并存**
    - scalar/vector sub-property: 完全 elide (Rect / Ellipse pattern)
    - nested-group sub-property (Dashes/Taper/Wave): 始终保留 3-child empty LIST tdgp
  - Color cdat 12 float64 (dim=4 + padding) 的精确 component order 待复核 — V2.2 lowering 可先按"4 × float64 BE in 0-255 range, 顺序待定"实现 + 用 typed setter 单元测试 round-trip 验证 AE 行为
  - V2.2 spec §3.6 提案 `StrokeNode Color=[0,0,0,1] black / Width=2 / Opacity=100` defaults 合理（与 base elision 一致）

### RE-S6 finding: PropertyStream 1-keyframe encoding (Layer Position 2D)

- Date: 2026-05-22
- Source: `tmp_debug/re_v22/kf_1_ae2020.aep` (kind=kf_1) — ShapeLayer + 1 Rect + Layer Position setValueAtTime(0, [960,540])
- Method: `tmp_debug/dump_kf` (新工具，dump first tdbs→list→{lhd3, ldat}); cross-check 用 V1 既有 `parse_keyframe.go`
- **lhd3 (52 B)** hex `00d00bee000000000000000100000001000000800000000400000001000000040000000000000000000000000000000000000000`:
  - `[0x00..0x03]` magic `00d00bee` (BE u32 = 13634542)
  - `[0x04..0x07]` `00000000` (reserved / 0)
  - `[0x08..0x0B]` **numKeyframes** = u32 BE = 1 (与 V1 `parse_keyframe.go` 既定 layout 一致)
  - `[0x0C..0x0F]` `00000001` (常量？; 待跨 fixture 复核)
  - `[0x10..0x13]` **bytesPerKeyframe (bpk)** = u32 BE = 128 (= 0x80; V1 parser 既定)
  - `[0x14..0x17]` `00000004` (待解；可能 = dims*2)
  - `[0x18..0x1B]` `00000001`、`[0x1C..0x1F]` `00000004` (待解；可能 spatial-style flag + bpk re-hint)
  - `[0x20..0x33]` zeros (padding)
  - **关键 negative**: TickRate **不在 lhd3**；TickRate 来自 comp's `cdta @0x08` (`parseCtx.tickRate`) — V1 `parseKeyframes` 也是用 ctx.tickRate 推算 time
- **ldat per-keyframe (bpk=128 B; layout = spatial-style for Position 2D)**:
  - `[0x00..0x03]` **time** = u32 BE in comp ticks → seconds = ticks / tickRate (验证: kf_1 time=0 → 0/30720 = 0.0s ✓)
  - `[0x04]` inInterp byte (0x01=linear, 0x02=bezier, 0x03=hold) = `0x01`
  - `[0x05]` outInterp byte = `0x01`
  - `[0x06]` flags = `0x00`
  - `[0x07]` **header07** = `0x07` (spatial-style 标记; layoutFor returns valueOff=0x38, spatialStyle=true)
  - `[0x08..0x37]` reserved + temporal ease (per V1 parse_keyframe layout):
    - `[0x18..0x1F]` inSpd (f64 BE) `[0x20..0x27]` inInf
    - `[0x28..0x2F]` outSpd `[0x30..0x37]` outInf
  - `[0x38..0x47]` **value** = 2 × f64 BE = (960, 540) ✓ 与 setValueAtTime 输入一致
  - `[0x48..0x57]` in-spatial-tangent = 2 × f64 BE = (0, 0)
  - `[0x58..0x67]` out-spatial-tangent = 2 × f64 BE = (0, 0)  
    - 实测出现 `8000000000000000` (negative-zero) 不是正零 — 不影响数值，但是 byte-for-byte 重写需保留
  - `[0x68..0x7F]` padding zeros
- Classification: [serialization encoding] — 与 V1 既有 `parse_keyframe.go` 既定 layout 完全吻合；本 RE 主要作 V2.2 正向 lowering 路径的 freeze
- 影响: `lower_property_stream.go` lowerVec2Stream 走 spatial-style 编码 (bpk=0x80, header07=0x07, valueOff=0x38)；时间编码用 ctx.tickRate × seconds (V1 `write_keyframe.go` 既有路径)

---

## 9. 关联文档

- 反馈源: `workshop/feedback/{gpt,deepseek,claude}`
- V2.1 实践基础: `workshop/scars/ae25-acceptance-gate.md`
- V2.1 plan: `workshop/plans/v2-1-foundation-plan.md`
- V3 direction: `workshop/specs/v3-direction.md`
- 当前架构: `workshop/specs/architecture.md`
- 项目 board: `workshop/board.md`
