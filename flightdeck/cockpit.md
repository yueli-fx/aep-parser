# Cockpit — aep-parser

**Last updated**: 2026-06-13 by claude（**Wiggle Paths**（UI 名 Wiggle Paths，内部 match-name = `ADBE Vector Filter - Roughen`，发现型 canAddProperty 探针自证）从零 ship——矢量滤镜 vein **第 10 次**复用，**常用矢量滤镜家族完整收齐**。8 子流建模 4 headline：`Roughen Size`/`Roughen Detail`/`Temporal Freq`(=Wiggles/Second)/`Random Seed`（各 1D scalar，含 animated；Points enum/Correlation/两 Phase elide）。`AddWigglePaths`/`WigglePathsNode`（SetSize/SetDetail/SetWigglesPerSecond/SetRandomSeed，NewWigglePathsNode 默认 Size=0 identity）。`TestMGWiggle_AEShipGate_AE2020/2025` 双版本渲染像素 PASS：400×400 白 Rect + Size=60/Detail=30/WPS=4/Seed=9 → 毛糙噪声边，顶边逐列 topY spread=41（干净方块≈0）**两版完全一致=seed 决定性跨版可复现**·4 值 resave 全读回。新 ground truth：时间随机滤镜单帧由 seed+相位决定性。commit 753f1b3。当日 9 连：RoundCorners b9a4dc3·Offset 3e2ba12·Merge 959e904·ZigZag 55e9fff·gradient方向 aac0f12·PolyStar f5bf6e5·PuckerBloat 43e9c3f·Twist bfa1674·Wiggle Paths 753f1b3。）
**Active focus**: **From-scratch MG 工程能力 — 主线 S1–S6 已闭环 + 质感件全清**（`specs/2026-06-12-from-scratch-mg-roadmap.md`，用户终极目标「不开 AE 纯 Go 生成完整 MG 动画」端到端达成）。ease ✅ 表达式 ✅ Trim ✅ precomp ✅ Repeater ✅ Round Corners ✅ Offset Paths ✅ Merge Paths ✅ ZigZag ✅ gradient 方向 ✅ PolyStar ✅ Pucker&Bloat ✅ Twist ✅ Wiggle Paths ✅（**常用矢量滤镜家族 10 件完整收齐 + gradient 方向 + 星形**）端到端组合 gate ✅。**余下 = 按需 polish**（非阻塞）：表达式语汇 gate（loopOut/wiggle/跨层引用）· animated trim/repeater · precomp anchor/scale 参数化 · 唯一剩的矢量滤镜 Wiggle Transform（`ADBE Vector Filter - Wiggler`，带嵌套 Transform 组同 Repeater，需求驱动）· gradient 余项（radial/HiLite/G-Stroke 方向）· PolyStar Polygon 型。每 slice 渲染像素级双版本 gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1` 无人值守）。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**MG roadmap 主线 S1–S6 已闭环**（`specs/2026-06-12-from-scratch-mg-roadmap.md`）。下一步皆**按需点名即开**（无强制主线）：

1. **其余矢量滤镜/形状**（需求驱动，蓝本已**十次验证** Trim/Repeater/RoundCorners/Offset/Merge/ZigZag/PuckerBloat/Twist/WigglePaths + PolyStar 形状，常用家族完整收齐——详 `incidents/trim-paths-vector-filter-re.md`）：只剩 Wiggle Transform（`ADBE Vector Filter - Wiggler`，带嵌套 Transform 组同 Repeater）· PolyStar Polygon 型（需 Type slot + 独立模板）。
2. **表达式语汇 gate**：loopOut / wiggle / thisComp.layer 跨层引用（S2 仅验单 `time*90`）。
3. **gradient 余项**：Grad Type（radial）· HiLite · G-Stroke 方向（同 G-Fill vein，G-Stroke 模板需重抽）· direction read-back（hydrate）。
4. 技债（顺修）：`encodePathTimeTable` 容量分页同病（path >4 kf 前必修，详 `incidents/lhd3-keyframe-capacity-pages.md`）· animated trim/repeater（line-draw / count-up 真动画）+ precomp 多层嵌套已通未单独 gate · precomp anchor/scale 未参数化 · 种子模板 32bpc→8bpc 评估。

**需求驱动候选**（点名即开工）：encodeBezier AE-native 字节 · EG W deferred 控件 · mask 剩余写功能（SetMaskPath / animated path / mode·color·feather 参数化）· SetEffectParam 扩库 · RQ Set\* slice-5/6/7/8（Alpha by design）。

**仍 fixture/RE-gated（需外部输入）**：Layr Transform 3D 通道（需 3D layer 支持）· 暂搁项（environmentLayer / ligature / maskFeatherFalloff / CMS chunk 创建）· ValueText（schema-db）。详 `plans/coverage.md` § 暂搁 / 不可达。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3，与新建 Camera/Light 层相关）。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
