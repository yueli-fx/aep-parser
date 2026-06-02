---
status: active
when_to_read: implementing PropertyBase.Remove / Duplicate / MoveTo or any structural mutation of a property group's children; deciding whether a given property group permits add/remove/reorder; wiring the INDEXED_GROUP predicate
applies_to: [property, property-group, indexed-group, remove, duplicate, moveto, effect-parade, mask-parade, structural-write, p3-3c]
last_updated: 2026-06-03
---

# PropertyBase Remove / Duplicate / MoveTo — AE 行为契约 + chunk 机制

RE'd 2026-06-03 (AE 2020, `17.7x45`) via `test_data/re_property_struct.jsx`
（modes: baseline / remove / duplicate / move）。目标域 = canonical indexed
group **ADBE Effect Parade**（最常用，3 effects: Gaussian Blur / Tint / Fill）。

## 可变性谓词（最关键发现）

AE ScriptingAPI 对 `PropertyBase.remove()` / `.moveTo()` 的唯一门槛：

> **node 的 `parentProperty.propertyType` 必须 == `PropertyType.INDEXED_GROUP`**。

否则直接 throw（fixed Transform 叶子实测）：

```
Position.remove() -> THROW: "无法使用 remove 方法属性，因为父属性不是
INDEXED_GROUP（例如，parentProperty.propertyType != PropertyType.INDEXED_GROUP）"
```

propertyType 枚举值（实测）：`PROPERTY=6212` / `NAMED_GROUP=6213` /
`INDEXED_GROUP=6214`。

- Effect Parade.propertyType = **6214**（INDEXED_GROUP）；`canAddProperty('ADBE Fill')=true`
- Transform Group.propertyType = 6213（NAMED_GROUP）；`canAddProperty('ADBE Position')=false`
- effect 自身（Gaussian Blur 等）propertyType=6213 —— **但它是 INDEXED parent 的
  child**，所以可删/可移。谓词看的是 **parent**，不是 node 自身。

INDEXED_GROUP 不存盘 flag —— 由 **固定 match-name 集合** 判定（抄 py-aep
`_INDEXED_GROUP_MATCH_NAMES`，default = NAMED_GROUP）：

```
ADBE Effect Parade
ADBE Mask Parade
ADBE Effect Mask Parade
ADBE Text Animators
ADBE Root Vectors Group
```

**所有这些 indexed group 的 child 都是 group payload**（effect=sspc, mask atom=tdgp,
shape item=tdgp, animator=tdgp）—— 没有裸叶子。故 Remove/Duplicate/MoveTo 实现
挂在 `*AEPropertyGroup`（非 `*Property`）即覆盖 100% 真实场景。

## Chunk 机制（baseline→after byte diff，全部 pure pair 操作）

Effect Parade tdgp 直接 children 布局：

```
[0] tdsb (4B)          ← group subprop flags（不动）
[1] tdsn (14B)         ← group display-name struct（不动）
[2] tdmn "ADBE Gaussian Blur 2"
[3] LIST:sspc          ← effect 1 wrapper
[4] tdmn "ADBE Tint"
[5] LIST:sspc          ← effect 2 wrapper
[6] tdmn "ADBE Fill"
[7] LIST:sspc          ← effect 3 wrapper
[8] tdmn "ADBE Group End"   ← sentinel（不动）
```

| op | children 数 | AE 做了什么 |
|---|---|---|
| **Remove**(Tint) | 9→7 | 删掉 `tdmn "ADBE Tint"` + 其 `sspc` 这一对。**无 count chunk 联动** |
| **Duplicate**(GaussianBlur) | 9→11 | 在 source **紧后** 插入新 `tdmn "ADBE Gaussian Blur 2"` + `sspc`；**match-name 相同**；display name 跑出 " 2" 后缀（见下）。**无 count** |
| **MoveTo**(Fill→idx1) | 9→9 | Fill 的 tdmn+sspc 移到最前，其余后移。**纯 reorder，无 index chunk** |

→ 三个 op 都是父 group LIST 内 **tdmn+payload 对** 的 splice / insert / reorder。
**没有任何计数/索引 chunk 要维护**。完全被既有 `mergePosition`（删 pair）/
`separatePosition`（clone pair）helper 的机制覆盖。

实现策略 `rebuildIndexedGroupChunk`：按 mutate 后的 scene `Children` 顺序重发
parent.chunk.Children，保留 prefix（tdsb/tdsn）+ suffix（Group End），每个 child
复用**原 chunk 指针对** → opaque 内容原样带走（CLAUDE.md #5）。

## Duplicate 的 display-name 后缀（slice 2 待解）

`.duplicate()` 后 clone 的 **display name** = "高斯模糊 2"（源 = "高斯模糊"），
但 **match-name 不变**。display name 存在 effect 内层 `tdgp → tdsn → Utf8`
（length-variable）。**未确认** clone 的 tdsn 里是 AE 存了 " 2" 还是 runtime
auto-dedup（py-aep auto_name 逻辑暗示可能是 runtime 派生）。Duplicate slice 落地
前要先 byte-check 这点 —— 若 AE 存了，需 length-variable 改 Utf8；若 runtime 派生，
byte-identical clone 即可，AE 打开自动去重。

## 落地状态

- **Remove + MoveTo**：`mutate_property_structural.go`，Go round-trip + AE 2020/2025
  双版本 ship-gate（`property_structural_shipgate_test.go`）。Effect Parade 已 gate；
  Mask Parade / Root Vectors / Text Animators 同机制、Go round-trip 过，但未单独
  ship-gate → Alpha。
- **Duplicate**：deferred（display-name 后缀机制待 byte-check）。

相关：[[separate-dimensions-write-mechanics]]（同 tdmn+payload pair 机制的姊妹案例）、
[[ae-deletelayer-re]]（同 V2.1 atomic invariants pattern）。
