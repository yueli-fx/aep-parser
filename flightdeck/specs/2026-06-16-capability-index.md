---
status: active
summary: 源码内 aep:cap 结构化 tag → cmd/capindex 自动生成可秒查的能力+API 索引(JSON+人读表) + version(minver) 支持,CI 防漂移;取代手写易过期的 coverage.md
last_updated: 2026-06-16
graduate: true
---

# 源码驱动的能力/API 索引地基 (capindex)

## 背景 / 动机

项目知识库目前靠手写 `plans/coverage.md`(416 行)+ 散落 incident 描述能力边界。问题:**知识与代码两张皮、系统性滞后**(已有记忆 [board-status-drift]),用起来既不准也查不快。2026-06-16 的火焰会话是活样本——agent 全程不知道库到底能做什么(gradient alpha / blend mode / `AnimateEffectParam` 其实都有,却一路靠现 grep + 瞎猜)。

用户的两条硬目标:
1. **不再出现过期知识**——单一真相源 = 代码,自动生成,机器防漂移。
2. **能力可秒取**——问"我想用 xxx",一次查询即得"能不能做 / 怎么调 / 验到几级 / 哪个 gate / 边界坑 / 需要哪版 AE",而非翻文档。

本 spec 只做**地基**:源码内结构化 tag → 自动生成可查索引 + CI 防漂。其余工作流(incidents 全量审核 / CLAUDE.md 审核 / coverage.md 退役 / showcase 升级)挂在地基之上,各起独立 spec。

## 目标 / 非目标

**目标**
- 每个 `internal/aep` 导出符号携带源码内 `aep:cap` 结构化 tag(状态元信息的单一真相源)。
- `cmd/capindex` 从源码自动抽 API 表面 + 交叉 ship-gate 测试,生成 `docs/capabilities.json`(机读可查)+ `docs/capabilities.md`(人读表)。
- CI 强制:(a) 每个导出符号有 tag;(b) tag 声明的 gate 测试真实存在;(c) 生成物 == 已提交(防漂);(d) 防假绿——`verify≥ae-accept` 必须有真实 gate。
- 支持版本元信息(`minver`),将来开放"AE 2024+ 才能用"的函数一标即显。
- 一个查询路径(`go run ./cmd/capindex -q <term>` + 直接 grep JSON),让 agent 秒答能力问题。

**非目标(本 spec 不做)**
- 不重写 incidents / CLAUDE.md / showcase(各自独立 spec,但本 spec 的 `incident=` 链接核验为它们的审核埋点)。
- 不立即退役 coverage.md(P2 全量标注完成后另起 spec 退役,避免真相源真空期)。
- 不改任何运行时代码语义——纯文档/工具/CI 增量。

## 数据模型: `aep:cap` tag

每个导出符号 doc comment 末尾加一行(或多行折行)机读 tag。doc comment 仍是 docgen 的文档源(英文为源),`aep:cap` 是其尾部机读附注。

```go
// AddEffect appends a built-in effect to the layer's Effect Parade and returns
// the parsed *Effect (ready for SetEffectParam).
//
// aep:cap domain=effect tier=stable verify=render-pixel minver=2020
//   gate=TestAddEffect_AEShipGate_AE2020,TestAddEffect_AEShipGate_AE2025
//   boundary="param 控制走 SetEffectParam;per-effect typed helper 未做"
//   incident=add-effect-splice-re  alias="effect,特效,blur,glow"
func AddEffect(layer *Layer, effectMatchName string) (*Effect, error) { ... }
```

**字段**(`key=value`,value 含空格用双引号;列表用逗号):

| 字段 | 必填 | 取值 | 含义 |
|---|---|---|---|
| `domain` | ✓ | layer-create · layer-set · shape · gradient · keyframe · effect · text · mask · comp · project · render-queue · eg · expr · io · structural · meta | 分类桶(分组 + 过滤) |
| `tier` | ✓ | stable · alpha · planned · missing · negative | 成熟度。negative=已证不可达/negative-finding;missing/planned=缺口占位(允许挂在"占位符符号"或 domain 级条目上) |
| `verify` | ✓ | none · roundtrip · ae-accept · render-pixel | 已达到的最高验证级别(对齐交付准则四级) |
| `minver` | — | 2020(默认,可省) · 2022 · 2024 · 2025 … | 该能力需要 AE ≥ 此版本才能用 |
| `gate` | 条件 | 测试函数名,逗号分隔 | ship-gate 测试。`verify` 为 ae-accept/render-pixel 时必填且必须存在 |
| `boundary` | — | 自由文本 | 已知规模/组合边界、未做子项 |
| `incident` | — | incident slug,逗号分隔 | gotcha 来源,须解析到 `incidents/<slug>.md` |
| `alias` | — | 关键词,逗号分隔(含中文) | 搜索别名,服务"秒取" |

`minver` 与 `gate` 正交:`minver`=功能要求的 AE 版本;`gate`=我们已验证过的 AE 版本(测试名里编码)。

**tier 与 verify 是两条正交轴**(2026-06-16 P2 wave2 压测修正,原 `stable⟹ae-accept` 耦合规则作废):
- `tier` = **API 成熟度**(签名是否锁定):`stable`=锁定的公共 API(CLAUDE.md 核心 R/W + 已收口结构性 op)· `alpha`=可能改/删 · `planned/missing/negative`=占位/缺口/不可达。
- `verify` = **证据级别**(验到几级):`none`<`roundtrip`<`ae-accept`<`render-pixel`。
- 二者独立:length-preserving 核心 setter(如 `SetOpacity`)= **stable + roundtrip**(API 锁定但无专门 AE gate,改字节不改 size 故低风险);text animator = **alpha + render-pixel**(已 render-gated 但 accessor 未接、API 可能改)。

**CI 校验规则**(防假绿在 verify⟹gate,不在 tier 耦合):
- `stable`/`alpha` ⟹ `verify ∈ {roundtrip, ae-accept, render-pixel}`(至少 round-trip;`verify=none` 仅 planned/missing/negative 或 `domain=meta`)。
- `verify ∈ {ae-accept, render-pixel}` ⟹ `gate` 非空(声明 AE 验证必须给出 gate 测试名)。
- `planned/missing/negative` ⟹ `verify=none` 且无 gate。
- `domain=meta`(getter/reader/alias/枚举/plumbing)⟹ `verify ∈ {none, roundtrip}`、无 gate(读类无"AE 接受"语义)。
- **渲染类能力(颜色/opacity/可见效果)的 stable 声明仍须 render-pixel gate**(红线4);此约束按 domain 语义人工把关,不由 tier 机械强制。

## 生成器 `cmd/capindex`

- **复用 docgen 的 go/doc 抽取**(同库已有 `cmd/docgen` 从 facade doc comment 生成 docs;capindex 同样走 go/ast/go/doc 解 `internal/aep`)。
- 自动抽:符号名、kind(func/method/type/const)、签名、doc 摘要(去掉 `aep:cap` 行)、关联 `Example*`(docgen 已有此映射,复用)。
- 解析 `aep:cap` tag → 结构体。
- **自动交叉 ship-gate**:扫 `internal/**/*_test.go`,建测试函数名集合;校验每个 `gate=` 名真实存在、且函数体内无顶层 `t.Skip`(skipped 的 gate 视为未验证 → 报错或降级)。
- 解析失败/校验失败 → 非零退出(CI 卡)。

## 输出产物

- `docs/capabilities.json` — 数组,每条 `{symbol, kind, domain, tier, verify, minver, gate[], signature, summary, example, boundary, incident[], alias[]}`。grep/jq 友好。**这是 agent "秒取"的查询目标**。
- `docs/capabilities.md` — 按 domain 分组的人读表,带状态徽章(🟢stable/🟡alpha/⬜planned/❌missing/🚫negative)+ "需 AE ≥ X" 列。头部 `DO NOT EDIT`。
- API 表沿用 docgen 现有 `docs/*.gen.md` + `docs/docs_index.json`(已有 `Type.Name → file/anchor/signature/summary`,见记忆 [docs-index])。capindex 条目按 symbol 与之交叉链接,不重复造 API 表。

## 查询机制("光速答")

- 主路径:`go run ./cmd/capindex -q "<term>"` —— 按 `alias`/`domain`/`symbol`/`summary` 模糊匹配,输出每条:符号、签名、一行示例、tier+verify+minver、gate、boundary。
- 兜底:直接 `grep`/`jq` `docs/capabilities.json`(agent 实际最常用这条)。
- 效果:agent 回答"能不能用 xxx"= 一次查询,不再读 coverage.md。

## 防漂移(CI + go generate)

- `//go:generate go run ./cmd/capindex` 重生成 `docs/capabilities.{json,md}`。
- CI 测试(同 docgen 的 generated-vs-committed 模式,警惕 [docgen-alias-blindspot] 假绿):
  1. 重生成结果 == 已提交(漂移即 fail)。
  2. 每个 `internal/aep` 导出符号都有 `aep:cap` tag(漏标即 fail;trivial getter/alias 可走显式 `aep:cap domain=meta tier=...` 或白名单,但默认强制)。
  3. 每个 `gate=` 测试存在且未 skip。
  4. tier×verify 一致性 + `incident=` slug 可解析。

## 分阶段

- **P1 地基**(本 spec 核心):tag schema + `cmd/capindex` + 4 项 CI 校验 + `-q` 查询,先在 `layer-create` domain(`NewShapeLayer/NewSolidLayer/NewNullLayer/NewAdjustmentLayer/NewCameraLayer/NewLightLayer/NewPrecompLayer/NewTextLayer` 等,全 ship-gated、tier 清晰)端到端跑通。产出可查、CI 绿。
- **P2 全量标注**:给所有 facade 导出符号打 tag,**逐个对着 ship-gate 测试核实** ← 这步本身就是"以源码为准的全量能力审核",会揪出假绿与缺口(可顺手标 `tier=missing/planned` 占位)。完成即开 CI 强制 tag 全覆盖。
- **P3+(独立 spec)**:incidents 全量审核(靠本 spec 的 `incident=` 反向链接核 + 过时清理)· CLAUDE.md 审核 · coverage.md 退役 · showcase 占位符/积攒阈值/review 提醒升级。

## 与现有 docgen / coverage.md 的关系

- docgen 不动:继续管 API 文档(签名 + doc 正文)。capindex 是其上的**状态/能力层**,共用同一批符号与 go/doc 基础设施。
- coverage.md:P2 完成前保持 active 作为过渡真相源;P2 后独立 spec 退役(内容已被 tag 吸收)。

## 验收标准

- [ ] `aep:cap` tag schema 文档化(本 spec 即契约,graduate)。
- [ ] `cmd/capindex` 跑通,`layer-create` domain 全标注,生成 `docs/capabilities.{json,md}`。
- [ ] 4 项 CI 校验测试存在且绿。
- [ ] `go run ./cmd/capindex -q "solid"` 秒出 `NewSolidLayer` 全卡。
- [ ] `go vet ./... && go test ./...` 绿。

## 风险 / 未决

- **标注工作量大**:P2 给数百符号打 tag 是大头(但它=用户要的审核,有价值)。地基(P1)先小范围证流水线。
- **"导出符号"边界**:facade 有大量 typed alias / 枚举常量,是否每个都要独立 tag,还是 alias 类走 `domain=meta` 批量。P1 落地时定白名单策略。
- **gate→symbol 映射**:靠 tag 显式声明(`gate=`),不靠启发式猜 → 准确但需人填;CI 只验"声明的 gate 存在",不验"该 symbol 真被该 gate 覆盖"(后者需更深 AST,暂不做)。
