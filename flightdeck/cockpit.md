# Cockpit — aep-parser

**Last updated**: 2026-06-12 by claude（**SetText 变长零 refuse ship**：空串+多段落+多 run+手动 kerning 全解封。多 run=collapse-to-first-run、kerning=drop `/8` slot，均 AE 亲手产出 ground truth 证实（re_text_multirun / re_text_kern_resize）。gate：段落/空串 2020+2025、多 run+kerning 2024+2025 全 PASS（24+ API fixture 故 AE-24 stamped，2020 前向拒开 N/A）。length-variable SetText 不再拒任何输入。详 incident § 修法 v2/v3/v4）
**Active focus**: **需求驱动期，无 active 主线**——大 arc 全收口：V3 框架（M1-M8，spec 已归档）· 结构性创建 vein（New\* 全家族 + AddEffect/SetEffectParam/AddMask，全 Stable）· Essential Graphics W（2026-06-12 ship）。ship-gate 自助（agent 跑 `scripts/ae_run.ps1` 双版本无人值守）。剩余候选见 ## 下一步；历史脉络靠 `git log` + `archive/` + coverage.md。

## 进行中

<!-- AUTO:inprogress -->
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**主线已全收口（2026-06-12）**：结构性创建 vein + EG W + Alpha→Stable 审计批次全部落地，结构性写主力 API 全 Stable。**无 active 主线**——以下候选均需求驱动，点名即开工：

1. ~~**SetText 变长 refuse 集解封**~~ **全 ship（2026-06-12）= length-variable SetText 零 refuse**：空串 + 多段落 + 多 run + 手动 kerning 全解封。统一规律——整文本替换时 AE 把「跟旧字符耦合的东西」都丢/collapse：多 run→collapse 到 run[0] 样式、kerning→drop 整个 `/8` slot（`multirun_setvalue`/`kern_grow` = AE 亲手产出的 ground truth）。gate：段落/空串双版本（2020+2025）PASS；多 run + kerning **AE 2024+2025 双版本 PASS**（fixture 需 AE 24+ API 故 AE-24 stamped，AE 2020 前向拒开 N/A）。详 `incidents/text-btdk-length-variable-write-scoping.md` § 修法 v2/v3/v4。
2. **encodeBezier AE-native 字节**（纯优化）：写 AE-native 值（shph[3] open=0x09 等；@0x0C/@0x1C 待 n=5..8 RE 确证容量）→ shape/mask 写路径 byte-identical、mask patch 可化简。AE 已容忍现值（开放 path 双版本 gated），无紧迫性。详 `incidents/add-mask-create-re.md` §finding-4。
3. **残余 Alpha 的 gate 补齐**：RQ AddItem/RemoveItem · AddMarker · InsertLayer 家族（各有既注明的 gate 缺口/限制）。
4. **EG W deferred 控件**：point/dropdown/text/Transform 源 controller + RemoveEssentialProperty。详 `incidents/essential-graphics-write-re.md`。
5. **各 vein 需求驱动剩件**：RemoveMask / 既有 mask 路径改写 / animated mask path · SetEffectParam 扩库 + point 单位换算 helper · 导入 solid footage 归 Solids folder。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
