# Cockpit — aep-parser

**Last updated**: 2026-06-04 by claude（docgen pilot 验收通过 + ship：property.md/marker.md 生成物化、drift gate 上 `go test`、directive 验证、铁律调和；pilot plan landed；剩余 ~11 文件推广待续）
**Active focus**: **docgen 推广** —— 生成器 `cmd/docgen` 已 ship + 加固（inline field comment / example 注释 / JSON 索引 / package-level func 渲染）。**已生成物化 8 个文件**：property（70 节）/ marker / footage / mask / effect / project / composition / constants；+ `docs/docs_index.json`；drift gate `TestDocsUpToDate` 守 `go test ./...`。决策已定：**doc comment 英文为源**。**剩 3 个大 wrinkle 文件**：layer（1192）/ text（928）/ shape（含 builder 教程）+ json（特殊，JSON 导出说明非符号文档，待评估）。README 不生成（手写导航）。
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

**docgen 推广剩 3 个大文件 + json**（已落 8 个）。各自方案：
- **layer.md**（1192）：linchpin 类型，最大。补 ~6 个裸字段（Index/Name/Type/StartTime/Duration/Stretch + 子集合）；text 的 `SetRun*` setter 都在 `*Layer` 上 → 自然渲在这。完整 camera/light/material/iris setter 会全出（completeness）。先推这个，text 可 cross-ref。
- **text.md**（928）：多类型（TextSource/TextStyleRun/TextParagraph + 枚举）；`TextEncodedByteLen` 走 funcs；`Layer.SetText`/`SetRun*` 已在 layer.md → text.md 加 cross-ref 注。
- **shape.md**：ShapePath/ShapePrimitive/ShapePrimitiveKind roots + **V2.2 Builder API 手写教程 ~100 行**（中文）→ 翻成英文进 tail include（或评估是否仍准）。
- **json.md**：JSON 导出格式说明，非符号文档 → 大概率保持手写 or 走纯 include，待评估。
每文件流程同 recipe：核 doc comment 英文成熟度 → 补缺 + Example + `_includes` → manifest 加节 → 生成过 drift gate。

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
