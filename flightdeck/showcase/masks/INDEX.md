---
showcase: masks
direction: 遮罩裁切 — 纯 Go 从零建彩色形状层 → AddMask（圆/三角/星/inverted）裁切，AE 实渲只露 mask 形状
capabilities: [add-mask, bezier-path, mask-inverted, mask-clipping]
gates: [TestAddMask_AEShipGate_AE2020, TestAddMask_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/masks  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# masks — 遮罩裁切 showcase

## 这个方向测什么

2×2 网格，每格一个彩色方块（320×320），各 `AddMask` 一个不同形状的 bezier path，AE 渲染时把方块裁成 mask 轮廓（mask 内露填充、mask 外透明露 BG）。证明 `aep.AddMask(layer, name, BezierPath)`（+ `SetInverted`）从零创建遮罩并裁切。坐标 = source-less（shape）层的 layer-local 像素（1:1 读回，无需换算）。AddMask 需 parsed 层 → 先建方块、`Reopen`、再 AddMask。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go, tracked) | 构建 masks.aep（5 层 = 4 遮罩方块 + BG） |
| render.jsx | 渲染脚本(tracked) | `saveFrameToPng(0)` + dump 每层 mask 名/inverted/closed/顶点数 → png/.done |
| masks.aep | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| masks.png | 渲染帧 (gitignored) | frame 0 |

## 布局（2×2，逐格眼验）

| 格 | 图层 | mask 形状 | 应看到 |
|---|---|---|---|
| 左上 | 01_Circle | 4 顶点 bezier 圆（kappa 切线） | teal **圆盘**（非方块） |
| 右上 | 02_Triangle | 3 顶点三角 | amber **三角形** |
| 左下 | 03_Star | 10 顶点 5 角星 | pink **五角星** |
| 右下 | 04_InvertedHole | inverted 方形 mask | green 方框中间**挖方洞**（透出 BG） |

> 审核要点：每个方块被裁成对应 mask 轮廓 = AddMask 生效；右下绿框中间透出暗 BG = `SetInverted` 生效（保留 mask 外、挖掉 mask 内）。

## 溯源

AddMask + bezier path + inverted + 坐标单位：`incidents/add-mask-create-re.md`；gate：`add_mask_shipgate_test.go`。
