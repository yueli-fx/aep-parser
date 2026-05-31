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
- `(c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error)` (V3 Phase 5C + 5C.1) — Deep-clone of src into c at atIdx (0-based); new ID via `proj.allocItemID()`; clone block = adaptive deep-copy of src's [Layr + Ewst + leaf followers] with ldta deltas (ID @0x00..0x03 / TrackMatteNone @0x6B / ParentID @0x84..0x87 zeroed / explicit matte @0xA0..0xA3 zeroed); atomic mutation (snapshot + rollback on re-parse warning). **Same-Project** (5C): src from sibling comp; SourceID + Name verbatim. **Cross-Project** (5C.1): src from a different `*Project` — imports src's reachable item closure (footage + precomp, transitively, via `locateItemBlockByID` recursing into folder Sfdr) into c's Project at root level with fresh item IDs, dedup'ing file-backed footage by Path, then remaps the clone's SourceID @0x28 + AlternateSourceID through a srcItemID→destItemID map; seven-way dest snapshot + warnings-as-failure rollback covers both phases (closure import + layer splice); folders not recreated; assert-based gate (no AE-native cross-Project op → no byte-diff baseline). Shared `spliceLayerClone` core (sourceRemap = identity same-Project / itemIDMap cross-Project). Refuse-cases: nil src / dest lacks itemList / atIdx OOR / src detached / same-comp redirect / non-AV / direct pre-comp loop (same-Project only) / src lacks back-ref / corruption / (cross) dest|src no rootFold / dangling closure source. **12/12 ship-gate PASS** — same-Project 6/6 (AE 2020+2025 × basic/footage/precomp) + cross-Project 6/6 (AE 2020+2025 × footage/precomp/dedup), 2026-05-29; AE accepts Go-emitted files, sources resolve (imported/dedup'd), footage not duplicated on path match.
- `(p *Project) DuplicateComposition(src *Composition, name string) (*Composition, error)` (V3 Phase 5D) — same-Project comp deep-clone mirroring AE `CompItem.duplicate()`; new comp item ID via `allocItemID()` (idta @0x10); each layer gets a fresh ID (ldta @0x00); intra-comp ParentID @0x84 + TrackMatteLayerID @0xA0 **remapped** through a srcLayerID→dupLayerID map (the per-comp-clone machinery vs InsertLayer's zero-reset); SourceID @0x28 verbatim (footage/precomp items SHARED, not cloned); name = caller-supplied (length-variable Utf8); rootFold splice after src's sibling run + reparse + atomic rollback. Refuse-cases: nil src / project back-ref missing / src itemList back-ref missing / src not in this Project / empty name / src Item not in rootFold / layer ldta too short. **2/2 ship-gate PASS** (AE 2020 + AE 2025, 2026-05-29; AE confirms dup's parent ref resolves within the dup, sources shared). Scoped to comps (only item kind with a scripting-RE path); generic `DuplicateItem` umbrella + footage/folder deferred (no scripting API). Unblocks Phase 5C.1 cross-Project InsertLayer.

### V2.2 alpha ShapeLayer 写路径 (2026-05-25, iter-7/8)

- `(c *Composition) NewShapeLayer(name string) (*ShapeLayer, error)` — 新建空 ShapeLayer，AE 2025 接受 + comp.layers.length=1
- `(s *ShapeLayer) RootGroup() *VectorGroup` — 取顶层 Contents 容器
- `(g *VectorGroup) AddRect() (*RectNode, error)` / `AddFill() (*FillNode, error)` — 加入参数化 shape kid
- Setter: `RectNode.SetSize([w, h]) / FillNode.SetColor([r,g,b,a])` static-only
- **AE 接受 gate**: 通过 embed tolerance.aep 抽出的 boilerplate 字节 (Transform Group 1842B + Rect 448B + Fill 426B + **Ellipse 730B** in `internal/aep/templates/`)；详 `../scars/v2-2-aelayer-structure.md`

#### V2.2.1 (2026-05-29) — Ellipse embed bytes + AE 2020 ldta 地基修复
- `(g *VectorGroup) AddEllipse() (*EllipseNode, error)` + `SetSize / SetPosition` — ✅ **AE 2020+2025 双版本 ship-gate PASS**。embed `v2_2_shape_ellipse_body.bin` + overwrite Ellipse Size/Position cdat（offset 0, f64 BE）。`TestV2_2_Ellipse_AEShipGate_AE20{20,25}`（assert-based：AE 接受层不 silent-drop + re-save cdat 保留值）。
- **AE 2020 地基 bug 修复**：`buildLdtaBytes` 此前硬编码 164B ldta，AE 2020 判**所有** shape 图层（含 Rect+Fill）损坏并跳过；从未发现因 AE-2020 shape gate 长期 skip。改为按 target 分支（capability matrix `LdtaSize`：160 AE2020/22 / 164 AE25）。`TestLowerShapeLayer_LdtaSizeByTarget`。**Rect+Fill 在 AE 2020 现亦有效**（同 ldta 路径，Ellipse gate 已证该路径）。详 `../incident-reports/ae2020-shape-ldta-164-corrupt.md`。
#### V2.2.1 子项② (2026-05-29) — Path embed+splice + ldat 编码 RE 修复
- `(g *VectorGroup) AddPath() (*PathNode, error)` + `SetVertices / SetClosed` — ✅ **AE 2020+2025 双版本 ship-gate PASS**。embed `v2_2_shape_path_body.bin`（AE-native scaffolding）+ splice `encodeBezier` 几何。`TestV2_2_Path_AEShipGate_AE20{20,25}`（distinct 三角形，re-save + 解析器反归一化解 anchor 验证）。
- **ldat 顶点编码 bug 修复**：`encodeBezier` 曾存 `[anchor, in_i, out_i]`（本顶点 in/out），AE 实为 `[anchor, anchor+out_i, anchor_{i+1}+in_{i+1}]`（本顶点 out 控制点 + 下一顶点 in 控制点，wrap mod n，bbox 归一化）。`TestEncodeBezier_LdatMatchesAELayout`，验证与 AE-native 字节一致。详 RE：`../sketches/2026-05-29-path-embed-re-findings.md`。
- **from-scratch path 崩溃 AE 2020**（0::42）→ 必须 embed（同 Ellipse 教训，但 path 是变长几何 splice，非 overwrite-in-place）。

#### V2.2.1 子项③ (2026-05-29) — Stroke embed + Fill/Stroke Color 编码 RE 修复
- `(g *VectorGroup) AddStroke() (*StrokeNode, error)` + `SetColor/SetWidth/SetOpacity` + Cap/Join/Miter（子项⑩）+ BlendMode/CompositeOrder（子项⑪）+ `Taper()/Wave()`（子项⑫）— ✅ **AE 2020+2025 双版本 ship-gate PASS**。embed `v2_2_shape_stroke_body.bin`（含 Taper 6 + Wave 3 active slot；Dashes 仍空 placeholder）+ overwrite cdat。`TestV2_2_Stroke_AEShipGate_AE20{20,25}` + `TestV2_2_StrokeTaperWave_AEShipGate_AE20{20,25}`。
- **shape 颜色编码 RE 修复**：AE 存 `[A,R,G,B]×255` f64（非原始 `[r,g,b,a]×1.0`）。`encodeShapeColorBE` 统一 Fill+Stroke。修了长期 deferred 的"Fill Color 编码不准"——`lowerFillNode` 此前写原始 RGBA，可见色错。`TestLowerFillNode_ColorEncodingARGB255` + Ellipse gate re-save Fill 颜色校验。
- 清理：移除 from-scratch 死代码 `nodeBodyTdgp` / `emptySubPropPlaceholder`（5 个 shape kind 全 embed）。
- **V2.2.1 全部 5 shape kind（Rect/Ellipse/Path/Fill/Stroke）+ 地基 ldta + 颜色编码均 ship**。

#### V2.2.1 子项④ (2026-05-29) — shape keyframe 持久化（Size + Color + Ellipse Position）
- **Rect/Ellipse Size**（non-spatial Vec2）+ **Fill/Stroke Color**（spatial-style dim4 ARGB×255）+ **Ellipse Position**（spatial motion-path Vec2 bpk 104）keyframe 持久化 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_RectKf_*` + `TestV2_2_FillKf_*` + `TestV2_2_EllKf_*`；re-save 解 numKf+值）。Ellipse Position 仅 AE 自动 ~0 spatial 切线与原生不同（AE recompute），值往返正确；`valueLayout.motionPath` 在 0x08 写标志。
- 机制 `injectAnimatedStream`：static tdbs 的 cdat ↔ animated `LIST(list)(lhd3+ldat)`；patch tdb4 标志（@0x05 `&=~1`、@0x44 `=1`、@0x4f `&=~1`）。ldat 与 AE 原生字节一致。`encodeKeyframes` non-spatial（value@0x08 bpk 88）/ spatial（value@0x38）两布局。**坑**：`rifx.IDTdb4` 是大写 legacy，实际小写 `tdb4`。
- **仍 deferred keyframe**:
  - **Path**（bezier keyframe，逐帧 shap）：V2.3+ 级别。

#### V2.2.1 子项⑤ (2026-05-30) — Layr Transform Position keyframe 持久化（Path B：combined ADBE Position）
- **Layr Position**（spatial 真运动路径）keyframe 持久化 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_LayrPosKf_*`；re-save 解 numKf+值）。
- **磁盘形态**：combined `ADBE Position`（comp=3 spatial dim-3，bpk **128** = 0x38 + 3·3·8），value@0x38 X / @0x40 Y / @0x48 Z(=0)；motion-path 标志@0x08。runtime API 仍是 2D（`[2]float64`，Z 钉 0）。复用 `encodeKeyframes`（spatial 分支已泛化到 dim3）+ `injectAnimatedStream`，新 helper `injectAnimatedLayerPosition`（`lower_layer.go`）。
- **关键 RE 发现（坑）**：AE 仅在 Position 被 **set/animated** 时才写 combined `ADBE Position` cdat；默认/未触碰时只留分离维 `Position_0/_1`（旧 tolerance 模板即此态，故无 combined slot）。新模板 `templates/v2_2_transform_group_body.bin` 改从 `v2_2_shape_transform_pos.aep`（position 设为静态非默认值 [500,300] 的 shape 图层）提取 → 含 combined Position 静态 cdat（17 children，旧 15 + combined Position）。重生：`gen_shape_transform_pos.jsx` → `extract_transform_group`。
- 模板为 **shared**（所有 from-scratch shape 图层 transform group 都用它）→ swap 后重跑全部 shape ship-gate（Ellipse/Path/Stroke/RectKf/FillKf/EllKf）双版本均仍 PASS。

#### V2.2.1 子项⑥ (2026-05-30) — Layr Transform 全通道 keyframe + static 持久化（Anchor/Scale/Rotation/Opacity）
- **Anchor / Scale / Rotation / Opacity** keyframe + 静态持久化 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_XfKf_*`；AE 读回 user 单位 anchor 50/60·scale 150/200·rot 90·opacity 50 正确 + re-save numKf/bpk 往返）。
- **各通道磁盘编码**（RE 自 `v2_2_transform_kf_re.aep`）：
  - **Anchor**：3D spatial motion-path（bpk-128，value@0x38，[x,y,0]）— 与 Position 同布局。
  - **Scale**：3D **non-spatial**（bpk-128 = 0x08+5·3·8，value@0x08），值 **÷100**，Z(depth)=**1.0**（=100%）。
  - **Rotation**：1D non-spatial（bpk-48，value@0x08），degrees 原值。
  - **Opacity**：1D non-spatial（bpk-48，value@0x08），值 **÷100**。
- **归一化非对称（坑）**：写盘归一（scale/opacity ÷100），但 parser 读**原始**盘值（`list_props`: Scale `[1.2,1.3,1]`、Opacity `0.8`）→ 不反归一。ship-gate 用 AE `keyValue` 验 user 单位（AE 自己反归一），Go re-parse 只验 numKf/bpk。
- 新 helper（`lower_layer.go`）：`lowerTransformVec2Spatial`（Anchor/Position）、`lowerTransformScale`、`lowerTransformScalar`（Rotation scale=1 / Opacity scale=0.01）；animated→inject，static→overwrite cdat。
- 模板再扩到 **25 children**（源 `v2_2_shape_transform_full.aep`：5 通道全设静态非默认值；旧默认值被 AE elide）。模板 shared → 全部既有 shape ship-gate 重跑双版本仍 PASS。新增 public API `ShapeLayer.AnchorPoint()`（补齐 5 通道 shorthand；alpha）。
- **仍 deferred**: 3D 通道（Orientation/RotateX/Y/Position_Z）；Path keyframe（V2.3）。

#### V2.2.1 子项⑦ (2026-05-30) — Rect Position + Roundness 持久化（static + keyframe）
- **Rect Position + Roundness** static + keyframe 持久化 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_RectSubKf_*`；AE 读回 pos 40/50·roundness 20 + re-save numKf/bpk 往返）。
- **磁盘编码**（RE 自 `v2_2_rect_subprops.aep`）：**Rect Position** = spatial Vec2 motion-path（bpk-104，value@0x38，[x,y]）— 与 Ellipse Position 同布局；**Roundness** = 1D non-spatial（bpk-48，value@0x08，原值）。
- 新 helper `lowerShapeVec2 / lowerShapeScalar`（`lower_shape_node.go`）：animated→inject，static→overwrite cdat。`lowerRectNode` 现持久化 Size+Position+Roundness 三流。
- 富化 rect body 模板（9 children；源 `v2_2_shape_rect_full.aep`，Size/Position/Roundness 全设静态非默认 → AE 不 elide）。仅 rect body 变更（其它 4 shape body 字节不变）。
#### V2.2.1 子项⑧ (2026-05-30) — Stroke Opacity + Width keyframe 持久化
- **Stroke Opacity + Width** keyframe 持久化 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_StrokeKf_*`；AE 读回 opacity 50·width 20 + re-save numKf/bpk）。
- **磁盘编码**（RE 自 `v2_2_stroke_kf_re.aep`）：两者均 **1D non-spatial（bpk-48，value@0x08，原值无归一化）**。注意：Stroke Opacity 存**原始 %**（100/50），**不**像 Layr Opacity ÷100。Width = 原始 px。
- stroke body 模板**已含** Opacity/Width cdat slot（无需富化模板）→ 仅把 animated 路径从「first-kf 折叠为 static」改为真正 `lowerShapeScalar` inject。static 路径字节不变（`encode1D`==`encodeF64sBE`）。
#### V2.2.1 子项⑭ (2026-05-31) — Gradient fill (SetGradient: color + alpha stops)
- `(g *VectorGroup) AddGradientFill()` → `GradientFillNode`（`SetColorStops`/`SetAlphaStops` ≥2 stop + 范围校验、`Gradient()` getter）— ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_GradientFill_AEShipGate_AE20{20,25}`：一层 rect+gradient-fill，3 色标 red/green/blue，re-save 经 read 路径解码校验）。
- **磁盘编码**：色标存 `ADBE Vector Grad Colors` 的 `LIST(GCst) → LIST(GCky) → Utf8` prop.map XML（version='4'）。色标数组 6 float `[off,mid,r,g,b,1]`，alpha 3 float `[off,mid,a]`，Alpha Stops 在 Color Stops 前 + 各带 Stops Size，尾 `Gradient Colors=1.0`。`EncodeGradientXML` = `ParseGradientXML` 的逆，round-trip 自洽。
- **length-variable 写**：覆 Utf8 XML 改字节长 → **rifx.Chunk.Write 自动 bottom-up 重算所有 LIST size**（同 `Footage.SetPath` 机制），GCst 的 tdb4-124B 非冗余长度头，无需手动 fixup。
- **read 早已 ship**（`ParseGradientXML` + `Property.Gradient`，`TestGradient_FixturePyAep`）；本子项补 write。
- **deferred（elision/coupling + ScriptingAPI 封锁）**：Grad Type/Start Pt/End Pt（模板源 fixture 为默认被 AE elide，无 slot；AE 套默认线性 ramp）；动画色标；gradient **stroke**（G-Stroke，同 GCst 路径，直接接力）。
- **跨版本关键发现**：唯一带色标的 fixture 是 AE 25.6-saved（AE 2020 拒开整个项目文件），但 **from-scratch AE25-shaped 渐变体 AE 2020 仍接受**——渐变格式 version-portable，一份 AE25 模板服务双版本 gate。详 `incident-reports/gradient-fill-write-re.md`。RE/模板源 `v2_2_gradient_src.aep`（← py-aep gradient.aep）。
#### V2.2.1 子项⑬ (2026-05-31) — Stroke Dashes (single Dash+Gap pair, hidden-until-enabled)
- `(s *StrokeNode) Dashes()` → `StrokeDashes`（`Enable/Disable`、`SetDash`/`SetGap` 自动 enable + 拒负、`Enabled/Dash/Gap` getter）— ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_StrokeDashes_AEShipGate_AE20{20,25}`：一层 rect+stroke，Dash=18/Gap=7，re-save cdat 解码校验）。
- **磁盘编码**：Dash 1 / Gap 1 均 OneD float64-BE @ cdat[0:8]，嵌套于 `ADBE Vector Stroke Dashes` group `LIST(tdgp)` 内（同 Taper/Wave，`findGroupBody` 下钻 + `overwriteShapeStreamCdat`）。默认 Dash/Gap=10。
- **enable 语义 = 模板切换**：solid stroke 的 Dashes 组是空 placeholder（无 Dash/Gap leaf）。enable 时 `lowerStrokeNode` 切到第二嵌入模板 `v2_2_shape_stroke_dashed_body.bin`（携 Dash 1/Gap 1 slot）；solid `v2_2_shape_stroke_body.bin` **不变** → 现有 stroke/enum/taper-wave gate 零回归（`DisabledStaysSolid` round-trip 证 solid 路径 byte 一致）。hydrate 以 Dash/Gap leaf 存在与否回判 enabled。
- **deferred**：Dash 2/3 + Gap 2/3（AE 只 emit enabled pair，每对需独立模板变体）、**Offset**（AE 端 hidden-until-enabled 且 `setValue` 抛 hidden-property，script-ungettable，无 slot 可建模——RE `v2_2_stroke_dashed.done` 实证）。static-only（不建模 keyframe）。
- RE：`re_stroke_dtw.jsx` + `v2_2_stroke_dashed.aep`（dashed 模板源），详 `incident-reports/stroke-line-cap-join-miter-re.md` Dashes addendum。
#### V2.2.1 子项⑫ (2026-05-31) — Stroke Taper + Wave (static, %/Wavelength-mode)
- `(s *StrokeNode) Taper()/Wave()` → `StrokeTaper` / `StrokeWave`，各 getter/setter — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_StrokeTaperWave_AEShipGate_AE20{20,25}`：一层 rect+stroke，Taper 6 + Wave 3 标量设非默认，re-save cdat 解码校验全 9 值）。
- **Taper**（6 字段）：Start/End Length、Start/End Width、Start/End Ease，默认全 0。**Wave**（3 字段）：Amount(默认 0)、Wavelength(默认 100)、Phase(默认 0)。全 OneD float64-BE @ cdat[0:8]，嵌套于 group `LIST(tdgp)` 内（比顶层 stroke 标量深一层；`findGroupBody` 下钻 + 复用 `overwriteShapeStreamCdat`）。
- **模板**：`gen_shape_all_full.jsx` 扩展设 Taper+Wave（Units 留 % → 9 active slot emit）→ 仅 stroke body 变更（rect/ellipse/fill/path body md5 不变，零回归）。
- **deferred（elision/coupling 陷阱，见 incident-report）**：Taper Length Units + StartWidthPx/EndWidthPx（% 模式被 elide）、Wave Units + Cycles（Wavelength 模式被 elide / Cycles hidden）。**Stroke Dashes** 已 ship 见子项⑬。
- RE：`re_stroke_dtw.jsx`（probe+set 枚举三组子属性），详 `incident-reports/stroke-line-cap-join-miter-re.md` Taper/Wave addendum。static-only（不建模 keyframe）。
#### V2.2.1 子项⑪ (2026-05-31) — Shape enum sweep: Direction / Blend Mode / Composite Order / Fill Rule (static)
- Rect/Ellipse `Direction`、Fill `BlendMode`/`CompositeOrder`/`FillRule`、Stroke `BlendMode`/`CompositeOrder` getter/setter — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_ShapeEnums_AEShipGate_AE20{20,25}`：一层 rect+ellipse+fill+stroke 全设 enum，re-save cdat 校验 Direction=3/FillRule=2/BlendMode=3/CompositeOrder=2）。
- **类型**：`ShapeDirection`（Normal 1/Reversed 3）、`ShapeBlendMode`（AE 1-based index，Normal=1，不枚举全表）、`ShapeCompositeOrder`（AbovePrevious 1/BelowPrevious 2）、`FillRule`（NonzeroWinding 1/EvenOdd 2）。全 OneD float64-BE @ cdat[0:8]，默认皆 1（RE 同 `re_shape_enums.jsx`，详 incident-report）。
- **模板**：rect/ellipse/fill/stroke body 全部从单一 `v2_2_shape_all_full.aep`(`gen_shape_all_full.jsx`，每 prop 设非默认)重抽 → 4 模板含 enum slot（path body 不变）。**重跑全部 shape ship-gate 双版本**（Ellipse/EllKf/FillKf/FillOpKf/RectKf/RectSubKf/Stroke/StrokeKf/ShapeEnums × AE2020+2025）均 PASS，模板 swap 零回归。
- RE gotcha：addProperty reindex 使旧 handle 失效（须按 matchName 重取）；详 incident-report。
#### V2.2.1 子项⑩ (2026-05-31) — Stroke Line Cap / Line Join / Miter Limit (static)
- `(s *StrokeNode) LineCap/LineJoin/MiterLimit` getters + `SetLineCap/SetLineJoin/SetMiterLimit` — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_Stroke_AEShipGate_AE20{20,25}` 扩展：Cap=Projecting(3)/Join=Round(2)/Miter=12，re-save cdat 解码校验）。
- **类型**：`StrokeLineCap`（Butt 1/Round 2/Projecting 3，默认 Butt）、`StrokeLineJoin`（Miter 1/Round 2/Bevel 3，默认 Miter）enum；`MiterLimit` float64（默认 4，setter 拒 <1）。
- **磁盘编码**（RE 自 `re_stroke_linecap.jsx`，详 `incident-reports/stroke-line-cap-join-miter-re.md`）：三者均 1D OneD，float64-BE @ cdat[0:8]，enum 存 1-based index。matchName 即文档 `ADBE Vector Stroke Line {Cap,Join} / Miter Limit`（旧 "property-not-found" groundwork 是错的）。
- **富化 stroke body 模板**（21 children；源 `v2_2_shape_stroke_full.aep`，Cap/Join/Miter 全设非默认 → AE 不 elide）。仅 stroke body 变更（其它 4 shape body 字节不变）。
- **AE 行为**：Cap+Join+Miter 由 AE 绑定一起写；Miter 在 Join≠Miter 时 *live setValue* 被 reset，但 open+resave 保留我们写的值（双版本 gate 均读回 Miter=12）。static-only（不建模 keyframe）。
#### V2.2.1 子项⑨ (2026-05-30) — Fill Opacity 持久化（static + keyframe）
- **Fill Opacity** static + keyframe 持久化 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_FillOpKf_*`；AE 读回 opacity 40 + re-save numKf/bpk；FillKf 重跑双版本仍 PASS）。
- **磁盘编码**（RE 自 `v2_2_fill_kf_re.aep`）：1D non-spatial（bpk-48，value@0x08，**原始 %**，同 Stroke Opacity）。之前 lowerFillNode 完全丢弃 opacity（连 static 都没写）。
- 富化 fill body 模板（7 children；源 `v2_2_shape_fill_full.aep`，Fill Opacity 设静态 60）。仅 fill body 变更。`lowerFillNode` 现持久化 Color + Opacity。
- **仍 deferred（shape-node 次要）**: Stroke Dashes Dash 2/3·Gap 2/3·Offset（单对已 ship 见子项⑬）。〔Stroke Dashes 见子项⑬；Taper/Wave 见子项⑫；Line Cap/Join/Miter 见子项⑩；Rect/Ellipse Direction + Fill/Stroke BlendMode·CompositeOrder + Fill Rule 见子项⑪〕
- **次要子属性**（多数 runtime-only）: Layr Transform Anchor/Scale/Rotation/Opacity。
  - Fill Color 编码: cdat scalar 跟 JSX 0-1 input 不对齐（tolerance 0.5 → 0x406fe0... ≈ 255），可见色可能错
  - Layr Transform 的 Anchor / Scale / Rotation / Opacity keyframe: runtime-only 不持久化（Position keyframe 已 ship，见子项⑤）
  - Rect/Ellipse Direction、Rect Position/Roundness: runtime-only 不持久化

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
