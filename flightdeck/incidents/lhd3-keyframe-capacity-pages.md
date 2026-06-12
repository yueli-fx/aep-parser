---
status: active
when_to_read: AE 2025 rejects a Go-written project as corrupt (项目文件似乎已损坏/读取无效) when a property has >4 keyframes; touching encodeKeyframes lhd3 header fields; assuming lhd3 @0x0C/@0x1C are constants; extending any keyframe-list writer (path time table / mask om-s)
applies_to: [lhd3, keyframe, capacity, pages, encodeKeyframes, ae2025-reject, scale-boundary, position-kf, mg-roadmap]
last_updated: 2026-06-12
resolved_by:
---

# lhd3 容量字段按 4-keyframe 分页 — AE 2025 校验 reject

## Signature
- symptom: AE 2025 打开 Go 写的工程报「该项目文件似乎已损坏 (读取……无效)」；同一文件 AE 2020 正常打开；属性带 >4 个关键帧时触发，≤4 个不触发
- error_type: —
- where: internal/serializer/lower_property_stream.go encodeKeyframes (lhd3 头)
- trigger: 任何属性流写入 ≥5 个关键帧后在 AE 2025 打开（MG roadmap S1 规模 gate 首撞）

## 症状/复现

MG ease gate（6 关键帧 Position）AE 2020 PASS、AE 2025 reject。bisect 矩阵（`tmp_debug/mg_bisect`）：2kf linear OK、2kf **ease OK**、**6kf linear REJECT** → 触发器 = 关键帧数量，与 ease/bezier 无关。「2 关键帧覆盖边界」（交付准则红线 2）不只是未验证——是真实硬边界。

## 根因

lhd3 @0x0C 与 @0x1C **不是常量**（历史注释 "observed constant" 来自 2-keyframe fixture 的巧合）：

- `@0x0C` = ceil(n/4)：容量**页数**，每页 4 个关键帧
- `@0x1C` = 4 × 页数：**槽容量**

我们恒写 1 / 4 → 6 kf 声称容量 4 → AE 2025 校验 count ≤ capacity 不过判损坏；AE 2020 无此校验（宽松重算）。对照 fixture = AE 2025 原生 6kf Position（`gen_native_sixkf.jsx`），diff 工具 `tmp_debug/mg_bisect/dumpkf`。

**未修但已观察**：原生 ldat 每块 @0x10 有 f64 segment 长度缓存（等距 300.0；末块为 epsilon 残值），我们写 0——双版本接受且渲染正确（运行时重算缓存），暂不复刻。`encodePathTimeTable`（path 时间表 bpk64）@0x0C 同样恒写 1——path 多关键帧 >4 时大概率同病，**扩 path keyframe 规模前先按本 incident 修它**。

## 修法

`encodeKeyframes`：`pages := (n+3)/4`；`@0x0C = pages`、`@0x1C = 4*pages`（commit 同 S1 ease gate）。验证：bisect 4/4 OPEN ok + `TestMGEase_AEShipGate_*` 双版本渲染像素 PASS（6kf 中点插值 x=950 精确命中）。

## Cases
- 2026-06-12 首次（MG roadmap S1：ease + 规模 gate 同场发现；ease 路径反而无辜——interp 字节硬编码 linear 的问题在同 commit 一并修复 `writeKeyframeBlock` per-side bezier）
