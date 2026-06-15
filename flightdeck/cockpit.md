# Cockpit — aep-parser

**Last updated**: 2026-06-15 by claude（**优先级1 动画关键帧全收口 + 优先级4 mask 主体收口**。本轮收尾两个 priority-1 遗留：①**path 几何 >4 顶点**——建 n=5 五边形 mask+shape 实测 AE 2020/2025 **接受 + resave 逐字节相同**，incident 的 nextPow2 容量假设**证伪**（AE 实际 cap=n，我们 emit 本就 byte-faithful），纯 gate 零代码（`TestMGPentagonPath`）；②**gradient STROKE 色标动画**——`GradientStrokeNode.AddGradientKeyframe` 复用 fill 的 `animateGradientStops` + 同 `ADBE Vector Grad Colors` 流，`TestGradientStrokeAnim_AEShipGate` 双版本 PASS（stroke ramp R/B/G→G/R/B 翻转）。**优先级1 无遗留**。亦做了一轮 walkaround 体检（修 15 个 dangling-ref：landed/→archive/、charts/→references/）。前情同日：**animated mask path（`aep.SetMaskPathKeyframes`）**落地——既有 mask 轮廓改成 N≥2 关键帧动画（om-s time-table tdbs + 每帧 shap）。关键 RE（`dump_oms_tdb4` 实测）：**静态 mask tdb4（`maskShapeTdb4`）与 AE 自存 animated shape tdb4（`re_path_anim.aep`）逐字节同基底，仅差 @0x05/@0x44/@0x4f 三个 static→animated flag offset**——即 shape-path 动画机制（`spliceAnimatedPath`）逐字节迁移到 mask，无新字节语义。复用 `encodePathTimeTable` + 从 `makeMaskShapeOmS` 抽出的 `makeMaskShap`（单帧 mask-strictness shap，顶点数逐帧可变）。`TestMGMaskPathKf_AEShipGate` 双版本 PASS〔白 600×600 card，mask t=0 左半 → t=2s 右半，两关键帧 render reveal L↔R 翻转，静态 mask 不可能；numKeys==2 读回，resave 保留 2 path kf〕。新 `MaskPathKey` 输入类型 + docgen 补登 SetMaskPath（前漏）+ SetMaskPathKeyframes。**mask 路径写静态+动画全收口**，剩 maskFeatherFalloff（位置未 RE，可能不可达）。前情同日：**SetMaskPath** 既有 mask 路径改写落地——既有 mask 轮廓改成 N≥2 关键帧动画（om-s time-table tdbs + 每帧 shap）。关键 RE（`dump_oms_tdb4` 实测）：**静态 mask tdb4（`maskShapeTdb4`）与 AE 自存 animated shape tdb4（`re_path_anim.aep`）逐字节同基底，仅差 @0x05/@0x44/@0x4f 三个 static→animated flag offset**——即 shape-path 动画机制（`spliceAnimatedPath`）逐字节迁移到 mask，无新字节语义。复用 `encodePathTimeTable` + 从 `makeMaskShapeOmS` 抽出的 `makeMaskShap`（单帧 mask-strictness shap，顶点数逐帧可变）。`TestMGMaskPathKf_AEShipGate` 双版本 PASS〔白 600×600 card，mask t=0 左半 → t=2s 右半，两关键帧 render reveal L↔R 翻转，静态 mask 不可能；numKeys==2 读回，resave 保留 2 path kf〕。新 `MaskPathKey` 输入类型 + docgen 补登 SetMaskPath（前漏）+ SetMaskPathKeyframes。**mask 路径写静态+动画全收口**，剩 maskFeatherFalloff（位置未 RE，可能不可达）。前情同日：**SetMaskPath** 既有 mask 路径改写落地——`aep.SetMaskPath(layer, mask, path)` 复用 `makeMaskShapeOmS` 重建 om-s swap 进 atom，顶点数可变〔rect 4→triangle 3，走 AddMask 的 n≠4 lhd3 补丁不崩〕。`TestMGMaskPath_AEShipGate` 双版本 PASS〔三角内白 2/2 + 原 rect 角暗 3/3 + resave 顶点=3〕。前情同日：**Mask Opacity/Feather/Expansion** 经 synthesis-insert 第 3 落点〔mask atom，继 material leaves / shape filter 子流〕——`Mask.SetOpacity/SetFeather/SetExpansion`，parser 加 atomTdgp backref + AE-native leaf 模板〔mask 严格性须实抽〕。Opacity `TestMGMaskOpacity_AEShipGate` 双版本 PASS〔白 rect + 同尺寸 mask 100% vs 50% → lum 255/135；resave 经本库 parser 读回值存活〕。Feather/Expansion 同 wire + Go round-trip。前情同日：**优先级3 形状剩余主体收口**：**Fill/Stroke Opacity** 双版本 render-gate〔`TestMGOpacity_AEShipGate` 100% vs 50% → lum 255/135，white over 暗底〕；其余次要属性甄别归档〔Fill Rule 间接覆盖 · Blend Mode/Composite Order 写路径就绪 render-gate 按需 · Shape Direction 实心视觉无效不单独 gate〕。本日优先级3 累计：Offset Copies + Trim Type + Merge 全4模式 + Fill/Stroke Opacity 全双版本。前情同日：**Merge 全 4 模式 gated**——Venn 判别床（两部分重叠椭圆+Merge+Fill-on-top）一帧 4 卡覆盖 Add/Subtract/Intersect/Exclude，`TestMGMergeModes_AEShipGate` 双版本 PASS〔三区签名全互异：花生/左月牙/透镜/双月牙；纯 gate 零新代码〕。前情同日：**Trim Type**（Individually/Simultaneously）经 synthesis-insert 第 2 次落地——enum leaf〔4-child tdbs〕同 scalar-with-range 走通用 `spliceShapeLeafBeforeGroupEnd`。`TestMGTrimType_AEShipGate` 双版本 PASS〔Individually=路径按序 trim→左满右空，与「Simultaneously=各半弧」相反，render 实锤纠直觉；left-ring 4/4 + right-empty 4/4 + resave Type=2，png 20470b 两版一致〕。前情同日：**Offset Copies** 经 synthesis-insert 落地——把 `SetMaterialOption`（material leaves）那套**首次推广到矢量滤镜 body**：默认 Amount-only body 不变，仅当 `SetCopies≠1` 时 splice `ADBE Vector Offset Copies` leaf〔`spliceShapeLeafBeforeGroupEnd` + 覆写 cdat〕。`TestMGOffsetCopies_AEShipGate` 双版本 PASS〔even-odd 同心环签名：外白环 d≈300 + gap 暗 d≈260 + 白心，单副本不可能；png 10777b 两版一致〕。蓝本可推 Trim Type / Offset 其余子流。详 `trim-paths-vector-filter-re.md` § Offset Copies。前情：优先级2（3D）全部收官。逐 commit 见 git log。）

**Active focus**: **剩余能力 roadmap 主体已走完，库进入需求驱动稳态**（`specs/2026-06-14-remaining-capability-roadmap.md`）。优先级 1-5 全收口：**动画关键帧 ✅ 全收口（2026-06-15）** · **3D ✅ 全收官** · **形状剩余 ✅ 主体收口** · **mask ✅ 主体收口** · **表达式 ✅ 机制全 ship**（`SetExpression`/`SetExpressionEnabled` + 4 idiom 语汇 gate；剩 linear()/ease() 显式按需）。**唯一真剩余 = 优先级6 文字多 run/段落（btdk splicing 未 RE，effort 高，等需求驱动）**；其余皆「按需 / 不可达」（maskFeatherFalloff、gradient stroke 嵌套 Dashes/Taper/Wave、附录 A negative-finding）。机制库：parse-the-clone + synthesis-insert + animateGradientStops/spliceAnimatedPath。每渲染/可见类双版本 ship-gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1`）。基本图形搁置。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**主线 = `specs/2026-06-14-remaining-capability-roadmap.md`**（模板起点，非 from-scratch）。优先级顺序（用户 2026-06-14 定）：**动画关键帧 > 3D 图层 > 形状剩余 > mask > 表达式 > 文字图层**。

**优先级1 动画关键帧 ✅ 全收口（2026-06-15）**（lhd3 容量分页 / temporal ease / animated Trim / animated gradient 色标 全双版本，2026-06-14~15）。**path 几何 >4 顶点分页 ✅**（n=5 五边形 mask+shape `TestMGPentagonPath` 双版本 PASS + AE resave 逐字节相同——nextPow2 假设证伪，AE 用 cap=n，零代码改）。**gradient STROKE 色标动画 ✅**（`GradientStrokeNode.AddGradientKeyframe`，复用 fill 的 `animateGradientStops` + 同 `ADBE Vector Grad Colors` 流；`TestGradientStrokeAnim_AEShipGate` 双版本 PASS，stroke ramp R/B/G→G/R/B 翻转）。**无遗留**。

**优先级2 3D 图层 ✅ 全部收官（2026-06-15）**：enable / 相机推拉 / Z 视差 / RotateY 透视 / 相机 DoF / **光照** / **阴影** 七项全渲染 gate 双版本 PASS；showcase `3d-camera` 用户真机验收 complete。整个 3D transform group 从零零新 serializer 代码（transform 模板本就是 6-axis 3D schema）；光照零新代码；阴影靠新 `aep.SetMaterialOption` synthesis-insert（material leaves）。RotateX/Orientation/RotateZ 同路径按需补 gate（低优先）。详 `incidents/layer-3d-enable-bit-materializes.md`。

**➡ 优先级3 形状图层剩余〔进行中〕**（shape 矢量滤镜主体已收齐；剩 elided 子流 + 未单独 gate 的模式，多为 synthesis-insert / enum 补值小活）：
- ~~**Offset Copies**~~ ✅ **2026-06-15**（synthesis-insert 首发矢量滤镜；scalar-with-range leaf）。剩 Offset Line Join / Miter / Copy Offset 同 splice 路径，按需。
- ~~**Trim Type**（Simultaneously/Individually）~~ ✅ **2026-06-15**（synthesis-insert 第 2 次，enum leaf；Individually=左满右空 render 签名）。
- ~~**Merge Add/Intersect/Exclude 模式**~~ ✅ **2026-06-15**（Venn 判别床一帧 4 卡双版本 gate；纯 gate 零新代码）。
- ~~**shape 次要子属性**~~ ✅ **2026-06-15**（Fill/Stroke Opacity 双版本 render-gate；Fill Rule 间接覆盖；Blend Mode/Composite Order 写就绪 render-gate 按需；Shape Direction 实心视觉无效不单独 gate。详 `stroke-line-cap-join-miter-re.md` § secondary sub-properties）。

**优先级3 主体收口**。剩**纯按需项**（非阻塞）：Gradient stroke 嵌套组 Dashes/Taper/Wave · Offset Line Join/Miter/Copy Offset · Blend Mode/Composite Order render-gate · ZigZag Points/Twist Center 等 elided 子流（synthesis-insert 蓝本现成）。

**优先级4 mask ✅ 主体收口（2026-06-15）**：结构性 op 全收口（Add/Remove/Duplicate/Move）· mode/color/inverted · **Feather/Opacity/Expansion ✅** · **SetMaskPath ✅** · **animated mask path（SetMaskPathKeyframes）✅**（om-s 多 shap + tdbs 时间表；RE 实测静态 mask tdb4 == animated shape tdb4 同基底，仅差 3 flag offset，shape-path 动画机制逐字节迁移；左半→右半 reveal L↔R 翻转双帧 render gate 双版本 PASS）。剩**纯按需/可能不可达**：**maskFeatherFalloff**（位置未 RE，先探可能不可达）。详 `incidents/add-mask-create-re.md` § SetMaskPathKeyframes。
- **Gradient stroke 嵌套组 Dashes/Taper/Wave**（实心描边的三组已 ship 子项⑫⑬；gradient stroke 缺）。

**优先级5 表达式 ✅ 机制全 ship**（核实代码：`SetExpression`/`SetExpressionEnabled` tdb4 @0x77/@0x78 已修 + `expression_shipgate` + `expr_vocab_shipgate` 4 idiom 双版本——看板旧措辞「SetExpression 栽」已滞后）。剩 `linear()`/`ease()` remap 显式按需（机制已证内容无关，边际值低）。详 `incidents/expression-enable-byte-pair.md`。

**➡ 新优先级（用户 2026-06-15 改）：表达式 + 效果 > 文字**。文字基础（NewTextLayer + 单段单 run SetText）够用，**完整文字动画接入留到下一阶段**。当前挖 expr+effects 深化的真缺口（实测确认）：
- **〔A〕expression 驱动 effect param ✅（2026-06-15）**——本以为「字节已通」是**假绿**：`SetExpression` append Utf8 到末尾，对带 tdum/tduM 的 effect param 落在 tduM 后 → AE 2020 判损坏跳层。**真 bug**：Utf8 必须插 cdat 后 / tdum-tduM 前（AE-native dump 确认）。修 `back_property.go::SetExpression`。`TestExprEffect_AEShipGate` 双版本 PASS（Gaussian Blur Blurriness=`time*40`，AE 求值 valueAtTime(2.5)=100、blur 增长 lum 0→62）。**把表达式从「只能挂无 tdum/tduM 属性」扩到任意属性含 effect param**。详 `expression-enable-byte-pair.md` § Utf8 位置二次纠错。
- **➡〔B 大头·下一步〕animated effect param（from-scratch keyframe 合成）**——实测 `InsertKeyframe` 对 static effect param 报「insert from scratch not supported」（`propertyBack` 只有 cdat 无 ldat/lhd3）。需把 static cdat 转 animated keyframe 容器（机制类比 shape 的 `injectAnimatedStream`+`encodeKeyframes`，但走 parsed Property 路径）。**这是 effects 最大真缺口**（动画 blur/slider 驱动）。
- 表达式 `linear()/ease()/valueAtTime`：低边际（内容无关已证），按需。
- 效果库扩充（>30）/ per-effect typed helper / reference-param effects（Set Matte 等需 tdpi remap）：按需。

**文字多 run/段落**（btdk splicing 未 RE）= 留到下一阶段完整文字动画。

**蓝本（synthesis-insert 推广到矢量滤镜，2026-06-15 验透 2 类 leaf）**：clone elided leaf 模板 → `spliceShapeLeafBeforeGroupEnd`（GroupEnd 前插 (tdmn,tdbs) pair）→ 覆写 cdat，仅当值≠默认。已验 scalar-with-range（Offset Copies 6-child）+ enum（Trim Type 4-child）两类。不污染默认 body、不需 hydration（filter 靠 opaque chunk 穿越 Reopen）。剩 ZigZag Points / Twist Center / Repeater Order / Offset Line Join·Miter·Copy Offset 同路径按需。详 `trim-paths-vector-filter-re.md` § Offset Copies / Trim Type。

完整清单（每层细项 + 优先级4-6 mask/表达式/文字 + 不可达附录）见 roadmap spec。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
