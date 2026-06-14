---
showcase: gradient
direction: 渐变填充 + 描边 + ramp 方向/类型/高光 — 纯 Go 从零生成 3×3：线性填充(横/纵/对角) · 径向填充(2停/3停/+HiLite高光) · 渐变描边(线性/径向/对角三停)，验证 SetColorStops + Start/End Pt 方向 + SetGradientType(linear/radial) + SetHighlightLength/Angle，G-Fill 与 G-Stroke 全 parity
capabilities: [gradient-fill, gradient-stroke, color-stops, multi-stop, ramp-direction, start-point, end-point, gradient-type, radial, hilite, highlight]
gates: [TestShapeGradient_AEShipGate, TestMGGradientDir_AEShipGate_AE2020, TestMGGradientDir_AEShipGate_AE2025, TestMGGradientRadial_AEShipGate_AE2020, TestMGGradientRadial_AEShipGate_AE2025, TestMGGradientHilite_AEShipGate_AE2020, TestMGGradientHilite_AEShipGate_AE2025, TestMGGradStrokeGeom_AEShipGate_AE2020, TestMGGradStrokeGeom_AEShipGate_AE2025, TestV2_2_GradientStroke_AEShipGate_AE2020, TestV2_2_GradientStroke_AEShipGate_AE2025]
status: 待review
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/gradient  +  scripts/ae_run.ps1 render.jsx"
---

# gradient — 渐变填充 + 描边 + 方向/类型/高光 showcase

## 这个方向测什么

九个矩形（3×3）覆盖 **stop 颜色 + 方向（Start/End Pt）+ 类型（线性/径向）+ 径向高光（HiLite）+ 填充 vs 描边**：

- **第一排 线性填充**：横 / 纵 / 对角，证 `SetStartPoint`/`SetEndPoint` 控方向。
- **第二排 径向填充（+高光）**：2停同心 / 3停同心 / 2停**+HiLite**（亮心右移），证 `SetGradientType(radial)` + `SetHighlightLength/SetHighlightAngle`。
- **第三排 渐变描边**：线性横向 / 径向 / 对角三停，证 `GradientStrokeNode` 的方向+类型与 G-Fill 完全 parity（18px 描边环）。

证明 `SetColorStops` 写颜色停、方向/类型/高光 setter 在 **G-Fill 与 G-Stroke** 上都生效（S5 gradient direction + 2026-06-14 radial type / HiLite / G-Stroke geom）。AE 2020 渲染 frame 0 → `gradient.png` 眼验。

> 覆盖边界：linear/radial 渐变 + 方向 + 多停 + **HiLite 高光** + **G-Stroke 方向/类型** 均已双版本像素 gate。G-Stroke HiLite 独立像素 gate（边际，机制已被 G-Fill 证）/ type·direction·highlight read-back / stroke geometry(Width/Cap/…) 仍 deferred（见 cockpit），本 showcase 不含。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `gradient.aep`（10 层 = 9 渐变 + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(0)` |
| `gradient.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| `gradient.png` | 渲染帧 (gitignored) | frame 0 |

## 布局（3×3）

| 位置 | 图层 | 应看到 | 控制 |
|---|---|---|---|
| 左上 | 01_Horizontal | 红→蓝 横向线性填充 | Start[-150,0] End[150,0] |
| 中上 | 02_Vertical | 品红→青 纵向线性填充 | Start[0,-105] End[0,105] |
| 右上 | 03_Diagonal | 琥珀→绿 对角线性填充 | Start[-150,-105] End[150,105] |
| 左中 | 04_Radial | 红心→蓝边 **同心径向** 填充 | Radial · 心[0,0] 半径[150,0] |
| 中中 | 05_RadialMulti | 白心→橙→紫边 三停**径向**填充 | Radial · 3 stop |
| 右中 | 06_RadialHiLite | 红心→蓝边径向，但**亮心右移** | Radial · HiLite Length70 Angle0 |
| 左下 | 07_StrokeLinear | 红→蓝 横向线性 **描边环** | G-Stroke · Start[-150,0] End[150,0] |
| 中下 | 08_StrokeRadial | 金→紫 **径向描边环** | G-Stroke · Radial · 半径[220,0] |
| 右下 | 09_StrokeDiag3 | 红→黄→蓝 对角三停 **描边环** | G-Stroke · 3 stop 对角 |

> 审核要点：① 04 与 06 对比——06 的亮心明显**偏右**（HiLite 高光生效，04 居中）；② 第三排是**描边环**（中心镂空），环上的颜色随方向/类型变化（07 左红右蓝、08 金紫径向对称、09 对角彩虹），证 G-Stroke 与 G-Fill 同源。

## 溯源

渐变写入 + 方向 + 类型 + HiLite + G-Stroke geom RE：`incidents/gradient-fill-write-re.md`；gate：`shape_gradient_shipgate_test.go`、`mg_gradient_dir_shipgate_test.go`、`mg_gradient_radial_shipgate_test.go`、`mg_gradient_hilite_shipgate_test.go`、`mg_gradstroke_geom_shipgate_test.go`、`shape_gradient_stroke_shipgate_test.go`。
