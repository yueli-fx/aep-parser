# ⚠ lhd3 容量字段按 4-keyframe 分页 — AE 2025 校验 reject

lhd3 容量字段按 4-keyframe 分页 — AE 2025 校验 reject

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

**未修但已观察**：原生 ldat 每块 @0x10 有 f64 segment 长度缓存（等距 300.0；末块为 epsilon 残值），我们写 0——双版本接受且渲染正确（运行时重算缓存），暂不复刻。

**`encodePathTimeTable` 同病已修（2026-06-14）**：path 时间表 lhd3（bpk64，`lower_shape_node.go`）此前 @0x0C 恒写 1、@0x1C 恒写 4 —— 同 `encodeKeyframes` 旧坑，path >4 关键帧 AE 2025 判损坏。改为 `pages=(n+3)/4`、`@0x0C=pages`、`@0x1C=4*pages`（同 commit）。验证：`TestV2_2_PathKf_AEShipGate_AE20{20,25}` 从 3 关键帧 bump 到 **6 关键帧**（ceil(6/4)=2 页），双版本 PASS（AE 接受、JSX 断言 numKeys≥6 未截断、Go re-decode resaved ≥6 shap + time-table kfl 存活）。

**仍未修 / 另一轴**：path **几何** lhd3（per-shap，@0x0C=nVerts、@0x1C=16）是**顶点数**轴而非关键帧数轴——`encodeBezier` 写 @0x1C=16，AE-native 疑用 `cap=nextPow2(nVerts)`（`path-keyframe-write-re.md` §104）。shape gate 只验过 nVerts≤4（AE 容忍偏差，mask 才严格）；**>4 顶点的 path 是否同样需分页未测**，本次 gate 顶点数刻意保持 ≤4 隔离此轴。需求驱动时单独 RE。

## 修法

`encodeKeyframes`：`pages := (n+3)/4`；`@0x0C = pages`、`@0x1C = 4*pages`（commit 同 S1 ease gate）。验证：bisect 4/4 OPEN ok + `TestMGEase_AEShipGate_*` 双版本渲染像素 PASS（6kf 中点插值 x=950 精确命中）。

## Cases
- 2026-06-12 首次（MG roadmap S1：ease + 规模 gate 同场发现；ease 路径反而无辜——interp 字节硬编码 linear 的问题在同 commit 一并修复 `writeKeyframeBlock` per-side bezier）
- 2026-06-14 第二处（`encodePathTimeTable` path 时间表同坑，roadmap 优先级1 首项）：`encodeKeyframes` 那次只修了标量/矢量流，path 时间表 lhd3 漏修，恒写 1/4。修法相同（pages 化）；gate 从 3kf bump 到 6kf 双版本 PASS。`encodeKeyframes`（标量/矢量）+ `encodePathTimeTable`（path）两条 keyframe 路径容量分页**全闭合**。
- 2026-06-17 第三处验证（**mask path 动画 >4kf**，无代码改）：`encodePathTimeTable` 被 `makeMaskShapeOmSAnimated` 复用，但 **mask 严格性**（AE 急切解码 mask outline，容量字段错硬崩 0::42）下 >4kf 此前未单独 gate（`TestMGMaskPathKf` 只验 2kf=单页）。把它从 2kf bump 到 6kf（=2 页），AE 2020+2025 双版本 PASS（numKeys=6 读回不截断 + render 翻转 + resave 保留 6kf）→ 确认容量分页在 mask 上同样成立。详 [[add-mask-create-re]] § 2026-06-17。

---

## [合并] 关键帧字节布局两种 — 必须走 layoutFor 分发（原 `keyframe-byte-layout-dispatcher`,2026-06-16 折入）

容量分页(上)是 lhd3 头的轴;**block 内字节布局**是另一轴,同样**不统一**。两种 layout 互斥,按 property 类型决定——**不要在 caller 端手算 offset**。

- **Spatial-style**(4D color / Position / Anchor 3D):block 头 byte `0x07 = 0x07` 或 `0x01 + dims≥2`;scalar ease @0x18/0x20/0x28/0x30;values @0x38;bpk = 0x38 + 3·N·8。
- **Non-spatial**(Opacity 1D / Scale 3D / Mask Feather 2D):头 byte `0x07 = 0x00`;values @0x08;per-component ease @ `0x08 + (N+i)·8`;bpk = 0x08 + 5·N·8。

**How to apply**:加新 keyframe property 类型 → 走 `layoutFor(header07, dims)` 集中分发,**不新加 if/else 链**;改写回 → 用同一 `layoutFor` 算 offset 保 read/write 对称;调试 ease/value 错位 → 先 dump block 头 byte `0x07` 确认 layout。**Why**:早期 caller 端到处手算 offset → 改一种漏改另一种 → silent 错位;`layoutFor`(`parse_keyframe.go`)是唯一权威分发点(被 back_keyframe/write_property 多处调用)。
