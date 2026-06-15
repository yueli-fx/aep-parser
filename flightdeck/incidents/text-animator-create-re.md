---
status: active
when_to_read: implementing/extending text animators (AddTextOpacityAnimator / AnimateTextRangeOffset); adding a new animator property type (Position/Scale/Color); wondering where Text Animators live vs btdk; building a property *Property over a spliced chunk to reuse AnimateScalarKeyframes
applies_to: [text, text-animator, kinetic-typography, range-selector, ADBE Text Animators, ADBE Text Animator, synthesis-insert, indexed-group, AnimateScalarKeyframes, mutate-text-animator]
last_updated: 2026-06-15
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

## 现状 / 边界

- **已 ship**：Opacity、**Position 3D**、**Scale 3D**、**Rotation**、**Fill Color** 动画器、Range Selector（参数化
  Start/End/Offset）、offset 关键帧扫光（reveal / slide-in / shrink-in / spin-in / colour-wipe 通用）。各双版本渲染 gate PASS。
- **Alpha**：动画器叶子无 typed accessor（chunk-only，Reopen 后属性树重建但 animator 叶子
  不带 back-ref，同 `property-indexed-group-structural-re.md` 的 Root Vectors）。多动画器 append +
  Opacity/Position 混排已测。
- **按需扩**（同 vein，抽模板 + 一个 facade）：Range Advanced（Mode/Shape/Smoothness/基于…）；多 Selector；
  Wiggly/Expression Selector；**animate leaf 本身**（关键帧驱动 Position/Scale/Rotation/Color 值，非仅
  Range Offset）。同布局的免费近邻：Fill Opacity / Stroke Color / Stroke Width / Skew /
  Rotation X·Y（各 1D/color，照 overwriteScalar/VectorCdat 抽模板即可）。
- **gate 签名速查**（每 leaf 类型选作用面）：Opacity→全帧亮度 spread；Position→ink 垂直质心；
  Scale→ink 面积（像素数）；Rotation→单字 ink 包围盒长宽比；**Color→ink 像素均值 RGB（绿通道单调）**。
- structural op（Remove/Duplicate/Move 对 text animators）走 `mutate_property_structural.go`，
  仍仅 Go round-trip = Alpha 未单独 ship-gate。

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
