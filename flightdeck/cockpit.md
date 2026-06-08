# Cockpit — aep-parser

**Last updated**: 2026-06-09 by claude
**Active focus**: **V3 M8 方案②（真·物理分包）执行中** — plan `plans/2026-06-07-v3-m8-physical-split-plan.md`（active）。P0 基线+inventory ✅ · P1 抽 `internal/codec` ✅ · **P2 back-ref 接口化 ✅ 完成**（9 类倒置为接口：Composition/Marker/Mask/Footage/Keyframe/Project/Layer/RenderQueueItem/**Property**；OM/RenderQueue 容器/PropertyGroup 按 §F 故意保 concrete——无 writer 接口，P3 统一解耦）。**Task 2.3 收口 ✅：全 `scene_*.go` 零 rifx import/code-token**。**下一步 = P3 git mv 物理分包**。每 commit 绿 + byte-identical round-trip。

## 进行中

<!-- AUTO:inprogress -->
- [2026-05-22-v3-direction.md](specs/2026-05-22-v3-direction.md) — V3 direction：scene-graph IR + capability matrix + serializer split（Phase 1-5 + 包重组方案① + M8 scene→rifx 白名单清零 已落；真·物理分包(方案②接口倒置)执行中 P2 7/10） — [note: 大 arc 暂停 — 真·物理分包(方案②接口倒置)执行中，见 M8 plan P2 7/10；结构性 Phase 1-5 + 包重组方案① + scene→rifx 白名单清零 已落]
- [2026-05-26-py-aep-parity-design.md](specs/2026-05-26-py-aep-parity-design.md) — py-aep parity API 全覆盖路线图（P1/P2 已落；P3 §3A RQ R/W+结构性 + DimensionsSeparated R/W 双向(static+animated) + 3C PropertyBase Remove/MoveTo/Duplicate + 3G comp marker 增删 已落，剩 ValueText）
- [2026-05-27-v3-deep-think.md](specs/2026-05-27-v3-deep-think.md) — V3 deep think：open questions / risk register / migration strategy
- [2026-06-07-v3-m8-physical-split-design.md](specs/2026-06-07-v3-m8-physical-split-design.md) — V3 M8 方案② 真·物理分包设计（A 先行 + B′ back-ref 接口）：scene/serializer/codec 物理拆包，back-ref 作 scene 内 writer 接口（serializer 实现）→ 保留全部方法 API（不破 API）、scene 编译期零 rifx；eager length-preserving patch 经接口；opaque 延后到 C；全程保 byte-exact 回归门 — [note: brainstorm + 三家外审两轮整合。方向 = A 先行（保 byte-exact，C 日后独立）+ B′（back-ref 接口、不破 API、无侧表/无 Document god-object）。前置：M8 scene→rifx 白名单清零（2026-06-07 已落）]
- [2026-06-07-v3-m8-physical-split-plan.md](plans/2026-06-07-v3-m8-physical-split-plan.md) — V3 M8 方案② 物理分包实现计划（A 先行 + B′）：P0 基线+inventory → P1 抽 internal/codec → P2 单包内 back-ref 接口化（concrete→XWriter） → P3 git mv 物理分包 → P4 下游+收口+双版本 ship-gate
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
- [m8-setter-inventory.md](plans/m8-setter-inventory.md) — V3 M8 Task 0.3 产出 — 全量 Set* 分类表（A 类 back-ref setter / B 类 pure-graph setter）+ 每 backref 结构的 XWriter 接口方法清单。P2（back-ref 接口倒置）的逐类输入。 — [note: 从 internal/aep 实测枚举（277 个 Set* + 10 个 *Backrefs 结构 + 全部 `.back.` 字节访问点），逐方法读 body 分类，非按名猜。spec §3 分类规则 + edge-case adjudication。]
<!-- /AUTO -->

## 下一步

**P2 全部完成**（2026-06-09）。两批 landed：
- **RenderQueue 子系统**（commits ba2865b→5c2eeec）：RQ/OM/Guide settings chunk-alias → scene 独占 copy（单一真相源）+ WriteAEP 单点 sync（`syncRenderQueue`/`syncGuides`）；`SetComment` → `RenderQueueItemWriter`；类型化 offset `codec.RenderSettingOffset`；attach 断言扩 RQ/OM/Guide。
- **Property**（commits 73649c9 A + 8ca1e9c B，最后一类）：A=4 纯 setter（SetStaticValue/SetExpressionEnabled/SetExpression/SetLockedRatio）逻辑迁 `(*propertyBackrefs)` + scene delegate；B=`back *propertyBackrefs` → `PropertyWriter` 接口 + `propertyBack()` helper，全 reader/parse/keyframe-stream/separate-dims 经 helper 取 concrete。**判定**：InsertKeyframe/DeleteKeyframe（重建 scene Keyframes 切片）+ SetDimensionsSeparated（增删 follower Property 节点）= 结构性 → 留 scene-method stopgap（同 RQ AddItem/RemoveItem，D-U3），**不进** PropertyWriter。验证：`rect.SetSize` 委托链 + 全 separate-dims（2D/3D/merge/animated）round-trip 绿。

**P2 收口 ✅**：全 `scene_*.go` 零 rifx import/code-token；OM/RenderQueue-容器/PropertyGroup back 故意保 concrete（§F——无 writer 接口，P3 统一解耦）。

## 下一步（P3 物理分包）

1. **P3** git mv 物理拆包（编译期硬边界）：`scene_*.go`→`internal/scene`、`{parse_,lower_,write_,back_,mutate_}*.go`→`internal/serializer`、`internal/aep` 收为薄 facade。逐 task：3.1 建 scene 包（迁类型+accessor+writers 接口+WriteJSON，处理 export 可见性、DAG 断言 scene⊥rifx）→ 3.2 建 serializer 包（迁 parse/lower/write/back/mutate，核 init/global-var 顺序）→ 3.3 收 aep 为 facade（Open/New* re-export + 类型别名）。**关键 P3 待解**：OM/RQ/PropertyGroup 的 concrete back 跨包后需统一解耦（接口 or 导出 accessor，Task 3.1 Step 2）；结构性 op（AddItem/RemoveItem/InsertKeyframe/DeleteKeyframe/SetDimensionsSeparated/New*/Delete* 等）的 method-vs-free-function + public API 保全（spec 开放题）。每 task 独立 commit + byte-identical + DAG 断言。
2. **P4** 下游切换 + 退役 AST 守卫（scene⊥rifx 改编译期保证）+ CLAUDE.md 多包描述 + 双版本 ship-gate 终验。
3. docgen 次要 follow-on（非阻塞）：README 英文化 + 类型级 / package-func Example 渲染。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
