---
showcase: shape-primitives
direction: 形状图元 + 描绘算子 — 纯 Go 从零生成四种参数图元（Rect/Ellipse/Star/Rounded Rect）与 Fill/Stroke/Gradient 描绘变体并 AE 实渲
capabilities: [rect, ellipse, star, rounded-rect, fill, stroke, stroke-width, stroke-line-join, gradient-fill, gradient-ramp-direction]
gates: [TestNewShapeLayer_AEShipGate, "shape rect/ellipse/star gates", TestMGGradientDir_AEShipGate, "stroke line-cap/join gate"]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/shape-primitives  +  scripts/ae_run.ps1 render.jsx"
---

# shape-primitives — 形状图元 + 描绘算子 showcase

## 这个方向测什么

不开 AE，纯 Go 从零拼一个 4×2 网格：上排 = 四种**参数化图元**（矩形 / 椭圆 / 星形 / 圆角矩形），下排 = **描绘算子变体**（纯描边 / 填充+描边 / 圆角连接描边星 / 渐变填充）。覆盖 shape primitives 与 Fill/Stroke/Gradient 的从零组合 + 描边宽度/连接/渐变 ramp 方向。AE 2020 渲染 frame 0 → `shape_primitives.png` 逐格眼验。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `shape_primitives.aep`（9 层 = 8 格 + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(0)` → `shape_primitives.png` |
| `shape_primitives.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| `shape_primitives.png` | 渲染帧 (gitignored) | frame 0 |

## 布局（4 列 × 2 行）

| 格 | 图层 | 应看到 | 能力 |
|---|---|---|---|
| 行1·列1 | 01_Rect | 白矩形 | Rect + Fill |
| 行1·列2 | 02_Ellipse | 青椭圆 | Ellipse + Fill |
| 行1·列3 | 03_Star | 琥珀六角星 | Star 6pt |
| 行1·列4 | 04_RoundedRect | 粉圆角矩形 | Rect Roundness=45 |
| 行2·列1 | 05_StrokeOnly | 绿色空心环（无填充） | Ellipse + Stroke W=22 |
| 行2·列2 | 06_FillStroke | 琥珀方块 + 白描边 | Fill + Stroke |
| 行2·列3 | 07_StrokedStar | 粉色空心星（圆角连接） | Star + Stroke + LineJoin=Round |
| 行2·列4 | 08_GradientFill | 红→蓝对角渐变方块 | Gradient Fill + ramp Start/End Pt |

## 溯源

形状层创建：`incidents/v2-2-aelayer-structure.md`、`multi-layer-silent-drop.md`。描边枚举：`stroke-line-cap-join-miter-re.md`。渐变 + ramp 方向：`gradient-fill-write-re.md`。
