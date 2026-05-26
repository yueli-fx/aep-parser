# Board — aep-parser

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 接下来 + 最近 5 条归档 + PASS count 单一权威源。

**Last updated**: 2026-05-26 by claude (V2.2 alpha 闭环；新路线 = py-aep API 全 parity；PASS = **175 / 0 FAIL**)
**Active focus**: py-aep parity Phase 1 — filter views + frame-time accessor + Project single-field chunks（~83 新 API，5-7 天估算）

## Next session

1. 验证 PASS = **175** + `go vet ./...` clean
2. 起 Phase 1 Task 1A（最低悬果） — `internal/aep/composition_views.go` 加 15 个 filter view (TextLayers/ShapeLayers/CameraLayers/LightLayers/NullLayers/SolidLayers/AdjustmentLayers/ThreeDLayers/GuideLayers/SoloLayers/AVLayers/CompositionLayers/FootageLayers/FileLayers/PlaceholderLayers) + test + docs sync。详 [P1 plan Task 1A](plans/2026-05-26-py-aep-parity-p1-plan.md#task-1a--compitem-filter-views15-apipure-helper)
3. Task 1A 完工 → 进 1B (Project filter + LayerByID + EffectNames) → 1C (frame-time accessor) → ...
4. 全局策略 + Phase 2/3 路线 → [py-aep parity spec](specs/2026-05-26-py-aep-parity-design.md)

## In flight

- **py-aep parity Phase 1** — 9 个子任务（1A-1I），目标 +50 PASS / +83 API。spec [`specs/2026-05-26-py-aep-parity-design.md`](specs/2026-05-26-py-aep-parity-design.md)，plan [`plans/2026-05-26-py-aep-parity-p1-plan.md`](plans/2026-05-26-py-aep-parity-p1-plan.md)

## Blockers

无。

## Deferred

- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化）— 跟 py-aep parity 并行；用户用 AE create fixture 后可起
- **V3 capability framework** — Layer.Remove/Duplicate/Move/PropertyBase 结构性 ops 都靠这套；进 Phase 3
- `environmentLayer` 360° 素材 / `ligature` OT liga 字体 / `maskFeatherFalloff` 位置未 RE
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"

## Recently finished

- **2026-05-26 workshop 格式 v0.8.0 对齐 + py-aep parity 路线落地** — board.md 重组按 v0.8.0 模板；specs/plans 加 YYYY-MM-DD- 前缀；V2.1 入 finish/；新增 py-aep parity spec + Phase 1 plan（PASS 175 不变）。
- **2026-05-26 V2.2 Phase 6 docs sync + v0.8.0 frontmatter migration** — `docs/shape.md` 加 V2.2 alpha Builder API section + 限制声明；`coverage.md` + `v2-2 spec §6.4a/§6.5/§10` finalize；10 scars + 3 playbooks 全补 frontmatter。PASS 175 不变。
- **2026-05-25 V2.2 Phase 5 iter-8** — embed Rect+Fill body bytes（`v2_2_shape_rect_body.bin` 448B + `v2_2_shape_fill_body.bin` 426B）；variants #2-7 全 PASS layers=1。Phase 5 ship gate 完整闭环。PASS 175。
- **2026-05-25 V2.2 Phase 5 iter-7** — transplant 法 isolate + embed Transform Group bytes（`v2_2_transform_group_body.bin` 1842B）；variant #2 empty ShapeLayer 首次 PASS layers=1，跨第一道 silent-drop 闸门。
- **2026-05-25 V2.2 Phase 5 iter-5b** — Gide + Ewst layer-skel boilerplate 补齐（PASS 175），pivot byte-patching → semantic RE。详 [scars/v2-2-aelayer-structure.md](scars/v2-2-aelayer-structure.md)。

## Hanging tasks

无。
