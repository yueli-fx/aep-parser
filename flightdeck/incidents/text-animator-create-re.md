---
status: active
when_to_read: implementing/extending text animators (AddTextOpacityAnimator / AnimateTextRangeOffset); adding a new animator property type (Position/Scale/Color); wondering where Text Animators live vs btdk; building a property *Property over a spliced chunk to reuse AnimateScalarKeyframes
applies_to: [text, text-animator, kinetic-typography, range-selector, ADBE Text Animators, ADBE Text Animator, synthesis-insert, indexed-group, AnimateScalarKeyframes, mutate-text-animator]
last_updated: 2026-06-16
resolved_by:
---

# Text Animators (kinetic typography) — create-from-scratch RE + 落地

RE'd + shipped 2026-06-15 (用户加项「effects 之后做文字」的真前沿——文字基础早 ship，
真缺的是 Text Animators 引擎)。`AddTextOpacityAnimator` + `AnimateTextRangeOffset`，
**AE 2020 + AE 2025 双版本渲染像素 ship-gate PASS**（`text_animator_shipgate_test.go`）。

## 结构（RE 自 `re_text_animator.jsx`，AE 2020/2025）

Text Animators 住**图层属性树**（tdgp），**不在 btdk**（btdk 只是字符串 + layout 缓存）。
嵌套路径：

```
Layr → (outer tdgp) → tdmn "ADBE Text Properties" → tdgp
  ├ ADBE Text Document (btds/btdk)
  ├ ADBE Text Path Options
  ├ ADBE Text More Options
  └ tdmn "ADBE Text Animators" → tdgp   ← INDEXED group（6214）= splice 目标
       └ tdmn "ADBE Text Animator" → tdgp  ← 一个动画器（splice 的 payload）
            ├ tdmn "ADBE Text Selectors" → tdgp (INDEXED)
            │    └ tdmn "ADBE Text Selector" → tdgp  ← Range Selector
            │         ├ ADBE Text Percent Start / End / Offset (1D scalar)
            │         ├ ADBE Text Index Start / End / Offset
            │         └ ADBE Text Range Advanced (Units/Mode/Shape/Smoothness/…)
            └ tdmn "ADBE Text Animator Properties" → tdgp  ← 可动画叶子
                 └ ADBE Text Opacity / Position 3D / Scale 3D / Fill Color / …（~120 个）
```

propertyType: 6214 INDEXED_GROUP · 6213 NAMED_GROUP · 6212 PROPERTY。

## 关键 finding

1. **重度 elision**：AE 只存非默认值。fixture 设 Opacity=0 + Range End=50 → 存盘只剩
   这两个 cdat，Start/Offset/全部 Index*/全部 Range Advanced/其余 ~119 个属性全 elide。
   故模板用 `RE_TXANIM_MODE=template`（Start=20/End=80/Offset=10/Opacity=0 **全非默认**）
   生成，让每个 cdat slot materialize，`AddTextOpacityAnimator` 再按参数覆写。
2. **scalar 布局** = LIST:tdbs{tdsb, tdsn, tdb4(124B), cdat(40B), tdum, tduM}，
   **值在 cdat[0:8] BE f64**（40490000…=50.0 印证）。
3. **从零 NewTextLayer 不带 Animators 组**——只有 Document/Path Options/More Options
   （`probe_text_animator_scaffold`）。故首个动画器 splice **整个** `ADBE Text Animators`
   组进 Text Properties（在 Group End 前）；后续动画器 append 进既有组。
4. **animated offset = 标准 1D 非空间关键帧流**（`RE_TXANIM_MODE=animated` ground truth）：
   offset 变 `LIST:list{lhd3(52B,count@0x08,bpk@0x10=0x30=48), ldat(2×48)}`，value@0x08
   time@0x00，tdb4 @0x05 由 01→00 翻转——**与 `AnimateScalarKeyframes` 逐字节一致**。
   故扫光复用既有 1D 动画机制，零新关键帧代码。

## 落地（`mutate_text_animator.go`）

- 模板 = 外层 `ADBE Text Animators` 组（含一个 opacity 动画器，4 slot 全在），
  `extract_text_animator` 抽到 `templates/text_animators_opacity_body.bin`（embed-AE-bytes，
  同 AddEffect/shape-body vein）。
- `AddTextOpacityAnimator(layer, opacity, start, end, offset)`：定位 Text Properties chunk →
  无 Animators 组则 splice 整组 / 有则 append 内层动画器 → 按 match-name 递归覆写 4 个
  cdat[0:8]。原子 undo。
- `AnimateTextRangeOffset(layer, tickRate, kfs)`：定位 offset 的 tdbs → `parseLeafProperty`
  建带 back-ref 的 *Property → `AnimateScalarKeyframes`（把 spliced chunk 包成 parsed
  Property 是复用现有动画机制的通法，可推广到任意 spliced scalar）。

## ship-gate（红线4 渲染像素）

`text_animator_shipgate_test.go`：从零 comp + 全帧 BG + 文字「ABCDEF」+ Opacity-0 动画器
（Start=0/End=100, Offset 关键帧 0→100）。**颜色/位置无关签名** = 全帧亮度 spread（max−min）：
- t=0 offset=0 → 全选中 → opacity 0 → 全隐 → 均匀 BG → **spread=0**
- t=1 → 半揭示 → **spread=151**
- t=2 offset=100 → 全显 → 字形出现 → **spread=168**

单调揭示（隐→现，同一层跨帧）排除「掉层」+「静态」。AE 2020/2025 双版本 PASS，readback
确认 animator + Range Selector + offset 2 关键帧(0→100) + opacity=0，resave offset 存活
为 2-kf bpk-48 容器。文字层 Position 不经 Go accessor 暴露（fresh+reopened 都 nil），gate
用 verify JSX 设位（采样 fixture 关切，非被测能力）。

## Position 3D 动画器扩展（2026-06-15，第二个 leaf 类型）

同 vein 扩 Position（kinetic typography 的 slide-in / drop-in），双版本渲染 gate PASS
（`text_animator_position_shipgate_test.go`）。关键：

1. **vtype 表**（自 postemplate dump，`RE_TXANIM_MODE=postemplate`）：Opacity=6417(1D)、
   Position 3D=**6413(spatial 3D)**、Scale 3D=6414(3D)、Rotation=6417(1D)、Fill Color=6418(color)。
2. **spatial 3D cdat = 72B**：三个 BE f64 @ [0:8]/[8:16]/[16:24]（=x/y/z），尾 48B 全 0
   （spatial tangent 槽）。tdb4 头 `db990003000f0003`（3 分量），tdbs 无 tdum/tduM（scalar 才有）。
   故 `overwriteVectorCdat(root, name, vals)` 按 `cdat[8*i:]` 写 N 个 double，通杀任意分量数。
3. **每 leaf 类型一个模板**（elision）：`templates/text_animators_position_body.bin`
   （postemplate 模式：Position=[30,-40,0] + Start/End/Offset 全非默认 → 全 slot materialize）。
   `extract_text_animator <src> <out>` 已加参数化。
4. **动画复用零新代码**：经典 slide-in = **静态** Position 位移 + Range Offset 关键帧扫光
   （`AnimateTextRangeOffset` 操作 Range Selector，与被驱动 leaf 类型无关）。字形始终可见（opacity 100），
   扫光时整块**垂直位移** → gate 签名 = ink 垂直质心单调迁移（t0=187→t1=312→t2=447，Δ=260px，
   AE2020/2025 逐像素一致）。证 MOTION（区别于 Opacity gate 的 appearance）。
5. **机制泛化**：splice 逻辑抽 `spliceTextAnimator(tp, tmpl)`、Range 抽 `setRangeSelector`、
   模板缓存改按 body 指针的 `sync.Map`（`animatorTemplate(body)`）。下一个 leaf 类型 = 抽模板 +
   一个 facade，复用全部。

## Scale 3D 动画器扩展（2026-06-15，第三个 leaf 类型）

`AddTextScaleAnimator`，双版本渲染 gate PASS（`text_animator_scale_shipgate_test.go`）。
泛化机制零改动，只抽模板 + 一个 facade（验证了 Position 留下的泛化设计）。findings：

1. **Scale 3D cdat = 120B**（≠ Position 的 72B；vtype 6414 ThreeD 非空间）：三 BE f64 @ [0:24]
   = sx/sy/sz，后 96B 扩展槽（per-dim min/max ease 等），tdbs 多一个 tdum 下界。`overwriteVectorCdat`
   仍只写 [0:24]，长度无关——通用。
2. 模板 `templates/text_animators_scale_body.bin`（`RE_TXANIM_MODE=scaletemplate`：Scale=[40,60,100]
   + Start/End/Offset 全非默认）。
3. **gate 签名 = ink 面积**（区别于 Position 的质心、Opacity 的 spread）：Scale=220% shrink-in，
   扫光时字形从 220%→100% → ink 像素数单调缩（t0=1070→t1=693→t2=313，AE2020/2025 一致）。
   证 SIZE 变化，文字全程可见。

## Rotation 动画器扩展（2026-06-15，第四个 leaf 类型）

`AddTextRotationAnimator`，双版本渲染 gate PASS（`text_animator_rotation_shipgate_test.go`）。
Go 侧真·免费（Rotation = vtype 6417 1D scalar，cdat 40B f64 @ [0:8]，**与 Opacity 同布局**，
degrees 直存——`overwriteScalarCdat` 照搬）。难点全在 gate 签名：

- **旋转不改面积/亮度/质心** → 上述签名全失效。解法：**单个不对称字符 "L" 的 ink 包围盒长宽比
  (w/h)**。0°→竖(w/h<1)、90°→横(w/h>1)、45°→近方，长宽比单调。verify JSX 设 fontSize=240
  放大单字以稳定测量。t0=1.52→t1=0.81→t2=0.65（AE2020/2025 一致），方向无关、mid 居中。
- 模板 `templates/text_animators_rotation_body.bin`（`RE_TXANIM_MODE=rottemplate`：Rotation=45
  + Start/End/Offset 全非默认）。Rotation tdbs 无 tdum/tduM（4 children，≠ Opacity 的 6）。

## Fill Color 动画器扩展（2026-06-15，第五个 leaf 类型）

`AddTextColorAnimator(layer, r, g, b, a, rangeStart, rangeEnd, rangeOffset)`，双版本渲染 gate
PASS（`text_animator_color_shipgate_test.go`）。泛化机制零改动，只抽模板 + 一个 facade + 复用
`overwriteVectorCdat`。findings：

1. **Fill Color = vtype 6418 color，cdat = 96B**（12 BE f64）：颜色 `[A,R,G,B]×255` @ [0:32]，
   后 64B（8 f64）全 0。**与 shape Fill/Stroke 同一 on-disk 编码**（`encodeShapeColorBE`），不是 effect-param 的
   ARGB 变体。RE 自 `RE_TXANIM_MODE=colortemplate`：Fill=[0.2,0.4,0.8,1] → disk [255,51,102,204] 逐字节印证 order。
   tdbs 只有 tdsb/tdsn/tdb4(124B,4 分量)/cdat —— **无 tdum/tduM**（同 Position/Rotation；scalar 才有）。
2. facade 入参 r,g,b,a 为 0..1，`overwriteVectorCdat(payload, name, []float64{a*255, r*255, g*255, b*255})`
   只写 [0:32]，长度无关——通用（同 Scale 120B 只写 [0:24]）。
3. 模板 `templates/text_animators_color_body.bin`（colortemplate：Fill 非默认 + Start/End/Offset 非默认）。
4. **gate 签名 = ink 像素均值 RGB 的绿通道**（区别于 Opacity 亮度 / Position 质心 / Scale 面积 / Rotation 长宽比）：
   红色覆写 + 白底文字，扫光时字形从红→白。verify JSX 强制 base `fillColor=[1,1,1]` 使红覆写成唯一颜色信号。
   t0=(248,0,0)红 → t1=(241,120,120)粉 → t2=(236,236,236)白，绿通道单调 0→120→236、红通道全程高
   （AE2020/2025 逐像素**完全一致**）。证 COLOR（静态红会全程绿低、掉层无 ink）。

## animate leaf 本身（2026-06-15，1D scalar：Opacity/Rotation）

`AnimateTextOpacity` / `AnimateTextRotation`，双版本渲染 gate PASS
（`text_animator_leaf_anim_shipgate_test.go`）。**新机制维度**：此前所有动画靠 Range Offset 扫光、
leaf 值静态；这条让 **leaf 值本身关键帧化**——全部被选字符**同步**走一条值曲线（脉冲/连续旋转/淡入淡出，
扫光做不到）。findings：

1. **零新关键帧代码**：定位第一个 animator 的 leaf tdbs（`scalarTdbs(animators, matchName)`）→
   `parseLeafProperty` 包成带 back-ref 的 *Property → `AnimateScalarKeyframes`。**与 `AnimateTextRangeOffset`
   逐字一致**，只是目标从 Range Offset 换成被驱动 leaf（通法 `animateTextScalarLeaf`）。
2. **wiring 反转 = 证据**：gate 把 Range Offset **留静态**（numKeys=0）、Rotation leaf 关键帧化（numKeys=2），
   正好是 spin-in offset-sweep gate 的镜像 —— 证明动的是 leaf 而非 selector。单字 "L" 旋 0→90，
   aspect (w/h) t0=0.65→t1=0.81→t2=1.52 单调（AE2020≡AE2025 逐像素一致）。
3. **前置依赖**：leaf 必须已存在（先 AddText*Animator）且仍 static（已 animated 用 InsertKeyframe）。
   非文字层 / 无 animator / 无该 leaf 均 refuse。
4. **3D/4D leaf animate（Position/Scale/Color）= 下一 slice**：走 `AnimateVectorKeyframes`（spatial block，
   color [A,R,G,B]×255），但该路径 RE 自 **effect** color/point params，text leaf 的 animated block layout
   未验证——**须独立 gate**，不可假定同 effect。本 slice 只交付已证的 1D scalar 路径（红线：不堆未验证机制）。

## animate 3D/4D leaf（2026-06-15，Position 3D + Fill Color；Scale 3D 暂搁）

`AnimateTextPosition`（[x,y,z] px）+ `AnimateTextColor`（[r,g,b,a] 0..1）。双版本渲染 gate PASS
（`text_animator_vec_leaf_anim_shipgate_test.go`）。**RE-first**：先抽 AE ground truth
（`RE_TXANIM_MODE=animatedvec`：Position/Scale/Color 各 keyframe 2 帧）逐字对照 `AnimateVectorKeyframes`，
**不假定同 effect**。findings（每 leaf 的 animated keyframe block，bpk=lhd3 @0x10）：

| leaf | vtype | bpk | value 偏移 | header @0x04 / marker @0x08 | 匹配 `vectorKeyframeLayout` |
|---|---|---|---|---|---|
| Position 3D | 6413 spatial | 128 | **0x38** | 0007 / **3** | ✓ `(3)` spatial marker3 |
| Fill Color | 6418 color | 152 | **0x38** | 0001 / **2** | ✓ `(4)` color marker2，值=[A,R,G,B]×255 |
| **Scale 3D** | 6414 ThreeD | 128 | **0x08** | 00 / **无 marker** | ✓ `AnimateVectorKeyframesNonSpatial`（非 spatial 块） |

1. **Position/Color 逐字匹配 effect color/point 的 spatial block** → 复用 `AnimateVectorKeyframes` 零改动
   （`animateTextVectorLeaf` = scalarTdbs→parseLeafProperty→AnimateVectorKeyframes）。单测断言 bpk + value@0x38
   逐字节核对 AE ground truth（128/Position·152/Color），再双版本 render gate。
2. **Scale 3D = 非 spatial 多维块**（value@0x08、headerByte 0x00、无 marker，≠ Position 的 spatial）。
   **codec 早已支持**：`bytesPerKeyframe` 非 spatial = `0x08+5*dim*8`（dim3=128 ✓）、`writeKeyframeBlock`
   非 spatial 分支 value@0x08 + per-component ease。只是 `AnimateVectorKeyframes` 硬编码 spatial layout。
   解法 = 抽 `animateVectorKeyframes(p,tr,kfs,nonSpatial)` 私有核 + `AnimateVectorKeyframesNonSpatial`
   公开入口（layout = `{dim, headerByte:0x00, spatial:false}`）。**零新字节代码**——复用既有非 spatial 编码器。
   `AnimateTextScale` 走它，gate 签名 = ink 面积（100→150 ⇒ 2940→5879 px，AE2020≡AE2025）。
3. **on-disk 单位**：Position 值原样存 px（kf2 [100,50,0] 存 raw），Color facade 把 [r,g,b,a]0..1 转 [a,r,g,b]×255
   （与 `AddTextColorAnimator` 一致）。gate 签名：Color = ink 均值 R 降 B 升（248/0→0/248）；Position = ink 垂直质心
   下移（367→617，Δ=250px）。AE2020≡AE2025 逐像素一致。

## 现状 / 边界

- **已 ship**：Opacity、**Position 3D**、**Scale 3D**、**Rotation**、**Fill Color** 动画器、Range Selector（参数化
  Start/End/Offset）、offset 关键帧扫光（reveal / slide-in / shrink-in / spin-in / colour-wipe 通用）、
  **animate leaf 本身（全覆盖：1D scalar `AnimateTextOpacity`/`AnimateTextRotation` + 3D/4D spatial `AnimateTextPosition`/`AnimateTextColor` + 非 spatial 3D `AnimateTextScale`，全字符同步值曲线）**。各双版本渲染 gate PASS。
  **animate-leaf 方向已收口**——每个有 Add\*Animator facade 的 leaf 类型都能 animate。
- **Alpha**：动画器叶子无 typed accessor（chunk-only，Reopen 后属性树重建但 animator 叶子
  不带 back-ref，同 `property-indexed-group-structural-re.md` 的 Root Vectors）。多动画器 append +
  Opacity/Position 混排已测。
- **免费近邻已收口（2026-06-16，见下节）**：Fill Opacity / Stroke Opacity / Stroke Width / Stroke
  Color / Skew 五个双版本渲染 gate PASS；Rotation X/Y evidence-based defer（2D 视觉惰性）。
- **selector 家族已收口（2026-06-16）**：Range Advanced（`SetTextRangeAdvanced`，Amount gated）· 多 Selector
  （`AddTextRangeSelector`，gated）· Wiggly（`AddTextWigglySelector`，gated）· **Expressible**
  （`AddTextExpressibleSelector`，2026-06-17 render-gated，原 defer 翻案）全 ship。**animate leaf 已全覆盖**——1D 复用
  `animateTextScalarLeaf`、spatial 多维复用 `animateTextVectorLeaf(...,false,...)`、非 spatial 多维复用
  `animateTextVectorLeaf(...,true,...)`，新 leaf 类型零额外 animate 代码。
- **gate 签名速查**（每 leaf 类型选作用面）：Opacity→全帧亮度 spread；Position→ink 垂直质心；
  Scale→ink 面积（像素数）；Rotation→单字 ink 包围盒长宽比；**Color→ink 像素均值 RGB（绿通道单调）**。
- structural op（Remove/Duplicate/Move 对 text animators）走 `mutate_property_structural.go`，
  **2026-06-16 双版本 ship-gate PASS**（`text_animator_struct_shipgate_test.go`：3 动画器
  Opacity/Skew/Fill Color × remove-middle/move-last-to-front/duplicate-first × AE2020+2025
  = 6/6，readback 动画器顺序逐项匹配；gotcha：ScriptingAPI 枚举完整 leaf schema，verify 按非默认
  值辨识 driven leaf）。详 `[[property-indexed-group-structural-re]]` 落地状态。

## Free-neighbor leaves 收口（2026-06-16）

用户点名「文字免费近邻收口」。RE 确认 7 个候选 leaf 全有效（`probe_text_anim_neighbors.jsx`，
AE 2020），vtype 完全落在既有机制：

| leaf | match-name | vtype | 复用 |
|---|---|---|---|
| Fill Opacity / Stroke Opacity / Stroke Width / Skew / Skew Axis / Rotation X / Rotation Y | 6417 1D scalar | `overwriteScalarCdat`（40B cdat f64@[0:8]，同 Opacity/Rotation） |
| Stroke Color | 6418 color | `overwriteVectorCdat`（96B cdat [A,R,G,B]×255，同 Fill Color） |

落地：一次 AE 调用批量 author 7 个 materialized .aep（`gen_text_anim_neighbor_templates.jsx`，
每 leaf 一个 fresh project，Range Start/End/Offset 全非默认 → 全 slot materialize），逐个
`extract_text_animator <src> <out>` 抽 `templates/text_animators_<key>_body.bin`。Go 侧抽两个私有
helper（`addTextScalarLeafAnimator` / `addTextColorLeafAnimator`），7 个 facade 各一行委托。

**gotcha**：加 Rotation X（或 Y）时 AE **auto-materialize 一个伴生 `ADBE Text Rotation`（Z=0，默认）**
——rotx/roty 模板比纯 scalar 大 242B（多一个 tdmn+tdb4 124B+cdat 40B）。`overwriteScalarCdat`
按 match-name 精确定位 `ADBE Text Rotation X`/`Y`，伴生 Z 留默认（无副作用）。

**5 个双版本渲染 gate PASS**（`text_animator_neighbor_shipgate_test.go`，表驱动 + 通用
`verify_text_animator_neighbor.jsx`，每 leaf 一个作用面签名，AE2020≡AE2025 逐数字一致）：
- Fill Opacity → 全帧亮度 spread（隐 0 → 显 199）
- Stroke Opacity → stroke ink 面积（隐 0 → 显 4362）
- Stroke Width → stroke ink 面积（粗 10797 → 细 1780）
- Stroke Color → ink 均值绿通道（红 0 → 白 243）
- Skew → 单字 ink bbox 长宽比（剪切宽 2.00 → 正立 0.65）

stroke 三连的 gate 在 verify jsx 把 base text 设 fill-off + 白 stroke（采样 fixture 关切，
非被测写），让 stroke ink 成唯一信号。

**Rotation X / Y = evidence-based defer（reachable-but-visually-inert）**：facade 写值正确、
AE 接受、值 round-trip 存活（`TestTextRotationXY_RoundTrip` Go 自验），但在**平面 2D 文字层
里视觉完全无效**——实测 bbox 三帧逐像素相同（高 160/160/160，宽 104/104/104，n 全 947）。
per-character 3D 旋转需先 enable「逐字 3D」（Per-character 3D，独立的 3D 能力，未支持）。
故 facade 标 Alpha/write-only，**不进渲染 gate**（红线4：不可像素门禁则不假绿）。同 Repeater
Order「可达但视觉惰性」家族的诚实收口。Skew Axis 同理（Skew=0 时独立不可门禁，未单出 facade）。

机制证：**1D scalar / color leaf 加 text animator 已是纯模板抽取 + 一行 facade**，Go 侧真·免费；
真正的工作量与门槛全在「每 leaf 设计一个可像素门禁的作用面签名」。

## Wiggly Selector（2026-06-16）+ Expressible Selector（defer）

RE（`probe_text_special_selectors.jsx`）：`ADBE Text Selectors` 接受两种特殊选择器（皆 NAMED 6213）：
- **Wiggly Selector**（`ADBE Text Wiggly Selector`，10 params：Mode/Wiggly Max·Min Amount/Range Type2/
  Temporal Freq/Character Correlation/Temporal·Spatial Phase/Wiggly Lock Dim/Random Seed）——选区随时间
  随机摆动（但 per-seed 确定）。
- **Expressible Selector**（`ADBE Text Expressible Selector`，2 params：Range Type2/Expressible Amount）——
  Amount 由表达式驱动。

**Wiggly 已 ship**：`AddTextWigglySelector(layer)` splice 一个全 elided 的 wiggly selector（AE 应用默认
Temporal Freq 2/Max 100/Min 0 → 自动摆动，无关键帧）进 Selectors 组。**双版本 render-gate PASS**
（`text_wiggly_selector_shipgate_test.go`，**时间变化签名**：opacity-0 + 空 range（End=0 选 0 字）+ wiggly →
wiggle 独驱选区，render 3 帧两两 frameDiff 大（0-1=1660·1-2=2247·0-2=1333，每帧 spread=199 有字）→ 证选区随
时间 re-select；AE2020≡AE2025 逐数字一致——确定性 wiggle）+ round-trip（wiggly tdmn 存活）。
gotcha：probe 里 `canSetExpression`/无效 match-name 抛 uncaught → AE 脚本错误 modal → ae_run exit 2（非
flake，是真 modal）；每 addProperty 独立 step()/去掉 canSetExpression 后绿。

**Expressible Selector = SHIPPED（2026-06-17，原 defer 已翻案）**：`AddTextExpressibleSelector(layer, amountExpr)`
双版本 render-gate PASS（`text_expressible_selector_shipgate_test.go`）。原 defer 理由「库表达式未渲染验证」**已过时**
——SetExpression 早经 S2 + 2026-06-15 Utf8-order 修复渲染验证（含 effect param，见 [[expression-enable-byte-pair]]）。

RE 关键发现（`re_text_expressible.jsx`）：
- **Expressible Amount 是表达式专属 param**：fresh 时 expressionEnabled=false、expression 空、且**读 `.value` 直接抛**
  （无静态值）。故空表达式 = inert selector，facade 强制 amountExpr 非空。
- **模板必须先 materialize**：默认 Expressible Selector 两 param（Range Type2 + Amount）**全 elided**，
  extract 出来只有 `tdsb+tdsn+GroupEnd`（108B）——没有 Amount tdbs 可写。解法：RE fixture 里先给 Amount 设个
  表达式（`amt.expression="selectorValue"`）逼 AE 持久化 Amount tdbs，再 extract（516B，含 cdat+Utf8+tdum/tduM，
  Utf8 已在 canonical 位）。`AddTextExpressibleSelector` splice 该模板 + `scalarTdbs`→`parseLeafProperty`→
  `SetExpression(amountExpr)`+`SetExpressionEnabled(true)` 覆写。
- **gate = 空间差分**（非时间）：两 comp 各 "ABCDEFGH" + opacity-0 animator + Expressible Selector，
  L=`textIndex<=4?100:0`（隐 ABCD → 渲 EFGH）、R=`textIndex>4?100:0`（隐 EFGH → 渲 ABCD），断言两者 ink 质心
  相距 >100px（可见字集不同 = 表达式真驱动选区）。坑：文字 center-justify，可见字集左右方向是排版假象，故用 |Δ| 不看符号。
  AE2020≡AE2025 逐数字一致（|Δ|=188，确定性）。**selector 家族至此全收口。**

## 多 Selector（2026-06-16）

`AddTextRangeSelector(layer, start, end, offset)` —— 给首个动画器加第 2+ 个 Range Selector。
RE 确认 `ADBE Text Selectors` 是 **INDEXED_GROUP（ptype 6214）**，干净接受多个 `ADBE Text Selector`
（`probe_text_multi_selector.jsx`：2 选择器各自 Start/End/Offset+Advanced）。机制 = splice 一个单
selector 模板（`templates/text_selector_body.bin`，extract 自任一 animator fixture）进 Selectors 组
Group End 前 + 覆写 Start/End/Offset——同 add-animator 的 indexed-group append vein。

**多选择器按各自 Mode 组合**（默认 Add=并集；Mode 经 `SetTextRangeAdvanced` 设——但注意它只作用首个
选择器，第 2 个的 Mode 暂走默认 Add）。**双版本 render-gate PASS**（`text_multi_selector_shipgate_test.go`，
A/B 差分单帧：两层皆 Opacity-0 + 首选择器 [0,50]——TXTA 仅 1 选择器 → 后半可见 spread=199 / TXTB 加第 2
选择器 [50,100] → 并集全隐 spread=0，AE2020≡AE2025）+ round-trip（selector count==2 存活）。

## Range Advanced（Selector 高级参数，2026-06-16）

`SetTextRangeAdvanced(layer, TextRangeAdvanced)` —— 一次设全 10 个 Range Selector「高级」子参数
（住 `ADBE Text Range Advanced` 子组，全 vtype 6417 1D scalar，枚举=1-based index）：

| match-name | 默认 | 含义 |
|---|---|---|
| ADBE Text Range Units / Type2 | 1 | Units（1=%/2=Index）/ Based On（1=Chars/2=ExclSpaces/3=Words/4=Lines）|
| ADBE Text Selector Mode | 1 | Mode（1=Add/2=Subtract/3=Intersect/4=Min/5=Max/6=Difference）|
| ADBE Text Selector Max Amount | 100 | **Amount %**（动画器作用强度）|
| ADBE Text Range Shape | 1 | Shape（1=Square/2=RampUp/3=RampDown/4=Triangle/5=Round/6=Smooth）|
| ADBE Text Selector Smoothness | 100 | Smoothness %（**仅 Shape=Square 生效**）|
| ADBE Text Levels Max/Min Ease | 0 | Ease High/Low |
| ADBE Text Randomize Order / Random Seed | 0 | 随机化 |

**机制**：Advanced 组在新建动画器上 elided（仅 tdsb/tdsn/GroupEnd）。setter 用一个 AE-native
**全 materialize** 模板（`templates/text_range_advanced_body.bin`）整组 replace elided Advanced 组的
children，再逐 cdat[0:8] 覆写为 caller 值（`DefaultTextRangeAdvanced()` 给默认基线）。

**RE gotcha（两条）**：
1. **enum set 使 live ExtendScript ref 失效**（同 shape addProperty stale-ref）——materialize 模板
   的 jsx 必须每个 set 自己 step()/重导航，否则跑到第 6 个炸。
2. **Smoothness 在 AE UI 里 Shape≠Square 时被隐藏**，`setValue` 抛「属性被隐藏」。但**Shape 非默认时
   AE 仍持久化 Smoothness 的 cdat slot**（值=默认 100）——故单 fixture（Shape=2）即拿到全 10 slot，
   无需双 fixture 合并（一度以为要 synthesis-merge，实测 fixture A 已含全 10）。

**Amount 双版本 render-gate PASS**（`text_range_advanced_shipgate_test.go`，A/B 差分单帧：两文字层仅
Amount 异，皆 Opacity-0 全选——TXTA Amount=100 → 隐（top spread=0）/ TXTB Amount=20 → 显（bottom
spread=159），AE2020≡AE2025）。**其余 9 参数 round-trip 验证**（全 cdat 存活，`TestTextRangeAdvanced_RoundTrip`），
**render gate evidence-defer**：Mode 需多 Selector 才有视觉、Smoothness 仅 Shape=Square、Units/BasedOn/
Randomize/Seed 改的是「选中哪些字」非简单亮度、Shape/Ease 是 falloff 轮廓 polish——均无干净独立像素签名，
机制（写+AE 接受+存活）已证。

## 相关
- [[property-indexed-group-structural-re.md]] — `ADBE Text Animators` 是 INDEXED_GROUP，
  结构性 op 共享机制；本案是其「从零创建 child」的补全
- [[text-btdk-length-variable-write-scoping.md]] — 文字「文档结构」侧（btdk），与本案（属性树）正交
- [[effect-param-elision-synthesis-lite.md]] — 同 default-elision + materialize-on-write 思路
- [[add-effect-splice-re.md]] — 同 (tdmn, payload) splice 进 indexed group 的 vein

## Cases
- 2026-06-15 首次：RE + AddTextOpacityAnimator + AnimateTextRangeOffset，双版本渲染 gate PASS
- 2026-06-15 扩 Position：AddTextPositionAnimator（spatial 3D 72B cdat + overwriteVectorCdat +
  spliceTextAnimator 泛化），slide-in 双版本渲染 gate PASS（ink 垂直质心迁移 Δ=260px）
- 2026-06-15 扩 Scale：AddTextScaleAnimator（3D 120B cdat，泛化机制零改动只抽模板+facade），
  shrink-in 双版本渲染 gate PASS（ink 面积 1070→313）
- 2026-06-15 扩 Rotation：AddTextRotationAnimator（1D scalar 同 Opacity，Go 免费；gate 签名 =
  单字 "L" ink 包围盒长宽比 1.52→0.65），spin-in 双版本渲染 gate PASS
- 2026-06-15 扩 Fill Color：AddTextColorAnimator（color 96B cdat=[A,R,G,B]×255 @[0:32]，复用
  overwriteVectorCdat；gate 签名 = ink 均值 RGB 绿通道 0→120→236），colour-wipe 双版本渲染 gate
  PASS（AE2020≡AE2025 逐像素一致）
- 2026-06-15 animate leaf 本身（1D scalar）：AnimateTextOpacity/AnimateTextRotation（`animateTextScalarLeaf`
  = scalarTdbs→parseLeafProperty→AnimateScalarKeyframes，零新关键帧代码）。gate 反转 wiring（Offset 静态、
  Rotation leaf 关键帧 0→90），aspect 0.65→1.52 双版本渲染 gate PASS。3D/4D leaf animate 留下一 slice
- 2026-06-15 animate 3D/4D leaf：AnimateTextPosition + AnimateTextColor（`animateTextVectorLeaf`→AnimateVectorKeyframes）。
  RE-first 抽 AE ground truth（animatedvec）逐字核对：Position(bpk128,value@0x38,marker3)/Color(bpk152,value@0x38,marker2)
  逐字匹配 effect spatial block，单测断言 bytes + 双版本 render gate PASS（Color R/B 248/0→0/248、Position 质心
  367→617）。Scale 3D 非 spatial 块（value@0x08）≠ AnimateVectorKeyframes → 暂搁
- 2026-06-15 收口 animate Scale 3D leaf（非 spatial）：`AnimateVectorKeyframesNonSpatial`（codec 早已支持非 spatial
  多维编码器，只是 AnimateVectorKeyframes 硬编码 spatial → 抽私有核 + 非 spatial 公开入口，零新字节代码）。
  `AnimateTextScale` 走它，单测断言 value@0x08/bpk128/0x38 留空，双版本 render gate PASS（ink 面积 2940→5879）。
  **animate-leaf 方向全收口**（1D+3D/4D spatial+非 spatial 3D 全 ship）
- 2026-06-16 Wiggly Selector：`AddTextWigglySelector`（splice elided wiggly selector，AE 默认参数自动摆动）。
  时间变化签名双版本 render-gate PASS（3 帧两两 frameDiff 大、确定性）+ round-trip。Expressible Selector
  evidence-defer（表达式驱动，库表达式支持未验证）。详上节
- 2026-06-16 多 Selector：`AddTextRangeSelector`（`ADBE Text Selectors` 是 INDEXED 6214，splice 单
  selector 模板 + 覆写 Start/End/Offset）。A/B 差分双版本 render-gate PASS（1 选择器后半可见 199 / 2 选择器
  并集全隐 0）+ round-trip（count==2）。详上节
- 2026-06-16 Range Advanced：`SetTextRangeAdvanced`（全 10 子参数一次设，全 materialize 模板 replace
  elided Advanced 组 + 逐 cdat 覆写）。Amount A/B 差分单帧双版本 render-gate PASS（top=0/bottom=159），
  其余 9 round-trip 验证 + render gate evidence-defer。gotcha：enum set 失效 live ref（每 set 重导航）；
  Smoothness Shape≠Square 隐藏但 slot 仍持久化（单 fixture 即全 10 slot）。详上节
- 2026-06-16 free-neighbor leaves 收口：5 个新 Add*Animator（Fill/Stroke Opacity·Stroke Width·Stroke
  Color·Skew）双版本渲染 gate PASS（表驱动 + 通用 verify jsx，每 leaf 一个作用面签名，AE2020≡AE2025）；
  Rotation X/Y evidence-based defer（2D 视觉惰性，bbox 三帧全同 → 需逐字 3D；facade 保留+round-trip 自验+标
  Alpha/write-only）。Go 侧抽 2 私有 helper，新 leaf = 抽模板 + 一行 facade。详上节
