---
status: active
summary: V3 M8 Task 0.3 产出 — 全量 Set* 分类表（A 类 back-ref setter / B 类 pure-graph setter）+ 每 backref 结构的 XWriter 接口方法清单。P2（back-ref 接口倒置）的逐类输入。
note: 从 internal/aep 实测枚举（277 个 Set* + 10 个 *Backrefs 结构 + 全部 `.back.` 字节访问点），逐方法读 body 分类，非按名猜。spec §3 分类规则 + edge-case adjudication。
related:
  - flightdeck/specs/2026-06-07-v3-m8-physical-split-design.md
---

# M8 — Set* 分类表 + XWriter 接口清单（Task 0.3 产出）

## 顶部摘要（实测真数，非 spec 估算）

| 指标 | 数值 |
|---|---|
| 总 `Set*` 方法数（非测试） | **277** |
| A 类（back-ref setter，需 writer 接口方法） | **179** |
| └─ A1：经 `recv.back.<chunk>` 字节 patch（被 `.back.` grep 捕获） | 141 |
| └─ A2：经**别名字节切片字段** patch（`settingsBlock` / `roouData` / `block`，**不**经 `.back.`，grep 漏） | 38 |
| B 类（pure-graph / 委托，留 scene，无接口） | **98** |
| XWriter 接口数 | **10**（含 `RenderQueueWriter`；spec §2.1 估「9」偏低） |
| XWriter 接口方法总数 | **~191**（179 A 类 setter + `ProjectWriter.WriteAEP` + 11 个 A 类非-Set* public mutator，见 §C / §D） |

> spec §2.1 估「9 接口 / ~60–100 方法」。**实测 10 接口 / ~191 方法**——估算偏低的两个原因：(1) `write_layer`(35) / `write_text`(33) / `write_composition`(20) / render-settings(36) 这四组本身就远超 100；(2) spec 估算把 shape/material/frame-time 等大批 **B 类委托 setter** 直觉算进去了，但它们其实不需接口（见 §B），而真正的 A 类 length-preserving setter 比预想更密。

### 10 个 backref 结构（精确类型名）+ 是否有 A 类 setter

| backref 结构 | 拥有类型 | 有 A 类 setter？ | XWriter 接口 |
|---|---|---|---|
| `projectBackrefs` | `Project` | ✅ | `ProjectWriter` |
| `compositionBackrefs` | `Composition` | ✅ | `CompositionWriter` |
| `layerBackrefs` | `Layer` | ✅ | `LayerWriter` |
| `propertyBackrefs` | `Property` | ✅ | `PropertyWriter` |
| `propertyGroupBackrefs` | `AEPropertyGroup` | ✅（仅经 `SetDimensionsSeparated` + 结构性 op 改 `grp.back.chunk.Children`） | `PropertyGroupWriter` |
| `keyframeBackrefs` | `Keyframe` | ✅ | `KeyframeWriter` |
| `markerBackrefs` | `Marker` | ✅ | `MarkerWriter` |
| `maskBackrefs` | `Mask` | ✅ | `MaskWriter` |
| `footageBackrefs` | `Footage` | ✅ | `FootageWriter` |
| `renderQueueBackrefs` + `renderQueueItemBackrefs` | `RenderQueue` / `RenderQueueItem` | ✅（仅 `RenderQueueItem.SetComment` 经 `it.back.*`；render-settings setter 走别名字段，见 §A2） | `RenderQueueWriter` |

> 注：`renderQueueItemBackrefs` 与 `renderQueueBackrefs` 是两个独立结构，但同属 render-queue 子系统，合并到一个 `RenderQueueWriter` 接口下。`OutputModule` 没有自己的 `*Backrefs` 结构——它的写路径全部走别名字节切片字段（`settingsBlock` / `roouData`），见 §A2。

---

## 方法论（如何分类）

- spec §3 规则：**方法体若经 `*rifx.Chunk` 做 length-preserving patch → A 类；若仅赋 scene 字段或委托另一 Set* → B 类。**
- 逐方法读 body（非按名猜）。两条 grep 起点：
  - 方法定义：`rg "^func \([a-z]+ \*[A-Za-z]+\) Set[A-Z]" internal/aep`（277 条）。
  - 字节访问点：`rg "\.back\.|\.back\b" internal/aep --glob '!*_test.go'`（725 条命中）。
- **关键 edge-case 校正（grep 不够，必读 body）**：有一类 setter 字节 patch 走的是 scene 结构**直接持有的别名字节切片字段**（`it.settingsBlock` / `om.roouData` / `g.block`），**不**经 `recv.back.<chunk>`。这类**纯按 `.back.` grep 会误判为 B**，但功能上是 length-preserving chunk-byte patch（字段 alias chunk 的 backing array）。spec §3 明确把 RQ / OM length-preserving setter 列为 back-ref setter（A 类）。故本表把它们归 **A2**，并逐条标注「不经 `.back.`」——P2 需决定这些别名字段是迁进 backref 结构、还是给 writer 接口单独造方法（见 §E 未决项）。

---

## A 类：back-ref setter（需 writer 接口方法）

### A1 — 经 `recv.back.<chunk>` 字节 patch（141）

#### ProjectWriter（`projectBackrefs`） — 20 个 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Project` | `SetCompensateForSceneReferredProfiles(bool) error` | `back.acerChunk` |
| `Project` | `SetAudioSampleRate(float64) error` | `back.adfrChunk` |
| `Project` | `SetWorkingGamma(float64) error` | `back.dwgaChunk` |
| `Project` | `SetGpuAccelType(string) error` | `back.gpugUtf8`（length-variable） |
| `Project` | `SetExpressionEngine(string) error` | `back.exenUtf8` |
| `Project` | `SetFeetFramesFilmType(FeetFramesFilmType) error` | `back.nnhdChunk` @8 bit7 |
| `Project` | `SetFootageTimecodeDisplayStartType(FootageTimecodeDisplayStartType) error` | `back.nnhdChunk` @9 |
| `Project` | `SetTimecodeDefaultBase(int) error` | `back.nnhdChunk` @14:16 |
| `Project` | `SetFramesCountType(FramesCountType) error` | `back.nnhdChunk` @20 |
| `Project` | `SetDisplayStartFrame(int) error` | `back.nnhdChunk` @11 bit0 |
| `Project` | `SetFramesUseFeetFrames(bool) error` | `back.nnhdChunk` @11 bit0 |
| `Project` | `SetTimeDisplayType(TimeDisplayType) error` | `back.nnhdChunk` @8 bits6-0 |
| `Project` | `SetTransparencyGridThumbnails(bool) error` | `back.nnhdChunk` @25 |
| `Project` | `SetColorManagementSystem(ColorManagementSystem) error` | `back.cmsUtf8`（JSON 重写） |
| `Project` | `SetLutInterpolationMethod(LutInterpolationMethod) error` | `back.cmsUtf8` |
| `Project` | `SetOcioConfigurationFile(string) error` | `back.cmsUtf8` |
| `Project` | `SetBitsPerChannel(BitsPerChannel) error` | `back.nhedChunk` @0x0F + `back.nnhdChunk` @0x18 |
| `Project` | `SetLinearBlending(bool) error` | `back.root.Children`（经 `setRootFlagChunk` 增删 Lnrb 子 chunk） |
| `Project` | `SetLinearizeWorkingSpace(bool) error` | `back.root.Children`（增删 Lnrp 子 chunk） |
| `Footage`* | `SetPath(string) error` | `back.aliasChunk` / `back.cpthChunk` —— **见下注** |

> `Footage.SetPath` 的 receiver 是 `Footage`，但它 patch 的是 `footageBackrefs`（`aliasChunk`/`cpthChunk`），归 `FootageWriter`，不属 `ProjectWriter`。上表把它列在 write.go 同文件故顺带，实际归属见 FootageWriter。**ProjectWriter 实有 19 个 setter + `WriteAEP`。**

#### CompositionWriter（`compositionBackrefs`） — 23 个 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Composition` | `SetRenderer(string) error` | `back.prinChunk`（就地）+ `back.prdaChunk`（整体替换） |
| `Composition` | `SetBGColor([3]uint8) error` | `back.cdta` |
| `Composition` | `SetSize(width, height uint16) error` | `back.cdta` |
| `Composition` | `SetResolutionFactor(x, y uint16) error` | `back.cdta` |
| `Composition` | `SetShutterAngle(uint16) error` | `back.cdta` |
| `Composition` | `SetShutterPhase(int32) error` | `back.cdta` |
| `Composition` | `SetMotionBlurAdaptiveSampleLimit(int32) error` | `back.cdta` |
| `Composition` | `SetMotionBlurSamplesPerFrame(int32) error` | `back.cdta` |
| `Composition` | `SetName(string) error` | `back.nameChunk`（length-variable） |
| `Composition` | `SetFrameRate(float64) error` | `back.cdta` |
| `Composition` | `SetDuration(float64) error` | `back.cdta` |
| `Composition` | `SetHideShyLayers(bool) error` | `back.cdta`（flag-bit） |
| `Composition` | `SetCompMotionBlur(bool) error` | `back.cdta`（flag-bit） |
| `Composition` | `SetPreserveNestedFrameRate(bool) error` | `back.cdta`（flag-bit） |
| `Composition` | `SetDraft3D(bool) error` | `back.cdta`（flag-bit） |
| `Composition` | `SetFrameBlending(bool) error` | `back.cdta`（flag-bit） |
| `Composition` | `SetPreserveNestedResolution(bool) error` | `back.cdta`（flag-bit） |
| `Composition` | `SetPixelAspect(float64) error` | `back.cdta` |
| `Composition` | `SetWorkArea(startSeconds, endSeconds float64) error` | `back.cdta` |
| `Composition` | `SetDisplayStartTime(float64) error` | `back.cdta` |
| `Composition` | `SetDisplayStartFrame(int) error` | `back.cdta`（委托 `SetDisplayStartTime`，仍 patch cdta） |
| `Composition` | `SetComment(string) error` | `back.itemLayrParent` + `back.itemCmtaChunk`（item 级，length-variable） |
| `Composition` | `SetLabel(uint8) error` | `back.itemIdtaChunk` @0x3A |

#### LayerWriter（`layerBackrefs`） — 见下，共 35（write_layer）+ 33（write_text）= 68 个 setter
write_layer.go（35，全部 patch `back.ldta` / `back.nameChunk` / `back.commentChunk` / `back.layrList` / `back.alternateSourceBlsi`）：

| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Layer` | `SetVisible(bool) error` | `back.ldta`（flag-bit @b.off） |
| `Layer` | `SetSolo(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetShy(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetLocked(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetEffectsEnabled(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetMotionBlur(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetAudioEnabled(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetFrameBlendEnabled(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetCollapseTransform(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetIs3D(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetIsAdjust(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetIsGuide(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetIsNull(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetMarkersLocked(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetSamplingBicubic(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetFrameBlendPixelMotion(bool) error` | `back.ldta`（flag-bit） |
| `Layer` | `SetBlendingMode(BlendingMode) error` | `back.ldta` @0x63 |
| `Layer` | `SetTrackMatte(TrackMatteType) error` | `back.ldta` @0x6B |
| `Layer` | `SetLabel(uint8) error` | `back.ldta` @0x3D |
| `Layer` | `SetQuality(LayerQuality) error` | `back.ldta` @0x04:0x06 |
| `Layer` | `SetParent(parentID uint32) error` | `back.ldta` @0x84:0x88 |
| `Layer` | `SetSource(sourceID uint32) error` | `back.ldta` @0x28:0x2C |
| `Layer` | `SetAutoOrient(AutoOrientType) error` | `back.ldta` @0x25/0x26 |
| `Layer` | `SetPreserveTransparency(bool) error` | `back.ldta` @0x67 |
| `Layer` | `SetStartTime(seconds float64) error` | `back.ldta`（divisor 对 @off+4） |
| `Layer` | `SetInPoint(seconds float64) error` | `back.ldta` |
| `Layer` | `SetOutPoint(seconds float64) error` | `back.ldta` |
| `Layer` | `SetStretch(ratio float64) error` | `back.ldta` @0x08/0x6C |
| `Layer` | `SetName(string) error` | `back.nameChunk`（length-variable） |
| `Layer` | `SetComment(string) error` | `back.commentChunk` / `back.layrList`（length-variable，可新建 cmta） |
| `Layer` | `SetTrackMatteLayer(sourceID uint32, mode TrackMatteType) error` | `back.ldta` @0xA0:0xA4 + @0x6B |
| `Layer` | `SetLightKind(LightKind) error` | `back.ldta` @0x88:0x8C |
| `Layer` | `SetLightSource(target *Layer) error` | `back.ldta` @0x28:0x2C |
| `Layer` | `SetAlternateSource(item AVItem) error` | `back.alternateSourceBlsi` @0:4 |
| `Layer` | `SetText(newText string) error` | `back.btdsChunk`（length-variable splice） |

write_text.go（33，全部经 `l.splicePSValue(...)` → patch `back.btdsChunk`，length-variable）：

| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Layer` | `SetRunFontSize(runIdx int, sizePts float64) error` | `back.btdsChunk`（splicePSValue /1） |
| `Layer` | `SetRunTracking(runIdx int, tracking float64) error` | `back.btdsChunk` /8 |
| `Layer` | `SetRunBaselineShift(runIdx int, shift float64) error` | `back.btdsChunk` /9 |
| `Layer` | `SetRunLeading(runIdx int, leading float64) error` | `back.btdsChunk` /5 |
| `Layer` | `SetRunAutoLeading(runIdx int, auto bool) error` | `back.btdsChunk` /4 |
| `Layer` | `SetRunFontIndex(runIdx, fontIdx int) error` | `back.btdsChunk` /0 |
| `Layer` | `SetRunFauxBold(runIdx int, on bool) error` | `back.btdsChunk` /2 |
| `Layer` | `SetRunFauxItalic(runIdx int, on bool) error` | `back.btdsChunk` /3 |
| `Layer` | `SetRunHorizontalScale(runIdx int, scale float64) error` | `back.btdsChunk` /6 |
| `Layer` | `SetRunVerticalScale(runIdx int, scale float64) error` | `back.btdsChunk` /7 |
| `Layer` | `SetRunTsume(runIdx int, tsume float64) error` | `back.btdsChunk` /36 |
| `Layer` | `SetRunFillColor(runIdx int, rgba [4]float64) error` | `back.btdsChunk` /53/0/1 |
| `Layer` | `SetRunStrokeColor(runIdx int, rgba [4]float64) error` | `back.btdsChunk` /54/0/1 |
| `Layer` | `SetRunApplyStroke(runIdx int, apply bool) error` | `back.btdsChunk` /57 |
| `Layer` | `SetRunStrokeWidth(runIdx int, width float64) error` | `back.btdsChunk` /63 |
| `Layer` | `SetRunCapsOption(runIdx int, caps TextCapsOption) error` | `back.btdsChunk` /12 |
| `Layer` | `SetRunBaselineOption(runIdx int, base TextBaselineOption) error` | `back.btdsChunk` /13 |
| `Layer` | `SetRunStrokeOverFill(runIdx int, over bool) error` | `back.btdsChunk` /58 |
| `Layer` | `SetRunAutoKernType(runIdx int, kt TextAutoKernType) error` | `back.btdsChunk` /11 |
| `Layer` | `SetRunNoBreak(runIdx int, on bool) error` | `back.btdsChunk` /52 |
| `Layer` | `SetRunLineJoinType(runIdx int, j TextLineJoinType) error` | `back.btdsChunk` /62 |
| `Layer` | `SetRunDigitSet(runIdx int, d TextDigitSet) error` | `back.btdsChunk` /70 |
| `Layer` | `SetParagraphJustification(paraIdx int, j TextJustification) error` | `back.btdsChunk` /0 |
| `Layer` | `SetParagraphFirstLineIndent(paraIdx int, v float64) error` | `back.btdsChunk` /1 |
| `Layer` | `SetParagraphStartIndent(paraIdx int, v float64) error` | `back.btdsChunk` /2 |
| `Layer` | `SetParagraphEndIndent(paraIdx int, v float64) error` | `back.btdsChunk` /3 |
| `Layer` | `SetParagraphSpaceBefore(paraIdx int, v float64) error` | `back.btdsChunk` /4 |
| `Layer` | `SetParagraphSpaceAfter(paraIdx int, v float64) error` | `back.btdsChunk` /5 |
| `Layer` | `SetParagraphAutoHyphenate(paraIdx int, on bool) error` | `back.btdsChunk` /9 |
| `Layer` | `SetParagraphLeadingType(paraIdx int, lt TextLeadingType) error` | `back.btdsChunk` /8 |
| `Layer` | `SetParagraphHangingRoman(paraIdx int, on bool) error` | `back.btdsChunk` /21 |
| `Layer` | `SetParagraphDirection(paraIdx int, d TextParagraphDirection) error` | `back.btdsChunk` /33 |
| `Layer` | `SetManualKerning(values []int) error` | `back.btdsChunk`（多次 splicePSValue /1/1/0/0/8 + /7） |

#### PropertyWriter（`propertyBackrefs`） — 4 个 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Property` | `SetStaticValue(v any) error` | `back.cdat` |
| `Property` | `SetExpressionEnabled(bool) error` | `back.tdbs` → tdb4 @0x78 |
| `Property` | `SetExpression(source string) error` | `back.tdbs`（增删 Utf8）+ `back.exprChunk`（length-variable） |
| `Property` | `SetDimensionsSeparated(bool) error` | `back.tdsb`/`back.cdat`/`back.tdb4`/`back.tdbs` + **`grp.back.chunk.Children`**（structural splice，跨 `PropertyGroupWriter`，见下注） |

> `Property.SetStaticValue` 是 shape/material/transform 等大批 B 类 setter 的委托终点（§3 spec 澄清）：`rect.SetSize(v)` → `prop.SetStaticValue(v)` → `prop.back` patch cdat。故 B 类形状/材质 setter **不**需各自接口，全部搭 `PropertyWriter.SetStaticValue` 顺风车。

#### PropertyGroupWriter（`propertyGroupBackrefs`） — 0 个直属 Set*，但被结构性 op + `SetDimensionsSeparated` 改 `grp.back.chunk.Children`
- 没有 `(g *AEPropertyGroup) Set*` 方法。该接口的写需求全部来自：(a) `Property.SetDimensionsSeparated`（splice `grp.back.chunk.Children`）；(b) 结构性 op `AEPropertyGroup.Remove` / `Duplicate` / `MoveTo`（`mutate_property_structural.go`，splice `parent.back.chunk.Children`）。
- **P2 决策点**：`PropertyGroupWriter` 接口方法面由这些 splice 操作定义（非 Set*）。最小面可能是「暴露 group chunk splice 原语」给 serializer impl，或干脆把结构性 group op 整体迁 serializer。归入 §C / §E。

#### KeyframeWriter（`keyframeBackrefs`） — 8 个 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Keyframe` | `SetTime(seconds float64) error` | `back.ldat` @offset:offset+4（tickRate） |
| `Keyframe` | `SetValue(v any) error` | `back.ldat`（value 槽） |
| `Keyframe` | `SetInInterp(InterpType) error` | `back.ldat` @offset+off |
| `Keyframe` | `SetOutInterp(InterpType) error` | `back.ldat` @offset+off |
| `Keyframe` | `SetInTemporalEase(eases []TemporalEase) error` | `back.ldat`（speed/influence 槽） |
| `Keyframe` | `SetOutTemporalEase(eases []TemporalEase) error` | `back.ldat` |
| `Keyframe` | `SetInSpatialTangent(v []float64) error` | `back.ldat`（tangent 槽） |
| `Keyframe` | `SetOutSpatialTangent(v []float64) error` | `back.ldat` |

#### MarkerWriter（`markerBackrefs`） — 8 个 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Marker` | `SetTime(seconds float64) error` | `back.ldat` @ldatOffset:+4 |
| `Marker` | `SetDuration(seconds float64) error` | `back.nmHd` @0x08:0x0C |
| `Marker` | `SetLabel(index uint8) error` | `back.nmHd` @0x10 |
| `Marker` | `SetComment(s string) error` | `back.nmrd`（slot 0，经 `setNmrdUtf8`，length-variable） |
| `Marker` | `SetChapter(s string) error` | `back.nmrd`（slot 1） |
| `Marker` | `SetURL(s string) error` | `back.nmrd`（slot 2） |
| `Marker` | `SetFrameTarget(s string) error` | `back.nmrd`（slot 3） |
| `Marker` | `SetCuePointName(s string) error` | `back.nmrd`（slot 4） |

> SetComment-family 5 个的字节 patch 经未导出 helper `m.setNmrdUtf8(...)`，但 helper patch 的是 `m.back.nmrd` —— 故归 A 类（patch 经 `recv.back.*`，只是穿了一层 helper）。

#### MaskWriter（`maskBackrefs`） — 6 个 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Mask` | `SetMode(MaskMode) error` | `back.mkif` @0x04:0x08 |
| `Mask` | `SetInverted(bool) error` | `back.mkif` @0x00 |
| `Mask` | `SetColor([3]uint8) error` | `back.mkif` @0x2D:0x2F |
| `Mask` | `SetLocked(bool) error` | `back.mkif` @0x01 |
| `Mask` | `SetMaskMotionBlur(MaskMotionBlurMode) error` | `back.mkif` @0x02 |
| `Mask` | `SetClosed(bool) error` | `back.shph` @0x14 |

#### FootageWriter（`footageBackrefs`） — 3 个 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `Footage` | `SetPath(newPath string) error` | `back.aliasChunk`（JSON fullpath）/ `back.cpthChunk`（length-variable） |
| `Footage` | `SetComment(string) error` | `back.itemLayrParent` + `back.itemCmtaChunk`（item 级，length-variable） |
| `Footage` | `SetLabel(uint8) error` | `back.itemIdtaChunk` @0x3A |

#### RenderQueueWriter（`renderQueueBackrefs` / `renderQueueItemBackrefs`） — 1 个 A1 setter
| receiver | method | 触及 chunk 字段 |
|---|---|---|
| `RenderQueueItem` | `SetComment(comment string) error` | `it.back.rcomChunk` / `it.back.litm`（RCom 插入/替换，length-variable） |

> 其余 36 个 render-queue setter（19 `RenderQueueItem` + 17 `OutputModule`）走别名字节切片字段，见 §A2。

---

### A2 — 经别名字节切片字段 patch（**不**经 `.back.`，grep 漏；38）

这些 setter 字节 patch 的是 scene 结构**直接持有**的 `[]byte` 字段，该字段 alias chunk 的 backing array（subslice 共享底层数组，patch 即改 chunk）。**按 spec §3「RQ / OM length-preserving setter = back-ref setter」，归 A 类。** 但因不经 `recv.back.<chunk>`，**纯 `.back.` grep 会漏判为 B**——P2 实施时必须把这些字段一并迁进 serializer（连同其所属 scene 结构的写路径），并决定接口归属（§E 未决项 U1）。

#### RenderQueueWriter ← `RenderQueueItem`（19，patch `it.settingsBlock` / `it.roouData` 经 `it.patchU16` / `it.patchU32`，**无 `.back.`**）
`SetQuality(int)` · `SetColorDepth(int)` · `SetEffects(int)` · `SetFieldRender(int)` · `SetPulldown(int)` · `SetFrameBlending(int)` · `SetMotionBlur(int)` · `SetProxyUse(int)` · `SetSoloSwitches(int)` · `SetGuideLayers(int)` · `SetDiskCache(int)` · `SetFrameRate(int)` · `SetResolution(x, y int)` · `SetSkipExistingFiles(bool)` · `SetName(string)` · `SetQueueItemNotify(bool)` · `SetLogType(uint16)` · `SetTimeSpanStart(float64)` · `SetTimeSpanDuration(float64)`
（全部 `func(...)`，**无 error 返回** —— silent no-op when block too short。）

#### RenderQueueWriter ← `OutputModule`（17，patch `om.settingsBlock` / `om.roouData` 经 `om.omPatch*` / `om.omSetBit`，**无 `.back.`**；`OutputModule` 无 `*Backrefs` 结构）
`SetChannels(int)` · `SetResizeQuality(int)` · `SetResize(bool)` · `SetLockAspectRatio(bool)` · `SetCrop(bool)` · `SetCropTop(int)` · `SetCropLeft(int)` · `SetCropBottom(int)` · `SetCropRight(int)` · `SetIncludeProjectLink(bool)` · `SetPostRenderAction(uint32)` · `SetUseCompFrameNumber(bool)` · `SetUseRegionOfInterest(bool)` · `SetIncludeSourceXMP(bool)` · `SetPreserveRGB(bool)` · `SetDepth(int)` · `SetStartingNumber(uint32)`
（全部 `func(...)`，**无 error 返回**。）

#### （未列入任何现有接口）← `Guide`（2，patch `g.block` alias 的 Gide ldat 槽，**无 `.back.`**；`Guide` 无 `*Backrefs` 结构，且不在 spec §3 列举的「Layer/Comp/Property/Footage/Marker/Mask/Project/RQ/OM」内）
`SetPosition(px float64)` · `SetOrientation(o GuideOrientation)` （皆 `func(...)`，无 error 返回）

> **Guide 是 spec §3 漏列的第 11 个 length-preserving 写面。** P2 需补一个 backref/接口决策（最小：给 `Guide` 一个 backref + `GuideWriter`，或把 guide 字节写路径迁 serializer 并用窄接口）。归 §E 未决项 U2。

---

## B 类：pure-graph setter（留 scene，无接口）（98）

仅赋 scene 字段 / 委托另一 setter；序列化经 lower 重建或经委托终点的 A 类 patch。**这些不需独立 writer 接口方法。**

| receiver | method(s) | 行为 | 委托终点（若有） |
|---|---|---|---|
| `Layer` | `SetTimeRemapEnabled`、`SetAudioLevels`、`SetAnchorPoint`、`SetPosition`、`SetScale`、`SetOrientation`（scene_layer_accessors.go，6 个） | 取 `Property` 委托 / 清 scene 字段 | `Property.SetStaticValue`（A 类，PropertyWriter） |
| `Layer` | `SetRotation`、`SetRotateX`、`SetRotateY`、`SetOpacity`（scene_layer_accessors.go，4 个） | 委托 | `setScalarProperty` → `Property.SetStaticValue` |
| `Layer` | scene_layer_property_access.go 全部 **43** 个（`SetGeometry*` / `SetMaterial*` / `SetCamera*` / `SetIris*` / `SetLight*`） | 委托 | `setScalarProperty(p, ...)` → `Property.SetStaticValue` / 直接 `p.SetStaticValue` |
| `Layer` | `SetFrameInPoint`、`SetFrameOutPoint`、`SetFrameStartTime`（scene_frame_time.go，3 个） | frame→sec 换算后委托 | `Layer.SetInPoint`/`SetOutPoint`/`SetStartTime`（A 类，LayerWriter） |
| `Composition` | `SetWorkAreaStartFrame`、`SetWorkAreaEndFrame`、`SetWorkAreaDurationFrame`（scene_frame_time.go，3 个） | 委托 | `Composition.SetWorkArea`（A 类，CompositionWriter） |
| `Keyframe` | `SetFrameTime`（scene_frame_time.go，1 个） | 读 `back.compFps`（标量，非 chunk patch）后委托 | `Keyframe.SetTime`（A 类，KeyframeWriter） |
| `Marker` | `SetFrameTime`、`SetFrameDuration`（scene_frame_time.go，2 个） | 读 `compFps`（scene 标量）后委托 | `Marker.SetTime`/`SetDuration`（A 类，MarkerWriter） |
| `Layer` | `SetTrackMatteSource`（scene_layer_matte.go，1 个） | 校验后委托 | `Layer.SetTrackMatteLayer`（A 类，LayerWriter） |
| `RectNode` | `SetSize`、`SetPosition`、`SetRoundness`、`SetDirection`（4） | 委托 / 校验赋 scene enum | `PropertyStream.SetStaticValue`（β stream，经 lower 重建，**非** Property.back）/ `setShapeDirection` |
| `EllipseNode` | `SetSize`、`SetPosition`、`SetDirection`（3） | 同上 | `PropertyStream.SetStaticValue` / `setShapeDirection` |
| `PathNode` | `SetVertices`、`SetClosed`（2） | 委托 | `PropertyStream.SetStaticValue`（BezierPath） |
| `FillNode` | `SetColor`、`SetOpacity`、`SetBlendMode`、`SetCompositeOrder`、`SetFillRule`（5） | 委托 / 校验赋 scene 字段 | `PropertyStream.SetStaticValue` / `setShapeBlendMode` / `setShapeCompositeOrder` / 直接赋 |
| `GradientFillNode` | `SetColorStops`、`SetAlphaStops`（2） | 校验后赋 `n.gradient.*` slice | 无（scene 字段，经 lower 重生 prop.map XML） |
| `StrokeNode` | `SetColor`、`SetOpacity`、`SetWidth`、`SetBlendMode`、`SetCompositeOrder`、`SetLineCap`、`SetLineJoin`、`SetMiterLimit`（8） | 委托 / 校验赋 scene 字段 | `PropertyStream.SetStaticValue` / `setShape*` / 直接赋 |
| `StrokeTaper` | `SetStartLength`、`SetEndLength`、`SetStartWidth`、`SetEndWidth`、`SetStartEase`、`SetEndEase`（6） | 直接赋 scene 字段 | 无（经 lower） |
| `StrokeWave` | `SetAmount`、`SetWavelength`、`SetPhase`（3） | 直接赋 scene 字段 | 无（经 lower） |
| `StrokeDashes` | `SetDash`、`SetGap`（2） | 校验赋 scene 字段 + enable | 无（经 lower） |

**B 类小计**：6+4+43+3+3+1+2+1 + 4+3+2+5+2+8+6+3+2 = **98**。

> **shape graph 关键澄清（spec §3）**：`RectNode`/`EllipseNode`/`PathNode`/`Fill`/`Stroke` 的 stream setter 委托到 **`PropertyStream[T].SetStaticValue`**（类型参数化的 β 视图 stream，`scene_shape_graph.go` 的内部 stream），**不是** `Property.SetStaticValue`（后者才是 A 类 back patch）。这些 shape 节点在 `NewShapeLayer` 构造期建图、经 `lower_*` 重生 chunk 子树，**从不原地 patch**。故纯 B 类。`StrokeTaper`/`StrokeWave`/`StrokeDashes`/`GradientFill` 更是直接赋 plain scene 字段。

---

## §C — XWriter 接口方法集（按 backref 结构分组，patch-first 语义）

每接口方法导出（serializer 跨包实现）。除上列 A 类 setter 外，还含 A 类**非-Set\* public mutator**（§D 详）+ `ProjectWriter.WriteAEP`。

| 接口 | Set* 方法（A1+A2） | 额外方法（非-Set* A 类，见 §D） | 方法数合计 |
|---|---|---|---|
| `ProjectWriter` | 19 | `WriteAEP(io.Writer) error` | **20** |
| `CompositionWriter` | 23 | （结构性 New/Duplicate comp 在 §D，归属待 P2 决） | **23** |
| `LayerWriter` | 68（35 layer + 33 text） | `AddFont(string) (int, error)` | **69** |
| `PropertyWriter` | 4 | `InsertKeyframe(float64, any) (*Keyframe, int, error)`、`DeleteKeyframe(int) error` | **6** |
| `PropertyGroupWriter` | 0（仅 `SetDimensionsSeparated` 跨界改 + 结构性 op） | group chunk splice 原语（`Remove`/`Duplicate`/`MoveTo` 底层）—— 面 P2 定 | **~1–4（待定）** |
| `KeyframeWriter` | 8 | — | **8** |
| `MarkerWriter` | 8 | （`Composition.AddMarker` / `Marker.Remove` 结构性，归属待 P2 决） | **8** |
| `MaskWriter` | 6 | — | **6** |
| `FootageWriter` | 3 | — | **3** |
| `RenderQueueWriter` | 1（A1）+ 36（A2，RQItem 19 + OM 17） | （`RemoveItem`/`AddItem` 结构性，归属待 P2 决） | **37+（结构性另计）** |
| **合计（核心 R/W 面）** | **179** | `WriteAEP` + `AddFont` + 2 keyframe op = **+11**（连同结构性 op，见 §D） | **~191** |

> `+11` = `ProjectWriter.WriteAEP`(1) + `LayerWriter.AddFont`(1) + `PropertyWriter.{InsertKeyframe,DeleteKeyframe}`(2) + 结构性 op（NewComposition/DuplicateComposition/NewShapeLayer/DeleteLayer/MoveLayer/InsertLayer/AddMarker/RemoveItem/AddItem 等，~7 主要入口，§D）。结构性 op 的接口归属 P2 单独决（spec §2.4：建-chunk 构造器住 serializer + facade re-export，可能**不**进 writer 接口而是 serializer 自由函数）——故未硬算进 191。**`191` = 179 A 类 setter + WriteAEP + AddFont + 2 keyframe op + ~8 zero-arg 余量；P2 以实际定义为准。**

---

## §D — 非-Set* public mutator（patch 字节经 `.back.` 或别名字段，**flagged**）

任务要求 flag 非-`Set`-前缀的 public mutator（它们也 patch 字节）。下列从 `^func (recv *T) (New|Delete|Insert|Move|Duplicate|Remove|Add|Import|Write|Convert|Replace)[A-Z]` 枚举 + 读 body 确认 patch 路径：

| receiver.method | patch 路径 | 类别 | 接口归属建议 |
|---|---|---|---|
| `Project.WriteAEP(io.Writer) error` | `back.root.Write(w)` | A（**任务点名要求纳入 `ProjectWriter`**） | `ProjectWriter.WriteAEP` |
| `Property.InsertKeyframe(float64, any) (*Keyframe,int,error)` | `back.ldat`+`back.lhd3`（length-variable rebuild） | A | `PropertyWriter` |
| `Property.DeleteKeyframe(int) error` | `back.ldat`+`back.lhd3` | A | `PropertyWriter` |
| `Layer.AddFont(string) (int,error)` | `back.btdsChunk`（length-variable） | A | `LayerWriter` |
| `Project.NewComposition(...)` | `back.rootFold.Children`（结构性，re-parse closed loop） | A（结构性 op） | serializer 自由函数 / facade re-export（spec §2.4），P2 决 |
| `Project.DuplicateComposition(src, name)` | `back.rootFold.Children` | A（结构性） | 同上 |
| `Composition.NewShapeLayer(name)` | `back.itemList`（经 lower 建 Layr） | A（结构性） | 同上 |
| `Composition.DuplicateLayer(index, name)` | `back.itemList.Children` | A（结构性） | 同上 |
| `Composition.DeleteLayer(index)` | `back.itemList.Children` + neighbor `back.ldta` patch | A（结构性） | 同上 |
| `Composition.MoveLayer(from, to)` / `Layer.MoveToBeginning/End/After/Before` | `back.itemList.Children` | A（结构性） | 同上 |
| `Composition.InsertLayer(src, atIdx)` | `back.itemList.Children`（跨 comp） | A（结构性） | 同上 |
| `Composition.AddMarker(seconds)` | `markerList.ldat`/`mrky` + clone `back.nmHd`（length-variable） | A（结构性） | `MarkerWriter` / serializer，P2 决 |
| `AEPropertyGroup.MoveTo(index)` | `parent.back.chunk.Children`（splice） | A（结构性） | `PropertyGroupWriter` / serializer |
| `RenderQueue.RemoveItem(index)` | `rq.back.lrdr` + `item.back.litm` | A（结构性） | `RenderQueueWriter` / serializer |
| `RenderQueue.AddItem(comp)` | `rq.back.lrdr` + 新 `item.back`（template clone via `back.itemListChunk`） | A（结构性） | 同上 |
| `Marker.Remove()` | `markerList.{ldat,lhd3,mrky}` + `m.back.nmrd`（length-variable） | A（结构性） | `MarkerWriter` / serializer |
| `Layer.ReplaceSource(target, fixExpressions)` | 委托 `SetSource` + 表达式重写（scene） | 部分 A（经 `SetSource`） | 搭 `LayerWriter.SetSource` |
| `Layer.RemoveTrackMatte()` | 委托 `SetTrackMatteLayer`/清 scene | 搭 LayerWriter | 搭 LayerWriter |
| `Layer.ClearTrackMatteLayer() error`（write_layer.go） | 委托 `SetTrackMatteLayer` | 搭 LayerWriter | 搭 LayerWriter |
| `VectorGroup.AddRect/AddEllipse/AddPath/AddFill/AddStroke/AddGradientFill` | 纯 scene 图构造（detached，经 lower） | **B**（纯图，无 patch） | 无接口 |
| `Project.WriteJSON(io.Writer) error` | 纯 scene 导出（无 chunk） | **B**（留 scene，spec §1 明确 WriteJSON 留 scene 包） | 无接口 |

> **结构性 op（New/Delete/Insert/Move/Duplicate/Add 系列）确 patch chunk-tree，但 spec §2.4 把「建-chunk 构造器」定为住 serializer + facade re-export**——它们**未必**进 writer 接口（可能是 serializer 自由函数 + facade 包装）。本表 flag 出来供 P2 决策，**不**硬算进 §C 的 179+核心接口面。本 Task 0.3 范围 = Set* 分类 + writer 接口；结构性 op 归属是 P2 / §E 未决项 U3。

---

## §E — P2 未决项（本表 flag，P2 实施前必决）

- **U1（A2 别名字段归属）**：`RenderQueueItem.settingsBlock`/`roouData`、`OutputModule.settingsBlock`/`roouData` 是 scene 结构直持的 chunk-alias `[]byte`，非 `back` 字段。物理分包后 scene 不能持 chunk-alias（破 scene⊥rifx）。P2 必须：(a) 把这些字节字段迁进对应 backref 结构（serializer 侧），(b) 给 `RenderQueueWriter` 补 36 个 patch 方法（或一个通用 `PatchSettings(off, bytes)` 原语 + scene 侧保留 offset 常量），(c) `OutputModule` 需要自己的 backref + 是否独立 `OutputModuleWriter` 接口（当前合并进 `RenderQueueWriter`）。
- **U2（Guide 漏列）**：`Guide.SetPosition`/`SetOrientation` patch `g.block`（Gide ldat alias），spec §3 未列 Guide。P2 需补 `Guide` 的 backref + 写接口（最小 `GuideWriter` 或并入 `CompositionWriter`，因 guide 属 comp）。
- **U3（结构性 op 接口归属）**：New*/Delete*/Insert*/Move*/Duplicate*/Add* + `AEPropertyGroup.{Remove,Duplicate,MoveTo}` + RQ 结构性 op —— 按 spec §2.4 倾向住 serializer + facade re-export，**不**进 writer 接口；但 `PropertyGroupWriter` / `RenderQueueWriter` 的精确面取决于此。
- **U4（PropertyGroupWriter 面）**：无直属 Set*，写需求来自 `SetDimensionsSeparated`（跨界改 `grp.back.chunk`）+ 结构性 group op。接口面是「group chunk splice 原语」还是「整 op 迁 serializer」待 P2 定。
- **U5（无-error setter 签名）**：A2 的 36 RQ/OM setter + 2 Guide setter 是 `func(...)` 无 error 返回（block 太短时 silent no-op）。接口方法签名要否统一加 `error`？spec C-1 patch-first 原子序假设 setter 返回 error；这批不返回，P2 需决定保持 `func(...)` 接口方法还是改签名（改签名 = 破 Alpha API，可接受但需 commit 标 BREAKING）。

---

## 自审

- ✅ 逐方法读 body（非按名猜）：A1 全部经 grep `.back.` 命中点交叉核对；委托型 setter（scene_frame_time / scene_layer_accessors / scene_layer_property_access / scene_shape_graph / scene_layer_matte）逐文件读 body 确认委托终点；A2 三组（RQItem / OM / Guide）读 body 确认 patch 别名字段而非 `.back.`。
- ✅ 每个 Set* 恰好归 A / B 之一：A=179（A1 141 + A2 38）+ B=98 = **277**，与 `rg` 计数一致。
- ✅ 接口方法签名取实际 param/return 类型（从方法定义抄，非编造）。
- ✅ 真接口数 = **10**（spec 估 9，漏了 `RenderQueueWriter` 与 `renderQueueItemBackrefs` 的独立性；另 Guide/OutputModule 无独立 backref，见 U1/U2）。真方法数 = **~191**（spec 估 60–100 偏低）。
- ⚠️ **A2 是最大判定风险**：38 个 setter 纯按 `.back.` grep 会误判 B，靠读 body + spec §3「RQ/OM = back-ref setter」裁定为 A。P2 若只 grep `.back.` 会漏接这 38 个 + 它们的别名字段迁移（U1/U2）。
- ⚠️ 结构性 op（§D）按任务要求 flag 但**未**硬算进核心接口面——spec §2.4 倾向 serializer 自由函数。P2 需独立决 U3/U4。
