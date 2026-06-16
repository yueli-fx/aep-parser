# Cockpit — aep-parser

**Last updated**: 2026-06-16 by claude（**capindex P2 DONE — 写/做面全标注 468/468 = 100%**(`docs/capabilities.{json,md}` 470 条)。8-agent Workflow(581k token)并行标 scene 写面 + 收尾。**write-surface-first**(用户批准):getter/reader/const **豁免**,只强制写/做面;`TestWriteSurfaceFullyTagged` CI 锁定零漏标。verify 诚实分布:roundtrip 280(length-preserving)/ render-pixel 149(实测采样像素 gate)/ ae-accept 35 / meta 6。审核修正:**tier/verify 解耦**(SetOpacity=stable+roundtrip 合法)· manual-gate orphans 诚实标 · **写面定义扩面**(纳入 shape 节点全族/render-queue/text-run,此前漏)。`-q "蒙版/位置/表达式/色彩管理"` 秒答。commits 4ae1358..0cd4626(~13 个)。遗留 cosmetic:render-queue 部分 directive-first 放置(功能等价)。）

**Active focus**: **知识库地基重建 arc（当前主线）**——三原则:源码为准/自动化/秒查。capindex(`specs/2026-06-16-capability-index.md`,graduate)**P1+P2 DONE**:源码内 `aep:cap` tag → 自动生成可秒查能力/API 索引,**写/做面 100% 标注 + CI 强制零漏标**。getter/const 按 write-surface-first 豁免。**下一步 = `knowledge-consolidation`**(parked→可启动):coverage.md 退役(写面真相源已迁入 tag,已解锁)/ incidents 合并清理 / CLAUDE.md 瘦身 / 记忆系统退役 / 根目录 exe 清理。**能力查询**:`go run ./cmd/capindex -q <词>` 或 grep `docs/capabilities.json`(取代 coverage.md)。—— **库能力主线**早已全收口、需求驱动稳态(roadmap 1-6 + Text Animators 全 ship,详 `specs/2026-06-14-remaining-capability-roadmap.md`);每渲染类双版本 AE ship-gate(红线4)。火焰演示=番外。

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

**➡ capindex P2 完成。下一步 = `knowledge-consolidation`**(`specs/knowledge-consolidation.md`,idea→可启动):
- **A coverage.md 退役**(已解锁):写面真相源已迁入源码 tag + `docs/capabilities.{json,md}`;coverage.md/coverage-detail.md 残值审计后退役/精简为指针。
- **B incidents 审计**:靠 `incident=` 反向链接核 + 过时清理 + 相似合并。
- **C CLAUDE.md 瘦身** · **D 记忆系统退役**(迁 flightdeck) · **E 根目录 exe 清理**(无依赖可先做)。
**遗留小项**(capindex):render-queue tag directive-first 放置统一为 END(cosmetic,功能等价)· orphan 决议(11 manual-gate op:编码 JSX gate 为 Go test 恢复 ae-accept,或维持 stable+roundtrip)。
**能力查询**:`go run ./cmd/capindex -q <词>`(写/做面全覆盖)。

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
