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

- **[必修·阻塞] path/keyframe 容量分页（lhd3）** —— 属性 >4 关键帧时 AE 2025 判损坏。`encodePathTimeTable` + `encodeKeyframes` 的 lhd3 header（@0x0C/@0x1C 非常量，是 page 容量字段）未做分页。**任何「上规模动画」前必修**。详 `lhd3-keyframe-capacity-pages.md`。
- **temporal ease 普及** —— 部分属性 keyframe 首版只 linear（shape path 子项⑮ deferred temporal ease）。补 ease in/out 字节（keyframe 两布局已有 `layoutFor` dispatcher）。
- **animated 矢量滤镜** —— Trim / Repeater 的参数动画（static 已 ship，animated 没做）。
- **animated gradient 色标** —— gradient stops 目前 static-only。
- **animated mask path** —— 见优先级 4（mask），与此层耦合。

## 优先级 2 — 3D 图层

**最大整块空白**。解锁真·拉镜（相机推轨 + 景深视差）、希区柯克变焦——AE 相机只对 3D 图层起作用，2D 图层相机动了也没用。camera/light option 已全做（2026-06-14），就缺「图层 3D 化 + Z 轴」这一环。

- **图层 3D flag（ldta）** —— RE ldta 里的 3D-enable bit（让一个 AV/shape 层变 3D）。模板起点：从一个 AE 存的 3D 层 fixture diff 出 flag 位。
- **Transform 3D 通道合成** —— Position_Z / Orientation / Rotate X / Rotate Y。当前 parser 能读但 tree 里是空 group（AE 默认 elide），写需 synthesis-insert（同 Light Color/Iris 法）。详 `transform-group-default-omission.md`（已确认是 AE default-omission，非 parser 截断）。
- **Material Options** —— 3D 层的 Casts Shadows/Accepts Lights/Ambient 等（`re_material_options.jsx` fixture 已存，未 ship）。
- **3D-render 像素 gate** —— 有 3D 层后，camera/light 的 DoF/bokeh/光照才能像素级深验（现仅 DOM 值）。
- 产物里程碑：一个纯模板编辑产出的「带景深视差的推拉镜」showcase。

## 优先级 3 — 形状图层剩余

shape 矢量滤镜家族主体已收齐；剩 elided 子流 + 未单独 gate 的模式。多数是 synthesis-insert 或 enum 补值的小活。

- **Stroke 嵌套组 Dashes / Taper / Wave**（实心描边）+ **Gradient stroke 同三组** —— 嵌套 group 写，模板带默认值。
- **Trim Type**（Simultaneously/Individually，elide 无 slot，需 synthesis-insert）。
- **Offset Paths 4 个 elided 子流**（Line Join / Miter / Copies / Copy Offset，现只模 headline Amount）。
- **Merge Add/Intersect/Exclude 模式** —— 已能写 enum 但未单独 gate。
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
