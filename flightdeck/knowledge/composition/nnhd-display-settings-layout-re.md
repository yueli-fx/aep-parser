# ⚠ nnhd display settings: AE reads legacy nhed, not nnhd (+ feet = frames-per-foot)

SUMMARY: nnhd display settings: AE reads legacy nhed, not nnhd (+ feet = frames-per-foot)
READ WHEN: 实现/调试 project 显示设置 setter（SetTimeDisplayType / SetFramesCountType / SetFeetFramesFilmType / SetFramesUseFeetFrames / SetFootageTimecodeDisplayStartType / SetTimecodeDefaultBase / SetTransparencyGridThumbnails）；写了 nnhd 字节 Go round-trip 绿但 AE DOM 不反映；以为 nnhd 是唯一 display-settings 来源；以为 feetFramesFilmType 是 nnhd byte8 bit7；新增任何 nhed/nnhd 双写头字段

---

## Signature
- symptom: `Go writes nnhd display-setting bytes, round-trip green, but AE DOM reads back the default (timeDisplayType/framesCountType/feetFramesFilmType/framesUseFeetFrames unchanged on reopen)`
- error_type: —
- where: internal/serializer/back_project.go (project display setters) + internal/scene/scene_project_settings.go (nnhd layout)
- trigger: 从零或 fixture-mutate 设 project 显示设置后用 AE app.project DOM 复核

## 症状/复现

showcase project-settings 的 AE-DOM readback 查出：`SetTimeDisplayType(Frames)` / `SetFramesCountType(Start1)` / `SetFeetFramesFilmType(MM16)` 写完字节 Go round-trip 全绿，但 AE 打开后 `app.project.*` 读回**默认值**，我方写入完全被忽略（红线4a false green）。`SetFootageTimecodeDisplayStartType`（nnhd byte9）当时「看似通过」实为**巧合**（我方 from-scratch seed 的 nhed[9] 恰好 = 设的值）。

复现：`go run ./tmp_debug/nnhd_verify_gen`（建 4 设置的工程）→ `scripts/ae_run.ps1` 跑 `test_data/generators/verify_nnhd.jsx` 读 DOM → `verify_nnhd.done` 显示 `NO ... got=<default>`。

## 根因

**两个独立错误叠加，py-aep 的 nnhd 布局两处是错的：**

1. **AE 从 legacy `nhed`（32B）读显示设置，不是 `nnhd`（40B）。** 工程根下 RIFX 同时有 `nhed`（旧紧凑头）和 `nnhd`（新扩展头），两者镜像同一组设置。AE 引擎读的是 `nhed`；`nnhd` 是冗余副本。只写 `nnhd` → AE reopen 时拿 `nhed` 里的旧值 → 静默忽略。（`SetBitsPerChannel` 早就双写 nhed[0x0F]+nnhd[0x18]，所以 bpc 一直对——其它 display setter 漏了这一课。）

2. **`feetFramesFilmType` 不是 byte8 bit7，是「每英尺帧数」(frames-per-foot)。** py-aep 说 byte8 bit7=feet 是虚构的；真值在 py-aep 标「unknown 0x00000010」的 `nnhd[16-19]`（u32 BE）/ `nhed[13]`（单字节）：**35mm = 16 (0x10)，16mm = 40 (0x28)**（电影实拍常识：35mm 每英尺 16 格、16mm 每英尺 40 格）。旧 feet setter 写 byte8 bit7 → 既写错位置、又**污染 byte8**（与 timeDisplay 共字节 = cockpit「互斥清位」症状的来源）。

`timeDisplayType`（byte8 整字节 0/1）和 `framesCountType`（byte20 0/1/2）的 **nnhd 位置 py-aep 其实是对的**——它们「失败」纯粹是因为只写了 nnhd 没写 nhed（错误 #1）；feet 还额外踩了错误 #2。

### nhed(32B) ↔ nnhd(40B) 字段映射（RE 2026-06-14，AE 2020 自存 byte-diff 实证）

| 字段 | nhed off | nnhd off | 编码 |
|---|---|---|---|
| time_display_type | 8 | 8 | 0=Timecode 1=Frames（整字节，非位打包）|
| footage_timecode_start | 9 | 9 | 0=Start0 1=UseSourceMedia |
| frames_use_feet | 11 bit0 | 11 bit0 | bool |
| timecode_default_base | 12 (单字节) | 14-15 (u16 BE) | 1-999（nhed 单字节，>255 截断）|
| feet_frames_film_type | 13 (单字节) | 16-19 (u32 BE) | frames/foot：35mm=16 16mm=40 |
| frames_count_type | 14 | 20 | 0=Start0 1=Start1 2=TimecodeConv |
| bits_per_channel | 15 | 24 | 已有（早就双写）|
| transparency_grid_thumbnails | 16 | 25 | bool（best-effort 映射，已双写）|

AE 2020 新建工程 DOM 默认：timeDisplayType=Frames、framesCountType=Start1、feetFramesFilmType=MM16、framesUseFeetFrames=true（与我方 from-scratch seed 默认不同，做 RE 时别拿默认当无操作）。

**DOM 可读性（决定能否 readback-gate）**：`app.project` 暴露 timeDisplayType / framesCountType / feetFramesFilmType / framesUseFeetFrames / footageTimecodeDisplayStartType / **transparencyGridThumbnails** / displayStartFrame —— 这些可 DOM-gate。**`timecodeDefaultBase` 无对应 DOM property**（probe_proj_props.jsx 实证 `"timecodeDefaultBase" in app.project === false`）→ binary-only，只能字节 round-trip 验，无法 AE-DOM readback 证实（同 rq-comment 类）；nhed[12] 单字节映射是 best-effort（>255 截断）。transparencyGridThumbnails 的 nnhd[25]/nhed[16] 映射经 AE 自存 true/false byte-diff 实证（非 best-effort）。

## 修法

`internal/serializer/back_project.go`：加 `mirrorNhed(off, v)` helper，每个 display setter 在写 nnhd 后**同步写 nhed 对应 offset**（按上表）。`SetFeetFramesFilmType` 改写 frames-per-foot（nnhd u32 @0x10 + nhed[13]）而非 byte8 bit7；`SetTimeDisplayType` 改成整字节 `byte(v)`（不再 `&0x80` 保留虚构的 feet bit）。reader（`scene_project_settings.go`）`FeetFramesFilmType()` 改读 `nnhd[16-19]` u32（40→MM16 否则 MM35）；其余 reader 读 nnhd 即可（写时已与 nhed 同步，AE 也保持两者一致）。新增接口方法 `NnhdUint32`（scene_writers.go + back_project.go 实现）。

**验证（必须 AE-DOM，不能只 Go round-trip——这正是 false-green 的源头）**：
- `go run ./tmp_debug/nnhd_verify_gen` → AE2020 **和** AE2025 跑 `verify_nnhd.jsx`，4 字段全 `OK`（2026-06-14 双版本实测 PASS）。
- 回归测试：`internal/serializer/project_settings_internal_test.go::TestProjectSettings_NhedNnhdMirror`（合成 nhed+nnhd chunk，断言每个 setter 双写正确 offset）。

RE 工具（本地，`tmp_debug/` + `test_data/` 均 gitignored，按本文从零可重建）：`tmp_debug/dump_nhed_nnhd.go`（同 dump 两头）· `tmp_debug/nnhd_verify_gen/`（建验证工程）· `test_data/generators/re_nnhd_settings.jsx`（AE 自存两值变体供 byte-diff）· `test_data/generators/verify_nnhd.jsx`（DOM readback gate）。永久回归走 `TestProjectSettings_NhedNnhdMirror`（tracked）。

## Cases
- 2026-06-14 首次：showcase project-settings AE-DOM readback 暴露；RE 出 nhed-primary + frames-per-foot 双根因；7 个 display setter 全改双写；AE2020+2025 双版本 DOM gate 全绿。commit 见 git log。
