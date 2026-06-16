---
status: active
summary: 扩 capindex 抽取到 docgen 三包(facade+scene+codec),给全量公共能力符号(Set*/getter/结构性/Add*)打 aep:cap tag、逐个对 ship-gate 交叉核实(揪假绿/标缺口),最后开 CI 全覆盖强制
last_updated: 2026-06-16
implements: specs/2026-06-16-capability-index.md
---

# capindex P2:全量能力标注审核

> **执行约定**:子任务用 `- [ ]` 勾选追踪。每个标注 wave 收尾必跑 `go run ./cmd/capindex`(重生成)+ `go test ./cmd/capindex/`(校验)+ `go vet ./... && go test ./...`(全套)+ commit。**一律用中文**汇报。

**Goal**:把项目能力知识从手写 coverage.md 彻底迁到源码内 `aep:cap` tag——**全量公共能力符号**逐个对 ship-gate 交叉核实后打 tag,`capindex` 生成可秒查的 `docs/capabilities.{json,md}`,CI 强制全覆盖 + 防漂 + 防假绿。这步本身 = 用户要的"以源码为准的全量能力审核"(会揪出假绿、标出缺口)。

**Architecture**:`cmd/capindex` 抽取范围从单包 `internal/aep` 扩到 docgen 的三包集(`internal/aep` facade + `internal/scene` 真类型/方法 + `internal/codec` 值类型),复用 docgen 的多包 `loadPackage`/merge 模式。"哪些符号必须有 tag" = docgen `docgen.json` 策展的公共面(roots 类型的导出方法 + funcs 列表),非三包全部导出符号。

**Tech Stack**:Go `go/ast`/`go/doc`/`go/parser`/`go/printer`(同 docgen);`aep:cap` 单行 doc-comment directive;TDD + 逐 wave commit;CI drift guard(`TestGeneratedUpToDate`)。

---

## 范围修正(对 spec 的关键澄清)

spec 字面写"capindex 解 `internal/aep`",基于"facade 含所有符号"的假设。**实测推翻该假设**:按架构红线 #2/#3,公共 R/W 的主体 `Set*`/getter 是 **scene 类型方法**(住 `internal/scene`,facade 仅靠 type-alias 方法提升,不再定义)。光 `Layer` 就 229 个方法、`Composition` 58、`Project` 52、`Property` 32。"只扫 internal/aep"会漏掉 341 个 `Set*` + 全部 getter——而这恰是"我想用 xxx 功能"主场景。

**修正**:capindex 抽取对齐 docgen 三包集;tag 落在符号定义处(scene 方法的 tag 写进 scene 源码 doc comment,与 docgen 文档源一致)。本修正落地后回写 spec 一行(范围澄清)。

## 公共能力面规模(实测)

| 来源 | 计数 | 性质 |
|---|---|---|
| `internal/aep/facade.go` funcs | 71 | 结构性 op / Add* / Open·Write(真自由函数,P1 已标 8 个 New*) |
| `internal/aep/aliases.go` funcs | 22 | shape node 构造器(`New*Node`)+ `WrapShapeLayer`/`NewVectorGroup`/`Capabilities` |
| `internal/aep/facade_codec.go` funcs | 3 | `ParseGradientXML`/`EncodeGradientXML`/`NewPropertyStream` |
| scene root 类型方法 | Layer 229 · Composition 58 · Project 52 · Property 32 · Footage 15 · Mask 9 · … | Set*/getter/property access(能力主体) |
| 三包 type/const | type 8(facade)+ codec types · const 501(枚举值) | meta lane |

全量公共面约 **400–600 符号**。两条标注通道:
- **能力通道**(domain 专属 + verify/gate):写/做/特性符号(`Set*`/`Add*`/`New*`/结构性/`Animate*`/per-domain feature)。**verify 级别逐个对 ship-gate 核实 = 审核本体**。
- **meta 通道**(`domain=meta`,无 gate):getter/reader、type alias、枚举 const、构造器。批量低开销,满足 CI 全覆盖。

## 每个能力符号的核实配方(audit recipe)

1. 定位符号、定 `domain`。
2. 定 `tier`+`verify`:查是否有 ship-gate 覆盖 → grep gate 测试名;**coverage.md 仅作起点提示、非真相**(红线:board status drift);确认 gate 真实存在且非 `t.Skip`。
3. 渲染类能力(颜色/opacity/可见效果)→ verify 必须 `render-pixel`(红线 #4),否则降 `roundtrip`/标 `alpha`。
4. 写 tag(能力通道全字段;meta 通道极简)。
5. wave 收尾:重生成 + `go test ./cmd/capindex/` + 全套 + commit。

---

## Wave 0 — 工具扩展(capindex 多包抽取 + 公共面门禁 + meta 通道)✅ DONE(commit Wave0)

**Files:** `cmd/capindex/extract.go`· `surface.go`(新)· `tag.go`· `main.go`· 各 `*_test.go`

- [x] **0.1 多包抽取**:`extractEntries(dirs...)` 多包 + facade-priority 去重 + `Entry.Pkg`。测试:`SetOpacity` 抽出(recv=Layer,pkg=scene)。
- [x] **0.2 公共面门禁** `surface.go`:读 `docs/docgen.json` roots/funcs → must-tag = aep funcs ∪ root 类型方法 ∪ docgen funcs。
- [x] **0.3 meta 通道**:`validateCap` 容 `domain=meta`(verify=none|roundtrip、无 gate)+ `validDomains` 闭集校验(揪 domain 拼写错)。
- [x] **0.4 `-coverage` flag**:公共面 536 符号进度仪表(暂不 CI 强制,Wave 9 flip)。
- [x] **0.5** 重生成 + 测试 + 全套绿 + commit。

## Wave 1 — facade 全量自由函数 ✅ DONE(98/536,commit db96367)

facade.go + aliases.go + facade_codec.go + scene_application.go 全部公共自由函数已标 + ship-gate 交叉核实。domain:layer-create 8 / effect 6 / mask 6 / text 21 / shape 17 / gradient 2 / structural 14 / comp 3 / project 1 / eg 1 / render-queue 2 / layer-set 1 / keyframe 2 / meta 15。

**审核发现(归档,供 Wave9 + orphan 决议):**
- **manual-gate orphans(11)**:`DeleteLayer`/`DuplicateLayer`/`InsertLayer`/`MoveLayer`/`MoveToBeginning`/`MoveToEnd`/`MoveAfter`/`MoveBefore`/`DuplicateComposition`/`InsertKeyframe`/`DeleteKeyframe` —— CLAUDE.md/coverage 称 Stable,但**无自动 Go `_AEShipGate` test**(AE 接受系 manual JSX 验证,fixtures gitignored)。按交付准则诚实标 `alpha/verify=roundtrip` + boundary。**待决**:(a) 编码 JSX gate 为 Go test 恢复 ae-accept/stable;(b) 扩 capindex 支持 manual-gate 引用;(c) 维持 roundtrip。
- **read-tier 张力**:read/infra(Open/FromReader/Reopen/Parse/getter)无 "AE 接受"语义 → 归 `meta`(stable+roundtrip);tier×verify 的 `stable⟹ae-accept|render-pixel` 规则仅适用 write/render 域。
- **AnimateTextOpacity / AddTextRotationX·YAnimator** 诚实降 `roundtrip`(opacity leaf 共享已 render-gated 机制但自身无 gate;RotX/Y 2D 视觉惰性 write-only)。

## 策略修正(2026-06-16,用户批准)

- **tier/verify 解耦**(commit 63af412):两条正交轴,`stable+roundtrip` 合法(length-preserving 核心 setter);防假绿守卫在 `verify∈{ae-accept,render-pixel}⟹gate`。详 spec。
- **getter 豁免(write-surface-first)**:root 类型方法**仅写/做动词前缀**(Set/Add/Remove/Delete/Insert/Move/Duplicate/Animate/Replace/Clear/Enable/Disable/Toggle/Apply/Reset/Make)强制 tag;纯 getter/reader/navigation **豁免**(不禁止,以后可 additive 补)。公共面 536→**313**(223 getter 豁免,`-coverage` 报 exempt 数,非静默)。CI 全覆盖(Wave 9)= 写面全覆盖。
- **压测样本**(commit 63af412):SetOpacity/SetName/SetExpression/SetColorManagementSystem(minver=2024)/SetFrameRate/getter —— schema 经全形态压测确认足够。
- **剩余写面 ≈ 210**:Layer 128 · Comp 25 · Project 18 · Marker 10 · Mask 9 · Keyframe 9 · Footage 5 · Property 4 · Guide 2。

## Wave 2 — Layer Set*/getter(229,最大块;子拆)

- [ ] **2a** transform:Position/Scale/Rotation/Opacity/AnchorPoint/SetDimensionsSeparated 相关 — `domain=layer-set`。
- [ ] **2b** 3D/material:3D enable、Orientation/RotateX·Y·Z、material option — `domain=layer-set`(对 3D gate 现状核实,多为 alpha/write-only,见 cockpit backlog)。
- [ ] **2c** time/flags/quality/blend/matte:inPoint/outPoint/startTime、blendingMode、trackMatte、quality、各 flag — `domain=layer-set`。
- [ ] **2d** getter/reader(Name/Width/Index/…)→ `domain=meta`(批量)。
- [ ] **2e** property-access 方法(`Property(path)`/group 导航)→ `domain=keyframe` 或 `meta`。
- [ ] 重生成 + 测试 + commit(可按子拆多次 commit)。

## Wave 3 — Property/Keyframe(Property 32 + 关键帧 API)

- [ ] `Property` 方法 + keyframe writer(AddKeyframe*/SetStaticValue/SetSpatial*/ease)— `domain=keyframe`,对 keyframe 双布局 + 规模 gate 核实(`layoutFor`、from-scratch-mg-roadmap S1)。

## Wave 4 — Composition(58)+ Project(52)

- [ ] Composition Set*/视图/便利 — `domain=comp`。
- [ ] Project Set*/settings/item — `domain=project`。

## Wave 5 — Shape graph nodes(FillNode/StrokeNode/Gradient*/各 primitive node 方法)

- [ ] 各 node 的 `Set*`/stops — `domain=shape`/`gradient`,对 shape 渲染像素 gate 核实(红线 #4:shape 颜色曾假绿)。

## Wave 6 — Text(TextSource/run/paragraph/range 方法)

- [ ] `domain=text`,对 animator/selector 双版本 render-gate 核实;Expressible Selector/Rotation X·Y 等 evidence-defer 标 `alpha`/`negative` + boundary。

## Wave 7 — Effect 类型方法 + Wave 8 — Mask/Footage/RenderQueue/EG/Guide

- [ ] Effect 方法 — `domain=effect`。
- [ ] Mask(`domain=mask`)· Footage(`domain=io`)· RenderQueue(`domain=render-queue`,多 alpha)· EssentialGraphics(`domain=eg`,搁置项标 boundary)· Guide。

## Wave 9 — meta 扫尾 + 开 CI 全覆盖强制

- [ ] type alias / 枚举 const / 剩余 getter 批量 `domain=meta`。
- [ ] `surface.go` 门禁升级:`TestPublicSurfaceFullyTagged` —— 公共面每符号必须有合法 tag(漏标即 fail)。替换 P1 的 `TestLayerCreateFullyTagged`(或并存,后者成子集)。
- [ ] 全套绿 + commit `feat(capindex): P2 done — 公共面全标注 + CI 全覆盖强制`。
- [ ] 回写 spec 范围修正一行;cockpit Next 推进到 `knowledge-consolidation`(coverage.md 退役解锁)。

---

## 验收标准 — ✅ DONE(2026-06-16)

- [x] capindex 抽取覆盖 facade+scene(codec 经 facade 再导出);`-q "opacity"` 秒出 `SetOpacity`(scene 方法)全卡。
- [x] **写/做面 468/468 = 100% 合法 tag**,`TestWriteSurfaceFullyTagged` CI 强制绿(getter/const 按 write-surface-first 豁免)。
- [x] 每 `gate=` 真实存在且非 skip;tier×verify 一致;`incident=` slug 可解析(capindex 校验)。
- [x] `docs/capabilities.{json,md}` == 重生成(drift guard 绿);docgen 文档随新增 doc comment 重生成。
- [x] `go vet ./... && go test ./...` 全绿。
- [x] 审核副产物:tier/verify 解耦 · manual-gate orphans 诚实 stable+roundtrip · render-pixel claim spot-check 实测采样像素 · verify 分布(roundtrip 280 / render-pixel 149 / ae-accept 35 / meta-none 6)。

## 完成纪要

- **8-agent Workflow**(581k token)并行标 scene 写面 + 收尾补漏(text-run 34 + WriteJSON)。
- **写面定义扩面**:docgen-root → "任意 exported 能力类型的写方法"(此前漏 shape 节点全族 / render-queue / OutputModule / text-run)。
- **工具硬化**:`isWriteMethod` CamelCase 边界(排除 Enabled/Added getter 误判);`parseCapTag` 单行契约(防 prose 误析);+Write 动词。
- **遗留 cosmetic**:`scene_render_queue_writers.go` 部分 tag directive-first 放置(功能等价、go/doc 已 strip、测试绿),可后续清理。
- **下一步**:`knowledge-consolidation`——coverage.md 退役(已解锁)/ incidents 审核 / CLAUDE.md 瘦身 / 记忆退役。

## 风险 / 注记

- **体量大**:400–600 符号、多 wave、跨多 context window。plan 即追踪载体,逐 wave commit 让进度可续。
- **coverage.md 仅作提示**:其 done/deferred 标记系统性滞后(board-status-drift),核实一律以代码 + ship-gate test 为准。
- **scene 源码改动面广**:大量 scene 方法加 `aep:cap` doc 行——符合"导出符号 doc comment 是文档载体"约定,顺带补全 docgen 输出。
- **getter 取舍**:若用户后续要轻量化,可把纯 getter 留 meta、只精标写/做面;当前按"全量"推进(用户明确要完整能力表)。
