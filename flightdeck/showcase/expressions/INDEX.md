---
showcase: expressions
direction: 表达式激活 + 语汇 — 纯 Go 从零生成 time*N 旋转轨道（上排）+ 四语汇 idiom（下排：跨层引用 / loopOut / wiggle / slider effect-param 引用）
capabilities: [set-expression, set-expression-enabled, rotation-expression, time-driven, cross-layer-ref, loopout, wiggle, effect-param-ref, slider-control]
gates: [TestExpression_AEShipGate_AE2020, TestExpression_AEShipGate_AE2025, TestExprVocab_AEShipGate_AE2020, TestExprVocab_AEShipGate_AE2025]
status: 待review
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/expressions  +  scripts/ae_run.ps1 render.jsx (renders t=1s)"
---

# expressions — 表达式激活 + 语汇 showcase

## 这个方向测什么

**上排（time*N 轨道）**：三个「轨道」层各一个向上偏移的圆点，Rotation 挂 `time*N` 表达式让点绕中心 pip 旋转。t=1s 时三个转速（90/180/270 °/s）的点各转到 90°/180°/270°（右/下/左），证明值是表达式实时算的。

**下排（四语汇 idiom）**——每个 idiom **最好在 AE 里拖时间轴看运动**（png 只是 t=1s 快照）：

| 图层 | 颜色 | 表达式 | 拖时间轴看到 | t=1s 快照位置 |
|---|---|---|---|---|
| LINK | 青 | `thisComp.layer("LEAD").transform.position + [0,-110]` | 紧贴橙色 leader 上方 110px 同步移动 | leader 正上方 |
| LOOP | 绿 | `loopOut("cycle")`（位置关键帧 0..1s 上下 bob） | 每秒循环上下弹 | bob 区间内 |
| WIG | 品红 | `wiggle(3, 70)` | 围绕锚点抖动 | 偏离锚点 |
| SLD | 黄 | `[effect(1)(1), 820]`（层上 Slider Control = 1500） | 静止在 slider 值 x=1500 处 | x≈1500 |

（橙色 LEAD 是 LINK 的引用目标，自身横向扫动的关键帧动画。）

> ⚠ 覆盖边界（交付准则）：time*N / loopOut / wiggle / 跨层引用 / effect-param 引用五类语汇均已双版本渲染像素 gate（`expression_shipgate_test.go` + `expr_vocab_shipgate_test.go`）。effect-param 必须**按索引** `effect(1)(1)` 引用（实例名=match-name、参数名非"Slider"）。`linear()`/`ease()` 等仍按需补。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `expressions.aep`（12 层 = 3 轨道点 + 3 pip + LEAD/LINK/LOOP/WIG/SLD + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(1.0)` → t=1s |
| `expressions.aep` | 产出工程 (gitignored) | 1920×1080，30fps×4s，AE2020 target |
| `expressions.png` | 渲染帧 (gitignored) | **t=1s** 快照 |

> 审核要点：**打开 .aep 拖时间轴**——LINK 跟随、LOOP 循环弹、WIG 抖动、SLD 定在 slider 位、上排三点转速不同。这些是表达式实时求值的证据，单帧 png 看不全。

## 溯源

S2 表达式激活 RE（`expressionEnabled` 字节对）+ 语汇扩展（loopOut/wiggle/跨层/effect-param）：`incidents/expression-enable-byte-pair.md`；gate：`expression_shipgate_test.go`、`expr_vocab_shipgate_test.go`。
