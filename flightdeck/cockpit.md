# Cockpit — aep-parser

**Last updated**: 2026-06-09 by claude（收口会话：incidents 审计 + docgen 多包修复 + docs/ 全自动化（删纯人写页）+ 根/docs README 去过时；测试分散决议 B=不分散）
**Active focus**: **V3 arc 实质收尾** — M8 物理分包完成（`internal/{rifx,codec,scene,serializer}` + `aep` 薄 facade，已归档 `archive/plans/2026-06-07-v3-m8-physical-split-plan.md`）。**可达字段覆盖 ~99% 已 ship**（M1-M8 框架基本实现，见 `plans/coverage.md` 末）；剩余前沿均 fixture/RE-gated 或架构不可达。**本会话收口**：incidents 审计（0 过期 + 5 路径 refresh，`ba3b40b`）、**docgen 多包修复**（修 stage-1 起静默空类型文档的回归，`1bc4ad4`，恢复 +11.8k 行；教训见 `incidents/docgen-alias-blindspot-false-green.md`）。**当前无 active 大 plan**，下一步见下。

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

1. **测试分散 = 已决 B（不分散，2026-06-09）**：黑盒端到端测试（114 个 `aep_test`）按 Go 惯例留 `internal/aep` facade（测公开 API、天然跨层）；纯子系统白盒单测在 M8 时已随迁 serializer/scene。**无后续动作**。
2. **docs/ = 已全自动（2026-06-09，`4bff35a`）**：删两个纯人写页（README.md + json.md）；docs/ 仅余 docgen 生成的 11 类型 .md + docs_index.json + docgen.json + `_includes/`（管线源）。docgen 多包扫描已修（`1bc4ad4`）。根 README + docs link 已去过时（`5e14c13`/`3c99306`）。
3. **immediate follow-up（非阻塞）**：serializer 结构性 op 自由函数仍带与 facade 重复的富 doc comment（doc home 已是 facade）→ trim 为简短内部注释。
4. **V3 剩余前沿（均 fixture/RE-gated，需用户提供 fixture 或新发现才动）**：Layr Transform 3D 通道（需 3D layer 支持，backlog 顶）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建 …）· ValueText（schema-db 依赖）。详 `plans/coverage.md` § 暂搁/不可达。
5. **（offer，未决）根 README 全量功能刷新**：V1 时代 support 表缺大批 V2/V3 功能（不是错，是缺）；要做另开一轮。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
