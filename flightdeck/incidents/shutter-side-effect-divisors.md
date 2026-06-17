---
status: active
when_to_read: implementing Comp.SetShutterAngle / SetShutterPhase; debugging unexpected cdta @0x18/@0x20/@0x28 changes after shutter setter
applies_to: [composition, shutter, cdta, work-area, ae-cosmetic, side-effect]
last_updated: 2026-05-22
---

# shutter setter 在 AE 端触发 work-area divisor 重编码

## 现象

JSX 设 `comp.shutterPhase = -90` 或 `comp.shutterAngle = 360`，AE 保存 .aep 后 cdta 不只 0xB4..0xB7（phase）/ 0xAE..0xAF（angle）的真值字节变，**外加** work-area divisor 三处（@0x18 / @0x20 / @0x28）从默认 `0x00000258` (600) 重编码为 `0x00005DA8` (24008)。

`re_cdta_probe.aep` 实证：

```
A_baseline (默认):
  0x20..0x23 = 00 00 02 58 = 600   (workAreaStart divisor)
  0x28..0x2B = 00 00 02 58 = 600   (workAreaEnd divisor)

E_shutter_phase_neg90:
  0x20..0x23 = 00 00 5D A8 = 24008 ❗
  0x28..0x2B = 00 00 5D A8 = 24008 ❗

F_shutter_angle_360:
  0x20..0x23 = 00 00 5D A8 = 24008 ❗
  0x28..0x2B = 00 00 5D A8 = 24008 ❗
```

## 原因

AE 内部在精确表示 shutter 的分数角度时挑了更细的 tick base。`24008 / 600 ≈ 40.01`，跟 NTSC frame rate (29.97) 相关的 LCM。属于 AE 自己的 numeric optimization。

## 对我们 setter 的影响

我们的 `Composition.SetShutterPhase` / `SetShutterAngle` 只写 phase/angle 真值 byte（@0xB4 / @0xAE），**不复制** divisor 重编码。

测过：roundtrip 跑通 + AE 重新打开我们写的 .aep 行为正确（既有测试 `TestCompositionSetters` PASS）。所以 divisor 重编码是 AE 的 cosmetic optimization，不是语义必要。**保持现状**，不为此引入 setter 内的 divisor 重写逻辑。

## 提示

要复现/验证别的 setter 是否有类似 AE-side re-encoding，做法：

1. 写 JSX 在新 comp 上 set 单一属性
2. 跑 `go run ./tools/debug/dump_cdta probe.aep` 对比 baseline
3. 看是否有目标 byte 之外的 byte 也变了
4. 如果"额外变化字节"是 0/1/divisor 这种结构性数字 → 多半是 AE 内部 re-encoding，**别学 AE**；保持 setter 单点写

## 关联

- `re_cdta_probe.{jsx,aep}` 是这块的实证 fixture（永久保留）
- `CompItem.resolutionFactor` 跟 shutter 互不干扰（设它不触发 divisor 重编码）
