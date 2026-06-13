---
status: active
when_to_read: implementing or extending any shape vector-filter (Trim / Repeater / Round Corners / Offset / Merge / ZigZag) via AddTrim/AddRepeater/AddRoundCorners/AddOffsetPaths/TrimNode/RepeaterNode/RoundCornersNode/OffsetPathsNode; needing the shape-stack render order (why a filter cuts/duplicates/rounds/grows the shapes) or AE's ellipse path start vertex / winding; descending into a filter's nested group (Repeater Transform); deciding from-scratch vs embed-template for a new shape filter; reasoning about which filter sub-streams AE elides; a Repeater/Trim/RC/Offset match-name that returns null
applies_to: [trim-paths, repeater, round-corners, offset-paths, vector-filter, shape-filter, ADBE-Vector-Filter-Trim, ADBE-Vector-Filter-Repeater, ADBE-Vector-Filter-RC, ADBE-Vector-Filter-Offset, trim-start, trim-end, trim-offset, trim-type, repeater-copies, repeater-transform, roundcorner-radius, offset-amount, shape-stack-order, ellipse-path-winding, embed-template, findGroupBody, lower-shape-node, mg-roadmap, s3, s5, ship-gate, ae2020, ae2025, render-pixel]
last_updated: 2026-06-13
resolved_by:
---

# Trim Paths (`ADBE Vector Filter - Trim`) — 矢量滤镜走既有 shape-body 模板 vein

## Signature
- symptom: 无 bug（clean slice）。本条是 RE 记录 + 后续 shape-filter（Repeater/Merge/Offset/Round/ZigZag）的复用蓝本
- where: internal/serializer/lower_shape_node.go `lowerTrimNode` + templates/v2_2_shape_trim_body.bin · internal/scene/scene_shape_graph.go `TrimNode`
- trigger: MG roadmap S3 — 线描 reveal（stroke 半圈）从零生成

## RE 发现（v2_2_trim.aep，AE 2020；tmp_debug/gen_shape_trim.jsx）

Trim Paths 是 **Vectors Group 内、和 shape/fill/stroke 平级的一个矢量滤镜节点**（不是 layer effect）。
match-name `ADBE Vector Filter - Trim`，body = 标准 `LIST(tdgp)`，4 个子流（probe 自证）：

| 子流 match-name | 默认 | 存储 | 建模 |
|---|---|---|---|
| `ADBE Vector Trim Start` | 0 | f64 BE @cdat[0:8]，**原始百分比**（20→20.0） | ✅ Start() |
| `ADBE Vector Trim End` | 100 | 同上（70→70.0） | ✅ End() |
| `ADBE Vector Trim Offset` | 0 | f64 BE @cdat[0:8]，**度数**（30→30.0） | ✅ Offset() |
| `ADBE Vector Trim Type` | 1 (Simultaneously) | — | ❌ 默认被 AE elide（无 cdat slot，同 gradient/taper 的省略陷阱） |

→ 全部套既有 embed-body vein：`extract_shape_bodies` 抽 body（988B/9 children）→ `lowerTrimNode` clone +
`lowerShapeScalar` 覆写 Start/End/Offset 的 cdat。静态走 cdat 覆写；animated 自动经 `injectAnimatedStream`
flip 成 1D 非空间关键帧容器（和 Rect Roundness / Fill Opacity 同路）——line-draw reveal（End 0→100 keyframe）天然支持。

## 两个非显然 ground truth（render 出来看，不靠推理）

1. **Shape-stack 渲染顺序**：trim 要剪到 stroke，add 顺序必须 **[Ellipse, Stroke, Trim]**（`lowerVectorGroup` 里
   Children[0]=底=先处理，Children[len-1]=顶=后处理；bottom-up = path→stroke→trim）。RE fixture 就是这个顺序，
   渲染 v2_2_trim.aep 确认 stroke 被剪成弧（不是整圈）。**别按「operator 影响下方」的直觉空想**——直接 saveFrameToPng 看。
2. **AE 椭圆 path 起点/绕向**：起点在 **顶部 12 点**、**顺时针**。故 Start20/End70（50% 跨度）= 底部「U」弧（顶部两端开口）；
   Start0/End50 = 右半圈（top→right→bottom 在、left 不在）。gate 即用 left-present(FULL) vs left-absent(HALF) 当 trim 证据。

## Gate（红线4 渲染像素双版本）

`TestMGTrim_AEShipGate_AE2020/2025` PASS：FULL（End100）整圈 L+R 都在；HALF（End50）top/right/bottom 在、**left 不在**；
trim End resave 读回 50/100 存活。verify_mg_trim.jsx + mg_trim_shipgate_test.go。

## 复用蓝本

Repeater / Merge Paths / Offset Paths / Round Corners / ZigZag 都是同类矢量滤镜节点——预期同 vein：
probe match-name → 抽 body 模板 → clone+cdat 覆写。唯一变量是各自的子流集合与 elision 边界（先用 all-non-default fixture 逼 AE 不 elide）。

## 复用确认 — Repeater（S5, 2026-06-12）✅ 蓝本成立

`ADBE Vector Filter - Repeater` 一刀套上述 vein 落地（`AddRepeater`/`RepeaterNode`，`templates/v2_2_shape_repeater_body.bin`）。body 结构（gen_shape_repeater.jsx RE）：
- **顶层 1D**：`ADBE Vector Repeater Copies`（f64 raw count）+ `ADBE Vector Repeater Offset`（起始 copy index）→ `lowerShapeScalar`（含 animated）。
- **`ADBE Vector Repeater Order`（Composite enum）默认 elide 无 slot**——未建模（同 Trim Type）。
- **嵌套 `ADBE Vector Repeater Transform`（tdgp 子组）**：descend 经 `findGroupBody`（同 Stroke Taper/Wave），内含 Anchor/Position/Scale（Vec2 @cdat[0:16]）+ Rotation/Opacity 1/Opacity 2（1D @cdat[0:8]）→ `overwriteShapeStreamCdat` 静态覆写。

两个 Repeater 专属坑：
1. **match-name 是 `ADBE Vector Repeater Anchor`，不是 `...Anchor Point`**（JSX setValue 用错名 → `property()` 返 null → 「null 不是对象」）。Position/Scale/Rotation/Opacity 1/Opacity 2 名如其字。
2. **probe JSX 递归进 Transform 子组时 `"" + pr.value` 对 Vec2 数组抛「数字结果无效（除以零？）」**（valueOf 数值转换陷阱，[[effect-param-elision-synthesis-lite]] finding 4 同源）——dump 数组值必须 `value.join(",")`，否则抛在 save 之前 → 整个 fixture 没存成。

Gate：`TestMGRepeater_AEShipGate_AE2020/2025` PASS——一个白点 Copies=5 + Transform Position=[300,0] → 渲染帧 5 个点在 x=360..1560、4 个间隙全暗；Copies/Position resave 读回。verify_mg_repeater.jsx + mg_repeater_shipgate_test.go。

→ 蓝本三步（probe→抽 body→cdat 覆写）+ 「nested 组用 findGroupBody descend」对带子组的滤镜成立；Merge/Offset/Round/ZigZag 照此推进。

## 复用确认 — Round Corners（S5, 2026-06-13）✅ 蓝本第 4 次成立（迄今最简滤镜）

`ADBE Vector Filter - RC` 一刀套蓝本三步落地（`AddRoundCorners`/`RoundCornersNode`，`templates/v2_2_shape_roundcorners_body.bin`）。body 结构（gen_shape_roundcorners.jsx 探针自证）：
- **单子流** `ADBE Vector RoundCorner Radius`（1D f64 BE @cdat[0:8]，**原始像素**，默认 10）→ `lowerShapeScalar`（含 animated flip）。无嵌套组、无 enum、无 elision 陷阱——是迄今最干净的滤镜（body 仅 402B/5 children = tdsb+tdsn+tdmn+tdbs-list+GroupEnd）。
- 探针设值 60（非默认）逼 AE 不 elide → 抽出 Radius cdat slot。

渲染 ground truth（render 出来看，红线4）：stack 顺序 **[Rect, Fill, RoundCorners]**（filter 在 stack 顶）→ Round Corners 圆掉 Rect path 的角，Fill 填出圆角卡片。400×400 白 Rect + Radius=150 → 渲染成 squircle（直边在、四角被切）。gate 即用「内部+四直边中点白(5/5) vs 四原始尖角被切暗(4/4)」当圆角证据——比纯数值 round-trip 强（尖角不切=假绿）。**别空想 filter 影响上/下方，直接 saveFrameToPng 看**（本次一次命中 [Rect,Fill,RC]，但仍渲染确认）。

Gate：`TestMGRoundCorners_AEShipGate_AE2020/2025` 双版本渲染像素 PASS（Radius=150 resave 读回 + squircle 像素）。verify_mg_roundcorners.jsx + mg_roundcorners_shipgate_test.go。AE 2025 首跑 exit-2 冷启动 splash（内置 warm-retry 也栽），手动 warmup_quit.jsx 预热后同样 PASS——确认是冷启 flake 非数据 reject（reject 会 warm 后再栽）。

## 复用确认 — Offset Paths（S5, 2026-06-13）✅ 蓝本第 5 次成立

`ADBE Vector Filter - Offset` 同 RC 路：`AddOffsetPaths`/`OffsetPathsNode`，`templates/v2_2_shape_offset_body.bin`（408B/5 children）。探针（gen_shape_offset.jsx）见 **5 子流**：`ADBE Vector Offset Amount`（**默认 10**，非 0！1D f64 px，headline）· `Offset Line Join`(默认1=Miter) · `Offset Miter Limit`(4) · `Offset Copies`(1) · `Offset Copy Offset`(1)。只设 Amount 非默认 → 仅 Amount slot 发射，其余 4 个 elide（同 RC 模 headline 1 个、其余暂搁）→ `lowerShapeScalar`。

渲染 ground truth：[Rect, Fill, Offset] → Offset 把 Rect path 向外长（+Amount px/边，负值缩）。400×400 白 Rect + Amount=60 → 渲染成 ~520×520 方（每边 +60）。gate 用「四原始边外侧一圈带变白(grown 5/5) + offset 外更远点仍暗(bounded 4/4)」当增长证据——区分「没长(原边外暗)」「长太多/失控(远点白)」。眼验：方块明显变大、直边在、有界。默认 Line Join=Miter 角基本尖（轻微圆是 AE offset 外角行为）。

Gate：`TestMGOffset_AEShipGate_AE2020/2025` 双版本渲染像素 PASS。verify_mg_offset.jsx + mg_offset_shipgate_test.go。AE 2025 同样先 warmup_quit.jsx 预热避冷启 exit-2。
