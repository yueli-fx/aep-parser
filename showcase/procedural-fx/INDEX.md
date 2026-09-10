---
showcase: procedural-fx
direction: 程序化 FX 生成器火焰(spec procedural-fx-generator)— 纯 Go 从零拼多层原生效果栈,AE 双版本实渲出有层次、翻腾的火焰；用户真机验收通过
capabilities: [fractal-noise, tritone, turbulent-displace, glo2-glow, mask-feather, concentric-mask, add-blend, multi-layer-composite, animate-effect-param, animate-effect-param-vec, adjustment-layer, effect-stack-recipe]
gates: [TestFlameDemo_AEShipGate_AE2020, TestFlameDemo_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-18
regenerate: "go run ./showcase/procedural-fx  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# procedural-fx — 程序化火焰 showcase（v3，多层合成，用户验收通过）

## 这个方向测什么

证 spec `procedural-fx-generator`:**库能不能确定性造出一个有层次、像样的火焰**。不开 AE,纯 Go 多层合成一条原生效果栈。**v3 = 多层合成出层次感**(技法 T3 additive-depth,`docs/fx-techniques.md`):

- **黑底** + **3 个火层**(`Fractal Noise(竖拉+高对比)→ Tritone(三档色温)→ Turbulent Displace`),各层**不同噪声 scale**(粗大火舌 / 中 / 细芯)。
- **同心 mask**(核层小 / 中层中 / 外层全)+ 各层 Tritone **越内越热** → 径向温度分区:**外深红 wispy → 中橙 → 内黄白热芯柱**。
- **Add 混合** → 重叠处叠出白热芯;顶部 **Glo2 Glow 调整层** → 泛光。
- 运动**共相**(所有层同速率 Evolution+Offset,关键帧)→ 翻腾上升、**不失相抖动**。

覆盖 `AddEffect`(FractalNoise/Tritone/TurbulentDisplace/Glo2)+ `SetEffectParam`(elided 物化、color [A,R,G,B])+ `AnimateEffectParam`/`AnimateEffectParamVec` + `AddMask`/`SetFeather` + `Layer.SetBlendingMode` + `NewAdjustmentLayer`。AI 永不碰字节;这条配方是将来 AI 层调的参数空间,这里手动设死。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `flame.aep`(1080×1920,AE2020 target,5 层:黑底+3 火层+Glow 调整层) |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng` → `flame.png`(强制 8bpc) |
| `flame.aep` / `flame.png` | 产出 (gitignored) | |

## 应看到

一个**有层次的火苗**:外圈深红 wispy 火舌 → 中橙 → 内黄白热芯柱,边缘羽化,黑底发光,在翻腾上升。t=1/t=3 内部纹理明显不同且**全程平稳无抖动**。

## gate 实测 + 用户验收(2026-06-18)

- **双版本 ship-gate PASS**:`TestFlameDemo_AEShipGate_AE2020/AE2025` 均绿,**两版本逐像素一致**(bright=2019 / warm=1843 / 两帧 diff=1446)。
- **用户真机验收:✅ 过了**(2026-06-18)。这是火焰从 v1 被否(单层、无色温/Glow)→ v2(单层修色温/Glow,层次不足)→ **v3 多层合成有层次** 的收尾。

## 历史(踩坑记录,详 `checklists/techniques/build-good-fire.md`)

- v1 被否:单层 Tint 2 色,缺色温/Glow/深度。
- v2:单层 Tritone+Glo2 修掉色温/辉光,但**单层无层次**(用户:层次来自多层混合)。
- v3:多层 Add 合成。两个关键踩坑→解法:① **团块**→同心 mask 分温度区;② **抖动**(~3s 起失相闪烁)→层运动共相、只静态属性分层。
