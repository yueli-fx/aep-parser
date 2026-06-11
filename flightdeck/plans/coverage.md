---
status: active
implements: archive/specs/2026-05-26-py-aep-parity-design.md
summary: 字段覆盖概览（精简入口）
---

# Coverage 概览（精简入口）

> 要查具体 AE attribute → Go 字段对应，看 [coverage-detail.md](coverage-detail.md) 详细交叉表。
> 本文档只列：什么已 ship / 什么暂搁 / 什么不可达 / negative findings。

兼容声明：AE 2020 读下限 + AE 24+ 字段渐进写。所有 setter 都是 length-preserving（除少数 length-variable 替换：name / comment / expression / 字体名 / 文本内容）。

测试基线 / PASS count 见 `../cockpit.md` `Last updated` 行（**唯一权威**）。

---

## ✅ 已 ship

### V2 结构性创建 (2026-05-22)

- `aep.NewProject(target ...AETarget) *Project` — 全新空 project，零参 = TargetAE2020；支持 TargetAE2020 / 2022 / 2025
- `proj.NewComposition(name, w, h, fps, duration) (*Composition, error)` — 在 root folder 新建空 comp，原子（warning/error → rollback），跟 `Open(...)` 出来的 comp 同构（所有 Set\* 立即可用）
- AE 2020 + AE 2025 ship gate PASS（`AE_SHIP_GATE=1 go test -run TestV2_1_AEShipGate -v`）
- 详 `../incidents/ae25-acceptance-gate.md` 5 阶段 RE findings + canonical seed 策略

### Project / Composition / Item

- Project: `BitsPerChannel` R/W
- Composition（cdta）: `Name / FrameRate / Duration / Size / BGColor / ShutterAngle / ShutterPhase / MotionBlurAdaptive / MotionBlurSamplesPerFrame / WorkArea / DisplayStartTime / DisplayStartFrame / PixelAspect / ResolutionFactor` R/W
- Composition（PRin LIST）: `Renderer` **R/W**（binary match_name；`SetRenderer` 接受 binary 或 ExtendScript 名；prin 改名 length-preserving + prda 换模板 structural；ship-gate AE 2025 4/4 + AE 2020 Ernst/Escher 绿）
- Composition（cdta flag bits）: `HideShyLayers / CompMotionBlur / Draft3D / FrameBlending / PreserveNestedFrameRate / PreserveNestedResolution` R/W
- Item: `Name / Comment / Label` R/W（Composition + Footage 共用）
- Footage: `Path` R/W
- Composition markers: 全 8 个 setter
- **Composition marker 增删 (P3 §3G, stable)**: `Marker.Remove()` + `Composition.AddMarker(seconds) (*Marker, error)`。结构性 splice：ldat 16B block + lhd3 count + mrky Nmrd 三处联动；AddMarker 走 clone-template（拷现存 marker 的 ldat block + NmHd，规避 opaque 默认值 RE）。AE 2020+2025 双版本 ship-gate 2/2 PASS（`TestMarker_AEShipGate_*`，Remove m0 + Add@4.0）。限制：AddMarker 需 comp 已有 ≥1 marker（空 comp seed 暂搁）；tail-insert 未做 time 排序（add@4.0 在末尾，AE 无需 resort）。Layer marker 增删共享 `markerList` infra 但未单独 ship。fixture `test_data/re_compmarker.aep`
- **Composition guides (P3, py-aep parity, Alpha)**: `Composition.Guides []*Guide` R + `Guide.SetPosition / SetOrientation` W（length-preserving in-place ldat patch，block 别名 chunk bytes）。`GuideOrientation` 枚举（binary 2=horizontal / 1=vertical，py-aep logical 0/1 经 `.String()` 桥接）；JSON 导出 `guides[]{orientation,position}`。无 AE scripting 等价物（标尺参考线 UI-only，不影响渲染）。结构性增删 guide 暂搁。AE 接受未在 app 内验，标 Alpha。fixture `test_data/guides.aep`（py-aep 样本）
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
- **Essential Graphics R (P3 §2.8, py-aep parity)**: `Composition.MotionGraphicsTemplateName`（默认 "Untitled"）+ `EssentialGraphicsControllers []*EssentialGraphicsController{Name, Type, UUID}` + 便捷 `MotionGraphicsTemplateControllerCount() / ...Names()`。`EGControllerType` 枚举（1=Checkbox/2=Slider/4=Color/5=Point/6=Text/8=Comment/9=MultiDim/10=Group/13=Dropdown）。从 Item-level `LIST:CIF3 → CpS2(模板名) + CCtl×N(controller: CpS2 名 + Utf8 uuid + CTyp 类型)` 解。JSON 导出 `motion_graphics_template_name` + `essential_graphics[]`。R-only（W 是 .mogrt binding，结构性，deferred）。10 个 fixture 验证（`eg_*.aep`，py-aep 样本）。
- **3D Orientation R fidelity (P3, py-aep parity)**: 3D 层 `ADBE Orientation` 存在 `otst` wrapper（tdbs + otky）。`parse_properties.go::parseOrientationProperty` 修正：Components=3（tdb4 dim byte 误报 1）、静态 cdat 按**小端**解（otst 内 cdat 是 LE，`decodeCdatValueLE`）、keyframe 的 X/Y/Z 从 **otda**（大端）取。fixtures `orientation_5_0_0/0_279_0/with_keyframes`。⚠ animated orientation 的 easing/tangents 仍按旧 1D layout 未校验；tree 里 otst 仍显示为空 group（by design）。详 `incidents/transform-group-default-omission.md`。
- **Property dimensions separation R (P3, py-aep parity)**: `DimensionsSeparated()`（tdsb byte 3 `_enable_flags` bit 1）+ `IsSeparationLeader() / IsSeparationFollower() / SeparationDimension()`（纯 match-name：leader=="ADBE Position"，follower=="ADBE Position_0/1/2"→dim 0/1/2，非 follower 返回 -1）。fixture `transform_separated.aep`。
- **Property dimensions separation W (P3 §3C, stable)**: `Property.SetDimensionsSeparated(bool)` — 结构性 separate↔merge 双向 toggle。**static Position（2D + 3D）双版本 ship-gate 6/6 PASS** + **animated Position（3D，leader 近线性 path-ease）双版本 ship-gate 8/8 PASS**（AE 2020 + AE 2025 × separate/merge × uniform/非均匀），合计核心 R/W 已双版本验收 → stable。Go 输出 chunk 树 byte-structural 等同 AE 自存。字节机制：static separate = leader 翻 tdsb byte2→0x08+bit1、值重置默认 `[w/2,h/2,0]`、真值迁入 Position_0/1（3D 再合成 `Position_2`，靠 `layer.Is3D` 区分非 Components）；merge = leader 取回 `[X,Y,Z]`、删全部 follower（AE resave 会重新预分配 zeroed Pos0/1 → merged 两种等价表示）。animated 走 keyframe 流迁移（`separatePositionAnimated` / `mergePositionAnimated`，leader 3D spatial bpk=128 ↔ follower 1D temporal bpk=48，speed=值中心差分×(100/6)、influence 承载时距）。fallible 的 Position_2 re-parse/合成在所有就地改动前 → 失败即无副作用。**refuse 子集**：animated 2D 层、leader 自定义 spatial-path temporal ease（首切片仅 3D + 近线性）。详 `incidents/separate-dimensions-write-mechanics.md`。注：3D 层 Transform Group 解出的属性比 py-aep 少不是 bug —— AE 省略未改的默认属性（py-aep 合成 schema，我们不合成 `Elided`），见 `incidents/transform-group-default-omission.md`。
- **PropertyBase Remove / MoveTo W (P3 §3C, Alpha)**: `AEPropertyGroup.Remove()` / `MoveTo(index)` — 结构性删/重排 INDEXED_GROUP 直接子节点；`IsIndexedGroup()` 谓词 = 固定 match-name 集（Effect Parade / Mask Parade / Effect Mask Parade / Root Vectors Group / Text Animators，抄 py-aep `_INDEXED_GROUP_MATCH_NAMES`；其余皆 NAMED_GROUP，子节点固定不可删/移，调用即 error，镜像 AE ScriptingAPI refuse）。**Effect Parade 双版本 ship-gate 4/4 PASS**（AE 2020 + AE 2025 × remove-middle / move-last-to-front，AE resave 回读 + Go 二次解析确认保留）；Mask Parade / Root Vectors / Text Animators 同机制 + Go round-trip 过但未单独 gate → Alpha。字节机制：纯父 group LIST 内 **tdmn+payload pair splice / reorder**，**无 count/index chunk 联动**（`rebuildIndexedGroupChunk` 按 mutate 后 scene `Children` 序重发 parent.chunk.Children，保留 prefix tdsb/tdsn + suffix Group End，每子复用**原 chunk 指针对** → opaque 原样带走，CLAUDE.md #5）。V2.1 atomic（snapshot chunk LIST + scene children + Effect/Mask flat 镜像 + Project.Warnings；新增 parser warning 即全回滚并 error）。**Duplicate deferred**（clone display-name " 2" 后缀存盘 vs runtime 派生未 byte-check）。详 `incidents/property-indexed-group-structural-re.md`。
- **AddEffect W (2026-06-10, Stable——2026-06-11 全库 gate 后升级，含 RemoveEffect)**: `aep.AddEffect(layer, effectMatchName) (*Effect, error)` + `aep.SupportedEffects() []string` — 给 layer 的 `ADBE Effect Parade` 追加一个内置效果，返回解析好的 `*Effect`（可即时调参——`Property.SetStaticValue` 对 effect 参数生效，已 ship）。**AE 2020 + AE 2025 双版本 ship-gate 24/24 PASS**（`TestAddEffect_AEShipGate_AE20{20,25}` 子测，全 12 效果 × 2 版本，payload 1.7KB→20.6KB 全覆盖；AE 接受不损坏、回读 4 effect 顺序正确、AE resave 保留。2026-06-10 先按 payload 体量取样 5 效果 10/10，2026-06-11 补齐余 7 个）。字节机制：parade `LIST(tdgp)` 内 effect = `(tdmn[40B], LIST:sspc)` pair，以 `ADBE Group End` tdmn 哨兵收尾；AddEffect 在哨兵前 splice 一对（**同 `DuplicatePropertyGroup` 的 pair-splice 机制，但 sspc payload 来自嵌入 AE-native 模板**而非克隆兄弟）。LIST size 由 `rifx.Chunk.Write` bottom-up 自动重算（同 gradient/SetPath length-variable 路径）。原子（snapshot parade chunk + scene children + flat Effects + warnings → 任何 parser warning 即回滚）。
  - **效果库 = 30 内置效果**（match-name → 嵌入模板 `internal/serializer/templates/effect_adbe_*.bin`）。Wave 1（12，fixture `re_effect_library.aep`）：Gaussian Blur 2 / Fill / Tint / Brightness & Contrast 2 / Tritone / Easy Levels2 / Pro Levels2 / HUE SATURATION / Box Blur / Glo2 / Invert / Exposure2。Wave 2（17，2026-06-11，fixture `re_effect_library2.aep`，提取器 `tmp_debug/extract_effect_lib` 可复用）：Drop Shadow / Sharpen / Mosaic / Noise / Transform(Geometry2) / Gradient Ramp / Fractal Noise / Motion Tile / Directional Blur / Linear Wipe / Wave Warp / Curves(CurvesCustom) / Slider·Point·Color·Angle·Checkbox Control（表达式控制；Layer Control 仍排除——layer 引用参数）。Wave 3（1，2026-06-12，fixture `re_effect_param_types.aep`）：Point3D Control。版本可移植——AE-2020 字节 AE 2025 接受（同 gradient 发现）。**30 个全部 Go round-trip 验证**（`TestAddEffect_AllTemplates_RoundTrip` 迭代 `SupportedEffects()` 自动覆盖）+ **AE 双版本 ship-gate**（wave 1 24/24 + wave 2 34/34 2026-06-11；Point3D Control 经 SetEffectParam gate 2026-06-12 双版本 add+读回覆盖）。
  - **配套 API**：`aep.RemoveEffect(layer, index)`（AddEffect 的逆，delegate 到 ship-gate-green `RemovePropertyGroup`，index 校验，Go round-trip 验证）+ 30 个 `aep.Effect*` typed match-name 常量（`EffectGaussianBlur`…`EffectPoint3DControl`，registry 单源 + `TestEffectConstants_MatchRegistry` 防漂移）。
  - **`aep.SetEffectParam(layer, fx, paramMatchName, value)`（2026-06-11，Alpha，synthesis-lite；2026-06-12 控件类型补齐）**：设效果参数静态值；目标参数被 default-elide（默认实例无 tdbs 流可写）时先物化 `(tdmn, tdbs)` 再写值——镜像 AE 自身「值≠默认才持久化」语义，无需动已 gate 的 add 模板。**泛型 per-control-type 模板全类型覆盖**：值流 tdbs 形态按 control type 而非参数——任意效果的 **scalar/enum/boolean/angle/color/2D point/3D point/slider 参数即设即用**（从宿主 sspc 自带的 pard 定义 patch tdmn/tdsn/min-max；parT 永不 elide；模板源 = GB fixture + `re_effect_param_types.aep` 五个表达式控制）。**AE 2020 + AE 2025 双版本 ship-gate PASS**（all-Go-built 文件 13 项：GB 三参数 per-param + Drop Shadow 五参数泛型物化（含 cross-effect color/angle patch）+ 五个表达式控制 per-param，input+resave 重开读回全对）。**值编码 = on-disk StaticValue 单位**：color = [A,R,G,B]×255；point = 层坐标空间分数（有源层 = 源尺寸、source-less = comp 尺寸，z 除以 height——RE `re_effect_param_types_units.aep`）。⚠ 跨版本默认值漂移（AE 2025 把 GB Repeat Edge Pixels 默认改 true）→ resave 可合法重新 elide，断言看重开读回值不看流存在性。详 `incidents/effect-param-elision-synthesis-lite.md`。
  - **Parade auto-create（2026-06-10 Phase 2 落地）**：无效果的 parsed 层（AE 只在 ≥1 effect 时才存 parade，所以**所有无效果层都没有 parade**——不止 from-scratch 层）AddEffect 自动 splice 空 parade（tdsb 0x01 + tdsn "-_0_/-" + Group End，AE-native 形态）进 Layr property tree，锚点 = `ADBE Transform Group` tdmn 之前（AE 实际 emit 顺序，RE re_shape_effect/re_effect_library）。**AE 2020 + AE 2025 双版本 ship-gate 2/2 PASS**（`TestAddEffectAutoParade_AEShipGate_AE20{20,25}`：100% Go-built 文件 NewProject→NewShapeLayer→Reopen→AddEffect 全链路）。**tdpi host-layer remap fix**：效果模板 sspc 内每个参数 tdbs 带 4B `tdpi` = 提取 fixture 的宿主层 ID，AE 打开时校验，悬空即拒（「无法在合成中找到图层 ID=N」）——Phase-1 gate 当初能过是 baseline 宿主层 ID 碰巧 ==15；现 clone 后全部 remap 到目标层 ID（`retargetEffectHostLayer`，白盒回归 `TestAddEffect_RetargetsTdpiHostLayer`）。refuse：camera/light 层（AE 不允许效果）+ 未 parse 的 from-scratch 层（无 property tree；报错指引 `aep.Reopen`）。
  - **`aep.Reopen(p)`（2026-06-10，「fresh-layer setter 墙」统一方案 a）**：WriteAEP 到内存 + FromReader 往返，把结构性 New* 建出的 built 层升级为 parsed 层（property tree + chunk back-refs 齐全），一行解锁全部 parsed-layer-only 写路径（AddEffect auto-create / Camera*/Light* setter…）。byte-stable（`TestReopen_WriteStable`：reopen 后再写与原写 byte-identical）。
  - **仍 deferred**：fresh 层不经 Reopen 直接加效果（write-time re-lowering 墙 + scene 禁持 rifx chunk，需 pending-effects 表——有 Reopen 后价值低）；含 layer/path-reference 参数的效果（sspc 内 tdpi 指向非宿主层，需完整 id remap）+ 30 之外继续扩库 + per-effect typed param helper（今为 raw `SetStaticValue` by match-name）。
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
  - `LinearizeWorkingSpace` R/W — lnrp 同 lnrb pattern；**chunk 写法 AE-byte-identical** 但 ScriptingAPI 读 false (OCIO/CMS联动 quirk，AE 自己 set 后 reload 也读 false，详 `incidents/project-flag-chunks-lnrb-lnrp.md`)
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

### Layer 创建（footage-backed：Solid / Null / Adjustment，2026-06-10）

- **`aep.NewSolidLayer(comp, name, w, h, rgb)` / `aep.NewNullLayer(comp, name)` / `aep.NewAdjustmentLayer(comp, name)`（Stable——2026-06-11 审计升级）** → `*Layer` — footage-backed 图层创建。**AE 2020 + AE 2025 双版本 ship-gate PASS**（`TestNewSolidNull_AEShipGate_AE20{20,25}`：all-Go-built 工程 + solid/null/adjustment 三层 → AE 接受、颜色/尺寸/flag 读回正确、resave 保留、Go re-parse 确认；无 ID 巧合——模板 footage ID 14/16/18 vs dest 分配 13/15/17 错位）。机制：**embed 整个模板工程**（`templates/solidnull_2020.aep`，AE 2020 fixture）+ **复用 cross-Project `insertLayerCrossProject`**（footage 闭包导入 fresh item ID + SourceID remap），后置 patch 全部 length-preserving：layer name（managed splice）/ time span re-home 0→comp duration / opti name @0x1A（256B 固定缓冲）/ opti 颜色 ARGB 4×f32 @0x0A / sspc W/H u16 @0x20/0x24 / ldta pad 到 capability。返回的层**已是 parsed 层**（InsertLayer 内 re-parse）——transform/AddEffect 等 setter 即刻可用，无需 Reopen。
- **`Footage.SolidColor [3]float64` R**（opti "Soli" ARGB @0x0A，alpha 恒 1 不暴露）+ `MainSource()` 的 `SolidSource.Color` 现已填充。
- **ID-namespace scar 扩展**（`nextitemid-must-include-layer-ids` 续）：dummy-comp 模板的 service 层（DLay/SLay/CLay/SecL，ID 2..12）不进 `c.Layers`，旧 `initDerived`/`NewComposition` 看不见 → Go-built 工程里 `allocItemID` 发出 2/3 与 service 层撞号 → **AE 2025 拒收**（`unexpected match name searched for in group`；AE 2020 宽容）。修复：`initDerived` + `NewComposition` 都扫 itemList 全部 layer-list chunk ID（`maxLayerIDInItemList`）。

### Layer 创建（source-less，2026-06-10）

- **`aep.NewCameraLayer(comp, name)` / `aep.NewLightLayer(comp, name)`（Stable——2026-06-11 审计升级）** → `*Layer` — 源-less 图层创建（Camera / Light）。**AE 2020 + AE 2025 双版本 ship-gate PASS**（`TestNewCameraLight_AEShipGate_AE20{20,25}`：fresh comp + camera + light → AE 接受、类型正确〔`Cam1(cam)`/`Light1(light)`〕、resave 保留、Go re-parse 确认 `LayerTypeCamera`/`LayerTypeLight`）。机制：**embed-whole-Layr**——克隆从 `re_cameralight.aep` 提取的 AE-native Camera/Light Layr（`templates/layer_{camera,light}_body.bin`，4-child：ldta + Utf8 + tdgp〔Transform + Camera/Light Options〕+ Gide），只 patch 每实例字段（ldta ID @0x00 / time span / ParentID @0x84=0 / size pad 到 capability / Utf8 name），splice 复用 NewShapeLayer 机制（Ewst + fvdv followers + atomic rollback）。AE-internal flag 字节全保真 → 规避 shape 层那些 from-scratch silent-drop 坑。版本可移植（AE-2020-form 模板双版本接受）。**setter 暂搁**：fresh 层无 scene tree，Camera*/Light* 值继承模板默认，需 write+reopen 才能调（同 effect Phase 2 的 scene-vs-chunk 墙）。详 `incidents/camera-light-layer-create-re.md`。
- **`aep.NewTextLayer(comp, name)`（2026-06-11, Stable——同日审计升级）** → `*Layer` — 文本图层创建（新建图层类型最后一块，**全部图层类型可建齐**）。**AE 2020 + AE 2025 双版本 ship-gate 2/2 PASS**（`TestNewTextLayer_AEShipGate_AE20{20,25}`：fresh comp + 两个文本层 T1="A"/T2=SetText("B") → AE 接受、`instanceof TextLayer`、`sourceText.value.text` 读回正确、resave 保留、Go re-parse 确认）。机制：同 camera/light 的 **embed-whole-Layr**（模板 = `re_text.aep` 层 baseline_A：point text "A"，YouYuan 88px 单 run；`templates/layer_text_body.bin`，14.5KB，btdk 文本 blob 原样嵌入——审计 tdpi=0/sourceID=0）。**增量**：创建时即接 btds back-ref + 解码 TextSource → fresh 层**无需 Reopen** 就能读 `TextSource` / 做等长 `SetText`（优于 camera/light 的 setter 体验）。T2 gate probe 同时证明 **AE 载入时重算 btdk 布局缓存**（stale "A" 字形度量下正确读回 "B"）——任意长度文本写的解封路径 + 三处长度耦合计数详 `incidents/text-btdk-length-variable-write-scoping.md`。
- **`Layer.SetText` 升级 length-variable（2026-06-11，Stable 核心 R/W 语义扩展、签名不变）**：单段落/单 run/无手动 kerning 文档（NewTextLayer 产物即此形态）支持**任意长度**改字——串 splice + 段落/run 字符计数（UTF-16 units 含尾 `\r`）同步 + 内嵌 btdk size header 修正（AddFont 同机制），布局缓存留给 AE 载入重算。**AE 2020 + AE 2025 双版本 ship-gate 2/2 PASS**（`TestSetTextVariable_AEShipGate_AE20{20,25}`：Go-built 工程，"A"→"Hello AEP parser"（17 units）+ "A"→"你好世界"（CJK 5 units），AE 读回精确、resave 保留）。等长且**段落剖面相等**（逐段 UTF-16 计数一致，不止总数——边界移动会写坏段落计数）走原地快路径，任意结构可写；变长对多段落/多 run/带 kerning 表/空串 refuse（需 entry splicing，未 RE）。随手修复：`DecodePSString` 代理对解码（astral 字符 round-trip，原逐 unit `rune(cu)` 会打成 U+FFFD）。

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
- 路径动画: `PathKeyframes` R only（既有 mask 的顶点重写 = 结构性 ❌；但**创建时路径可参数化**——见 AddMask）
- **AddMask W (2026-06-11, Alpha)**: `aep.AddMask(layer, name, path BezierPath) (*Mask, error)` — 给 layer 的 `ADBE Mask Parade` splice 一个 from-scratch mask atom（`(tdmn, mkif[48B], tdgp)` **三件套**，无嵌入模板，全 Go 构造；parade auto-create 复用 `spliceEmptyParade`，锚点 = Effect Parade 之前否则 Transform Group 之前）。创建时静态路径任意参数化（顶点/切线/open-closed，公共 API 收层像素，内部按源 item 分数 vs source-less 裸像素换算）；返回的 `*Mask` back-refs 即时可用（SetMode/SetInverted/SetColor…）。**AE 2020 + AE 2025 双版本 ship-gate 4/4 PASS**（AE-native fixture + 100% Go-built × closed/open，顶点像素级读回 + resave 保留）。配套修复：mask `Closed` 解析改 shph[3] bit3（旧 @0x14 读法对 AE-native open mask 误读）+ `Mask.Name` 加 atom tdsn fallback（AE 面板名存 tdsn，omtn 恒空）。refuse：camera/light / 未 Reopen 的 fresh 层 / 空 path。deferred：RemoveMask（atom 三件套不满足 RemovePropertyGroup 的 pair 假设）/ 既有 mask 路径改写 / animated path / mode·color 创建参数。详 `incidents/add-mask-create-re.md`。

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

### Render Queue R-only (P3 §3A slice-1, 2026-06-01)

py-aep parity P3 首刀。纯 reader，无写、无 ship-gate（byte-identical round-trip 由 opaque preservation 保证，已测）。RE 全部对照 py-aep `binary/render_chunks.py` + golden JSON 对账（`test_data/rq_*.aep` 来自 py-aep samples）。

- `Project.RenderQueue *RenderQueue`（无 LRdr 时 nil；空队列 non-nil + 0 items）+ `RenderQueue.NumItems() / .Items`
- `RenderQueueItem.{Comp（via comp_id@0x08→Composition）, Status（raw u32@0x0C）, Name（template_name@0x5A win-1252）, Comment（RCom→Utf8）, LogType（raw u16@0x50）, QueueItemNotify（flag@0x07 bit2）, ElapsedSeconds（u32@0x89A）, TimeSpanStart/Duration（按 time_span_source@2148 解析：LENGTH_OF_COMP/WORK_AREA/CUSTOM）, RenderSettings, OutputModules, NumOutputModules()}`
- `RenderSettings`（slice-2，R）：Quality/ColorDepth/Effects/FieldRender/Pulldown/FrameBlending/MotionBlur/ProxyUse/SoloSwitches/GuideLayers/DiskCache/FrameRate/Resolution[2]/SkipExistingFiles — 值用 py-aep NUMBER 语义（0xFFFF→-1 current-settings）；offset 对 4 个判别 fixture 交叉验证
- `OutputModule.{Name（Als2 后 Utf8[0]）, FileTemplate（Als2 后 Utf8[1]）, FullPath（alas JSON fullpath）, Settings}`
- `OutputModuleSettings`（slice-3，R）：128B `OutputModuleSettingsItem`（Channels/ResizeQuality/Resize/LockAspectRatio/Crop+4 边/OutputAudio/IncludeProjectLink/PostRenderAction/ConvertToLinear/4 flag-bit）+ 154B `Roou`（VideoCodec/FormatID/StartingNumber/Width/Height/Depth/VideoOutput/AudioSampleRate/AudioBitDepth/AudioChannels/AudioEnabled）。om-settings ldat 在每个 LItm-item 的 `list` 内；按 OM index 配对；2 个正交 flag fixture 交叉验证
- `OutputModule.FormatOptions *FormatOptions`（slice-4，R）：Ropt 按 format_code 分发的 6 个二进制格式 —— Cineon(sDPX)/Jpeg(JPEG)/OpenExr(oEXR)/Targa(TPIC)/Tiff(TIF )/Png(png!)。XML 格式（AVI/H264/QuickTime 无 Ropt 变体）→ nil。offset 对 cineon 4 个 fixture-name-encoded 值交叉验证（converted_white_point 存归一化 253/255）
- JSON 导出 `render_queue` 节点（snake_case 自有 schema，含 `render_settings` + OM `settings` + `format_options`）
- chunk family：`LRdr → {list→lhd3+ldat(RenderSettingsItem 2246B×N), LItm→per-item [RCom]+list(ldat=OMSettings 128B×M)+'LOm '(Roou 组+Ropt)}`；新 rifx ID `LRdr/LItm/'LOm '/RCom/Roou/Ropt/Rout`
- **RQ reader R-only 主体完成**（slice-1 结构 + 2 render settings + 3 OM settings + 4 format options）
- `RenderQueueItem` 渲染设置 **W**（slice-5，**Alpha** — 未 ship-gate）：`SetQuality/SetColorDepth/SetEffects/SetFieldRender/SetPulldown/SetFrameBlending/SetMotionBlur/SetProxyUse/SetSoloSwitches/SetGuideLayers/SetDiskCache/SetFrameRate/SetResolution/SetSkipExistingFiles`。length-preserving in-place ldat patch（settingsBlock 别名 chunk bytes，sentinel -1↔0xFFFF），非结构性故不走 ship-gate；round-trip 字节验证。AE 接受未在 app 内验，标 Alpha
- `OutputModule` 设置 **W**（slice-6，**Alpha**）：128B 块 `SetChannels/SetResizeQuality/SetResize/SetLockAspectRatio/SetCrop+4 边/SetIncludeProjectLink/SetPostRenderAction/SetUseCompFrameNumber/SetUseRegionOfInterest/SetIncludeSourceXMP/SetPreserveRGB` + Roou `SetDepth/SetStartingNumber`。flag-bit 走 read-modify-write；round-trip 验证
- `RenderQueueItem` item 级 **W**（slice-7/8，**Alpha**）：`SetName`（64B 定长 win-1252）`SetQueueItemNotify`（flag bit）`SetLogType` + `SetTimeSpanStart/SetTimeSpanDuration`（切 CUSTOM source + 分数化 1e6+gcd，decimal 值精确 round-trip）
- `RenderQueueItem.SetComment` **W**（length-variable，**双版本 ship-gate PASS**）：RCom 是 wrapper leaf 内嵌单 Utf8（`"Utf8"+u32 len+UTF-8 payload`，奇数 payload 补偶 pad）。已有 RCom → 原地换 payload；无则在 item 的 `list` 前插入新 RCom。空串 + 无 RCom = no-op。**`RenderQueueItem.comment` 无任何版本 ScriptingAPI**（纯二进制字段=面板 Comment 列），故 gate 走"AE 打开接受 + resave byte 层保留"而非 readback；AE 2020 + AE 2025 均接受且保留。**RQ value 读+写主体完成**
- `RenderQueue.RemoveItem(index)` / `AddItem(comp)` **W**（结构性，**双版本 ship-gate PASS**）：
  - **Remove**：删 item 联动四处 —— LItm 移除 `[RCom?]+list+'LOm '`、settings ldat splice 2246B 块 + lhd3 count -1、Rout per-item block 删 + header 比例减。Go 输出 LRdr 子树 byte-structural 等同 AE `item.remove()`。
  - **Add**：clone 末位 item + remap comp_id（@0x08）—— append settings block / `[list,'LOm ']` / Rout block + lhd3/header 增。空队列无模板 → refuse。
  - settings ldat 跨 item 共享 → 增删后重挂 item 的 settingsBlock 别名。AE 2020+2025 读回 numItems±1 + comp 链接正确。**RQ 结构性增删收尾完成**。详 `incidents/render-queue-delete-mechanics.md`
- **defer**：XML format options（AVI/H264 从 Als2 旁 JSON，独立解析路径）/ Cineon·Png HDR10 metadata / Format·OutputAudio·Color 派生枚举映射 / SkipFrames（派生）/ `file` 模板变量解析 / LogType·Status·PostRenderAction 的 3xxx 命名空间枚举 / 任何写。详 `plans/2026-06-01-py-aep-p3-renderqueue-reader-plan.md`
- gotcha：comp duration 两套表示分歧 → [`incidents/cdta-duration-two-representations.md`](../incidents/cdta-duration-two-representations.md)

### V2.2 alpha ShapeLayer 写路径 (2026-05-25, iter-7/8)

- `(c *Composition) NewShapeLayer(name string) (*ShapeLayer, error)` — 新建空 ShapeLayer，AE 2025 接受 + comp.layers.length=1
- `(s *ShapeLayer) RootGroup() *VectorGroup` — 取顶层 Contents 容器
- `(g *VectorGroup) AddRect() (*RectNode, error)` / `AddFill() (*FillNode, error)` — 加入参数化 shape kid
- Setter: `RectNode.SetSize([w, h]) / FillNode.SetColor([r,g,b,a])` static-only
- **AE 接受 gate**: 通过 embed tolerance.aep 抽出的 boilerplate 字节 (Transform Group 1842B + Rect 448B + Fill 426B + **Ellipse 730B** in `internal/aep/templates/`)；详 `../incidents/v2-2-aelayer-structure.md`

#### V2.2.1 (2026-05-29) — Ellipse embed bytes + AE 2020 ldta 地基修复
- `(g *VectorGroup) AddEllipse() (*EllipseNode, error)` + `SetSize / SetPosition` — ✅ **AE 2020+2025 双版本 ship-gate PASS**。embed `v2_2_shape_ellipse_body.bin` + overwrite Ellipse Size/Position cdat（offset 0, f64 BE）。`TestV2_2_Ellipse_AEShipGate_AE20{20,25}`（assert-based：AE 接受层不 silent-drop + re-save cdat 保留值）。
- **AE 2020 地基 bug 修复**：`buildLdtaBytes` 此前硬编码 164B ldta，AE 2020 判**所有** shape 图层（含 Rect+Fill）损坏并跳过；从未发现因 AE-2020 shape gate 长期 skip。改为按 target 分支（capability matrix `LdtaSize`：160 AE2020/22 / 164 AE25）。`TestLowerShapeLayer_LdtaSizeByTarget`。**Rect+Fill 在 AE 2020 现亦有效**（同 ldta 路径，Ellipse gate 已证该路径）。详 `../incidents/ae2020-shape-ldta-164-corrupt.md`。
#### V2.2.1 子项② (2026-05-29) — Path embed+splice + ldat 编码 RE 修复
- `(g *VectorGroup) AddPath() (*PathNode, error)` + `SetVertices / SetClosed` — ✅ **AE 2020+2025 双版本 ship-gate PASS**。embed `v2_2_shape_path_body.bin`（AE-native scaffolding）+ splice `encodeBezier` 几何。`TestV2_2_Path_AEShipGate_AE20{20,25}`（distinct 三角形，re-save + 解析器反归一化解 anchor 验证）。
- **ldat 顶点编码 bug 修复**：`encodeBezier` 曾存 `[anchor, in_i, out_i]`（本顶点 in/out），AE 实为 `[anchor, anchor+out_i, anchor_{i+1}+in_{i+1}]`（本顶点 out 控制点 + 下一顶点 in 控制点，wrap mod n，bbox 归一化）。`TestEncodeBezier_LdatMatchesAELayout`，验证与 AE-native 字节一致。详 RE：`../archive/specs/2026-05-29-path-embed-re-findings.md`。
- **from-scratch path 崩溃 AE 2020**（0::42）→ 必须 embed（同 Ellipse 教训，但 path 是变长几何 splice，非 overwrite-in-place）。

#### V2.2.1 子项③ (2026-05-29) — Stroke embed + Fill/Stroke Color 编码 RE 修复
- `(g *VectorGroup) AddStroke() (*StrokeNode, error)` + `SetColor/SetWidth/SetOpacity` + Cap/Join/Miter（子项⑩）+ BlendMode/CompositeOrder（子项⑪）+ `Taper()/Wave()`（子项⑫）— ✅ **AE 2020+2025 双版本 ship-gate PASS**。embed `v2_2_shape_stroke_body.bin`（含 Taper 6 + Wave 3 active slot；Dashes 仍空 placeholder）+ overwrite cdat。`TestV2_2_Stroke_AEShipGate_AE20{20,25}` + `TestV2_2_StrokeTaperWave_AEShipGate_AE20{20,25}`。
- **shape 颜色编码 RE 修复**：AE 存 `[A,R,G,B]×255` f64（非原始 `[r,g,b,a]×1.0`）。`encodeShapeColorBE` 统一 Fill+Stroke。修了长期 deferred 的"Fill Color 编码不准"——`lowerFillNode` 此前写原始 RGBA，可见色错。`TestLowerFillNode_ColorEncodingARGB255` + Ellipse gate re-save Fill 颜色校验。
- 清理：移除 from-scratch 死代码 `nodeBodyTdgp` / `emptySubPropPlaceholder`（5 个 shape kind 全 embed）。
- **V2.2.1 全部 5 shape kind（Rect/Ellipse/Path/Fill/Stroke）+ 地基 ldta + 颜色编码均 ship**。

#### V2.2.1 子项④ (2026-05-29) — shape keyframe 持久化（Size + Color + Ellipse Position）
- **Rect/Ellipse Size**（non-spatial Vec2）+ **Fill/Stroke Color**（spatial-style dim4 ARGB×255）+ **Ellipse Position**（spatial motion-path Vec2 bpk 104）keyframe 持久化 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_RectKf_*` + `TestV2_2_FillKf_*` + `TestV2_2_EllKf_*`；re-save 解 numKf+值）。Ellipse Position 仅 AE 自动 ~0 spatial 切线与原生不同（AE recompute），值往返正确；`valueLayout.motionPath` 在 0x08 写标志。
- 机制 `injectAnimatedStream`：static tdbs 的 cdat ↔ animated `LIST(list)(lhd3+ldat)`；patch tdb4 标志（@0x05 `&=~1`、@0x44 `=1`、@0x4f `&=~1`）。ldat 与 AE 原生字节一致。`encodeKeyframes` non-spatial（value@0x08 bpk 88）/ spatial（value@0x38）两布局。**坑**：`rifx.IDTdb4` 是大写 legacy，实际小写 `tdb4`。
- **shape path keyframe** 已 ship 见子项⑮（2026-05-31，linear，逐帧 shap + time-table tdbs）。

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
#### V2.2.1 子项⑮ (2026-05-31) — Shape path keyframe write (animated ADBE Vector Shape, linear)
- `(p *PathNode) Path().AddKeyframeLinear(t, BezierPath)` — 逐帧 bezier 形状路径动画 — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_PathKf_AEShipGate_AE20{20,25}`：3 linear 关键帧、各帧顶点数/几何不同；AE 开文件不损坏不 drop、JSX 确认 path numKeys≥2、re-save 后 Go re-decode resaved om-s 确认 ≥3 shap + time-table kfl 存活）。
- **磁盘形态（== 动画 mask path）**：`numKeys≥2` 时 `lowerPathNode` 走 `spliceAnimatedPath`——克隆 embed `v2_2_shape_path_body.bin` 的 om-s，①value tdbs → time-table tdbs（**保留 tdb4** + patch @0x05/@0x44/@0x4f flag，删 cdat，接 `LIST(kfl){lhd3,ldat}` 时间表，每帧 64B block）；②omks 单 shap 克隆成 N shap，各 `spliceShapGeometry`（`encodeBezier` bbox 归一化几何）。`numKeys≤1` 走原 static splice。
- **time-block 64B 布局（byte-match `re_path_anim.aep`）**：time u32@0x00 / inInterp@0x04=1 / outInterp@0x05=1 / const 0x01@0x07 / const u32 2@0x08 / f64 1.0@0x10（除末帧）/ runtime 指针@0x38 置零。详 `incidents/path-keyframe-write-re.md`（含上个会话三处误读的纠错）。
- **三关踩坑（gate 走通前）**：首跑 exit 2 被 cockpit "默认当 flake" 误导（实为字节错触发 "项目文件似乎已损坏（跳过部分）"）→ 改 time-block + tdb4 字节；JSX 导航太浅（path 嵌 3 层）→ 递归 `findByMatch`；测试 `findShipChunk(IDOmS)` 按 ID 找不到 LIST → 新 `findShipListByForm`。
- **deferred**：temporal ease（首版 linear only）、mask path write（只做 shape path）、open path 的 shph closed 语义。
#### V2.2.1 子项⑯ (2026-06-10) — Gradient stroke (AddGradientStroke: color + alpha stops, R/W)
- `(g *VectorGroup) AddGradientStroke()` → `GradientStrokeNode`（`SetColorStops`/`SetAlphaStops`，复用 fill 共享 `setGradientColorStops`/`setGradientAlphaStops` free function）+ reader `hydrateGradientStrokeNode`（G-Stroke 节点局部降级，对齐 G-Fill）— ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_GradientStroke_AEShipGate_AE20{20,25}`：rect+gradient-stroke，3 色标 + 非默认 alpha ramp，re-save 经 read 路径解码校验 color + alpha）。
- **复用**：磁盘编码同 G-Fill（`ADBE Vector Grad Colors` GCst→GCky→Utf8）；lower 经共享 `lowerGradientStops`、reader 经共享 `hydrateGradientStops`（第二个 gradient 节点 = DRY 阈值，fill 三处本体重构复用）。模板从 `v2_2_gradient_src.aep` 的 G-Stroke 节点提取（`v2_2_shape_gradstroke_body.bin`）。
- **deferred**：stroke 几何（width/cap/join/miter/dashes/taper/wave，保持模板原值）+ ramp geometry（同 fill）+ 动画色标。
- 详 `../archive/specs/2026-06-10-gradient-stroke-write.md`（两轮外部 review 定稿）+ `incidents/gradient-fill-write-re.md`（共用 GCst 路径与 elision 教训）。
#### V2.2.1 子项⑭ (2026-05-31) — Gradient fill (SetGradient: color + alpha stops)
- `(g *VectorGroup) AddGradientFill()` → `GradientFillNode`（`SetColorStops`/`SetAlphaStops` ≥2 stop + 范围校验、`Gradient()` getter）— ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_GradientFill_AEShipGate_AE20{20,25}`：一层 rect+gradient-fill，3 色标 red/green/blue，re-save 经 read 路径解码校验）。
- **磁盘编码**：色标存 `ADBE Vector Grad Colors` 的 `LIST(GCst) → LIST(GCky) → Utf8` prop.map XML（version='4'）。色标数组 6 float `[off,mid,r,g,b,1]`，alpha 3 float `[off,mid,a]`，Alpha Stops 在 Color Stops 前 + 各带 Stops Size，尾 `Gradient Colors=1.0`。`EncodeGradientXML` = `ParseGradientXML` 的逆，round-trip 自洽。
- **length-variable 写**：覆 Utf8 XML 改字节长 → **rifx.Chunk.Write 自动 bottom-up 重算所有 LIST size**（同 `Footage.SetPath` 机制），GCst 的 tdb4-124B 非冗余长度头，无需手动 fixup。
- **read 早已 ship**（`ParseGradientXML` + `Property.Gradient`，`TestGradient_FixturePyAep`）；本子项补 write。
- **deferred（elision/coupling + ScriptingAPI 封锁）**：Grad Type/Start Pt/End Pt（模板源 fixture 为默认被 AE elide，无 slot；AE 套默认线性 ramp）；动画色标；gradient **stroke**（G-Stroke，同 GCst 路径，直接接力）。
- **跨版本关键发现**：唯一带色标的 fixture 是 AE 25.6-saved（AE 2020 拒开整个项目文件），但 **from-scratch AE25-shaped 渐变体 AE 2020 仍接受**——渐变格式 version-portable，一份 AE25 模板服务双版本 gate。详 `incidents/gradient-fill-write-re.md`。RE/模板源 `v2_2_gradient_src.aep`（← py-aep gradient.aep）。
#### V2.2.1 子项⑬ (2026-05-31) — Stroke Dashes (single Dash+Gap pair, hidden-until-enabled)
- `(s *StrokeNode) Dashes()` → `StrokeDashes`（`Enable/Disable`、`SetDash`/`SetGap` 自动 enable + 拒负、`Enabled/Dash/Gap` getter）— ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_StrokeDashes_AEShipGate_AE20{20,25}`：一层 rect+stroke，Dash=18/Gap=7，re-save cdat 解码校验）。
- **磁盘编码**：Dash 1 / Gap 1 均 OneD float64-BE @ cdat[0:8]，嵌套于 `ADBE Vector Stroke Dashes` group `LIST(tdgp)` 内（同 Taper/Wave，`findGroupBody` 下钻 + `overwriteShapeStreamCdat`）。默认 Dash/Gap=10。
- **enable 语义 = 模板切换**：solid stroke 的 Dashes 组是空 placeholder（无 Dash/Gap leaf）。enable 时 `lowerStrokeNode` 切到第二嵌入模板 `v2_2_shape_stroke_dashed_body.bin`（携 Dash 1/Gap 1 slot）；solid `v2_2_shape_stroke_body.bin` **不变** → 现有 stroke/enum/taper-wave gate 零回归（`DisabledStaysSolid` round-trip 证 solid 路径 byte 一致）。hydrate 以 Dash/Gap leaf 存在与否回判 enabled。
- **deferred**：Dash 2/3 + Gap 2/3（AE 只 emit enabled pair，每对需独立模板变体）、**Offset**（AE 端 hidden-until-enabled 且 `setValue` 抛 hidden-property，script-ungettable，无 slot 可建模——RE `v2_2_stroke_dashed.done` 实证）。static-only（不建模 keyframe）。
- RE：`re_stroke_dtw.jsx` + `v2_2_stroke_dashed.aep`（dashed 模板源），详 `incidents/stroke-line-cap-join-miter-re.md` Dashes addendum。
#### V2.2.1 子项⑫ (2026-05-31) — Stroke Taper + Wave (static, %/Wavelength-mode)
- `(s *StrokeNode) Taper()/Wave()` → `StrokeTaper` / `StrokeWave`，各 getter/setter — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_StrokeTaperWave_AEShipGate_AE20{20,25}`：一层 rect+stroke，Taper 6 + Wave 3 标量设非默认，re-save cdat 解码校验全 9 值）。
- **Taper**（6 字段）：Start/End Length、Start/End Width、Start/End Ease，默认全 0。**Wave**（3 字段）：Amount(默认 0)、Wavelength(默认 100)、Phase(默认 0)。全 OneD float64-BE @ cdat[0:8]，嵌套于 group `LIST(tdgp)` 内（比顶层 stroke 标量深一层；`findGroupBody` 下钻 + 复用 `overwriteShapeStreamCdat`）。
- **模板**：`gen_shape_all_full.jsx` 扩展设 Taper+Wave（Units 留 % → 9 active slot emit）→ 仅 stroke body 变更（rect/ellipse/fill/path body md5 不变，零回归）。
- **deferred（elision/coupling 陷阱，见 incident-report）**：Taper Length Units + StartWidthPx/EndWidthPx（% 模式被 elide）、Wave Units + Cycles（Wavelength 模式被 elide / Cycles hidden）。**Stroke Dashes** 已 ship 见子项⑬。
- RE：`re_stroke_dtw.jsx`（probe+set 枚举三组子属性），详 `incidents/stroke-line-cap-join-miter-re.md` Taper/Wave addendum。static-only（不建模 keyframe）。
#### V2.2.1 子项⑪ (2026-05-31) — Shape enum sweep: Direction / Blend Mode / Composite Order / Fill Rule (static)
- Rect/Ellipse `Direction`、Fill `BlendMode`/`CompositeOrder`/`FillRule`、Stroke `BlendMode`/`CompositeOrder` getter/setter — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_ShapeEnums_AEShipGate_AE20{20,25}`：一层 rect+ellipse+fill+stroke 全设 enum，re-save cdat 校验 Direction=3/FillRule=2/BlendMode=3/CompositeOrder=2）。
- **类型**：`ShapeDirection`（Normal 1/Reversed 3）、`ShapeBlendMode`（AE 1-based index，Normal=1，不枚举全表）、`ShapeCompositeOrder`（AbovePrevious 1/BelowPrevious 2）、`FillRule`（NonzeroWinding 1/EvenOdd 2）。全 OneD float64-BE @ cdat[0:8]，默认皆 1（RE 同 `re_shape_enums.jsx`，详 incident-report）。
- **模板**：rect/ellipse/fill/stroke body 全部从单一 `v2_2_shape_all_full.aep`(`gen_shape_all_full.jsx`，每 prop 设非默认)重抽 → 4 模板含 enum slot（path body 不变）。**重跑全部 shape ship-gate 双版本**（Ellipse/EllKf/FillKf/FillOpKf/RectKf/RectSubKf/Stroke/StrokeKf/ShapeEnums × AE2020+2025）均 PASS，模板 swap 零回归。
- RE gotcha：addProperty reindex 使旧 handle 失效（须按 matchName 重取）；详 incident-report。
#### V2.2.1 子项⑩ (2026-05-31) — Stroke Line Cap / Line Join / Miter Limit (static)
- `(s *StrokeNode) LineCap/LineJoin/MiterLimit` getters + `SetLineCap/SetLineJoin/SetMiterLimit` — ✅ **AE 2020+2025 双版本 ship-gate PASS**（`TestV2_2_Stroke_AEShipGate_AE20{20,25}` 扩展：Cap=Projecting(3)/Join=Round(2)/Miter=12，re-save cdat 解码校验）。
- **类型**：`StrokeLineCap`（Butt 1/Round 2/Projecting 3，默认 Butt）、`StrokeLineJoin`（Miter 1/Round 2/Bevel 3，默认 Miter）enum；`MiterLimit` float64（默认 4，setter 拒 <1）。
- **磁盘编码**（RE 自 `re_stroke_linecap.jsx`，详 `incidents/stroke-line-cap-join-miter-re.md`）：三者均 1D OneD，float64-BE @ cdat[0:8]，enum 存 1-based index。matchName 即文档 `ADBE Vector Stroke Line {Cap,Join} / Miter Limit`（旧 "property-not-found" groundwork 是错的）。
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
| `lineOrientation` 横/竖排切换 | layer-local 坐标重排 + 多字段连锁 |
| `MaskPropertyGroup.rotoBezier` | 切换重写整个 shape 顶点表示（+16 字节，4500+ byte-diff） |
| Motion Graphics Template / EP 模板 binding（除 `alternateSource`） | 跨 chunk 复杂结构，P3 罕用 |
| Project 渲染设置（`gpuAccel / colorSpace / expressionEngine`） | P3，AE 24+ 大多锁定为 default。（`Composition.renderer` R/W 已 ship — `SetRenderer` 双版本 ship-gate 绿） |
| Adobe World-Ready composer 切换 | P3 |
| 手动 kerning **首次启用** | 结构性添加（需 AE 先 emit `/8` slot） |
| Camera `FilmSize` setter | ldta `@0x98` 持久化但 ScriptingAPI 不暴露写路径（详 `incidents/camera-filmsize-ldta-write-blocked.md`） |

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
| `Property.valueText` / `propertyParameters`（§3H） | 通用不可达：内置枚举 label 不在文件（pard 只存 nbOptions 计数），需 Adobe 不公开的 per-effect×版本 schema DB；AE 26.0-only API；py-aep 自己也没做。仅自定义 Dropdown Menu Control 的 label 在 `pdnm` chunk 可 RE（有界子集，当前不做）。详 `incidents/valuetext-needs-schema-db.md` |

---

## ❓ 剩余可探方向（非 candidate 列表，需要新发现才动）

可达字段约 99% 已 ship。继续动需要：

1. **ldta `@0x60-0x82` / `@0x8C-0x9F` 零值区 probe** — 高密度 JSX layer-flag 探针，可能挖出 1-2 个零散 flag 或全 negative。
3. **Footage proxy 字段** — 大部分结构性。
4. **Project nhed/nnhd 扩展字段** — 除 BitsPerChannel 外的字节，可能持 ColorSpace / Working Color Profile。
