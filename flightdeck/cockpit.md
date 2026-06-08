# Cockpit — aep-parser

**Last updated**: 2026-06-09 by claude
**Active focus**: **V3 M8 方案②（真·物理分包）执行中** — plan `plans/2026-06-07-v3-m8-physical-split-plan.md`（active）。P0 基线+inventory ✅ · P1 抽 `internal/codec` ✅ · **P2 back-ref 接口化 7/10**（已倒置 Composition / Marker / Mask / Footage / Keyframe / Project / Layer；余 PropertyGroup / RenderQueue / Property，Property 最后）。剩 3 类都带设计待决（见 inventory §E U1/U4）。每 commit 绿 + byte-identical round-trip。

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

**§E 设计待决已审定**（inventory §F，用户批准；commit 0383642）：U1 RQ/OM/Guide settings → scene 独占 copy 单一真相源 + WriteAEP 单点 sync（同构现有 syncShapeLayerChunks）+ 类型化 offset；U2 Guide 无接口；U3 结构性 op 住 serializer；U4 砍 PropertyGroupWriter；U5 公共签名不变、内部 primitive 返 bool。**净结果：实质待倒置 = RenderQueue 子系统 + Property；OM/Guide/PropertyGroup 无需接口。**

1. **RenderQueue 倒置**（下一步，最重，唯一引入 write-时同步新机制）。已读全 RQ 写面，路径明确：
   - **setter 体不改**（`patchU16`/`omPatch*` 已写 `settingsBlock`/`roouData`，字段从别名变 copy 后自动成「改 scene copy」；已同步解码字段 `it.RenderSettings.X=v`，getter 不受影响）。
   - **parse**：`item.settingsBlock`/`om.settingsBlock`/`om.roouData`/`guide.block` 从 alias 改 `append([]byte(nil),...)` copy（单一真相源）。
   - **WriteAEP 加 `syncRenderQueue` 步骤**：serializer 持原别名 slice（`renderQueueItemBackrefs` 增 `settingsSlice`；新建 `outputModuleBackrefs` 持 settings+roou slice，OM 加 back；guide 经 comp 序列化路径按 index 配对），scene copy 单点拷回 chunk。byte-identical：unmutated 时 copy==原字节。
   - **类型化 offset**：`type RenderSettingOffset int` + 现成 `codec.Rs*`/`Oms*`/`Rouo*` 常量改此类型。
   - `SetComment`（唯一 A1 length-variable）→ `RenderQueueItemWriter{SetComment}`；结构性 AddItem/RemoveItem 暂 stopgap（P3 迁 serializer）。
2. **Property**（RQ 之后）：4 setter + InsertKeyframe/DeleteKeyframe → PropertyWriter；`SetDimensionsSeparated` 经 propertyBackrefs 增持父 group chunk 自行 splice。
3. **PropertyGroup**：仅核查无 scene→rifx 残留（无接口）。
   每步独立 commit + byte-identical + setter 单测 + attach 断言 + 我独立复核。
2. P2 收口（Task 2.3：scene_* rifx 引用清零核查）→ **P3** git mv 物理分包（编译期硬边界）→ **P4** 下游切换 + 退役 AST 守卫 + CLAUDE.md + 双版本 ship-gate。
3. docgen 次要 follow-on（非阻塞）：README 英文化 + 类型级 / package-func Example 渲染。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
