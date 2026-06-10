# Cockpit — aep-parser

**Last updated**: 2026-06-11 by claude（③ **效果参数 elision RE**：AE 按值≠默认持久化（negative finding：`setValue(默认)` 不物化）→ 参数化定向 synthesis-lite 路线，新 incident `effect-param-elision-synthesis-lite` ② **New\* 图层家族六个全升 Stable**（审计落地，含 NewTextLayer doc 过期段修正）① **AddEffect 库 wave-2 扩到 29**：新增 17 个 parameter-only 内置效果（Drop Shadow/Sharpen/Mosaic/Noise/Transform/Gradient Ramp/Fractal Noise/Motion Tile/Directional Blur/Linear Wipe/Wave Warp/Curves + 5 表达式控制），双版本 ship-gate 34/34 PASS，全库 29/29 无未 gate 模板；新可复用提取器 `tmp_debug/extract_effect_lib`；commit `322ea00`。同日早些：effect 全 12 gated + AddEffect/RemoveEffect 升 Stable、NewTextLayer ship、SetText 升级 length-variable）
**Active focus**: **结构性创建 vein（纯代码前沿）** — 结构性创建路径纯代码可推 + ship-gate 自助（agent 跑 `scripts/ae_run.ps1` 双版本无人值守）。2026-06-10 落：AddEffect（12 效果库）→ Camera/Light → parade auto-create + `aep.Reopen` → Solid/Null/Adjustment（详 `incidents/new-layer-types-scoping.md`）。2026-06-11 落：effect 全 12 模板双版本 gated + AddEffect/RemoveEffect 升 Stable + **NewTextLayer**（embed-whole-Layr，**新建图层类型全部建齐**；fresh 层免 Reopen 可读 TextSource/等长 SetText；btdk 改字长解封路径详 `incidents/text-btdk-length-variable-write-scoping.md`）。⚠ scar：Go-built 工程 allocItemID 撞 service 层 ID 2..12 → AE 2025 拒收，已修（详 `incidents/nextitemid-must-include-layer-ids.md`）。M8 物理分包已落。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split。M1-M8 框架已实现（结构性 mutation Phase 1-5 + 包重组① + M8 真·物理分包② 全落）；可达字段 ~99% ship，剩余前沿 fixture/RE-gated（详 plans/coverage.md）。 — [note: 大 arc 实质收尾（2026-06-09）—— M8 物理分包完成（design+plan 已归档）；剩 ValueText / Layr 3D 通道等 fixture-gated 边缘项，无 active 大 plan。]
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**结构性创建 vein 继续推**（均纯代码可推 + ship-gate 自助，无需用户输入）。按价值/风险排序：

1. **AddEffect 参数化 = synthesis-lite 实现**（RE 已齐，2026-06-11，详 `incidents/effect-param-elision-synthesis-lite.md`）：AE 按「值≠默认」持久化参数（非 touched 标记，`setValue(默认)` 也被 elide——negative finding 堵死换模板路线）；物化参数 tdbs 小且**无 tdpi**，splice 即可。方案 = `EnsureEffectParameter`：set 时按需 splice 参数级 `(tdmn, tdbs)` 模板再写值，输出 = AE 自身语义，**无需重 gate 29 个 add 模板**。待决：per-param 模板 vs 泛型 per-control-type 模板（~7 个，需 gate 证 tdb4 非参数特定）。（✅ 扩库 wave-2 已落，库 = 29 全部双版本 gated；再扩库 = 需求驱动，⚠ 引用参数效果 tdpi 指向非宿主层详 add-effect incident finding 5）
2. **AddMask**（机制已证）：Mask Parade 同 INDEXED_GROUP splice + parade auto-create 模式可复用；但 mask path 写是 structural（暂搁），v1 只能 fixed-shape mask——等 mask-path-write 解封再做。
3. ~~New\* 图层家族 Alpha→Stable 审计~~（✅ 2026-06-11 落：六个 New\* 全升 Stable，gate 证据 = 三个 feature commit 的双版本 PASS 断言 + gate 测试在册；顺带修 NewTextLayer doc comment 过期的「等长 SetText」段）。
4. **SetText 变长 refuse 集解封**（需求驱动再做）：多段落（splice 段落 dict entry）/ 多 run（计数分配 = AE 行为 RE）/ 空串。详 `incidents/text-btdk-length-variable-write-scoping.md` § v1 守卫。
5. **Solid 系后续小件**（需求驱动再做）：`SetSolidColor`/`SetSolidSize` 独立 setter（机制已 RE：opti @0x0A ARGB + sspc @0x20/0x24，今只在创建参数暴露）；导入的 solid footage 落 dest 根而非 Solids folder（AE 接受，仅整理性差异）。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **结构性创建 vein**（见 ↑下一步 1-4）— **当前主线**，纯代码可推。
2. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
3. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
