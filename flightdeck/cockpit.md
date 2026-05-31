# Cockpit — aep-parser

**Last updated**: 2026-05-31 by claude（**Gradient fill write (SetGradient) landed on main**（子项⑭，承 Stroke Dashes 之后；用户选 "build it, gate empirically"）：`VectorGroup.AddGradientFill()` → `GradientFillNode`（`SetColorStops`/`SetAlphaStops` ≥2 stop + 范围校验、`Gradient()` getter），static 色标。**read 早已 ship**（`ParseGradientXML`+`Property.Gradient`），本条补 **write**。色标存 `ADBE Vector Grad Colors` 的 `LIST(GCst)→LIST(GCky)→Utf8` prop.map XML（v4；色标 6-float `[off,mid,r,g,b,1]`，alpha 3-float，Alpha Stops 在前+各带 Stops Size，尾 `Gradient Colors=1.0`）。`EncodeGradientXML` = `ParseGradientXML` 逆，round-trip 自洽。**length-variable 写几乎免费**：覆 Utf8 XML 改长 → `rifx.Chunk.Write` 自动 bottom-up 重算所有 LIST size（同 `Footage.SetPath`）；GCst 的 tdb4 非冗余长度头。embed 模板 `v2_2_shape_gradfill_body.bin`（仅含 Grad Colors；Grad Type/Start/End 在 fixture 为默认被 elide → 无 slot，AE 套默认线性 ramp，deferred）。**关键跨版本发现**：唯一带色标 fixture 是 AE 25.6-saved（AE 2020 拒开整个项目），但 **from-scratch AE25-shaped 渐变体 AE 2020 仍接受**（渐变格式 version-portable），一份 AE25 模板服务双版本 gate。**新 Gradient ship-gate 双版本 AE2020+2025 均 PASS**（`TestV2_2_GradientFill_AEShipGate_*`，3 色标 RGB re-save 解码校验）。Go round-trip 全绿，Stable API。详 `incident-reports/gradient-fill-write-re.md` + coverage 子项⑭。**deferred**：Grad Type/Start/End（elided）、动画色标、gradient **stroke**（G-Stroke，同 GCst 路径直接接力）。〔前一条：Stroke Dashes，子项⑬，commit d29fcd0〕详 logbook。

〔历史：**Stroke Dashes landed on main**（子项⑬，承 Taper/Wave 之后）：`StrokeNode.Dashes()` → `StrokeDashes`（`Enable/Disable`、`SetDash`/`SetGap` 自动 enable + 拒负、`Enabled/Dash/Gap` getter），单 Dash+Gap 对，static。Dash 1/Gap 1 均 OneD float64-BE @ cdat[0:8]，嵌套于 `ADBE Vector Stroke Dashes` group `LIST(tdgp)`（同 Taper/Wave，`findGroupBody` 下钻 + `overwriteShapeStreamCdat`）。**enable = 模板切换**：solid stroke 的 Dashes 组是空 placeholder，故引入第二嵌入模板 `v2_2_shape_stroke_dashed_body.bin`（携 Dash 1/Gap 1 slot，源 `v2_2_stroke_dashed.aep`）；solid `v2_2_shape_stroke_body.bin` **字节不变** → 现有 stroke/enum/taper-wave gate 零回归（`DisabledStaysSolid` round-trip 证 solid 路径 byte 一致，无需重跑）。hydrate 以 Dash/Gap leaf 存在与否回判 enabled。**新 Dashes ship-gate 双版本 AE2020+2025 均 PASS**（`TestV2_2_StrokeDashes_AEShipGate_*`，Dash=18/Gap=7 resave 解码）。Go round-trip 全绿，Stable API。详 `incident-reports/stroke-line-cap-join-miter-re.md` Dashes addendum + coverage 子项⑬。**deferred**：Dash 2/3·Gap 2/3（每对需独立模板变体）+ **Offset**（AE 端 hidden 且 `setValue` 拒，script-ungettable，无 slot 可建模——`v2_2_stroke_dashed.done` 实证）。〔前一条：Stroke Taper + Wave，子项⑫，commit ae23664〕详 logbook。〕

〔历史：**shape-node enum sweep**（commit 8c8f43d，承 Stroke Line Cap/Join/Miter 之后）：Rect/Ellipse `Direction` + Fill `BlendMode`/`CompositeOrder`/`FillRule` + Stroke `BlendMode`/`CompositeOrder` getter/setter（typed enum + 校验）。全 OneD float64-BE @ cdat[0:8]。rect/ellipse/fill/stroke 4 模板从单一 `v2_2_shape_all_full.aep` 重抽（含 enum slot）→ **重跑全部 shape ship-gate（9 gate × AE2020+2025）均 PASS，模板 swap 零回归**。Go round-trip 全绿，Stable API。RE 详 `incident-reports/stroke-line-cap-join-miter-re.md`（含 enum addendum）+ coverage 子项⑪。〔前一条：Stroke Line Cap/Join/Miter，commit 4beb7fd〕详 logbook。〕
**Active focus**: **Path Keyframe arc（in-flight，Phase 0 待做）**。逐帧 bezier path 动画持久化。范围已与用户锁定：①首版 **linear only**（不做 temporal ease，结构稳定后再补）；②**只做 shape path**，不顺带 mask path write；③**Phase 0 RE 先行**。成功标准：`Path().AddKeyframe*` 数据完整写出 + hydrate 完整读回 + 多帧时间/几何正确 + AE2020+2025 双版本 gate 过 + 仅 linear。

## Next session

1. **Path Keyframe Phase 0 — RE（直接开干，任务 #4）**。写 `test_data/re_path_anim.jsx`：AE 给一个 shape layer 的 path 属性（`ADBE Vector Shape - Group` 下的 `ADBE Vector Shape`）加 2-3 个 linear 关键帧（`pathProp.setValueAtTime(t, new Shape())`，顶点数/位置随帧变），跑 AE 2025 存盘 → `go run ./tmp_debug/dump_body` 或新 dump 工具确认结构。**要确认的 4 点**：(a) `om-s → omks` 下是否**每帧一份 shap 几何**（== mask 多帧同构）；(b) `om-s.tdbs.kfl` 的 time/interp 表布局（64B block，time@0x00 = 1/8000s ticks）；(c) 每帧 `shph` 的 bbox 是否独立归一化；(d) `lhd3`(52B) 多帧/多顶点下哪些字段变（当前单帧是 verbatim 复制的魔数）。结论决定 Phase 2 lower 走 from-scratch 还是扩 embed 模板。

> **关键复用点（已探明，省下个对话重新摸索）**：
> - Runtime 模型 `PathNode.path = PropertyStream[BezierPath]` 已支持关键帧（`AddKeyframeLinear/WithEase` 现成）→ **API 表面几乎零新增**。
> - **磁盘结构已知**：mask path 动画在 `parse_mask.go`（`decodeMask` + `readMaskPathTimes`）已完整 RE+parse；shape path 同构（`parse_shape.go` 注释明说）。Phase 1 hydrate 直接借 mask 那套读多 shap + tdbs.kfl times。
> - **first-kf fallback 位置**：`lower_shape_node.go:454-456`（lowerPathNode）+ `lower_property_stream.go:102-116`（LowerPathStream，无调用方）。shape path 实走 `lowerPathNode`（embed 模板 splice）。
> - **复用边界**：time/ease 表 = spatial 标量 ease 同布局（可复用）；几何值 = 每帧一个 shap（path 专属，要新写多 shap emit）。
> - **最大风险**：从零构造多帧 om-s（tdbs.kfl + N shap）能否被 AE 接受——shape path 从零构造曾**崩 AE 2020(0::42)** 故现状走 embed 单帧模板；embed 模板撑不了多帧（无 tdbs.kfl、单 shap）。详设计提案见本对话历史 / logbook。
> - 已建任务 #4–#7（Phase 0–3）。

> **长线 backlog**（大 arc 或缺 runtime setter）：

1. ~~**Path keyframe**~~ — 🛫 **in-flight**（见上方 Active focus + Next session；Phase 0 RE 待做，任务 #4–#7）。
2. **Layr Transform 3D 通道**（Orientation / Rotate X/Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。
3. ~~**Stroke Line Cap/Join/Miter**~~ — ✅ **已 ship**（2026-05-31, commit 4beb7fd, 见 Last-updated + logbook + coverage 子项⑩）。
4. ~~**Gradient W**（SetGradient）~~ — ✅ **已 ship**（2026-05-31, 子项⑭, gradient **fill** 色标 write；Grad Type/Start/End + 动画色标 + gradient **stroke** deferred。见 Last-updated + logbook + coverage 子项⑭ + `incident-reports/gradient-fill-write-re.md`）。
5. ~~**Stroke Dashes**~~ — ✅ **已 ship**（2026-05-31, 子项⑬, 单 Dash+Gap 对 + 模板切换；Dash 2/3·Gap 2/3·Offset deferred，见 Last-updated + logbook + coverage 子项⑬）。

**其它候选**：泛型 `DuplicateItem`（无 scripting API）、`ImportComposition`（需求驱动）。其余 deferred R-only（DisplayColorSpace / ValueText 等）见 logbook § Deferred。

## Hanging tasks

无。
