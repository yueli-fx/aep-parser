# Cockpit — aep-parser

**Last updated**: 2026-06-15 by claude（优先级2 **3D transform 全闭环**：Z 视差 `TestLayer3DParallax`〔NEAR 300/FAR 99，3.03×〕+ RotateY 透视 `TestLayer3DRotateY`〔梯形 1.45×〕双版本渲染 PASS。**关键发现**：整个 3D transform group 从零写入零新 serializer 代码——嵌入 transform 模板本就是完整 6-axis 3D schema，`SetIs3D`+既有 setter 即可。先前同日：3D enable+相机推拉 gate · 优先级1 animated gradient 色标〔用户造 fixture 解锁〕+ animated Trim。逐 commit 见 git log。）

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

**优先级2 3D 图层（主体闭环 — 解锁真·拉镜）**：
- ~~图层 3D flag（enable）~~ ✅ `TestLayer3DEnable`（DOM）+ ~~相机推拉~~ ✅ `TestLayer3DCamDolly`（渲染 426→124px）。
- ~~Z 视差~~ ✅ `TestLayer3DParallax`（两层不同 Z，NEAR 300 / FAR 99，3.03×）。
- ~~RotateY 透视 tumble~~ ✅ `TestLayer3DRotateY`（梯形 1.45×）。
- **关键发现**：整个 3D transform group（enable+Z+旋转+朝向）**从零零新 serializer 代码**——`SetIs3D`+reopen 后既有 transform setter，因嵌入 transform 模板本就是完整 6-axis 3D schema。详 `incidents/layer-3d-enable-bit-materializes.md`。
- **下一项（按需）**：Material Options 从零 3D 层 gate（fixture setter 已工作）· RotateX/Orientation/RotateZ 补 gate（同路径）· 3D-render DoF/光照像素深验 · 推拉镜+视差 showcase（需用户真机验收才能标 complete）。

**用户真机验收待办**：本批次 6 个能力（animated trim/gradient + 3D enable/dolly/parallax/rotateY）均 agent 眼验，**未经用户真机复核**——攒 showcase 时请用户过目。

完整清单（每层细项 + 不可达附录 + 搁置项）见 roadmap spec。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
