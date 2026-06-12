# Cockpit — aep-parser

**Last updated**: 2026-06-12 by claude（红线4 首案破案：orbit 渲染塌暗红 = Layer Styles placeholder tdsb bit0 误置 enabled，修 serializer flag words + orbit gate 升级渲染像素 gate，AE 2020+2025 PASS。详 `incidents/layer-styles-tdsb-enabled-bit.md`；同日前序（交付准则、mask op 全收口、marker Stable）见 `git log` + `archive/`。）
**Active focus**: **需求驱动期，无 active 主线**——大 arc 全收口：V3 框架（M1-M8，spec 已归档）· 结构性创建 vein（New\* 全家族 + AddEffect/SetEffectParam，全 Stable）· mask 结构性 op（Add/Remove/Duplicate/Move 全 Stable 双版本 gated）· Essential Graphics W（2026-06-12 ship）。ship-gate 自助（agent 跑 `scripts/ae_run.ps1` 双版本无人值守）。剩余候选见 ## 下一步；历史脉络靠 `git log` + `archive/` + coverage.md。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**无 active 主线**——大 arc 全收口（V3 框架 M1-M8 · 结构性创建 vein · EG W · mask 结构性 op · length-variable SetText 零 refuse · Alpha→Stable 审计批次；脉络见 `git log` + `archive/` + coverage.md）。以下候选均需求驱动，点名即开工：

1. **encodeBezier AE-native 字节**（纯优化）：写 AE-native 值（shph[3] open=0x09 等；@0x0C/@0x1C 待 n=5..8 RE 确证容量）→ shape/mask 写路径 byte-identical、mask patch 可化简。AE 已容忍现值（开放 path 双版本 gated），无紧迫性。详 `incidents/add-mask-create-re.md` §finding-4。
2. **EG W deferred 控件**：point/dropdown/text/Transform 源 controller + RemoveEssentialProperty。详 `incidents/essential-graphics-write-re.md`。
3. **mask 剩余写功能**（结构性 op 已全收口 Add/Remove/Duplicate/Move）：既有 mask 路径改写 `SetMaskPath`（值级，非结构性）/ animated mask path（om-s 多 shap + tdbs 时间表）/ mask mode·color·feather 创建参数化。详 `incidents/add-mask-create-re.md` §Deferred。
4. **各 vein 零散剩件**：SetEffectParam 扩库 + point 单位换算 helper · 导入 solid footage 归 Solids folder · RQ Set\* slice-5/6/7/8 仍 Alpha（render settings / OutputModule / item 级 setter，非结构性 in-place patch，按设计不走 ship-gate）。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。（shape 颜色渲染 RE 已破案收口 2026-06-12：根因 = Layer Styles placeholder tdsb bit0 全写 0x01 → AE 渲染 10 种 style 全开（红 solidFill+bevel 罩面），颜色编码本身无辜；修 `makeTdsbFlags`/`emptyPropGroupFlags` 按 AE 原生值（body 0x03 / fx 0x02），orbit gate 升级为渲染像素 gate（JSX 渲帧+Go 采样 5 点 ±8），AE 2020+2025 双版本 PASS。详 `incidents/layer-styles-tdsb-enabled-bit.md`。）
