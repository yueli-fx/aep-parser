---
status: active
when_to_read: implementing RenderQueue add/insert; debugging Rout chunk layout; extending RQ structural ops; reviewing mutate_render_queue.go
applies_to: [render-queue, RQItem, remove, structural-write, mutate, Rout, lhd3, LItm, LRdr]
last_updated: 2026-06-02
---

# Render Queue delete 写机制 — RE findings

`RenderQueue.RemoveItem(index)`（P3 §3A，结构性）。RE 来源：AE 2020 受控 2-item→1-item（`test_data/re_rq_delete.jsx`，`renderQueue.item(2).remove()`），`tmp_debug/dump_lrdr` diff。

## LRdr 容器结构（root 直接子 LIST:LRdr）

```
LRdr
 ├ Rhed (20B)                         ← 队列头（删 item 不变）
 ├ Rout (4B header + per-item block)  ← 每 item 的 flags
 ├ LIST:list → lhd3(52B) + ldat(2246B × N)  ← per-item RenderSettingsItem
 ├ LIST:LItm → per item: [RCom?] + LIST:list(OM settings) + LIST:'LOm '
 └ LIST:LSIf → ARsi                   ← （删 item 不变）
```

## 删 item i 的四处联动

1. **LItm**：移除该 item 的 `[RCom?] + LIST:list + LIST:'LOm '`（无 comment 时无 RCom）。`LOm ` 是该 item settings list 之后的下一个 `LOm ` sibling。
2. **settings ldat**：splice 掉 `[i*2246 : (i+1)*2246]` 这 2246B 块。⚠ 每个 item 的 `settingsBlock` 别名共享同一 ldat 底层数组 —— 删中间块后必须**重建 ldat.Data + 重挂存活 item 的 settingsBlock 别名**到新偏移，否则后续 item 读到错位字节。（OM 的 128B settings 在各 item 自己的 list 里，不共享，无此问题。）
3. **settings lhd3**：@0x08 和 @0x0C 两个 count 字段都 -1（@0x10 = 2246 stride 不变）。
4. **Rout**：4B header + 均匀 per-item block。header = item 数比例值（2 item=0x0a, 1 item=0x05 → 5×count）。删法：`stride=(len-4)/count`、splice 掉 `[4+i*stride : 4+(i+1)*stride]`、`header -= header/count`。

`Rhed` / `LSIf:ARsi` 删 item 时不变。

## 验证

Go `RemoveItem(1)` 输出的 LRdr 子树 **byte-structural 等同** AE 自身 `item(2).remove()`（dump_lrdr diff 空）。双版本 ship-gate PASS（`render_queue_remove_shipgate_test.go`：AE 2020 + 2025 读回 numItems=1 + 存活 comp=RQA）。

## ADD（clone + remap）

`RenderQueue.AddItem(comp)` 走 **clone 末位 item + remap comp_id**（同 InsertLayer 思路），不从零合成：deep-clone 模板的 2246B settings block（改 @0x08 comp_id = comp.ID）+ `[list, 'LOm ']` 组 + Rout per-item block，append 三处 + lhd3/Rout header 增。**双版本 ship-gate PASS**（`render_queue_remove_shipgate_test.go`：AE 2020 + 2025 接受 cloned item，读回 numItems+1 + item(n).comp = 目标 comp）。

- 克隆的 OM 保留模板的 output path/template（AE 接受；fresh add 会按 comp 命名输出 —— deferred 细节，不影响接受/comp 链接）。
- 空队列无模板可克隆 → refuse（同 NewComposition 需 template 的限制）。
- ldat append 触发 realloc → **全部** item（含既有）的 settingsBlock 别名重挂。

## scope / 未知

- **Rout per-item stride 仅在单 OM/item 上 RE 过**（20B = 5×u32）。多 output module 的 item 是否 per-OM 扩展 Rout 未验 —— `mutate_render_queue.go` 用计算 stride `(len-4)/count` + 比例 header，对均匀 item 正确；多 OM 混合需补 RE。
- Rout per-item block 内容（item1 `00000013..` vs item2 `40000013..` 不同）疑为 per-item 进度/状态；clone 时照搬，AE 接受未现问题。语义未细 RE。
