---
showcase: procedural-fx
direction: 程序化 FX 生成器 v1 火焰(spec procedural-fx-generator Phase 0)— 纯 Go 从零拼一条原生效果栈,AE 双版本实渲出可辨认的、翻腾的火焰
capabilities: [fractal-noise, tint, turbulent-displace, mask-feather, animate-effect-param, effect-stack-recipe]
gates: [TestFlameDemo_AEShipGate_AE2020, TestFlameDemo_AEShipGate_AE2025]
status: 待review
last_updated: 2026-06-18
regenerate: "go run ./flightdeck/showcase/procedural-fx  +  scripts/ae_run.ps1 render.jsx"
---

# procedural-fx — 程序化 FX 生成器 v1 火焰 showcase

## 这个方向测什么

证 spec `procedural-fx-generator` 的命门(Phase 0):**库能不能确定性造出一个一眼看得出是火焰的东西**。不开 AE,纯 Go 在一个固态层上叠一条 AE 原生效果栈:

`Fractal Noise(竖向拉伸+高对比)→ Tint(黑→暗红 / 白→橙黄)→ Turbulent Displace(有机扰动)+ 羽化水滴形 mask(火苗轮廓)`,再给 Fractal Noise / Turbulent Displace 的 **Evolution 打关键帧** → 火焰翻腾。

AI 永不碰字节:这条配方将来是 AI 层调的参数空间,这里每个旋钮手动设死。覆盖 `AddEffect` + `SetEffectParam`(含 elided param 物化、color [A,R,G,B] 0-255)+ `AnimateEffectParam`(从零给效果参数打关键帧,实测 AnimateEffectParam 未撞 elision 缺口)+ `AddMask`/`SetFeather`。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `flame.aep`(1080×1920 竖构图,AE2020 target,单 Flame 层) |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(2.0)` → `flame.png`(强制 8bpc) |
| `flame.aep` | 产出工程 (gitignored) | |
| `flame.png` | 渲染帧 (gitignored) | t=2s 中段 |

## 应看到

一个**水滴形火苗**:下宽上尖,竖向橙黄火舌 + 黑色间隙,边缘羽化柔和,透明底。t=1 与 t=3 两帧内部纹理明显不同(在翻腾)。

## gate 实测(2026-06-18)

`TestFlameDemo_AEShipGate_AE2020/AE2025` 双版本 PASS:bright≈1473 / warm≈1141 / 两帧 diff≈1472(运动活)。agent 已 Read png 眼验像火焰。**status=待review:等用户真机打开 flame.aep 复核火焰质量后才翻 complete**(agent 眼验 ≠ 用户验收)。

## 已知可改进(Phase 1 参数化时)

底部更红、芯部更亮、加 Glow 泛光会更像真火;当前是「可辨认火焰」的最小可行版,够证命门。
