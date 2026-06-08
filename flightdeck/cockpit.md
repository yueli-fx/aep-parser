# Cockpit — aep-parser

**Last updated**: 2026-06-08 by claude
**Active focus**: **V3 M8 方案②（真·物理分包）执行中** — plan `plans/2026-06-07-v3-m8-physical-split-plan.md`（active）。P0 基线+inventory ✅ · P1 抽 `internal/codec` ✅ · **P2 back-ref 接口化 5/10**（已倒置 Composition / Marker / Mask / Footage / Keyframe；余 Layer / Property / Project / PropertyGroup / RenderQueue）。每 commit 绿 + byte-identical round-trip。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① + M8 scene→rifx 白名单清零 已落；真·物理分包(方案②接口倒置)执行中 P2 5/10） — [note: 大 arc 暂停 — 真·物理分包(方案②接口倒置)执行中，见 M8 plan P2 5/10；结构性 Phase 1-5 + 包重组方案① + scene→rifx 白名单清零 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落；§3H ValueText defer）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

1. **续 P2 接口化**（plan Task 2.2 Step 7）：剩 5 类 concrete→XWriter，**Property 最后**（shape/stroke/fill setter 全委托 `Property.SetStaticValue`，接口化后自动经 `prop.back.WriteStaticValue`）。每类独立 commit + byte-identical + setter 单测 + attach 完整性断言。
2. P2 收口（Task 2.3：scene_* rifx 引用清零核查）→ **P3** git mv 物理分包（编译期硬边界）→ **P4** 下游切换 + 退役 AST 守卫 + CLAUDE.md + 双版本 ship-gate。
3. docgen 次要 follow-on（非阻塞）：README 英文化 + 类型级 / package-func Example 渲染。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
