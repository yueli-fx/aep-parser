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

## 字节 delta（separated → merged）— **不是 separate 的逆**

观测对：`re_sepdim_merge_before.aep`（separated 3D）→ AE toggle false → `*_after.aep`。

1. **Leader**：tdsb `00000803` → `00000001`（byte2 `0x08→0x00`、清 bit1）；cdat 头 3×f64 ← **从 followers 取回的真值** `[100,200,50]`。tdb4 不变。
2. **ALL followers 删除**：Position_0/1/2 的 `tdmn + tdbs` 全部移除（tdgp 15→9 children）。**不是清零保留** —— 物理删除。

## merged 有两种合法表示（AE 都接受）

- **fresh-merged**（建层后从未分离）：leader + Position_0/1 **预分配**（zeroed cdat、tdsb=0x03），无 Position_2。`re_separate_dims_before.aep` 13 children。
- **merged-after-separate**：leader-only，无任何 follower。`re_sepdim_merge_after.aep` 9 children。

本仓 merge 写 leader-only 形（删全部 follower）。**ship-gate 关键发现**：AE resave 一个我们写的 merged 文件时会**重新预分配** zeroed Position_0/1（归一到 fresh-merged 形）。所以 merge 的 preservation 断言不能查 "follower 不存在"，要查 `DimensionsSeparated()==false` + **Position_2 不存在**（merged 永不含 Pos2）。

## 2D vs 3D — 靠 `Layer.Is3D`，不是 Components

- **2D 层 Position 也是 3 分量**（tdb4 `db990003000f0003`、cdat 72B、value `[x,y,0]`），所以 `Components` **无法**区分 2D/3D leader。
- 分离时：3D 加 `Position_2`，**2D 不加**（仅 X/Y）。判定用 `layer.Is3D`。
- ⚠ ExtendScript 对 2D separated 层读 `Position_2.value` 返回 **合成的 0**（物理不存在）—— ship-gate 不能靠 AE 读 Position_2 判 2D；靠 Go round-trip + byte-structural diff 证 "2D 无 Pos2"。

## 反直觉点（错题）

- **merged 态已含 Position_0/_1**（zeroed cdat、tdsb=0x03），并非"分离时才创建 X/Y"。
- **follower tdsb bit1 语义不是 dimensionsSeparated**：merged 态 follower 是 `0x03`、separated 态是 `0x01`，与 leader 相反。`DimensionsSeparated()` 只读 **leader** 的 bit1。
- **py-aep fixture 对（before_separation / transform_separated）不是干净 toggle 对**：py-aep 合成完整 transform schema，且其 fixture merged 态无 leader chunk、separated 态无 Position_2。**别拿 py-aep 样本做写路径 RE 基准** —— 用自建 AE 受控 before/after。

## 实现要点（`mutate_property_separate.go`）

- 唯一 fallible 步（合成 Position_2 的 `parseLeafProperty` re-parse）放在所有就地字节改动**之前** → 失败即零副作用，无需 rollback 机制。
- Property → Layer 反向引用：`AEPropertyGroup` 加 unexported `layer`（仅 parseLayer 在 root 上设），`Property.ownerLayer()` 走 parentTreeGroup→root 取；用于把合成的 Position_2 追加进 `Layer.Properties`。
- leader 默认值复用已有 `assignTransformDefaults` 填的 `Property.DefaultValue`（Position = `[w/2,h/2,0]`），无需 mutate 时再算 comp 维度。
- 序列化从 rifx.Chunk 树重算 LIST size，splice `grp.chunk.Children` 即可（同 `mutate_layer_insert` 模式）。

## scope / 暂搁

shipped：**静态 Position 的 separate（2D + 3D）+ merge 双向**，AE 2020 + AE 2025 ship-gate 6/6 PASS（`property_separate_shipgate_test.go`：3D sep / merge / 2D sep × 2 版本）。Go 输出 chunk 树 byte-structural 等同 AE 自存。

未做：**animated Position**（keyframe 流的迁移/合并未 RE）—— 当前 `StaticValue` 非 `[]float64` / follower 非 scalar 时一律 refuse。
