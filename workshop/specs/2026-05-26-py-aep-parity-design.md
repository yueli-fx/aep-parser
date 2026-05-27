# py-aep parity — API 全覆盖路线图

**Status**: spec, ready for phased plans
**Created**: 2026-05-26
**Goal**: aep-parser **API 覆盖 ≥ py-aep**（[forticheprod/py-aep](workshop/reference/py-aep)，~20k LOC）+ 保留我们既有优势（length-preserving 写、AE 2020/2025 双 ship gate、V2.2 ShapeLayer 创建、文本完整 setter）。

参照源: `workshop/reference/py-aep/`（v0.x，自标 "save() highly experimental"，写区 alpha）。

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
| Gradient | ✅ XML 解析（color_stops/alpha_stops）| ❌ |
| Project 多字段 | ✅ ~10 个（nnhd/head/CMS/XMP/flag chunks）| 🚧 只 BitsPerChannel |
| Layer 结构性 | ✅ remove/duplicate/copy_to_comp/move_*| ❌（仅 keyframe add/remove） |

### 1.2 工作原则（不动 [`../../CLAUDE.md`](../../CLAUDE.md) 核心铁律）

1. **length-preserving 默认路径不动** — 新加字段优先 splice；结构性 ops 走 V2.x capability framework
2. **public API 不动** — 现有 `Layer.SetXxx` 系列保留；新加镜像 py-aep 命名（English snake_case 改 Go PascalCase）的 typed pair
3. **单 `internal/aep` package** — 不引子包；新功能进同 package
4. **文档铁律** — 每个 phase 完成同步 [docs/](../../docs/)、[coverage.md](../plans/coverage.md)、[coverage-detail.md](../plans/coverage-detail.md)、[board.md](../board.md)
5. **alpha 写区严格 ship-gate** — 任何新结构性写都跑 AE 2020 + AE 2025 双开 fixture 校验（详 [`scars/ae25-acceptance-gate.md`](../scars/ae25-acceptance-gate.md)）

### 1.3 显式 non-goals（py-aep 有但我们不抄）

- **ExtendScript 1-based indexing** — Go 0-based 不改
- **Pythonic iterator protocol** — Go 用 slice 直接 range，无 `__iter__`
- **Property descriptor metaclass** — Go 写 typed getter/setter pair（已是项目惯例，~53 个）
- **save(new_path) 全重写** — 跟 length-preserving invariant 冲突；写新文件走 `WriteAEP` 现有 API
- **Runtime-only / ScriptingAPI-only 字段** — `dropFrame` / `fontLocation` / `selection` 等已经在 [`scars/runtime-only-fields.md`](../scars/runtime-only-fields.md) 证实不可写；不重复 RE
- **Expression evaluation** — py-aep 也明确不支持，符号执行不在 scope

---

## 2. API 全表（对照 [py-aep ExtendScript coverage](workshop/reference/py-aep/docs/extendscript_coverage.md)）

每行: ✅ = 已 ship | 🟢 = 部分 | 🟡 = R only | ❌ = 缺 | 🗑️ = 暂搁 (runtime-only/structural-blocked)
列含义: **py-aep** = py-aep 状态 / **我们** = 当前 / **目标** = parity 后

### 2.1 Project 域

| API | py-aep | 我们 | 目标 | Phase | 备注 |
|---|---|---|---|---|---|
| `BitsPerChannel` | ✅ | ✅ R/W | ✅ | done | |
| `Revision` (head @ next_item_id+revision) | ✅ R/W | ❌ | ✅ R only | P1 | user-action counter；写无意义 |
| `LinearBlending` (lnrb flag chunk) | ✅ toggle | ❌ | ✅ R/W | P1 | toggle = add/remove chunk |
| `LinearizeWorkingSpace` (lnrp) | ✅ toggle | ❌ | ✅ R/W | P1 | 同上 |
| `CompensateForSceneReferredProfiles` (acer.value) | ✅ R/W | ❌ | ✅ R/W | P1 | bool, 1 byte |
| `AudioSampleRate` (adfr.value) | ✅ R/W enum{22050,32000,44100,48000,96000} | ❌ | ✅ R/W | P1 | f64 |
| `WorkingGamma` (dwga) | ✅ R/W enum{2.2, 2.4} | ❌ | ✅ R/W | P1 | f64 + validate |
| `GpuAccelType` (gpug.Utf8) | ✅ R/W enum | ❌ | ✅ R/W | P1 | Utf8 length-variable, splice OK |
| `ExpressionEngine` (ExEn LIST→Utf8) | ✅ R/W enum{"extendscript","javascript-1.0"} | ❌ | ✅ R/W | P1 | 设过/未设两路径 |
| `FeetFramesFilmType` (nnhd) | ✅ R/W enum{16mm/35mm} | ❌ | ✅ R/W | P2 | nnhd 字节布局未 RE |
| `FootageTimecodeDisplayStartType` (nnhd) | ✅ R/W | ❌ | ✅ R/W | P2 | 同上 |
| `TimecodeDefaultBase` (nnhd) | ✅ R/W [1..999] | ❌ | ✅ R/W | P2 | int |
| `FramesCountType` (nnhd) | ✅ R/W enum | ❌ | ✅ R/W | P2 | |
| `DisplayStartFrame` (nnhd, 0/1) | ✅ R/W | ❌ | ✅ R/W | P2 | |
| `FramesUseFeetFrames` (nnhd) | ✅ R/W | ❌ | ✅ R/W | P2 | |
| `TimeDisplayType` (nnhd) | ✅ R/W enum | ❌ | ✅ R/W | P2 | |
| `TransparencyGridThumbnails` (nnhd) | ✅ R/W | ❌ | ✅ R/W | P2 | |
| `ColorManagementSystem` (CMS JSON, AE 24+) | ✅ R/W enum | ❌ | ✅ R/W | P2 | JSON Utf8 chunk |
| `LutInterpolationMethod` (CMS) | ✅ R/W | ❌ | ✅ R/W | P2 | |
| `OcioConfigurationFile` (CMS) | ✅ R/W | ❌ | ✅ R/W | P2 | |
| `WorkingSpace` (ws Utf8) | ✅ R only | ❌ | ✅ R only | P2 | ICC blob 写需 Adobe ICC files, R only |
| `DisplayColorSpace` (dcs Utf8) | ✅ R only | ❌ | ✅ R only | P2 | 同上 |
| `XmpPacket` | ✅ R/W (raw XML) | ❌ | 🟡 R only | P2 | XML mutation 风险高，初版 R only |
| `EffectNames` (Pefl/pjef list) | ✅ R only | ❌ | ✅ R only | P1 | list of effect names used |
| `Compositions / Folders / Footages` filter | ✅ | 🟢 only Compositions | ✅ | P1 | 加 Folders / Footages helper |
| `RootFolder` | ✅ | 🟢（in items tree）| ✅ | P1 | typed accessor |
| `LayerByID(id)` | ✅ | ❌ | ✅ | P1 | 跨 comp 查找 |
| `ImportPlaceholder(name, w, h, fps, dur)` | ✅ | ❌ | ✅ | P2 | NewProject pattern |
| `Save(path)` | 🟡 alpha 全重写 | 🟢 WriteAEP 任意 io.Writer | — | done | 我们已有 |

### 2.2 CompItem 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Name / FrameRate / Duration / Size / BGColor / Shutter*` | ✅ | ✅ R/W | ✅ | done |
| `PixelAspect / ResolutionFactor / WorkArea*` | ✅ | ✅ R/W | ✅ | done |
| `DisplayStartTime` | ✅ | ✅ R/W | ✅ | done |
| `DisplayStartFrame` | ✅ | ❌ | ✅ R/W | P1 | 我们有 Time，加 Frame 双轨 |
| `WorkAreaStartFrame / WorkAreaDurationFrame` | ✅ | ❌ | ✅ R/W | P1 | frame-time 伴生 |
| `DropFrame` | ✅ R only | 🗑️ | 🗑️ | done | runtime-only |
| `Renderer` | ✅ R/W | 🟢 R only | ✅ R/W | P3 | prda 长度随 renderer 变（结构性） |
| `Markers` flat list | ✅ | 🟢 markerProperty only | ✅ | P1 | 加 `Composition.Markers` flat slice |
| `NumLayers / HasAudio / ActiveCamera` | ✅ | ❌ | ✅ | P1 | pure helper |
| `TextLayers / ShapeLayers / CameraLayers / LightLayers` | ✅ | ❌ | ✅ | P1 | filter helper |
| `NullLayers / SolidLayers / AdjustmentLayers` | ✅ | ❌ | ✅ | P1 | |
| `ThreeDLayers / GuideLayers / SoloLayers` | ✅ | ❌ | ✅ | P1 | |
| `AVLayers / CompositionLayers / FootageLayers / FileLayers / PlaceholderLayers` | ✅ | ❌ | ✅ | P1 | 按 source 类型 |
| `TimeScale` | ✅ R only | 🟢 TickRate | ✅ | P1 | 重命名/别名 |
| `MotionGraphicsTemplateName` | ✅ R/W | ❌ | ✅ R/W | P3 | Essential Graphics |
| `MotionGraphicsControllers` | ✅ R only | ❌ | ✅ R only | P3 | |
| `MotionGraphicsTemplateController{Count,Names}` | ✅ | ❌ | ✅ | P3 | |
| `Guides` (ruler 辅助线) | ✅ R only | ❌ | ✅ R/W | P3 | UI-only chunk |
| `Time` (current time) | ✅ R/W | ❌ | ✅ R/W | P2 | UI state, low priority |
| `FrameDuration` (total in frames) | ✅ | ❌ | ✅ | P1 | helper |

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
| `LightSource` (light source layer ref, AE 24+) | ✅ R/W | ❌ | ✅ R/W | P2 | light-source linking |
| `FrameInPoint / FrameOutPoint / FrameStartTime / FrameTime` | ✅ | ❌ | ✅ R/W | P1 | frame-time 伴生 |
| `Index` (within comp, 0-based) | ✅ | 🟢 implicit | ✅ | P1 | accessor |
| `LayerType` (string discriminator) | ✅ R only | 🟢 typed dispatch | ✅ | P1 | accessor returning class name |
| `HasVideo / HasAudio / AudioActive / AudioActiveAtTime(t)` | ✅ | ❌ | ✅ | P1 | helper |
| `Active / ActiveAtTime(t)` | ✅ | ❌ | ✅ | P1 | (visible && t in range) |
| `AdjustmentLayer / EnvironmentLayer / GuideLayer / ThreeDLayer / ThreeDPerChar` (typed) | ✅ | 🟢 ldta bit R/W | ✅ | P1 | rename to py-aep style helpers |
| `Width / Height` (from source) | ✅ | ❌ | ✅ | P1 | proxy to source |
| `HasTrackMatte / IsTrackMatte / TrackMatteLayer` | ✅ | 🟢 R/W TrackMatteLayer | ✅ | P1 | add helpers |
| `AutoName / IsNameFromSource` | ✅ | ❌ | ✅ | P1 | helper |
| `SetTrackMatte(layer, type)` | ✅ | 🟢 SetTrackMatteLayer | ✅ | P1 | 增 type 参数 |
| `RemoveTrackMatte()` | ✅ | ❌ | ✅ | P1 | |
| `ReplaceSource(newSource, fixExpressions=false)` | ✅ | 🟢 SetSource (low-level) | ✅ | P2 | fixExpressions 是 ScriptingAPI 行为 |
| `ContainingComp` | ✅ | 🟢 implicit | ✅ | P1 | back-ref |
| `Marker` (PropertyGroup) + `Markers` (flat list) | ✅ | 🟢 layer.Markers | ✅ | done |
| `Effects / Masks / Text / Transform` PropertyGroup accessors | ✅ | 🟢 Property tree | ✅ | P2 | 加 PropertyGroup hierarchy |
| `Remove()` | ✅ | ❌ | ✅ | P3 | 结构性，V3 capability |
| `Duplicate()` | ✅ | ❌ | ✅ | P3 | 结构性 |
| `CopyToComp(comp)` | ✅ | ❌ | ✅ | P3 | 结构性 |
| `MoveAfter/MoveBefore/MoveToBeginning/MoveToEnd` | ✅ | ❌ | ✅ | P3 | 结构性 |
| `SetParentWithJump(layer)` | ✅ | ❌ | ✅ | P3 | preserve world transform |
| `CanSetCollapseTransformation / CanSetTimeRemapEnabled` | ✅ | ❌ | ✅ | P2 | capability query |
| `TimeRemapEnabled / SetTimeRemap*` | ✅ | 🟢 R only | ✅ R/W | P2 | enable=structural |
| `ThreeDModelLayer`（Cinema 4D / GLB） | ✅ | ❌ | ✅ R only | P2 | |

### 2.4 Property / PropertyBase / PropertyGroup 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `MatchName / Name` | ✅ | ✅ R | ✅ | done |
| `Components / Dimensions` | ✅ | ✅ R (Components) | ✅ | done |
| `LockedRatio` (tdsb) | ✅ R only | ❌ | ✅ R/W | P2 |
| `IsSpatial` (tdb4) | ✅ R | ❌ | ✅ R | P1 |
| `Animated / Color / Integer / NoValue / Vector` (tdb4 flags) | ✅ R | 🟢 partial via Components | ✅ R | P1 |
| `DefaultValue / LastValue / NbOptions` | ✅ R | ❌ | ✅ R | P2 | tdb4 字节 RE |
| `MinValue / MaxValue / UnitsText` | ✅ R | ❌ | ✅ R | P2 | 内置 schema |
| `Value` (current) | ✅ R/W | ✅ R/W StaticValue | ✅ | done |
| `CanVaryOverTime` | ✅ R | ❌ | ✅ R | P1 | |
| `DimensionsSeparated` | ✅ R/W | ❌ | ✅ R/W | P3 | structural toggle |
| `Expression / ExpressionEnabled / CanSetExpression` | ✅ R/W | ✅ R/W | ✅ | done |
| `PropertyControlType` | ✅ R | ❌ | ✅ R | P2 | enum |
| `PropertyValueType` | ✅ R | ❌ | ✅ R | P2 | enum |
| `IsModified / Active / Elided / IsName Set` | ✅ R | ❌ | ✅ R | P2 | |
| `PropertyIndex / PropertyDepth` | ✅ R | ❌ | ✅ R | P2 | |
| `ParentProperty` (PropertyGroup back-ref) | ✅ R | ❌ | ✅ R | P2 | |
| `Selected / SelectedKeys` | ✅ R | 🗑️ | 🗑️ | done | runtime-only |
| `EssentialPropertySource` | ✅ R | ❌ | 🗑️ | — | runtime/EG |
| `AlternateSource / CanSetAlternateSource` (Layer-level, on Property) | ✅ R | 🟢 Layer.AlternateSource | ✅ | done |
| `ValueText` (formatted value) | ✅ R | ❌ | ✅ R | P3 | |
| **PropertyGroup ops**: `Properties / Property(key) / NumProperties / CanAddProperty` | ✅ | ❌ | ✅ | P2 | hierarchical access |
| `PropertyBase.Remove / Duplicate / MoveTo` | ✅ R/W | ❌ | ✅ R/W | P3 | 结构性 |
| `Keyframe.Time / FrameTime / Value / InInterpType / OutInterpType / TemporalEase / SpatialTangent` | ✅ | ✅ R/W | ✅ | done |
| `InsertKeyframe / DeleteKeyframe` | 🚧 limited | ✅ | ✅ | done |
| **Gradient** (ADBE Vector Grad Colors XML 解析 → color_stops + alpha_stops) | ✅ R/W | ❌ | ✅ R/W | P2 | XML in cdat |
| **MaskPropertyGroup** | ✅ | 🟢 Mask 类型化 | ✅ | done |

### 2.5 Footage / Source 域

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Path` | ✅ R/W | ✅ R/W | ✅ | done |
| `Width / Height / Duration / FrameRate / FrameDuration / PixelAspect` | ✅ R | 🟢 partial | ✅ R | P1 | proxy to source |
| `FootageMissing / HasAudio` | ✅ R | ❌ | ✅ R | P1 | |
| `StartFrame / EndFrame` | ✅ R | ❌ | ✅ R | P1 | |
| `AssetType` (placeholder/solid/file) | ✅ R | 🟢 implicit | ✅ R | P1 | discriminator |
| `MainSource` (typed: FileSource / SolidSource / PlaceholderSource) | ✅ | ❌ | ✅ | P2 | 加 typed sources |
| `File` (FootageItem.file convenience) | ✅ R | 🟢 via Source.Path | ✅ | P1 | |
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
| `EssentialGraphicsController` (controllers + override UUIDs) | 🚧 controllers exposed | ❌ | 🚧 controllers R | P3 |
| 自动 controller-override 解析 | ❌ | ❌ | ❌ | — |

### 2.9 杂项

| API | py-aep | 我们 | 目标 | Phase |
|---|---|---|---|---|
| `Application` 顶层 (version / project / etc) | 🚧 | 🟢 `*Project` 等价 | ✅ | P1 | 加 `Application` 或 `App` wrapper |
| `Application.Version` | ✅ R | ❌ | ✅ R | P1 | 从 head 解 |
| `Viewer / ViewOptions / View` | ✅ R only | ❌ | 🗑️ | — | UI state, runtime |
| `ImportOptions` | ❌ py-aep 也没 | ❌ | ❌ | — |

---

## 3. Phase 规划

每个 phase 独立可 ship；先做 P1（低风险、零 invariant 冲突）再分头进 P2/P3。

### Phase 1 — 低悬果：filter views + frame-time 伴生 + Project single-byte chunks（**第一波**）

**Target**: 50+ 新 API，零结构性写，纯 reader + 现有 chunk slice 设字段。

子任务（详 `workshop/plans/2026-05-26-py-aep-parity-p1-plan.md`）：

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
- **2B** CMS JSON (AE 24+) `ColorManagementSystem / LutInterpolationMethod / OcioConfigurationFile`；`WorkingSpace / DisplayColorSpace` R only；XMP R only
- **2C** Gradient (XML in cdat): `Gradient { ColorStops, AlphaStops }` + `ADBE Vector Grad Colors` Property 类型识别 → 解析 XML → 暴露结构化访问 + 写回（XML 重序列化）
- **2D** 类型化 Footage Sources: `FileSource / SolidSource / PlaceholderSource` 区分 + `Footage.MainSource` typed return + `MinValue / MaxValue / UnitsText / DefaultValue / LastValue / NbOptions / PropertyControlType / PropertyValueType` Property 读
- **2E** ImportPlaceholder: `Project.ImportPlaceholder(name, w, h, fps, dur)` — 跟 NewComposition 同模式
- **2F** ReplaceSource (Layer 级): `Layer.ReplaceSource(newSource, fixExpressions=false)` — splice source id
- **2G** PropertyGroup 链式访问: `Layer.PropertyGroup("ADBE Transform Group").Property("ADBE Opacity")` — public API 加 hierarchical accessor
- **2H** Marker properties on Layer + PropertyGroup access for Transform/Effects/Masks/Text
- **2I** LightSource (light layer 关联另一个 layer, AE 24+)
- **2J** ThreeDModelLayer R only
- **2K** TimeRemap R/W (enable=structural)

**估算**: 3-5 个会话；nnhd RE 需要 fixture probe，Gradient XML 需要找 fixture + 写测试。

### Phase 3 — 大域 / 结构性：Render Queue + Layer 增删移 + Composition.Renderer 写 + Essential Graphics + Guides

子任务：

- **3A** Render Queue 全栈 (`RenderQueue / RenderQueueItem / OutputModule`)，含全 enum + Settings + format options × 7 — 整个新域 (~3-5k LOC)
- **3B** Layer 结构性 ops: `Remove / Duplicate / CopyToComp / MoveAfter / MoveBefore / MoveToBeginning / MoveToEnd / SetParentWithJump`（全 V3 capability framework 内做）
- **3C** PropertyBase 结构性: `Remove / Duplicate / MoveTo / DimensionsSeparated R/W`
- **3D** Composition.Renderer 写（跨 renderer 切换 → prda 长度变结构性）
- **3E** Essential Graphics R only (controllers + override UUIDs，不做自动解析)
- **3F** Guides R/W (ruler 标尺辅助线，UI-only chunk)
- **3G** Composition Markers ：comp-level marker 增删（已 R/W setter；增删是结构性）
- **3H** Property.ValueText (formatted, requires AE schema database)

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

- Phase 1 plan: [`../plans/finish/2026-05-26-py-aep-parity-p1-plan.md`](../plans/finish/2026-05-26-py-aep-parity-p1-plan.md) (executed, archived 2026-05-26)
- Phase 2/3 plans: 写完 Phase 1 后再起，按需切
- Board `Active focus` 反映**当前 phase**
- Coverage docs 每个子任务完工同步更新

---

## 6. 相关文档

- 参照源: [`workshop/reference/py-aep/`](../reference/py-aep)
- 当前架构概览: 项目根 `CLAUDE.md` § 数据流 + § 硬约束
- 覆盖矩阵: [`../plans/coverage.md`](../plans/coverage.md) / [`../plans/coverage-detail.md`](../plans/coverage-detail.md)
- V2.2 ShapeLayer alpha: [`2026-05-22-v2-2-layer-creation-design.md`](2026-05-22-v2-2-layer-creation-design.md)
- V3 方向: [`2026-05-22-v3-direction.md`](2026-05-22-v3-direction.md)
- 已知陷阱: [`../scars/`](../scars)
