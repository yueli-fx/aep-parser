# Cockpit — aep-parser

**Last updated**: 2026-06-01 by claude（P3 §3A RQ read+write 8 slice 全 ship + commit；Renderer W RE 探明记 sketch）
**Active focus**: **py-aep parity P3** — §3A Render Queue **read+write 主体完成**（8 commit `fb657cb`→`54f65f1`）：reader（结构/RenderSettings/OMSettings/FormatOptions）+ writer（render settings / OM settings / item scalar / time span，**Alpha** length-preserving）。剩 RQ 结构性增删 + comment 写 + 派生枚举/XML format options。

## Next session

1. **Renderer W 结构性 arc**（RE 已探明 → `sketches/2026-06-01-renderer-write-re-findings.md`）：prin 改名 length-preserving + prda 变长替换 = 结构性 → 抽 4 prda 模板 + atomic SetRenderer + **AE 2020+2025 双版本 ship-gate**（AE 在 `E:\adobe`）。⚠ 进入 ship-gate 慢档。
2. 其余 P3 同需 ship-gate（DimensionsSeparated / RQ 增删）或大 RE（ValueText / Essential Graphics / Guides）。详 `plans/2026-06-01-py-aep-p3-renderqueue-reader-plan.md` + spec §2.6。

## Backlog（长线大 arc / 缺 runtime setter）

**暂停中的大 arc（spec 部分完成，非 pending 非 done）**：

- **py-aep parity P3** — **§3A Render Queue read+write 已落**（2026-06-01，8 slice）；剩 Essential Graphics / Guides / Composition.Renderer W（RE 已探）/ Property.ValueText / DimensionsSeparated / RQ 结构性增删。详 `specs/2026-05-26-py-aep-parity-design.md`。
- **V3 收尾** — M8 物理分包（scene/serializer 拆包，scene→rifx 残留 5 项白名单清零）+ 通用 capability matrix 完整化 + ShapeGraph/EffectSchema。详 `specs/2026-05-22-v3-direction.md`（结构性 Phase 1-5 + 包重组方案① 已落）。

**单条候选**：

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API，需纯 RE）。
3. `ImportComposition`（需求驱动）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 [sketches/deferred-backlog.md](sketches/deferred-backlog.md) + [plans/coverage.md](plans/coverage.md)。

（已落条目历史见 `git log` + [landed/HISTORY.md](landed/HISTORY.md)，不在 cockpit 留存。）

## Hanging tasks

无。
