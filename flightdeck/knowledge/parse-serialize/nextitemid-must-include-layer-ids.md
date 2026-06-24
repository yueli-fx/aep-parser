# ⚠ initDerived must walk LAYER IDs when computing nextItemID

SUMMARY: initDerived must walk LAYER IDs when computing nextItemID
READ WHEN: implementing any allocItemID / monotonic-ID logic; computing max(used IDs) across a parsed project; debugging head-counter collisions after structural mutation; reviewing initDerived or any function summing IDs across project state; AE 2025 rejects a Go-built file with "unexpected match name searched for in group"

---

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

## 第二回（2026-06-10）：service 层也在同一 namespace —— AE 2025 直接拒收

NewSolidLayer 暴露同一 scar 的更深层：comp 的 **service 层**（DLay / SLay×6 /
CLay×3 / SecL，dummy-comp 模板里恒占 ID 2..12；真实 AE 文件也有）**不进
`c.Layers`**（opaque 保留），所以修过的 initDerived 仍看不见它们。Go-built
工程（NewProject + NewComposition）的 nextItemID = 2 → `allocItemID` 发出
2/3 给 footage item + 层 → 与 service 层撞号。

**表象差异**：AE 2020 宽容（gate PASS）；**AE 2025 open 即抛
`内部验证错误 {unexpected match name searched for in group}`**（mode 3，JSX
catch 到）。同字节文件插进 parsed AE-native 工程（nextItemID 高）则双版本全
过 —— bisect 关键一刀。

**为何以前没炸**：真实 AE 文件 service ID 很小（comp 创建后立刻分配），永远
< 已有 max(ID)；唯一能让 nextItemID 落进 2..12 的是「除了 dummy comp 几乎空
白」的 Go-built 工程，而旧的结构性创建（shape/camera/light）层 ID 走
`maxLayerIDInItemList+1`（≥13）恰好绕开 —— `spliceLayerClone`/`allocItemID`
路径第一次踩上去。

**修复（两个 chokepoint）**：`initDerived` per-comp 加扫
`maxLayerIDInItemList(cb.itemList)`（盖住 parse / Reopen 路径）+
`NewComposition` 落 comp 后同样 bump（盖住 Go-built 路径）。副作用：多 comp
Go-built 工程的第二个 comp item ID 从 2 变 13（单调性不变，AE 无所谓）；
受影响 AE gate（V2_1 head counter 字节变化）复跑双版本 PASS。

## 第三回（2026-06-22）：跨-comp 用户层 ID 撞号 —— UI 编辑/删层炸（render/脚本全过）

Booyah comp ② 全工程复刻（多 comp 从零）暴露同一 scar 的又一面：**层 ID 用每-comp 计数**。
`NewShapeLayer` / `newTemplatedLayer` 用 `maxLayerIDInItemList(thisComp)+1` 发层 ID —— 每个 comp
的首个用户层都落在 **13**（紧跟模板 service 层 2..12）。一个工程里两个 comp 各有一个从零用户层
（comp① shape + comp② text）→ **两层共用 ID 13**。

**表象差异（关键）**：与前两回不同，这回 **AE open + render + 所有 ExtendScript op
（setValue/remove/duplicate/precompose/save）全过**，唯独**真机 UI 编辑文字 / 删图层抛
`内部验证失败 {unexpected match name searched for in group} (29::0)`**。机理：UI 删/改走
**ID-keyed 查找**（操作 + undo），撞号 ID 13 解析到**错的那个层**，于是在错层的属性组里搜匹配名搜不到
→ 报错；而脚本走**直接对象引用**（`comp.layer(1)`）绕过 ID 查找，故自动化复现不出来（坑了好几轮）。

**修复**：新 `allocLayerID(p, itemList)` = 全局单调 `allocItemID` + clamp 到 ≥ 本 comp service 层 max。
单 comp 工程返回值与旧公式**逐字节相同**（service-max+1），故已 ship 单-comp gate 不变（NewTextLayer
双版本 gate 复跑 PASS）；多-comp 才改（comp② text 层 13→15）。守卫：`TestNewLayer_CrossCompIDsDistinct`。
两个 chokepoint：`mutate_layer_new.go::NewShapeLayer` + `mutate_layer_camera.go::newTemplatedLayer`。

**教训**：render-pixel + Go round-trip + ExtendScript 全绿 **≠ AE UI 可编辑** —— UI 校验走 ID 查找比脚本严，
「结构有效能渲染但 ID 撞号」这类洞唯有**真机 UI 操作**能暴露（红线1 的 UI 变体；详 [[text-animator-create-re]] 末）。

**用户真机验收（2026-06-22）：修用户层 ID 后报错消失。** 故 **service 层（DLay/SLay…2..12）跨-comp 撞号
= 实证良性**：AE 视 service 层为 comp-内部、不参与全局唯一性校验（仅**用户层** ID 需全局唯一）。所以 NewComposition
不必 re-ID service 层（一度担心要做的更大改动，证实不需要）。即:全局唯一性只对 c.Layers 里的用户层强制。

## Related

- 修复 commit: 2026-05-28 V3 Phase 5B（同 commit 一起 ship）
- Phase 5B plan: [`../plans/2026-05-28-v3-phase5b-duplicatelayer-explicit-matte-plan.md`](../archive/plans/2026-05-28-v3-phase5b-duplicatelayer-explicit-matte-plan.md)
- 受影响代码: `internal/serializer/parse.go::initDerived`
- 受影响 invariant: Inv-9 (monotonic ID alloc, no reuse) in `incidents/ae25-acceptance-gate.md`
