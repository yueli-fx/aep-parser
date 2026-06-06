# Cockpit — aep-parser

**Last updated**: 2026-06-07 by claude（V3 M8 方案② spec 经三家外审**三轮**整合，定稿 **B′**：`specs/2026-06-07-v3-m8-physical-split-design.md`（status active）。轮1-2 驱动 A→B′ pivot（back-ref 作 scene 内 writer 接口、不破方法 API、无侧表/god-object）；轮3 精度/诚实加固（patch-first 原子序 C-1、attach 仅构造期、C-7 取舍、AttachWriter 暴露、attach 完整性断言、New* 位置、codec 不暴露 rifx）无新架构异议→判定收敛。含 §10 三轮 disposition。待 writing-plans）
**Active focus**: **V3 M8 方案②（真·物理分包）** —— 设计 `specs/2026-06-07-v3-m8-physical-split-design.md`（**A 先行 + B′**，用户通过）+ 实现计划 `plans/2026-06-07-v3-m8-physical-split-plan.md` 已落。**下一步：执行 plan**（P0 基线+inventory → P1 codec 抽包 → P2 back-ref 接口化[最危险] → P3 git mv 分包 → P4 收口+双版本 ship-gate）。执行方式待用户选：subagent-driven（推荐）/ inline。抉择终态：硬编译边界 / B′ back-ref 接口·不破 API / scene 内接口 serializer 实现 / A 先行（opaque+C 日后独立）。
> 收尾记录：**py-aep parity P3 R/W 域已收尾**——§3H ValueText 2026-06-04 **defer/won't-implement**（通用不可达，需 Adobe 不公开 schema DB；详 `incidents/valuetext-needs-schema-db.md`）；DimensionsSeparated animated W **已升 stable**（2026-06-07，feature ship-gate 8/8 早已落，本次仅 doc/classification sync）。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① + M8 scene→rifx 白名单清零 已落；真·物理分包(方案②接口倒置)待启） — [note: 大 arc 暂停 — M8 前置解耦已落（scene→rifx 残留 5 项白名单 2026-06-07 清零，守卫严格禁 import）；真·物理分包需方案② 接口倒置破环，待启；结构性 Phase 1-5 + 包重组方案① 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落；§3H ValueText defer）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**V3 M8 方案② 设计 + 实现计划均已落，下一步执行 plan。** 待办：
1. **执行 `plans/2026-06-07-v3-m8-physical-split-plan.md`**：P0 基线+inventory（Task 0.3 出 Set* 分类表 + XWriter 接口清单，P2 依赖它）→ P1 抽 internal/codec（facade re-alias 保零-diff）→ **P2 单包内 back-ref 接口化[最危险]**（10 类 concrete→XWriter，逐类独立 commit，byte-identical + setter 单测 + attach 完整性断言兜底）→ P3 git mv 物理分包（DAG 编译期硬边界）→ P4 下游+退役 AST 守卫+CLAUDE.md+双版本 ship-gate。执行方式待用户选：subagent-driven（推荐）/ inline。
2. **docgen 次要 follow-on**（非阻塞）：README.md 英文化 + 链接核对；类型级 Example / package-func Example 渲染（pilot defer 项）。
3. Backlog 单条候选：Layr Transform 3D 通道（需先 3D layer 支持，V2.3）。

> py-aep P3 R/W 域已收尾：DimensionsSeparated animated W 已升 **stable**（2026-06-07）；ValueText defer/won't-implement。剩余 Backlog 最大候选 = Layr Transform 3D 通道（需先做 3D layer 支持，V2.3）。

## Backlog（单条候选 / 缺 runtime setter）

> 大 arc（py-aep parity P3 / V3 收尾）现为 active specs，见 ## 进行中（V3 那条带 `note: 大 arc 暂停`）。

**单条候选**：

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API，需纯 RE）。
3. `ImportComposition`（需求驱动）。
4. **Property synthesis**（暂搁，可选大 feature）：AE 省略未改的默认属性，py-aep 合成完整 transform schema 我们不合成。补合成 + `Elided` 是唯一让 transform group 对齐 py-aep 长度的路。次要 fidelity：animated orientation 的 easing/tangents（旧 1D layout 未校验）。详 `incidents/transform-group-default-omission.md`。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 [specs/deferred-backlog.md](specs/deferred-backlog.md) + [plans/coverage.md](plans/coverage.md)。

（已落条目历史见 `git log`，不在 cockpit 留存。）

## Hanging tasks

无。
