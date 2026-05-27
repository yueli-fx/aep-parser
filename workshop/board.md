# Board — aep-parser

**Last updated**: 2026-05-27 by claude (review fixup #2 + P2b 2C Gradient R ship + Script Alert dialog rule)
**Active focus**: py-aep parity — P1 / P2a (sans Task 5) / P2b (sans DisplayColorSpace + Gradient W) / P2c (full) all ship

## Next session

**待 user 决定 next direction**：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需要 fixture（py-aep 样本有多 Utf8 keyframe 数据可参考）
- **V3**: runtime IR + capability framework — 方向性规划

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

- **2026-05-27 P2b 2C Gradient R + review fixup #2** — Gradient XML R 真正接通：fix 上一 session 的 dead-code（XML 写成从 cdat 解，实际在 `GCst → GCky → Utf8`）。新增 `parseGradientStopsProperty` 处理 GCst 包装。rifx 加 `IDGCst / IDGCky`。fixture 走 py-aep `samples/models/property/gradient.aep`（其本身带 1707/1789B 真实 prop.map XML）。新 `TestGradient_FixturePyAep` PASS。**清理**：`gradient.go` 删手写 `itoa` → `strconv.Itoa`；删 trivial `parseGradientXML` private wrapper（只留 `ParseGradientXML`）。**Script Alert 自动化**：`scripts/ae_dialog_rules.json` 加 `script-alert` 规则（windowTitle "Script Alert" + OCR fallback；Enter dismiss），解决 user 反馈的 JSX 弹窗需手动 OK 问题。`re_gradient.jsx` 改 `.done` 契约 + `app.quit()`（替代原 `alert()`），跳过曾经 throw 的 G-Fill 步骤（保留 G-Stroke RE 路径）。PASS 238 (+1)。docs 同步：spec §2.4 Gradient 行 ❌ → ✅ R，coverage 加 P2b 2C 段，deferred 删 Gradient 占位行。
- **2026-05-27 P2c followup DefaultValue/LastValue/NbOptions + pard infra** — 新 `pard` chunk 解析基础设施（`parse_effect_pard.go`）：从 sspc 内 parT LIST 提取 effect 参数定义（lastValue / defaultValue / nbOptions / minValue / maxValue），支持 Scalar/Angle/Boolean/TwoD/Enum/Slider/ThreeD 7 种 control type。`Property` 新增 `DefaultValue / LastValue / NbOptions` 字段 + transform 硬编码默认值表（`property_defaults.go`）。parser 在 `collectEffects` 后自动 merge pard 元数据，在 `parseComposition` 后对 transform 属性赋默认值。PASS 232 (+3)。spec §2.4 表 1 行从 ❌ → ✅ done。
- **2026-05-27 P2c followup Property metadata reads** — ControlType/ValuePropertyType 枚举推导 + MinValue/MaxValue tdum/tduM 解码 + UnitsText 静态 map + PropertyIndex/PropertyDepth。PASS 229 (+6)。
- **2026-05-27 P2c PropertyGroup hierarchy MVP** — AEPropertyGroup 树 + PropertyBase interface + 11 typed group accessor。PASS 224 (+5)。
- **2026-05-27 P2a/P2b review fixup** — 删 ImportPlaceholder + DisplayColorSpace stub；CMS enum 校验。PASS 219。

## Hanging tasks

无。
