---
status: active
summary: 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
last_updated: 2026-06-18
---

# roundtrip→ae-accept 补验 arc 完成记录 + 残值 backlog(53)

> 这份 spec 是 **roundtrip→ae-accept 补验 arc 的收尾文档**:上半是已完成记录(为什么做、做了什么、得到的通用洞察),下半是**剩 53 个 `verify=roundtrip` 残值的 someday-backlog**——大半本就不该升,真要继续按下表挑一项。真相源仍是 capindex(`go run ./cmd/capindex -q <词>`);本表是人读的导航。

## 一、Arc 完成记录(批1-28)

**起因**:库能力主线早已全收口,但 capindex 里 **279 个写/做能力仅 `verify=roundtrip`**——只 Go 解析器自读回,从没让真 AE 消化(红线1「Go round-trip 0 警告 ≠ AE 接受」的假绿温床)。本 arc 系统性按域做综合 ship-gate fixture,一次真 AE 往返验一批,能 DOM readback 的值验、读不回的封顶 acceptance-preservation。

**结果**:`ae-accept` 35→**261**,`roundtrip` 279→**53**。按域(批次):

- layer-set 主体 AV-flag/field/name(批1-3)· camera/light 17(批4)· classic material 8(批5a)· Composition 22(批6)· text 27(批7/7b/7c)· shape 几何+wiggle 20(批8/8b/8c)· marker 9(批9)· mask 选项 8(批10)· keyframe 12 mutate(批11)· structural 8 op(批12)· layer-set transform/time 7(批13)· track-matte classic(批14)· AutoOrient/CollapseTransform(批15)· text-font(批16)· source-swap(批17)· EffectsEnabled/IsNull(批18)· DuplicateComposition(批19)· project 设置 9(批20)· comp idta(批21)· footage idta(批22)· shape transform 收口(批23)· **audio 2(批24)· frameblend 2(批25)· track-matte explicit 4 单版本(批26)· render-queue value-setter 36(批28)**。
- **收口域**(roundtrip 清零或仅留已记残值):shape · footage-io · audio · frameblend · track-matte · render-queue value-setter · mask · keyframe(留 1)· expr(closed)。

**修的 4 个真 bug(都是 round-trip 绿但 AE 不认)**:① layer SetComment 假绿(ldta @0x3C has-comment flag + cmta double-NUL)· ② **item cmta 位置**(批21,必须紧跟 Utf8 非末尾,否则 AE 打不开文件)· ③ SetFrameRate+SetDuration 时长被 newFps/oldFps 缩放(incident)· ④ btdk point 值须 FormatPSReal(批7/7b,否则 AE 读 值/65536)。

**通用洞察(已各自落 incident/契约)**:
- **default-omission**:from-scratch 层 elide 默认属性,mutate 既有属性的 setter 拿到 nil → 载体须先 materialize 非默认占位(批13/24 模板)。
- **opaque preservation**:AE2020 保留它不识别的 AE24+ 属性字节 → AE24+ 写能力可经低版本 resave-preservation 双版本验(批7c)。
- **验证洁癖洞**:既有 render gate 常已 set+DOM 读回某 setter,只是 tag 停 roundtrip;自跑确认绿后补标即可(批4/8/19/20)。
- **单版本 ae-accept 例外(契约,2026-06-18 用户授权)**:物理不可达双版本者可凭单版本 mint ae-accept,但 tag 须写清版本下限(`minver=` + boundary 首句 `ae-accept=AE2025+ 单版本验证`)。详 `checklists/delivery-contract.md`。首例批26 track-matte explicit。
- **`.enabled=true` ≠ settable**:判属性可写必须真 setValue 试,别看 `.enabled`(批27 钓鱼坑)。
- **acceptance+resave-preservation**:无 ScriptingAPI readback 的二进制字段(RQ settings),ae-accept 上限 = AE 开不损坏 + resave 后 Go 重解析字节存活(批28,同 SetComment)。

## 二、残值 backlog(53,someday-when-needed)

按「该不该升 / 能不能升」分四类。**A=本就不该升(N/A)**,**B=实勘判定不可达(负结论,别再试)**,**C=false-green/无 DOM(已记 incident)**,**D=可做但低 ROI(真要继续从这挑)**。

### A. 本就不该升 ae-accept(read/helper/writer 入口,不产 AE 摄入字节)— 12

- **meta 10**:`Open` `Parse` `ParseReader` `FromReader` `Reopen` `Version` `NewPropertyStream` `HasAlternateSourceSlot` `ParseGradientXML` `EncodeGradientXML`。读/解析/helper,ae-accept 概念不适用。
- **io 2**:`WriteAEP` `WriteJSON`。写入口本身(它们就是 round-trip 的「写」端),无独立 AE 验对象。

> **处理**:这 12 个**不是补验目标**,应视为 roundtrip-N/A。若将来 capindex 加 `verify=na` 档可迁过去;否则永久留 roundtrip 当背景噪声。

### B. 实勘判定不可达 AE-gate(负结论,incident 已记,别重试)— 11

- **layer-set material-advanced 8**:`SetMaterialReflection` `SetMaterialGlossiness` `SetMaterialFresnel` `SetMaterialTransparency` `SetMaterialTranspRolloff` `SetMaterialIndexOfRefraction` `SetMaterialAppearsInReflections` `SetMaterialShadowColor`。
- **layer-set 3D-geometry 3**:`SetGeometryBevelDirection` `SetGeometryPlaneCurvature` `SetGeometryPlaneSubdivision`。

> **处理**:incident `material-advanced-props-hidden` 实锤——Advanced-3D 组里列出+`.enabled=true` 但 `setValue` 被 AE 拒「父级属性被隐藏」,solid+挤出 shape × AE2020+AE2025 四组合全拒,无渲染器 un-hide。**唯一可能路径** = 用户手动在 AE UI 造非默认材质/几何的挤出工程存 .aep(脱离无人值守),ROI 极低。除非有人专门要这 11 个,别碰。

### C. false-green / 无 DOM(已记 incident,封顶 roundtrip)— 8

- **layer-set 2**:`SetStretch`(incident `layer-setstretch-ae-recomputes-span`,AE 按 in/out span 重算)· `SetTimeRemapEnabled`(incident `layer-settimeremapenabled-needs-keyframes`,需 2 identity 关键帧)。
- **comp 1**:`SetDraft3D`(incident `comp-setdraft3d-false-green`,cdta @0x8A bit0 字节对但 DOM 读 false)。
- **project 1**:`SetLinearizeWorkingSpace`(incident `project-flag-chunks-lnrb-lnrp`,lnrp 字节对+AE 接受,但 DOM 即便 AE 自存也读 false)。
- **render-queue 2**:`SetPreserveRGB`(批28,AE2025 resave 清 @0x07 bit7,输出色彩管理 gate)· `SetQueueItemNotify`(批28,AE2020 resave 清 @0x07 bit2,版本相关)。
- **expr 1**:`SetExpressionEnabled`(closed decision,incident `expression-enable-byte-pair`;机制已双版本 gated,单独 enable bit 不补)。

> **处理**:这些**字节写入端正确**,缺的是 AE 消化端反映/AE 主动归一化——非 bug,是 AE 语义。已各记 incident。**不升**。

### D. 可做但低 ROI(真要继续 arc,从这挑)— 22

- **keyframe 1**:`SetLockedRatio`(tdsb @0x02 bit4,无干净 DOM 入口,可试 acceptance-preservation)。
- **project 9(部分可仿批28 acceptance-preservation gate)**:`SetWorkingGamma` `SetAudioSampleRate` `SetTimecodeDefaultBase` `SetTransparencyGridThumbnails` `SetGpuAccelType`(机器相关)· **CMS 4**:`SetColorManagementSystem` `SetOcioConfigurationFile` `SetLutInterpolationMethod` `SetCompensateForSceneReferredProfiles`(AE24+,需既有 CMS chunk,AE2020 开不了→单版本或不可达)· `SetPath`(需 footage 载体)。
  - **怎么做**:二进制 header 字段无 DOM 的,仿 `TestRenderQueueSettings`(批28)——from-scratch 或既有 project 载体设非默认→AE 开+resave→Go 重解析验存活=acceptance-preservation 升 ae-accept。CMS 4 大概率单版本(AE2025)。
- **layer-set 4(niche/特殊载体)**:`SetMarkersLocked`(无 DOM,acceptance)· `SetLightSource`(AE24+ 环境灯,需 environment-light + 源层两载体)· `SetAlternateSource` `ClearAlternateSource`(blsi/EG Essential-Props slot,AE24+,incident `altsource-wrapper-precomp`)。
- **comp 2(Guide layer)**:`SetOrientation` `SetPosition`(Guide 层,无直接 DOM;疑 acceptance-preservation 可达)。
- **text 4(硬尾,结构性)**:`AnimateTextOpacity` `AddTextRotationXAnimator` `AddTextRotationYAnimator`(text animator 结构性 write-only,render-gate 已 defer,incident `text-animator-create-re`)· `SetManualKerning`(incident `kerning-first-enable`,需 AE-native kerning slot,难)。
- **render-queue 2(结构性)**:`AddItem` `RemoveItem`(`RemoveItem` 有 `render_queue_remove_shipgate_test` 但 tag 停 roundtrip=洁癖洞,核实后可补标;`AddItem` 需结构性 acceptance gate)。

> **处理(优先级)**:① **render-queue AddItem/RemoveItem 洁癖洞**——最便宜,核实既有 remove gate 是否真验后补标。② **project header acceptance-preservation**(仿批28,中等)。③ comp Guide 2 / layer-set niche 4(需特殊载体,中等)。④ text 硬尾 4(结构性,最贵)。

## 三、起手式(将来某天接 D 类)

1. `go run ./cmd/capindex -q <symbol>` 看当前 boundary。
2. 仿模板:值验→`layer_xform_static_shipgate_test.go`;acceptance-preservation→`render_queue_settings_shipgate_test.go`(批28)。
3. 单版本判据 + tag 写法 → `checklists/delivery-contract.md` 单版本例外段。
4. 配方/陷阱 → 已归档的 `plans/2026-06-17-roundtrip-ae-accept-backfill.md`(landed/)。
