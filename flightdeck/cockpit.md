# Cockpit — aep-parser

**Last updated**: 2026-06-09 by claude（V3 arc 收尾 + 文档治理收口；cockpit 清理）
**Active focus**: **V3 arc 实质收尾** — M8 物理分包已落（`internal/{rifx,codec,scene,serializer}` + `aep` facade）；**可达字段 ~99% 已 ship**，剩余前沿均 fixture/RE-gated 或架构不可达（详 `plans/coverage.md`）。docs 已全自动（docgen 多包，`incidents/docgen-alias-blindspot-false-green.md` 记教训）。**当前无 active 大 plan**，下一步见下。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split。M1-M8 框架已实现（结构性 mutation Phase 1-5 + 包重组① + M8 真·物理分包② 全落）；可达字段 ~99% ship，剩余前沿 fixture/RE-gated（详 plans/coverage.md）。 — [note: 大 arc 实质收尾（2026-06-09）—— M8 物理分包完成（design+plan 已归档）；剩 ValueText / Layr 3D 通道等 fixture-gated 边缘项，无 active 大 plan。]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落，剩 ValueText）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

无 active 大 plan、无简单待办——2026-06-09 收口会话已落：M8 物理分包 + 文档全自动化（docgen 多包修复 / docs 删纯人写页 / 双 README 瘦身）+ serializer 注释去重 + incidents 审计（详 git log）。测试分散决议 **B=不分散**（黑盒端到端测试按 Go 惯例留 facade）。

**真·下一步（均需外部输入，非「顺手」）**：V3 剩余前沿全 **fixture/RE-gated** —— Layr Transform 3D 通道（需 3D layer 支持，backlog 顶）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。推进需用户提供 AE fixture 或起一轮 RE。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
