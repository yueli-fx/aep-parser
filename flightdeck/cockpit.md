# Cockpit — aep-parser

**Last updated**: 2026-06-15 by claude（effects 深化 arc 全收口：跨效果 enum gate · `AnimateEffectParamVec`〔color/point〕· `SetEffectLayerParam`+Set Matte · 库 31→41；expr 视为完成。归档 done showcase spec，精简看板。）

**Active focus**: **库进入需求驱动稳态**——剩余能力 roadmap 主体（优先级 1-5：动画关键帧 / 3D / 形状 / mask / 表达式）+ 表达式·效果深化 arc 全收口（详 `specs/2026-06-14-remaining-capability-roadmap.md` + git log）。**唯一真剩余主线 = 文字多 run/段落**；其余皆「按需 / 不可达」。机制库：parse-the-clone + synthesis-insert + animate(Gradient/Vector/Path)。每渲染/可见类双版本 AE ship-gate（红线4，`scripts/ae_run.ps1` 自助）。基本图形搁置。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**➡ 下一阶段 = 文字多 run / 段落**（用户既定顺序，effects 之后）：多 paragraph / 多 run / 带 kerning / 空串改字——当前 refuse，需 btdk 段落·run entry splicing（未 RE，effort 高）= 完整文字动画接入。文字基础（NewTextLayer + 单段单 run SetText）已够用。详 `incidents/text-btdk-length-variable-write-scoping.md`。

**按需 backlog（非阻塞，需求驱动）**：
- shape：Gradient stroke 嵌套组 Dashes/Taper/Wave · Offset Line Join/Miter/Copy Offset · Blend Mode/Composite Order render-gate · ZigZag/Twist 等 elided 子流（synthesis-insert 蓝本现成，`trim-paths-vector-filter-re.md`）
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
