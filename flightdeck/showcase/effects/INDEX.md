---
showcase: effects
direction: 加效果 + 改参数 — 纯 Go 从零建图形层 → AddEffect(内置效果库) + SetEffectParam 调参，AE 实渲逐格看可见效果差
capabilities: [add-effect, set-effect-param, gaussian-blur, drop-shadow, invert, tint, tritone, wave-warp, brightness-contrast, fractal-noise, gradient-ramp, mosaic, directional-blur]
gates: [TestAddEffect_AEShipGate_AE2020, TestAddEffect_AEShipGate_AE2025, "set_effect_param_shipgate_test.go (SetEffectParam 双版本)"]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/effects  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# effects — 加效果 + 改参数 showcase

## 这个方向测什么

4×3 网格，每格一个**相同源 token**（teal 方块 + amber 描边，给效果留「边缘 + 双色」作用面），分别施加一个内置效果并 `SetEffectParam` 调参。左上第一格是无效果参照，其余 11 格各一种效果。覆盖 `AddEffect`（30 效果库）+ `SetEffectParam`（scalar / color / angle / bool 各控件型）。`AddEffect` 需 parsed 层 → 生成器先建全部源层、`Reopen` 升级、再 splice 效果。AE 2020 渲染 frame 0 → `effects.png` 逐格对比眼验（红线4：看可见效果差，不靠值 round-trip）。每个效果 + `SetEffectParam` 本身已过双版本 ship-gate；本网格验证它们**组合 + 多效果实渲**可见。

## ⚠ 真实边界（诚实标注 · 交付准则）

- **不是所有 effect 参数都「即设即渲」**：HueSaturation 的 master hue（angle）`SetEffectParam` 设值后 **Go round-trip 绿、AE 却不应用色相**（frame 0 颜色无变化）= 假绿，已从 showcase 剔除、换成确定渲染的 Wave Warp。`SetEffectParam` 的「即设即用」覆盖**经 ship-gate 的参数**（Gaussian Blur / Drop Shadow / 表达式控制等），**不等于任意效果任意参数都被 AE 引擎应用**——这正是红线4d 的活体样本。
- **Mosaic 块数**是 control-type 1（无 generic 模板），`SetEffectParam` 调不了；用 AE 默认块网格（仍可见马赛克化）。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go, tracked) | 构建 effects.aep（13 层 = 12 效果格 + BG） |
| render.jsx | 渲染脚本(tracked) | AE `saveFrameToPng(0)` + dump 每效果参数名 → effects.png/.done |
| effects.aep | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| effects.png | 渲染帧 (gitignored) | frame 0 |

## 布局（4 列 × 3 行，逐格眼验）

| 格 | 图层 | 效果 + 参数 | 应看到 |
|---|---|---|---|
| 行1·列1 | 01_Source | 无（参照） | teal 方块 + amber 描边 |
| 行1·列2 | 02_GaussianBlur | Gaussian Blur 模糊度=30 | 边缘糊散开发光感 |
| 行1·列3 | 03_DropShadow | Drop Shadow 距离=30/方向=135°/柔和=14 | 右下方黑色投影 |
| 行1·列4 | 04_Invert | Invert（默认） | teal→红、amber→蓝（反色） |
| 行2·列1 | 05_Tint | Tint 着色量=100 | 去色成灰阶 |
| 行2·列2 | 06_Tritone | Tritone（默认） | 棕褐三色调映射 |
| 行2·列3 | 07_WaveWarp | Wave Warp（默认波形） | 上下边波浪起伏（几何扭曲） |
| 行2·列4 | 08_Brightness | Brightness&Contrast 亮+60/对比+40 | 整体变亮（亮青 + 亮黄边） |
| 行3·列1 | 09_FractalNoise | Fractal Noise（默认） | 黑白分形噪声填满 token |
| 行3·列2 | 10_GradientRamp | Gradient Ramp（默认） | 白→灰线性渐变填满 |
| 行3·列3 | 11_Mosaic | Mosaic（默认块网格） | 马赛克大色块 |
| 行3·列4 | 12_DirectionalBlur | Directional Blur 方向=45°/长度=35 | 对角方向拖影模糊 |

> 审核要点：逐格对比第一格 01_Source（未加效果）——每格的可见差异 = 该效果生效。中间 07_WaveWarp 的波浪边、04_Invert 的红蓝反色、03_DropShadow 的投影是最直观的「effect 真渲染」证据。

## 溯源

AddEffect + 30 效果库 + parade auto-create：`incidents/add-effect-splice-re.md`；SetEffectParam 参数物化 + 控件类型 + 假绿边界：`incidents/effect-param-elision-synthesis-lite.md`；gate：`add_effect_shipgate_test.go`、`set_effect_param_shipgate_test.go`。
