# Cockpit — aep-parser

**Last updated**: 2026-06-04 by claude（docgen pilot 验收通过 + ship：property.md/marker.md 生成物化、drift gate 上 `go test`、directive 验证、铁律调和；pilot plan landed；剩余 ~11 文件推广待续）
**Active focus**: **docgen 推广** —— 生成器 `cmd/docgen` 已 ship + 加固（inline field comment / example 注释 / JSON 索引）。**已生成物化 4 个文件**：property（70 节）/ marker / footage / mask；+ `docs/docs_index.json`（符号 KV 索引）；drift gate `TestDocsUpToDate` 守 `go test ./...`。决策已定：**doc comment 英文为源**。下一步继续推剩余 docs（clean: shape/constants/text；wrinkle: composition/layer/project/effect；recipe 见 docgen spec § Pilot outcome）。
> 并行收尾：**py-aep parity P3 R/W 域实质收尾**——§3H ValueText 2026-06-04 **defer/won't-implement**（通用不可达，需 Adobe 不公开 schema DB；详 `incidents/valuetext-needs-schema-db.md`）；唯一小尾 DimensionsSeparated animated 升 stable doc-sync。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-04-docgen-from-comments-design.md](specs/2026-06-04-docgen-from-comments-design.md) — docgen：docs/*.md 从 Go doc comment 自动生成。**pilot 验收通过 + ship**（生成器 + property.md/marker.md 生成物化 + drift gate + directive 验证）；剩余 ~11 文件推广进行中，recipe + 每文件 wrinkle 见 § Pilot outcome
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① 已落；M8 物理分包待启） — [note: 大 arc 暂停 — M8 物理分包待启（scene/serializer 拆包，scene→rifx 残留 5 项白名单待清零）；结构性 Phase 1-5 + 包重组方案① 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落；§3H ValueText defer）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**docgen 推广下一个文件**（已落 property/marker/footage/mask）：按 recipe 推下一个 **clean 域**（shape / constants / text），再啃 wrinkle 文件（composition 的 `NewComposition` 跨类型 + bespoke 分组拍平、layer 1192 行、project）。每文件：核 doc comment 英文成熟度 → 补缺 + Example 函数 + `_includes` 表格 → manifest 加节 → 生成 + `go test ./cmd/docgen` 过 drift gate。注：constants.md 是跨类型 enum 聚合，可能不直映单 root（待评估按多 enum root 渲 or 走 include）。

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
