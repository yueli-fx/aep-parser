# Board — aep-parser

> 文档分工见 [../CLAUDE.md](../CLAUDE.md) 场景触发器表。本文件 = 现在在做啥 + 接下来 + 最近 5 条归档 + PASS count 单一权威源。

**Last updated**: 2026-05-26 by claude (V2.2 alpha 完整闭环 — Phase 5 ship gate + Phase 6 docs sync done; PASS = **175 / 0 FAIL**)
**Active focus**: V2.2 alpha 已闭环；下一程未定（V2.2.1 vs V3 方向）

## Next session

1. 验证 PASS = **175** + `go vet ./...` clean
2. 选 V2.2.1 任一条做（见 [v2-2 plan](plans/2026-05-22-v2-2-layer-creation-plan.md) V2.2.1 候选段）：
   - Ellipse/Path/Stroke embed bytes（用户先用 AE 写 `re_v2_2_{ellipse,path,stroke}.aep` fixture）
   - Fill Color 编码 RE（tolerance cdat 跟 JSX 0..1 输入不对齐）
   - keyframe 持久化（embed body 现只 static cdat slot；加 LIST(list) lhd3/ldat 注入路径）
3. 或定 V3 方向 — 见 [v3-direction](specs/v3-direction.md) + [architecture](specs/architecture.md)

## In flight

无 — V2.2 alpha 闭环后下一程主线未选。

## Blockers

无。

## Deferred

- `environmentLayer` — 需 360° 素材
- `ligature` — 需 OT `liga` 字体 fixture（默认 false 无 diff）
- `maskFeatherFalloff` — 位置未 RE
- V2.2.1 Ellipse/Path/Stroke embed — 等用户 AE create fixture
- Composition.SetRenderer / ldta 零值区 probe / Footage proxy / Project nhed 扩展 — 见 [coverage.md](plans/coverage.md) "剩余可探方向"

## Recently finished

- **2026-05-26 V2.2 Phase 6 docs sync + v0.8.0 frontmatter migration** — `docs/shape.md` 加 V2.2 alpha Builder API section + 限制声明；`coverage.md` + `v2-2 spec §6.4a/§6.5/§10` finalize；10 scars + 3 playbooks 全补 `when_to_read`/`applies_to`/`last_updated` 满足 v0.8.0 hard-fail。PASS 175 不变。
- **2026-05-25 V2.2 Phase 5 iter-8** — embed Rect+Fill body bytes（`v2_2_shape_rect_body.bin` 448B + `v2_2_shape_fill_body.bin` 426B）；variants #2-7 全 PASS layers=1。Phase 5 ship gate 完整闭环。PASS 175。
- **2026-05-25 V2.2 Phase 5 iter-7** — transplant 法 isolate + embed Transform Group bytes（`v2_2_transform_group_body.bin` 1842B）；variant #2 empty ShapeLayer 首次 PASS layers=1，跨第一道 silent-drop 闸门。
- **2026-05-25 V2.2 Phase 5 iter-5b** — Gide + Ewst layer-skel boilerplate 补齐（PASS 175），pivot byte-patching → semantic RE。详 [scars/v2-2-aelayer-structure.md](scars/v2-2-aelayer-structure.md)。
- **2026-05-25 V2.2 Phase 5 iter-5** — Root Vectors Group 5-层嵌套 emit + `WrapShapeLayer` 语义变更（PASS 174）。

## Hanging tasks

无。
