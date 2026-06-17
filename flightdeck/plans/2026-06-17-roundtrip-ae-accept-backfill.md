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
- `ae-accept` 35→**78**;`roundtrip` 279→**236**。

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
- **comp 37**:⚠ **带着 item-comment idta-flag 嫌疑去**——`Composition.SetComment` 走同款 `EncodeCmta`(现已 double-NUL)但 item 无 ldta@0x3C,大概率需某 idta flag(同 SetComment 真因)。先 RE:AE 原生设 comp comment→存盘→`diff` idta 找 has-comment flag。其余 comp setter(duration/framerate/bg/resolution…)DOM 可读。
- **text 33**:btdk 字段,DOM 可读(font/size/justification/tracking…),部分已有 text gate,核实重复。
- **shape 23**:⚠ render-pixel 类(颜色/描边),按红线4 验渲染像素不只值。
- **keyframe 13 / mask 8 / structural 8**:渲染或结构接受验。
- **project 19 / meta 16 / io 4**:多 header/flag,DOM 读不回的→acceptance。
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
