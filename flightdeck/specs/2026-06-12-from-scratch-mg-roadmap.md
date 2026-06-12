---
status: active
summary: 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
last_updated: 2026-06-12
---

# From-scratch MG 工程能力 roadmap

## 目标（用户 2026-06-12 点名）

**不用 AE 就能创建完整 .aep 工程**——AI 直接生成 MG（motion graphics）动画。即把库的强项从「读真实 .aep + 局部改」扩展到「从零拼完整工程」，每一步按交付准则（CLAUDE.md #7）确权：**渲染像素级 + AE 2020/2025 双版本 gate**，演示禁混未验证能力。

基线（orbit demo 已确权）：NewComposition + NewShapeLayer + Rect/Ellipse/Fill/Stroke/GradientFill + 静态值 + 2 关键帧线性旋转 + MoveToEnd 层序 → 渲染像素全对。

## Slice 排序（价值 × MG 刚需 × 风险）

### S1 — ease 关键帧 + 关键帧规模确权 ✅ DONE 2026-06-12
- `AddKeyframeWithEase` API 已存在但从未 AE-gated；gate 过程抓出**两个真 bug**：
  1. `writeKeyframeBlock` interp 字节硬编码 Linear → ease 表被 AE 忽略（修：per-side 非零 ease → Bezier；codec 校验 influence ∈ (0,1] 分数）。
  2. lhd3 @0x0C/@0x1C 不是常量而是 **4-keyframe 分页容量**——恒写 1/4 致 >4 kf 被 AE 2025 判损坏（AE 2020 宽松）。「2 关键帧边界」是真硬边界。详 `incidents/lhd3-keyframe-capacity-pages.md`。
- ✅ `TestMGEase_AEShipGate_*` AE 2020+2025 双版本渲染像素 PASS：LIN 中点 x=950 / EAS 慢出滞后 x=411 / SCL 6kf 插值 x=950，DOM 读回 influence 90/10 + BEZIER。
- 遗留：ldat 块 @0x10 segment 长度缓存未复刻（AE 重算，双版本接受）；`encodePathTimeTable` 容量字段同病未修（path >4kf 前必修）。

### S2 — 表达式激活 RE ✅ DONE 2026-06-12
- 翻案：历史「@0x78 反语义 disabled 位」解读是错的——真相 = **@0x77 disabled 位 + @0x78 has-expression 标记**两字节对（@0x78 不同步时 AE 直接丢表达式文本）。修 SetExpression（同步 @0x78）+ SetExpressionEnabled（写 @0x77）+ parse。详 `incidents/expression-enable-byte-pair.md`。
- ✅ `TestExpression_AEShipGate_*` AE 2020+2025 双版本渲染像素 PASS：ON 层 `time*90` 实际求值（rotation@2s=180、dot 渲染转到锚点下方）、OFF 层同表达式保文本不求值、resave 双态存活。
- followup（开 S6 端到端前做）：常用 MG 表达式语汇 gate（loopOut / wiggle / thisComp.layer 引用链）——单表达式 `time*90` 已确权，语汇覆盖未验。

### S3 — Trim Paths（线描动画，MG 标配）✅ DONE 2026-06-12
- Trim Paths（`ADBE Vector Filter - Trim`）= Vectors Group 内与 shape/fill/stroke 平级的矢量滤镜节点。套既有 embed-body vein：`extract_shape_bodies` 抽 `v2_2_shape_trim_body.bin` → `lowerTrimNode` clone + `lowerShapeScalar` 覆写 Start/End/Offset（f64 BE @cdat[0:8]；Start/End 原始%、Offset 度数）。Trim Type 默认 elide 未建模。`AddTrim`/`TrimNode`（static→cdat 覆写、animated→`injectAnimatedStream` flip 关键帧容器，line-draw reveal 天然支持）。
- ✅ `TestMGTrim_AEShipGate_*` AE 2020+2025 双版本**渲染像素** gate PASS：FULL(End100) 整圈 L+R、HALF(End50) 右半弧 top/right/bottom 在·left 不在；trim End resave 读回 50/100。两个 RE ground truth（render 验证非空想）：add-order [Ellipse,Stroke,Trim] trim 剪 stroke；AE 椭圆 path 起点顶部 12 点顺时针。详 `incidents/trim-paths-vector-filter-re.md`。
- 遗留：animated trim（line-draw 真动画 End 0→100 keyframe）路径已通但未单独 gate；Repeater/Merge/Offset/Round/ZigZag 同类矢量滤镜复用此 vein（蓝本见 incident）。

### S4 — precomp 嵌套（工程结构刚需）✅ DONE 2026-06-12
- `aep.NewPrecompLayer(parent, child, name)`：precomp 层 = 普通 AV 层，唯一标识 = ldta @0x28 SourceID 指向 CompItem（源 comp 已存在，无 footage item 要造）。复用 camera/light 的 `newTemplatedLayer`（embed `layer_precomp_body.bin` 单 Layr）+ clone 后 `SetSource(child.ID)` + cycle guard。修 `newTemplatedLayer` backref 缺 ldta（SetSource 静默失败靠 ID 巧合，字节输出无变）。
- ✅ `TestMGPrecomp_AEShipGate_*` AE 2020+2025 双版本**渲染像素** gate PASS：parent 唯一层=child precomp，AE 读回 source=CompItem、child 绿方块+蓝 BG 透出、resave SourceComposition 解到 child。**防 ID 巧合**：parent 先建 child 后建（child.ID≠模板 stale 1）。详 `incidents/precomp-layer-source-id-re.md`。
- 遗留：anchor/scale 未参数化（child≠1920×1080 需手设 transform）· collapse transformation / time-remap 未做 · 多层嵌套（祖孙）未单独 gate（cycle guard 已覆逻辑）。

### S5 — 零散质感件（按需）
- **Repeater（`ADBE Vector Filter - Repeater`）✅ DONE 2026-06-12**：径向/网格复制，MG 高频。套 S3 trim 矢量滤镜 vein（顶层 Copies/Offset + 嵌套 Transform 子组 findGroupBody descend）。`AddRepeater`/`RepeaterNode`。✅ `TestMGRepeater_AEShipGate_*` 双版本渲染像素 PASS（Copies=5 + Position[300,0] → 5 点成行、间隙暗、resave 读回）。坑：match-name `ADBE Vector Repeater Anchor`（非 Anchor Point）。详 `incidents/trim-paths-vector-filter-re.md` § 复用确认。
- Gradient Start/End Pt：现模板 elided → 只有默认水平 ramp，方向不可控（orbit BG 已暴露）。**未做**。
- Rounded Corners / Merge Paths / Offset Paths / ZigZag（同矢量滤镜 vein，蓝本已三次验证）/ 文本动画器（大坑，单列）。**未做**。

## 不做 / 边界

- 文本动画器（Text Animators）、3D、粒子类不入本 roadmap（btdk/3D 各有暂搁前置）。
- 每 slice 独立 plan + 独立 ship-gate；roadmap 不预支「组合 = 各 slice 之和」——最终「AI 生成完整 MG 工程」需要一个端到端组合 gate（orbit demo 的升级版，含 ease+trim+precomp 全要素）作为收口 slice S6。

### S6 — 端到端组合 gate（收口）✅ DONE 2026-06-12 — roadmap 主线闭环
- 一个纯 Go 工程 `MGX_Scene` 组合全部已 ship slice：precomp(S4 的 `MGX_Badge`=ellipse+stroke+Trim End=50 半弧·S3) + ease 关键帧动画 MOVER(S1) + Repeater 4-copy DOTS(S5) + BG。
- ✅ `TestMGCombo_AEShipGate_AE2020/2025` 双版本**渲染像素** PASS（mid-frame t=2s）：MOVER ease 滞后 x=411（线性中点 950）· BADGE precomp 透出 trim 右半弧（左切）· DOTS 4 点+间隙暗 · AE 读回 4 层/BADGE source=CompItem/MOVER 2 eased kf influence 90/DOTS Copies=4 · resave 全工程存活。
- **意义**：坐实交付准则的「组合/端到端是独立交付项」——单点 gate 不为「从零拼完整工程」背书，本 gate 证明四要素在一个 AE 接受的工程里协同渲染正确。**用户终极目标（不开 AE 纯 Go 生成完整 MG 动画）端到端达成**。`mg_combo_shipgate_test.go`（无新写路径，纯组合）。

## 评审纪要

（空）
