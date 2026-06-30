---
showcase: transform-values
direction: 改 transform 字段值 — 纯 Go 从零建形状层 → SetPosition/Scale/Rotation/Opacity 改值，AE 实渲逐格看变换
capabilities: [set-position, set-scale, set-rotation, set-opacity, transform-setter]
gates: [TestV2_2_XfKf_AEShipGate_AE2020, TestV2_2_XfKf_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/transform-values  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# transform-values — 改 transform 字段值 showcase

## 这个方向测什么

同一个源形状（teal 方块 + amber 描边），在网格每格施加不同 Layer transform 值（Position / Scale / Rotation / Opacity），渲染 frame 0 一眼看出每格变换。证明 `l.Position()/.Scale()/.Rotation()/.Opacity().SetStaticValue(...)` 改值在 **AE 渲染中真生效**（不是只值 round-trip）。落盘单位：Scale/Opacity = 百分比（200=200%、40=40%），Rotation = 度。**全 6 格 AE 实渲眼验通过**。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go, tracked) | 构建 transform_values.aep（7 层 = 6 变体 + BG） |
| render.jsx | 渲染脚本(tracked) | `saveFrameToPng(0)` + dump 每层 transform 五通道值 → png/.done |
| transform_values.aep | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| transform_values.png | 渲染帧 (gitignored) | frame 0 |

## 布局（同一方块的 transform 变体，逐格眼验）

| 格 | 图层 | 变换 | 应看到 |
|---|---|---|---|
| 行1·列1 | 1_Source | 默认 | 标准方块（基准） |
| 行1·列2 | 2_Scale200 | Scale 200% | 放大方块 |
| 行1·列3 | 3_Scale50 | Scale 50% | 缩小方块 |
| 行2·列1 | 4_Rotate45 | Rotation 45° | 菱形（方块转 45°） |
| 行2·列2 | 5_Opacity40 | Opacity 40% | 半透明暗方块（融入 BG） |
| 行2·列3 | 6_MovePos | Position 偏移 | 偏离格中心的方块 |

> 审核要点：对照 1_Source 基准——放大 / 缩小 / 旋转菱形 / 半透 / 移位逐格可见 = transform setter 渲染生效。

## 溯源

Layer transform 5 通道（Position/Anchor/Scale/Rotation/Opacity）持久化：`flightdeck/plans/coverage.md` § V2.2.1 子项⑥；gate：`TestV2_2_XfKf_AEShipGate_*`（AE 读回 user 单位 scale/rot/opacity 正确）。
