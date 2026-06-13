---
showcase: layers
direction: 图层创建 — 纯 Go 从零 NewSolidLayer（颜色+尺寸）堆成同心色框，并创建 Null/Adjustment 非渲染层
capabilities: [new-solid-layer, solid-color, solid-dimensions, new-null-layer, new-adjustment-layer, layer-stacking]
gates: [TestNewSolidNull_AEShipGate_AE2020, TestNewSolidNull_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/layers  +  scripts/ae_run.ps1 render.jsx"
---

# layers — 图层创建 showcase

## 这个方向测什么

从零创建多种图层类型。可见部分 = 四个 `NewSolidLayer`（各带颜色 + 尺寸，居中堆叠 → 同心色框），证明 Solid 的颜色/尺寸/堆叠顺序。另创建 `NewNullLayer` + `NewAdjustmentLayer`——这些是**非渲染层类型**，不出现在平面渲染里，但 `render.jsx` 的 .done 日志 dump 出层名 + 类型标志证明它们已建。

## ⚠ 边界（诚实标注）

- **可见渲染只有 Solid**。Null / Adjustment（以及未在此添加的 Camera / Light）是非渲染层——创建成功、列在层面板、但平面 frame 不显示。
- Solid 等图层的 `Position` 同样未物化（不可 `SetPosition`），故用**居中不同尺寸**堆出同心效果，而非摆位。
- Adjustment 层需配合 effect 才影响下方层；本 showcase 未加 effect，故它只是「已创建」。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `layers.aep`（4 Solid + Null + Adjustment = 6 层） |
| `render.jsx` | 渲染脚本(tracked) | 渲 frame 0 + dump 每层名/类型到 .done |
| `layers.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| `layers.png` | 渲染帧 (gitignored) | frame 0 |

## 布局

| 层（前→后） | 类型 | 尺寸 | 颜色 | 渲染 |
|---|---|---|---|---|
| Solid1_front | Solid | 560×320 | 琥珀 | 最前 |
| Solid2 | Solid | 1000×560 | 粉 | |
| Solid3 | Solid | 1440×800 | 蓝 | |
| Solid4_back | Solid | 1920×1080 | 暗 | 最后（满屏底） |
| Null_helper | Null | — | — | 非渲染（仅 dump 证实） |
| Adjustment_top | Adjustment | — | — | 非渲染（仅 dump 证实） |

> 审核要点：画面四层同心色框 = Solid 颜色+尺寸+堆叠生效；.done 日志列出 Null/Adjustment = 非渲染层已创建。Camera/Light/Precomp 见各自 incident / precomp-nesting showcase。

## 溯源

图层创建 RE：`incidents/new-layer-types-scoping.md`、`camera-light-layer-create-re.md`；gate：`new_solid_null_shipgate_test.go`。
