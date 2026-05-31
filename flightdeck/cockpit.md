# Cockpit — aep-parser

**Last updated**: 2026-05-31 by claude（**Gradient fill write (SetGradient) landed on main**（子项⑭，承 Stroke Dashes 之后；用户选 "build it, gate empirically"）：`VectorGroup.AddGradientFill()` → `GradientFillNode`（`SetColorStops`/`SetAlphaStops` ≥2 stop + 范围校验、`Gradient()` getter），static 色标。**read 早已 ship**（`ParseGradientXML`+`Property.Gradient`），本条补 **write**。色标存 `ADBE Vector Grad Colors` 的 `LIST(GCst)→LIST(GCky)→Utf8` prop.map XML（v4；色标 6-float `[off,mid,r,g,b,1]`，alpha 3-float，Alpha Stops 在前+各带 Stops Size，尾 `Gradient Colors=1.0`）。`EncodeGradientXML` = `ParseGradientXML` 逆，round-trip 自洽。**length-variable 写几乎免费**：覆 Utf8 XML 改长 → `rifx.Chunk.Write` 自动 bottom-up 重算所有 LIST size（同 `Footage.SetPath`）；GCst 的 tdb4 非冗余长度头。embed 模板 `v2_2_shape_gradfill_body.bin`（仅含 Grad Colors；Grad Type/Start/End 在 fixture 为默认被 elide → 无 slot，AE 套默认线性 ramp，deferred）。**关键跨版本发现**：唯一带色标 fixture 是 AE 25.6-saved（AE 2020 拒开整个项目），但 **from-scratch AE25-shaped 渐变体 AE 2020 仍接受**（渐变格式 version-portable），一份 AE25 模板服务双版本 gate。**新 Gradient ship-gate 双版本 AE2020+2025 均 PASS**（`TestV2_2_GradientFill_AEShipGate_*`，3 色标 RGB re-save 解码校验）。Go round-trip 全绿，Stable API。详 `incident-reports/gradient-fill-write-re.md` + coverage 子项⑭。**deferred**：Grad Type/Start/End（elided）、动画色标、gradient **stroke**（G-Stroke，同 GCst 路径直接接力）。〔前一条：Stroke Dashes，子项⑬，commit d29fcd0〕详 logbook。

〔历史：**Stroke Dashes landed on main**（子项⑬，承 Taper/Wave 之后）：`StrokeNode.Dashes()` → `StrokeDashes`（`Enable/Disable`、`SetDash`/`SetGap` 自动 enable + 拒负、`Enabled/Dash/Gap` getter），单 Dash+Gap 对，static。Dash 1/Gap 1 均 OneD float64-BE @ cdat[0:8]，嵌套于 `ADBE Vector Stroke Dashes` group `LIST(tdgp)`（同 Taper/Wave，`findGroupBody` 下钻 + `overwriteShapeStreamCdat`）。**enable = 模板切换**：solid stroke 的 Dashes 组是空 placeholder，故引入第二嵌入模板 `v2_2_shape_stroke_dashed_body.bin`（携 Dash 1/Gap 1 slot，源 `v2_2_stroke_dashed.aep`）；solid `v2_2_shape_stroke_body.bin` **字节不变** → 现有 stroke/enum/taper-wave gate 零回归（`DisabledStaysSolid` round-trip 证 solid 路径 byte 一致，无需重跑）。hydrate 以 Dash/Gap leaf 存在与否回判 enabled。**新 Dashes ship-gate 双版本 AE2020+2025 均 PASS**（`TestV2_2_StrokeDashes_AEShipGate_*`，Dash=18/Gap=7 resave 解码）。Go round-trip 全绿，Stable API。详 `incident-reports/stroke-line-cap-join-miter-re.md` Dashes addendum + coverage 子项⑬。**deferred**：Dash 2/3·Gap 2/3（每对需独立模板变体）+ **Offset**（AE 端 hidden 且 `setValue` 拒，script-ungettable，无 slot 可建模——`v2_2_stroke_dashed.done` 实证）。〔前一条：Stroke Taper + Wave，子项⑫，commit ae23664〕详 logbook。〕

〔历史：**shape-node enum sweep**（commit 8c8f43d，承 Stroke Line Cap/Join/Miter 之后）：Rect/Ellipse `Direction` + Fill `BlendMode`/`CompositeOrder`/`FillRule` + Stroke `BlendMode`/`CompositeOrder` getter/setter（typed enum + 校验）。全 OneD float64-BE @ cdat[0:8]。rect/ellipse/fill/stroke 4 模板从单一 `v2_2_shape_all_full.aep` 重抽（含 enum slot）→ **重跑全部 shape ship-gate（9 gate × AE2020+2025）均 PASS，模板 swap 零回归**。Go round-trip 全绿，Stable API。RE 详 `incident-reports/stroke-line-cap-join-miter-re.md`（含 enum addendum）+ coverage 子项⑪。〔前一条：Stroke Line Cap/Join/Miter，commit 4beb7fd〕详 logbook。〕
**Active focus**: 无 active 实现线。

## Next session

1. **从下方长线 backlog 选下一条实现线**（用户定方向）。

> **长线 backlog**（大 arc 或缺 runtime setter）：

1. **Path keyframe**（逐帧 bezier）— V2.3+ 级大 arc。
2. **Layr Transform 3D 通道**（Orientation / Rotate X/Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。
3. ~~**Stroke Line Cap/Join/Miter**~~ — ✅ **已 ship**（2026-05-31, commit 4beb7fd, 见 Last-updated + logbook + coverage 子项⑩）。
4. ~~**Gradient W**（SetGradient）~~ — ✅ **已 ship**（2026-05-31, 子项⑭, gradient **fill** 色标 write；Grad Type/Start/End + 动画色标 + gradient **stroke** deferred。见 Last-updated + logbook + coverage 子项⑭ + `incident-reports/gradient-fill-write-re.md`）。
5. ~~**Stroke Dashes**~~ — ✅ **已 ship**（2026-05-31, 子项⑬, 单 Dash+Gap 对 + 模板切换；Dash 2/3·Gap 2/3·Offset deferred，见 Last-updated + logbook + coverage 子项⑬）。

**其它候选**：泛型 `DuplicateItem`（无 scripting API）、`ImportComposition`（需求驱动）。其余 deferred R-only（DisplayColorSpace / ValueText 等）见 logbook § Deferred。

## Hanging tasks

无。
