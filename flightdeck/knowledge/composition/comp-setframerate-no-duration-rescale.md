# ⚠ comp SetFrameRate 不 rescale duration ticks (与 SetDuration 同 comp → AE 读错时长)

SUMMARY: comp SetFrameRate 不 rescale duration ticks (与 SetDuration 同 comp → AE 读错时长)
READ WHEN: 实现/调试 Composition.SetFrameRate 或 SetDuration;AE 打开后合成时长被 newFps/oldFps 缩放;给 comp 同时改帧率+时长;批量 comp-settings fixture

---

## Signature
- symptom: `AE comp.duration 读回被 newFps/oldFps 缩放;SetFrameRate(24)→SetDuration(8) 同 comp,AE DOM duration=6.41666 而非 8`
- error_type: —
- where: `internal/serializer/back_composition.go` SetFrameRate / SetDuration
- trigger: 同一 comp 先 SetFrameRate(改帧率) 再 SetDuration(设时长)

## 症状/复现

批6 comp-settings gate 勘出。`SetFrameRate(24)` 后 `SetDuration(8)`,Go round-trip 读回 Duration=8(自洽),但 **AE DOM `comp.duration` 读回 6.41666 = 8×24/30**(AE2025 + AE2020 双版本一致复现)。值被 `newFps/oldFps` 缩放。

## 根因

- `back.SetFrameRate(fps)` **只**写 fps 字段 cdta @0x9C(whole)/@0x9E(frac) + 更新 `b.FrameRateHz`。**不动** `ticksPerFrame@0x06`,**不 rescale** duration 的权威源 `MasterTicks@0x2C`。
- `back.SetDuration(sec)` 写 `MasterTicks@0x2C = round(sec × nominalTickRate)`,其中 `nominalTickRate = ticksPerFrame@0x06 × round(FrameRateHz)`。
- 序列:comp 建于 30fps → `ticksPerFrame@0x06` 是 30 基准。`SetFrameRate(24)` 改 fps 字段=24 但 @0x06 仍 30 基准。`SetDuration(8)` 用 `b.FrameRateHz=24` 算 `nominalTickRate = tpf(30基准) × 24`,写 ticks = 8 × tpf × 24。
- AE 读时按 @0x06(30 基准)还原:`8 × tpf × 24 / (tpf × 30) = 8 × 24/30 = 6.4`。完全吻合实测 6.41666。
- 旁证:`SetFrameRate` 自身也不 rescale 现有 duration ticks → 单独 `SetFrameRate(24)` 也会让 AE 把现有时长按新率重解释(批6 CfgF 只验 frameRate 值,未验 duration)。@0xB0 是 360 shutter 参考非 duration(见 [[cdta-0xB0-shutter-ref-not-duration]]);tickrate per-comp 见 [[tickrate-per-composition]]。

## 修法

- **workaround(批6 gate 采用,已 ship)**:在 comp 的**创建帧率**上设 duration(`NewComposition` 已带 fps 参数,@0x06 与 fps 一致),**勿创建后再改帧率**;若必须既测帧率又测时长,分到不同 comp(SetFrameRate→CfgF 单验,SetDuration→CfgA 在 30fps 单验)。
- **proper fix(未做,backlog)**:`SetFrameRate` 应同步 rescale `ticksPerFrame@0x06` + `MasterTicks@0x2C`(及可能的 work-area/displayStart 时间字段),使改帧率时**保持时长不变**(AE GUI 改帧率的语义)。改前须双版本 gate 复验所有时间字段。

## Cases
- 2026-06-17 首次 — 批6 comp-settings ae-accept gate(`TestCompSettings_AEShipGate_AE2020/AE2025`)。修法=解耦到不同 comp;SetDuration/SetFrameRate cap boundary 已加 ⚠ 互斥注 + 指本 incident。
