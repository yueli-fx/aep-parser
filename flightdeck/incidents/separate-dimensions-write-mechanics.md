---
status: active
when_to_read: implementing SetDimensionsSeparated merge direction / 2D / animated; debugging Position separation byte layout; extending separation to other multidimensional properties; reviewing mutate_property_separate.go
applies_to: [property, dimensions-separated, separate-dimensions, position, structural-write, mutate, tdsb, cdat]
last_updated: 2026-06-02
---

# Separate Dimensions 写机制 — RE findings

`Property.SetDimensionsSeparated(true)`（P3 §3C）不是翻一个 bit，是带**值迁移 + 默认合成 + 组件数分支**的结构性重构。RE 来源：AE 2020 受控 before/after（`test_data/re_separate_dims.jsx` 同一工程 toggle 一次，存两份），`tmp_debug/list_item_chunks` diff。

## 字节 delta（merged → separated，静态 3D Position）

观测对：`re_separate_dims_before.aep`（merged, value=[100,200,50]）→ AE toggle → `*_after.aep`。

1. **Leader `ADBE Position`**（就地，72B cdat 长度不变）：
   - tdsb `00000001` → `00000803`：byte2 `0x00→0x08`、byte3 bit1（dimensions_separated）置位。
   - cdat 头 3×f64 BE `[100,200,50]` → **property 默认** `[w/2, h/2, 0]`（comp 1920×1080 = `[960,540,0]`）。tdb4 不变。
   - **leader 值分离后不再权威** —— AE 读 per-axis followers。
2. **Position_0 / Position_1** follower（就地）：tdsb `00000003` → `00000001`（清自身 bit1）；cdat 头 8B 由零填入迁移值（X / Y）。
3. **Position_2** 新增（结构性，+~306B 含 LIST overhead）：merged 态**不存在**（AE 只预分配 X/Y）。合成 = 克隆 Position_1 的 `tdmn + tdbs(6 children: tdsb/tdsn/tdb4/cdat/tdum/tduM)`，tdmn 改名 `ADBE Position_2`、cdat 头填 Z、tdsb 清 bit1。插在 Position_1 的 tdbs 之后。

## 反直觉点（错题）

- **merged 态已含 Position_0/_1**（zeroed cdat、tdsb=0x03），并非"分离时才创建 X/Y"。只有 Z（`Position_2`）是分离时才出现 → 组件数分支：3D 加 1 个 follower，2D 不加。
- **follower tdsb bit1 语义不是 dimensionsSeparated**：merged 态 follower 是 `0x03`、separated 态是 `0x01`，与 leader 相反。py-aep / 本仓的 `DimensionsSeparated()` 只读 **leader** 的 bit1，follower 自身 bit1 不参与该语义。
- **py-aep fixture 对（before_separation / transform_separated）不是干净 toggle 对**：py-aep 合成完整 transform schema（三个 golden 都显示 Position+follower），且其 fixture 的 merged 态无 leader chunk、separated 态无 Position_2（疑似 2D 或不同 AE 版本/构造）。**别拿 py-aep 样本做写路径 RE 基准** —— 用自建 AE 受控 before/after。

## 实现要点（`mutate_property_separate.go`）

- 唯一 fallible 步（合成 Position_2 的 `parseLeafProperty` re-parse）放在所有就地字节改动**之前** → 失败即零副作用，无需 rollback 机制。
- Property → Layer 反向引用：`AEPropertyGroup` 加 unexported `layer`（仅 parseLayer 在 root 上设），`Property.ownerLayer()` 走 parentTreeGroup→root 取；用于把合成的 Position_2 追加进 `Layer.Properties`。
- leader 默认值复用已有 `assignTransformDefaults` 填的 `Property.DefaultValue`（Position = `[w/2,h/2,0]`），无需 mutate 时再算 comp 维度。
- 序列化从 rifx.Chunk 树重算 LIST size，splice `grp.chunk.Children` 即可（同 `mutate_layer_insert` 模式）。

## scope / 暂搁

shipped：**仅静态 3D Position 的 separate 方向**，AE 2020 + AE 2025 ship-gate PASS（`property_separate_shipgate_test.go`）。

未做（各需独立 RE + ship-gate）：merge（separate→merged）、2D Position（无 Position_2）、animated Position（keyframe 流迁移）。当前对这些一律 refuse。
