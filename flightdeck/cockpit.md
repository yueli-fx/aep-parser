# Cockpit — aep-parser

**Last updated**: 2026-06-09 by claude（M8 物理分包 P0-P4 完成并归档；下一步=follow-up trim / 大方向待定向）
**Active focus**: **V3 M8 物理分包 ✅ 完成（2026-06-09）** — `internal/{rifx,codec,scene,serializer}` + `aep` 薄 facade 落地，DAG 经包级 arch_boundary 守卫 + dag_boundary 强制，全程 byte-identical(183 fixture)。plan 已归档 `archive/plans/2026-06-07-v3-m8-physical-split-plan.md`（含 P4.3 ship-gate 经 byte-identity 等效验收）。**当前无 active 大 plan**；V3 大 arc 其余暂停，下一步见下。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① + M8 scene→rifx 白名单清零 已落；真·物理分包(方案②接口倒置)执行中 P2 7/10） — [note: 大 arc 暂停 — 真·物理分包(方案②接口倒置)执行中，见 M8 plan P2 7/10；结构性 Phase 1-5 + 包重组方案① + scene→rifx 白名单清零 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落，剩 ValueText）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [2026-06-07-v3-m8-physical-split-design.md](specs/2026-06-07-v3-m8-physical-split-design.md) — V3 M8 方案② 真·物理分包设计（A 先行 + B′ back-ref 接口）：scene/serializer/codec 物理拆包，back-ref 作 scene 内 writer 接口（serializer 实现）→ 保留全部方法 API（不破 API）、scene 编译期零 rifx；eager length-preserving patch 经接口；opaque 延后到 C；全程保 byte-exact 回归门 — [note: brainstorm + 三家外审两轮整合。方向 = A 先行（保 byte-exact，C 日后独立）+ B′（back-ref 接口、不破 API、无侧表/无 Document god-object）。前置：M8 scene→rifx 白名单清零（2026-06-07 已落）]
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
- [m8-setter-inventory.md](plans/m8-setter-inventory.md) — V3 M8 Task 0.3 产出 — 全量 Set* 分类表（A 类 back-ref setter / B 类 pure-graph setter）+ 每 backref 结构的 XWriter 接口方法清单。P2（back-ref 接口倒置）的逐类输入。 — [note: 从 internal/aep 实测枚举（277 个 Set* + 10 个 *Backrefs 结构 + 全部 `.back.` 字节访问点），逐方法读 body 分类，非按名猜。spec §3 分类规则 + edge-case adjudication。]
<!-- /AUTO -->

## 下一步

**M8 物理分包 arc 全完**（P0-P4，2026-06-09）。plan 已归档 `archive/plans/2026-06-07-v3-m8-physical-split-plan.md`；逐 commit 历程见 git log（`a347e46` serializer 抽出 + `ed4669c` arch 守卫/CLAUDE.md #3）+ 设计 `specs/2026-06-07-v3-m8-physical-split-design.md`。

1. **immediate follow-up（非阻塞，单独 commit）**：serializer 结构性 op 自由函数仍带与 facade 重复的富 doc comment（doc home 已是 facade）→ trim 为简短内部注释。
2. **大方向待用户定向**（当前无 active 大 plan）：① 续 V3 大 arc（`specs/2026-05-22-v3-direction.md`，opaque 子表/M9+）；② backlog 顶：Layr Transform 3D 通道（Orientation/Rotate X·Y/Position_Z，需 3D layer 支持 V2.3）；③ py-aep parity 剩 ValueText（schema-db 依赖，见 `incidents/valuetext-needs-schema-db.md`）。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
