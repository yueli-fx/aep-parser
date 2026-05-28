---
when_to_read: implementing any allocItemID / monotonic-ID logic; computing max(used IDs) across a parsed project; debugging head-counter collisions after structural mutation; reviewing initDerived or any function summing IDs across project state
applies_to: [nextItemID, allocItemID, head-counter, id-allocation, initDerived, layer-id, parse, monotonic-id, inv-9, duplicate-layer, new-shape-layer]
last_updated: 2026-05-28
---

# initDerived must walk LAYER IDs when computing nextItemID

V3 Phase 5B 暴露的隐患：`Project.initDerived` 只走 `Compositions / Footage / Folders` 求 `max(ID) + 1` 当 `nextItemID`，**漏了 layer**。AE 把 layer ID 也分配在同一个 head counter 下 —— 一个 comp 内 layer ID 普遍 `> footage ID`（因为 AE 添加 solid 时是先建 footage 再建 layer，layer 拿后一个号），所以 `max(layer)` 可能 `>` `max(folder/comp/footage)`。

## 表象

`re_trackmatte_ae24.aep` 解析后：
- comp + 7 footage IDs：max ≈ 66
- 7 layer IDs：[55, 57, 59, 61, 63, 65, 67] —— **max = 67**
- `initDerived` 算出 `nextItemID = 66 + 1 = 67`（漏 layer）
- 接着 `DuplicateLayer → allocItemID()` 返回 67 —— **跟某 layer 的 ID 撞了**
- WriteAEP + re-parse 后，两个 layer 共用 ID 67；`LayerByID(67)` / `for l.ID == 67` 命中错的那个，clone 的 matte / name 字段全部"消失"

## 根因

```go
// 错的代码（initDerived 漏 layers）：
for _, c := range p.Compositions { if c.ID > maxID { maxID = c.ID } }
for _, f := range p.Footage      { if f.ID > maxID { maxID = f.ID } }
for _, fo := range p.Folders     { if fo.ID > maxID { maxID = fo.ID } }
p.nextItemID = maxID + 1   // ← 漏掉 c.Layers 的 ID
```

**错误假设**：item-level（Composition / Footage / Folder）的 ID 永远 ≥ 层级嵌套的 layer ID。

**真相**：AE 用单一 head counter 分配 `id_n = n`（monotonic）给项目里的**任何 entity**（item 或 layer）。AE 添加层时（addSolid 等），footage 先拿一个号，layer 立刻拿下一个号 —— `layer.ID == footage.ID + 1`。一旦某个 layer 的 ID 比所有 item 的 ID 都大，旧的 initDerived 就漏了。

## 修复

`initDerived` 也走 `c.Layers`：

```go
for _, c := range p.Compositions {
    if c.ID > maxID { maxID = c.ID }
    for _, l := range c.Layers {
        if l.ID > maxID { maxID = l.ID }
    }
}
```

固定在 `parse.go::initDerived` 里，跟 footage / folder 三个 loop 合一处。Phase 5B 的 `TestDuplicateLayer_ExplicitMatte_RoundTrip` 是 first triggering case；以前的 dup 测试用 `re_delete_layer_baseline.aep`（max(layer) ≤ max(footage)）所以没暴露。

## 适用范围

任何 "走 project state 求 max ID" 的代码都要把 `c.Layers` 算进来：
- `Project.allocItemID` 派生于 `nextItemID` → ✅ 修了上游
- 未来的 `Project.DuplicateItem` / `InsertLayer` 跨 comp 走 ID-collision check → 看到这条 scar 直接走 `initDerived` 的 maxID 计算逻辑，别再 reinvent
- Inv-9（monotonic ID, no reuse）成立的前提是 nextItemID 严格 `> max(used)` —— 这条 scar 是 Inv-9 的 "computation correctness" 补丁

## 测试

`TestDuplicateLayer_ExplicitMatte_RoundTrip` 走 `re_trackmatte_ae24.aep` 这条路径 —— 修复后 clone 的 ID 是 68（不撞），round-trip 后 LayerByID 找到 clone 自身（不是 ID==67 的解 layer）。

不必加 dedicated regression test：`re_trackmatte_ae24.aep` 解析时就走 fix 后的 initDerived，dup test 之间 cover。

## Related

- 修复 commit: 2026-05-28 V3 Phase 5B（同 commit 一起 ship）
- Phase 5B plan: [`../plans/2026-05-28-v3-phase5b-duplicatelayer-explicit-matte-plan.md`](../plans/2026-05-28-v3-phase5b-duplicatelayer-explicit-matte-plan.md)
- 受影响代码: `internal/aep/parse.go::initDerived`
- 受影响 invariant: Inv-9 (monotonic ID alloc, no reuse) in `scars/ae25-acceptance-gate.md`
