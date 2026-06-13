# Cockpit — aep-parser

**Last updated**: 2026-06-14 by claude（**showcase 全量收官**：15 个 🖼 可视方向用户真机验收全 ✅ complete；📋 B 类读值档 4 个 from-scratch 可行已补 🔍 待review（comp-settings / project-settings / camera-light / essential-graphics，verify.jsx dump DOM 值核对），3 个 from-scratch 不可表达（markers/render-queue 需 canonical seed、media-replace 需真 footage，已注明非缺陷=fixture-mutation 能力）。**B 类 readback 系统性查出多个「字节 round-trip 绿但 AE-DOM 不反映」的 settings setter**（comp SetShutterAngle/Phase ×1.2、SetResolutionFactor AE 除零；project nnhd byte8 SetTimeDisplayType/FeetFrames/FramesCountType；camera/light 选项 setter from-scratch 全 elide）→ 独立 RE/修候选。前序本日补 3 个 🖼 看图档——stroke-detail 重建[闭合星方案，弃塌缩的 open-path]7a7f63d · keyframe-channels[四通道关键帧渲 t=2s 插值] · animated-path[闭合 Path 几何 morph 横条→正方→竖条]。全部 AE2020 实渲眼验 + readback 数证，🔍 待review。**A 类可视方向 = 15 个全齐**；余下皆 B 类读值档（按需）。新发现 2 个假绿边界：open-path+stroke 渲染塌缩[stroke ship-gate 只验闭合 rect 值不验渲染]、Taper 闭合环无起止渲等宽[值绿无视觉]。上批 2026-06-13：**showcase 规约立 + 8 方向回填**：用户立规「大阶段（独立 plan/spec arc）必出从零示例供审核，小阶段攒批」——规则入 `rules.md` § Showcase + `checklists/showcase.md`。`flightdeck/showcase/<方向>/`（按能力方向分区，每区 tracked `INDEX.md`+`gen.go`+`render.jsx`，产物 `*.aep/*.png` gitignore，`go run` 可重生成）。已回填 8 方向全部 AE 实渲眼验：shape-filters / shape-primitives / keyframes-ease / expressions / precomp-nesting / gradient / text / layers。诚实标注 text·precomp·layers 的 from-scratch 边界（Position 未物化不可摆位、文字无颜色/字号 setter）。commit 238703c→314c07a。**前置**：**Wiggle Transform**（`ADBE Vector Filter - Wiggler`）从零 ship——矢量滤镜 vein **第 11 次**复用、**所有常用 shape 矢量滤镜全部收齐（vein 闭合）**。带嵌套 Transform 组（同 Repeater，经 findGroupBody descend）：顶层 `Xform Temporal Freq`(Wiggles/Sec)/`Random Seed`（1D scalar 含 animated）+ 嵌套组 4 通道抖动幅度 Anchor/Position/Scale(Vec2)/Rotation(1D 静态覆写)。`AddWiggleTransform`/`WiggleTransformNode`（+WigglerTransform 子，默认全零幅度=identity）。`TestMGWiggleTransform_AEShipGate_AE2020/2025` 双版本渲染像素 PASS：200×200 白 Rect 标称中心 (960,540) + Pos 幅度 [220,220]/Rot 70/Seed 8 → 白块质心位移到 (1092.9,404.6)、离中心 **189.7px 两版完全一致=seed 决定性跨版可复现**·6 值 resave 全读回。坑：Wiggler Scale 默认 [0,0]（抖动幅度非绝对值）。commit df0b874。当日 10 连：RoundCorners b9a4dc3·Offset 3e2ba12·Merge 959e904·ZigZag 55e9fff·gradient方向 aac0f12·PolyStar f5bf6e5·PuckerBloat 43e9c3f·Twist bfa1674·Wiggle Paths 753f1b3·Wiggle Transform df0b874。另：纯 Go 12-格滤镜 showcase 经 AE 实渲验收（test_data/demo_shape_filters.*）。）
**Active focus**: **From-scratch MG 工程能力 — 主线 S1–S6 已闭环 + 质感件全清**（`specs/2026-06-12-from-scratch-mg-roadmap.md`，用户终极目标「不开 AE 纯 Go 生成完整 MG 动画」端到端达成）。ease ✅ 表达式 ✅ Trim ✅ precomp ✅ Repeater ✅ Round Corners ✅ Offset Paths ✅ Merge Paths ✅ ZigZag ✅ gradient 方向 ✅ PolyStar ✅ Pucker&Bloat ✅ Twist ✅ Wiggle Paths ✅ Wiggle Transform ✅（**所有常用 shape 矢量滤镜 11 件全部收齐 = vein 闭合 + gradient 方向 + 星形**）端到端组合 gate ✅。**余下 = 按需 polish**（非阻塞）：表达式语汇 gate（loopOut/wiggle/跨层引用）· animated trim/repeater · precomp anchor/scale 参数化 · gradient 余项（radial/HiLite/G-Stroke 方向）· PolyStar Polygon 型。**矢量滤镜家族已无候选**。每 slice 渲染像素级双版本 gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1` 无人值守）。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-13-full-showcase-coverage.md](specs/2026-06-13-full-showcase-coverage.md) — 把所有已 ship 能力补齐 showcase 供用户逐个真机验收（代码过≠效果对）；已回填 8 个 from-scratch 可视方向，列出剩余缺口 + 非可视能力的验证方式，交接下个对话逐方向补
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**[当前焦点] showcase 全量补全 + 用户逐个真机校验**（`specs/2026-06-13-full-showcase-coverage.md`，2026-06-13 用户定「要所有示例」）：**🖼 看图档（A 类可视方向）已全部补齐 = 15 方向全 待review**。本批（2026-06-14）补 3 个：**stroke-detail**（重建，闭合星方案；Dashes/Join/Miter/Wave 实渲对，Cap·Taper 诚实暂缺 7a7f63d）· **keyframe-channels**（四通道 Position/Scale/Rotation/Opacity 关键帧，渲 t=2s 插值 + readback）· **animated-path**（闭合 Path morph 横条→正方→竖条，渲 t=2s + extent readback）。**待补皆 📋 读值档（B 类，按需点名即补，spec 定为用户驱动）**：render-queue / essential-graphics / markers / comp-settings / project-settings / camera-light / media-replace（dump 值 readback 不看图）。**关键发现（假绿边界，红线4d）**：① `SetEffectParam` 对未单独 gate 的参数可能假绿（HueSaturation master hue 物化绿但 AE frame0 不应用 → 换 WaveWarp）；② **open-path（AddPath+SetClosed(false)）+stroke 渲染塌缩**（几何 round-trip 绿但 AE 渲微小图形；3 个 stroke ship-gate 全建闭合 rect、只验值不验渲染像素、从不用 open path → 唯一渲染被证实的 stroke 几何 = 闭合形状；故 showcase 全用闭合星，Line Cap 需端点→暂缺）；③ **Taper 在闭合环无起止 → AE 渲等宽轮廓**（值 round-trip 但无视觉，实渲确认）。**solid 层不可摆位**（Position 未物化）→ 结构性 op showcase 用同心环 readback 佐证。**B 类用户真机 review 结果（2026-06-14）**：comp-settings / project-settings ✅ complete；camera-light 类型已更正（灯光=**环境光**非点光、相机=双节点默认值，待用户最终确认）；**essential-graphics 🛑 3 控件版展开「基本图形」面板崩溃 AE**（DOM readback 全过=假绿，EG ship-gate 只验 1 slider 且从不开面板）→ 已退回单 slider，**待用户重验单 slider 面板是否仍崩**。

**下一步**：① 用户重验 EG 单 slider 面板（崩→整个 from-scratch EG 退未验证；不崩→单 slider 成立、多控件标 RE）+ 确认 camera-light；② **setter/EG 边界 RE 候选（独立修）**：**EG 多控件/color/checkbox 面板崩溃**（CIF3/CCtl/OvG2/CprC，gate 需补「开面板」验证）· comp SetShutterAngle/Phase 单位（×1.2）· SetResolutionFactor AE 除零 · project nnhd byte8 位打包（SetTimeDisplayType/FeetFramesFilmType 互斥清位）+ SetFramesCountType · camera/light 选项 setter from-scratch elide（需 property synthesis 或扩模板）· 从零 SetLightType 缺失（默认只能环境光）；③ markers/render-queue/media-replace from-scratch 不可表达（需 seed/footage），要 showcase 须改 fixture-mutation 形态（暂不做）。

**MG roadmap 主线 S1–S6 已闭环**（`specs/2026-06-12-from-scratch-mg-roadmap.md`）。其余下一步皆**按需点名即开**（无强制主线）：

1. **形状/滤镜剩项**（矢量滤镜家族**已 11 件全收齐、vein 闭合无候选**——蓝本详 `incidents/trim-paths-vector-filter-re.md`）：只剩 PolyStar Polygon 型（需 Type slot + 独立模板）· 各滤镜 deferred 子流（Trim Type/ZigZag Points/Merge 其余模式/Wiggle Correlation·Phase 等，需求驱动）。
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
