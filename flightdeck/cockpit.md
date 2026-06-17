# Cockpit — aep-parser

**Last updated**: 2026-06-17 by claude（本会话连推：effects wave 11 layer-ref + wave 12 收回 wave-6 parked 经典(库 203→216,全大写 match-name) · mask 域全收口(>4kf 容量分页 6kf + maskFeatherFalloff mkif @0x03 翻案"不可达")。均双版本 gate。另删 deferred-backlog spec。· expr backlog 撤项（按函数补 gate = closed decision，机制已闭环）。· 启动 roundtrip→ae-accept 补验 arc：layer-set 批1-3 共 17 setter 升 ae-accept(35→53)，RE 修好 SetComment from-scratch 假绿(ldta @0x3C has-comment flag)，立跨对话 drain 清单 plan。）

**Active focus**: **roundtrip→ae-accept 补验 arc**(2026-06-17 起;详 `## 下一步` + `plans/2026-06-17-roundtrip-ae-accept-backfill.md`)。库能力主线早已全收口,此 arc 把"标 stable 但只 Go round-trip、从没让真 AE 消化"的写能力按域补双版本 AE 验(验证洁癖洞)。已推进至批7b(ae-accept 35→**118**,roundtrip→**196** 跌破 200):layer-set 主体(批1-3)+ camera/light 17(批4)+ classic material 8(批5a)+ Composition 22(批6)+ **text SetRun* style 13(批7)+ SetParagraph indent 5(批7b)**。修 3 个真 bug:SetComment 假绿(RE)· SetFrameRate+SetDuration(incident)· **btdk FormatPSReal**(tsume + 5 段落,复发 incident `btdk-point-value-needs-formatpsreal`)。**不变量**:知识单一家 = flightdeck + CLAUDE.md(auto-memory 已退役);能力真相源 = capindex(`go run ./cmd/capindex -q <词>`,CI 强制零漏标);每渲染类双版本 AE ship-gate(红线4)。火焰=番外(`incidents/procedural-fx-over-vector.md`)。

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
- **批7/7b ✅**(13 SetRun* style + 5 SetParagraph indent 双版本;修 tsume + 5 段落 FormatPSReal bug)。残 text:SetRun* enum/index(FontIndex/DigitSet/AutoKernType/BaselineOption/LineJoinType/NoBreak)· SetParagraph bool/enum 4(AutoHyphenate/LeadingType/HangingRoman/Direction)· ManualKerning/AddFont/animator。多可仿批7/7b 扩 gate(whole-doc DOM 或 resave-preservation)。
- **下一批 = shape 23(渲染像素红线4,大块)** 或 text 残扩 gate → keyframe/mask → marker(SetChapter/URL/CuePoint 归错 comp domain)→ project/meta → render-queue 40(多 acceptance)。layer-set bool 残 + 6b(comp idta) 按需。
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
