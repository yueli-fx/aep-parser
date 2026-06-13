---
showcase: comp-settings
direction: 合成设置 setter — Composition.Set*（运动模糊/工作区/背景色/嵌套帧率等），纯 Go 从零设值后 AE DOM readback 核对（📋 读值档，不看图）
capabilities: [comp-motion-blur, motion-blur-samples, work-area, bg-color, hide-shy-layers, preserve-nested-framerate]
gates: [shutter-side-effect-divisors, cdta-duration-two-representations]
status: 待review
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/comp-settings  +  scripts/ae_run.ps1 verify.jsx"
---

# comp-settings — 合成设置 setter showcase（📋 读值档）

## 这个方向测什么

`Composition.Set*` 合成级设置，纯 Go 从零设值后 **AE 打开 + DOM readback 核对**（这些设置无渲染视觉 → 读值不看图）。仅收**AE-DOM readback 完全一致**的 setter。

## 验证方式（📋 读值，非看图）

`verify.jsx` 打开 .aep，把 comp DOM 各字段 dump 到 `.done`，**你读日志核对下表**。无 png。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `comp_settings.aep` + 应用各 Set* |
| `verify.jsx` | 读值脚本(tracked) | dump comp DOM → `.done`（无 png） |
| `comp_settings.aep` | 产出工程 (gitignored) | 1920×1080，30fps×10s，AE2020 |
| `comp_settings.done` | readback 日志 (gitignored) | 用户读这个核值 |

## 设值 → 期望 readback（AE2020 实测一致）

| Set* 调用 | DOM 字段 | 期望值 |
|---|---|---|
| SetCompMotionBlur(true) | motionBlur | true |
| SetMotionBlurSamplesPerFrame(24) | motionBlurSamplesPerFrame | 24 |
| SetMotionBlurAdaptiveSampleLimit(192) | motionBlurAdaptiveSampleLimit | 192 |
| SetBGColor([40,20,80]) | bgColor | [0.157,0.078,0.314] |
| SetWorkArea(1.0, 6.0) | workAreaStart / workAreaDuration | 1 / 5 |
| SetHideShyLayers(true) | hideShyLayers | true |
| SetPreserveNestedFrameRate(true) | preserveNestedFrameRate | true |
| （NewComposition 30fps×10s） | frameRate / duration | 30 / 10 |

> 审核要点：`.done` 八行值与上表逐一对上即通过。

## ⚠ 已发现边界（诚实标注 · 本档刻意排除）

本 showcase 的 AE-DOM readback **首次**核到两个**之前仅字节 round-trip、从未 AE-DOM 验证**的 setter 有问题(红线4a「Go round-trip ≠ AE 接受」活样本)：

- **SetShutterAngle / SetShutterPhase**：设 172 / -86，AE DOM 读回 **206 / -103**（一致地 ×≈1.198）= 单位/编码未对齐。
- **SetResolutionFactor(2,2)**：AE 读 `comp.resolutionFactor` 抛**「数字结果无效（除以零）」** = 写入值让 AE 分辨率计算除零。

二者已从本档排除,**需独立 RE/修**(byte 写对但 AE 语义不对)。其余合成 setter（SetSize/SetName/SetFrameRate/SetDuration/SetPixelAspect/SetDisplayStartTime/SetDraft3D/SetFrameBlending 等）未纳入本批,按需再扩。

## 溯源

cdta 字段：`incidents/shutter-side-effect-divisors.md`（shutter 副作用）· `cdta-duration-two-representations.md`（duration/work-area）· 覆盖矩阵 `plans/coverage.md`。
