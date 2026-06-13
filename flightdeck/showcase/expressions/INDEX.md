---
showcase: expressions
direction: 表达式激活 — 纯 Go 从零生成 time*N 旋转表达式驱动的轨道动画，渲染 t=1s 时三个点各被表达式转到不同角度
capabilities: [set-expression, set-expression-enabled, rotation-expression, time-driven]
gates: [TestExpression_AEShipGate_AE2020, TestExpression_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/expressions  +  scripts/ae_run.ps1 render.jsx (renders t=1s)"
---

# expressions — 表达式激活 showcase

## 这个方向测什么

三个「轨道」层，每层一个向上偏移 260px 的圆点；在层的 **Rotation 上挂 `time*N` 表达式**让点绕中心旋转。渲染 **t=1s** 时，三个不同转速（90/180/270 °/s）的点恰好转到不同角度——证明值是**表达式实时算的**、不是静态关键帧。灰色小 pip 标各轨道中心，便于读角度。表达式在 `Reopen` 后设到 `Rotation()`（`SetExpression` + `SetExpressionEnabled`），与 S2 ship-gate 同路。

> ⚠ 覆盖边界（交付准则）：S2 gate 只验过 `time*N` 这一类纯时间表达式。`loopOut`/`wiggle`/跨层引用等语汇**未** gate，本 showcase 不含。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `expressions.aep`（7 层 = 3 轨道点 + 3 中心 pip + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(1.0)` → t=1s |
| `expressions.aep` | 产出工程 (gitignored) | 1920×1080，30fps×4s，AE2020 target |
| `expressions.png` | 渲染帧 (gitignored) | **t=1s** |

## 布局（t=1s 时点相对各自中心 pip 的角度 = 表达式证据）

| 图层 | 颜色 | 中心 x | 表达式 | t=1s 角度 | 点应在中心的 |
|---|---|---|---|---|---|
| ORB_90 | 橙 | 480 | `time*90` | 90° | 右侧 |
| ORB_180 | 青 | 960 | `time*180` | 180° | 下方 |
| ORB_270 | 粉 | 1440 | `time*270` | 270° | 左侧 |

> 审核要点：点不在中心 pip 的正上方（初始偏移位）= 表达式已把旋转算到对应角度。

## 溯源

S2 表达式激活 RE（`expressionEnabled` 字节对）：`incidents/expression-enable-byte-pair.md`；gate：`expression_shipgate_test.go`。
