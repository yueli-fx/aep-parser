---
showcase: rain
direction: Native rain recipe showcase from shape/repeater streaks plus native mist, trim splashes, and echo persistence
capabilities: [shape path, fill, repeater, trim paths, layer position keyframes, precomp nesting, native effects]
gates: [TestRainShowcaseGeneratorBuildsNativeRainSpine]
status: 待review
last_updated: 2026-07-04
regenerate: "go run ./showcase/rain + AE render.jsx"
---

# rain — showcase

## 这个方向测什么

从 `build-good-rain.md` 的 stock-AE spine 出发，纯 Go 从零生成一个无第三方插件的 rain showcase。它不尝试等价复刻 Motionbox Rain Day 的 Trapcode Particular / Unmult 输出，而是验证可迁移的 native 雨线骨架：shape path + fill + repeater + 下落关键帧，再叠 native mist、Echo trail、wet-ground trim splashes。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go) | 纯 Go 构建 `rain.aep` |
| render.jsx | 渲染脚本 | AE saveFrameToPng at t=2.0s -> `rain.png` |
| rain.aep | 产出工程 (gitignored) | `RAIN` + `RAIN_STREAKS` 两个 comp |
| rain.png | 渲染帧 (gitignored) | 中间帧视觉审核图 |

## 布局

| 区域/层 | 期望效果 |
|---|---|
| `RAIN_STREAKS/NearFast` | 粗、亮、长的前景雨线，斜向快速下落 |
| `RAIN_STREAKS/MidSheet` | 中景密集雨幕，repeater copies 构成雨帘 |
| `RAIN_STREAKS/FineBack` | 背景细雨线，较暗较密 |
| `RAIN_STREAKS/BrightCuts` | 局部亮雨线，打破均匀网格 |
| `RAIN/NativeMist` | Fractal Noise 原生雾气，Add 混合 |
| `RAIN/EchoTrail` | Echo + Blur + Glow 给雨线拖尾和湿润感 |
| `RAIN/Splash*` | 地面椭圆 trim arc，表示落点水花 |

## 边界

- 这是 plugin-free native showcase，不是 Motionbox Rain Day exact clone。
- Exact reference render 仍需要 Trapcode Particular 与 Unmult。
- agent 渲染眼验后也只能保持 `待review`，等待用户真机复核。
