# Cockpit — aep-parser

**Last updated**: 2026-06-12 by claude（**shape path open/closed 收口**：修 `decodeShapePath` closed 误判公共 API bug（shph[0x14]→shph[3]）+ 开放 shape path 双版本 ship-gate PASS（功能正确，AE 容忍 encodeBezier 偏差字节）；writer byte-faithfulness 降纯优化 follow-up。详 `incidents/add-mask-create-re.md` §finding-4）
**Active focus**: **结构性创建 vein（纯代码前沿）** — 结构性创建路径纯代码可推 + ship-gate 自助（agent 跑 `scripts/ae_run.ps1` 双版本无人值守）。2026-06-10 落：AddEffect（12 效果库）→ Camera/Light → parade auto-create + `aep.Reopen` → Solid/Null/Adjustment（详 `incidents/new-layer-types-scoping.md`）。2026-06-11 落：effect 全 12 模板双版本 gated + AddEffect/RemoveEffect 升 Stable + **NewTextLayer**（embed-whole-Layr，**新建图层类型全部建齐**；fresh 层免 Reopen 可读 TextSource/等长 SetText；btdk 改字长解封路径详 `incidents/text-btdk-length-variable-write-scoping.md`）。⚠ scar：Go-built 工程 allocItemID 撞 service 层 ID 2..12 → AE 2025 拒收，已修（详 `incidents/nextitemid-must-include-layer-ids.md`）。M8 物理分包已落。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split。M1-M8 框架已实现（结构性 mutation Phase 1-5 + 包重组① + M8 真·物理分包② 全落）；可达字段 ~99% ship，剩余前沿 fixture/RE-gated（详 plans/coverage.md）。 — [note: 大 arc 实质收尾（2026-06-09）—— M8 物理分包完成（design+plan 已归档）；剩 ValueText / Layr 3D 通道等 fixture-gated 边缘项，无 active 大 plan。]
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**结构性创建 vein 实质收口（2026-06-12）**——纯代码可推的前沿全部落地，剩余条目均需求驱动或 fixture-gated。候选下一步：AddMask/SetEffectParam Alpha→Stable（待用例积累）· 需求驱动条目（↓4/5）。

✅ **shape path open/closed parser bug 已修（2026-06-12，确证公共 API bug）**：dump AE-native `v2_2_shape_path_re.aep`（含开放 path）发现 `decodeShapePath` 用恒 0x01 的 shph[0x14] 读 closed → 公共 API `Layer.ShapePaths[].Closed` + JSON 对**所有**开放 shape path 误报 true，已改读 shph[3]（同 hydrate）+ 回归测试 + fixture 验证。同时**纠正上一轮**：AE-native shape path 的 closed/lhd3 字段与 mask 完全一致（shph[3] open=0x09 / @0x14=4 恒 / @0x18=1 恒），encodeBezier 是通用偏差非 mask 独有，只是 shape 侧 AE 容忍。详 `incidents/add-mask-create-re.md` §三 finding 4（含 ground-truth 表）。

✅ **开放 shape path 写功能性 = 双版本 ship-gated（2026-06-12）**：`TestV2_2_PathOpen_AEShipGate_AE2020/_AE2025` PASS（21.8s/20.1s）——AE 2020+2025 接受当前 encodeBezier 开放 path 偏差字节（shph[3]=0x00 等）、层未丢、顶点精确读回、AE resave 后开放标志存活。**即偏差值是 AE 容忍的纯 byte 差异、非功能 bug**，关闭「开放 shape path 写 AE 接受性未知」缺口。

📌 **降级为纯优化 follow-up（需求驱动）**：让 encodeBezier 写 AE-native 值（shph[3] open=0x09 / @0x14=4 / @0x18=1 确定；@0x0C/@0x1C 待 n=5..8 RE 确证 nextPow2 容量），shape/mask 写路径即 byte-identical、mask patch 可化简——但 AE 已容忍且 round-trip 功能正确，无紧迫性。

1. ~~**AddEffect 参数化**~~（✅ 完成：2026-06-11 `aep.SetEffectParam` synthesis-lite 双版本 gated；2026-06-12 泛型 per-control-type 模板**全 8 类**（scalar/enum/bool/angle/color/2D/3D/slider）——任意效果参数即设即用，免逐效果提取；效果库 30 全双版本 gated（#30 Point3D Control）。详 `incidents/effect-param-elision-synthesis-lite.md`）。剩余需求驱动：Alpha→Stable 待用例积累；再扩库（⚠ 引用参数效果 tdpi 指向非宿主层详 add-effect incident finding 5）；point 参数像素→分数便捷换算 helper（今为 raw on-disk 单位，编码已 doc）。
2. ~~**AddMask**~~（✅ 2026-06-11 落，超原计划：原以为只能 fixed-shape，实际 from-scratch atom 让**创建时路径任意参数化**（复用 shap 发射 + mask 专用字节修正），双版本 gate 4/4，Alpha 待用例积累升 Stable。剩余需求驱动：RemoveMask（atom 三件套不满足 pair 假设需专用实现）/ 既有 mask 路径改写 / animated mask path / mode·color 创建参数。⚠ follow-up：encodeBezier lhd3 @0x14/@0x1C 对 n≠4 顶点的 shape path 可能同样错（mask 实测掀出，shape 侧 gated fixture 或全是 n=4——需核）。详 `incidents/add-mask-create-re.md`）
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
