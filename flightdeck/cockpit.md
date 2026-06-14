# Cockpit — aep-parser

**Last updated**: 2026-06-14 by claude（立 `remaining-capability-roadmap` spec：把未做/半做/不可达能力按用户优先级〔动画关键帧>3D>形状剩余>mask>表达式>文字〕集中归档，模板起点非 from-scratch，基本图形搁置。先前同日：camera/light option 全清 + 双版本 gate〔Camera 13/13 + Light 10/10〕。逐 commit 见 git log。）

**Active focus**: **剩余能力 roadmap（模板起点，非 from-scratch）**（`specs/2026-06-14-remaining-capability-roadmap.md`）。MG from-scratch 主线 + 质感件 + camera/light option 全字段已闭环；现转入按优先级补齐剩余能力：**动画关键帧 > 3D 图层 > 形状剩余 > mask > 表达式 > 文字**。起点是现有模板/真实 .aep（绕开 from-scratch silent-drop），机制复用 parse-the-clone + synthesis-insert。每能力渲染/可见类双版本 ship-gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1`）。基本图形搁置。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**主线 = `specs/2026-06-14-remaining-capability-roadmap.md`**（模板起点，非 from-scratch）。优先级顺序（用户 2026-06-14 定）：**动画关键帧 > 3D 图层 > 形状剩余 > mask > 表达式 > 文字图层**。

**当前层 = 优先级1 动画关键帧**，第一项硬阻塞：
- **lhd3 keyframe 容量分页**（`encodePathTimeTable`/`encodeKeyframes` 的 lhd3 @0x0C/@0x1C page 容量字段）——属性 >4kf 时 AE 2025 判损坏，**任何上规模动画前必修**。详 `incidents/lhd3-keyframe-capacity-pages.md`。
- 同层后续：temporal ease 普及 · animated trim/repeater · animated gradient 色标。

完整清单（每层细项 + 不可达附录 + 搁置项）见 roadmap spec。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
