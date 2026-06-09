---
status: idea
summary: Deferred backlog 细项（从退役 logbook §Deferred 迁入）
---

# Deferred backlog（细项）

> 从退役的 flightdeck/logbook.md §Deferred 迁入。大 arc 见 cockpit Backlog + specs/。


- **V2.2.1 剩余 keyframe / 子属性** — 〔Layr Position keyframe ✅ 子项⑤、shape path keyframe ✅ 子项⑮〕各 shape 次要子属性（Fill/Stroke Opacity·BlendMode·CompositeOrder、Stroke Line Cap/Join/Miter/Dashes/Taper/Wave、Direction、Layr Transform Anchor/Scale/Rotation/Opacity，多 runtime-only）。详 coverage.md 子项④。
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- **Essential Graphics W**（创建 controller / 绑定 override UUID）+ **Gradient stroke W**（GradientStroke 节点未实现）—— py-aep parity 路线图（归档 `archive/specs/2026-05-26-py-aep-parity-design.md`）收尾后仅存的真未实现写线头
- ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](../plans/coverage.md) "剩余可探方向"（`Composition.SetRenderer` 已 ship，移出本表）
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [incidents/project-flag-chunks-lnrb-lnrp.md](../incidents/project-flag-chunks-lnrb-lnrp.md)
