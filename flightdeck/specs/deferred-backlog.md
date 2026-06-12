---
status: idea
summary: Deferred backlog 细项（从退役 logbook §Deferred 迁入）
---

# Deferred backlog（细项）

> 从退役的 flightdeck/logbook.md §Deferred 迁入。大 arc 见 cockpit Backlog + specs/。


- **V2.2.1 剩余 keyframe / 子属性** — 〔Layr Position keyframe ✅ 子项⑤、shape path keyframe ✅ 子项⑮〕各 shape 次要子属性（Fill/Stroke Opacity·BlendMode·CompositeOrder、Stroke Line Cap/Join/Miter/Dashes/Taper/Wave、Direction、Layr Transform Anchor/Scale/Rotation/Opacity，多 runtime-only）。详 coverage.md 子项④。
- **V3 大子项残余**（capability matrix 完整化 / ShapeGraph / EffectSchema）— 框架 M1-M8 已全落（结构性 ops 全 ship，V3 spec 已归档 2026-06-12）；这三个大子项无需求驱动不做
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- ~~**Essential Graphics W**~~ ✅ 2026-06-12 全程 ship（SetMotionGraphicsTemplateName + AddEssentialProperty，双版本 gate，Stable）；剩 point/dropdown/text/Transform 源 controller + RemoveEssentialProperty 需求驱动，详 `incidents/essential-graphics-write-re.md`（Gradient stroke W 已于 2026-06-10 ship，见 coverage 子项⑯）
- ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](../plans/coverage.md) "剩余可探方向"（`Composition.SetRenderer` 已 ship，移出本表）
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [incidents/project-flag-chunks-lnrb-lnrp.md](../incidents/project-flag-chunks-lnrb-lnrp.md)
