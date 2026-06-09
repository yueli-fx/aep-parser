# Cockpit — aep-parser

**Last updated**: 2026-06-09 by claude
**Active focus**: **V3 M8 方案②（真·物理分包）执行中** — plan `plans/2026-06-07-v3-m8-physical-split-plan.md`（active）。P0 基线+inventory ✅ · P1 抽 `internal/codec` ✅ · **P2 back-ref 接口化 ✅ 完成**（9 类倒置为接口：Composition/Marker/Mask/Footage/Keyframe/Project/Layer/RenderQueueItem/**Property**；OM/RenderQueue 容器/PropertyGroup 按 §F 故意保 concrete——无 writer 接口，P3 统一解耦）。**Task 2.3 收口 ✅：全 `scene_*.go` 零 rifx import/code-token**。**P3.0 结构性 op method→free function ✅ 完成**（7 组 7 commit，全绿 + byte-identical + docgen 重生成）。**下一步 = P3.1 git mv 物理分包**。每 commit 绿 + byte-identical round-trip。

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

**P3 关键决议**（2026-06-09 用户拍板）：「mechanical git mv」假设证伪——结构性 op 是 scene 方法但需 serializer 访问 backref（跨包=import 环），且多为 Stable ship-gated，spec §2.4/§F D-U3 的「改 free function」会破 Stable 签名。用户**批准破原 #2 约束 + 写新约束**（commit 2ef731e：CLAUDE.md #2 新增「结构性 op 语义稳定、调用形态可随分包改 facade 自由函数，标 BREAKING 不算违约」）。→ 采 **Path B**（结构性 op → serializer 自由函数 + facade re-export）。

1. **P3.0：结构性 op method→free function ✅ 完成**（2026-06-09，单包内转，每组独立 commit，BREAKING + docgen 重生成，全程绿 + byte-identical round-trip）。7 组 landed：
   - ✅ `RenderQueue.{AddItem,RemoveItem}`（c6eb51b，模式验证）。
   - ✅ Composition layer-list `{DeleteLayer,MoveLayer,InsertLayer,DuplicateLayer,NewShapeLayer}`（5915d51）。
   - ✅ Layer 自定位 `{MoveToBeginning,MoveToEnd,MoveAfter,MoveBefore}`（bd8beee；委托 MoveLayer=serializer，故必转）。
   - ✅ Composition 生命周期 `{NewComposition,DuplicateComposition}`（14f29ac）。
   - ✅ Marker `{AddMarker, Remove→RemoveMarker}`（75e1939）。
   - ✅ Property 结构性 `{InsertKeyframe,DeleteKeyframe,SetDimensionsSeparated}` + `AEPropertyGroup.{Remove→RemovePropertyGroup, MoveTo→MovePropertyGroup, Duplicate→DuplicatePropertyGroup}`（c666f30）。
   - **判据（本次确立，比原「结构性就转」更精准）**：方法→serializer 自由函数 **当且仅当它调 serializer-only 代码**（结构性自由函数 / 建 chunk / 碰 concrete backref）。scene→scene 纯委托者**留方法**（scene 内合法）。
   - **偏离 1**：`Layer.{ReplaceSource,RemoveTrackMatte,ClearTrackMatteLayer}`（原 待转 列含之）**未转**——三者纯委托到 Set*（scene 方法，经 writer 接口，不碰 concrete backref）→ 留方法。`ClearTrackMatteLayer` 虽在 write_layer.go，但与同文件 Set* 方法一样 P3.1 随迁 scene。转之纯属多余 BREAKING。
   - **偏离 2**：`AEPropertyGroup.Duplicate` 原 待转 列**漏列**，与 Remove/MoveTo 同族（INDEXED_GROUP chunk-pair splice）→ 本次补转，否则 P3.1 git-mv `mutate_property_structural.go` 会断。
   - **完成验证**：scene_*.go 零调用任何结构性自由函数（grep 证）+ arch boundary guard（scene⊥rifx、codec⊥scene）绿 + `go vet ./...` + `go test ./...` 全绿。纯图构造（VectorGroup.AddRect 等 detached）留 scene。
2. **P3.1-3.3（← 下一步）** git mv 物理拆包：`scene_*.go`→`internal/scene`、`{parse_,lower_,write_,back_,mutate_}*.go`→`internal/serializer`、`internal/aep` 薄 facade（类型别名 + Open/FromReader + 全部结构性自由函数 re-export）。处理 export 可见性（AttachWriter plumbing 导出）、init/global-var 顺序、DAG 断言。OM/RQ/PropertyGroup concrete back 跨包后随 mutate_/back_ 迁 serializer（同包，无需接口）。**关键修正（P3.0 副产物）**：`write_*.go` 等 serializer-bound 文件里**留存的 scene 类型方法**（全部 `Set*` core R/W + 偏离 1 的 `ReplaceSource/RemoveTrackMatte/ClearTrackMatteLayer` 委托者 + Layer 的 `locateInComp/locatePair` 等 unexported helper）**不能随文件 git-mv 进 serializer**（方法不能跨包定义在 scene 类型上——即驱动 Path B 的同一 Go 墙）→ 这些方法须**先析出到 scene_ 文件**，仅 chunk-building/concrete-backref 函数体（P3.0 已转成的自由函数）进 serializer。故 write_*.go 的 git-mv 非整体搬，需按「方法 vs 自由函数/接口实现」二分。
3. **P4** 下游切换 + 退役 AST 守卫（scene⊥rifx 改编译期保证）+ CLAUDE.md #3 多包描述 + 双版本 ship-gate 终验。
4. docgen 次要 follow-on（非阻塞）。

## Backlog（单条候选）

1. **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持（V2.3）。**当前最大候选**。
2. 泛型 `DuplicateItem`（无 scripting API）· `ImportComposition`（需求驱动）· **Property synthesis**（暂搁大 feature，详 `incidents/transform-group-default-omission.md`）。

其余 deferred R-only（DisplayColorSpace / ValueText 等）+ shape 次要子属性见 `specs/deferred-backlog.md` + `plans/coverage.md`。

## Hanging tasks

无。
