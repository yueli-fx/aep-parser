# Cockpit — aep-parser

**Last updated**: 2026-06-18 by claude（**roundtrip→ae-accept 补验 arc 收尾 + landed**:批1-28 完成,ae-accept 35→**261**/roundtrip 279→**53**;末三批=批26 track-matte explicit 单版本破例(契约)·批27 material-advanced 11 负结论(`.enabled`≠settable)·批28 RQ value-setter 36 acceptance-preservation。剩 53 残值移交 `specs/2026-06-18-roundtrip-ae-accept-residual.md`。plan 已 archive。）

**Active focus**: **无活跃 arc**——读写主线 + roundtrip→ae-accept 补验 arc(批1-28,landed 2026-06-18)均收口。下一步**需求驱动**:残值 backlog(剩 53 roundtrip 的分类 + someday 优先级)见 `specs/2026-06-18-roundtrip-ae-accept-residual.md`;真要继续补验从该 spec § 二.D 挑(最便宜=RQ AddItem/RemoveItem 洁癖洞 / project header acceptance)。**不变量**:知识单一家 = flightdeck + CLAUDE.md(auto-memory 已退役);能力真相源 = capindex(`go run ./cmd/capindex -q <词>`,CI 强制零漏标);每渲染类双版本 AE ship-gate(红线4);单版本 ae-accept 例外见 `checklists/delivery-contract.md`。火焰=番外(`incidents/procedural-fx-over-vector.md`)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-18-procedural-fx-generator.md](specs/2026-06-18-procedural-fx-generator.md) — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
- [2026-06-18-roundtrip-ae-accept-residual.md](specs/2026-06-18-roundtrip-ae-accept-residual.md) — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
<!-- /AUTO -->

## 下一步

**无活跃任务**——补验 arc 已收尾 landed。下一步**需求驱动**。

- **残值 backlog(剩 53 roundtrip)** → `specs/2026-06-18-roundtrip-ae-accept-residual.md`(完成记录 + 逐项分类 + someday 优先级)。真要继续从该 spec § 二.D 挑;最便宜=**RQ AddItem/RemoveItem 洁癖洞核实** / **project header acceptance-preservation**(仿批28 `TestRenderQueueSettings`)。
- 域分布查询:`pwsh -c "(gc docs/capabilities.json -raw|ConvertFrom-Json)|?{$_.cap.verify -eq 'roundtrip'}|group {$_.cap.domain}"`;能力真相源 `go run ./cmd/capindex -q <词>`。
- 能力 roadmap 早已全收口(effects/mask/expr,详 capindex + `archive/specs/`)。

## Backlog

- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
