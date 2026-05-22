# AE 25 接受自建 .aep 需要的 3 类隐藏字段

Phase 6 ship gate 2026-05-22 实测，AE 25 打开 builder 产 .aep 有三道关。每道都让 AE 表现不同：

## 症状梯度

| Stage | builder 缺什么 | AE 25 反应 |
|---|---|---|
| 1 | cdta 二级时间字段全零 | **崩溃** (AfterFX.exe 进程崩) |
| 2 | head[12..15] / head[16..19] 跟实际 itemID 不匹配 | 同 Stage 3 表现 |
| 3 | Item LIST 后面缺 8 个 sibling chunks (FEE / fvdv / fiop / ftts / foac / fiac / fipc / fifl) | **"After Effects 错误: 文件数据丢失"** |

3 道都修才能打开。每道只修一道 AE 还报后续阶段错误。

## 字段 RE 来源

`test_data/fdta_probe/AE2025_2comp.aep`（AE 自己存的 2-comp 文件）+ `test_data/re_tickrate.aep` 各 RE_fps_* comp 对照。

### Stage 1: cdta 二级时间字段（fps-依赖）

cdta_layout.go 新增常量 + framerate_canonical.go `canonicalFpsTiming` 表：

| offset | 含义 | 30fps | 29.97fps | 50fps | 59.94fps |
|---|---|---|---|---|---|
| @0x06 (u16) | ticks_per_frame | 1024 | 800 | 512 | 400 |
| @0x08 (u32) | TickRate = tpf×fps | 30720 | 23976 | 25600 | 23976 |
| @0x10 (u32) | const 600 | 600 | 600 | 600 | 600 |
| @0x18 (u32) | TickRate mirror | = @0x08 | | | |
| @0x2C (u32) | tpf×5×fps_nominal_whole | 153600 | 120000 | 128000 | 120000 |
| @0x30 (u32) | TickRate mirror | = @0x08 | | | |
| @0xB8 (u32) | duration_frames mirror of @0xB0 | | | | |

NTSC nominal_whole: 29.97→30 / 59.94→60 / 23.976→24（@0x2C 计算用此，不是 floor(fps)）。

23.976 fixture 缺，用 1000 t/f 数学推算（fixture 出来补实证）。

### Stage 2: head chunk counters

Root `head` chunk (20 字节) `[12..15]` / `[16..19]` 两个 uint32 BE counter。`NewProject(TargetAE2025)` 从模板继承 `[1, 1]`。AE 期望 counter ≥ 当前最大 itemID。

修：`WriteAEP` 入口先 `syncHeadCounters()` —— 把两个 counter 拉到 `max(currentValue, nextItemID)`。

counter B 语义没完全 RE（AE2025_2comp 是 49, dummy_comp 是 25, re_batch 是 5229 —— 跟 save sequence 或总属性数相关）。pragmatic: 设为 nextItemID 够开。

### Stage 3: Item LIST sibling chunks

每个 Item LIST 在 Fold 里必须跟 8 个 sibling chunks（**不是** Item LIST 的 children，是 Item 在 Fold 中的紧邻 siblings）：

```
LIST Fold:
  fdta
  LIST Item  ← comp 1
  LIST FEE  (1 child: ppSn 8B)   ← sibling 1
  fvdv (4 B = 0x00000003)         ← sibling 2
  fiop (1 B = 0x00)                ← sibling 3
  ftts (4 B = 0x00000000)          ← sibling 4
  foac (1 B = 0x00)                ← sibling 5
  fiac (1 B = 0x00)                ← sibling 6
  fipc (2 B = 0x0000)              ← sibling 7
  fifl (4 B = 0x00000000)          ← sibling 8
  LIST Item  ← comp 2
  LIST FEE ...                     (重复 sibling 1-8)
```

修：`ensureTemplateCompItemChunks` 多提取一组 `templateItemSiblingChunks` from dummy_comp。NewComposition 在 `rootFold.Children` append Item 后再 append 8 个 clone。

8 个 chunks 跨 comp 完全相同（AE2025_2comp 两个 comp 字节相同），无需 per-item 修正。

## 语义猜测

8 个 sibling 前缀都是 `f*`，估计 footage/folder metadata cache。`ppSn` = 0x4062C000_00000000 (float64 = 150.5)，可能 panel zoom。AE 内部缓存，opaque 即可。

## 教训

- builder-from-scratch 不能只看 parser 反向：parser 容错（缺字段当默认 0），AE 严格（缺字段不开 / crash）。
- 高代价 RE 路径: 写 builder → 跑 AE ship gate → 找崩溃原因 → 字节对比 AE-saved fixture → 反推字段。每修一道走一遍。
- AE 25 的"文件数据丢失"提示 = item siblings 缺；崩溃 = cdta 时间字段空。两个错误信号互补可区分根因。

## 关联 commits / 文件

- `internal/aep/cdta_layout.go` 新增 7 offsets
- `internal/aep/framerate_canonical.go` `canonicalFpsTiming` + `lookupFpsTiming`
- `internal/aep/new_composition.go` `templateItemSiblingChunks` + Fold append
- `internal/aep/write.go` `syncHeadCounters`
