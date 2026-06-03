---
status: active
when_to_read: implementing SetDimensionsSeparated merge direction / 2D / animated; debugging Position separation byte layout; extending separation to other multidimensional properties; reviewing mutate_property_separate.go
applies_to: [property, dimensions-separated, separate-dimensions, position, structural-write, mutate, tdsb, cdat]
last_updated: 2026-06-04
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

**animated Position separate 方向已实现**（2026-06-04，`separatePositionAnimated`，3D linear-path-ease 首切片）：Go round-trip 通过，且 separate 后 4 个 Position* tdbs **byte-identical AE 自存 after fixture**（`tmp_debug/verify_sep_anim`）。**merge animated 方向 + AE 双版本 ship-gate 待做**（plan `2026-06-04-py-aep-p3-dimsep-animated-plan.md` Phase 3-4）。当前 merge animated 仍 refuse（mergePosition 拒绝非 scalar follower）。

## animated 子方向 — keyframe 流迁移映射（RE 2026-06-04）

结构与 static 同构，仅 stream 类型互换：separated 态 leader 退回 **static 默认**（kf 流清空，tdsb/cdat 同 static separate）+ Position_0/1/2 变 **animated**（per-axis 1D temporal kf 流）+ 合成 Pos2；merged 态 leader 是 animated 3D spatial kf 流、Position_0/1 static-0 占位。

**leader（3D spatial, bpk=128）→ follower（1D non-spatial, bpk=48）逐 kf 映射**（AE 2020 `re_separate_dims_anim.jsx` before/after，`tmp_debug/dump_sepdim_kf` 解码，9/9 kf-side 全对上含符号）：
```
follower.value[axis] = leader.value[axis]
follower.out_speed   = leader.outSpatTan[axis] × 100      (= /influence)
follower.in_speed    = −leader.inSpatTan[axis]  × 100
follower.influence   = 0.01 (相邻段 side) | 0 (首 kf in / 末 kf out)
follower interp      = bezier 双侧
```
`spatialTan = speed × influence`，常数 100=1/0.01。leader spatial tangent 是 AE 存盘现成值（auto-bezier `(P_next−P_prev)/6`），**读取不重算**。merge 反向：`outSpatTan=out_speed/100`、`inSpatTan=−in_speed/100`。

⚠ 未决（impl byte-check）：leader 自身 path temporal ease 非默认时映射待验（首切片限 path-ease≈linear）；100/0.01 是否随时距变 → ship-gate 用第二 fixture 防 fixture-specific。

### 实现确认的字节细节（2026-06-04 byte-check `re_sepdim_anim_after.aep`，已 byte-identical 复现）

**follower kfl block**（bpk=48，1D non-spatial）：time@0x00（ticks，per-comp tickRate）/ in interp@0x04=0x02(bezier) / out@0x05=0x02 / @0x06=0x00 / **hdr07@0x07=0x08**（非 0x00！）/ value@0x08 / in_speed@0x10 / in_inf@0x18 / out_speed@0x20 / out_inf@0x28。`speed = spatTan×100` 在 float64 下 **bit-exact**（`33.333…×100` 的 bits 等于 AE 存盘 bits）。lhd3 52B：magic+count@0x08+`1`@0x0C+bpk@0x10+`4 1 4`@0x14/0x18/0x1C。

**follower tdbs static→animated**：tdb4 `@0x05 &^= 0x01`、`@0x44 = 0x01`（= 既有 shape `injectAnimatedStream` 配方）；tdsb `byte3 &^= 0x02`（0x03→0x01）；cdat child 原位换成 LIST(kfl)（保留 tdum/tduM）。

**leader animated→static-separated**：tdb4 `@0x05 |= 0x01`、`@0x44 = 0x00`、`@0x4f |= 0x01`（与 follower 方向相反，注意 @0x4f）；tdsb byte2=0x08+byte3 bit1（同 static separate）；LIST(kfl) child 原位换成 **72B cdat = [default(3 f64), kf0.inSpatTan(3), kf0.outSpatTan(3)]**（AE 把首 kf 的 in/out 空间切线塞进 cdat 尾，非零）。
