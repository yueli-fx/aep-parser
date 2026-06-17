# Cockpit — aep-parser

**Last updated**: 2026-06-18 by claude（**火焰 Phase 0 命门=质量未过**:7 任务 done + 双版本 gate 绿,但用户真机否决「只有形态、不像火焰」→ roadmap 改顺序=先 Phase 2 学真实样本再重做、过用户关后才参数化。Phase 0 plan 已 archive、showcase procedural-fx 标 ❌质量未过、incident procedural-fx-over-vector +Case2。用户已放样本 `samples/Colorful Fire Ball`(+.rar,已 gitignore)。）

**Active focus**: **火焰程序化 FX 番外**(`specs/2026-06-18-procedural-fx-generator.md`)。Phase 0(纯 Go 手搓确定性火焰)命门**质量未过**——库能确定性造可辨认火焰 + 效果栈/param/animate/mask 全 gated(狭义达成),但**手搓配方只到「可辨认」、到不了「好」**(用户真机否决:缺白热芯/色温渐变/Glow/舔动)。**核心教训**:好视觉的效果栈+参数必须从真实人做的样本学,凭空调参到玩具档。**下一步=Phase 2(学样本)**:用户已放真实火焰工程 `samples/Colorful Fire Ball`,解析抽「好火焰」效果栈+参数 → 重做 → 再过用户关,之后才参数化(原 Phase 1,已推后)。读写主线 + roundtrip→ae-accept 补验 arc 早收口(残值 backlog 见 `specs/2026-06-18-roundtrip-ae-accept-residual.md`,需求驱动)。**不变量**:知识单一家 = flightdeck + CLAUDE.md;能力真相源 = capindex(`go run ./cmd/capindex -q <词>`);每渲染类双版本 AE ship-gate(红线4);火焰=番外(`incidents/procedural-fx-over-vector.md` Case2)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-18-procedural-fx-generator.md](specs/2026-06-18-procedural-fx-generator.md) — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
- [2026-06-18-roundtrip-ae-accept-residual.md](specs/2026-06-18-roundtrip-ae-accept-residual.md) — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
<!-- /AUTO -->

## 下一步

**Phase 2 — 学真实火焰样本(待开工)**。用户已把真实火焰工程放 `samples/Colorful Fire Ball`(+ `.rar`,已 gitignore;非 source、供解析学习)。

- **解析样本 .aep** → 抽「好火焰」真实**效果栈 + 参数范围 + 分层**;对照 v1 缺的:白热芯 / 由内到外色温渐变 / Glow 泛光 / 向上舔动细节。优先纯 AE 原生效果做的工程(查样本是否含 Particular 等插件 / 是否素材视频,若是则要用户另给纯原生的)。
- **用学到的配方重做火焰** → 双版本 AE ship-gate + showcase 自渲 Read png 自验 → 呈用户**真机复核**。过了才翻 showcase `procedural-fx` complete、才谈参数化(原 Phase 1,已推后到「有好火焰之后」)。
- 依据:`specs/2026-06-18-procedural-fx-generator.md` § 更新(2026-06-18) + `archive/plans/2026-06-18-phase0-flame-deterministic.md` § 用户验收结论 + `incidents/procedural-fx-over-vector.md` Case2。

**次要(需求驱动)**:残值 backlog(剩 53 roundtrip)→ `specs/2026-06-18-roundtrip-ae-accept-residual.md`;能力真相源 `go run ./cmd/capindex -q <词>`。

## Backlog

- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
