---
showcase: project-settings
direction: 工程设置 setter — Project.Set*（位深/线性混合/表达式引擎/素材时间码），纯 Go 从零设值后 AE app.project DOM readback 核对（📋 读值档，不看图）
capabilities: [bits-per-channel, linear-blending, expression-engine, footage-timecode-display]
gates: [project-flag-chunks-lnrb-lnrp]
status: 待review
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/project-settings  +  scripts/ae_run.ps1 verify.jsx"
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

## 设值 → 期望 readback（AE2020 实测一致）

| Set* 调用 | DOM 字段 | 期望值 |
|---|---|---|
| SetBitsPerChannel(BPC16) | bitsPerChannel | 16 |
| SetLinearBlending(true) | linearBlending | true |
| SetExpressionEngine("javascript-1.0") | expressionEngine | "javascript-1.0" |
| SetFootageTimecodeDisplayStartType(UseSourceMedia) | footageTimecodeDisplayStartType | FTCS_USE_SOURCE_MEDIA |

> 审核要点：`.done` 四行全 `-> OK` 即通过。

## ⚠ 已发现边界（诚实标注 · 本档刻意排除）

本 showcase 的 AE-DOM readback 核到三个**之前仅字节 round-trip、从未 AE-DOM 验证**的 setter 不反映到 AE DOM(红线4a 活样本,与 comp-settings 的 shutter 同类):

- **SetTimeDisplayType** 设 Frames → AE DOM 仍 Timecode。
- **SetFeetFramesFilmType** 设 MM16 → AE DOM 仍 MM35（默认）。
- **SetFramesCountType** 设 Start1 → AE DOM 不符。

线索:**timeDisplayType + feetFramesFilmType 共用 nnhd byte 8**(bit7=feet / bits6-0=timeDisplay),二者都失败;而**相邻的 footageTimecodeDisplayStartType(byte 9)成功** → 疑 byte 8 位打包 / 两 setter 互相清位。需独立 RE/修。其余工程 setter（SetWorkingGamma/SetAudioSampleRate/SetGpuAccelType/CMS 系列等）未纳入本批,按需再扩。

## 溯源

工程标志 chunk（lnrb/lnrp 等 presence-encoded）：`incidents/project-flag-chunks-lnrb-lnrp.md`；nnhd 布局见 `scene_project_settings.go` 注释；覆盖矩阵 `plans/coverage.md`。
