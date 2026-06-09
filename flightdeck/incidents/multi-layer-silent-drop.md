---
status: active
when_to_read: implementing any NewX layer-creation path that inserts a Layr into a comp Item LIST; debugging "AE accepts the file but comp.layers shows only 1 of N layers" silent-drop; adding multiple user layers to one composition; reviewing what sibling chunks delimit a layer's serialized unit
applies_to: [new-shape-layer, multi-layer, silent-drop, layer-siblings, fvdv, fiop, ftts, foac, fiac, fipc, fifl, itemList, ewst, ship-gate, ae2020, ae2025, structural-write]
last_updated: 2026-06-01
---

# 单合成多 ShapeLayer 被 AE silent-drop（layer-level sibling chunks 缺失）

## 症状

`NewComposition` + N×`NewShapeLayer`（同一 comp）后 `WriteAEP`。AE **接受文件**（不报"文件数据丢失"/不崩溃），但 `comp.layers.length == 1` —— 只保留**第一个** user layer，其余 silent-drop。单层（每 comp 1 个 ShapeLayer）一直正常，所以 V2.2 单层 ship-gate 没暴露。AE 2020 / 2025 同样 drop（非版本特定）。

## 根因

AE 写每个 user Layr 的序列化单元是：

```
LIST Layr            ← 图层本体
LIST Ewst            ← 空 0-child sibling
fvdv fiop ftts foac fiac fipc fifl   ← 第 1 组 layer-level siblings
fvdv fiop ftts foac fiac fipc fifl   ← 第 2 组（重复）
```

`NewShapeLayer` 旧实现只补了 `Ewst`，漏了后面 **14 个 layer-level sibling chunk**（两组 7 个）。AE 用这些 sibling **界定一个 layer 单元的结束**；缺了，AE 无法把第 2 个 Layr 起始与第 1 个分开，于是从 `comp.layers` 里 drop 掉第一个之后的所有层。单个 user layer 时它是 comp 内末位 layer，AE 宽容了，所以单层没暴露。

这与 `ae25-acceptance-gate.md` Stage 3 的 **Item-level** sibling（Fold 里每个 Item LIST 后跟 `FEE + fvdv/fiop/ftts/foac/fiac/fipc/fifl`）同源——同一套 `f*` 属性 chunk，只是这次在 **Layr level**（无 FEE，两组 7 个）。

## sibling 字节（跨 layer / 跨 AE 版本恒定）

| chunk | size | bytes |
|---|---|---|
| fvdv | 4 | `00 00 00 03` |
| fiop | 1 | `00` |
| ftts | 4 | `00 00 00 00` |
| foac | 1 | `00` |
| fiac | 1 | `00` |
| fipc | 2 | `00 00` |
| fifl | 4 | `00 00 00 00` |

两组，字节相同。AE 2020 baseline 写出的值在 AE 2025 原样接受。模板（`2020_dummy_comp.aep`）的 service layers（DLay/SLay/CLay/SecL）后面本就带这两组，user Layr 漏的就是它们。

## 定位过程（bisection，少走弯路用）

逐字段试错代价大，最终靠两刀切定位：

1. **byte-diff ldta**：ours 的 3 个 user Layr ldta 与 AE baseline **byte-identical**（连 ID 13/14/15、name 都一致）→ 排除 ldta。
2. **逐字段 hex-patch**（head counter B / idta@0x3a / cdta@0xb2,ba / idta@0x51-53）→ 全部无效，排除 comp 头部 metadata。
3. **round-trip bisection（决定性）**：用 parser 读 AE baseline（认 3 层）再 `WriteAEP` 写回 → AE 仍认 3 层。**证明 writer 无问题，缺陷在 builder**。
4. **结构 diff（builder comp vs round-tripped AE comp）**：归一化树 diff 暴露 ours 的 user Layr 缺 `fvdv/fiop/ftts/foac/fiac/fipc/fifl`，而 service layers 都有 → 锁定。

## 修复

- `internal/serializer/lower_layer_siblings.go` — `lowerLayerSiblings()` primitive，返回两组 7 个 sibling chunk（常量字节）。
- `internal/serializer/mutate_layer_new.go` — `NewShapeLayer` 插入单元从 `[Layr, Ewst]` 改为 `[Layr, Ewst] + lowerLayerSiblings()`（16 chunk），shift tail by n。
- `internal/serializer/new_shape_layer_test.go` — `TestNewShapeLayer_MultiLayer_EmitsLayerSiblings` 断言每个 user Layr 后跟 `Ewst + 14 siblings`。

双版本 ship-gate PASS（AE 2020 + 2025，3-layer + 8-layer 单合成均 `comp.layers == N`）。RE fixture：`test_data/re_multi_shape_layer.jsx`（AE 自建 3 shape layer baseline）。

## 教训

- **单层通过 ≠ 多层通过**：sibling/delimiter 类 boilerplate 在 N=1 时常被 AE 宽容，N≥2 才暴露。结构性写的 ship-gate 应至少覆盖 N=2。
- **silent-drop 不报错**：AE 接受文件但丢内容，比"文件数据丢失"更隐蔽。验证必须读回 `comp.layers.length` 实际值，不能只看"打开不报错"。
- **round-trip 是切 writer/builder 的利器**：parser 读 AE 文件 → writer 写回 → AE 验证，能把"writer 丢信息"和"builder 缺信息"一刀分开。
