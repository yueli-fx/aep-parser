# Cockpit — aep-parser

**Last updated**: 2026-06-04 by claude（docgen 全 arc 完成 + landed：生成器 + 11 个符号 doc 全部生成物化 + json 英文化；spec done → landed/specs/。docgen 焦点结束）
**Active focus**: **开放（docgen 已收官）** —— docgen arc 全部落地（生成器 `cmd/docgen` + 5 能力 + 11 个生成 doc + `docs_index.json` KV 索引 + drift gate）。下一焦点未定：候选 = V3 M8 物理分包 / py-aep P3 尾（ValueText defer、DimSep animated doc-sync）/ docgen 次要 follow-on（README 英文化）。
> 并行收尾：**py-aep parity P3 R/W 域实质收尾**——§3H ValueText 2026-06-04 **defer/won't-implement**（通用不可达，需 Adobe 不公开 schema DB；详 `incidents/valuetext-needs-schema-db.md`）；唯一小尾 DimensionsSeparated animated 升 stable doc-sync。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① 已落；M8 物理分包待启） — [note: 大 arc 暂停 — M8 物理分包待启（scene/serializer 拆包，scene→rifx 残留 5 项白名单待清零）；结构性 Phase 1-5 + 包重组方案① 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落；§3H ValueText defer）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**docgen 收官，焦点开放 —— 由用户定下一步。** 候选：
1. **V3 M8 物理分包**（scene/serializer 拆包，scene→rifx 残留 5 项白名单清零）—— 最大未启 arc。
2. **py-aep P3 尾**：DimensionsSeparated animated 升 stable 的 doc-sync（feature 已 ship-gate 8/8，标 Alpha）；ValueText 已 defer。
3. **docgen 次要 follow-on**（非阻塞）：README.md 英文化 + 链接核对；类型级 Example / package-func Example 渲染（pilot defer 项）。
4. Backlog 单条候选：Layr Transform 3D 通道（需先 3D layer 支持，V2.3）。

> 旁路小尾（py-aep P3，非阻塞）：DimensionsSeparated animated 已 ship-gate 8/8（landed），升 stable 的 doc-sync 暂跳过，feature 标 **Alpha**；或转 Backlog 最大候选 Layr Transform 3D 通道（需先做 3D layer 支持，V2.3）。

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
