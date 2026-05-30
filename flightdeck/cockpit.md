# Cockpit — aep-parser

**Last updated**: 2026-05-31 by claude（**shape-node enum sweep landed on main**（commit 8c8f43d，承 Stroke Line Cap/Join/Miter 之后）：Rect/Ellipse `Direction` + Fill `BlendMode`/`CompositeOrder`/`FillRule` + Stroke `BlendMode`/`CompositeOrder` getter/setter（typed enum + 校验）。全 OneD float64-BE @ cdat[0:8]。rect/ellipse/fill/stroke 4 模板从单一 `v2_2_shape_all_full.aep` 重抽（含 enum slot）→ **重跑全部 shape ship-gate（9 gate × AE2020+2025）均 PASS，模板 swap 零回归**。Go round-trip 全绿，Stable API。RE 详 `incident-reports/stroke-line-cap-join-miter-re.md`（含 enum addendum）+ coverage 子项⑪。〔前一条：Stroke Line Cap/Join/Miter，commit 4beb7fd〕详 logbook。）
**Active focus**: 无 active 实现线。

## Next session

1. **从下方长线 backlog 选下一条实现线**（用户定方向）。

> **长线 backlog**（大 arc 或缺 runtime setter）：

1. **Path keyframe**（逐帧 bezier）— V2.3+ 级大 arc。
2. **Layr Transform 3D 通道**（Orientation / Rotate X/Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。
3. ~~**Stroke Line Cap/Join/Miter**~~ — ✅ **已 ship**（2026-05-31, commit 4beb7fd, 见 Last-updated + logbook + coverage 子项⑩）。
4. **Gradient W**（SetGradient）— 大 arc。groundwork：默认 gradient 被 AE elide（须自定义 stops 强制 emit）；无现成 in-repo fixture；存储 `GCst > GCky > Utf8(prop.map XML)`（parse_properties.go）。需 gradient-fill fixture + XML RE + 序列化器 + API + 双版本 gate。
5. **Stroke Dashes/Taper/Wave** — 嵌套组，低价值，缺 runtime setter（嵌套组结构比 OneD enum 复杂）。〔Fill/Stroke BlendMode·CompositeOrder + Rect/Ellipse Direction + Fill Rule 已 ship，见 Last-updated + coverage 子项⑪〕

**其它候选**：泛型 `DuplicateItem`（无 scripting API）、`ImportComposition`（需求驱动）。其余 deferred R-only（DisplayColorSpace / ValueText 等）见 logbook § Deferred。

## Hanging tasks

无。
