# Cockpit — aep-parser

**Last updated**: 2026-05-31 by claude（**Stroke Taper + Wave landed on main**（子项⑫，承 enum sweep 之后）：`StrokeNode.Taper()/Wave()` → `StrokeTaper`(Start/End Length·Width·Ease 6 标量) / `StrokeWave`(Amount·Wavelength·Phase 3 标量)，static getter/setter。全 OneD float64-BE @ cdat[0:8]，嵌套于 Taper/Wave group `LIST(tdgp)`（`findGroupBody` 下钻 + 复用 `overwriteShapeStreamCdat`）。stroke body 模板从 `gen_shape_all_full.jsx`(扩展设 Taper+Wave，Units 留 % → 9 active slot)重抽 → **仅 stroke .bin 变更**（rect/ellipse/fill/path md5 不变，零回归）。**新 Taper/Wave ship-gate + 全部 stroke-touching ship-gate（Stroke/StrokeKf/ShapeEnums/RectSubKf/FillOpKf）双版本 AE2020+2025 均 PASS**。Go round-trip 全绿，Stable API。RE 自 `re_stroke_dtw.jsx`，详 `incident-reports/stroke-line-cap-join-miter-re.md` Taper/Wave addendum + coverage 子项⑫。**deferred**：Taper Length Units/Px + Wave Units/Cycles（% / Wavelength 模式被 elide）；**Stroke Dashes**（变长 N×Dash/Gap + Offset hidden-until-enabled，下一条接力）。〔前一条：shape-node enum sweep，commit 8c8f43d〕详 logbook。

〔历史：**shape-node enum sweep**（commit 8c8f43d，承 Stroke Line Cap/Join/Miter 之后）：Rect/Ellipse `Direction` + Fill `BlendMode`/`CompositeOrder`/`FillRule` + Stroke `BlendMode`/`CompositeOrder` getter/setter（typed enum + 校验）。全 OneD float64-BE @ cdat[0:8]。rect/ellipse/fill/stroke 4 模板从单一 `v2_2_shape_all_full.aep` 重抽（含 enum slot）→ **重跑全部 shape ship-gate（9 gate × AE2020+2025）均 PASS，模板 swap 零回归**。Go round-trip 全绿，Stable API。RE 详 `incident-reports/stroke-line-cap-join-miter-re.md`（含 enum addendum）+ coverage 子项⑪。〔前一条：Stroke Line Cap/Join/Miter，commit 4beb7fd〕详 logbook。〕
**Active focus**: 无 active 实现线。

## Next session

1. **从下方长线 backlog 选下一条实现线**（用户定方向）。

> **长线 backlog**（大 arc 或缺 runtime setter）：

1. **Path keyframe**（逐帧 bezier）— V2.3+ 级大 arc。
2. **Layr Transform 3D 通道**（Orientation / Rotate X/Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。
3. ~~**Stroke Line Cap/Join/Miter**~~ — ✅ **已 ship**（2026-05-31, commit 4beb7fd, 见 Last-updated + logbook + coverage 子项⑩）。
4. **Gradient W**（SetGradient）— 大 arc。groundwork：默认 gradient 被 AE elide（须自定义 stops 强制 emit）；无现成 in-repo fixture；存储 `GCst > GCky > Utf8(prop.map XML)`（parse_properties.go）。需 gradient-fill fixture + XML RE + 序列化器 + API + 双版本 gate。
5. **Stroke Dashes** — 变长嵌套组（N×Dash/Gap 对 + Offset，hidden-until-enabled，需 enable/reveal runtime 模型）。RE 已知结构（`re_stroke_dtw.jsx`，7 个 pre-existing slot，AE 只 emit enabled pair）。〔~~Taper/Wave~~ 已 ship 见子项⑫；嵌套组 cdat 下钻法（`findGroupBody`）可复用，但 Dashes 变长 + hidden 语义比 Taper/Wave 复杂〕

**其它候选**：泛型 `DuplicateItem`（无 scripting API）、`ImportComposition`（需求驱动）。其余 deferred R-only（DisplayColorSpace / ValueText 等）见 logbook § Deferred。

## Hanging tasks

无。
