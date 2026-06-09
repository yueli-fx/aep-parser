# Cockpit — aep-parser

**Last updated**: 2026-06-10 by claude（AddEffect W ship — 12-effect 库，AE 双版本 ship-gate 10/10，incident `add-effect-splice-re.md`）
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
2. **新建图层类型**：当前仅 `NewShapeLayer`。缺 Solid / Null / Text / Camera / Light / Adjustment。Camera/Light 像 shape 层（无 source，靠 ldta+属性树，embed-template 可推）；Solid/Null 需先建 footage source item。
3. **AddEffect 扩库 / 参数化**：更多内置效果（注意带 layer/path 引用参数的效果需 sspc id remap，类似 cross-Project InsertLayer）；per-effect typed param helper（今为 raw `SetStaticValue` by match-name）。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **结构性创建 vein**（见 ↑下一步 1-3）— **当前主线**，纯代码可推。
2. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
3. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
