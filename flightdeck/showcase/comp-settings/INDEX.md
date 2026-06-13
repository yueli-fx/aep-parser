---
showcase: comp-settings
direction: 合成设置 setter — Composition.Set*（运动模糊/工作区/背景色/嵌套帧率等），纯 Go 从零设值后 AE DOM readback 核对（📋 读值档，不看图）
capabilities: [comp-motion-blur, motion-blur-samples, work-area, bg-color, hide-shy-layers, preserve-nested-framerate]
gates: [shutter-side-effect-divisors, cdta-duration-two-representations]
status: complete
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
| SetShutterAngle(172) | shutterAngle | 172（**已修**，见下） |
| SetShutterPhase(-86) | shutterPhase | -86（**已修**） |
| SetCompMotionBlur(true) | motionBlur | true |
| SetMotionBlurSamplesPerFrame(24) | motionBlurSamplesPerFrame | 24 |
| SetMotionBlurAdaptiveSampleLimit(192) | motionBlurAdaptiveSampleLimit | 192 |
| SetBGColor([40,20,80]) | bgColor | [0.157,0.078,0.314] |
| SetWorkArea(1.0, 6.0) | workAreaStart / workAreaDuration | 1 / 5 |
| SetHideShyLayers(true) | hideShyLayers | true |
| SetPreserveNestedFrameRate(true) | preserveNestedFrameRate | true |
| （NewComposition 30fps×10s） | frameRate / duration | 30 / 10 |

> 审核要点：`.done` 各行值与上表逐一对上即通过。

## ✅ 已修复（2026-06-14，本档 readback 暴露的真 bug）

**SetShutterAngle/Phase ×1.2 假绿已根治**（incident `cdta-0xB0-shutter-ref-not-duration.md`）：根因不在 shutter setter,而是 **cdta @0xB0 历史误标为 duration**——实为「shutter 角度 360° 参考常量」。NewComposition 往 @0xB0 写真帧数(300=10s×30fps),AE 算 shutterAngle = stored×360/@0xB0 = 172×360/300 = 206.4 → 假绿。**连带发现**:parser 也从 @0xB0 读 duration → **误读所有真实 AE 工程的时长**(10s/30fps 读成 12s)。修复:parser 改读 @0x2C(MasterTicks=真时长 ticks),NewComposition 往 @0xB0 写常量 360。双版本未跑但 AE2020 实测 shutter 172/-86 + 5 个 AE-native 时长全对。

## ⚠ 仍存边界（本档刻意排除）

- **SetResolutionFactor(2,2)**：AE 读 `comp.resolutionFactor` 抛「数字结果无效（除以零）」——**AE 自身 scripting `comp.resolutionFactor=[2,2]` 也抛同错**,是 AE 侧深坑,非单纯 writer bug,待独立 RE。

其余合成 setter（SetSize/SetName/SetFrameRate/SetDuration/SetPixelAspect/SetDisplayStartTime/SetDraft3D/SetFrameBlending 等）未纳入本批,按需再扩。

## 溯源

cdta 字段：`incidents/shutter-side-effect-divisors.md`（shutter 副作用）· `cdta-duration-two-representations.md`（duration/work-area）· 覆盖矩阵 `plans/coverage.md`。
