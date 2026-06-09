---
status: active
summary: py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated，animated 限 3D+linear) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落。剩余均 deferred / fixture·RE-gated：ValueText(won't-do) · RQ output-module 路径名+全 settings enum 完整化 · Renderer W · EG W · Guides · Gradient stroke W · TimeRemap）
---

# py-aep parity — API 全覆盖路线图

**Status**: **部分完成（partial / paused）— 更新 2026-06-09**。
- **P1 ✅ 全落**（landed `2026-05-26-py-aep-parity-p1-plan`）
- **P2 ✅ 大部分**（landed P2a/P2b plans）；剩 deferred 少数：2C Gradient **W**（2026-05-31 已 ship gradient fill write，stroke/Type-Start-End 仍 deferred）、2E ImportPlaceholder（AE 拒收，删除）、2K TimeRemap enable（结构性）、Composition.Time（low-pri）、ReplaceWithPlaceholder/Solid
- **P3 🚧 进行中**——已落：**§3A Render Queue R/W**（`Project.RenderQueue` 读 + `AddItem/RemoveItem` 结构性 + RenderQueueItem/OutputModule 各 15+/12+ 值 setter + `SetComment`（length-variable, ship-gated）；详 `../plans/2026-06-01-py-aep-p3-renderqueue-reader-plan.md`）、**§3C PropertyBase Remove/Duplicate/MoveTo + DimensionsSeparated R/W**（static + animated，animated 限 3D layer + ~linear path-ease）、**§3G comp marker 增删**（双版本 ship-gate）。剩余 deferred（多 fixture·RE-gated）：RQ output-module 路径名 + 全 settings enum / format options×7 完整化、Composition.Renderer W、Essential Graphics W、Guides、Gradient stroke W、§3H ValueText（won't-do）。**例外**：Layer 结构性 ops（Remove/Duplicate/CopyToComp/Move）已由 **V3 Phase 2-5 做掉**（见下表 ✅）
**Created**: 2026-05-26
**Goal**: aep-parser **API 覆盖 ≥ py-aep**（[forticheprod/py-aep](../charts/py-aep)，~20k LOC）+ 保留我们既有优势（length-preserving 写、AE 2020/2025 双 ship gate、V2.2 ShapeLayer 创建、文本完整 setter）。

参照源: `flightdeck/charts/py-aep/`（v0.x，自标 "save() highly experimental"，写区 alpha）。

---

## 1. 战略基调

### 1.1 我们 vs py-aep 立场对比

| 维度 | py-aep | aep-parser (我们) |
|---|---|---|
| 写策略 | 全 RIFX 重写（save 新文件，不允许覆盖） | length-preserving splice 默认 + V2.x 结构性创建（NewProject/NewComp/NewShapeLayer）|
| API 镜像 | 全 ExtendScript surface（snake_case） | 选择性 ExtendScript-similar + 额外 typed accessor pair |
| 文本 | "Many attrs missing"（自承认） | 完整：16 typed setter + 22 字段 per-run + 14 字段 per-paragraph + FontAxes R + AddFont |
| Camera/Light | ✅ kind + props | ✅ 13+11 typed accessor pair（kind、shadow、light props） |
| ShapeLayer 创建 | 无（只能解析） | ✅ V2.2 alpha NewShapeLayer + AddRect/Ellipse/Path/Fill/Stroke |
| NewProject/NewComposition | 无（save 整体）| ✅ V2.1 双 ship gate PASS |
| Render Queue | ✅ 完整 + format options | ❌ |
| Essential Graphics | 🚧 controllers exposed | ❌ |
| Gradient | ✅ XML 解析（color_stops/alpha_stops）| ✅ R (P2b 2C)；W 待 fixture |
| Project 多字段 | ✅ ~10 个（nnhd/head/CMS/XMP/flag chunks）| 🚧 只 BitsPerChannel |
| Layer 结构性 | ✅ remove/duplicate/copy_to_comp/move_*| ❌（仅 keyframe add/remove） |

### 1.2 工作原则（不动 [`../../CLAUDE.md`](../../CLAUDE.md) 核心铁律）

1. **length-preserving 默认路径不动** — 新加字段优先 splice；结构性 ops 走 V2.x capability framework
2. **public API 不动** — 现有 `Layer.SetXxx` 系列保留；新加镜像 py-aep 命名（English snake_case 改 Go PascalCase）的 typed pair
3. **单 `internal/aep` package** — 不引子包；新功能进同 package
4. **文档铁律** — 每个 phase 完成同步 [docs/](../../docs/)、[coverage.md](../plans/coverage.md)、[coverage-detail.md](../plans/coverage-detail.md)、[cockpit.md](../cockpit.md)
5. **alpha 写区严格 ship-gate** — 任何新结构性写都跑 AE 2020 + AE 2025 双开 fixture 校验（详 [`incidents/ae25-acceptance-gate.md`](../incidents/ae25-acceptance-gate.md)）

### 1.3 显式 non-goals（py-aep 有但我们不抄）

- **ExtendScript 1-based indexing** — Go 0-based 不改
- **Pythonic iterator protocol** — Go 用 slice 直接 range，无 `__iter__`
- **Property descriptor metaclass** — Go 写 typed getter/setter pair（已是项目惯例，~53 个）
- **save(new_path) 全重写** — 跟 length-preserving invariant 冲突；写新文件走 `WriteAEP` 现有 API
- **Runtime-only / ScriptingAPI-only 字段** — `dropFrame` / `fontLocation` / `selection` 等已经在 [`incidents/runtime-only-fields.md`](../incidents/runtime-only-fields.md) 证实不可写；不重复 RE
- **Expression evaluation** — py-aep 也明确不支持，符号执行不在 scope

---

## 2. API 全表（对照 [py-aep ExtendScript coverage](../charts/py-aep/docs/extendscript_coverage.md)）

每行: ✅ = 已 ship | 🟢 = 部分 | 🟡 = R only | ❌ = 缺 | 🗑️ = 暂搁 (runtime-only/structural-blocked)
列含义: **py-aep** = py-aep 状态 / **我们** = 当前 / **目标** = parity 后

### 2.1 Project 域

| API | py-aep | 我们 | 目标 | Phase | 备注 |
|---|---|---|---|---|---|
| `BitsPerChannel` | ✅ | ✅ R/W | ✅ | done | |
| `Revision` (head @ next_item_id+revision) | ✅ R/W | ✅ R | ✅ R only | done | P1 1D；user-action counter；写无意义 |
| `LinearBlending` (lnrb flag chunk) | ✅ toggle | ✅ R/W | ✅ R/W | done | P1 1D；toggle = add/remove chunk |
| `LinearizeWorkingSpace` (lnrp) | ✅ toggle | ✅ R/W | ✅ R/W | done | P1 1D；同上 |
| `CompensateForSceneReferredProfiles` (acer.value) | ✅ R/W | ✅ R/W | ✅ R/W | done | P1 1D；bool, 1 byte |
| `AudioSampleRate` (adfr.value) | ✅ R/W enum{22050,32000,44100,48000,96000} | ✅ R/W | ✅ R/W | done | P1 1D；f64 |
| `WorkingGamma` (dwga) | ✅ R/W enum{2.2, 2.4} | ✅ R/W | ✅ R/W | done | P1 1D；f64 + validate |
| `GpuAccelType` (gpug.Utf8) | ✅ R/W enum | ✅ R/W | ✅ R/W | done | P1 1D；Utf8 length-variable, splice OK |
| `ExpressionEngine` (ExEn LIST→Utf8) | ✅ R/W enum{"extendscript","javascript-1.0"} | ✅ R/W | ✅ R/W | done | P1 1D；设过/未设两路径 |
| `FeetFramesFilmType` (nnhd) | ✅ R/W enum{16mm/35mm} | ✅ R/W | ✅ R/W | done | P2b 2A |
| `FootageTimecodeDisplayStartType` (nnhd) | ✅ R/W | ✅ R/W | ✅ R/W | done | P2b 2A |
| `TimecodeDefaultBase` (nnhd) | ✅ R/W [1..999] | ✅ R/W | ✅ R/W | done | P2b 2A |
| `FramesCountType` (nnhd) | ✅ R/W enum | ✅ R/W | ✅ R/W | done | P2b 2A |
| `DisplayStartFrame` (nnhd, 0/1) | ✅ R/W | ✅ R/W | ✅ R/W | done | P2b 2A |
| `FramesUseFeetFrames` (nnhd) | ✅ R/W | ✅ R/W | ✅ R/W | done | P2b 2A |
| `TimeDisplayType` (nnhd) | ✅ R/W enum | ✅ R/W | ✅ R/W | done | P2b 2A |
| `TransparencyGridThumbnails` (nnhd) | ✅ R/W | ✅ R/W | ✅ R/W | done | P2b 2A |
| `ColorManagementSystem` (CMS JSON, AE 24+) | ✅ R/W enum | ✅ R/W | ✅ R/W | done | P2b 2B |
| `LutInterpolationMethod` (CMS) | ✅ R/W | ✅ R/W | ✅ R/W | done | P2b 2B |
| `OcioConfigurationFile` (CMS) | ✅ R/W | ✅ R/W | ✅ R/W | done | P2b 2B |
| `WorkingSpace` (CMS JSON `baseColorProfile.colorProfileName`) | ✅ R only | 🟢 R only | 🟢 R only | done | P2b 2B |
| `DisplayColorSpace` (separate chunk) | ✅ R only | 🗑️ deferred | 🗑️ deferred | deferred | P2b 2B；separate chunk 位置未 RE，旧 stub 已删 |
| `XmpPacket` | ✅ R/W (raw XML) | 🟡 R only | 🟡 R only | done | P2c followup#2；trailing UTF-8 after RIFX，已通过 `root.Trailing` 透传 roundtrip。`Project.XmpPacket()` R only；W deferred（AE 可能校验 XML 结构） |
| `EffectNames` (Pefl/pjef list) | ✅ R only | ✅ R only | ✅ R only | done | P1 1B |
| `Compositions / Folders / Footages` filter | ✅ | ✅ | ✅ | done | P1 1B |
| `RootFolder` | ✅ | ✅ | ✅ | done | P1 1B |
| `LayerByID(id)` | ✅ | ✅ | ✅ | done | P1 1B |
| `ImportPlaceholder(name, w, h, fps, dur)` | ✅ | 🗑️ deferred | 🗑️ deferred | deferred | P2a Task 5；opti format 未 RE，合成 builder AE 拒收，2026-05-27 删除 |
| `Save(path)` | 🟡 alpha 全重写 | 🟢 WriteAEP 任意 io.Writer | — | done | 我们已有 |

### 2.2 CompItem 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Name / FrameRate / Duration / Size / BGColor / Shutter*` | ✅ | ✅ R/W | ✅ | done |
| `PixelAspect / ResolutionFactor / WorkArea*` | ✅ | ✅ R/W | ✅ | done |
| `DisplayStartTime` | ✅ | ✅ R/W | ✅ | done |
| `DisplayStartFrame` | ✅ | ✅ R/W | ✅ R/W | done | P1 1C |
| `WorkAreaStartFrame / WorkAreaDurationFrame` | ✅ | ✅ R/W | ✅ R/W | done | P1 1C |
| `DropFrame` | ✅ R only | 🗑️ | 🗑️ | done | runtime-only |
| `Renderer` | ✅ R/W | 🟢 R only | ✅ R/W | P3 | prda 长度随 renderer 变（结构性） |
| `Markers` flat list | ✅ | ✅ | ✅ | done | P1 1C |
| `NumLayers / HasAudio / ActiveCamera` | ✅ | ✅ | ✅ | done | P1 1F |
| `TextLayers / ShapeLayers / CameraLayers / LightLayers` | ✅ | ✅ | ✅ | done | P1 1A |
| `NullLayers / SolidLayers / AdjustmentLayers` | ✅ | ✅ | ✅ | done | P1 1A |
| `ThreeDLayers / GuideLayers / SoloLayers` | ✅ | ✅ | ✅ | done | P1 1A |
| `AVLayers / CompositionLayers / FootageLayers / FileLayers / PlaceholderLayers` | ✅ | ✅ | ✅ | done | P1 1A |
| `TimeScale` | ✅ R only | ✅ TickRate alias | ✅ | done | P1 1F |
| `MotionGraphicsTemplateName` | ✅ R/W | ✅ R | ✅ R/W | R done (P3) | Essential Graphics；W deferred（结构性 .mogrt binding） |
| `MotionGraphicsControllers` | ✅ R only | ❌ | ✅ R only | P3 | |
| `MotionGraphicsTemplateController{Count,Names}` | ✅ | ❌ | ✅ | P3 | |
| `Guides` (ruler 辅助线) | ✅ R only | ❌ | ✅ R/W | P3 | UI-only chunk |
| `Time` (current time) | ✅ R/W | ❌ | ✅ R/W | P2 | UI state, low priority |
| `FrameDuration` (total in frames) | ✅ | ✅ | ✅ | done | P1 1C |

### 2.3 Layer / AVLayer / 类型化 Layer 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Name/Comment/Label/InPoint/OutPoint/StartTime/Stretch/Parent/Source/TrackMatte/AutoOrient` | ✅ | ✅ R/W | ✅ | done |
| `Visible/Solo/Shy/Locked/EffectsActive/MotionBlur/AudioEnabled/FrameBlend.*` | ✅ | ✅ R/W | ✅ | done |
| `CollapseTransform/Is3D/IsAdjust/IsGuide/IsNull/MarkersLocked/PreserveTransparency` | ✅ | ✅ R/W | ✅ | done |
| `Quality/SamplingQuality/BlendingMode/EnvironmentLayer` | ✅ | 🟢 大部分 | ✅ | done |
| `AlternateSource` (Media Replacement) | ❌ | ✅ R/W | ✅ | done | 我们有 py-aep 没 |
| `Transform.{Anchor,Position,Scale,Rotate,Orientation,Opacity}` | ✅ R/W (via property) | ✅ typed setter | ✅ | done |
| `AudioLevels` | ✅ | ✅ R/W | ✅ | done |
| Camera 13 fields (Zoom/DOF/Focus/Aperture/BlurLevel/Iris×8) | ✅ | ✅ R/W | ✅ | done |
| Light 11 fields (Color/Intensity/Cone*/Falloff*/Shadow*) + LightKind | ✅ | ✅ R/W | ✅ | done |
| `LightSource` (light source layer ref, AE 24+) | ✅ R/W | ✅ R/W | ✅ R/W | done | P2a Task 2 |
| `FrameInPoint / FrameOutPoint / FrameStartTime / FrameTime` | ✅ | ✅ R/W | ✅ R/W | done | P1 1C |
| `Index` (within comp, 0-based) | ✅ | 🟢 implicit | ✅ | P1 | accessor |
| `LayerType` (string discriminator) | ✅ R only | ✅ via `Layer.Type` enum | ✅ | done | P1; typed dispatch in `inferLayerType` |
| `HasVideo / HasAudio / AudioActive / AudioActiveAtTime(t)` | ✅ | ✅ | ✅ | done | P1 1E |
| `Active / ActiveAtTime(t)` | ✅ | 🗑️ | 🗑️ | done | N/A (AE runtime) |
| `AdjustmentLayer / EnvironmentLayer / GuideLayer / ThreeDLayer / ThreeDPerChar` (typed) | ✅ | 🟢 ldta bit R/W | ✅ | P1 | rename to py-aep style helpers |
| `Width / Height` (from source) | ✅ | ✅ | ✅ | done | P1 1E |
| `HasTrackMatte / IsTrackMatte / TrackMatteLayer` | ✅ | 🟢 R/W TrackMatteLayer | ✅ | P1 | add helpers |
| `AutoName / IsNameFromSource` | ✅ | ✅ | ✅ | done | P1 1E |
| `SetTrackMatte(layer, type)` | ✅ | 🟢 SetTrackMatteLayer | ✅ | P1 | 增 type 参数 |
| `RemoveTrackMatte()` | ✅ | ✅ | ✅ | done | P1 1E |
| `ReplaceSource(newSource, fixExpressions=false)` | ✅ | ✅ | ✅ | done | P2a Task 4 |
| `ContainingComp` | ✅ | 🟢 implicit | ✅ | P1 | back-ref |
| `Marker` (PropertyGroup) + `Markers` (flat list) | ✅ | 🟢 layer.Markers | ✅ | done |
| `Effects / Masks / Text / Transform` PropertyGroup accessors | ✅ | ✅ R | ✅ | done | P2c — Layer.TransformGroup/AudioGroup/EffectsParade/MaskParade/etc. |
| `Remove()` | ✅ | ✅ | ✅ | done | V3 Phase2 DeleteLayer，双版本 ship-gate PASS |
| `Duplicate()` | ✅ | ✅ | ✅ | done | V3 Phase3 DuplicateLayer |
| `CopyToComp(comp)` | ✅ | ✅ | ✅ | done | V3 Phase5C InsertLayer（+ 5C1 cross-Project） |
| `MoveAfter/MoveBefore/MoveToBeginning/MoveToEnd` | ✅ | ✅ | ✅ | done | V3 Phase4 MoveLayer |
| `SetParentWithJump(layer)` | ✅ | ❌ | ✅ | P3 | preserve world transform（未做） |
| `CanSetCollapseTransformation / CanSetTimeRemapEnabled` | ✅ | ✅ R | ✅ | done | P2c followup#2；纯派生 capability query — `Layer.AVSource()` 解 source 后判 type/duration |
| `TimeRemapEnabled / SetTimeRemap*` | ✅ | 🟢 R only | ✅ R/W | P2 | enable=structural |
| `ThreeDModelLayer`（Cinema 4D / GLB） | ✅ | ✅ R only | ✅ R only | done | P2a Task 1 |

### 2.4 Property / PropertyBase / PropertyGroup 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `MatchName / Name` | ✅ | ✅ R | ✅ | done |
| `Components / Dimensions` | ✅ | ✅ R (Components) | ✅ | done |
| `LockedRatio` (tdsb) | ✅ R only | ✅ R/W | ✅ R/W | done | P2a Task 3 |
| `IsSpatial` (tdb4) | ✅ R | ✅ R | ✅ R | done | P1 1G |
| `Animated / Color / Integer / NoValue / Vector` (tdb4 flags) | ✅ R | ✅ R | ✅ R | done | P1 1G |
| `DefaultValue / LastValue / NbOptions` | ✅ R | ✅ R | ✅ R | done | P2c followup; transform defaults table + pard chunk parsing |
| `MinValue / MaxValue / UnitsText` | ✅ R | ✅ R | ✅ R | done | P2c followup; tdum/tduM chunk + unitsTextMap |
| `Value` (current) | ✅ R/W | ✅ R/W StaticValue | ✅ | done |
| `CanVaryOverTime` | ✅ R | ✅ R | ✅ R | done | P1 1G |
| `DimensionsSeparated` | ✅ R/W | ✅ R/W | ✅ R/W | done (P3 §3C) | R + W both directions (Alpha); static Position 2D+3D, AE 2020+2025 ship-gate 6/6; animated done (3D layer + ~linear path-ease only). See `incidents/separate-dimensions-write-mechanics.md` |
| `Expression / ExpressionEnabled / CanSetExpression` | ✅ R/W | ✅ R/W | ✅ | done |
| `PropertyControlType` | ✅ R | ✅ R | ✅ R | done | P2c followup; derived from tdb4 flags |
| `PropertyValueType` | ✅ R | ✅ R | ✅ R | done | P2c followup; derived from tdb4 flags |
| `IsModified / Active / Elided / IsNameSet` | ✅ R | ✅ R | ✅ R | done | P2c followup#2；Property + AEPropertyGroup. Elided=false placeholder (no synthesis yet)；IsNameSet 用 `Name != MatchName` proxy（tdsn decode deferred） |
| `PropertyIndex / PropertyDepth` | ✅ R | ✅ R | ✅ R | done | P2c followup; via parentTreeGroup |
| `ParentProperty` (PropertyGroup back-ref) | ✅ R | ✅ R | ✅ R | done | P2c followup; ParentGroup() |
| `Selected / SelectedKeys` | ✅ R | 🗑️ | 🗑️ | done | runtime-only |
| `EssentialPropertySource` | ✅ R | ❌ | 🗑️ | — | runtime/EG |
| `AlternateSource / CanSetAlternateSource` (Layer-level, on Property) | ✅ R | 🟢 Layer.AlternateSource | ✅ | done |
| `ValueText` (formatted value) | 🚧 py-aep 也没做 | ❌ | 🗑️ deferred | P3 §3H | **通用不可达**：内置枚举 label 不在文件（需 Adobe 不公开 schema DB），AE 26.0-only API；仅自定义 Dropdown Menu Control 子集可 RE。2026-06-04 决策 defer，详 `../incidents/valuetext-needs-schema-db.md` |
| **PropertyGroup ops**: `Properties / Property(key) / NumProperties / CanAddProperty` | ✅ | ✅ R (sans CanAddProperty) | ✅ | done | P2c — `AEPropertyGroup.Property/Group/ChildByIndex/NumProperties/ParentGroup/PropertyByPath`；CanAdd 结构性 P3 |
| `PropertyBase.Remove / Duplicate / MoveTo` | ✅ R/W | ❌ | ✅ R/W | P3 | 结构性 |
| `Keyframe.Time / FrameTime / Value / InInterpType / OutInterpType / TemporalEase / SpatialTangent` | ✅ | ✅ R/W | ✅ | done |
| `InsertKeyframe / DeleteKeyframe` | 🚧 limited | ✅ | ✅ | done |
| **Gradient** (ADBE Vector Grad Colors XML 解析 → color_stops + alpha_stops) | ✅ R/W | ✅ R | ✅ R/W | P2b 2C R done | XML 在 `GCst → GCky → Utf8`（不在 cdat）；写回 deferred |
| **MaskPropertyGroup** | ✅ | 🟢 Mask 类型化 | ✅ | done |

### 2.5 Footage / Source 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Path` | ✅ R/W | ✅ R/W | ✅ | done |
| `Width / Height / Duration / FrameRate / FrameDuration / PixelAspect` | ✅ R | 🟢 partial | ✅ R | P1 | proxy to source |
| `FootageMissing / HasAudio` | ✅ R | ✅ R | ✅ R | done | P1 1H |
| `StartFrame / EndFrame` | ✅ R | ✅ R | ✅ R | done | P1 1H |
| `AssetType` (placeholder/solid/file) | ✅ R | ✅ R | ✅ R | done | P1 1H |
| `MainSource` (typed: FileSource / SolidSource / PlaceholderSource) | ✅ | ✅ R | ✅ | done | P2c followup；`Footage.MainSource()` 返回 typed interface。SolidSource.Color 零值占位（sspc 颜色字段未 RE） |
| `File` (FootageItem.file convenience) | ✅ R | ✅ R | ✅ | done | P1 1H |
| `ReplaceWithPlaceholder / ReplaceWithSolid` | ✅ R/W | ❌ | ✅ R/W | P2 | 结构性 source 替换 |
| `Proxy / UseProxy / ProxySource` | ❌ py-aep 没| ❌ | ❌ | — | py-aep 也没 |

### 2.6 Render Queue / Output Module 域（**全新域，~3-5k LOC 估算**）

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Project.RenderQueue` | ✅ | ❌ | ✅ R/W | P3 |
| `RenderQueue.Items / NumItems` | ✅ | ❌ | ✅ R | P3 |
| `RenderQueueItem.Status / Render / Comment / Name / Settings` | ✅ R/W | ❌ | ✅ R/W | P3 |
| `RenderQueueItem.Comp / OutputModules / NumOutputModules` | ✅ R | ❌ | ✅ R | P3 |
| `RenderQueueItem.TimeSpan{Start,Duration} / SkipFrames / ElapsedSeconds` | ✅ R/W | ❌ | ✅ R/W | P3 |
| `RenderQueueItem.LogType / QueueItemNotify` | ✅ R/W | ❌ | ✅ R/W | P3 |
| Render Settings 全 enum：ColorDepth / DiskCache / Effects / FieldRender / FrameBlending / FrameRateSetting / GuideLayers / MotionBlur / ProxyUse / Pulldown / Quality / SoloSwitches / TimeSpanSource | ✅ R/W | ❌ | ✅ R/W | P3 |
| `OutputModule.FileTemplate / FormatOptions / Settings` | ✅ R/W | ❌ | ✅ R/W | P3 |
| Format options 结构: Cineon / Jpeg / OpenExr / Png / Targa / Tiff / Xml | ✅ R/W | ❌ | ✅ R/W | P3 |
| `RenderQueue.CanQueueInAME / Rendering / Templates` | ❌ runtime | ❌ | 🗑️ | — |

### 2.7 Text 域（**我们领先**）

| API | py-aep | 我们 |
|---|---|---|
| `TextDocument` 全字段 | ✅ | ✅ |
| Per-run typed setter (×16) | 🚧 "many missing" | ✅ |
| Per-paragraph typed setter (×14) | 🚧 | ✅ |
| `FontAxes` R / variable fonts | ❌ | ✅ R |
| `AddFont` | ❌ | ✅ |
| `ManualKerning + Kerning` | 🚧 partial | ✅ R/W |
| `IsBoxText + BoxBounds` | ✅ | ✅ |
| `CharacterRange / ComposedLineRange / ParagraphRange` | ❌ | ❌ | — | 这是 runtime API |

### 2.8 Essential Graphics 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `EssentialGraphicsController` (controllers + override UUIDs) | 🚧 controllers exposed | ✅ controllers R | 🚧 controllers R | done (P3) | CIF3→CCtl: Name/Type/UUID + comp template name；10 fixture |
| 自动 controller-override 解析 | ❌ | ❌ | ❌ | — |

### 2.9 杂项

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Application` 顶层 (version / project / etc) | 🚧 | 🟢 `*Project` 等价 | ✅ | P1 | 加 `Application` 或 `App` wrapper |
| `Application.Version` | ✅ R | ✅ R | ✅ R | done | P1 1I；`internal/aep/application.go::Application.Version()` 从 head 解 |
| `Viewer / ViewOptions / View` | ✅ R only | ❌ | 🗑️ | — | UI state, runtime |
| `ImportOptions` | ❌ py-aep 也没 | ❌ | ❌ | — |

---

## 3. Phase 规划

每个 phase 独立可 ship；先做 P1（低风险、零 invariant 冲突）再分头进 P2/P3。

### Phase 1 — 低悬果：filter views + frame-time 伴生 + Project single-byte chunks（**第一波**）

**Target**: 50+ 新 API，零结构性写，纯 reader + 现有 chunk slice 设字段。

子任务（详 `flightdeck/plans/2026-05-26-py-aep-parity-p1-plan.md`）：

- **1A** CompItem filter views (15+)：`TextLayers / ShapeLayers / CameraLayers / LightLayers / NullLayers / SolidLayers / AdjustmentLayers / ThreeDLayers / GuideLayers / SoloLayers / AVLayers / CompositionLayers / FootageLayers / FileLayers / PlaceholderLayers`
- **1B** Project filter views: `Folders / Footages / RootFolder / LayerByID(id) / EffectNames`
- **1C** Frame-time 伴生 accessor pair: `Layer.{FrameInPoint, FrameOutPoint, FrameStartTime, FrameTime} R/W` + `Composition.{DisplayStartFrame, WorkAreaStartFrame, WorkAreaDurationFrame, FrameDuration, FrameTime} R/W` + `Keyframe.FrameTime R/W` + `Marker.FrameTime R/W` + `Marker.FrameDuration R/W`
- **1D** Project single-field chunks: `Revision R / LinearBlending R/W / LinearizeWorkingSpace R/W / CompensateForSceneReferredProfiles R/W / AudioSampleRate R/W / WorkingGamma R/W / GpuAccelType R/W / ExpressionEngine R/W`
- **1E** Layer convenience: `Index / LayerType / HasVideo / HasAudio / AudioActive / Active / ActiveAtTime(t) / Width / Height / HasTrackMatte / IsTrackMatte / AutoName / IsNameFromSource / ContainingComp / RemoveTrackMatte`
- **1F** Composition convenience: `NumLayers / HasAudio / ActiveCamera / Markers (flat) / TimeScale (alias TickRate)`
- **1G** Property tdb4 flag readers: `IsSpatial / IsAnimated / IsColor / IsInteger / IsNoValue / IsVector / CanVaryOverTime`
- **1H** Footage convenience: `FootageMissing / HasAudio / StartFrame / EndFrame / AssetType / File`
- **1I** Application wrapper: `aep.App` 顶层（指向 Project 等价，与 py-aep `parse()` 返回值对齐）+ `App.Version`

**估算**: ~1-2 个会话；80%+ 新 API 都是 helper function 或单 byte/chunk 读，无 RE。

### Phase 2 — 中等域：nnhd 全展开 + CMS + Gradient + 类型化 Sources

子任务：

- **2A** nnhd 字节布局 RE：`FeetFramesFilmType / FootageTimecodeDisplayStartType / TimecodeDefaultBase / FramesCountType / DisplayStartFrame / FramesUseFeetFrames / TimeDisplayType / TransparencyGridThumbnails`（一次 RE 出 8 个字段）
- **2B** CMS JSON (AE 24+) `ColorManagementSystem / LutInterpolationMethod / OcioConfigurationFile`（setter 在 `cmsUtf8 == nil` 时拒写）；`WorkingSpace` R only；`DisplayColorSpace` 暂搁（separate chunk 位置未 RE）；XMP R only
- **2C** Gradient (XML in cdat): `Gradient { ColorStops, AlphaStops }` + `ADBE Vector Grad Colors` Property 类型识别 → 解析 XML → 暴露结构化访问 + 写回（XML 重序列化）
- **2D** 类型化 Footage Sources: `FileSource / SolidSource / PlaceholderSource` 区分 + `Footage.MainSource` typed return + `MinValue / MaxValue / UnitsText / DefaultValue / LastValue / NbOptions / PropertyControlType / PropertyValueType` Property 读
- **2E** ImportPlaceholder: `Project.ImportPlaceholder(name, w, h, fps, dur)` — 跟 NewComposition 同模式。**2026-05-27 暂搁**：合成 opti tag AE 不识别，需真 placeholder fixture 才能 RE 出 opti/sspc/idta 字段
- **2F** ReplaceSource (Layer 级): `Layer.ReplaceSource(newSource, fixExpressions=false)` — splice source id
- **2G** PropertyGroup 链式访问: `Layer.PropertyGroup("ADBE Transform Group").Property("ADBE Opacity")` — public API 加 hierarchical accessor
- **2H** Marker properties on Layer + PropertyGroup access for Transform/Effects/Masks/Text
- **2I** LightSource (light layer 关联另一个 layer, AE 24+)
- **2J** ThreeDModelLayer R only
- **2K** TimeRemap R/W (enable=structural)

**估算**: 3-5 个会话；nnhd RE 需要 fixture probe，Gradient XML 需要找 fixture + 写测试。

### Phase 3 — 大域 / 结构性：Render Queue + Layer 增删移 + Composition.Renderer 写 + Essential Graphics + Guides

子任务：

- **3A** Render Queue 全栈 (`RenderQueue / RenderQueueItem / OutputModule`) ✅ **R/W 大部分 done**——read + `AddItem/RemoveItem` 结构性 + RQItem/OM 值 setter（15+/12+）+ `SetComment`（ship-gated）；残留 deferred：output-module 路径名、全 settings enum、format options × 7
- **3B** Layer 结构性 ops: `Remove / Duplicate / CopyToComp / MoveAfter / MoveBefore / MoveToBeginning / MoveToEnd / SetParentWithJump`（全 V3 capability framework 内做）
- **3C** PropertyBase 结构性: `Remove / Duplicate / MoveTo / DimensionsSeparated R/W` ✅ **done**——`RemovePropertyGroup / DuplicatePropertyGroup / MovePropertyGroup`（facade 自由函数，INDEXED_GROUP：effect/mask parade · root vectors · text animators，详 `../incidents/property-indexed-group-structural-re.md`）；`DimensionsSeparated R/W` separate↔merge 双向 static Position 2D+3D（双版本 ship-gate 6/6）+ animated（限 3D layer + ~linear path-ease，详 `../incidents/separate-dimensions-write-mechanics.md`）
- **3D** Composition.Renderer 写（跨 renderer 切换 → prda 长度变结构性）
- **3E** Essential Graphics R only (controllers + override UUIDs，不做自动解析)
- **3F** Guides R/W (ruler 标尺辅助线，UI-only chunk)
- **3G** Composition Markers ：comp-level marker 增删 ✅ **done（stable）**——`Marker.Remove()` + `Composition.AddMarker`（clone-template 规避 opaque RE），AE 2020+2025 ship-gate 2/2 PASS（2026-06-04）。限制：AddMarker 需 ≥1 现存 marker（空 comp seed 暂搁）+ tail-insert 未做 time 排序。详 `../plans/2026-06-04-py-aep-p3-3g-comp-marker-structural-plan.md`
- **3H** Property.ValueText 🗑️ **deferred / won't-implement（2026-06-04）**——通用不可达：内置枚举 label 不在 .aep（pard 只存 nbOptions 计数），需 Adobe 不公开的 per-effect×版本 schema DB；AE 26.0-only ScriptingAPI；py-aep 自己也没实现。仅自定义 Dropdown Menu Control 的 label 在 `pdnm` chunk 可 RE（有界子集，payoff 窄，不做）。详 `../incidents/valuetext-needs-schema-db.md`

**估算**: V3 capability framework 完成后才动；多个会话独立子项，可并行。

---

## 4. 风险与开放问题

### 4.1 已知 RE 工作

- nnhd 字节布局完全未知（P2 prereq）
- ADBE Vector Grad Colors cdat 内的 XML schema（P2 prereq）
- Render Queue/Output Module RQHF/RQMF chunk family 未解（P3 prereq）
- Essential Graphics override UUID 跨 chunk 关联（P3 prereq）

### 4.2 跟我们 invariant 冲突

- **save(new_path)**: py-aep 设计是写新文件，整 RIFX 重新组装 chunks。我们的 `WriteAEP(io.Writer)` 已支持任意 writer（包括 file）。**结论**：跟我们设计兼容，不抄 py-aep 的"save 不能 overwrite"约束
- **Layer.Remove() / Duplicate()**: py-aep 走全 RIFX rewrite，我们 length-preserving splice 不支持长度变化的整 chunk 增删。**结论**：进 V3 capability framework，**P3 前不动**
- **Property.Value setter for animated**: py-aep 自动 link/unlink keyframes。我们 SetStaticValue 现仅在 cdat 路径作；keyframed property 走 InsertKeyframe API。**结论**：API 差异保留，文档说明 keyframed path 走 InsertKeyframe

### 4.3 命名风格 trade-off

- py-aep `snake_case` Pythonic — Go 我们继续 `PascalCase`
- py-aep `frame_in_point` (字段) — 我们 `FrameInPoint` (method)
- py-aep `transform.opacity.value` 链式 — 我们 Go-style typed accessor `layer.SetOpacity(v)` + property tree 提供 escape hatch
- **结论**：API 表面用 Go 惯例，**不强镜像 snake_case 命名**

### 4.4 P1 优先级争议

P1 子项 ~50+ API，其中：
- **真低悬果（必做 P1）**: 1A/1B/1C/1F/1H 全是 helper，~30 API
- **可分歧（也许 P1 也许 P2）**: 1D Project chunks (8 个新 chunk 字段，少量 RE 风险), 1E Layer convenience (15 个), 1G Property flags (7 个)

**决议**: 都进 P1，因为 1D/1E/1G 单 byte/single field 即可解。

---

## 5. 进度跟踪

- Phase 1 plan: [`../plans/finish/2026-05-26-py-aep-parity-p1-plan.md`](../landed/plans/2026-05-26-py-aep-parity-p1-plan.md) (executed, archived 2026-05-26)
- Phase 2/3 plans: 写完 Phase 1 后再起，按需切
- Board `Active focus` 反映**当前 phase**
- Coverage docs 每个子任务完工同步更新

---

## 6. 相关文档

- 参照源: [`flightdeck/charts/py-aep/`](../charts/py-aep)
- 当前架构概览: 项目根 `CLAUDE.md` § 数据流 + § 硬约束
- 覆盖矩阵: [`../plans/coverage.md`](../plans/coverage.md) / [`../plans/coverage-detail.md`](../plans/coverage-detail.md)
- V2.2 ShapeLayer alpha: [`2026-05-22-v2-2-layer-creation-design.md`](../landed/specs/2026-05-22-v2-2-layer-creation-design.md)
- V3 方向: [`2026-05-22-v3-direction.md`](2026-05-22-v3-direction.md)
- 已知陷阱: [`../incidents/`](../incidents)
