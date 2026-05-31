# Cockpit — aep-parser

**Last updated**: 2026-05-31 by claude
**Active focus**: 无 active 线 — Path Keyframe arc 已收官 ship。下条从 backlog 选，等用户定方向。无 in-flight WIP，工作树 commit 后应为 clean。

## Next session

1. **从 backlog 选下一条（等用户定方向）**。

## Backlog（长线大 arc / 缺 runtime setter）

**暂停中的大 arc（spec 部分完成，非 pending 非 done）**：

- **py-aep parity P3** — Render Queue 全域（~3-5k LOC）/ Essential Graphics / Guides / Composition.Renderer W / Property.ValueText / DimensionsSeparated。详 `specs/2026-05-26-py-aep-parity-design.md`（P1/P2 已落，P3 未启动）。
- **V3 收尾** — M8 物理分包（scene/serializer 拆包，scene→rifx 残留 5 项白名单清零）+ 通用 capability matrix 完整化 + ShapeGraph/EffectSchema。详 `specs/2026-05-22-v3-direction.md`（结构性 Phase 1-5 + 包重组方案① 已落）。

**单条候选**：

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API，需纯 RE）。
3. `ImportComposition`（需求驱动）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 logbook § Deferred + `flight-plans/coverage.md`。

## Recently finished（≤5，详 logbook.md）

- 2026-05-31 Shape path keyframe write（子项⑮，linear）— 双版本 ship-gate PASS
- 2026-05-31 Gradient fill write SetGradient（子项⑭）— 双版本 ship-gate PASS
- 2026-05-31 Stroke Dashes（子项⑬，单 Dash+Gap 对）— 双版本 ship-gate PASS
- 2026-05-31 shape-node enum sweep（子项⑪，commit 8c8f43d）— 双版本 ship-gate PASS
- 2026-05-31 Stroke Line Cap/Join/Miter（子项⑩，commit 4beb7fd）— 双版本 ship-gate PASS

## Hanging tasks

无。
