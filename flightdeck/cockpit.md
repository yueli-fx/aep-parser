# Cockpit — aep-parser

**Last updated**: 2026-06-03 by claude（preflight 接力树上未提交的 3C PropertyBase 结构性工作：`AEPropertyGroup.Remove()` + `MoveTo()` —— INDEXED_GROUP 直接子节点删/重排，纯 tdmn+payload pair splice 无 count chunk，V2.1 atomic。复跑 dual ship-gate 4/4 PASS（Effect Parade × remove/move × AE2020/2025）—— `incidents/property-indexed-group-structural-re.md`。同回合跑 flightdeck 2.2 deck 迁移）
**Active focus**: **py-aep parity P3** — §3A Render Queue read+write **完成（含结构性增删）**。**SetRenderer / Guides R/W / 分离维度 R+W(双向) / 3D Orientation R / Essential Graphics R / RQ SetComment + Remove + Add / 3C PropertyBase Remove + MoveTo(ship-gated) 已落**。剩 ValueText（需 AE schema DB）+ DimensionsSeparated animated 子方向 + 3C Duplicate。

## Next session

1. **其余 P3**：`Property.ValueText`（formatted value，需 AE schema DB，最大 RE，多会话）；DimensionsSeparated animated 子方向（keyframe 流迁移）；**3C PropertyBase.Duplicate**（Remove/MoveTo 已 ship-gated 落地；Duplicate 待 clone display-name " 2" 后缀 byte-check —— 见 `incidents/property-indexed-group-structural-re.md` slice 2）；3G comp marker 增删。RQ 域已完整（read + value W + 结构性增删）。详 spec §3 Phase 3。
2. **Property synthesis**（可选大 feature，暂搁）：AE 省略未改的默认属性，py-aep 合成完整 transform schema 我们不合成。补合成 + `Elided` 是唯一让 transform group 对齐 py-aep 长度的路。详 `incidents/transform-group-default-omission.md`。次要 fidelity：animated orientation 的 easing/tangents（旧 1D layout 未校验）。
2. **ship-gate wrapper 改进已落**（`19d8922`）：splash Ignore action + grace 15→30s。后续若 OCR 被其他窗口遮挡（occlusion）仍可能误报，见 `incidents/ae-automation-occlusion-crashstate.md`。

## Backlog（长线大 arc / 缺 runtime setter）

**暂停中的大 arc（spec 部分完成，非 pending 非 done）**：

- **py-aep parity P3** — **§3A Render Queue read+write+结构性增删 完整**（2026-06-01 8 slice + 2026-06-02 SetComment/RemoveItem/AddItem ship-gated）；**DimensionsSeparated R/W 双向 2D+3D 已落**（2026-06-02）；**3C PropertyBase Remove + MoveTo 已落**（2026-06-03，Effect Parade 双版本 ship-gate）。剩 Property.ValueText（需 AE schema DB）+ DimensionsSeparated animated + 3C Duplicate + 3G marker 增删。详 `specs/2026-05-26-py-aep-parity-design.md`。
- **V3 收尾** — M8 物理分包（scene/serializer 拆包，scene→rifx 残留 5 项白名单清零）+ 通用 capability matrix 完整化 + ShapeGraph/EffectSchema。详 `specs/2026-05-22-v3-direction.md`（结构性 Phase 1-5 + 包重组方案① 已落）。

**单条候选**：

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API，需纯 RE）。
3. `ImportComposition`（需求驱动）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 [sketches/deferred-backlog.md](sketches/deferred-backlog.md) + [plans/coverage.md](plans/coverage.md)。

（已落条目历史见 `git log` + [landed/HISTORY.md](landed/HISTORY.md)，不在 cockpit 留存。）

## Hanging tasks

无。
