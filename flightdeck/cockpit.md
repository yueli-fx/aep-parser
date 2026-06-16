# Cockpit — aep-parser

**Last updated**: 2026-06-16 by claude（**知识库地基重建 arc 全部 DONE**:capindex P1+P2(写面 468/468=100% + CI 强制)+ knowledge-consolidation E/D/C/A/B 全收口。本轮收尾:**A** ~20 auto-memory 迁 flightdeck/CLAUDE.md 后清空 memory 目录 + CLAUDE.md 加"不用 auto-memory"规则;**B** CLAUDE.md #2/#3 长解释下沉 spec、留短规则+指针(硬约束 #1-7 全在);flame 番外 lesson 入 `incidents/procedural-fx-over-vector`。全套 go build/test 绿。~28 commits 4ae1358..(本轮)。**arc 完结 → 回归需求驱动库工作**。）

**Active focus**: **需求驱动稳态**（知识库地基重建 arc 已完结）。**知识单一家 = flightdeck + CLAUDE.md**(auto-memory 已退役);**能力真相源 = capindex**(`go run ./cmd/capindex -q <词>` / `docs/capabilities.{json,md}`,CI 强制写/做面零漏标;getter/const 豁免)。coverage.md=残值(暂搁/不可达/negative)、coverage-detail=AE-attribute 参考矩阵。—— **库能力主线**早已全收口、需求驱动(roadmap 1-6 + Text Animators 全 ship,详 `specs/2026-06-14-remaining-capability-roadmap.md`);机制库 parse-the-clone + synthesis-insert + animate;每渲染类双版本 AE ship-gate(红线4)。火焰=番外(程序化效果链,见 incident)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [2026-06-16-capability-index.md](specs/2026-06-16-capability-index.md) — 源码内 aep:cap 结构化 tag → cmd/capindex 自动生成可秒查的能力+API 索引(JSON+人读表) + version(minver) 支持,CI 防漂移;取代手写易过期的 coverage.md
<!-- /AUTO -->

## 下一步

**➡ 知识库地基重建 arc 已完结(capindex P1+P2 + knowledge-consolidation E/D/C/A/B 全 done)。回归需求驱动库工作。**

**可选收尾小项(非阻塞)**:
- capindex:render-queue tag directive-first→END(cosmetic,功能等价、go/doc 已 strip)· orphan(11 manual-gate op 如 DeleteLayer 编码 JSX gate 为 Go test → 恢复 ae-accept,需跑 AE)。
- landing housekeeping:capability-index spec(graduate)可 graduate-to-docs;knowledge-consolidation(done,一次性整合)可归档。
**库能力 backlog**(需求驱动,见下)· **能力查询**:`go run ./cmd/capindex -q <词>`。

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
