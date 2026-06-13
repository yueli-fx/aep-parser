---
showcase: keyframes-ease
direction: 时间缓动关键帧 — 纯 Go 从零生成同跨度位移动画的 linear / ease-out / ease-in-out 三种缓动，渲染中间帧让缓动表现为单帧内的水平位置差
capabilities: [position-keyframes, linear-interpolation, temporal-ease, ease-out, ease-in-out, multi-keyframe-stream]
gates: [TestMGEase_AEShipGate_AE2020, TestMGEase_AEShipGate_AE2025]
status: 待review
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/keyframes-ease  +  scripts/ae_run.ps1 render.jsx (renders t=2s)"
---

# keyframes-ease — 时间缓动关键帧 showcase

## 这个方向测什么

三个圆点从 x=200 跑到 x=1700（4 秒、同跨度），分别用 **linear / ease-out（起点慢出）/ ease-in-out（两端缓）** 三种时间缓动。关键在于：**渲染正中时刻 t=2s**，缓动就在单帧里变成可见的水平位置差——线性与对称 ease-in-out 落在中点 x≈950，而「起点慢出」的点严重滞后在左侧。这正是缓动「不是匀速」的视觉证据（红线4：渲染中帧验位置，不靠值 round-trip）。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `keyframes_ease.aep`（4 层 = 3 动画点 + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(2.0)` → t=2s 帧 |
| `keyframes_ease.aep` | 产出工程 (gitignored) | 1920×1080，30fps×4s，AE2020 target |
| `keyframes_ease.png` | 渲染帧 (gitignored) | **t=2s（中间帧）** |

## 布局（t=2s 时三点的位置 = 缓动证据）

| 图层 | 颜色 | 行 y | 缓动 | t=2s 期望位置 | 视觉读法 |
|---|---|---|---|---|---|
| LIN_linear | 橙 | 300 | 线性两帧 | x≈950（正中） | 匀速 → 跨度中点 |
| EOUT_easeOut | 青 | 540 | 起点 ease-out（influence 0.9） | x≪950（严重靠左） | 慢启动 → 中时刻还没到中点 |
| EIO_easeInOut | 粉 | 780 | 两端 ease-in-out（influence 0.6） | x≈950（正中） | 对称缓动 → 中时刻仍在中点，但首尾更慢 |

> 审核要点：橙、粉在画面水平中心，青点明显偏左 = 缓动生效（匀速三点会三点共线在中心）。

## 溯源

S1 ease 关键帧 + >2 关键帧规模 RE/gate：`mg_ease_shipgate_test.go`；关键帧字节布局：`incidents/keyframe-byte-layout-dispatcher.md`、容量分页 `lhd3-keyframe-capacity-pages.md`。
