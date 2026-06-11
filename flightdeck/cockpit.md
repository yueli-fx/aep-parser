# Cockpit — aep-parser

**Last updated**: 2026-06-12 by claude（**encodeBezier lhd3 n≠4 follow-up 核查（纯代码 negative finding）**：shape path 侧不受 mask n≠4 lhd3 bug 影响——静态 path gate 用 n=3 三角形 + 精确顶点读回、动画 gate frame 2 亦 n=3，双版本 PASS；AE 对 shape om-s 容忍 encodeBezier 的 n-依赖 lhd3，急切-decode 严格性 mask 独有。落 add-mask + path-keyframe 两 incident + 收 cockpit 候选。前一更新 = **SetEffectParam 控件类型补齐**：泛型 per-control-type 模板从 3 类扩到全 8 类（+angle/color/2D/3D/slider，touch-all fixture 一次提取）+ Point3D Control 扩库 #30，双版本 gate 13 项全 PASS（含 cross-effect color/angle pard-patch）。RE：point 参数 cdat = 层坐标空间分数（有源=源尺寸、source-less=comp 尺寸、z 除 height）；color cdat = ARGB×255。⚠ 坑：ExtendScript `""+colorValue` 拼接抛「除以零」——首轮 gate 假 reject，bisect 整文件才发现是 verify JSX 日志行炸了（陷阱已记 re-fixture.md）。详 `incidents/effect-param-elision-synthesis-lite.md`）
**Active focus**: **结构性创建 vein（纯代码前沿）** — 结构性创建路径纯代码可推 + ship-gate 自助（agent 跑 `scripts/ae_run.ps1` 双版本无人值守）。2026-06-10 落：AddEffect（12 效果库）→ Camera/Light → parade auto-create + `aep.Reopen` → Solid/Null/Adjustment（详 `incidents/new-layer-types-scoping.md`）。2026-06-11 落：effect 全 12 模板双版本 gated + AddEffect/RemoveEffect 升 Stable + **NewTextLayer**（embed-whole-Layr，**新建图层类型全部建齐**；fresh 层免 Reopen 可读 TextSource/等长 SetText；btdk 改字长解封路径详 `incidents/text-btdk-length-variable-write-scoping.md`）。⚠ scar：Go-built 工程 allocItemID 撞 service 层 ID 2..12 → AE 2025 拒收，已修（详 `incidents/nextitemid-must-include-layer-ids.md`）。M8 物理分包已落。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split。M1-M8 框架已实现（结构性 mutation Phase 1-5 + 包重组① + M8 真·物理分包② 全落）；可达字段 ~99% ship，剩余前沿 fixture/RE-gated（详 plans/coverage.md）。 — [note: 大 arc 实质收尾（2026-06-09）—— M8 物理分包完成（design+plan 已归档）；剩 ValueText / Layr 3D 通道等 fixture-gated 边缘项，无 active 大 plan。]
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**结构性创建 vein 实质收口（2026-06-12）**——纯代码可推的前沿全部落地，剩余条目均需求驱动或 fixture-gated。候选下一步：AddMask/SetEffectParam Alpha→Stable（待用例积累）· 需求驱动条目（↓4/5）。

✅ **encodeBezier lhd3 n≠4 核查已了结（2026-06-12，纯代码 negative finding）**：shape path 侧**不受** mask 那个 n≠4 lhd3 bug 影响——静态 path gate 本就用 n=3 三角形 + 精确顶点读回、动画 gate frame 2 亦 n=3，双版本 PASS；AE 对 "ADBE Vector Shape" om-s 容忍 encodeBezier 的 n-依赖 lhd3，mask 的急切-decode 严格性是 mask 独有。未 gate 的次要轴仅「开放 shape path 的 @0x18=0」（需求驱动）。详 `incidents/add-mask-create-re.md` §三 finding 4 + `incidents/path-keyframe-write-re.md` 末节。

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
