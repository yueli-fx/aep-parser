# Board — aep-parser

**Last updated**: 2026-05-27 by claude (CLAUDE.md 硬约束 V3 时代刷新 — #1/#2/#3 措辞更新 + 新加 #5 opaque preservation + #6 AE 接受 gate)
**Active focus**: V3 Phase 1 起步——`back *backrefs` shard 抽取 refactor。py-aep parity P2 范围全闭环。

## Next session

**V3 Phase 1 plan + 执行**：纯 refactor，把 `Property.tdb4 / cdat / ldat / ...`、`Layer.ldta`、`Composition.cdta` 等 chunk refs 从 struct top-level 抽到 `back *backrefs` shard。预期 1-2 session，PASS 259 全过、WriteAEP 字节级一致、public API 零 delta。spec: `specs/2026-05-27-v3-deep-think.md` § 4。

**起步 checklist**：
1. 读 `specs/2026-05-27-v3-deep-think.md` 全文（266 行）+ CLAUDE.md 新硬约束 #5 / #6
2. user 确认 4 个决策点（spec §6）—— 软迁移 / opaque shard 现加 / Phase 1 纯 refactor / 多周时间表
3. 写 `plans/2026-05-28-v3-phase1-backrefs-plan.md`
4. 执行（subagent-driven，1 文件 1 refactor 1 测试 pass）

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

**并行候选**：V2.2.1 ShapeLayer 拓展（其实最好压到 V3 Phase 5 一起做，避免 alpha API 重复）。

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

- **2026-05-27 CLAUDE.md 硬约束 V3 时代刷新** — #1 length-preserving 加上"结构性 ops 走 V2.1 atomic invariants"补丁；#2 拆 Stable / Alpha 两级（撤销新加但已知 broken 的 API 不算违反 — 上次删 `ImportPlaceholder` 时的犹豫不再存在）；#3 单 package 理由更新为"文件名规约 + lint > 子包 re-export 噪音"；新加 #5 opaque preservation（V2.2 教训显化）+ #6 AE 接受 gate（结构性写必须双版本 ship-gate）。配套 `specs/2026-05-27-v3-deep-think.md` § 1.1 / § 2.5 / § 8 引用这些新硬约束。**这之前 V3 决策得绕这些约束的旧版本走，现在显化后 V3 Phase 1 plan 可以直接遵循**。
- **2026-05-27 V3 deep-think spec** — `specs/2026-05-27-v3-deep-think.md`（266 行）。验证 M1-M8（M3 ShapeGraph 延后到 Phase 5；M4 effect schema 出 V3 scope）+ 7 个 open question 含推荐答案 + 7 行风险表 + Phase 1 纯 refactor 起步建议（back-ref 抽取，公开 API 零 delta，1-2 session 完成）+ 4 个决策点等 user 拍。
- **2026-05-27 P2c followup#2 — P2 parity 收尾** — 一波拿下 spec 里 P2 scope 剩下的 quick wins，没碰 V3/P3 结构性。新加 `Property.IsModified() / Enabled() / Active() / Elided() / IsNameSet()`（Enabled 读 tdsb byte3 bit0，IsModified 走 animated/expression/value≠default 三轨）+ `AEPropertyGroup.IsModified()`（indexed group 有 children = modified；否则递归）+ `Layer.AVSource() / CanSetCollapseTransformation() / CanSetTimeRemapEnabled()`（py-aep `AVLayer.can_set_*` parity）+ `Project.XmpPacket() R`（trailing UTF-8 after RIFX，已经 round-trip 走 `root.Trailing`，只缺 accessor）。spec §2.1 XmpPacket / §2.3 LayerType+CanSet+IsModified / §2.4 IsModified+ValueType / §2.5 MainSource / §2.9 Application.Version 等 7 行刷新成 ✅ done。新加 21 个 PASS test。**PASS 259**（基线 238 → 259，+21）。剩余 ❌ 全部 V3/P3：Render Queue / Essential Graphics / MotionGraphics / Layer structural ops / DimensionsSeparated / ValueText。
- **2026-05-27 P2b 2C Gradient R + review fixup #2** — Gradient XML R 真正接通：fix 上一 session 的 dead-code（XML 写成从 cdat 解，实际在 `GCst → GCky → Utf8`）。新增 `parseGradientStopsProperty` 处理 GCst 包装。rifx 加 `IDGCst / IDGCky`。fixture 走 py-aep `samples/models/property/gradient.aep`（其本身带 1707/1789B 真实 prop.map XML）。新 `TestGradient_FixturePyAep` PASS。**清理**：`gradient.go` 删手写 `itoa` → `strconv.Itoa`；删 trivial `parseGradientXML` private wrapper（只留 `ParseGradientXML`）。**Script Alert 自动化**：`scripts/ae_dialog_rules.json` 加 `script-alert` 规则（windowTitle "Script Alert" + OCR fallback；Enter dismiss），解决 user 反馈的 JSX 弹窗需手动 OK 问题。`re_gradient.jsx` 改 `.done` 契约 + `app.quit()`（替代原 `alert()`），跳过曾经 throw 的 G-Fill 步骤（保留 G-Stroke RE 路径）。PASS 238 (+1)。docs 同步：spec §2.4 Gradient 行 ❌ → ✅ R，coverage 加 P2b 2C 段，deferred 删 Gradient 占位行。
- **2026-05-27 P2c followup DefaultValue/LastValue/NbOptions + pard infra** — pard chunk 解析（`parse_effect_pard.go`）；7 种 control type；transform 默认值表。PASS 232 (+3)。

## Hanging tasks

无。
