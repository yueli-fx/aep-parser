# Cockpit — aep-parser

**Last updated**: 2026-06-16 by claude（**capindex P2 DONE(写/做面 468/468=100%,CI 强制)→ 转 knowledge-consolidation,E+D 落地**。capindex:8-agent Workflow 标 scene 写面 + write-surface-first(getter/const 豁免)+ tier/verify 解耦 + 写面定义扩面(shape 节点/render-queue/text-run)。verify 诚实分布 roundtrip280/render-pixel149/ae-accept35/meta6。**knowledge-consolidation**(active):**E** 删 12 个根目录 stray exe;**D** coverage 退役——coverage.md 415→75 行(写能力快照→capindex 指针,保留 暂搁/不可达/negative 残值,status=superseded)+ coverage-detail.md→reference + CLAUDE.md 改指 capindex。commits 4ae1358..d07bc56。）

**Active focus**: **知识库地基重建 arc（当前主线）**——三原则:源码为准/自动化/秒查。capindex(`specs/2026-06-16-capability-index.md`,graduate)**P1+P2 DONE**:源码内 `aep:cap` tag → 自动生成可秒查能力/API 索引,**写/做面 100% 标注 + CI 强制零漏标**。getter/const 按 write-surface-first 豁免。**下一步 = `knowledge-consolidation`**(parked→可启动):coverage.md 退役(写面真相源已迁入 tag,已解锁)/ incidents 合并清理 / CLAUDE.md 瘦身 / 记忆系统退役 / 根目录 exe 清理。**能力查询**:`go run ./cmd/capindex -q <词>` 或 grep `docs/capabilities.json`(取代 coverage.md)。—— **库能力主线**早已全收口、需求驱动稳态(roadmap 1-6 + Text Animators 全 ship,详 `specs/2026-06-14-remaining-capability-roadmap.md`);每渲染类双版本 AE ship-gate(红线4)。火焰演示=番外。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [2026-06-16-capability-index.md](specs/2026-06-16-capability-index.md) — 源码内 aep:cap 结构化 tag → cmd/capindex 自动生成可秒查的能力+API 索引(JSON+人读表) + version(minver) 支持,CI 防漂移;取代手写易过期的 coverage.md
- [knowledge-consolidation.md](specs/knowledge-consolidation.md) — 退役 auto-memory 系统(迁入 flightdeck/CLAUDE.md)+ incidents 过时清理与合并减量 + coverage.md/coverage-detail.md 退役(待 capindex 完成)+ CLAUDE.md 瘦身 + 根目录 stray exe 清理;目标=单一知识家 flightdeck + 精简 CLAUDE.md + 干净仓库
<!-- /AUTO -->

## 下一步

**➡ `knowledge-consolidation`(active)续做**(spec 字母轴):**E ✅**(exe 清理)· **D ✅**(coverage 退役)。剩:
- **A 记忆退役**:~20 条 auto-memory 按 A1 映射表迁入 flightdeck/CLAUDE.md(多条已重复→删),memory 目录清空 + CLAUDE.md 加"不用 auto-memory"规则。⚠ **memory 在 repo 外、删除不可 git 回滚 → 执行前须用户确认**(spec 风险条)。
- **C incidents 审核**(51 个):过时清理(grep 代码核实,board-status-drift)+ 同根因合并成多-Case incident + 对齐 capindex `incident=` 反链。**量大,workflow 候选**(同 capindex 并行模式)。
- **B CLAUDE.md 全量瘦身**:#2/#3 长 RE 解释下沉 references/incident,留短规则+指针(部分已做:能力源指针)。+ 写防复发约定(`go build -o tmp_debug/bin/`)。
**capindex 遗留小项**:render-queue tag directive-first→END(cosmetic,功能等价)· orphan(11 manual-gate op 编码 JSX gate 为 Go test 恢复 ae-accept)。
**能力查询**:`go run ./cmd/capindex -q <词>`。

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
