# V3 direction — scene-graph IR + capability matrix + serializer split

**Status**: Planning note, not a plan. No implementation yet. Drafted 2026-05-22 after V2.1 ship.

GPT 反馈整合 (`flightdeck/kneeboard/gpt`) + V2.1 Phase 6 实践经验 (`scars/ae25-acceptance-gate.md`)。

## 动机：V2.1 暴露的架构债

V2.1 写 `NewProject + NewComposition` 时，5 阶段 AE 接受 gate 一路下来发现：

1. **parser ≠ runtime model**: parser 读 cdta 16 个字段（W/H/fps/duration/bg/shutter…）；AE 实际依赖更多（@0x06 ticks_per_frame / @0x2C masterTicks / @0xB8 duration_mirror / Item Fold-level siblings 8 个 / head counters）。chunk tree 不是 source-of-truth，AE 内部 runtime 模型才是。
2. **per-version 差异是 serialization 关心，不是 runtime 关心**: AE 25 vs AE 2020 在 Item 字节级有 10 个 Material property groups / ldta 160→164B / FEE 0→1 children 差异。但这些都是 **back-compat readable** —— 用 AE 2020 minimum 子集 + AE 自己补默认就跨版本工作。"capability matrix" 比 "per-version dispatch" 自然。
3. **template-copy 难推广**: V2.1 靠 1 个 dummy_comp.aep 模板 + 字段 patch。V2.2 加 Layer / V2.3 加 Effect / V2.4 加 Shape 时，每种 archetype 一个模板 → template proliferation。Layer 类型 ~10+；Shape primitive ~5；Effect ~1000+。不可行。

V2 path: chunk-patch + template-copy 把 NewComposition 推到 ship；再往下走需要 runtime IR。

## 概念分层

Phase 7 文档已开始两层语言:

### Logical (runtime) layer

```
Project
├── Comp
│   ├── Layer (Shape / Text / AV / Camera / Light / Solid)
│   │   ├── PropertyStream (per-property: dimensions / keyframes / interpolation / expression)
│   │   ├── EffectInstance (matchName + params + streams)
│   │   ├── MaskGroup
│   │   └── …
│   ├── Marker
│   └── …
├── Footage
└── Folder
```

### Serialization layer

```
RIFX
├── head / nhed / nnhd / svap (project headers)
├── LIST Fold
│   ├── fdta
│   ├── LIST Item (cdta + DLay + PRin + …)
│   ├── FEE + fvdv + … (8 item-attr siblings)
│   └── …
└── chunk-level data: cdta / ldta / tdgp / tdbs / tdb4 / cdat / lhd3 / ldat …
```

**Boundary**: logical layer只 reads/writes runtime objects. serializer 负责 chunk-tree synthesis + parsing. 目前两者混在 `internal/aep/`，V3 应分包.

## V3 modules (proposal, no impl)

### M1. Runtime Object Model

`internal/aep/runtime/` (or `internal/scene/`)：

```go
type Project struct { /* logical only — no chunk refs */ }
type Composition struct { Layers []*Layer; … }
type Layer struct {
    Kind  LayerKind  // ShapeLayer / TextLayer / AVLayer / CameraLayer / LightLayer / SolidLayer
    Streams map[string]*PropertyStream
    Effects []*EffectInstance
    Masks   []*MaskGroup
    …
}
type PropertyStream struct {
    Dimensions int
    Keyframes  []Keyframe
    Interpolation InterpKind
    Expression  string
}
type EffectInstance struct {
    MatchName string  // e.g. "ADBE Gaussian Blur 2"
    Params    map[string]*PropertyStream
}
```

Chunk refs (`*rifx.Chunk`) 不出现在 runtime types.

### M2. Archetype 最小化

只保留"最小合法 runtime object"的种子，不保留 content templates:

| Runtime kind | Archetype (canonical seed) |
|---|---|
| Empty project | `templates/2020.aep` (已有) |
| Empty comp | `templates/2020_dummy_comp.aep` (已有) |
| Empty shape layer | TBD — 1 个 minimal ShapeLayer ldta + tdgp |
| Empty text layer | TBD — 1 个 minimal TextLayer ldta + Slin |
| Empty AV layer | TBD |
| Empty camera/light layer | TBD |

**不**为 rectangle / ellipse / glow / drop-shadow / trim path 等每个 primitive 维护 archetype。这些通过 graph composition + serializer 动态生成。

### M3. Shape graph

```
ShapeLayer
└── VectorGroup (RootGroup)
    ├── RectNode { size, position, roundness }
    ├── EllipseNode { size, position }
    ├── PolyStarNode { points, innerRadius, outerRadius }
    ├── PathNode { bezier vertices }
    ├── MergeNode { mode }
    ├── TrimNode { start, end, offset }
    ├── FillNode { color, opacity }
    ├── StrokeNode { color, width, dashes }
    ├── GradientFillNode { … }
    ├── TransformNode { anchor, position, scale, rotation, opacity, skew }
    └── (nested VectorGroup)
```

Serializer 把 node graph 编码为 AE 的 `tdgp` LIST + `tdmn` matchName + property streams。Parser 反向。

### M4. Effect: generic container

```go
type EffectInstance struct {
    MatchName string                    // identity
    Params    map[string]*PropertyStream // schema lookup at runtime
}
```

不为每个 effect 一个 archetype。effect 差异只在 matchName + params。Param schema 来自 (a) runtime metadata table or (b) opaque pass-through（库不识别的 effect 字节级 round-trip）。

### M5. PropertyStream IR

统一 animation/property abstraction：

```go
type PropertyStream struct {
    Dimensions    int            // 1/2/3/color
    StaticValue   []float64      // when no keyframes
    Keyframes     []Keyframe
    Interpolation InterpKind     // Hold/Linear/Bezier
    Expression    string         // ExtendScript / JS source
    ExpressionEnabled bool
}
type Keyframe struct {
    Time    float64        // in seconds (TickRate handled by serializer)
    Value   []float64
    InEase  []Ease
    OutEase []Ease
    InInterp, OutInterp InterpKind
    SpatialInTangent, SpatialOutTangent []float64
}
```

时间单位 runtime 用秒；serializer 根据 comp TickRate 转 ticks。

### M6. Graph mutation API

```go
project.CreateLayer(comp, ShapeLayer, ...)
project.DeleteLayer(layer)
layer.InsertEffect(effect, idx)
layer.RemoveEffect(idx)
shapeLayer.AttachNode(parentGroup, node)
project.CloneSubgraph(srcLayer, dstComp)
layer.Reparent(newParentLayer)
```

graph-level transactions，不是直接 binary surgery。原子性靠 runtime snapshot/restore。

### M7. Capability matrix

替代 `if AE2020 / if AE2025`：

```go
type AECapabilities struct {
    LdtaSize                int    // 160 / 164
    FEEHasPpSn              bool
    MaterialLightingGroups  bool   // AE 24+
    TdgpVariant             int    // 19 / 37 children layout
    SupportsRendererSetter  bool
    NTSCShutterScriptingQuirk bool  // stored × 1.2 on read
    // ...
}

var Capabilities = map[AETarget]AECapabilities{
    TargetAE2020: { LdtaSize: 160, FEEHasPpSn: false, MaterialLightingGroups: false, TdgpVariant: 19, ... },
    TargetAE2022: { LdtaSize: 160, FEEHasPpSn: true,  MaterialLightingGroups: false, TdgpVariant: 19, ... },
    TargetAE2025: { LdtaSize: 164, FEEHasPpSn: true,  MaterialLightingGroups: true,  TdgpVariant: 37, ... },
}
```

Serializer 读 capabilities 决定写什么字节。Logical layer 跨版本不变。

### M8. Serializer boundary

```
runtime types  ←→  serializer  ←→  rifx.Chunk tree  ←→  .aep bytes
```

Current `internal/aep/` 混了两侧。V3 拆:

- `internal/scene/` — runtime types + mutation API
- `internal/serializer/` (or keep `internal/aep/`) — chunk synthesis + parsing
- `internal/rifx/` — 不变（底层二进制 framing，已分离）

API surface: `aep.Open(...) (*scene.Project, error)` / `scene.Project.WriteAEP(...) error` / scene mutations are pure functions over scene types.

## 不在 V3 范围

- 完整 effect parameter schema 库 (~1000 effects). Effect remains opaque-pass-through for unrecognized ones.
- Expression engine. Strings stay strings.
- Render queue / output module / preview. AE-runtime-only concepts.
- Color management deep integration.

## V3 落地建议路径

1. **brainstorm** (`superpowers:brainstorming` skill) → `flightdeck/specs/v3-runtime-design.md`
2. **plan** (`superpowers:writing-plans`) → `flightdeck/flight-plans/v3-runtime-plan.md`，多 phase：
   - Phase 1: scene types + parse 闭环（不动 serializer，只 wrap）
   - Phase 2: 1 个 layer kind 完整 round-trip via scene API
   - Phase 3: capability matrix + serializer 分包
   - Phase 4: 余 layer kinds
   - Phase 5: effect / shape / text 各子系统
   - Phase 6: mutation API + cleanup
3. **执行** subagent-driven，per phase

预计跨越多周。期间 V2 path 继续 (V2.2 Layer 创建 可走当前 chunk-patch 套路，作为 RE 补充，但**不再深化**：V2.x 接下来每个 sub-project 都要权衡：是补 V3 brainstorm 输入，还是真的 V2 path 必要 ship）。

## scene→rifx 残留解耦项（方案① 重组遗留，方案② M8 清零）

2026-05-30 的 aep package 命名轴重组（方案①，`specs/2026-05-30-aep-package-reorg-design.md`）把 `internal/aep` 沿 `<stage>_<domain>` 收口，并加 `arch_boundary_test.go` 守卫「`scene_*` 禁 import rifx」。守卫枚举出 **5 个 `scene_*` 文件仍 import rifx**，均因持有 chunk 耦合逻辑，其干净归位正是 M8 serializer/scene 物理分包要做的。重组期暂入白名单（不阻塞、且守卫禁止新增 scene→rifx 耦合）。M8 落地时按下表迁出，迁完即清空白名单：

| scene_ 文件 | rifx 用法 | M8 迁往 |
|---|---|---|
| `scene_features.go` | Marker/Mask backref 结构持 `*rifx.Chunk`（ldat/nmHd/nmrd/mkif/shph） | `back_marker.go` / `back_mask.go` |
| `scene_project_settings.go` | root flag-chunk 读写（lnrb/lnrp/dwga，length-preserving patch） | `write_project_settings.go`（write 阶段） |
| `scene_project_views.go` | root LIST 导航 `findRootListByType`（Pefl/Pjef） | `parse_*`（读阶段导航） |
| `scene_property_flags.go` | `decodeTdumValue(*rifx.Chunk)` 解码 helper | `parse_properties.go` |
| `scene_property_group.go` | `AEPropertyGroup` 持 `*rifx.Chunk` + tdgp 树构建 | `parse_*`（chunk→tree 构建） |

> 这 5 项是 M8「scene 纯逻辑、零 chunk 引用」的具体待办；守卫白名单与本表是 M8 的 ready-made checklist。

## 决策记录

- 2026-05-22 V2.1 完工 + GPT 反馈 + escape-hatch 实证 → V3 brainstorm 当作"主线下个 phase 候选"。等用户确认起步。
- 2026-05-30 aep package 命名轴重组（方案①）ship；scene→rifx 残留 5 项入守卫白名单 + 上表，作为 M8 前置解耦清单。

## 关联文档

- 反馈源: `flightdeck/kneeboard/gpt`
- V2.1 实践: `flightdeck/incident-reports/ae25-acceptance-gate.md`
- 当前架构概览: 项目根 `CLAUDE.md` § 数据流 + § 硬约束（架构 ground truth 已并入 CLAUDE.md，原 architecture.md 内容拆分到 scars/ + playbooks/ + coverage.md）
- V2.1 plan: `flightdeck/landed/flight-plans/2026-05-22-v2-1-foundation-plan.md`
