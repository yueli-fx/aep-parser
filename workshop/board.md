# Board — aep-parser

**Last updated**: 2026-05-27 by claude (P2c followup#2 — P2 parity 收尾 + V3 brainstorm 起)
**Active focus**: py-aep parity P2 范围全闭环（剩余 ❌ 全部 V3/P3 结构性），开始 V3 深度思考

## Next session

**V3 深度思考**：runtime IR + capability framework。本 session 已起 brainstorm，详 `specs/2026-05-27-v3-deep-think.md`（如已写）或 `specs/2026-05-22-v3-direction.md`（原方向）。剩余 ❌ 行（Render Queue / Essential Graphics / Layer structural ops / DimensionsSeparated 等）全在 V3 范围。

**并行 R-only 仍 deferred**：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

**并行候选**：V2.2.1 ShapeLayer 拓展 / V3 brainstorm。

## In flight

无。

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"
- **`linearizeWorkingSpace` ScriptingAPI quirk** — chunk byte 跟 AE 自己写一致但 ScriptingAPI 读不到 true，归 OCIO/CMS-联动；详 [scars/project-flag-chunks-lnrb-lnrp.md](scars/project-flag-chunks-lnrb-lnrp.md)

## Recently finished

- **2026-05-27 P2c followup#2 — P2 parity 收尾** — 一波拿下 spec 里 P2 scope 剩下的 quick wins，没碰 V3/P3 结构性。新加 `Property.IsModified() / Enabled() / Active() / Elided() / IsNameSet()`（Enabled 读 tdsb byte3 bit0，IsModified 走 animated/expression/value≠default 三轨）+ `AEPropertyGroup.IsModified()`（indexed group 有 children = modified；否则递归）+ `Layer.AVSource() / CanSetCollapseTransformation() / CanSetTimeRemapEnabled()`（py-aep `AVLayer.can_set_*` parity）+ `Project.XmpPacket() R`（trailing UTF-8 after RIFX，已经 round-trip 走 `root.Trailing`，只缺 accessor）。spec §2.1 XmpPacket / §2.3 LayerType+CanSet+IsModified / §2.4 IsModified+ValueType / §2.5 MainSource / §2.9 Application.Version 等 7 行刷新成 ✅ done。新加 21 个 PASS test。**PASS 259**（基线 238 → 259，+21）。剩余 ❌ 全部 V3/P3：Render Queue / Essential Graphics / MotionGraphics / Layer structural ops / DimensionsSeparated / ValueText。
- **2026-05-27 P2b 2C Gradient R + review fixup #2** — Gradient XML R 真正接通：fix 上一 session 的 dead-code（XML 写成从 cdat 解，实际在 `GCst → GCky → Utf8`）。新增 `parseGradientStopsProperty` 处理 GCst 包装。rifx 加 `IDGCst / IDGCky`。fixture 走 py-aep `samples/models/property/gradient.aep`（其本身带 1707/1789B 真实 prop.map XML）。新 `TestGradient_FixturePyAep` PASS。**清理**：`gradient.go` 删手写 `itoa` → `strconv.Itoa`；删 trivial `parseGradientXML` private wrapper（只留 `ParseGradientXML`）。**Script Alert 自动化**：`scripts/ae_dialog_rules.json` 加 `script-alert` 规则（windowTitle "Script Alert" + OCR fallback；Enter dismiss），解决 user 反馈的 JSX 弹窗需手动 OK 问题。`re_gradient.jsx` 改 `.done` 契约 + `app.quit()`（替代原 `alert()`），跳过曾经 throw 的 G-Fill 步骤（保留 G-Stroke RE 路径）。PASS 238 (+1)。docs 同步：spec §2.4 Gradient 行 ❌ → ✅ R，coverage 加 P2b 2C 段，deferred 删 Gradient 占位行。
- **2026-05-27 P2c followup DefaultValue/LastValue/NbOptions + pard infra** — pard chunk 解析（`parse_effect_pard.go`）；7 种 control type；transform 默认值表。PASS 232 (+3)。
- **2026-05-27 P2c followup Property metadata reads** — ControlType/ValuePropertyType + MinValue/MaxValue + UnitsText + PropertyIndex/PropertyDepth。PASS 229 (+6)。
- **2026-05-27 P2c PropertyGroup hierarchy MVP** — AEPropertyGroup 树 + PropertyBase interface + 11 typed group accessor。PASS 224 (+5)。

## Hanging tasks

无。
