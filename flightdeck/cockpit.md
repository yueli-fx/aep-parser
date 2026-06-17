# Cockpit — aep-parser

**Last updated**: 2026-06-18 by claude（**火焰 Phase 2 进行中**:样本 #1 解析 → 配方 + **效果用途词典(角色→效果)**固化进 `checklists/build-good-fire.md`。v2 单层栈(Tritone 色温+Glo2 辉光)AE2025 自渲修掉 v1 三缺口,但**层次感不足**——用户定调「层次感来自多层混合,不是调参」。下一步 v3 = 多层合成结构。Phase 0 plan 已 archive、showcase ❌质量未过。）

**Active focus**: **火焰程序化 FX 番外**(`specs/2026-06-18-procedural-fx-generator.md`)。Phase 0(纯 Go 手搓确定性火焰)命门**质量未过**——库能确定性造可辨认火焰 + 效果栈/param/animate/mask 全 gated(狭义达成),但**手搓配方只到「可辨认」、到不了「好」**(用户真机否决:缺白热芯/色温渐变/Glow/舔动)。**核心教训**:好视觉的效果栈+参数必须从真实人做的样本学,凭空调参到玩具档。**下一步=Phase 2(学样本)**:用户已放真实火焰工程 `samples/Colorful Fire Ball`,解析抽「好火焰」效果栈+参数 → 重做 → 再过用户关,之后才参数化(原 Phase 1,已推后)。读写主线 + roundtrip→ae-accept 补验 arc 早收口(残值 backlog 见 `specs/2026-06-18-roundtrip-ae-accept-residual.md`,需求驱动)。**不变量**:知识单一家 = flightdeck + CLAUDE.md;能力真相源 = capindex(`go run ./cmd/capindex -q <词>`);每渲染类双版本 AE ship-gate(红线4);火焰=番外(`incidents/procedural-fx-over-vector.md` Case2)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-18-procedural-fx-generator.md](specs/2026-06-18-procedural-fx-generator.md) — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
- [2026-06-18-roundtrip-ae-accept-residual.md](specs/2026-06-18-roundtrip-ae-accept-residual.md) — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
<!-- /AUTO -->

## 下一步

**Phase 2 — 火焰 v3 多层合成(为「层次感」,待开工)**。v2 单层栈(Fractal Noise→Tritone→Turbulent Displace→Glo2+mask+黑底)AE2025 自渲:**色温/辉光/翻腾对了**(修掉 v1 三缺口),但**层次感不足**——用户定调:「**调再多参数没意义了,层次感来自多层混合**」。坐实单层天花板。

- **v3 = 搭多层合成结构**(非继续调参):同套「噪声→Displacement Map→Ramp」图层**复制 3–4 份**(不同位移/偏移/色温)+ **混合模式**(`Add` 叠热芯 / `Difference` 做负火出暗筋 / `Divide`)+ 内外焰分层 + 全局 Glow。本库可做(`NewSolidLayer`×N + `AddEffect` + `Layer.SetBlendingMode` + `SetEffectLayerParam` 指噪声层,全 gated)。
- **角色词典(效果→用途)+「层次感=合成结构」已固化** `checklists/build-good-fire.md`(本次新增,比参数表重要)。
- → v3 自渲自验 → 呈用户真机 → 过了才双版本 gate + showcase 升级 + checklist 升「验证配方」。
- v2 builder 在 `tmp_debug/flame2/`(local);依据 `checklists/build-good-fire.md` · `specs/...procedural-fx-generator.md` § 样本解析 #1。

**次要(需求驱动)**:残值 backlog(剩 53 roundtrip)→ `specs/2026-06-18-roundtrip-ae-accept-residual.md`;能力真相源 `go run ./cmd/capindex -q <词>`。

## Backlog

- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
