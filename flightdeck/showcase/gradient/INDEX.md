---
showcase: gradient
direction: 渐变填充 + ramp 方向 + ramp 类型 — 纯 Go 从零生成 4 种线性（横/纵/对角/三停）+ 2 种径向（2停/3停同心）渐变，验证 SetColorStops + Start/End Pt 方向 + SetGradientType(linear/radial)
capabilities: [gradient-fill, color-stops, multi-stop, ramp-direction, start-point, end-point, gradient-type, radial]
gates: [TestShapeGradient_AEShipGate, TestMGGradientDir_AEShipGate_AE2020, TestMGGradientDir_AEShipGate_AE2025, TestMGGradientRadial_AEShipGate_AE2020, TestMGGradientRadial_AEShipGate_AE2025]
status: 待review
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/gradient  +  scripts/ae_run.ps1 render.jsx"
---

# gradient — 渐变填充 + ramp 方向 + 类型 showcase

## 这个方向测什么

六个矩形（3×2）覆盖 **stop 颜色 + ramp 方向（Start/End Pt）+ 多停 + 类型（线性/径向）**：上排三种线性（横/纵/对角）+ 下排左三停线性、中/右两种径向同心。证明 `SetColorStops` 写颜色停、`SetStartPoint`/`SetEndPoint` 控方向、`SetGradientType(GradientLinear|GradientRadial)` 切线性↔径向（S5 gradient direction + 2026-06-14 radial type）。AE 2020 渲染 frame 0 → `gradient.png` 眼验。

> 覆盖边界：linear/radial 渐变 + 方向 + 多色停已双版本 gate。HiLite / G-Stroke 方向·类型 / type·direction read-back 仍 deferred（见 cockpit），本 showcase 不含。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `gradient.aep`（7 层 = 6 渐变 + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(0)` |
| `gradient.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| `gradient.png` | 渲染帧 (gitignored) | frame 0 |

## 布局（3×2）

| 位置 | 图层 | 应看到 | 控制 |
|---|---|---|---|
| 左上 | 01_Horizontal | 红→蓝 横向线性 | Start[-170,0] End[170,0] |
| 中上 | 02_Vertical | 品红→青 纵向线性 | Start[0,-130] End[0,130] |
| 右上 | 03_Diagonal | 琥珀→绿 对角线性 | Start[-170,-130] End[170,130] |
| 左下 | 04_ThreeStop | 红→黄→蓝 三停横向线性 | 3 stop @ 0/0.5/1 |
| 中下 | 05_Radial | 红心→蓝边 **同心径向** | Radial · 心[0,0] 半径[170,0] |
| 右下 | 06_RadialMulti | 白心→橙→紫边 三停**径向** | Radial · 3 stop |

> 审核要点：下排中/右两格应是**同心圆**（中心一色、向外环状过渡），不是横向带——这是 radial type 生效的证据。

## 溯源

渐变写入 + ramp 方向 + 类型 RE：`incidents/gradient-fill-write-re.md`；gate：`shape_gradient_shipgate_test.go`、`mg_gradient_dir_shipgate_test.go`、`mg_gradient_radial_shipgate_test.go`。
