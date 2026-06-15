# Cockpit — aep-parser

**Last updated**: 2026-06-15 by claude（**effects 深化 arc 全收口**：用户定「expr 尽量全通 + effects 尽量全通 → 文字」。expr 机制已全 ship〔视为完成，剩 linear()/ease() 内容无关已证按需〕；effects 四缺口全 ship 双版本 gate——①跨效果 enum render gate〔Invert Channel〕②`AnimateEffectParamVec` 动画 color/point effect param〔spatial block RE，byte-structural〕③`SetEffectLayerParam`+Set Matte〔layer-reference 参数=目标层 ID 写 tdpi〕④效果库 31→41〔wave5 十个 MG distort/generate/stylize/transition〕。commits 0cad0b8/dff6b22/d7ad64a/96b9d10。详 incidents + git log。）

**Active focus**: **剩余能力 roadmap 主体已走完 + 表达式/效果深化 arc 全收口，库进入需求驱动稳态**（`specs/2026-06-14-remaining-capability-roadmap.md`）。优先级 1-5 全收口：**动画关键帧 ✅** · **3D ✅** · **形状剩余 ✅** · **mask ✅** · **表达式 ✅ 机制全 ship**。**effects 深化 arc ✅（2026-06-15）**：跨效果 enum render gate · `AnimateEffectParamVec`〔animated color/point〕· `SetEffectLayerParam`+Set Matte〔layer-reference〕· 效果库 31→41〔wave5〕——四项全双版本 gate（详 ## 下一步）。**唯一真剩余主线 = 文字多 run/段落（btdk splicing 未 RE，effort 高）**；其余皆「按需 / 不可达」（maskFeatherFalloff、gradient stroke 嵌套 Dashes/Taper/Wave、附录 A negative-finding）。机制库：parse-the-clone + synthesis-insert + animateGradientStops/spliceAnimatedPath/animated-vector-effect-param。每渲染/可见类双版本 ship-gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1`）。基本图形搁置。

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

**➡ 优先级（用户 2026-06-15 定）：表达式 + 效果「尽量全通」→ 文字。effects arc 已全收口 ✅（2026-06-15）**：
- **表达式 ✅ 视为完成**——机制全 ship（`SetExpression`/`SetExpressionEnabled` + Utf8 位置修 + 4 idiom + effect-param 表达式驱动），剩 `linear()/ease()/valueAtTime` 内容无关已证，**按需**（边际值低）。
- **〔1〕跨效果 enum render gate ✅**——`SetEffectParam` 泛型 enum 模板跨效果材化此前仅 GB 自身 gate；`TestMGEffectEnum_AEShipGate` 双版本（Invert Channel Red(2)/Green(3) 渲染各异、R 通道跨 128 中线、resave 存活）。零新代码。commit 0cad0b8。
- **〔2〕animated color/point effect param ✅**——`aep.AnimateEffectParamVec`：scalar 动画扩到 color(4D)/point(2D·3D)。RE 发现这三类用 **SPATIAL** keyframe block（value@0x38，bpk 152/104/128，per-type @0x08 marker color=2/point=3），tdb4 三 flag 翻转 type-agnostic。`TestAnimEffectVec_AEShipGate` 双版本（Fill Color 红→蓝 + Gradient Ramp Start 点 L→R radial swap）。commit dff6b22。详 `effect-param-elision-synthesis-lite.md` § Vec 扩展。
- **〔3〕Set Matte / layer-reference 参数 ✅**——`aep.SetEffectLayerParam` + `EffectSetMatte`。RE：layer-ref 存为**目标层 ID 写在参数自己的 tdpi**（同 host 绑定指向别层）；setter = length-preserving 4B tdpi 改写。`TestSetMatte_AEShipGate` 双版本（左红/右黑 matte gating）。关掉「reference-param effects」deferred。commit d7ad64a。详 `add-effect-splice-re.md` § Reference-param effects。
- **〔4〕效果库 31→41 ✅**——wave 5 十个 MG 效果（Turbulent Displace/Roughen Edges/Echo/Radial Blur/4-Color Gradient/Checkerboard/Grid/Stroke/Corner Pin/Venetian Blinds）。一次 fixture 提取 + `TestAddEffectWave4_AEShipGate` 双版本（全 10 add+读回+resave）。commit 96b9d10。

**effects 剩纯按需**：per-effect typed param helper · default-elided layer-ref 物化（Set Matte -0001 已随模板带出，故已可用）· Displacement Map/Compound Blur 等同 layer-ref 机制按需 · 库继续扩。

**➡ 下一阶段 = 文字（用户既定顺序 effects 之后）**：**文字多 run/段落 / 多 paragraph / 带 kerning / 空串改字**（btdk 段落·run entry splicing 未 RE，effort 高）= 完整文字动画接入。文字基础（NewTextLayer + 单段单 run SetText）已够用。

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
