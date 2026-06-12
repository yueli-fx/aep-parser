# Cockpit — aep-parser

**Last updated**: 2026-06-12 by claude（MG roadmap S3 Trim Paths 收口：矢量滤镜节点套既有 embed-body vein，双版本渲染像素 gate PASS（FULL 整圈 vs HALF 右半弧·left 缺）。详 `incidents/trim-paths-vector-filter-re.md`。同日前序：S1 ease + S2 表达式激活。）
**Active focus**: **From-scratch MG 工程能力**（`specs/2026-06-12-from-scratch-mg-roadmap.md`，用户终极目标 = 不开 AE 纯 Go 生成完整 MG 动画）。S1 ease ✅ · S2 表达式 ✅ · S3 Trim Paths ✅ · 下一刀 S4 precomp 嵌套 → S5 Repeater 等 → S6 端到端组合 gate。每 slice 渲染像素级双版本 gate（红线4）。ship-gate 自助（`scripts/ae_run.ps1` 无人值守）。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**主线 = MG roadmap**（`specs/2026-06-12-from-scratch-mg-roadmap.md`；S1 ease ✅ S2 表达式 ✅ S3 Trim ✅）：

1. **S4 precomp 嵌套**（comp-as-layer-source，工程结构刚需）：RE precomp 层 ldta source-ID 指向 comp item + sibling（对照 JSX `comp.layers.add(otherComp)` fixture）→ 父 comp 渲染像素能看到子 comp 内容的 gate。
2. **S5 Repeater / gradient Start·End Pt / Rounded Corners**（Repeater/Merge/Offset/Round/ZigZag 同类矢量滤镜，复用 S3 trim 的 embed-body vein 蓝本，详 `incidents/trim-paths-vector-filter-re.md`）→ **S6 端到端组合 gate**（前置：S2 followup 表达式语汇 gate loopOut/wiggle/跨层引用）。
3. 技债（roadmap 路上顺修）：`encodePathTimeTable` 容量分页同病（path >4 kf 前必修，详 `incidents/lhd3-keyframe-capacity-pages.md`）· animated trim（line-draw 真动画）路径已通未单独 gate · 种子模板 32bpc→8bpc 评估。

**需求驱动候选**（点名即开工）：encodeBezier AE-native 字节 · EG W deferred 控件 · mask 剩余写功能（SetMaskPath / animated path / mode·color·feather 参数化）· SetEffectParam 扩库 · RQ Set\* slice-5/6/7/8（Alpha by design）。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
