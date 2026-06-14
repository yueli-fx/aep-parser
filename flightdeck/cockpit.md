# Cockpit — aep-parser

**Last updated**: 2026-06-14 by claude（**camera/light from-scratch option 全清 + 全部双版本 AE-DOM-gate**：Camera 13/13〔含 8 Iris*/Highlight* synthesis-insert〕+ Light 10/10〔含 Color synthesis-insert + Falloff/Shadow 收尾〕。Gotcha：Falloff Distance 默认 500，gate 用 750 避 elide。附带修 IrisHighlightSaturation 拼错 match-name latent bug〔Adobe 拼成 "Hightlight"〕。唯一剩 3D-render 像素深验。逐 commit 见 git log。）

**Active focus**: **From-scratch MG 工程能力 — 主线闭环 + 质感件全清，现处「按需 polish」期**（`specs/2026-06-12-from-scratch-mg-roadmap.md`）。常用 shape 矢量滤镜 11 件 + **gradient（方向/类型/高光，fill+stroke，写+读全闭环）** + 星形全收齐，端到端组合 gate ✅。工作流（2026-06-14 用户立规）：**每阶段写完即出 showcase**，agent 先**自验证**（render → Read png 对照意图）再给用户真机复核；gradient 与 expressions showcase 均 ✅ complete（用户真机过）。每 slice 渲染像素级双版本 gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1`）。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

无强制主线（MG roadmap S1–S6 闭环 + gradient/expressions showcase 收官）。下列皆**按需点名即开**：

**未修 RE 候选**：
- **EG 面板崩溃**（用户低优）——从零 EG 工程展开「基本图形」面板崩 AE（退回单 slider 也崩），DOM readback 假绿，根源 ship-gate 从不开面板。修法：补「真机开面板」gate + 比对 AE-native CIF3/CCtl/OvG2/CprC 字节。详 `incidents/essential-graphics-write-re.md`。

**polish / 技债**：
- **camera/light from-scratch option 已全清**（**Camera 13/13 + Light 10/10 全部双版本 AE-DOM-gate 实测 PASS**；含 Iris*/Highlight* + Light Color synthesis-insert + Falloff/Shadow 收尾，附带修 IrisHighlightSaturation 拼错 match-name latent bug）。**唯一剩**：3D-render 像素 gate（值已全程 DOM 验，但渲染像素深验需 3D layer 创建支持，本库未有）。`incidents/camera-light-layer-create-re.md`。
- **G-Stroke Dashes/Taper/Wave** 嵌套组（蓝本实心描边）· 各矢量滤镜 deferred 子流（`incidents/trim-paths-vector-filter-re.md`）。
- `encodePathTimeTable` 容量分页（path >4kf 前必修，`incidents/lhd3-keyframe-capacity-pages.md`）· animated trim/repeater · precomp anchor/scale 参数化。
- 表达式 `linear()`/`ease()` remap（边际值低，按需）。

**需求驱动**：encodeBezier AE-native 字节 · EG W deferred 控件 · mask 余项（**Feather/Opacity** 属性 + **SetMaskPath** / animated path；mode·color·inverted 等 mkif 字节已通）· SetEffectParam 扩库 · RQ Set\* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
