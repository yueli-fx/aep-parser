---
showcase: project-settings
direction: 工程设置 setter — Project.Set*（位深/线性混合/表达式引擎/素材时间码），纯 Go 从零设值后 AE app.project DOM readback 核对（📋 读值档，不看图）
capabilities: [bits-per-channel, linear-blending, expression-engine, footage-timecode-display, time-display-type, frames-count-type, feet-frames-film-type, frames-use-feet]
gates: [project-flag-chunks-lnrb-lnrp, nnhd-display-settings-layout-re]
status: complete
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/project-settings  +  scripts/ae-worker/ae_run.ps1 verify.jsx"
---

# project-settings — 工程设置 setter showcase（📋 读值档）

## 这个方向测什么

`Project.Set*` 工程级设置,纯 Go 从零设值后 **AE 打开 + `app.project` DOM readback 核对**(无渲染视觉 → 读值不看图)。仅收 **AE-DOM readback 完全一致**的 setter。

## 验证方式（📋 读值，非看图）

`verify.jsx` 打开 .aep,对每项做 `got === want` 显式判定写入 `.done`(每行带 `-> OK/NO`)。无 png。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `project_settings.aep` + 应用各 Set* |
| `verify.jsx` | 读值脚本(tracked) | dump app.project DOM + OK/NO 判定 → `.done` |
| `project_settings.aep` | 产出工程 (gitignored) | 1920×1080，AE2020 |
| `project_settings.done` | readback 日志 (gitignored) | 用户读这个核值 |

## 设值 → 期望 readback（AE2020 + AE2025 实测一致）

| Set* 调用 | DOM 字段 | 期望值（NON-default）|
|---|---|---|
| SetBitsPerChannel(BPC16) | bitsPerChannel | 16 |
| SetLinearBlending(true) | linearBlending | true |
| SetExpressionEngine("javascript-1.0") | expressionEngine | "javascript-1.0" |
| SetFootageTimecodeDisplayStartType(UseSourceMedia) | footageTimecodeDisplayStartType | FTCS_USE_SOURCE_MEDIA |
| SetTimeDisplayType(Timecode) | timeDisplayType | TIMECODE |
| SetFramesCountType(Start0) | framesCountType | FC_START_0 |
| SetFramesUseFeetFrames(true) | framesUseFeetFrames | true |
| SetFeetFramesFilmType(MM35) | feetFramesFilmType | MM35 |

> 审核要点：`.done` 第一行 `PASS` + 八行全 `-> OK` 即通过。

## ✅ 已修复（曾经的 false-green 边界）

本档曾排除 SetTimeDisplayType / SetFramesCountType / SetFeetFramesFilmType（字节 round-trip 绿但 AE DOM 不反映，红线4a）。**2026-06-14 RE 定位并修复**：AE 从 legacy `nhed` 头读显示设置（非 `nnhd`），且 feetFramesFilmType 真存「每英尺帧数」（35mm=16/16mm=40）而非 byte8 bit7。setter 现双写 nhed+nnhd。AE 2020 + 2025 双版本 DOM readback 全绿。详 `incidents/nnhd-display-settings-layout-re.md`。其余工程 setter（SetWorkingGamma/SetAudioSampleRate/SetGpuAccelType/CMS 系列等）未纳入本批,按需再扩。

## 溯源

工程标志 chunk（lnrb/lnrp 等 presence-encoded）：`incidents/project-flag-chunks-lnrb-lnrp.md`；nnhd 布局见 `scene_project_settings.go` 注释；覆盖矩阵 `plans/coverage.md`。
