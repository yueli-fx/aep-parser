# Cockpit — aep-parser

**Last updated**: 2026-06-11 by claude（NewTextLayer ship（双版本 gate 2/2，新建图层类型**全部建齐**）+ effect 库全 12 模板双版本 gated + AddEffect/RemoveEffect 升 Stable + btdk 改字长 scoping 落 `incidents/text-btdk-length-variable-write-scoping.md`——T2 probe 证明 AE 载入重算布局缓存）
**Active focus**: **结构性创建 vein（纯代码前沿）** — 结构性创建路径纯代码可推 + ship-gate 自助（agent 跑 `scripts/ae_run.ps1` 双版本无人值守）。2026-06-10 落：AddEffect（12 效果库）→ Camera/Light → parade auto-create + `aep.Reopen` → Solid/Null/Adjustment（详 `incidents/new-layer-types-scoping.md`）。2026-06-11 落：effect 全 12 模板双版本 gated + AddEffect/RemoveEffect 升 Stable + **NewTextLayer**（embed-whole-Layr，**新建图层类型全部建齐**；fresh 层免 Reopen 可读 TextSource/等长 SetText；btdk 改字长解封路径详 `incidents/text-btdk-length-variable-write-scoping.md`）。⚠ scar：Go-built 工程 allocItemID 撞 service 层 ID 2..12 → AE 2025 拒收，已修（详 `incidents/nextitemid-must-include-layer-ids.md`）。M8 物理分包已落。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split。M1-M8 框架已实现（结构性 mutation Phase 1-5 + 包重组① + M8 真·物理分包② 全落）；可达字段 ~99% ship，剩余前沿 fixture/RE-gated（详 plans/coverage.md）。 — [note: 大 arc 实质收尾（2026-06-09）—— M8 物理分包完成（design+plan 已归档）；剩 ValueText / Layr 3D 通道等 fixture-gated 边缘项，无 active 大 plan。]
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**结构性创建 vein 继续推**（均纯代码可推 + ship-gate 自助，无需用户输入）。按价值/风险排序：

1. **Length-variable 文本写（任意长度 SetText）**：最大未知已消除——NewTextLayer gate T2 probe 证明 AE 载入时重算 btdk 布局缓存（stale 字形度量被接受）。剩余 = 改串 + 段落/run 字符计数（2 处 PsNum）+ btdk size header 修正（AddFont 已有 splice 范本）；v1 限单段落。解封路径详 `incidents/text-btdk-length-variable-write-scoping.md`。
2. **AddEffect 扩库 / 参数化**：更多内置效果（⚠ 带 layer/path 引用参数的效果其 tdpi 指向非宿主层，需选择性 remap——盲 retarget-all 会写坏，详 incident finding 5）；per-effect typed param helper（今为 raw `SetStaticValue` by match-name）。（✅ 既有 12 模板已全部双版本 ship-gate，2026-06-11）
3. **AddMask**（机制已证）：Mask Parade 同 INDEXED_GROUP splice + parade auto-create 模式可复用；但 mask path 写是 structural（暂搁），v1 只能 fixed-shape mask——等 mask-path-write 解封再做。
4. **Solid 系后续小件**（需求驱动再做）：`SetSolidColor`/`SetSolidSize` 独立 setter（机制已 RE：opti @0x0A ARGB + sspc @0x20/0x24，今只在创建参数暴露）；导入的 solid footage 落 dest 根而非 Solids folder（AE 接受，仅整理性差异）。
5. **New\* 图层家族 Alpha→Stable 审计**（文档级，已获授权）：Camera/Light/Solid/Null/Adjustment/Text 均双版本 gated，按 CLAUDE.md #2 可升 Stable——逐一核对 gate 证据后一次 sweep（参照 AddEffect 升级先例）。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **结构性创建 vein**（见 ↑下一步 1-4）— **当前主线**，纯代码可推。
2. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
3. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
