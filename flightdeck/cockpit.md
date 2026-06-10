# Cockpit — aep-parser

**Last updated**: 2026-06-10 by claude（NewCameraLayer/NewLightLayer ship — source-less 图层创建，embed-whole-Layr，双版本 ship-gate PASS，incident `camera-light-layer-create-re.md`）
**Active focus**: **结构性创建 vein（新发现的纯代码前沿）** — 之前「可达字段 ~99% 已 ship」只指**字段** R/W；**结构性创建路径**（新建图层类型、给图层加效果）此前未被识别为前沿，实则**纯代码可推 + ship-gate 自助**（agent 跑 `scripts/ae_run.ps1` 双版本无人值守）。2026-06-10 落 **AddEffect**（`aep.AddEffect`/`SupportedEffects`，12 内置效果库，splice `(tdmn,sspc)` pair 进 Effect Parade，双版本 10/10）。M8 物理分包已落。下一步见下。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split。M1-M8 框架已实现（结构性 mutation Phase 1-5 + 包重组① + M8 真·物理分包② 全落）；可达字段 ~99% ship，剩余前沿 fixture/RE-gated（详 plans/coverage.md）。 — [note: 大 arc 实质收尾（2026-06-09）—— M8 物理分包完成（design+plan 已归档）；剩 ValueText / Layr 3D 通道等 fixture-gated 边缘项，无 active 大 plan。]
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**结构性创建 vein 继续推**（均纯代码可推 + ship-gate 自助，无需用户输入）。按价值/风险排序：

1. **AddEffect Phase 2 — Effect Parade auto-create**：让 `NewShapeLayer` 的 from-scratch 层也能加效果（当前 refuse parade-less 层）。需 RE 空 Effect Parade group 在 Layr property tree 的位置 + 空组字节，splice 进去。详 `incidents/add-effect-splice-re.md` § Phase 2。
2. **新建图层类型**（详 `incidents/new-layer-types-scoping.md` + `camera-light-layer-create-re.md`）：✅ **Camera/Light 已 ship**（2026-06-10，embed-whole-Layr，双版本 PASS）。剩：**NewTextLayer**（embed-whole-Layr 可行，但 btdk 文本 blob 复杂）；**Solid/Null/Adjustment** = footage-backed（共用 solid footage Item：`opti "Soli"` + Pin/sspc + Solids folder），走 InsertLayer closure-import + ID-remap，**AE-acceptance gate 高风险**（同 NewComposition 5 阶段）。fresh 层 setter 暂搁（scene-vs-chunk 墙，需 write+reopen）。
3. **AddEffect 扩库 / 参数化**：更多内置效果（注意带 layer/path 引用参数的效果需 sspc id remap，类似 cross-Project InsertLayer）；per-effect typed param helper（今为 raw `SetStaticValue` by match-name）。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **结构性创建 vein**（见 ↑下一步 1-3）— **当前主线**，纯代码可推。
2. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
3. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

### Agent 建议（2026-06-10 会话留，按价值/成本排序）

1. **★ 统一「fresh-layer setter 墙」**（跨切面，最高杠杆）：本会话两处撞同一墙——AddEffect Phase 2 与 Camera/Light setter 都因「from-scratch 层无 scene tree + scene 禁持 rifx chunk」而无法在内存态调值/加子结构。值得一个**统一方案**而非逐个 workaround：候选 (a) `Project.Reopen()`/`Layer.Reparse()` helper（write→parse 往返，把 built 层升级为 parsed 层，一行解锁所有 setter）；(b) serializer-side pending-mutation 表，lower 时重放。(a) 最省、最快见效。详 `incidents/add-effect-splice-re.md` + `camera-light-layer-create-re.md`。
2. **完成 effect 库 ship-gate**（便宜稳健）：当前 12 效果只有 5 个 AE 双版本 gated，余 7 个走同机制仅 Go round-trip。补跑 7×2 AE runs 即全 12 个「verified」，去掉「rides the mechanism」免责声明。
3. **AddMask**（机制已证）：Mask Parade 同 AddEffect 的 INDEXED_GROUP splice；但 mask path 写是 structural（暂搁），故 v1 只能 fixed-shape mask，价值有限——等 mask-path-write 解封再做。
4. **NewNullLayer 优先于其它 footage-backed**：Null（parenting 用）是 footage-backed 家族里最常用的；做 solid footage Item 时先瞄准 Null/Solid。

## Hanging tasks

无。
