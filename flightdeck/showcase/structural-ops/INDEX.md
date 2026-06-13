---
showcase: structural-ops
direction: 图层结构性 op — DuplicateLayer/MoveLayer/DeleteLayer/SetDimensionsSeparated，纯 Go 应用后 AE 实渲 + .done readback 双重验证
capabilities: [duplicate-layer, move-layer, delete-layer, set-dimensions-separated]
gates: [TestDeleteLayer_AEShipGate, TestDuplicateLayer_AEShipGate, TestMoveLayer_AEShipGate]
status: complete
last_updated: 2026-06-13
regenerate: "go run ./flightdeck/showcase/structural-ops  +  scripts/ae_run.ps1 render.jsx"
---

# structural-ops — 图层结构性 op showcase

## 这个方向测什么

对一组**居中同心方块**应用 4 种结构性 op，AE 渲染 + `.done` readback **双重验证**。op 全是 facade 自由函数：`aep.DuplicateLayer(c,i,name)` / `aep.MoveLayer(c,from,to)` / `aep.DeleteLayer(c,i)` / `aep.SetDimensionsSeparated(prop,true)`（需 parsed 层，`Reopen` 后）。

## ⚠ 真实边界（诚实标注 · 关键）

**Solid 层无法摆位**——solid 模板的 Position 是 merged 模式但无 leader（只有 Position_0/_1），AE 把所有 solid 堆到 comp 中心，只栈顶可见（这是项目已知硬限制，`layers` showcase 同款）。**故无法用「四象限分开摆位」直观演示**，改用**同心靶环**（不同尺寸+颜色+栈序，唯一 AE 认的 solid 杠杆）。结果：**Move / Delete 有像素特征可眼验，Duplicate / Separate 单帧无像素差、靠 `.done` readback 佐证**。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go, tracked) | 构建 structural_ops.aep（8 层同心环 + BG） |
| render.jsx | 渲染脚本(tracked) | `saveFrameToPng(0)` + dump 层 roster + DELETE/DUPLICATE/MOVE/Separate 四项 readback → png/.done |
| structural_ops.aep | 产出工程 (gitignored) | 1920×1080，AE2020 target |
| structural_ops.png | 渲染帧 (gitignored) | frame 0 |

## 布局（同心靶环，圆心 = 画面中心）

| 环（外→内） | 色 | op | 像素可见？ |
|---|---|---|---|
| 最外 | green | **Separate**（D = shape 层 Position 分离） | 框在（分离不可见，`.done` 证 `dimensionsSeparated=true`） |
| 大 | pink | **Delete**（删中间绿环 → 后面粉透出） | ✅ **中间无独立绿环** = Delete 生效 |
| 中 | teal | **Duplicate** | 不可见（克隆与源完全重叠，`.done` 证两层都在） |
| 中心 | violet | **Move**（violet 移到 amber 之前） | ✅ **中心是紫非琥珀** = Move 生效 |

> 审核要点：① 中心方块**紫色**（Move）；② **无中间独立绿环**、被粉填满（Delete）；③ `.done` 四项 readback 全 `true`（DELETE C_mid_green absent / DUPLICATE 两层在 / MOVE violetIdx<amberIdx / Separate separated=true）。Duplicate/Separate 单帧无像素特征是 solid 不可摆位的必然，靠 readback 验。

## 溯源

V3 结构性 op（adaptive block-splice + atomic rollback）：`flightdeck/plans/coverage.md` § "V3 Composition layer structural ops"；维度分离：`incidents/separate-dimensions-write-mechanics.md`。
