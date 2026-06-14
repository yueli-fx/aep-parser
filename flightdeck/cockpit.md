# Cockpit — aep-parser

**Last updated**: 2026-06-15 by claude（优先级1 **animated gradient 色标**闭环：用户手工造带关键帧色标 fixture（JSX 无法 authoring）解锁 RE，animated 布局=时间表(bpk=64)+每帧 Utf8；`AddGradientKeyframe`，kf0=R/B/G→kf1=G/R/B 双版本实渲交换 PASS+肉眼验。先前同日：优先级2 **3D flag 闭环**〔`SetIs3D`→AE 自动 materialize 3D 层，DOM gate + 相机推拉渲染 gate 双 PASS〕；优先级1 **animated Trim** reveal gate 双版本 PASS。逐 commit 见 git log。）

**Active focus**: **剩余能力 roadmap（模板起点，非 from-scratch）**（`specs/2026-06-14-remaining-capability-roadmap.md`）。MG from-scratch 主线 + 质感件 + camera/light option + 优先级1 animated 关键帧（trim✅）+ 优先级2 **3D-enable 双 gate 闭环**。现推进优先级2 余项：3D transform 通道**非默认值写入**（Z/旋转）。机制复用 parse-the-clone + synthesis-insert。每能力渲染/可见类双版本 ship-gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1`）。基本图形搁置。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**主线 = `specs/2026-06-14-remaining-capability-roadmap.md`**（模板起点，非 from-scratch）。优先级顺序（用户 2026-06-14 定）：**动画关键帧 > 3D 图层 > 形状剩余 > mask > 表达式 > 文字图层**。

**优先级1 动画关键帧**（基本收口）：
- ~~lhd3 keyframe 容量分页~~ ✅ · ~~temporal ease 普及~~ ✅（2026-06-14）。
- ~~animated 矢量滤镜（Trim End reveal）~~ ✅ 2026-06-15（三帧渲染 gate 双版本 PASS；Repeater/Offset 同 `lowerShapeScalar` 路径，按需补）。
- ~~animated gradient 色标~~ ✅ 2026-06-15（用户手工造 fixture 解锁 RE；`AddGradientKeyframe` + 时间表/多 Utf8 emit；kf0=R/B/G→kf1=G/R/B 双版本实渲交换 PASS）。deferred：gradient STROKE 色标动画 + 色标 ease。
- 遗留：path 几何 lhd3 >4 顶点分页（需求驱动）。

**当前层 = 优先级2 3D 图层**：
- ~~图层 3D flag（enable）~~ ✅ 2026-06-15。`SetIs3D` 翻 bit → AE 自动 materialize 完整 3D 层。DOM gate `TestLayer3DEnable` + 渲染 gate `TestLayer3DCamDolly`（相机推拉缩放）双版本 PASS。详 `incidents/layer-3d-enable-bit-materializes.md`。
- **下一项 = 3D transform 通道非默认值写入**（Z position / Orientation / Rotate X·Y）。enable 时 AE 自动建通道但写非默认值需 slot：3-comp Position 非 length-preserving。三路待选：(a) Is3D 时 emit 3-comp Position；(b) synthesis-insert 旋转通道（camera/light-option vein）；(c) parse-the-clone AE-materialized 3D 层。解锁**视差**（两层不同 Z）+ RotateY 透视 + 从零 3D 层的 Material Options。
- 后续：Material Options（fixture setter 已工作，缺从零 3D 层上的 gate）· 3D-render 像素深验 DoF/光照 · 推拉镜+视差 showcase。

完整清单（每层细项 + 不可达附录 + 搁置项）见 roadmap spec。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
