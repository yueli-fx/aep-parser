---
status: active
when_to_read: 调查 comp.Duration 与真实 AE 不符；AE 打开 Go 生成工程后 shutterAngle/shutterPhase 显示被放大（×duration/360）；以为 cdta @0xB0 是 duration；改 NewComposition / SetDuration / SetFrameRate 的时长写法；新增任何依赖 cdta 时间基的 setter
applies_to: [composition, cdta, duration, shutter-angle, shutter-phase, 0xB0, 0x2C, master-ticks, nominal-tickrate, parser-misread, false-green, new-composition, set-duration]
last_updated: 2026-06-14
supersedes: 部分修正 cdta-duration-two-representations.md（@0xB0 不是 duration）
---

# cdta @0xB0 是 shutter 角度 360° 参考常量，**不是 duration**

## 结论（RE 实证，2026-06-14）

cdta `@0xB0`（旧误标 `CdtaDuration`）+ 镜像 `@0xB8` 在**真实 AE 工程里恒为常量 360**，与时长无关。它是 **shutter 角度的 360° 参考**：AE 显示 `shutterAngle = stored@0xAE × 360 / @0xB0`。

**真实时长**在 `@0x2C`（旧名 `CdtaMasterTicks`）= `round(duration_seconds × nominalTickRate)`，其中 `nominalTickRate = ticksPerFrame@0x06 × round(fps)`。`duration = @0x2C / nominalTickRate`。

## 两个被它害的 bug（都是「自洽但与 AE 不符」的假绿）

1. **parser 误读所有真实 AE 工程时长**：旧 parser `Duration = @0xB0_frames / fps`。真实 AE @0xB0=360 → 10s/30fps 工程被读成 `360/30 = 12s`；4s/24fps → 15s；等。RE 实证 5 个 AE addComp 工程（dur 10/5/4/8s、fps 30/60/24/25）@0xB0 **全 = 360**，真时长全在 @0x2C（10s→307200=30720×10、5s→153600）。
2. **NewComposition 破坏 AE 的 shutter 显示**：旧 NewComposition 往 @0xB0 写真帧数（300=10s×30fps）。AE 打开后 `shutterAngle = 172 × 360/300 = 206.4`（显示 206），phase 同理 ×1.2。**shutter 字节本身（@0xAE/@0xB4）写得对**——victim 是 @0xB0。

为何长期没暴露：我们的 NewComposition 写 @0xB0=真帧数，parser 也读 @0xB0 → **自生成文件自洽**（写 300 读 300/30=10s）；只有跟真实 AE 对账（showcase comp-settings 的 shutter readback + AE 存盘工程时长）才暴露。

## 修复（已落地）

- `parse_composition.go`：`Duration` 改读 `@0x2C / nominalTickRate`；**@0xB0 仅作 legacy/synthetic fallback**（@0x2C=0 或 @0x06 缺失时）。
- `mutate_composition_new.go`：@0xB0/@0xB8 写**常量 360**（不再写帧数）；@0x2C 已正确写 `round(dur × NominalTickRate)`。
- `back_composition.go` SetDuration：写 @0x2C ticks；@0x06 缺失时 fallback 写 @0xB0 帧数（保 synthetic 测试）。
- `scene_composition_writers.go` SetFrameRate：duration 重算改读 @0x2C。
- `cdta_layout.go`：`CdtaDuration`→`CdtaShutterAngleMax`（0xB0）、`CdtaDurationMirror`→`CdtaShutterAngleMaxMir`（0xB8）；`CdtaMasterTicks`(0x2C) 注释改为「权威时长 ticks」。

## 验证

- AE2020 实测：fs from-scratch comp（shutter 172/-86）AE 读回 **172/-86**（修前 206/-103）。
- 5 个 AE-native 工程经我方 parser 读时长 = **10/10/4/5/8s**（修前 12/6/15/12/14.4s）。
- `go test ./...` 全绿（synthetic 测试经 @0xB0 fallback 保住）。
- **未跑双版本 ship-gate / 渲染像素**（这是 readback/解析层修复，非结构性写新增）；shutter 是 📋 读值能力。

## 关联

`cdta-duration-two-representations.md`（旧理解：@0xB0=frame-count + time-ratio 两表示——**本 incident 修正：@0xB0 根本不是 duration**，真时长在 @0x2C）· `tickrate-per-composition.md` · `shutter-side-effect-divisors.md`。RE 工具：`tmp_debug/re_dur.jsx`（AE 造多时长工程）+ `dump_cdta`。
