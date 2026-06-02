# Cockpit — aep-parser

**Last updated**: 2026-06-02 by claude（Property.SetDimensionsSeparated W 落地：merge→separate 结构性 toggle，static 3D Position，双版本 ship-gate PASS。Go 输出 chunk 树 byte-structural 等同 AE 自存。关键发现：值迁移 + Position_2 合成 + 组件数分支，非翻 bit —— `incidents/separate-dimensions-write-mechanics.md`）
**Active focus**: **py-aep parity P3** — §3A Render Queue read+write 主体完成。**SetRenderer / Guides R/W / 分离维度 R+W / 3D Orientation R / Essential Graphics R / RQ SetComment(ship-gated) 已落**。剩 RQ 结构性增删 + ValueText（+ DimensionsSeparated 的 merge/2D/animated 子方向）。

## Next session

1. **其余 P3**：需 ship-gate（RQ 结构性增删）或大 RE（ValueText 需 AE schema DB）。DimensionsSeparated W 的 merge→merged / 2D / animated 子方向暂搁（各需独立 RE + ship-gate，见 `incidents/separate-dimensions-write-mechanics.md`）。详 `plans/2026-06-01-py-aep-p3-renderqueue-reader-plan.md` + spec §2.6。
2. **Property synthesis**（可选大 feature，暂搁）：AE 省略未改的默认属性，py-aep 合成完整 transform schema 我们不合成。补合成 + `Elided` 是唯一让 transform group 对齐 py-aep 长度的路。详 `incidents/transform-group-default-omission.md`。次要 fidelity：animated orientation 的 easing/tangents（旧 1D layout 未校验）。
2. **ship-gate wrapper 改进已落**（`19d8922`）：splash Ignore action + grace 15→30s。后续若 OCR 被其他窗口遮挡（occlusion）仍可能误报，见 `incidents/ae-automation-occlusion-crashstate.md`。

## Backlog（长线大 arc / 缺 runtime setter）

**暂停中的大 arc（spec 部分完成，非 pending 非 done）**：

- **py-aep parity P3** — **§3A Render Queue read+write 已落**（2026-06-01，8 slice + 2026-06-02 SetComment ship-gated）；剩 Property.ValueText / DimensionsSeparated W / RQ 结构性增删（Guides R/W + Essential Graphics R + RQ SetComment 已落 2026-06-02；Composition.Renderer W 已 ship）。详 `specs/2026-05-26-py-aep-parity-design.md`。
- **V3 收尾** — M8 物理分包（scene/serializer 拆包，scene→rifx 残留 5 项白名单清零）+ 通用 capability matrix 完整化 + ShapeGraph/EffectSchema。详 `specs/2026-05-22-v3-direction.md`（结构性 Phase 1-5 + 包重组方案① 已落）。

**单条候选**：

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API，需纯 RE）。
3. `ImportComposition`（需求驱动）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 [sketches/deferred-backlog.md](sketches/deferred-backlog.md) + [plans/coverage.md](plans/coverage.md)。

（已落条目历史见 `git log` + [landed/HISTORY.md](landed/HISTORY.md)，不在 cockpit 留存。）

## Hanging tasks

无。
