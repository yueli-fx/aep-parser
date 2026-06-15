---
status: active
when_to_read: implementing or extending any shape vector-filter (Trim / Repeater / Round Corners / Offset / Merge / ZigZag / Pucker & Bloat / Twist / Wiggle Paths / Wiggle Transform) via AddTrim/AddRepeater/AddRoundCorners/AddOffsetPaths/AddMergePaths/AddPuckerBloat/AddTwist/AddWigglePaths/AddWiggleTransform/TrimNode/RepeaterNode/RoundCornersNode/OffsetPathsNode/MergePathsNode/PuckerBloatNode/TwistNode/WigglePathsNode/WiggleTransformNode; needing a UI-name↔match-name mismatch (Wiggle Paths = `ADBE Vector Filter - Roughen`, Wiggle Transform = `ADBE Vector Filter - Wiggler`); using canAddProperty to discover an unknown filter match-name; descending into a filter's nested Transform sub-group (Repeater / Wiggle Transform); needing the shape-stack render order (why a filter cuts/duplicates/rounds/grows/combines/bloats the shapes, incl. why a COMBINE filter needs the fill ABOVE it) or AE's ellipse path start vertex / winding; descending into a filter's nested group (Repeater Transform); deciding from-scratch vs embed-template for a new shape filter; reasoning about which filter sub-streams AE elides; a Repeater/Trim/RC/Offset/Merge/PB match-name that returns null; an ExtendScript addProperty live-ref going stale / ReferenceError when building a multi-shape fixture
applies_to: [trim-paths, repeater, round-corners, offset-paths, merge-paths, zigzag, pucker-bloat, vector-filter, shape-filter, ADBE-Vector-Filter-Trim, ADBE-Vector-Filter-Repeater, ADBE-Vector-Filter-RC, ADBE-Vector-Filter-Offset, ADBE-Vector-Filter-Merge, ADBE-Vector-Filter-Zigzag, ADBE-Vector-Filter-PB, merge-type, zigzag-size, zigzag-detail, puckerbloat-amount, trim-start, trim-end, trim-offset, trim-type, repeater-copies, repeater-transform, roundcorner-radius, offset-amount, shape-stack-order, fill-above-combine-filter, distort-filter, ellipse-path-winding, embed-template, findGroupBody, lower-shape-node, addproperty-stale-ref, mg-roadmap, s3, s5, ship-gate, ae2020, ae2025, render-pixel]
last_updated: 2026-06-15
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

## 复用确认 — Merge Paths（S5, 2026-06-13）✅ 蓝本第 6 次 + **两个新 ground truth**

`ADBE Vector Filter - Merge`：`AddMergePaths`/`MergePathsNode`，`templates/v2_2_shape_merge_body.bin`（376B/5 children）。单子流 `ADBE Vector Merge Type`（**非动画枚举** 1D f64 @cdat[0:8]：1=Merge 2=Add 3=Subtract 4=Intersect 5=Exclude，默认 1）→ 建模成普通字段（同 Fill blend mode，`overwriteShapeStreamCdat` 写枚举，**非** `lowerShapeScalar`）。

**新 ground truth ①（combine 型滤镜的 fill 位置反了）**：Merge 把下方多条 path 合成一条，**Fill 必须在 stack 顶（Merge 之上）** 才能画出合成结果——stack 顺序 `[Rect, Ellipse, Merge, Fill]`（Fill = Children 最高 index = 顶）。**Fill 放 Merge 下方渲染全黑**（fill 在合成前画了未合并的两条 path → ring 0/4 全暗）。这与 Trim/RC/Offset「paint 在 filter 下方」相反——因 Trim/RC/Offset 是改 path 几何、同一条 path 被下方 paint 引用；Merge 是「多 path→一 path」合成，需 paint 在合成之后。**别套前 5 个滤镜的 [shape,paint,filter] 直觉，render 出来看**：400×400 Rect − 200×200 同心 Ellipse(Subtract) → 白方块挖圆洞（眼验确认）。gate：ring 白 4/4 + 中心洞暗 3/3。

**新 ground truth ②（ExtendScript addProperty 返回的 live ref 会失效）**：`var r = sc.addProperty(...)` 返回的 PropertyBase live 引用，在**之后再 addProperty 兄弟属性时失效**（用它 `.property(...).setValue` 抛 ReferenceError）。修：先 add 完所有属性，再用 `sc.property(matchName)` 取（既有 trim/RC/offset fixture 就是这个模式，一直没踩坑因它们每次只在 add 后立即用一次）。两条同 match-name（两个 Rect）还会 `sc.property` 歧义——gate 用 Rect+Ellipse 异类避开。

Gate：`TestMGMerge_AEShipGate_AE2020/2025` 双版本渲染像素 PASS（Type=3 resave 读回 + 方块挖洞）。verify_mg_merge.jsx + mg_merge_shipgate_test.go。

**全 4 模式 gated（2026-06-15）✅**：`TestMGMergeModes_AEShipGate_AE2020/2025` 双版本 PASS——**Venn 判别床**一帧 4 卡（每卡两部分重叠椭圆 + Merge(mode) + Fill-on-top）覆盖 Add/Subtract/Intersect/Exclude。三区（左独占/重叠/右独占）签名全互异，一帧分清 4 模式：Add=W,W,W（花生/union）· Subtract(L−R)=W,暗,暗（左月牙）· Intersect=暗,W,暗（透镜）· Exclude=W,暗,W（双月牙/XOR）。**纯 gate 工作零新代码**（enum 写路径早通），但红线4 实渲眼验 4 模式全对（不靠「枚举写路径一致应同」假设）。verify_mg_merge_modes.jsx + mg_merge_modes_shipgate_test.go。**新工具法**：partial-overlap Venn pair = 一帧分清全部布尔模式的最省 gate 床（比 nested 强：nested 下 Subtract≡Exclude 同形）。

## 复用确认 — ZigZag（S5, 2026-06-13）✅ 蓝本第 7 次 — 常用矢量滤镜家族收齐

`ADBE Vector Filter - Zigzag`：`AddZigZag`/`ZigZagNode`，`templates/v2_2_shape_zigzag_body.bin`（708B/7 children）。探针 3 子流：`ADBE Vector Zigzag Size`（振幅 px，默认 5）+ `ADBE Vector Zigzag Detail`（每段隆起数/ridges，默认 10）+ `ADBE Vector Zigzag Points`（enum 默认 1，elide 未建模）。Size+Detail 双 1D scalar → `lowerShapeScalar`（同 Offset 单 headline，这里两个）。属 **distort 型**（改 path 几何），fill 在 filter 下方 stack `[Rect, Fill, ZigZag]`（与 Merge 的 combine 型相反，同 Trim/RC/Offset）。

渲染 ground truth：400×400 Rect + Size=40/Detail=8 → 每条边扭成尖齿（comic-book starburst，眼验确认实心内部+四边锯齿）。gate 用**逐列扫顶白 y 的 spread** 当锯齿证据（直边 spread≈0；zigzag spread=78 = ±40 振幅围绕原边 y=340，峰 y=301 谷 y=379）——比固定采样点稳（不依赖峰谷精确 x）。

Gate：`TestMGZigZag_AEShipGate_AE2020/2025` 双版本渲染像素 PASS（Size=40/Detail=8 resave 读回 + 锯齿边）。verify_mg_zigzag.jsx + mg_zigzag_shipgate_test.go。~~**deferred**：Points enum（Smooth/Corner）未建模（默认 elide）。~~ → **Points 已 ship**（synthesis-insert 第 3 次，见下）。

## 复用确认 — Pucker & Bloat（S5, 2026-06-13）✅ 蓝本第 8 次 — 同 Round Corners 最简

`ADBE Vector Filter - PB`：`AddPuckerBloat`/`PuckerBloatNode`，`templates/v2_2_shape_puckerbloat_body.bin`（412B/5 children）。探针单子流 `ADBE Vector PuckerBloat Amount`（1D f64 BE 百分比，默认 0=identity）→ `lowerShapeScalar`（含 animated flip）。无嵌套组、无 enum、无 elision（与 Round Corners 同构，唯二 headline-only 滤镜之一）。属 **distort 型**，fill 在 filter 下方 stack `[Rect, Fill, PuckerBloat]`（同 Trim/RC/Offset/ZigZag）。

渲染 ground truth：400×400 Rect + Amount=100（bloat）→ **四叶草/花瓣形**——每条直边外凸成圆瓣、四角全部内拉到中心（四瓣在 center 交汇）。gate 采样：center 白 + 四边外侧越界点白 4/4（边外凸越过原 400×400 边界）+ 四原始尖角内拉暗 4/4（角塌向中心）——精确 bloat 签名，与 plain rect / Round Corners / pucker / no-effect 全可区分。眼验四瓣花瓣确认（非数字 numerology）。

Gate：`TestMGPuckerBloat_AEShipGate_AE2020/2025` 双版本渲染像素 PASS（Amount=100 resave 读回 + 四瓣形）。verify_mg_puckerbloat.jsx + mg_puckerbloat_shipgate_test.go。负 Amount = pucker（凹/尖刺）同 slot 未单独 gate（同 vein 预期可推）。

## 复用确认 — Twist（S5, 2026-06-13）✅ 蓝本第 9 次 — 同 Round Corners/PuckerBloat 最简

`ADBE Vector Filter - Twist`：`AddTwist`/`TwistNode`，`templates/v2_2_shape_twist_body.bin`（402B/5 children，与 Round Corners **完全同构** = 单 headline scalar、无 enum、无嵌套组、无 elision 陷阱）。探针（gen_shape_twist.jsx）2 子流：`ADBE Vector Twist Angle`（**默认 10**，1D f64 BE 度数，headline）+ `ADBE Vector Twist Center`（Vec2 默认 [0,0]，elide 未建模）。只设 Angle 非默认 → 仅 Angle slot 发射 → `lowerShapeScalar`（含 animated flip）。NewTwistNode 默认 Angle=0（identity 无扭，语义 no-op；AE 默认是 10 但 lower 总覆写）。属 **distort 型**，fill 在 filter 下方 stack `[Rect, Fill, Twist]`（同 Trim/RC/Offset/ZigZag/PB）。

渲染 ground truth：400×400 Rect + Angle=150 → **风车/螺旋**——twist 中心钉住、离中心越远旋转越少（直边被扭成弧、四角甩离原轴对齐位）。眼验 silhouette（grid dump）确认非轴对齐方块。gate 用 twist 的**独有签名 = 左右镜像不对称**（mirror-asym=44 about x=center）：plain/pucker/bloat 方块全镜像对称 → broken/no-op twist 渲成对称方块会同时栽「mirror-asym≥20」+「四原始角清空 4/4」两条——比纯数值 round-trip 强（轴对称方块=假绿）。center 白钉住。

Gate：`TestMGTwist_AEShipGate_AE2020/2025` 双版本渲染像素 PASS（Angle=150 resave 读回 + 风车 + mirror-asym=44 两版一致）。verify_mg_twist.jsx + mg_twist_shipgate_test.go。commit bfa1674。~~**deferred**：Twist Center（Vec2，默认 elide）未建模。~~ → **Center 已 ship**（synthesis-insert 第 4 次 = 首个 Vec2 leaf，见下）。

## 复用确认 — Wiggle Paths（S5, 2026-06-13）✅ 蓝本第 10 次 — 常用矢量滤镜家族完整收齐

`ADBE Vector Filter - Roughen`（**UI 叫 Wiggle Paths，内部 match-name 是 Roughen**——发现型探针 `canAddProperty` 多候选自证）：`AddWigglePaths`/`WigglePathsNode`，`templates/v2_2_shape_wiggle_body.bin`（1326B/11 children）。8 子流（全 1D scalar），建模 4 个 headline：`ADBE Vector Roughen Size`（振幅 默认 10）+ `Roughen Detail`（默认 10）+ `ADBE Vector Temporal Freq`（=Wiggles/Second，默认 2）+ `ADBE Vector Random Seed`（默认 0）→ 各走 `lowerShapeScalar`（含 animated flip）。**deferred**：`Roughen Points`（enum）· `ADBE Vector Correlation`（默认 50）· `Temporal Phase` · `Spatial Phase`（默认 elide 未建模）。NewWigglePathsNode 默认 Size=0=identity（无扰）。属 **distort 型**，stack `[Rect, Fill, Wiggle]`（同 Trim/RC/Offset/ZigZag/PB/Twist）。

**新 ground truth（时间随机性滤镜）**：Wiggle 是**逐帧随机**（`Temporal Freq` 控制 churn 速度），但**单帧由 Random Seed + 相位决定性**——frame 0 渲出固定的毛糙边。gate 渲帧 0：400×400 Rect + Size=60/Detail=30/WPS=4/Seed=9 → 边缘毛糙噪声边界（眼验 silhouette：顶边 y=340 有缺口、y=310 外凸尖、y=760 下凸刺）。gate 用**顶边逐列 topmost-white-y 的 spread** 当毛糙证据（干净方块 spread≈0；wiggle spread=41，topY∈[312,353]）——**spread=41 两版（AE2020/2025）完全一致 = seed 决定性跨版本可复现**，比纯数值 round-trip 强（干净方块=假绿）。

Gate：`TestMGWiggle_AEShipGate_AE2020/2025` 双版本渲染像素 PASS（4 值 resave 全读回 + 毛糙边 spread=41 两版一致）。verify_mg_wiggle.jsx + mg_wiggle_shipgate_test.go。commit 753f1b3。

## 复用确认 — Wiggle Transform（S5, 2026-06-13）✅ 蓝本第 11 次 — **矢量滤镜家族全部收齐（vein 闭合）**

`ADBE Vector Filter - Wiggler`（UI 名 Wiggle Transform）：`AddWiggleTransform`/`WiggleTransformNode`，`templates/v2_2_shape_wiggletransform_body.bin`（2040B/9 children）。递归探针自证：5 顶层 scalar + 1 嵌套 Transform 组。建模 = 顶层 `ADBE Vector Xform Temporal Freq`(=Wiggles/Second，默认 2) + `ADBE Vector Random Seed`(默认 0)（各 1D scalar 含 animated）+ 嵌套 `ADBE Vector Wiggler Transform` 组（同 Repeater 经 `findGroupBody` descend + `overwriteShapeStreamCdat` 静态覆写）的 4 通道**抖动幅度**：`Wiggler Anchor`/`Position`/`Scale`(Vec2 @cdat[0:16]) + `Wiggler Rotation`(1D @cdat[0:8])。**deferred**：`Correlation`(默认 50)/`Temporal Phase`/`Spatial Phase`（默认 elide）。`WigglerTransform` 字段全 static（同 RepeaterTransform）。NewWiggleTransformNode 默认全零幅度=identity（不抖）。

**坑**：`Wiggler Scale`/`Anchor`/`Position` 默认值是 **[0,0]（抖动幅度，非绝对 transform 值）**——Scale 不是 [100,100]，零幅度=该通道不抖。

渲染 ground truth：Wiggle Transform 把整个 transform 随机抖动应用到下方 path（典型用在 Repeater 后散开副本）。gate 用**单形状**隔离：200×200 白 Rect 标称中心 (960,540) + Position 幅度 [220,220]/Rotation 70/Seed 8 → frame 0 白块**质心被位移到 (1092.9,404.6)、离中心 189.7px**——**两版（AE2020/2025）质心完全一致 = seed 决定性跨版可复现**。no-op（零幅度）质心居中 disp≈0。gate 阈值 disp≥50。

Gate：`TestMGWiggleTransform_AEShipGate_AE2020/2025` 双版本渲染像素 PASS（6 值 resave 全读回 + 质心位移 189.7px 两版一致）。verify_mg_wiggletransform.jsx + mg_wiggletransform_shipgate_test.go。commit df0b874。

## Animated filter scalar — Trim End reveal gated (2026-06-14)

静态滤镜 11 个收齐后，**animated 滤镜 scalar 路径**（`lowerShapeScalar`→`injectAnimatedStream`，与 Rect Roundness/Stroke Opacity/Fill Opacity 共享）一直**已 wired 但从未在渲染面 gate**——incident 早先那句「animated 自动经 injectAnimatedStream … 天然支持」是**未验证假设**（交付准则：Go round-trip ≠ AE 接受；可渲染能力须红线4 像素 gate）。

`TestMGTrimAnim_AEShipGate_AE2020/2025` PASS（commit 002264d）：单 shape 层 ellipse+白 stroke+Trim，Trim End 关键帧 0→100 over [t=0,t=4]，**渲染同一层三帧**断言 ring 扫开：
- t=0 (End=0)：空（left+right+bottom 缺；top 起始顶点可能留点，不断言）
- t=2 (End=50)：右半（top/right/bottom 在，left 缺）— AE 椭圆顶部起、顺时针
- t=4 (End=100)：整圈（left+right 都在）

**单层三帧的 left 侧单调 reveal（缺→缺→在）= 动画证据**——排除掉「层被 drop」和「静态 End=100」两种假绿。resave 重 parse 确认 Trim End 存活为 2-keyframe **bpk-48 1D 非空间**容器。双版本 PNG 逐字节一致（10581/19377/26276）。verify_mg_trimanim.jsx + mg_trimanim_shipgate_test.go。

**Repeater Copies/Offset · Offset Amount · PuckerBloat/Twist/RC Amount 等所有 headline scalar 走完全相同 `lowerShapeScalar` 路径**——AE 消化 animated filter scalar 已由本 gate 证明，各自 animated gate 按需补（边际价值低，非重复 RE）。**注**：animated gradient 色标**不**属此类（GCky 是独立 keyframe 容器，未 RE + JSX 无法 authoring，见 `gradient-fill-write-re.md` § Scope）。

## Offset Copies — shape-filter synthesis-insert 首发（优先级3，2026-06-15）✅

Offset Paths 的 4 个 elided 子流（Line Join / Miter / **Copies** / Copy Offset）现只模 headline Amount。**Copies** 用 **synthesis-insert** 落地——把 `SetMaterialOption`（material leaves）那套机制**首次推广到矢量滤镜 body**：

- **不污染默认**：`v2_2_shape_offset_body.bin`（仅 Amount slot）保持不变；仅当 `SetCopies` 被调用且值 ≠1 时，clone `ADBE Vector Offset Copies` leaf（tdmn + LIST:tdbs，6-child scalar-with-range，与 Amount 同构）splice 进 body，覆写 cdat。与 Dashes 的「solid body 不变 + 第二模板」同精神，但更轻（splice 单 leaf，不需整 body 变体）。
- **leaf 模板** `templates/v2_2_shape_offset_copies_leaf.bin`（318B，LIST(tdgp) 裹单个 (tdmn,tdbs) pair）由 `tmp_debug/extract_offset_copies_leaf` 从 `v2_2_offset_copies.aep`（JSX 设 Amount=40+Copies=3，逼 AE 不 elide）抽出。RE 确认落盘子序：Amount → Copies → GroupEnd（Line Join/Miter/Copy Offset 仍 elide）。
- **splice 落点** = `spliceShapeLeafBeforeGroupEnd`（GroupEnd 前；Copies canonical 在 Amount 后）。scene `OffsetPathsNode` 加 `copies float64 + copiesSet bool`（静态，非 PropertyStream——Copies 非动画需求）。
- **不需 hydration**：`collectShapeKids` 本就不 hydrate 任何 filter 节点（Trim/Merge/Offset 全靠 opaque chunk 穿越 Reopen）——Copies splice 在 from-scratch lower 期烘进字节，之后 Reopen→Write 作 opaque 保留。与所有现有 filter 一致。

**渲染 ground truth（先看图，红线4）**：400×400 Rect + **even-odd Fill** + Offset(Amount=40, Copies=3) → 3 副本叠在 half-width 240/280/320，even-odd 填出**白心方块(<240) + 暗 gap 环(240..280) + 白外环(280..320)**。gate 签名 = 外白环白(d≈300) + gap 暗(d≈260) + 白心——**单副本（任意大小实心方块）不可能有白→暗→白径向跳变**（实心方块=假绿，靠 even-odd 同心环识破）。`TestMGOffsetCopies_AEShipGate_AE2020/2025` 双版本 PASS（outer-ring 4/4 · gap 4/4 · centre 3/3 · resave Copies=3，png 10777b 两版一致）。

**蓝本推广**：此机制（clone leaf 模板 → spliceShapeLeafBeforeGroupEnd → 覆写 cdat，仅当非默认）直接复用于 **Trim Type**（elide 无 slot）· Offset 其余 3 子流 · ZigZag Points · Twist Center · Repeater Order 等所有 default-elided 矢量滤镜子流。`spliceShapeLeafBeforeGroupEnd` 是通用 helper（GroupEnd 前插 pair）；多 leaf 需 canonical 排序时再扩 ordinal map（同 `materialLeafOrder`）。

## Trim Type — synthesis-insert 第 2 次（enum leaf，优先级3，2026-06-15）✅

`ADBE Vector Trim Type`（AE「Trim Multiple Shapes」，默认 1=Simultaneously elide，2=Individually）用同 Offset Copies 的 synthesis-insert 落地——蓝本第 2 次套用，**确认对 enum leaf 同样成立**：

- scene `TrimNode += trimType TrimType`（`TrimTypeSimultaneously=1`/`TrimTypeIndividually=2`，`SetType`/`Type`）。lower 当 `Type != Simultaneously` 时 clone `ADBE Vector Trim Type` leaf（`v2_2_shape_trim_type_leaf.bin`，286B，**4-child tdbs = enum**，无 tdum/tduM）经 `spliceShapeLeafBeforeGroupEnd`（Offset 那次建的通用 helper）插入 + 覆写 cdat。
- RE：`tmp_debug/gen_shape_trim_type.jsx`（2 椭圆 + stroke + Trim End=50 + Type=2 逼 AE 不 elide）→ `v2_2_trim_type.aep` → `tmp_debug/extract_trim_type_leaf` 抽 leaf。落盘子序确认 Start/End/Offset/**Type**（Type 在 GroupEnd 前，splice 落点对）。

**渲染 ground truth（先看图，红线4——这次纠正了直觉）**：2 椭圆（屏 740/1180）+ stroke + Trim End=50：
- **Individually(2)** = 路径**按序** trim → 50% 时**左椭圆满圈、右椭圆空**（第一条 path = 前 50%，恰好满）。
- **Simultaneously(1，默认)** = 两 path **同步**各 50% → **两椭圆各右半弧**。
- 注意：UI 名与直觉相反，「Individually」反而是 sequential（左满右空），「Simultaneously」才是「各自半弧」。**别按字面推，render 出来看**（对比 demo `gen_trim_type_demo` 上下两行实锤）。
- gate 双判据：左椭圆**左缘(640)白**（满圈；Simul 此处暗）+ 右椭圆**右缘(1280)暗**（空；Simul 此处白）——组合排除 Simultaneously / 无trim / 掉层全部假绿。`TestMGTrimType_AEShipGate_AE2020/2025` 双版本 PASS（left-ring 4/4 · right-empty 4/4 · resave Type=2，png 20470b 两版一致）。

**synthesis-insert 蓝本现验 2 类 leaf**：scalar-with-range（Offset Copies，6-child）+ enum（Trim Type，4-child）。`spliceShapeLeafBeforeGroupEnd` + `overwriteShapeStreamCdat` 对两者通用。剩 ZigZag Points / Twist Center / Repeater Order / Offset Line Join·Miter·Copy Offset 等同路径按需。

## ZigZag Points — synthesis-insert 第 3 次（enum leaf，优先级3，2026-06-15）✅

`ADBE Vector Zigzag Points`（AE「Points」下拉，**默认 1=Corner elide，2=Smooth**）用同 Trim Type 的 synthesis-insert 落地——enum leaf 蓝本第 2 次套用确认稳定：

- scene `ZigZagNode += points ZigZagPoints`（`ZigZagPointsCorner=1`/`ZigZagPointsSmooth=2`，`SetPoints`/`Points`）。lower 当 `Points != Corner` 时 clone `ADBE Vector Zigzag Points` leaf（`v2_2_shape_zigzag_points_leaf.bin`，286B，**4-child tdbs = enum**，与 Trim Type leaf 同字节大小/结构）经 `spliceShapeLeafBeforeGroupEnd` 插入 + 覆写 cdat。落盘子序 RE 确认 Size/Detail/**Points**/GroupEnd（Points 在 GroupEnd 前，splice 落点对）。
- RE：`tmp_debug/gen_shape_zigzag_points.jsx`（Rect+Fill+ZigZag，Size=40/Detail=8/Points=2 逼 AE 不 elide + 渲染眼验）→ `v2_2_zigzag_points.aep` → `tmp_debug/extract_zigzag_points_leaf` 抽 leaf。probe 自证：默认 `Points=1`，3 子流 Size/Detail/Points。

**渲染 ground truth（先看图，红线4）**：一帧两卡同 Size=70/Detail=5、唯 Points 不同——左 **Corner(默认)** = 尖角 starburst（sharp 三角刺），右 **Smooth(2)** = 圆滑 scallop blob（rounded 鼓包、无尖点）。控制变量到只剩 Points，任何像素差全归因于 Points。gate 判据 = **apex 带宽**（顶边中段逐列 topmost-white-y，距全局最高 y ≤8px 的列数）：sharp 窄尖顶 flat 少、rounded 宽顶 flat 多——CORNER flat=22 vs SMOOTH flat=65（~3x），require ≥2x 余量。两卡 spread 同(138/140)=同振幅（证 zigzag 都活、排除「smooth 没 splice 退回 corner」）。`TestMGZigZagPoints_AEShipGate_AE2020/2025` 双版本 PASS（flat 22/65 两版一致 · png 58629b 两版逐字节同 · resave Points=2 存活）。verify_mg_zigzag_points.jsx + mg_zigzag_points_shipgate_test.go。

**synthesis-insert 蓝本累计 3 次全绿**：scalar-with-range（Offset Copies）· enum（Trim Type · **ZigZag Points**）。剩 Twist Center（Vec2，需 cdat[0:16] 而非 enum）/ Repeater Order / Offset Line Join·Miter·Copy Offset 等同路径按需。

## Twist Center — synthesis-insert 第 4 次（**首个 Vec2 leaf**，优先级3，2026-06-15）✅

`ADBE Vector Twist Center`（Vec2，默认 [0,0] elide）用同蓝本落地——确认 synthesis-insert **对 Vec2 同样成立**（前 3 次是 scalar/enum 的 cdat[0:8]，本次 cdat[0:16] 两 f64 BE）：

- scene `TwistNode += center [2]float64 + centerSet bool`（`SetCenter`/`Center`/`CenterSet`，static Vec2 同 Repeater Transform 字段，非动画）。lower 当 `CenterSet && Center≠[0,0]` 时 clone `ADBE Vector Twist Center` leaf（`v2_2_shape_twist_center_leaf.bin`，294B，**4-child tdbs**——Vec2 同 enum 的 4-child，无 tdum/tduM range chunk）经 `spliceShapeLeafBeforeGroupEnd` 插入 + `overwriteShapeStreamCdat(…, encodeF64sBE(c[0],c[1]))` 写 16B。落盘子序确认 Angle/**Center**/GroupEnd。**Vec2 写路径零新代码**——`overwriteShapeStreamCdat` 早被 Repeater/Wiggler Position 等 Vec2 用过，cdat 长度由 data 决定。
- RE：`tmp_debug/gen_shape_twist_center.jsx`（Rect+Fill+Twist，Angle=180 + Center=[150,0] 逼 AE 不 elide + 渲染眼验）→ `v2_2_twist_center.aep` → `tmp_debug/extract_twist_center_leaf`。probe 自证默认 Center=[0,0]。

**渲染 ground truth（先看图，红线4）**：一帧两卡同 Angle=180、唯 Center 不同——左 **CENTERED([0,0]默认)** = 点对称风车（两卷须 180° 对称、质心钉死中心）右 **OFFSET([150,0])** = 非对称（圆鼓大瓣在 pivot 侧上-右、尖卷在下-左、质心被甩离）。gate 判据 = **白像素质心位移**：CENTERED disp=0.8px（点对称证 twist 本身对称→位移全归因 Center）· OFFSET disp=119px（require CENTERED<25 + OFFSET≥50）。`TestMGTwistCenter_AEShipGate_AE2020/2025` 双版本 PASS（质心 0.8/119 两版一致 · png 44951b 两版同 · resave Center=[150,0] 16B Vec2 存活）。verify_mg_twist_center.jsx + mg_twist_center_shipgate_test.go。

**synthesis-insert 蓝本累计 4 次全绿 = 三类 leaf 全覆盖**：scalar-with-range（Offset Copies，6-child）· enum（Trim Type · ZigZag Points，4-child）· **Vec2（Twist Center，4-child cdat[0:16]）**。`spliceShapeLeafBeforeGroupEnd` + `overwriteShapeStreamCdat` 对三类通用。剩 Repeater Order（enum）/ Offset Line Join·Miter·Copy Offset（enum+scalar）等同三类之一，按需。

## Offset Line Join / Miter Limit / Copy Offset — synthesis-insert 第 5–7 次 + **多 leaf canonical 排序**（优先级3，2026-06-15）✅

Offset Paths 余下 3 个 elided 子流一次收齐——确认**单节点多 leaf 共存**的 canonical 排序：

- **canonical 落盘序**（RE 自证 `gen_shape_offset_extras.jsx` 全 5 子流非默认）：Amount → **Line Join** → **Miter Limit** → Copies → **Copy Offset** → GroupEnd。`spliceShapeLeafBeforeGroupEnd` 每次插在 GroupEnd 前 = **按调用序追加**，故 lowerOffsetPathsNode 只要**按 canonical 序依次 splice**（Line Join→Miter→Copies→Copy Offset）即天然有序，无需 ordinal map。
- scene `OffsetPathsNode +=` lineJoin(`StrokeLineJoin` 复用，Miter1/Round2/Bevel3) · miterLimit · copyOffset（各带 `*Set` flag，static）。Line Join leaf 286B/4-child enum；Miter/Copy Offset 各 318B/6-child scalar（同 Offset Copies）。leaf 用 `tmp_debug/extract_offset_leaf <aep> <leafName> <out>`（参数化，一 fixture 抽 3 leaf）。
- 各 setter 仅当**非默认**才 splice（Line Join≠Miter · Miter≠4 · CopyOffset≠1）——不污染默认 Amount-only body。

**渲染 ground truth（先看图，红线4）**：一帧 5 卡——上排 3 个 Rect200+Offset Amount=60 角形卡：**MITER(默认)** 尖角 miter 点（角尖像素 corner-fill 4/4 白）· **BEVEL(LineJoin=3)** 八边形切角（corner-fill 0/4）· **MITERLIM(Miter Limit=1)** 同切角（90°miter 比 1.41>limit1 → 裁成 bevel，corner-fill 0/4）；下排 2 个 Rect160+Amount=30+Copies=3 卡：**COPY1(默认 CopyOffset=1)** 右伸 172px · **COPY2(CopyOffset=2)** 右伸 202px（copy 步进外移一个 Amount）。`TestMGOffsetExtras_AEShipGate_AE2020/2025` 双版本 PASS（corner 4/0/0 · extent 172/202 两版一致 · png 12391b 两版同 · resave Line Join=3/Miter=1/Copy Offset=2/Copies=3 全存活）。verify_mg_offset_extras.jsx + mg_offset_extras_shipgate_test.go。眼验 5 卡（尖角 vs 八边形 ×2、小方 vs 大方）。

**synthesis-insert 蓝本累计 7 次全绿**：Offset Copies · Trim Type · ZigZag Points · Twist Center · **Offset Line Join · Miter Limit · Copy Offset**。三类 leaf（scalar/enum/Vec2）+ 单节点多 leaf canonical 排序全验。Offset Paths 5 子流**全收齐**。剩 Repeater Order（enum，可达性待核）/ Wiggle 调制参数（低价值）按需。

## Repeater Order — synthesis-insert 第 8 次 + **splice-before-group** + 「可达但视觉无效」实证（优先级3，2026-06-15）✅

`ADBE Vector Repeater Order`（Composite，Below=1 默认/Above=2）。**先纠错**：早先 Explore 审计称「JSX `property()` 返 null → 不可达」是**误判**——`gen_shape_repeater_order.jsx` 探针自证 Order 是 Repeater child[3]、真实 enum 属性、`setValue(2)` 成功读回 2。incident 原话「同 Trim Type」正确。

- **canonical 落盘序**：Copies → [Offset] → **Order** → Transform(组) → GroupEnd。Order 在 Transform **组之前**，故不能用 `spliceShapeLeafBeforeGroupEnd`（会落到 Transform 组后）——新增 `spliceShapeLeafBefore(body, "ADBE Vector Repeater Transform", …)`（插在指定 tdmn 前，缺省 fallback GroupEnd）。这是蓝本首次需要「splice 在某子组前」而非 GroupEnd 前。
- scene `RepeaterNode += order RepeaterOrder + orderSet`（`RepeaterOrderBelow=1`/`RepeaterOrderAbove=2`，`SetOrder`/`Order`/`OrderSet`）。leaf `v2_2_shape_repeater_order_leaf.bin` 286B/4-child enum（同 Trim Type）。`tmp_debug/extract_filter_leaf <aep> <filterMN> <leafMN> <out>`（通用化抽取器）。

**「视觉无效」实证（红线4 的诚实边界）**：Order 是**合成顺序**设置——同色填充下数学上**不可见**（同色多层 over 合成可交换：两层色 C、alpha a/b → 结果亮度 C(a+b(1-a)) 对 a,b 对称）。故**无像素可判别**。gate 据此诚实验证：一帧两卡（BELOW 默认 + ABOVE Order=2，repeater 几何全同：5 copies 重叠+缩放+opacity 100→20），断言**白覆盖恒等**（BELOW=7341 == ABOVE=7341 两版本完全一致）+ ABOVE resave Order=2 存活 + AE 接受。`TestMGRepeaterOrder_AEShipGate_AE2020/2025` 双版本 PASS。**这不是「假绿」——是 Order 本就无可见效果的正确实证**（眼验两卡渲染逐像素同）；over-claim 一个可见效果才违反交付准则。verify_mg_repeater_order.jsx + mg_repeater_order_shipgate_test.go。

**synthesis-insert 蓝本累计 8 次全绿**：+ Repeater Order。机制全谱：三类 leaf（scalar/enum/Vec2）· canonical 多 leaf 排序（Offset）· splice-before-group（Repeater Order）。**常用 shape 矢量滤镜的所有可写 elided 子流全收齐**。剩 Wiggle Paths/Transform 的 Correlation/Temporal·Spatial Phase（调制参数，低价值，待评估）。

## Wiggle 调制参数（Correlation / Temporal·Spatial Phase / Roughen Points）— 评估后 evidence-based defer（2026-06-15）

`gen_wiggle_modulation_probe.jsx` 探针自证 Wiggle Paths（Roughen）8 子流 + match-name + 默认全确认（建模 4：Size/Detail/Temporal Freq/Random Seed；**未建模 4**）：
| 子流 | 默认 | 类型 | 可达 |
|---|---|---|---|
| `ADBE Vector Roughen Points` | 1=Corner | enum | ✓（探针在 tree，可 setValue） |
| `ADBE Vector Correlation` | 50 | scalar | ✓（探针 `setValue(0)` 成功） |
| `ADBE Vector Temporal Phase` | 0 | scalar | ✓（real property） |
| `ADBE Vector Spatial Phase` | 0 | scalar | ✓（real property） |

Wiggle Transform（Wiggler）同族残留 `Correlation`/`Temporal Phase`/`Spatial Phase`（同义）。

**defer 依据（非「不可达」，是「按价值/可门禁性 evidence-based defer」）**：
1. **Temporal/Spatial Phase = 噪声函数相位参数**：固定帧改 phase = 读 seed 决定的噪声序列的不同采样 → 产出**统计同质的另一随机边**，无「正确像素」可断言，只是「另一随机实例」（同 Random Seed 的性质——已在 `TestMGWiggle` 单次门禁证 seed 决定性，不按值穷举）。**本质不可像素门禁**（无 categorical 判别量）。
2. **Correlation / Roughen Points = 可门禁但低价值**：Correlation（0=独立抖/锯齿、100=相干/平滑）+ Points（Smooth/Corner）确有视觉效果，但作用在**已混沌的 roughen 边**上，门禁信号弱（噪声边度量、不如 ZigZag 干净齿）；cockpit 明标低价值。leaf 形状（scalar/enum）已被蓝本证 8 次，**defer 不丢任何机制覆盖或能力类**，只搁具体低价值旋钮。
3. **方向实质收口**：synthesis-insert 机制全谱已证（三类 leaf + 多 leaf 排序 + splice-before-group），**所有 clean-visual elided 旋钮已 ship+gate**（ZigZag Points · Twist Center · Offset Line Join/Miter/Copy Offset · Repeater Order）。残留 = 本质不可门禁（Phase）或低价值噪声调制（Correlation/Points）。

实现路径现成（按需）：Points/Correlation/Phase 各 splice 对应 leaf（enum/scalar，`spliceShapeLeafBeforeGroupEnd`，注意 Roughen 子流 canonical 序 Size/Detail/Points/TemporalFreq/Correlation/TemporalPhase/SpatialPhase/Seed——Points 在 Detail 后，Correlation/Phase 在 Seed 前）。未做。

**家族小结（蓝本 11 次全绿 — vein 闭合）**：Trim · Repeater(+嵌套 Transform 组) · RoundCorners · Offset · Merge(combine·fill 在上) · ZigZag · Pucker&Bloat · Twist · Wiggle Paths · **Wiggle Transform(+嵌套 Transform 组)**——**所有常用 shape 矢量滤镜全部收齐，cdat-based vein 已无候选**。三步蓝本（probe→抽 body→cdat 覆写）+「nested 组 findGroupBody descend」对全部成立；唯二变量 = ① 子流集合/elision 边界（先 all-non-default fixture 逼 AE 不 elide）② **combine 型 fill 位置反**（Merge 需 fill 在 stack 顶）。新增工具法：**未知 filter match-name 用 `canAddProperty` 多候选发现 + 递归 walk dump 嵌套组**。
