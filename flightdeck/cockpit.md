# Cockpit — aep-parser

**Last updated**: 2026-06-04 by claude（P3 §3C DimensionsSeparated animated 实现 + AE 双版本 ship-gate 8/8 完，cut-3 修正落 `b104e6f`，plan landing 归档）
**Active focus**: **py-aep parity P3** — §3A Render Queue read+write **完成（含结构性增删）**。**SetRenderer / Guides R/W / 分离维度 R+W(双向) / 3D Orientation R / Essential Graphics R / RQ SetComment + Remove + Add / 3C PropertyBase Remove + MoveTo + Duplicate / 3G comp marker 增删(Remove+AddMarker, ship-gated) / 3C DimensionsSeparated animated R+W(双向, ship-gate 8/8, Alpha) 已落**。剩 ValueText（需 AE schema DB）+ DimensionsSeparated animated 升 stable doc-sync（小尾）。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① 已落；M8 物理分包待启） — [note: 大 arc 暂停 — M8 物理分包待启（scene/serializer 拆包，scene→rifx 残留 5 项白名单待清零）；结构性 Phase 1-5 + 包重组方案① 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落，剩 ValueText）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**`Property.ValueText`** — P3 最后大项（最大 RE，需 AE schema DB → 多会话）。先 RE：建 fixture 探 AE 各 effect/property 的 value-text schema 存储格式 + 定位字节。

> 小尾（非阻塞）：DimensionsSeparated animated 已实现 + 双版本 ship-gate 8/8 PASS（landed `landed/plans/2026-06-04-py-aep-p3-dimsep-animated-plan.md`），但升 stable 的 doc-sync（coverage / coverage-detail / docs / property.md）本会话按用户指示暂跳过，feature 现标 **Alpha**。需要时补这步即升 Stable。

## Backlog（单条候选 / 缺 runtime setter）

> 大 arc（py-aep parity P3 / V3 收尾）现为 active specs，见 ## 进行中（V3 那条带 `note: 大 arc 暂停`）。

**单条候选**：

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API，需纯 RE）。
3. `ImportComposition`（需求驱动）。
4. **Property synthesis**（暂搁，可选大 feature）：AE 省略未改的默认属性，py-aep 合成完整 transform schema 我们不合成。补合成 + `Elided` 是唯一让 transform group 对齐 py-aep 长度的路。次要 fidelity：animated orientation 的 easing/tangents（旧 1D layout 未校验）。详 `incidents/transform-group-default-omission.md`。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 [specs/deferred-backlog.md](specs/deferred-backlog.md) + [plans/coverage.md](plans/coverage.md)。

（已落条目历史见 `git log` + [landed/HISTORY.md](landed/HISTORY.md)，不在 cockpit 留存。）

## Hanging tasks

无。
