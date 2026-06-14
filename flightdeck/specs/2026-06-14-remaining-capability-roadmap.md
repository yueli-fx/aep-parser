---
status: active
summary: 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
last_updated: 2026-06-14
---

# 剩余能力 roadmap（模板起点）

> 2026-06-14 立项。把当前**没做 / 半做 / 不可达**的能力集中归档，按用户定的优先级排序，省得散落在各 incident / coverage 里忘掉。逐项做完即在此打勾 + 落对应 incident，做完一整层就 landing。

## 框架前提（重要，决定每项的难度）

- **不从零创建** —— 起点永远是**现有模板 / 真实 .aep**，在既有结构上扩/改。这绕开 from-scratch 那堆 silent-drop 坑（`multi-layer-silent-drop` / `ae2020-shape-ldta-164-corrupt`），工作性质 = 「parse-the-clone + 局部 splice / cdat overwrite」，难度比从零拼工程低一个量级。
- **复用已验证的两条机制**：① parse-the-clone（克隆模板跑读路径 parse → 点亮既有 setter）；② synthesis-insert（AE elide 的默认槽位 → splice AE-native leaf + 复位默认，组内**顺序敏感**、值须**可证明偏离默认**否则被 elide）。两者刚在 camera/light 验透（`camera-light-layer-create-re.md`）。
- **交付准则不变**：每项渲染/可见类能力须双版本 AE ship-gate（红线4 像素或至少 DOM-readback），值 round-trip 绿 ≠ 交付。
- **基本图形（Essential Graphics）搁置** —— 用户 2026-06-14 决定。剩 point/dropdown/text/Transform controller + Remove + **EG 面板崩溃未修 RE**（`essential-graphics-write-re.md`）全部押后，不在本 roadmap 优先级内。

---

## 优先级 1 — 动画关键帧

上规模动画的硬地基。当前 Layr Transform 全通道（Anchor/Scale/Rotation/Opacity）+ shape Size/Color/Position/path keyframe 已 ship，但有容量与覆盖缺口。

- ~~**[必修·阻塞] path/keyframe 容量分页（lhd3）**~~ ✅ **2026-06-14 全闭合**。`encodeKeyframes`（标量/矢量，早先随 S1 修）+ `encodePathTimeTable`（path 时间表，本次）的 lhd3 @0x0C/@0x1C 均 page 化（`pages=(n+3)/4`）。path gate 从 3kf bump 到 6kf 双版本 PASS。**遗留另一轴**：path 几何 lhd3 >4 **顶点**分页未测（顶点数轴 ≠ 关键帧数轴，需求驱动）。详 `lhd3-keyframe-capacity-pages.md`。
- ~~**temporal ease 普及**~~ ✅ **2026-06-14**。scalar/vector/color 早有 ease（`writeKeyframeBlock`）；唯一缺口 = shape **path**（`encodePathTimeTable` 恒写 linear）。修写 + 对称修 `hydratePathNode` 读回。Go round-trip + 双版本 AE gate（eased 关键帧 AE 读回 BEZIER）PASS。详 `path-keyframe-write-re.md`。注：influence 是分数 (0,1]。
- ~~**animated 矢量滤镜（Trim）**~~ ✅ **2026-06-14**。Trim End 关键帧 0→100 line-draw reveal，三帧渲染（空→右半→整圈）双版本 PASS（`TestMGTrimAnim_AEShipGate`）。代码路径早已通（`lowerShapeScalar`→`injectAnimatedStream`，与 Rect Roundness/Stroke Opacity 共享），缺的只是渲染面 gate。**Repeater/Offset Copies 走完全相同路径**（按需补 gate，边际价值低）。详 `trim-paths-vector-filter-re.md`。
- ~~**animated gradient 色标**~~ ✅ **2026-06-15**。用户手工在 AE GUI 造了带关键帧色标的 fixture（`v2_2_gradient_anim_src.aep`，JSX 无法 authoring 色标，唯此一途），RE 出 animated 布局 = 时间表(lhd3+ldat bpk=64)在 tdbs + 每帧一份 prop.map-XML Utf8 在 GCky（与 path 关键帧同构）。`AddGradientKeyframe` + `animateGradientStops`/`encodeGradientColorTimeTable`，双版本渲染 gate `TestGradientAnim_AEShipGate`（kf0=R/B/G → kf1=G/R/B 实渲交换，肉眼验）PASS。详 `gradient-fill-write-re.md` § animated color stops。**deferred**：gradient STROKE 色标动画 + 色标 ease（现仅 linear）。
- **animated mask path** —— 见优先级 4（mask），与此层耦合。

## 优先级 2 — 3D 图层

**最大整块空白**。解锁真·拉镜（相机推轨 + 景深视差）、希区柯克变焦——AE 相机只对 3D 图层起作用，2D 图层相机动了也没用。camera/light option 已全做（2026-06-14），就缺「图层 3D 化 + Z 轴」这一环。

- ~~**图层 3D flag（ldta）**~~ ✅ **2026-06-15**。3D-enable bit = ldta @0x26 **bit2**（早已读+写，`SetIs3D` length-preserving）。新发现：**翻 bit 一个动作就够**——AE 打开时自动 materialize 完整 3D 层（threeDLayer=true、Position 自动扩 2D→3D z=0、Orientation/RotateX/Y/Z + Material Options 全生成）。DOM-readback gate（`TestLayer3DEnable_AEShipGate`）双版本 PASS。详 `layer-3d-enable-bit-materializes.md`。
- ~~**可见 3D：相机推拉**~~ ✅ **2026-06-15**。从零 3D BOX(z=0) + `NewCameraLayer`，dolly 相机 Position Z（相机本就 3-comp 可设）→ box 随距离缩放（near z=-700→426px、far z=-2400→124px，3.41× 双版本逐字节一致）。**零新通道写代码**。渲染像素 gate `TestLayer3DCamDolly_AEShipGate` 双版本 PASS。详 `layer-3d-enable-bit-materializes.md`。
- ~~**Position Z 写入 = 视差**~~ ✅ **2026-06-15**。意外发现：从零 shape 层 Position **已是 3-comp 存盘**（`encode3D([x,y,0])`，Z 钉 0），故 reopen 后 `SetPosition([x,y,z])` length-preserving，**零新代码**。渲染 gate `TestLayer3DParallax_AEShipGate`：两同尺寸 3D box NEAR z=-800/FAR z=+1200 + Go 设相机 → NEAR 渲 300px、FAR 渲 99px（3.03× 双版本一致，肉眼验）。详 `layer-3d-enable-bit-materializes.md`。
- ~~**Rotate Y 透视 tumble**~~ ✅ **2026-06-15**（连带纠错：之前以为旋转/Orientation 通道不在从零树里需 synthesis——**错了**，`v2_2_transform_group_body.bin` 模板已含 Orientation/RotateX/Y/Z 全部 slot）。reopen 后 `SetRotateY` 覆写既有 cdat，length-preserving 零新代码。渲染 gate `TestLayer3DRotateY_AEShipGate`：RotateY=50°+近相机 → 梯形（左右边高 1.45× 双版本一致，肉眼验）。RotateX/Orientation/RotateZ 同路径按需补 gate。
- **小结**：整个 3D transform group（enable + Z 视差 + 旋转/朝向）**从零纯 Go 写入零新 serializer 代码**——`SetIs3D` + reopen 后既有 transform setter，因嵌入 transform 模板本就是完整 6-axis 3D schema。推拉镜/视差/透视全部渲染 gate 闭环。详 `layer-3d-enable-bit-materializes.md`。
- ~~**Material Options（光照 + 阴影）**~~ ✅ **2026-06-15**。**光照**零新代码（`SetLightKind(Point)` ldta @0x88 + AE 默认 Accepts Lights=ON）：`TestLayer3DLight_AEShipGate` 从零 POINT 光照 3D panel 出衰减梯度（near 185.6/far 88.0，2.11× 双版本）；`SetLightKind` 从零生效（4414=LightType.POINT）。**阴影**经 `aep.SetMaterialOption` synthesis-insert：从零 3D shape 的空 Material Options group splice `ADBE Casts Shadows`=On（synthesis-lite，同 SetEffectParam），`TestLayer3DShadow_AEShipGate` POINT 光投硬阴影到 wall（shadow 0 vs lit 114 双版本，AE 读回 materialCastsShadows=1）。catcher 默认接受阴影、light 自带 Casts Shadows 槽——均无需合成。其余系数（Diffuse/Specular…）同 `SetMaterialOption` 路径（Go round-trip 测，render-gate 按需）。详 `layer-3d-enable-bit-materializes.md`。
- ~~**3D-render DoF 像素 gate**~~ ✅ **2026-06-15**。相机景深 setter（DoF/Focus/Aperture/BlurLevel，之前仅 DOM）渲染验证：近层焦内锐（边带 1px）/ 远层失焦虚（边带 21-22px，>20× 双版本）。`TestLayer3DDoF_AEShipGate`。详 `layer-3d-enable-bit-materializes.md`。
- ~~产物里程碑：「景深视差推拉镜」showcase~~ ✅ **2026-06-15** `showcase/3d-camera`（视差 + 透视 tumble，用户真机验收 complete；景深另由 DoF gate 覆盖）。

## 优先级 3 — 形状图层剩余

shape 矢量滤镜家族主体已收齐；剩 elided 子流 + 未单独 gate 的模式。多数是 synthesis-insert 或 enum 补值的小活。

- **Stroke 嵌套组 Dashes / Taper / Wave**（实心描边）+ **Gradient stroke 同三组** —— 嵌套 group 写，模板带默认值。
- ~~**Trim Type**（Simultaneously/Individually）~~ ✅ **2026-06-15**（synthesis-insert 第 2 次，enum leaf；`TestMGTrimType_AEShipGate` 双版本 PASS，Individually=左满右空 render 签名；详 `trim-paths-vector-filter-re.md` § Trim Type）。
- **Offset Paths elided 子流**：~~**Copies**~~ ✅ **2026-06-15**（synthesis-insert 首次推广到矢量滤镜 body；`TestMGOffsetCopies_AEShipGate` 双版本 PASS，even-odd 同心环签名；详 `trim-paths-vector-filter-re.md` § Offset Copies）。剩 Line Join / Miter / Copy Offset（同 splice 路径，按需）。
- ~~**Merge Add/Intersect/Exclude 模式**~~ ✅ **2026-06-15**（Venn 判别床一帧 4 卡，`TestMGMergeModes_AEShipGate` 双版本 PASS；纯 gate 零新代码；详 `trim-paths-vector-filter-re.md` § Merge 全 4 模式）。
- **shape 次要子属性** —— Fill/Stroke Opacity·BlendMode·CompositeOrder、Shape Direction（部分 runtime-only，逐个甄别）。
- **PolyStar Polygon 型** —— 已于 2026-06-14 ship（独立模板）；此处仅留档确认无残余。

## 优先级 4 — mask

AddMask / RemoveMask + mode·color·inverted（mkif 字节）已 ship。剩属性与路径动画。

- **Mask Feather / Opacity 属性** —— mask 的羽化 + 不透明度（om-s 内属性，未做）。
- **SetMaskPath / 既有 mask 路径改写** —— 当前只能建 mask 不能改其 path；需 om-s 内 shap/shph 重写（与 shape path write 共享 `encodeBezier`）。
- **animated mask path** —— mask path keyframe（与优先级 1 的容量分页耦合，>4kf 需先修 lhd3）。
- **maskFeatherFalloff** —— 位置未 RE（可能落不可达，先探）。

## 优先级 5 — 表达式

**半残区，谨慎**。`SetExpression` 文本写 + enabled-byte（@0x77/@0x78）已 RE 修正（`expression-enable-byte-pair.md`），但仍有认知红线：Go round-trip 绿 ≠ AE 求值。

- **from-scratch / 模板表达式落地验证** —— 在模板层上挂表达式，双版本验 AE 真求值（不只读回文本）。
- **`linear()` / `ease()` remap** —— 边际值低（cockpit 已注），但属表达式常用；按需。
- **不做**：完整表达式引擎符号执行（`ReplaceSource` fixExpressions 已明确 out-of-scope）。

## 优先级 6 — 文字图层

NewTextLayer + 单段单 run length-variable SetText 已 ship（`text-btdk-length-variable-write-scoping.md`）。剩复杂文档结构。

- **多段落 / 多 run / 带 kerning / 空串改字** —— 当前 refuse，需 btdk 内段落·run entry splicing（未 RE）。
- **更多 TextDocument 字段** —— 按需扩（font/size/justification 等已部分有）。
- **变量字体写** —— ScriptingAPI no-op（`variable-fonts-write-noop.md`），落不可达。
- **ligature**（OT liga）—— 未 RE，可能不可达。

---

## 附录 A — 不可达 / ScriptingAPI 封锁（做不了或没法验，不在上面优先级内）

逐项有 negative-finding incident 背书，**不要重复尝试**：

- **DisplayColorSpace**（separate chunk 没 RE，旧 stub 已删）
- **ValueText / dropdown 选项 label**（需 schema DB，`valuetext-needs-schema-db.md`）
- **Camera FilmSize / SensorSize**（RO，write-blocked，`camera-filmsize-ldta-write-blocked.md`）
- **RQ comment**（无 scripting API 可读回验证，`rq-comment-no-scripting-api.md`）
- **linearizeWorkingSpace**（CMS/OCIO 联动，chunk 对但 API 读不到，`project-flag-chunks-lnrb-lnrp.md`）
- **dropframe / fontLocation / 变量字体**（runtime-only / 系统派生，`runtime-only-fields.md` + `variable-fonts-write-noop.md`）
- **environmentLayer 360° 素材 / ligature**（未 RE，疑不可达）

## 附录 B — 已搁置（用户决定，非技术阻塞）

- **Essential Graphics 进阶**：point/dropdown/text/Transform-sourced controller + RemoveEssentialProperty。
- **EG 面板崩溃未修 RE**：展开「基本图形」面板崩 AE，DOM readback 假绿，根源 ship-gate 从不开面板。详 `essential-graphics-write-re.md`。

## 附录 C — Render Queue（独立线，按需）

RQ R-only + 少量 W slice 已 ship；Set* slice-5~8 仍 Alpha。不在主优先级，需求驱动。
