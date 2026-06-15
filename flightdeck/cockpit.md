# Cockpit — aep-parser

**Last updated**: 2026-06-16 by claude（**文字 Range Advanced**：`SetTextRangeAdvanced` 一次设全 10 个 Selector 高级子参数（Units/BasedOn/Mode/Amount/Shape/Smoothness/Ease/Randomize/Seed，全 materialize 模板 replace elided 组 + 逐 cdat 覆写）。**Amount 双版本 render-gate PASS**（A/B 差分单帧：Amount=100 隐 spread 0 / Amount=20 显 spread 159，AE2020≡AE2025），余 9 round-trip 验证 + evidence-defer render gate（Mode 需多 Selector·Smoothness 仅 Square·余改选区非亮度）。gotcha：enum set 失效 live ref；Smoothness Shape≠Square UI 隐藏但 slot 仍持久化。commit a6f82a9。—— 同日：**文字 structural op gate**：Remove/MoveTo/Duplicate 双版本 ship-gate 6/6 PASS（`text_animator_struct_shipgate_test.go`，3 动画器 × 3 op × AE2020+2025，readback 顺序逐项匹配），Alpha→shipped；纯 gate 无 impl 改（复用 Effect Parade 同机制）。gotcha：AE ScriptingAPI 枚举完整 leaf schema，verify 按非默认值辨识 driven leaf。commit d47bbe7。—— 同日早些：**文字免费近邻收口 arc**：7 个 Add\*Animator over 既有 splice 引擎（抽模板 + 一行 facade），match-name 全 RE 确认（vtype 6417 1D / 6418 color）。**5 个双版本 render-gate PASS**（AE2020≡AE2025 逐数字：Fill Opacity spread 0→199 · Stroke Opacity area 0→4362 · Stroke Width 10797→1780 · Stroke Color green 0→243 · Skew aspect 2.00→0.65；表驱动 + 通用 verify jsx，每 leaf 一作用面签名）。**Rotation X/Y evidence-based defer**（2D 视觉惰性 bbox 三帧全同 → 需逐字 3D；facade 保留标 Alpha/write-only + `TestTextRotationXY_RoundTrip`）。Go 抽 2 私有 helper。commit 9585bf0。全量 suite 绿。）—— 上一条：**shape elided 子属性收口 arc**：synthesis-insert 蓝本从 2 次扩到 **8 次全绿**，覆盖三类 leaf（scalar/enum/Vec2）+ 多 leaf canonical 排序 + splice-before-group。新 ship 4 slice（每个双版本渲染像素 gate PASS）：**ZigZag Points**(enum Smooth/Corner，flat 22/65)·**Twist Center**(首个 Vec2 leaf，质心 disp 0.8/119)·**Offset Line Join/Miter Limit/Copy Offset**(5 卡一帧，Offset 5 子流全收齐)·**Repeater Order**(splice-before-group；纠正 Explore「不可达」误判=可达；同色合成 commutative 故视觉无效，实证 BELOW=ABOVE 白覆盖 7341==7341)。**Wiggle 调制参数(Correlation/Temporal·Spatial Phase/Roughen Points)** 评估后 evidence-based defer（Phase=噪声 reseed 本质不可门禁；Correlation/Points 低价值噪声调制，机制已证）。commits d1425a1/cd3aa9c/df488dc/5b469fe。全量 suite 绿。）

**Active focus**: **库主线全收口、进入需求驱动稳态**——剩余能力 roadmap 优先级 1-6（动画关键帧 / 3D / 形状 / mask / 表达式 / **文字**）+ 表达式·效果深化 arc + Text Animators 全 ship（详 `specs/2026-06-14-remaining-capability-roadmap.md` + git log）。**无剩余主线**；其余皆「按需 / 不可达」。机制库：parse-the-clone + synthesis-insert + animate(Scalar/Vector/Gradient/Path/TextRange)。每渲染/可见类双版本 AE ship-gate（红线4，`scripts/ae_run.ps1` 自助）。基本图形搁置。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**➡ 主线全 ship，无既定下一阶段——纯需求驱动。** 用户原定 expr→effects→文字 全收口（文字 = 基础 v4 + Text Animators 引擎，2026-06-15 双版本 gate PASS）。下面是按需 backlog，等用户点名或新需求。

**按需 backlog（非阻塞，需求驱动）**：
- 文字：动画器 5 类全 ✓ + **animate leaf 全收口** ✓ + **免费近邻收口** ✓（2026-06-16：~~Fill Opacity~~ ~~Stroke Opacity~~ ~~Stroke Width~~ ~~Stroke Color~~ ~~Skew~~ 5 个双版本 render-gate PASS；**Rotation X/Y evidence-based defer**——2D 层视觉惰性 bbox 三帧全同，需逐字 3D，facade 保留标 Alpha/write-only + round-trip 自验）；**structural op（Remove/Dup/Move）双版本 gate PASS** ✓ + **Range Advanced（10 子参数 SetTextRangeAdvanced，Amount 双版本 render-gate PASS，余 9 round-trip）** ✓（2026-06-16）；剩 多 Selector · Wiggly/Expression Selector。详 `incidents/text-animator-create-re.md`
- shape：**elided 子属性收口 arc 已完成**（2026-06-15）——ZigZag Points · Twist Center · Offset Line Join/Miter/Copy Offset · Repeater Order 全双版本 render-gate PASS（synthesis-insert 8×，Offset 5 子流全收齐）；Gradient stroke Dashes/Taper/Wave + Blend Mode/Composite Order 早已 ship（曾是看板 drift）。**仅剩** Wiggle Paths/Transform 调制参数（Correlation/Temporal·Spatial Phase/Roughen Points）evidence-based defer（Phase 本质不可像素门禁、Correlation/Points 低价值，机制已证、随时可做）。详 `trim-paths-vector-filter-re.md`
- mask：maskFeatherFalloff（位置未 RE，可能不可达）
- effects：per-effect typed param helper · 库继续扩 · Displacement Map/Compound Blur 等 layer-ref（同 Set Matte 机制）
- expr：linear()/ease()/valueAtTime remap（内容无关已证，边际低）
- 3D：RotateX/Orientation/RotateZ 补 gate（同路径，低优先）

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
