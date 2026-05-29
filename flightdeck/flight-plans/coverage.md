# Coverage 概览（精简入口）

> 要查具体 AE attribute → Go 字段对应，看 [coverage-detail.md](coverage-detail.md) 详细交叉表。
> 本文档只列：什么已 ship / 什么暂搁 / 什么不可达 / negative findings。

兼容声明：AE 2020 读下限 + AE 24+ 字段渐进写。所有 setter 都是 length-preserving（除少数 length-variable 替换：name / comment / expression / 字体名 / 文本内容）。

测试基线 / PASS count 见 `../board.md` `Last updated` 行（**唯一权威**）。

---

## ✅ 已 ship

### V2 结构性创建 (2026-05-22)

- `aep.NewProject(target ...AETarget) *Project` — 全新空 project，零参 = TargetAE2020；支持 TargetAE2020 / 2022 / 2025
- `proj.NewComposition(name, w, h, fps, duration) (*Composition, error)` — 在 root folder 新建空 comp，原子（warning/error → rollback），跟 `Open(...)` 出来的 comp 同构（所有 Set\* 立即可用）
- AE 2020 + AE 2025 ship gate PASS（`AE_SHIP_GATE=1 go test -run TestV2_1_AEShipGate -v`）
- 详 `../scars/ae25-acceptance-gate.md` 5 阶段 RE findings + canonical seed 策略

### Project / Composition / Item

- Project: `BitsPerChannel` R/W
- Composition（cdta）: `Name / FrameRate / Duration / Size / BGColor / ShutterAngle / ShutterPhase / MotionBlurAdaptive / MotionBlurSamplesPerFrame / WorkArea / DisplayStartTime / DisplayStartFrame / PixelAspect / ResolutionFactor` R/W
- Composition（PRin LIST）: `Renderer` **R only**（match-name；`ADBE Escher` = Advanced 3D / `ADBE Ernst` = Cinema 4D / `ADBE Standard` = Classic 3D）
- Composition（cdta flag bits）: `HideShyLayers / CompMotionBlur / Draft3D / FrameBlending / PreserveNestedFrameRate / PreserveNestedResolution` R/W
- Item: `Name / Comment / Label` R/W（Composition + Footage 共用）
- Footage: `Path` R/W
- Composition markers: 全 8 个 setter
- **Composition filter views (py-aep parity P1 1A)**: `TextLayers / ShapeLayers / CameraLayers / LightLayers / NullLayers / AdjustmentLayers / ThreeDLayers / GuideLayers / SoloLayers / AVLayers / CompositionLayers / FootageLayers / FileLayers / SolidLayers / PlaceholderLayers` — 15 个 filter helper，无写
- **Composition convenience (P1 1F)**: `NumLayers / HasAudio / TimeScale` — 3 个 helper（ActiveCamera + Markers field 早已 ship）
- **Layer convenience (P1 1E)**: `ContainingComp / HasVideo / HasAudio / AudioActive / AudioActiveAtTime / ActiveAtTime / Width / Height / HasTrackMatte / IsTrackMatte / AutoName / IsNameFromSource / RemoveTrackMatte` — 13 个 helper (Index / Type 早已直接 field 暴露)
- **ThreeDModelLayer R (P2a 2J)**: `LayerType3DModel` 枚举 + `Layer.IsThreeDModelLayer()` typed accessor；ldta byte `@0x83 == 0x05` (py-aep `LayerType.THREE_D_MODEL = 5`) 在 `inferLayerType` 派发。AE 24+ 3D Model layer 类型识别。read-only，无 fixture（synthetic byte-dispatch test）。
- **Application wrapper (P1 1I)**: `Application{Project} / Parse / ParseReader / Application.Version()` — py-aep `parse()` 入口对齐；Version 从 head chunk 解 (e.g. "17.7x45" = AE 2020 build 45)
- **Frame-time accessor (P1 1C)** — 用 owning comp 的 FrameRate 换算：
  - Layer: `InPoint() / OutPoint()` getters (filled gap of 既有 SetInPoint/SetOutPoint without R)；`FrameInPoint / FrameOutPoint / FrameStartTime` R/W (3 pair)
  - Composition: `DisplayStartFrame R` (SetDisplayStartFrame 早有)；`WorkAreaStartFrame / WorkAreaEndFrame / WorkAreaDurationFrame` R/W (3 pair)；`FrameDuration R`
  - Keyframe: `FrameTime` R/W (用 compFps，parser 注入)
  - Marker: `FrameTime / FrameDuration` R/W (用 compFps)
- **Property tdb4 flag readers (P1 1G)**: `IsSpatial / IsAnimated / IsColor / IsInteger / IsVector / IsNoValue / CanVaryOverTime` — 7 个 R only flag readers, 从 tdb4 metadata chunk (124B) 解 (offsets 来自 py-aep `binary/property_chunks.py::Tdb4Chunk`)；parser 新加 `Property.tdb4` 私有 ref。Standalone Property (tdb4=nil) 全 false fallback, IsAnimated 用 len(Keyframes) > 0.
- **Property tdsb flag readers/writers (P2a 2D, Task 3)**: `LockedRatio() / SetLockedRatio(v bool)` — 从 tdsb subprop flags chunk (4B) 读写 byte 2 bit 4 (py-aep: `locked_ratio` = bit 4)。parser 新加 `Property.tdsb` 私有 ref。Standalone Property (tdsb=nil) 返回 false；SetLockedRatio 返回 error。length-preserving。内部测试直接挂 tdsb 验证 reader/writer/邻位保留。
- **Layer ReplaceSource (P2a 2F, Task 4)**: `Layer.ReplaceSource(target AVItem, fixExpressions bool)` — 镜像 py-aep API；底层走既有 `SetSource` 路径。fixExpressions=true 时记 warning 到 Project.Warnings（不实现 symbolic execution）。
- **Gradient XML (P2b 2C)**: `Property.Gradient *Gradient` + `GradientColorStop / GradientAlphaStop` 类型 — R only。parser 识别 "ADBE Vector Grad Colors" 属性的 `GCst → GCky → Utf8` chunk 路径，解 `prop.map version="4"` XML（color stops + alpha stops + version）。rifx 加 `IDGCst / IDGCky` 常量。fixture: `flightdeck/charts/py-aep/samples/models/property/gradient.aep`（py-aep 自带样本）。**剩余**：XML 重序列化 / SetGradient / per-keyframe gradients（gcky 多 Utf8）—— P2 followup（需 fixture）。
- **Property derived state (P2c followup#2)**: `IsModified()` (animated/expression/value≠default) / `Enabled()` / `Active()` (alias) / `Elided()` (placeholder false) / `IsNameSet()` (Name != MatchName proxy) on Property；`IsModified()` on `AEPropertyGroup` (indexed groups modify on insert，否则递归)。Enabled 读 tdsb byte 3 bit 0（py-aep `_enable_flags` default = 1）。
- **AV Layer capability queries (P2c followup#2)**: `Layer.AVSource()` 解 SourceID 到 Composition/Footage AVItem；`Layer.CanSetCollapseTransformation()`（precomp 或 solid 源 → true）；`Layer.CanSetTimeRemapEnabled()`（源 Duration > 0 且非 still → true）。py-aep `AVLayer.can_set_*` parity。
- **Project XmpPacket R (P2c followup#2)**: `Project.XmpPacket() string` — 走 `root.Trailing`（AE 把 XMP UTF-8 XML 写在 RIFX 后面），WriteAEP 已 round-trip。W deferred（XML 结构校验风险）。
- **PropertyGroup hierarchy (P2c)**: `AEPropertyGroup` 树 + `PropertyBase` interface — 把 tdgp 层级在 parser 时镜像出来，与 flat `Layer.Properties / Effects / Markers / Masks` 并存（叶子是同 `*Property` instance，pointer identity 保留）。Layer API: `PropertyTree()` / `PropertyGroupByMatchName(name)` / `PropertyByPath(...matchNames)` + 11 个 typed group accessor。Group API: `Property(matchName) / Group(matchName) / ChildByIndex(i) / NumProperties / ParentGroup / PropertyByPath / PropertyIndex(child) / Depth()`。Leaf back-ref: `Property.ParentGroup()`。
- **Property metadata reads (P2c followup)**: `ControlType()` / `ValuePropertyType()` — 从 tdb4 flags 推导。`MinValue()` / `MaxValue()` — 从 tdum/tduM chunks 解码。`UnitsText()` — 静态 map。`PropertyIndex()` / `PropertyDepth()` — 委托 parentTreeGroup。`DefaultValue / LastValue / NbOptions` — transform 硬编码默认值表 + pard chunk 解析基础设施（7 种 control type extractor）。parser 在 collectEffects 后自动 merge pard 元数据。
- **Footage convenience (P1 1H)**: `AssetType / File / FootageMissing / HasAudio / StartFrame / EndFrame` — 6 个 helper. parser 新加 `Footage.sspcChunk` ref。**修复 latent bug**: 真实 AE sspc 222B 布局，Width/Height 在 @0x20/@0x24 (不是 @0/2)；synthetic 4B sspc fallback 保留。Real AE 文件 W/H 之前一直读 0，现 OK。
- **Project views (P1 1B)**: `Footages() / RootFolder() / LayerByID(id) / EffectNames()` — 4 个 API. EffectNames 从 root-level `Pefl` LIST → `pjef` Utf8 entries (rifx 加 IDPefl/IDPjef 常量)。LayerByID 是跨 comp lookup. RootFolder 拿 Folders[0]. Footages 是 .Footage slice 的命名 alias.
- **Project single-field setting chunks (P1 1D)**: 8 个 R/W (Revision 仅 R)，**AE 2020/2025 ship-gate PASS**（bisect 8/8 + combined run accept；AE-visible 值 byte-identical with Go-side input）
  - `Revision()` R — head[18..19] uint16 BE (per-save counter)
  - `LinearBlending` R/W — lnrb add/remove chunk (1 byte `0x01` payload, insert 在 `cpid` 之后)；AE ScriptingAPI 读出 true ✓
  - `LinearizeWorkingSpace` R/W — lnrp 同 lnrb pattern；**chunk 写法 AE-byte-identical** 但 ScriptingAPI 读 false (OCIO/CMS联动 quirk，AE 自己 set 后 reload 也读 false，详 `scars/project-flag-chunks-lnrb-lnrp.md`)
  - `CompensateForSceneReferredProfiles` R/W — acer[0] bool
  - `AudioSampleRate` R/W — adfr f64 BE, validate {22050/32000/44100/48000/96000}
  - `WorkingGamma` R/W — dwga[0] selector (0→2.2, ≠0→2.4)
  - `GpuAccelType` R/W — gpuG → Utf8 (length-variable splice)
  - `ExpressionEngine` R/W — ExEn → Utf8, validate {"extendscript", "javascript-1.0"}
  - rifx 加 IDAcer/IDAdfr/IDDwga/IDLnrb/IDLnrp/IDGpuG/IDExEn 常量
- **Project nnhd display settings (P2b 2A)**: 8 个 R/W，nnhd chunk (40 bytes) — py-aep `NnhdChunk` parity
  - `FeetFramesFilmType` R/W — byte 8 bit 7 (0=MM35, 1=MM16)
  - `FootageTimecodeDisplayStartType` R/W — byte 9 (0=Start0, 1=UseSourceMedia)
  - `TimecodeDefaultBase` R/W — bytes 14-15 u2 BE (1-999)
  - `FramesCountType` R/W — byte 20 (0=Start0, 1=Start1, 2=TimecodeConversion)
  - `DisplayStartFrame` R/W — derived from frames_count_type % 2
  - `FramesUseFeetFrames` R/W — byte 11 bit 0
  - `TimeDisplayType` R/W — byte 8 bits 6-0 (0=Timecode, 1=Frames)
  - `TransparencyGridThumbnails` R/W — byte 25 bool
- **Project CMS settings (P2b 2B, AE 24+)**: JSON Utf8 chunk — py-aep `NnhdChunk` parity
  - `ColorManagementSystem` R/W — 0=Adobe, 1=OCIO（enum 校验）
  - `LutInterpolationMethod` R/W — 0=Trilinear, 1=Tetrahedral（enum 校验）
  - `OcioConfigurationFile` R/W — string path
  - `WorkingSpace` R only — 从 CMS JSON `baseColorProfile.colorProfileName` 取
  - **Setters 仅在 `cmsUtf8 != nil` 时工作**：CMS chunk 的容器位置 / 是否需 LIST 包装尚未 RE，无 fixture 时拒绝写而非自动创建（避免产生 AE 拒收的文件）。`cmsSettings` JSON 解析失败时 emit `p.Warnings` 而非静默 fallback。`DisplayColorSpace` 暂搁（separate chunk 没 RE，旧 stub 永远返回 "None"，2026-05-27 删除）。
- **Footage discriminator**: 新增 `IsPlaceholder` 字段（opti tag = "Plac"）;`Footage.{IsSolid, IsPlaceholder}` 互斥三态（file / solid / placeholder）

### Layer

- 基础（ldta）: `Name / Comment / Label / InPoint / OutPoint / StartTime / Stretch / Parent / Source / TrackMatte / TrackMatteLayer / AutoOrient` R/W
- Flag bit: `Visible / Solo / Shy / Locked / EffectsEnabled / MotionBlur / AudioEnabled / FrameBlend{Enabled,PixelMotion} / CollapseTransform / Is3D / IsAdjust / IsGuide / IsNull / MarkersLocked / PreserveTransparency / Quality / SamplingBicubic / BlendingMode` R/W
- Markers: 全 8 个 setter
- AlternateSource（AE 18+ EGP slot）: R/W；首次启用 = 结构性 refused
- **Transform**: 8 typed setter (`SetAnchorPoint / SetPosition / SetScale / SetRotation / SetRotateX / SetRotateY / SetOrientation / SetOpacity`)
- **AudioLevels**: `SetAudioLevels([L, R])`
- **Camera 专属**: 13 typed accessor pair (`CameraZoom / DepthOfField / FocusDistance / Aperture / BlurLevel + Iris × 8`)，`SetCameraDepthOfField(bool)` 自动 0/1
- **Light 专属**: 11 typed accessor pair (`LightColor / Intensity / ConeAngle / ConeFeather / FalloffType / FalloffStart / FalloffDistance / CastsShadows / ShadowDarkness / ShadowDiffusion`) + `LightKind` ldta `@0x88` R/W；`SetLightCastsShadows(bool)`
- **LightSource R/W (P2a 2I, AE 24+)**: `Layer.LightSource() / SetLightSource(target *Layer)` — Environment-type light 指向同 comp 另一 layer 做光源；底层走 ldta `@0x28`（跟 AV `SourceID` 共用 slot，按 `Layer.Type` 重解释），sentinel `0xFFFFFFFF` = 无源。py-aep `LightLayer.light_source` parity，验证规则：caller 必须 Light / target 同 comp / target 不能 Light/Camera/3D/self。length-preserving，roundtrip + 6 validation test 全 PASS。
- **Material Options 3D AV**: 17 typed accessor pair + `MaterialCastsShadowsMode` 三态 enum（Off/On/Only）
- **Geometry Options 3D AV**: 3 typed accessor pair（PlaneCurvature / PlaneSubdivision / BevelDirection）

### Property / Keyframe

- `StaticValue` R/W（对 effect 参数也直接生效）/ `Expression` 完整 R/W（length-variable）/ `ExpressionEnabled` R/W（反语义解析）
- Keyframe: `Time / Value / InInterp / OutInterp / Temporal{In,Out}Ease / Spatial{In,Out}Tangent` 全 R/W（对 effect 参数 keyframe 也直接生效）
- 增删 keyframe: `InsertKeyframe(time, value)` + `DeleteKeyframe(i)`（要求 ≥ 1 既有 keyframe 作 layout 模板）

### Mask

- 顶层: `Mode / Inverted / Color / Closed / Locked / MotionBlur` R/W；`Feather / Opacity / Expansion` R + 顶层 setter
- 路径动画: `PathKeyframes` R only（顶点重写 = 结构性 ❌）

### Shape

- 自由路径: `Layer.ShapePaths` R only
- 参数化 Rect / Ellipse / Star: `ShapePrimitives` R/W（子字段 `*Property`）

### V3 Composition layer structural ops (2026-05-28)

AE-acceptance gate via `re_delete_layer_baseline.aep` 3-solid baseline + per-mode setup; ship-gate green AE 2020 + AE 2025.

- `(c *Composition) DeleteLayer(index int) error` (V3 Phase 2) — 0-based delete; adaptive 16-chunk splice (AE-saved layers + Go-built 2-chunk layers alike); ParentID / TrackMatteLayerID orphans reset to 0 on remaining layers; head counter NOT decremented (Inv-9 monotonic). Refuse-cases: index OOR / non-AV / sole layer / corruption. **8/8 ship-gate PASS** (AE 2020 + AE 2025 × baseline/middle/parent/matte; matte split per AE 23+ TrackMatteLayerID field).
- `(l *Layer) MoveAfter(other) / MoveBefore(other) / MoveToBeginning() / MoveToEnd() error` (V3 Phase 4 follow-up) — AE ScriptingAPI / py-aep parity wrappers over `Composition.MoveLayer`. Each finds the layer's current slice index via pointer identity in `l.comp.Layers` (Layer.Index is parse-time and can be stale post-mutation). Refuse: nil-other / self / cross-comp / missing comp back-ref. No ship-gate needed — delegate to already-validated MoveLayer.
- `(c *Composition) MoveLayer(from, to int) error` (V3 Phase 4) — 0-based reorder; same adaptive block-splice machinery as DeleteLayer/DuplicateLayer (Layr + Ewst + leaf followers moved atomically); no ID alloc, no follower mutation; `from == to` is no-op; refresh `Layer.Index` for every layer in c.Layers post-move. Refuse-cases: from/to OOR / comp lacks back-ref / source lacks Layr back-ref / backref corruption. **6/6 ship-gate PASS** (AE 2020 + AE 2025 × first_to_last/last_to_first/mid_swap).
- `(c *Composition) DuplicateLayer(index int, name string) (*Layer, error)` (V3 Phase 3) — 0-based dup at source's index, source pushed down; new ID via `proj.allocItemID()`; clone block = adaptive deep-copy of source's [Layr + Ewst + leaf followers] with fresh Data slices (concurrent-mutate safe); ldta @0x00..0x03 = new ID, all other body verbatim (SourceID @0x28 / ParentID @0x84 / TrackMatte @0x6B copied via verbatim clone); name = caller-supplied (length-variable Utf8 rewrite). F6: children's outgoing ParentID NOT rewritten — clone is a sibling shadow. Refuse-cases: empty name / index OOR / non-AV / TrackMatte != None (F2 quirk deferred to Phase 3.1) / corruption. **6/6 ship-gate PASS** (AE 2020 + AE 2025 × solo/dup_parent/dup_child; matte by-design refused).
- `(c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error)` (V3 Phase 5C) — Cross-comp deep-clone of src from same-Project sibling comp at atIdx (0-based); new ID via `proj.allocItemID()`; clone block = adaptive deep-copy of src's [Layr + Ewst + leaf followers] with ldta deltas (ID @0x00..0x03 / TrackMatteNone @0x6B / ParentID @0x84..0x87 zeroed / explicit matte @0xA0..0xA3 zeroed); SourceID + Name verbatim; atomic mutation (snapshot + rollback on re-parse warning). Refuse-cases: nil src / dest comp lacks itemList back-ref / atIdx OOR / src detached / same-comp redirect / cross-Project / non-AV / direct pre-comp loop / src lacks comp back-ref / corruption. **6/6 ship-gate PASS** (AE 2020 + AE 2025 × basic/footage/precomp, 2026-05-29; AE accepts Go-emitted file, clone refs source item verbatim with parent + track matte reset).

### V2.2 alpha ShapeLayer 写路径 (2026-05-25, iter-7/8)

- `(c *Composition) NewShapeLayer(name string) (*ShapeLayer, error)` — 新建空 ShapeLayer，AE 2025 接受 + comp.layers.length=1
- `(s *ShapeLayer) RootGroup() *VectorGroup` — 取顶层 Contents 容器
- `(g *VectorGroup) AddRect() (*RectNode, error)` / `AddFill() (*FillNode, error)` — 加入参数化 shape kid
- Setter: `RectNode.SetSize([w, h]) / FillNode.SetColor([r,g,b,a])` static-only
- **AE 接受 gate**: 通过 embed tolerance.aep 抽出的 3 处 boilerplate 字节 (Transform Group 1842B + Rect body 448B + Fill body 426B in `internal/aep/templates/`)；详 `../scars/v2-2-aelayer-structure.md`
- **V2.2 alpha 限制**（V2.2.1 候选）:
  - Ellipse / Path / Stroke: Go 端能 emit + parse，AE 会 silent drop（需各自 fixture + embed bytes）
  - Fill Color 编码: cdat scalar 跟 JSX 0-1 input 不对齐（tolerance 0.5 → 0x406fe0... ≈ 255），可见色可能错
  - Keyframe 持久化（Rect Size / Fill Color / Layr Position 全部）: 不持久化，first kf 作 static fallback
  - Layr Transform 的 Anchor / Scale / Rotation / Opacity: runtime-only 不持久化
  - Rect Direction / Position / Roundness: runtime-only 不持久化

### Text

- 基础: `Text` R/W (length-preserving)；`Fonts []string` + `AddFont(name)` 追加；`FontIndex` R/W
- per-run（22 setter）: `FontSize / FillColor / StrokeColor / StrokeWidth / ApplyStroke / Tracking / Leading / AutoLeading / BaselineShift / HorizontalScale / VerticalScale / Tsume / FauxBold / FauxItalic / CapsOption / BaselineOption / StrokeOverFill / AutoKernType / NoBreak / LineJoinType / DigitSet`
- per-paragraph（9 setter）: `Justification / FirstLineIndent / StartIndent / EndIndent / SpaceBefore / SpaceAfter / AutoHyphenate / LeadingType / HangingRoman / Direction`
- btdk-root: `SetManualKerning(values []int)`（要求 `/8` slot 已 emit）
- 字体 axes: `FontAxes [][]float64` R only（写绑定到 PostScript 名切换）

---

## 🗑️ 暂搁（fixture / 环境阻塞，遇到需要再做）

| 项 | 阻塞原因 |
| --- | --- |
| `AVLayer.environmentLayer` | 需 equirectangular 360° 视频素材 |
| `TextDocument.ligature` | 默认字体 ligature=false 设值无 diff；需带 OT `liga` feature 的字体 fixture |
| `maskFeatherFalloff` | JSX 设值后 mkif 字节零变化；疑似在 mask sub-property 树，需深挖 RE |
| 4D 颜色 32bpc 范围 (0..1 vs 0..255) | 需 32bpc 项目 fixture 验证 |
| mkif 残余字节 `@0x10 / @0x18 / @0x20-0x27` | 未 RE；roundtrip 走 `Mask.MkifRaw` 保留原字节，不假设 48 字节全已知 |
| **Property metadata (P2b 2D)** | 需 pard chunk reader + specs.py schema table 端口，复杂 RE 工作 |
| **ImportPlaceholder (P2a 2E)** | opti chunk format 未 RE；user fixture (`re_placeholder_ae20.aep`) 显示生成的 opti tag AE 不识别。2026-05-27 删除合成 builder，等真 fixture 出来后再起 |
| **CMS chunk 创建（P2b 2B）** | 文件原本没 CMS chunk 时 setter 拒写。CMS Utf8 在 root 下的精确位置 / 容器（裸 Utf8 vs LIST 包装）未 RE；空挂 AE 可能拒文件 |
| **`DisplayColorSpace`（P2b 2B）** | py-aep 提示 separate chunk，位置未 RE；旧 stub 永远 "None"，2026-05-27 删除 |

## ❌ 不可达（length-preserving 写约束之外 / AE 限制）

| 项 | 原因 |
| --- | --- |
| Layer / Effect / Mask vertex / ShapePrimitive **增删** | 结构性，破坏多个父 LIST 大小 |
| `Property.timeRemapEnabled` toggle | AE 加/删 2 个 identity keyframe（结构性） |
| `Property.dimensionsSeparated` | AE 拆 Position 成 3 个 1D 属性（+306 字节非局部改动） |
| `lineOrientation` 横/竖排切换 | layer-local 坐标重排 + 多字段连锁 |
| `MaskPropertyGroup.rotoBezier` | 切换重写整个 shape 顶点表示（+16 字节，4500+ byte-diff） |
| Motion Graphics Template / EP 模板 binding（除 `alternateSource`） | 跨 chunk 复杂结构，P3 罕用 |
| Project 渲染设置（`gpuAccel / colorSpace / expressionEngine`） | P3，AE 24+ 大多锁定为 default。`renderer` 已 ship R；setter 仍是 P3（prin 双段 NUL-sep + prda 长度随 renderer 变） |
| Adobe World-Ready composer 切换 | P3 |
| 手动 kerning **首次启用** | 结构性添加（需 AE 先 emit `/8` slot） |
| Camera `FilmSize` setter | ldta `@0x98` 持久化但 ScriptingAPI 不暴露写路径（详 `scars/camera-filmsize-ldta-write-blocked.md`） |

## ⚠ Negative findings（runtime-only / AE ScriptingAPI 限制）

| 项 | 结论 |
| --- | --- |
| `CompItem.dropFrame` | AE 不持久化（脚本可设可读，字节零变化）—— FrameRate NTSC 自动推断 |
| `TextDocument.fontLocation` | AE 不持久化（runtime 从系统字体注册表 join 路径） |
| Variable fonts axes **写** | `TextDocument.fontVariation` 不存在；唯一写途径 = 切已加载的 named-instance |
| `composerEngine` / `everyLineComposer` | AE 24+ 只接受 UNIVERSAL；其它写入抛错 |
| `setAlternateSource(item)` AE 脚本行为 | 自动包 wrapper precomp；我们 setter 不包装 |
| AE 24/25 ldta 加长 | 实测仍 164 字节（与 AE 23+ 一致），未见第三种长度 |
| cdta tail (≥ 0xCC) | 不存在 —— cdta 总长就是 0xCC=204 |

---

## ❓ 剩余可探方向（非 candidate 列表，需要新发现才动）

可达字段约 99% 已 ship。继续动需要：

1. **同-renderer 范围内 `Composition.SetRenderer`** — 跨 renderer 切换属结构性（prda 长度变），同 renderer 安全；价值小。
2. **ldta `@0x60-0x82` / `@0x8C-0x9F` 零值区 probe** — 高密度 JSX layer-flag 探针，可能挖出 1-2 个零散 flag 或全 negative。
3. **Footage proxy 字段** — 大部分结构性。
4. **Project nhed/nnhd 扩展字段** — 除 BitsPerChannel 外的字节，可能持 ColorSpace / Working Color Profile。
