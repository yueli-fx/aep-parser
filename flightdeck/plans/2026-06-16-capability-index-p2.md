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

## Wave 0 — 工具扩展(capindex 多包抽取 + 公共面门禁 + meta 通道)

**Files:** `cmd/capindex/extract.go`(改)· `cmd/capindex/surface.go`(新,读 docgen.json 定公共面)· `cmd/capindex/tag.go`(改 validateCap 容 meta)· `cmd/capindex/main.go`(改 pkgDir→pkgDirs)· 各 `*_test.go`

- [ ] **0.1 多包抽取**:`extractEntries` 由单 dir 改收多 dir(facade 优先),复用 docgen `loadPackage` 模式(`parser.ParseDir` 每包 + 原始 doc comment 副本保 `//aep:cap`)。合并:facade 自由函数优先,scene/codec 方法按 receiver 类型归集。
  - 测试先行:`SetOpacity`(scene `*Layer` 方法)被抽出、recv=Layer、能解析其 tag。
- [ ] **0.2 公共面门禁** `surface.go`:读 `docs/docgen.json` → 公共面 = 各 file 的 `roots` 类型的全部导出方法 ∪ `funcs` 列表 ∪ facade/aliases/facade_codec 的全部导出 func。**只有公共面符号"必须有 tag"**;非 root 的 scene 方法/内部 helper 不要求。
  - 测试:`Layer` 方法属公共面;某非 root scene 方法(如 writer 接口实现)不属。
- [ ] **0.3 meta 通道**:`validateCap` 容 `domain=meta`(tier=stable|alpha,verify=none|roundtrip,无 gate 不报错)。
  - 测试:`domain=meta tier=stable verify=none` 校验通过;`domain=layer-set tier=stable verify=roundtrip` 仍按原规则报错(stable⟹ae-accept|render-pixel)。
- [ ] **0.4 进度可见**:加 `-coverage` flag(或测试)报告"公共面 N 符号 / 已标 M / 未标 K",P2 期间作进度仪表(**暂不**作 CI 强制,Wave 9 才 flip)。
- [ ] **0.5** 重生成 + `go test ./cmd/capindex/` 绿 + 全套绿 + commit `feat(capindex): P2 wave0 — 多包抽取 + 公共面门禁 + meta 通道`。

## Wave 1 — facade 自由函数收尾(facade.go 71 + aliases 22 + codec 3)

P1 已标 8 个 `New*Layer`。本 wave 标完 facade 三文件剩余自由函数,domain 分布:`structural`(Delete/Duplicate/Insert/Move*Layer、Remove/Move/DuplicatePropertyGroup、AddMarker/RemoveMarker)·`keyframe`(InsertKeyframe/DeleteKeyframe/SetDimensionsSeparated)·`effect`(AddEffect/RemoveEffect/SupportedEffects/SetEffectParam/Animate*)·`text`(全部 `AddText*Animator`/`AddTextRangeSelector`/`AddTextWigglySelector`/`SetTextRangeAdvanced`/`AnimateText*`)·`shape`(aliases 的 `New*Node`/`WrapShapeLayer`/`NewVectorGroup`)·`gradient`(codec `ParseGradientXML`/`EncodeGradientXML`)·`meta`(`Capabilities`/`NewPropertyStream`/`Reopen`/`NewProject`)。

- [ ] 逐个按 recipe 标注 + 对 ship-gate 核实(text/shape/effect 多已有双版本 gate,见 `incidents/text-animator-create-re.md`、`trim-paths-vector-filter-re.md`)。
- [ ] 重生成 + 测试 + commit。

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

## 验收标准

- [ ] capindex 抽取覆盖 facade+scene+codec 三包;`go run ./cmd/capindex -q "opacity"` 秒出 `SetOpacity`(scene 方法)全卡。
- [ ] 公共面每符号有合法 tag,`TestPublicSurfaceFullyTagged` 绿。
- [ ] 每 `gate=` 真实存在且非 skip;tier×verify 一致;`incident=` slug 可解析。
- [ ] `docs/capabilities.{json,md}` == 重生成(drift guard 绿)。
- [ ] `go vet ./... && go test ./...` 全绿。
- [ ] 审核副产物:揪出的假绿 / 缺口逐条标 `tier=missing/planned/negative` + boundary(不假装"能用")。

## 风险 / 注记

- **体量大**:400–600 符号、多 wave、跨多 context window。plan 即追踪载体,逐 wave commit 让进度可续。
- **coverage.md 仅作提示**:其 done/deferred 标记系统性滞后(board-status-drift),核实一律以代码 + ship-gate test 为准。
- **scene 源码改动面广**:大量 scene 方法加 `aep:cap` doc 行——符合"导出符号 doc comment 是文档载体"约定,顺带补全 docgen 输出。
- **getter 取舍**:若用户后续要轻量化,可把纯 getter 留 meta、只精标写/做面;当前按"全量"推进(用户明确要完整能力表)。
