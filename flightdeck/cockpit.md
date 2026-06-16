# Cockpit — aep-parser

**Last updated**: 2026-06-17 by claude（**text Expressible Selector 收口**(text selector 家族全完):`AddTextExpressibleSelector(layer,amountExpr)` 双版本 render-gate(空间差分:L 隐 ABCD 渲 EFGH/R 隐 EFGH 渲 ABCD)。原 defer「库表达式未验证」翻案(S2 早证)。RE 坑:Amount 表达式专属(无静态值、`.value` 抛)、默认全 elided 须先在 fixture 设表达式 materialize 模板。前轮:**音频效果波收口**:+10 个音频处理效果入库(库 193→203),双版本 ae-accept ship-gate(`TestAddEffectAudio`)。音频效果只能挂有音频的层,故用用户给的 mp3 import 成音频层 RE+gate(fixture gitignored)。坑:同效果加两次第二个被 AE elide(686B vs 1502B),fixture 每效果只加一次;Flange match-name 是 `ADBE Aud_Flange`(下划线)。前一轮:**3D priority-2 真正收口**:RotateX/Orientation/RotateZ 双版本 render-pixel gate(`layer_3d_rotaxes_shipgate_test.go`,6 个全 PASS)。建 Orientation gate 时揪出真 correctness bug——3D Orientation 静态值在 otst 里存两份(cdat 小端 + otda 大端),AE 读 otda;旧 SetStaticValue 写 cdat 用了大端(字节翻转)且没碰 otda → 值 round-trip 绿但 AE 渲染 0(红线4 活样本)。修=cdatLE+otda 双写。SetRotateX/Y/Rotation/Orientation 升 render-pixel。commit e9a471e。）

**Active focus**: **需求驱动稳态**（知识库地基重建 arc 已完结）。**知识单一家 = flightdeck + CLAUDE.md**(auto-memory 已退役);**能力真相源 = capindex**(`go run ./cmd/capindex -q <词>` / `docs/capabilities.{json,md}`,CI 强制写/做面零漏标)。库能力主线早已全收口、需求驱动;机制库 parse-the-clone + synthesis-insert + animate;每渲染类双版本 AE ship-gate(红线4)。剩余 backlog 详 `specs/2026-06-14-remaining-capability-roadmap.md`。火焰=番外(见 `incidents/procedural-fx-over-vector.md`)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
<!-- /AUTO -->

## 下一步

**➡ 需求驱动稳态——无 active arc。等新需求,或从下列残项挑。** 能力查询:`go run ./cmd/capindex -q <词>`。

**剩余 backlog（非阻塞,需求驱动;全详 `specs/2026-06-14-remaining-capability-roadmap.md`）**：
- effects：~~音频效果波~~ ✅(2026-06-17,+10 个 Backwards/BT/Delay/Flange/HiLo/Modulator/EQ/Reverb/StereoMixer/Tone,库 193→203,双版本 ae-accept gate,详 incident wave10)· layer-ref 第二波(+4,按 Displacement Map 物化流程)。
- ~~text：Expressible Selector~~ ✅(2026-06-17,`AddTextExpressibleSelector(layer,amountExpr)` 双版本 render-gate;原 defer 翻案——表达式早已渲染验证。**text selector 家族全收口**)。
- mask：maskFeatherFalloff（位置未 RE,可能不可达）· expr：linear()/ease() remap（边际低）。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）· capindex 收尾小项（render-queue tag cosmetic · 11 manual-gate orphan 恢复 ae-accept,需跑 AE）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
