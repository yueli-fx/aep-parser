# ⚠ PropertyBase Remove / Duplicate / MoveTo — AE 行为契约 + chunk 机制

PropertyBase Remove / Duplicate / MoveTo — AE 行为契约 + chunk 机制

RE'd 2026-06-03 (AE 2020, `17.7x45`) via `test_data/generators/re_property_struct.jsx`
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
复用**原 chunk 指针对** → opaque 内容原样带走。

## Duplicate 的 display-name 后缀（slice 2 已解 2026-06-03）

byte-check 结论（`tmp_debug/dump_prop_tdsn` 对 baseline vs duplicate fixture）：

- **baseline**：每个 effect 内层 tdgp **无 tdsn** —— 默认显示名（"高斯模糊"）
  从 match-name **运行时派生**，不存盘。
- **duplicate**：clone（源紧后那个）内层 tdgp **多一个 length-variable `tdsn → Utf8`**
  = "高斯模糊 2"；源仍无 tdsn。

即 **AE 把去重后缀存进了 clone 的 tdsn**，且 base 是 AE 的**本地化**名（"高斯模糊"
= Gaussian Blur 的中文显示名）。要字节级复刻这后缀需 AE schema/本地化 DB ——
跟 `Property.ValueText` 同 blocker。

**实现决策（不复刻后缀）**：clone = 源 `(tdmn, payload)` 对的 verbatim deep-copy，
**不注入 tdsn**。理由：无 tdsn 的 clone 字节上等价于"同一 effect Add 两次"——
AE 完全合法，打开时运行时重算去重名（" 2"）。那后缀纯 cosmetic、AE 自己会算。
结构性 duplicate 是忠实的。**双版本 ship-gate 实测证实**：AE 2020 + 2025 都接受
该 clone（无数据丢失），读回 4 effects 顺序正确，resave 保留。

## 落地状态

- **Remove + MoveTo + Duplicate**：`mutate_property_structural.go`，Go round-trip +
  AE 2020/2025 双版本 ship-gate（`property_structural_shipgate_test.go`：remove /
  move / duplicate × 双版本 = 6/6 PASS）。**Effect Parade + Text Animators 已双版本 gate**
  （Text Animators：2026-06-16，`text_animator_struct_shipgate_test.go`——3 动画器
  Opacity/Skew/Fill Color，remove-middle / move-last-to-front / duplicate-first × 双版本
  = 6/6 PASS，AE readback 动画器顺序逐项匹配 + Go round-trip）。Mask Parade / Root Vectors
  同机制、Go round-trip 过，但未单独 ship-gate → Alpha。
  - **gate gotcha（Text Animators readback）**：AE ScriptingAPI 的 `animator.property("ADBE
    Text Animator Properties").property(1)` 枚举**完整 ~120 leaf schema**（永远 = "ADBE Text
    Anchor Point 3D"），**非 materialized 子集**——按 index 取 tag 会假阴。verify jsx 必须按
    **非默认值**辨识每个动画器的 driven leaf（Opacity=50 / Skew=20 / Fill Color=blue）。
- Duplicate clone 的 flat mirror（`Layer.Effects` / `Layer.Masks`）走 clone chunk
  **重解**（`collectEffects` / `decodeMask`），back-ref 指向 clone chunk 不 alias 源。

相关：[[separate-dimensions-write-mechanics]]（同 tdmn+payload pair 机制的姊妹案例）、
[[ae-deletelayer-re]]（同 V2.1 atomic invariants pattern）。
