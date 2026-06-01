# Cockpit — aep-parser

**Last updated**: 2026-06-02 by claude（Property DimensionsSeparated + IsSeparation{Leader,Follower}/SeparationDimension R 落地，2 测试绿；顺带发现 3D 层 Transform Group 解析截断 bug → incident）
**Active focus**: **py-aep parity P3** — §3A Render Queue read+write 主体完成（`fb657cb`→`54f65f1`，Alpha length-preserving）。**Composition.SetRenderer 已 ship**（`420020c`/`b13358b`，双版本 ship-gate 绿，off-alpha）。**Composition Guides R/W 已落**（Alpha）。**Property 分离维度 R 已落**（DimensionsSeparated + leader/follower/dimension）。剩 RQ 结构性增删 + comment 写 + 派生枚举/XML format options。

## Next session

1. **其余 P3**：需 ship-gate（DimensionsSeparated **W** / RQ 结构性增删）或大 RE（ValueText / Essential Graphics）。详 `plans/2026-06-01-py-aep-p3-renderqueue-reader-plan.md` + spec §2.6。
2. **3D 层 Transform Group 解析截断**（新发现，`incidents/transform-group-3d-truncation.md`）：Orientation 被误判为 0-child group，其后 Position_2/Scale/Rotate Z/Opacity 丢失，unseparated 更甚。独立 parser arc，修它是 3D transform 写路径的前置。
2. **ship-gate wrapper 改进已落**（`19d8922`）：splash Ignore action + grace 15→30s。后续若 OCR 被其他窗口遮挡（occlusion）仍可能误报，见 `incidents/ae-automation-occlusion-crashstate.md`。

## Backlog（长线大 arc / 缺 runtime setter）

**暂停中的大 arc（spec 部分完成，非 pending 非 done）**：

- **py-aep parity P3** — **§3A Render Queue read+write 已落**（2026-06-01，8 slice）；剩 Essential Graphics / Property.ValueText / DimensionsSeparated / RQ 结构性增删（Guides R/W 已落 2026-06-02；Composition.Renderer W 已 ship）。详 `specs/2026-05-26-py-aep-parity-design.md`。
- **V3 收尾** — M8 物理分包（scene/serializer 拆包，scene→rifx 残留 5 项白名单清零）+ 通用 capability matrix 完整化 + ShapeGraph/EffectSchema。详 `specs/2026-05-22-v3-direction.md`（结构性 Phase 1-5 + 包重组方案① 已落）。

**单条候选**：

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API，需纯 RE）。
3. `ImportComposition`（需求驱动）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 [sketches/deferred-backlog.md](sketches/deferred-backlog.md) + [plans/coverage.md](plans/coverage.md)。

（已落条目历史见 `git log` + [landed/HISTORY.md](landed/HISTORY.md)，不在 cockpit 留存。）

## Hanging tasks

无。
