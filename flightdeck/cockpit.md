# Cockpit — aep-parser

**Last updated**: 2026-06-16 by claude（**capindex P2 wave0+1**：抽取扩到 facade+scene 两包(揭示真实公共能力面 = **536 符号**,此前只见 facade、漏 341 个 `Set*`+全 getter);加 meta 通道 + `docs/docgen.json` 公共面门禁 + `-coverage` 仪表。**facade 全量自由函数已标 + ship-gate 交叉核实 = 98/536(18.3%)**。审核揪出 **manual-gate orphans**(DeleteLayer/MoveLayer/DuplicateComposition 等 11 个 doc 称 Stable 但无自动 Go `_AEShipGate` test → 诚实标 alpha/roundtrip + boundary,待决:编码 JSX gate)。commits:wave0 + db96367(wave1)。**余 438 scene 方法** = wave2+(pace/depth 待用户定)。）

**Active focus**: **知识库地基重建 arc（当前主线）**——三原则:源码为准/自动化/秒查。capindex(`specs/2026-06-16-capability-index.md`):源码内 `aep:cap` tag 自动生成可秒查能力/API 索引,**P1 done**(layer-create) → **P2 全量标注**(= 源码级能力审核,揪假绿/补缺口)→ CI 全覆盖强制 + coverage 退役。然后 `knowledge-consolidation`(parked,待 P2):记忆系统退役/incidents 合并清理/CLAUDE.md 瘦身/根目录 exe 清理。**能力查询**:`go run ./cmd/capindex -q <词>` 或 grep `docs/capabilities.json`(渐取代 coverage.md)。—— **库能力主线**早已全收口、需求驱动稳态(roadmap 1-6 + Text Animators 全 ship,详 `specs/2026-06-14-remaining-capability-roadmap.md`);机制库 parse-the-clone + synthesis-insert + animate(Scalar/Vector/Gradient/Path/TextRange);每渲染类双版本 AE ship-gate(红线4)。火焰演示=番外。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [2026-06-16-capability-index.md](specs/2026-06-16-capability-index.md) — 源码内 aep:cap 结构化 tag → cmd/capindex 自动生成可秒查的能力+API 索引(JSON+人读表) + version(minver) 支持,CI 防漂移;取代手写易过期的 coverage.md
- [2026-06-16-capability-index-p2.md](plans/2026-06-16-capability-index-p2.md) — 扩 capindex 抽取到 docgen 三包(facade+scene+codec),给全量公共能力符号(Set*/getter/结构性/Add*)打 aep:cap tag、逐个对 ship-gate 交叉核实(揪假绿/标缺口),最后开 CI 全覆盖强制
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**➡ capindex P2 wave2+:438 个 scene 方法标注审核**（Layer 229 / Comp 58 / Project 52 / Property 32 / Mask·Footage…)。**pace/depth 待用户定**:solo 多会话 vs Workflow 并行(成本/opt-in);全量含 getter vs 写/做面优先。配方见 `plans/2026-06-16-capability-index-p2.md` wave2-9。**orphan 决议**(11 个 manual-gate op):编码 JSX gate 为 Go test 恢复 ae-accept,还是维持 roundtrip。
**Wave 9 收尾**:meta 扫尾 + 开 CI 全覆盖强制 + 回写 spec。**然后** `knowledge-consolidation`(parked):记忆退役/incidents 合并/CLAUDE.md 瘦身/根 exe 清理。
**能力查询**:`go run ./cmd/capindex -q <词>`(facade 全覆盖,438 scene 方法陆续上)。

**原库能力 backlog（非阻塞,需求驱动;火焰=番外搁置）**：
- 文字：动画器 5 类全 ✓ + **animate leaf 全收口** ✓ + **免费近邻收口** ✓（2026-06-16：~~Fill Opacity~~ ~~Stroke Opacity~~ ~~Stroke Width~~ ~~Stroke Color~~ ~~Skew~~ 5 个双版本 render-gate PASS；**Rotation X/Y evidence-based defer**——2D 层视觉惰性 bbox 三帧全同，需逐字 3D，facade 保留标 Alpha/write-only + round-trip 自验）；**structural op（Remove/Dup/Move）双版本 gate PASS** ✓ + **Range Advanced（SetTextRangeAdvanced，Amount render-gate PASS）** ✓ + **多 Selector（AddTextRangeSelector）** ✓ + **Wiggly Selector（AddTextWigglySelector，双版本 render-gate PASS via 时间变化签名）** ✓（2026-06-16）；**selector 家族收口**。**Expressible Selector = evidence-defer**（Amount 表达式驱动，库表达式未渲染验证）。详 `incidents/text-animator-create-re.md`
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
