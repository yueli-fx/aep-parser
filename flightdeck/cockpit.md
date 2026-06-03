# Cockpit — aep-parser

**Last updated**: 2026-06-04 by claude（迁移 deck 2.3→3.0 + model-v4：sketches/debriefs 折叠、status 6→4、cockpit 进行中改为 active 集 AUTO 投影）
**Active focus**: **py-aep parity P3** — §3A Render Queue read+write **完成（含结构性增删）**。**SetRenderer / Guides R/W / 分离维度 R+W(双向) / 3D Orientation R / Essential Graphics R / RQ SetComment + Remove + Add / 3C PropertyBase Remove + MoveTo + Duplicate(ship-gated) 已落**。剩 ValueText（需 AE schema DB）+ DimensionsSeparated animated 子方向。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① 已落；M8 物理分包待启） — [note: 大 arc 暂停 — M8 物理分包待启（scene/serializer 拆包，scene→rifx 残留 5 项白名单待清零）；结构性 Phase 1-5 + 包重组方案① 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向 + 3C PropertyBase Remove/MoveTo/Duplicate 已落，剩 ValueText/animated 子方向/3G marker）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**其余 P3**：`Property.ValueText`（formatted value，需 AE schema DB，最大 RE，多会话）；DimensionsSeparated animated 子方向（keyframe 流迁移）；3G comp marker 增删。RQ 域已完整（read + value W + 结构性增删）；3C 结构性域（Remove/MoveTo/Duplicate）已完整。详 `specs/2026-05-26-py-aep-parity-design.md` §3 Phase 3。

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
