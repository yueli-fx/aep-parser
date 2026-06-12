---
status: active
when_to_read: implementing or extending NewPrecompLayer / nested-comp layers; debugging "precomp renders wrong/empty comp" or "AE drops the nested comp"; reasoning about how a layer references a comp source (ldta @0x28 SourceID); adding any newTemplatedLayer-based layer that needs a post-clone ldta setter (SetSource/SetParent); building a coincidence-proof source-ID gate
applies_to: [new-precomp-layer, precomp, nested-composition, source-id, ldta, 0x28, set-source, newTemplatedLayer, layer-backref, av-layer, structural-write, cycle-guard, mg-roadmap, s4, ship-gate, ae2020, ae2025, render-pixel, id-coincidence]
last_updated: 2026-06-12
resolved_by:
---

# Precomp 层 = AV 层 + ldta @0x28 SourceID → CompItem（复用 newTemplatedLayer）

## Signature
- symptom: 无外部 bug；一个内部 false-green 险些发生（见下「ldta backref 坑」）
- where: internal/serializer/mutate_layer_precomp.go `NewPrecompLayer` · mutate_layer_camera.go `newTemplatedLayer`（backref 加 ldta/nameChunk）
- trigger: MG roadmap S4 — comp 套 comp（工程结构刚需），纯 Go 生成嵌套合成

## RE 发现（re_precomp.aep，AE 2020；tmp_debug/gen_precomp.jsx）

`parent.layers.add(childComp)` 生成的 precomp 层就是**普通 AV 层**——4-child Layr（ldta 160B + Utf8 name + LIST(tdgp) Transform + LIST(Gide)），与 camera/light/shape 同形。**唯一把它标成 precomp 的是 `ldta @0x28 SourceID 指向一个 CompItem**（而非 footage item）。提取的 Layr：layerID=28 sourceID=1(child) parentID=0 **无 tdpi**（无宿主层绑定要 remap）。

→ 源 comp **已存在于项目**，所以不像 Solid/Null/Adjustment 要造/导 footage item。直接套 **camera/light 的 `newTemplatedLayer`**（embed 单 Layr + splice + 原子回滚），clone 后把 SourceID 改指向用户的 child comp 即可。parser 早已建模（`Layer.SourceComposition()` 经 `SourceID` 解到 CompItem；`inferLayerType` 给 AV 型）。

## ldta backref 坑（差点 false-green，gate-fixture-ID-coincidence 的同族）

`newTemplatedLayer` 原本只把 `layrList` 塞进 `layerBackrefs`，**不塞 `ldta`**——所以 clone 后 `layerBack(clone).SetSource(child.ID)` 报「no ldta chunk」**静默失败**，SourceID 保持模板里的 stale 值 1。偏偏 sanity 里 child 是第一个建的 comp（ID==1）→ 读回「正确」解到 child = **纯 ID 巧合**。修：`newTemplatedLayer` 构造 backref 时带上 `ldta`（已在手）+ `nameChunk`，SetSource 才真生效（字节输出对 camera/light 无变化——只多了内存指针，既有 gate 仍成立）。

**gate 防巧合**：parent **先**建（拿低 ID）、child **后**建（ID≠模板 stale 1），断言 `childR.ID > 1`。SetSource 若再次静默失败，precomp 会指向 ID 1（parent 自身/service 层）→ 渲不出 child 绿 → gate 红。

## 其它机制点
- **cycle guard**：`compReachable` 从 child 沿 precomp source 边 DFS，撞到 parent 即拒（AE 不允许循环合成引用）。`seen` 防图里既有环。
- **时间重置**：模板携 fixture child 的 span；`newTemplatedLayer` 重置到 0→parent duration。
- **anchor 陷阱**（gate 踩过）：模板 ldta/transform 的 anchor 来自 fixture（1920×1080 child = 960,540）。gate child 用 1920×1080 与模板同尺寸即 1:1 对齐；child 尺寸不同则 anchor 偏移需另设（V2.2 未参数化）。

## Gate（红线4 渲染像素双版本）
`TestMGPrecomp_AEShipGate_AE2020/2025` PASS：parent 唯一层 = child 的 precomp，AE 读回 source 是 CompItem「MGPREC_Child」；渲染帧 child 绿方块在 center、dark-blue BG 在 corner 都透出；resave 后 SourceComposition 仍解到 child。verify_mg_precomp.jsx + mg_precomp_shipgate_test.go。

## 遗留 / 复用
- precomp 层 anchor/scale 未参数化（child≠1920×1080 时需手设 transform）；collapse transformation / time-remap 未做。
- S6 端到端组合 gate 将叠 precomp + ease + trim 多要素；多层嵌套（祖孙）未单独 gate（cycle guard 已覆逻辑）。
