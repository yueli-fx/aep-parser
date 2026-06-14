# Cockpit — aep-parser

**Last updated**: 2026-06-14 by claude（gradient radial type 从零通（双版本旋转对称像素 PASS，模板重抽 15-child）· 表达式语汇 gate 四 idiom 双版本 PASS（slider「drop」误诊经 bisection 证伪）· 先前：nnhd display 家族修复 · SetLightKind 从零通。逐 commit 见 git log。）

**Active focus**: **From-scratch MG 工程能力 — 主线闭环 + 质感件全清，现处「按需 polish」期**（`specs/2026-06-12-from-scratch-mg-roadmap.md`）。所有常用 shape 矢量滤镜 11 件 + gradient 方向/**类型(radial)** + 星形全收齐，端到端组合 gate ✅。新工作流（2026-06-14 用户立规）：**每阶段写完即出 showcase 供真机审核**。gradient（+radial）与 expressions（+四语汇 idiom）showcase 已扩展，**🔍 待review**（等用户复核新内容）。每 slice 渲染像素级双版本 gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1`）。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

无强制主线（MG roadmap S1–S6 闭环 + showcase 收官）。下列皆**按需点名即开**：

**未修 RE 候选**：
- **EG 面板崩溃**（用户低优）——从零 EG 工程展开「基本图形」面板崩 AE（退回单 slider 也崩），DOM readback 假绿，根源 ship-gate 从不开面板。修法：补「真机开面板」gate + 比对 AE-native CIF3/CCtl/OvG2/CprC 字节。详 `incidents/essential-graphics-write-re.md`。
- **camera/light OPTION setter from-scratch elide**（Camera Zoom/Focus/Aperture · Light Intensity/Color/Cone…）——属性住 Options group，fresh layer 是 opaque 模板克隆无 scene tree，需 **property synthesis**（见 Backlog；`incidents/camera-light-layer-create-re.md`）。✅ **SetLightKind 从零已通**（2026-06-14，ldta @0x88，4 类 AE2020+2025 gate；模板默认其实是 parallel 非环境光）。

**polish / 技债**：
1. PolyStar Polygon 型（需 Type slot + 独立模板）· 各矢量滤镜 deferred 子流（蓝本 `incidents/trim-paths-vector-filter-re.md`）。
2. ✅ **表达式语汇 gate 已通**（2026-06-14）：loopOut（叠加在关键帧属性上）/ wiggle / 跨层引用 / **effect-param 引用（slider 绑定）四** idiom 双版本渲染像素 gate PASS，Go 零改动；`expr_vocab_shipgate_test.go`、详 `incidents/expression-enable-byte-pair.md`。slider idiom 一度误诊为「AddEffect 被 AE drop」，**bisection 证伪**——真因是表达式/JSX 须按**索引** `effect(1)(1)` 引用（效果实例名=match-name、参数显示名非"Slider"；按名引用 AE 静默回退静态值=假绿）。余项仅 `linear()`/`ease()` remap（边际值低，按需补）。
3. gradient 余项：✅ **Grad Type radial 已通**（2026-06-14，`TestMGGradientRadial_*` 双版本旋转对称像素 PASS；模板重抽 15-child 含 Grad Type/HiLite slot；enum 范围 [1,2] 非 1/3；`SetGradientType(GradientLinear|GradientRadial)`，详 `incidents/gradient-fill-write-re.md`）。余：HiLite 调参 · G-Stroke 方向/type · direction/type read-back。
4. `encodePathTimeTable` 容量分页（path >4kf 前必修，`incidents/lhd3-keyframe-capacity-pages.md`）· animated trim/repeater · precomp anchor/scale 参数化。✅ 种子模板 8bpc 已修（2026-06-14 576e78e：AE2020 seed 误为 32bpc，NewProject 现统一规整 8bpc=AE 默认）。

**需求驱动**：encodeBezier AE-native 字节 · EG W deferred 控件 · mask 剩余写（SetMaskPath / animated path / mode·color·feather）· SetEffectParam 扩库 · RQ Set\* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
