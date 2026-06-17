# Cockpit — aep-parser

**Last updated**: 2026-06-17 by claude（本会话连推：effects wave 11 layer-ref + wave 12 收回 wave-6 parked 经典(库 203→216,全大写 match-name) · mask 域全收口(>4kf 容量分页 6kf + maskFeatherFalloff mkif @0x03 翻案"不可达")。均双版本 gate。另删 deferred-backlog spec。· expr backlog 撤项（按函数补 gate = closed decision，机制已闭环）。· 启动 roundtrip→ae-accept 补验 arc，批1 layer-set 7 flag 双版本 gate。）

**Active focus**: **需求驱动稳态**（知识库地基重建 arc 已完结）。**知识单一家 = flightdeck + CLAUDE.md**(auto-memory 已退役);**能力真相源 = capindex**(`go run ./cmd/capindex -q <词>` / `docs/capabilities.{json,md}`,CI 强制写/做面零漏标)。库能力主线早已全收口、需求驱动;机制库 parse-the-clone + synthesis-insert + animate;每渲染类双版本 AE ship-gate(红线4)。两个能力 roadmap 已归档(`archive/specs/`),残项见下 backlog,真相源=capindex。火焰=番外(见 `incidents/procedural-fx-over-vector.md`)。

## 进行中

<!-- AUTO:inprogress -->

<!-- /AUTO -->

## 下一步

**➡ ACTIVE ARC:roundtrip→ae-accept 补验**(2026-06-17 起)。库 481 能力里 **272 个仅 verify=roundtrip**(只 Go 自读回,从没让真 AE 消化——红线1 假绿温床)。系统性按域做综合 fixture 批量补真 AE 验,把验证等级整体抬一档。**诚实约束**:能 JSX DOM readback 的→值验;读不回的(尤其 render-queue 40 个二进制专属字段,rq-comment incident)→上限只能 acceptance-preservation。
- **批1 ✅ DONE**(commit 5e98762):layer-set 7 个 AV-flag setter(Visible/Shy/Solo/Locked/MotionBlur/Quality/BlendingMode)双版本 gate;`ae-accept` 35→42。
- **下一批候选**:layer-set 剩余可值验项(InPoint/OutPoint/StartTime/Stretch/Parent/Label/AutoOrient…)→ comp 37 → text 33 → shape 23(渲染像素)→ keyframe/mask → project/meta(部分 acceptance)→ render-queue 40(多数仅 acceptance)。
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
