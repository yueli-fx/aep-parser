# Cockpit — aep-parser

**Last updated**: 2026-06-17 by claude（本会话连推补验 arc 批11-17 + chunk-probe 实验 + parse 提速 53×：**批11 keyframe 12 mutate**·**批12 structural 8 op 全收口**·**批13 layer-set transform/time 7**(SetPosition/Anchor/Scale/Opacity + SetFrame{In,Out}Point/StartTime)·**批14 track-matte**(classic SetTrackMatte 双版本;explicit 4 AE2025 单版本验、双版本不可达留 roundtrip)·**批15 layer-set flag**(SetAutoOrient + SetCollapseTransform 双版本;**SetStretch 抓出确认 false-green**——写 stretch 分数 AE 按 in/out span 重算回 100,降 stable→alpha,incident 记)·**批16 text-font**(AddFont+SetRunFontIndex 双版本,DOM textDocument.font=ArialMT)·**批17 source-swap**(SetSource+ReplaceSource 双版本,precomp 源 A→B)。ae-accept 164→**198**,roundtrip 150→**116**。**另**:诊断 + 修了 parse 性能(profile 实证 Open 把 raw *os.File 喂 parser → 每 chunk 数次 syscall 占 97%;改 os.ReadFile+bytes.Reader,8MB 工程 parse 2s→37ms ~53×,4656879)。**另**:chunk-probe 实验坐实「AE 保存丢弃任意未知 RIFX chunk(root+item,双版本)、打开不损坏」→ incident `ae-drops-unknown-chunks-on-resave`(元数据走 native comment / XMP-未支持)。陷阱:2D layer Position 解析为 3D(z=0)·transform 静态 setter 需 materialized 载体·SetInPoint=source-relative(AE=startTime+)·explicit matte TargetAE2025 fingerprint 被 AE2024 拒。**⚠ 待用户决策**:explicit-matte 4 是否破例 mint 单版本 ae-accept。）

**Active focus**: **roundtrip→ae-accept 补验 arc**(2026-06-17 起;详 `## 下一步` + `plans/2026-06-17-roundtrip-ae-accept-backfill.md`)。库能力主线早已全收口,此 arc 把"标 stable 但只 Go round-trip、从没让真 AE 消化"的写能力按域补双版本 AE 验(验证洁癖洞)。已推进至批17(ae-accept 35→**198**,roundtrip→**116**):layer-set 主体(批1-3)+ camera/light 17(批4)+ classic material 8(批5a)+ Composition 22(批6)+ text 27/33(批7/7b/7c)+ shape 20(批8/8b/8c)+ marker 9(批9)+ mask 选项 8(批10)+ keyframe 12 mutate(批11)+ structural 8 op 全收口(批12)+ layer-set transform/time 7(批13)+ track-matte classic 双版本(批14;explicit 4 单版本留 roundtrip)+ layer-set AutoOrient/CollapseTransform(批15;SetStretch 抓出 false-green 降 alpha)+ **text-font AddFont/SetRunFontIndex(批16)+ source-swap SetSource/ReplaceSource(批17)**。修 3 真 bug(SetComment 假绿 · SetFrameRate+Duration incident · btdk FormatPSReal incident:tsume+5 段落)。两通用洞察:**AE2020 opaque 保留不识别的 AE24+ 属性**(批7c)· **既有 render gate 常 set+DOM 读回 roundtrip-tag 的 setter = 验证洁癖洞**(批4/8,自跑确认绿后补标)。**不变量**:知识单一家 = flightdeck + CLAUDE.md(auto-memory 已退役);能力真相源 = capindex(`go run ./cmd/capindex -q <词>`,CI 强制零漏标);每渲染类双版本 AE ship-gate(红线4)。火焰=番外(`incidents/procedural-fx-over-vector.md`)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-17-roundtrip-ae-accept-backfill.md](plans/2026-06-17-roundtrip-ae-accept-backfill.md) — 把 capindex 里 261 个仅 verify=roundtrip 的写/做能力,按域做综合 fixture 批量补真 AE 验,升到 ae-accept(读不回的→acceptance)。layer-set 主体已清,SetComment 假绿已修。
<!-- /AUTO -->

## 下一步

**➡ ACTIVE ARC:roundtrip→ae-accept 补验**(2026-06-17 起)。库 481 能力里 **261 个仍仅 verify=roundtrip**(只 Go 自读回——红线1 假绿温床)。系统性按域做综合 fixture 批量补真 AE 验。**📋 逐项 drain 清单 + 验一批配方 + 复用资产 + 陷阱 → `plans/2026-06-17-roundtrip-ae-accept-backfill.md`(下个对话起手读它就能接着干)**。**诚实约束**:能 DOM readback→值验;读不回的(尤其 render-queue 40 个二进制字段)→封顶 acceptance-preservation。
- **批1 ✅**(5e98762):7 AV-flag(Visible/Shy/Solo/Locked/MotionBlur/Quality/BlendingMode)。
- **批2 ✅**:7 AV-field(InPoint/OutPoint/PreserveTransparency/SamplingBicubic/IsGuide/IsAdjust/Label)。
- **批3 ✅**:SetName(length-variable)/SetStartTime/SetParent(两层 fixture)。`ae-accept` 35→**52**,roundtrip→262。
- **批3 抓出 + 修好真假绿**:`Layer.SetComment` from-scratch 层 AE 读回空。RE **推翻"位置"假说**(cmta 本就该在 Layr 末尾),真因 = ldta **@0x3C has-comment flag**(此前未解字节)+ cmta **double-NUL**。已修 + 纳入批3 gate,`ae-accept` 52→**53**。incident RESOLVED(`layer-setcomment-cmta-append-position`)。**连带:item-level setItemComment 同走 EncodeCmta 但无 ldta@0x3C 等价物,补 comp 域须 RE idta flag**。
- **批4 ✅**(594409a):camera/light options **17 setter** 升 ae-accept(9 camera Zoom+8 Iris / 6 light / 2 spot)。**零新 fixture**——核实出"早被双版本 `TestNewCameraLight_AEShipGate_*` DOM-readback+resave 覆盖,只是 tag 停 roundtrip",自跑双版本确认真绿才标。`ae-accept` 53→**70**,roundtrip→**244**。
- **批5a ✅**(500a6d4):8 classic Material setter 双版本 ae-accept。关键 = author AE2020-native 载体(通用解法:3D/material/renderer 双版本 gate 须用目标低版本 author fixture)。残 material 8(ShadowColor + 7 ray-traced)+ geometry 3 判定基本不可达双版本(AE2020 classic 不暴露/禁用,无单版本 ae-accept 先例),搁置;详 plan。
- **批6 ✅**(comp 22 setter 双版本 from-scratch gate)。残 comp(→6b):SetLabel/SetComment(item-level idta,需真 comp 载体 + SetComment item-comment idta-flag RE)· SetDraft3D(@0x8A bit0 DOM 不反映,RE)。
- **批8-10 ✅**(shape 20 / marker 9 / mask 选项 8)。残 shape 3(Anchor/Scale match-name 需 RE)。
- **批11 ✅**(keyframe,96ba292):12 mutate setter(interp/ease/tangent/value/time/static/insert/delete)。新 gate keyframe_mutate。残 keyframe 1:SetLockedRatio(tdsb @0x02 bit4,无干净 DOM 入口,defer)。
- **批12 ✅**(structural,55a8d4d,**域全收口**):8 layer-list op(Delete/Duplicate/Insert/MoveLayer + 4 Move 封装)。新自动 gate structural_ops 取代旧人工 JSX-gate;载体须 solid/AV(非 AV 层被拒)。
- **批13 ✅**(layer-set transform/time,b13):7 setter(SetPosition/Anchor/Scale/Opacity + SetFrame{In,Out}Point/StartTime)。陷阱:transform 静态 setter 须 materialized 载体(default-omission)· SetInPoint=source-relative。
- **批14 ✅**(track-matte,4e2fbd2):classic SetTrackMatte 双版本 ae-accept;explicit 4 AE2025 单版本验、**双版本不可达(TargetAE2025 fingerprint 被 AE2024 拒)**留 roundtrip。**⚠ 待用户决策**:explicit-matte 4 是否破例 mint 单版本 ae-accept。
- **批15 ✅**(layer-set flag,b15):SetAutoOrient + SetCollapseTransform 双版本 ae-accept。**抓出 SetStretch 确认 false-green**(AE 按 in/out span 重算 stretch 回 100,降 alpha,incident `layer-setstretch-ae-recomputes-span`)。
- **批16 ✅**(text-font,b16):AddFont + SetRunFontIndex 双版本(DOM textDocument.font=ArialMT)。text 残 4=硬尾(3 animator + SetManualKerning)。
- **批17 ✅**(source-swap,b17):SetSource + ReplaceSource 双版本(precomp 源 A→B,DOM layer.source.name=B)。
- **roundtrip 剩 116,域分布**:render-queue 40(多无 ScriptingAPI,acceptance 顶,期望低)· layer-set 24(便宜双版本已榨干;残:material-advanced 8/3D-geometry 3=不可达双版本· trackmatte-explicit 4=单版本验· SetStretch=false-green-alpha· audio/frameblend/timeremap/effects/IsNull/alternate-source/light-source 需特殊载体· markers-locked 无 DOM)· project 19(多 header/flag→acceptance)· meta 10(**read/helper,ae-accept 不适用,非目标**)· comp 6(idta RE)· text 4(硬尾)· io 4 · shape 3(RE)· expr 2(closed)· keyframe 1(defer)。
- **下一批候选(中等,需载体)**:SetEffectsEnabled(加 effect 验 effectsActive)· SetTimeRemapEnabled(precomp 载体)· SetIsNull(solid→null)。再往后 audio/video 载体 + acceptance 封顶。
- **下一批候选**(剩多为硬骨头/低 ROI/不可达):layer-set bool 残(Stretch/AutoOrient/IsNull 可能 solid 验,audio/frameblend/timeremap/collapse 需 video/precomp 载体)· text-font(SetRunFontIndex+AddFont,plan 标 text 首选)· source/replace(需第二 source)· project/render-queue(多 acceptance 封顶)。material-advanced 8 + 3D-geometry 3 + trackmatte-explicit 4 = 实证不可达双版本(已 boundary 记)。
- 域分布查询:`pwsh -c "(gc docs/capabilities.json -raw|ConvertFrom-Json)|?{$_.cap.verify -eq 'roundtrip'}|group {$_.cap.domain}"`。

需求驱动残项(arc 外,按需):能力查询 `go run ./cmd/capindex -q <词>`。

**剩余 backlog（非阻塞,需求驱动;两个能力 roadmap 已全收口归档→ `archive/specs/`,真相源=capindex）**：
- effects：**全收口**（wave 11 layer-ref + wave 12 收回 wave-6 parked 经典 9 个:BEZMESH/MESH WARP/CHANNEL MIXER/RESHAPE/Vector Paint/Texturize/Color Link/Compound Arithmetic/Set Channels,库 216）。残项=modal-hang(Apply Color LUT/PS Arbitrary Map/Numbers,无法无人值守加)+ Vegas/Warp 等 AE2020 真不可用——**均不可达**(实证 canAdd=false/弹框)。
- mask：**域全收口** —— 结构性 op(Add/Remove/Dup/Move) + 选项(feather/opacity/expansion) + 路径(静/动含 >4kf 容量页) + maskFeatherFalloff(2026-06-17 RE mkif @0x03,翻案"不可达",双版本 gate)。**无残项**。
- expr：**全闭环**。SetExpression 字节机制双版本 gated(5 类性质各异 idiom:静态/跨层/带关键帧/时变/读 effect 参数,穷尽写入端字节情况)。`linear()/ease()/valueAtTime` 等未单独 gate = 验证洁癖残项**非能力缺口**——机制已证内容无关,风险在 AE 求值端非写入端,**不补**(closed decision,理由+若补注意点见 `incidents/expression-enable-byte-pair.md`)。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* **基本收尾**（OutputModule 实质字段补齐到 OutputAudio/ConvertToLinear,Alpha;剩 ColorSpaceWorking=CMS 假绿风险 · PostRenderCompID=结构性引用 · Rs runtime 字段=不该 setter,均不做）· capindex 收尾小项（render-queue tag cosmetic · 11 manual-gate orphan 恢复 ae-accept,需跑 AE）。

## Backlog

- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
