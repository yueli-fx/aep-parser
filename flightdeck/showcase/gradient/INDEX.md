---
showcase: gradient
direction: 渐变填充 + ramp 方向 — 纯 Go 从零生成四种线性渐变（横/纵/对角/三停彩虹），验证 SetColorStops + Start/End Pt 方向几何
capabilities: [gradient-fill, color-stops, multi-stop, ramp-direction, start-point, end-point]
gates: [TestShapeGradient_AEShipGate, TestMGGradientDir_AEShipGate_AE2020, TestMGGradientDir_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/gradient  +  scripts/ae_run.ps1 render.jsx"
---

# gradient — 渐变填充 + ramp 方向 showcase

## 这个方向测什么

四个矩形各一种线性渐变填充，覆盖 **stop 颜色 + ramp 方向（Start/End Pt）+ 多停**：横向 / 纵向 / 对角 / 三停彩虹。证明 `SetColorStops` 写颜色停，`SetStartPoint`/`SetEndPoint` 控制渐变线方向（这是 S5 gradient direction 能力）。AE 2020 渲染 frame 0 → `gradient.png` 眼验方向与停色。

> 覆盖边界：linear 渐变 + 方向 + 多色停已 gate。radial 类型 / HiLite / G-Stroke 方向仍 deferred（见 cockpit 下一步），本 showcase 不含。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `gradient.aep`（5 层 = 4 渐变 + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(0)` |
| `gradient.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| `gradient.png` | 渲染帧 (gitignored) | frame 0 |

## 布局（2×2）

| 位置 | 图层 | 应看到 | 方向控制 |
|---|---|---|---|
| 左上 | 01_Horizontal | 红→蓝 横向渐变 | Start[-170,0] End[170,0] |
| 右上 | 02_Vertical | 品红→青 纵向渐变 | Start[0,-130] End[0,130] |
| 左下 | 03_Diagonal | 琥珀→绿 对角渐变 | Start[-170,-130] End[170,130] |
| 右下 | 04_ThreeStop | 红→黄→蓝 三停横向 | 3 个 color stop @ 0/0.5/1 |

## 溯源

渐变写入 + ramp 方向 RE：`incidents/gradient-fill-write-re.md`；gate：`shape_gradient_shipgate_test.go`、`mg_gradient_dir_shipgate_test.go`。
