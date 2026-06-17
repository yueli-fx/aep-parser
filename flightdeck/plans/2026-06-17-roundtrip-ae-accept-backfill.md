---
status: active
summary: 把 capindex 里 261 个仅 verify=roundtrip 的写/做能力,按域做综合 fixture 批量补真 AE 验,升到 ae-accept(读不回的→acceptance)。layer-set 主体已清,SetComment 假绿已修。
last_updated: 2026-06-17
---

# Plan: roundtrip → ae-accept 补验 arc

## 一句话目标

481 能力里 **261 仍仅 `verify=roundtrip`**(只 Go 自读回,从没让真 AE 消化——红线1 假绿温床)。按域做综合 ship-gate fixture,一次 AE 往返验一批,双版本(AE2020+AE2025)。能 DOM readback 的→值验;读不回的(尤其 render-queue 二进制专属字段)→只能 acceptance-preservation。

## 怎么验一批(可复制配方)

1. **挑一批同域 setter**(优先 DOM 明确可读、载体简单 solid 即可的)。
2. **建 fixture**:`internal/aep/<域>_shipgate_test.go` —— 仿 `layer_av_flags_shipgate_test.go`(最干净模板)。流程:NewProject→NewComposition→NewSolidLayer→`Reopen`→批量调 setter→WriteAEP;写 `test_data/<x>_args.json`;`runAeRunShipGate`;读 done 验 PASS;parse resaved 验 Go 字段存活。
3. **verify JSX**:`test_data/verify_<x>.jsx` —— 仿 `verify_layer_av_flags.jsx`。DOM readback 逐字段 check,写 done(PASS/FAIL),resave。
4. **跑双版本**:`AE_SHIP_GATE=1 go test ./internal/aep/ -run 'TestX_AEShipGate_AE2025' -count=1 -v`(先 2025 验 fixture 对,再 AE2020)。两版本 exe 路径默认在 test 里。
5. **升 tag**:对应 `internal/scene/*_writers.go` 的 `//aep:cap` 注释,`verify=roundtrip`→`verify=ae-accept` + `gate=Test...AE2020,Test...AE2025`,boundary 尾"无专门 AE gate→round-trip"换成"双版本 AE gated(<fixture>)"。
6. **regen + commit**:`go generate ./cmd/capindex` → `go vet ./internal/...` → commit(jsx 须 `git add -f`,test_data 被 gitignore)。

## 拉某域待验 setter 清单

```
pwsh -c "(gc docs/capabilities.json -raw|ConvertFrom-Json)|?{$_.cap.verify -eq 'roundtrip' -and $_.cap.domain -eq 'comp'}|%{$_.symbol}"
```
(换 domain 值:layer-set/comp/text/shape/project/keyframe/mask/structural/render-queue/meta/io/expr)

## 进度

- ✅ **批1**(5e98762):layer-set 7 AV-flag — Visible/Shy/Solo/Locked/MotionBlur/Quality/BlendingMode。
- ✅ **批2**(83fff35):layer-set 7 AV-field — InPoint/OutPoint/PreserveTransparency/SamplingBicubic/IsGuide/IsAdjust/Label。
- ✅ **批3**(87cd03b + 0e700cf):SetName/SetStartTime/SetParent + **SetComment(修了假绿)**。
- ✅ **批4**(594409a):camera/light options **17 setter** 升 ae-accept — **零新 fixture,纯补标验证洁癖洞**。早被双版本真 AE gate `TestNewCameraLight_AEShipGate_AE2020/AE2025`(verify_camera_light.jsx DOM readback 逐项 near() + resave-parse 值存活)覆盖,只是 tag 一直停 roundtrip;本批**自跑双版本确认现在真绿**(AE2025 11.6s/AE2020 16.7s,23 项 DOM 值全对)才标。覆盖:9 camera(SetCameraZoom + 8×SetIris*) + 6 light(Color/FalloffType/FalloffStart/FalloffDistance/ShadowDarkness/ShadowDiffusion) + 2 spot(ConeAngle/ConeFeather)。
- ✅ **批5a**(500a6d4):**8 classic Material setter** 升 ae-accept(双版本)。**关键勘探**:唯一现成 material fixture `re_material_options.aep` 是 AE25 存的、**AE2020 直接拒开** → 无法双版本验;故 **author 一个 AE2020-native 载体** `re_material_classic_2020.aep`(AE2020 builder JSX 造 3D solid + 设 classic material 非默认值 materialize),两版本都能开。覆盖 LightTransmission/AcceptsShadows/AcceptsLights/Ambient/Diffuse/Specular/Shininess/Metal。新 gate `TestMaterialClassic_AEShipGate_AE2020/AE2025`。
- ✅ **批6**(comp 域):**22 Composition setter** 升 ae-accept(双版本,**from-scratch gate** `TestCompSettings_AEShipGate_AE2020/AE2025`,5 个 from-scratch comp DOM readback)。comp 设置写 cdta 固定字节无 elision,NewComposition 已双版本 → from-scratch 即载体,无需造文件。调试勘出 2 真问题:**SetFrameRate+SetDuration 同 comp → AE 读时长被 newFps/oldFps 缩放**(真 bug,新 incident `comp-setframerate-no-duration-rescale`,gate 解耦验);**SetDraft3D @0x8A bit0 AE DOM 不反映**(留 roundtrip 待 RE)。
- ✅ **批7**(text 域):**13 SetRun* style setter** 升 ae-accept(双版本)。载体 = from-scratch 单 run 文本(NewTextLayer+SetText("Ag")),每个 SetRun0 浮现为 **whole-doc textDocument** 属性,AE2020 也可读(characterRange 仅 per-run/AE2022+ 才需)。覆盖 FillColor/StrokeColor/ApplyStroke/StrokeWidth/StrokeOverFill/FauxBold/FauxItalic/BaselineShift/AutoLeading/Leading/HScale/VScale/Tsume。**修真 bug**:SetRunTsume 误用 FormatPSNumber(裸整数→AE 当 16.16 定点 /65536),改 FormatPSReal。
- ✅ **批7b**(text 段落):**5 SetParagraph indent/spacing** 升 ae-accept(双版本)。**又抓+修一个真 bug**:5 个 setter 同 tsume FormatPSReal bug(写 20 → AE 读 20/65536)。discovery gate 实证 → 改 FormatPSReal。AE2025 DOM 验 firstLineIndent/spaceBefore/spaceAfter,StartIndent/EndIndent(AE 无 leftMargin/rightMargin DOM)+ AE2020(无段落 DOM)靠双版本 resave-preservation。新 incident `btdk-point-value-needs-formatpsreal`(模式总结,复发两次)。
- ✅ **批7c**(text enum):**9 AE24+ run/paragraph enum** 升 ae-accept(双版本)。勘探坐实:虽标"AE24+ ScriptingAPI 才可写",**AE2020 opaque 保留了它不认识的 AE24+ enum 字节**(resave 后 Go 读回全对)→ 双版本 resave-preservation(AE2025 加 DOM)。覆盖 AutoKernType/BaselineOption/NoBreak/LineJoinType/DigitSet + AutoHyphenate/LeadingType/HangingRoman/Direction。**通用洞察:AE 旧版对不识别的新版属性 opaque 保留 → AE24+ 写能力可经低版本 resave-preservation 双版本验**。
- ✅ **批8**(shape wiggle):**9 Wiggle Paths/Transform procedural setter** 升 ae-accept(验证洁癖洞)。早被双版本 gate `mg_wiggle`/`mg_wiggle_modrt`/`mg_wiggletransform` 的 verify JSX **显式 DOM 值读回**(TemporalPhase/SpatialPhase/Correlation/RandomSeed/WigglesPerSecond),只是标 roundtrip(还是 tier=alpha)。自跑 3 gate 双版本确认绿才补标。boundary 留"本质不可像素门禁(procedural)"(随机/相位单帧无类别正确像素,ae-accept DOM 值验是正确上限)。
- ✅ **批8b/8c**(shape 几何):**11 个** 升 ae-accept。SetDirection(rect/ell,洁癖洞 shape_enums)·新 gate `TestShapeGeom` 覆盖 Rect Roundness/Position·Star Inner/Outer Roundness/Position·Repeater Offset/Rotation·Trim Start/Offset(Go-reparse streamCdat 按 match-name 验值存活,1D+2D)。残 shape 3:SetAnchor(Repeater/Wiggler)/SetScale(Wiggler)——Repeater Anchor match-name 实测错需 RE,Wiggler transform generic match-name 歧义,defer。
- ✅ **批9**(marker):**9 Marker field setter** 升 ae-accept(双版本)。SetChapter/CuePointName/URL/FrameTarget/Duration/FrameDuration/Label/Time/FrameTime —— marker_shipgate 只验了 SetComment。新 gate `TestMarkerFields`,载体 re_compmarker.aep(AddMark 需 clone 模板),4 marker 按 comment 区分,AE MarkerValue DOM 逐字段读回。这 9 个 domain-tag 误为 comp(实 Marker 方法)。
- ✅ **批10**(mask):**8 Mask 选项 setter** 升 ae-accept(双版本)。SetMode/Inverted/Locked/Color/MaskMotionBlur/Feather/Expansion/Closed。新 gate `TestMaskOpts`,from-scratch shape+AddMask 双 Reopen,AE Mask DOM 逐项读回。
- ✅ **批11**(keyframe,96ba292):**12 keyframe mutate setter** 升 ae-accept(双版本)。新 gate `TestKeyframeMutate_AEShipGate_AE2020/AE2025` + verify_kf_mutate.jsx,6 隔离 shape 层 DOM 逐字段读回:SetInInterp/SetOutInterp/SetInTemporalEase/SetOutTemporalEase(EASE)·SetTime/SetValue/SetFrameTime(VALT)·SetInSpatialTangent/SetOutSpatialTangent(TAN,**spatial tangent 首试即读回,无 auto-bezier 覆盖**)·Property.SetStaticValue(STAT)·InsertKeyframe(INS)·DeleteKeyframe(DEL)。create-path(AddKeyframe*)早被 mg_ease render-proven,mutate setter 字节布局同 → 渲染面传递性覆盖,本 gate 验值面。**陷阱**:layer Position 即便 2D 层也解析为 3D(z=0 存储),SetValue/Tangent/StaticValue slice 须长度 3,AE DOM 仍返 2 元素数组。残 keyframe 1:`SetLockedRatio`(tdsb @0x02 bit4,无干净 DOM 入口,defer)。
- ✅ **批12**(structural,55a8d4d):**8 layer-list 结构性 op** 升 ae-accept(**自动双版本 gate,域全收口**)。这 8 个此前只人工 JSX-gated(coverage.md 手记)、无自动 Go gate → 封顶 roundtrip。新 gate `TestStructuralOps_AEShipGate_AE2020/AE2025` + verify_structural_ops.jsx:9 隔离 comp(每 op 一个),from-scratch 跑 op 后 AE DOM 逐 comp 读回 layer 顺序。DeleteLayer/DuplicateLayer/InsertLayer/MoveLayer/MoveToBeginning/MoveToEnd/MoveAfter/MoveBefore。**陷阱**:Delete/Duplicate/InsertLayer 拒绝非 AV 层(shape/text/camera/light)→ 结构性载体须 solid。InsertLayer 自动 gate 仅覆盖同工程;cross-project 仍靠既有 assert-gate(coverage 6/6),boundary 注明。
- ✅ **批13**(layer-set transform/time,b13):**7 setter** 升 ae-accept(双版本)。新 gate `TestLayerXform_AEShipGate_AE2020/AE2025`:SetPosition/SetAnchorPoint/SetScale/SetOpacity(transform 静态)+ SetFrameStartTime/SetFrameInPoint/SetFrameOutPoint(frame→time)。单 shape 层 DOM 值读回。**两陷阱**:① scene `Layer.Set*`(transform 静态)mutate 既有属性,from-scratch 层 default-omission 会 elide → 载体须先经 **create-path(shape PropertyStream)materialize 非默认占位值**,Reopen 后 scene-mutate 到目标(占位≠目标=真测 mutate);② SetInPoint/SetOutPoint 写 source-relative trim,AE DOM inPoint/outPoint=comp-absolute=startTime+此值(批2 startTime=0 故重合)。
- ✅ **批14**(track-matte,4e2fbd2,硬骨头):**classic SetTrackMatte 双版本** ae-accept(AE2020+AE2025,mode byte @0x6B,AE2025 自迁移 classic→显式 DOM 仍读对);**explicit 4(SetTrackMatteSource/Layer/ClearTrackMatteLayer/RemoveTrackMatte)AE2025 单版本已验留 roundtrip**。新 gate `TestTrackMatteClassic_*`(双)+ `TestTrackMatteExplicit_AEShipGate_AE2025`(单)。**关键发现(可达性)**:explicit matte = AE23+ ldta @0xA0 slot,**仅 TargetAE2025 产此 slot**,而 TargetAE2025 fingerprint 被 **AE2024 forward-compat 拒**("文件使用 25.1...无法打开"),无 ≥AE23 中间 target → **双版本不可达**。按 no-single-version-ae-accept 惯例留 roundtrip,boundary 记 AE2025 实证(非 false-green)。**⚠ 待用户决策**:这 4 个是否破例 mint 单版本 ae-accept?(类同 5a ray-traced,但本批有正向 AE 实证)。
- ✅ **批15**(layer-set ldta flag,b15):**SetAutoOrient + SetCollapseTransform** 双版本 ae-accept(新 gate `TestLayerBool_AEShipGate_*`:autoOrient=ALONG_PATH 载体加 position 关键帧;collapseTransformation=true shape 续栅格化)。**抓出 SetStretch 确认 false-green**:写 @0x08/0x6C 分子分母 Go round-trip 绿,AE 读 stretch=100 不认(precomp 源也一样)——AE 按 in/out span 重算,setter 不调 outPoint。**SetStretch 降 stable→alpha**,留 roundtrip,incident `layer-setstretch-ae-recomputes-span`。
- ✅ **批16**(text-font,b16):**AddFont + SetRunFontIndex** 双版本 ae-accept。gate `TestTextFont_AEShipGate_*`:单 run 文本层 AddFont("ArialMT")+SetRunFontIndex→AE textDocument.font=ArialMT(默认 YouYuan 对比)。text roundtrip 6→4(残=3 animator + SetManualKerning 硬尾)。
- ✅ **批17**(source-swap,b17):**SetSource + ReplaceSource** 双版本 ae-accept。gate `TestLayerSource_AEShipGate_*`:MAIN 两 precomp 层源 A→B(ReplaceSource / SetSource),AE layer.source.name=B。
- ✅ **批18**(layer flag w/载体,b18):**SetEffectsEnabled + SetIsNull** 双版本 ae-accept。gate `TestLayerFlags2_AEShipGate_*`:EON(加 effect+switch on)effectsActive=true vs EOFF(SetEffectsEnabled(false))=false;NUL(solid 翻 null bit)nullLayer=true。**抓出 SetTimeRemapEnabled false-green**:enable 只设静态值 0.0(0 关键帧),AE/本库 getter 都以「2 identity 关键帧」表示启用→AE 读 false,getter/setter 不自洽。降 alpha,incident `layer-settimeremapenabled-needs-keyframes`。
- `ae-accept` 35→**200**;`roundtrip` 279→**114**。
- **layer-set 残 24 主要分类(便宜双版本 DOM 验已榨干)**:① 不可达双版本=material-advanced 8 + 3D-geometry 3(AE2020 无 ray-traced/Advanced-3D);② trackmatte-explicit 4(AE2025 单版本验、fingerprint);③ false-green=SetStretch(alpha);④ 需特殊载体=audio(SetAudioEnabled/Levels 需音频源)·frameblend(SetFrameBlendEnabled/PixelMotion 需视频源)·SetTimeRemapEnabled(需有时长 footage)·SetEffectsEnabled(需加 effect,effectsActive read-only)·SetIsNull(uncommon)·alternate-source 2(blsi/EG AE24+)·SetLightSource(AE24+ 环境灯);⑤ 无 DOM=SetMarkersLocked(acceptance)。
- **下一批候选(中等,需载体)**:SetEffectsEnabled(加 effect 验 effectsActive 对比)· SetTimeRemapEnabled(precomp 载体验 timeRemapEnabled)· SetIsNull(solid→null 验 nullLayer)。再往后=audio/video 载体 + acceptance 封顶。

## 待办(按 ROI / 难度排)

### A. layer-set 残项(同域,先清干净)
- **可值验(solid/简单载体)**:`SetStretch`(⚠ ratio↔AE 百分比映射先确认)、`SetAutoOrient`(2D 用 AlongPath,DOM `layer.autoOrient`)、`SetIsNull`(变 null,doc 说 uncommon,验 `layer.nullLayer`)、`SetMarkersLocked`(⚠ 无直接 DOM,可能 acceptance)、`SetEffectsEnabled`(⚠ `layer.effectsActive` read-only,需先加效果)、`SetCollapseTransform`(⚠ solid 不支持,需 precomp/shape 载体)、`SetAudioEnabled`/`SetFrameBlendEnabled`/`SetFrameBlendPixelMotion`(⚠ solid 无音频/帧混合,需视频素材或忽略)。
- **结构/引用类**:`SetSource`/`ReplaceSource`(需第二 source)、`SetTrackMatte`/`SetTrackMatteLayer`/`ClearTrackMatteLayer`/`SetTrackMatteSource`/`RemoveTrackMatte`(需两层 + matte,AE23+)、`SetAlternateSource`/`ClearAlternateSource`(需 blsi slot)。
- ~~**light/iris/camera-zoom 间接覆盖**~~ ✅ **批4 已清**:核实结论坐实"验证洁癖洞"——17 个早被 `TestNewCameraLight_AEShipGate_*` 实测覆盖,只补标。
- **残项(批4 后剩,需独立 fixture)**:
  - ~~`SetMaterial*` 8 classic~~ ✅ **批5a 已清**(500a6d4)。**勘探实证(覆盖原"版本劈裂"假说)**:
    - `re_material_options.aep` 是 **AE25 存的,AE2020 直接拒开**(不是"advanced prop 读不回",是整个 fixture 开不了)→ 双版本不可能用它。
    - 解法 = **author AE2020-native 载体** `re_material_classic_2020.aep`(builder `test_data/build_material_classic_2020.jsx`,AE2020 造 3D solid + 设 classic material 非默认值 materialize)。两版本都能开。**这是 3D/material/renderer 类需双版本 gate 的通用解法:载体必须用目标低版本 author**。
    - **AE2020 classic 渲染器实勘**:16 prop 暴露,**9 classic 可设**(本批 8 + CastsShadows 已 gated);**ShadowColor AE2020 不暴露(MISS)**;**7 ray-traced(Reflection/Glossiness/Fresnel/Transparency/TranspRolloff/IOR/AppearsInReflections)classic 禁用**(setValue 报"惰性/父级隐藏")。
    - **AcceptsShadows/AcceptsLights 默认=1**:Go 设 1 → AE resave elide → Go 读不回(正确行为),DOM readback 为权威证明。
  - **残:material 8(ShadowColor + 7 ray-traced)** 留 roundtrip(boundary 已文档化)。升级须 AE2025 单版本(**无单版本 ae-accept 先例**,全是双版本)或 **author AE2020 ray-traced 载体**(ray-traced 渲染器重、deprecated;ShadowColor/AppearsInReflections 是 2024+ AE2020 真没有)→ ROI 低,**判定基本不可达双版本,搁置**。
  - `SetGeometry*`(3,BevelDirection/PlaneCurvature/PlaneSubdivision):3D Geometry Options,需 Advanced 3D renderer(2024+),**AE2020 无 → 同 ray-traced 不可双版本,搁置**(boundary 待补文档)。
  - `SetLightSource`(AE24+ 环境灯专用,需 environment-light + 源层两载体):niche,最后做。
  - **复用资产**:`tmp_debug/scan_material_fixture`(Go dump 任 aep 每层 material prop 存在性)·`test_data/probe_material.jsx`(AE 跑,dump materialOption 树 matchName+value,版本探测用)——均本地未 tracked。

### B. 其它域(每域一/几个综合 fixture)
- ~~**comp**~~ ✅ **批6 主体已清**:22 cdta-level setter 双版本 ae-accept(from-scratch gate)。**残 comp 项(→ 6b)**:
  - `SetLabel`/`SetComment`:**item-level idta**(from-scratch comp 无 idta chunk,报 "no idta chunk reference")→ 须**解析真 comp 载体**(任一现成 .aep 的 comp)做 gate。SetComment 仍带 item-comment idta-flag 嫌疑(走 `EncodeCmta` 但 item 无 ldta@0x3C 等价物,大概率需某 idta has-comment flag,同 layer SetComment 真因)→ 先 RE:AE 原生设 comp comment→存盘→diff idta 找 flag。
  - `SetDraft3D`:@0x8A bit0,AE2025 `comp.draft3d` DOM 不反映(其余 5 个 @0x8B flag 全反映)→ 留 roundtrip,须 RE 正确 bit/字段。
- ~~**text 33**~~ ✅ **批7 主体已清**:13 SetRun* style 双版本(单 run whole-doc DOM)。**残 text(→7b/7c)**:
  - **SetRun* enum/index 残**:`SetRunFontIndex`(换字体,需多字体)·`SetRunDigitSet`·`SetRunAutoKernType`·`SetRunBaselineOption`(AE24+)·`SetRunLineJoinType`(AE24+)·`SetRunNoBreak` —— 多映射 textDocument 枚举属性,可仿批7 加进同 gate 验(部分 AE24+ 单版本)。
  - ~~**SetParagraph 5 indent/spacing**~~ ✅ **批7b**(含 bug 修)。~~**SetRun enum/index + SetParagraph bool/enum**~~ ✅ **批7c(9 个 AE24+ enum,opaque preservation 双版本)**。
  - **残 text(真硬尾,6 个,documented defer)**:`SetRunFontIndex`+`AddFont`(换字体,需先 AddFont 加第二字体再指,可仿批7c 加 gate=textDocument.font 验,**下个 text slice 首选**)· `SetManualKerning`(incident `kerning-first-enable`,需 AE-native kerning slot,难)· 3 个 text animator(`AddTextRotationX/YAnimator`/`AnimateTextOpacity`,结构性 write-only,render-gate 已 defer)。SetParagraphJustification 已 render-gated。
  - `SetManualKerning`(incident `kerning-first-enable`,难)·`AddFont`(字体表)·text animator(AddTextRotationX/YAnimator/AnimateTextOpacity,结构性,另批)。
- **shape**(批8 后残 ~14):
  - **验证洁癖洞候选(先查 verify 深度再自跑+retag,同批8)**:`SetDirection`(RectNode/EllipseNode,`shape_enums_shipgate` 已 set)· `SetRotation`(star,`mg_polygon` 已 set Rotation)——确认其 verify JSX 真读回值就可补标。
  - **几何参数(render-pixel 相关,需新 gate 或值读回)**:Rect SetRoundness/SetPosition · Star SetInner/OuterRoundness/SetPosition · Repeater SetOffset/Transform(SetAnchor/SetRotation)· Trim SetOffset/SetStart · WigglerTransform SetAnchor/SetScale。多数改渲染几何→红线4 理想 render-pixel,但 ae-accept(DOM 值读回)是合法升级;可仿 mg_* 的 RootGroup().AddX + 值读回。
- ~~**keyframe 13**~~ ✅ **批11**(残 SetLockedRatio 1,defer)。~~**mask 8**~~ ✅ 批10。~~**structural 8**~~ ✅ **批12 全收口**。
- **project 19 / meta 10 / io 4**:多 header/flag,DOM 读不回的→acceptance。**meta 10 多为 read/helper(Open/Parse/FromReader/Reopen/Version/NewPropertyStream/Encode·ParseGradientXML/HasAlternateSourceSlot)→ ae-accept 本不适用(不产出 AE 摄入字节),应保留 roundtrip 或另分类,非补验目标**。
- **comp 残 6**(SetComment/SetLabel 需 item-level idta RE · SetDraft3D @0x8A bit0 需 RE · DuplicateComposition 需结构性 comp gate · SetOrientation/SetPosition 域标可疑,疑似 camera/light mis-domain,先核实是否已被批4 camera gate 覆盖=洁癖洞)。
- **render-queue 40**:⚠ **多数无 ScriptingAPI**(`rq-comment` incident),DOM 读不回,上限=acceptance-preservation,不能值验。最后做、期望值最低。

## 复用资产 / 陷阱(下个对话必读)

- **fixture 模板**:`internal/aep/layer_av_flags_shipgate_test.go` + `test_data/verify_layer_av_flags.jsx`(单层);`layer_av_fields3_*`(两层 + parent)。
- **tmp_debug 工具**(本地,RE 用):`dump_layr <aep>`(Layr LIST children + cmta 字节)、`diff_ldta <a> <b>`(逐字节 diff 第一层 ldta)、`build_setcomment`(复现/round-trip)。RE fixture builder:`test_data/build_re_layer_comment.jsx`(AE 原生生成法,仿它建别的 RE fixture)。
- **SetComment 真因模式**(`incidents/layer-setcomment-cmta-append-position.md`):"字节全对≠AE认"——可能缺一个孤立 ldta/idta flag。RE 法:round-trip AE 原生 byte-identical 锁路径 → diff 找 flag。
- **载体限制**:solid 层无音频、不支持 collapse-transform;某些 setter 需 precomp/视频/两层载体。
- **红线**:值/字节对 ≠ AE 认(红线1);渲染类必验像素(红线4);from-scratch 是弱项,易暴露假绿(红线7)。
- **诚实约束**:不是 261 全能升 ae-accept——读不回的封顶 acceptance。产出是分层的,别刷假绿。

## 下个对话起手式

1. preflight 读 cockpit + 本 plan。
2. 选 A(layer-set 残项,建议先 light/material 核实间接覆盖,省重复)或直接 B-comp(带 item-comment RE)。
3. 按"怎么验一批"配方走。
