# Cockpit — aep-parser

**Last updated**: 2026-06-18 by claude（**「理解一个工程」研究 arc：效果字典 + 三轴本体冻结 + 理解自动化**。① `effects-dict` 基础设施:5 版本英文效果字典(matchName→名+类型+**默认值**)+ 可重生成 dumper(`scripts/dump_effects_dict.*`)+ checklist。② **`technique-ontology` spec 冻结(graduate)**:角色/技法/机制三轴本体 + schema v2,投大规模工程前的数据契约,经火焰/闪电/控制器三类压测。③ `fx-techniques` 14 技法重构到三轴(glow 跨现象复用实证)。④ **aepdissect 消费字典**:自动译参数 + 标非默认 = 配方信号(`✎ Glow Threshold=139 (default 153)` + `→ tuned:` 摘要)。核心洞察:.aep 用 **elision**(只存非默认)→ 字节即配方信号、免费降噪。火焰 v3 早先已用户验收 + 双版本 gate 过。）

**Active focus**: **Booyah Glitch 全工程复刻 = 理解金标准检验**(spec `2026-06-19-booyah-glitch-full-replication.md` + plan `2026-06-19-booyah-glitch-replication.md`)。把整个真实工程(12 comp/61 层/~100 mask/wiggle+Evolution 表达式/Curves)用咱们 Go API 从零重建——round-trip 只证读得回字节,**从零重建证真懂每个 chunk 怎么来的**。用户铁律:**不逃避未知字段**(撞墙 RE,物理不可写给实证负结论)。判据 = 结构保真(AE 接受 + DOM 读回值对账)+ 终帧渲染像素对照;范围 = 全 12 comp,DAG 叶→根逐 comp 验;方法 = 手写 per-comp 生成器、**原工程只读当取值神谕、chunk 全由 API 重建(绝不 copy 字节)**。是上游「技法内化框架」(`specs/2026-06-18-fx-technique-internalization.md`)的「复现判据 = 理解金标准」检验项,bar 拉到字节/工程级。**不变量**:知识单一家 = flightdeck + CLAUDE.md;能力真相源 = capindex(`go run ./cmd/capindex -q <词>`);每渲染类双版本 AE ship-gate(红线4)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-18-fx-technique-internalization.md](specs/2026-06-18-fx-technique-internalization.md) — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techniques + 现象配方 checklists/build-X)→ AE gate 验证。火焰=实例#1已验证。适配风/雨/雷电/转场=复用技法+增量。是 procedural-fx-generator 的「配方提取」引擎的通用化。
- [2026-06-18-procedural-fx-generator.md](specs/2026-06-18-procedural-fx-generator.md) — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
- [2026-06-18-roundtrip-ae-accept-residual.md](specs/2026-06-18-roundtrip-ae-accept-residual.md) — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
- [2026-06-18-technique-ontology.md](specs/2026-06-18-technique-ontology.md) — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized)+ 工程画像 schema + 受控词表 v0。schema 闭 instance 开,约束所有工程理解的产出形式,使跨工程可聚合出通用技巧。经火焰(生成型)/闪电(素材装配型)/控制器(lib-blocked)三类压测。
- [2026-06-19-booyah-glitch-full-replication.md](specs/2026-06-19-booyah-glitch-full-replication.md) — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle 表达式/Curves)用咱们的 Go API 从零完整复刻,作为理解金标准检验。判据=结构保真(AE 接受 + DOM 读回值对账)+ 终帧渲染像素对照;范围=全工程 DAG 叶子优先逐 comp 验;方法=手写 per-comp 生成器、原工程仅当取值神谕、chunk 全由 API 重建(绝不 copy 字节)。
- [2026-06-19-booyah-glitch-replication.md](plans/2026-06-19-booyah-glitch-replication.md) — 实现 Booyah Glitch 全工程复刻:Phase0 前置 spike(表达式 Evolution=time*N + wiggle / 单层 ~21 mask / Curves 曲线数据)→ Phase1-4 按 DAG 叶→根逐 comp 建(extract 值→写 gen_<comp>.go→AE 接受+verify.jsx DOM对账→覆盖账本→commit)→ 终帧 メインコンプ 像素对照原工程。
<!-- /AUTO -->

## 下一步

**Booyah Glitch 全工程复刻 arc — Phase 0 进行中**(plan `2026-06-19-booyah-glitch-replication.md`)。
- ✅ **Task 0.1 scaffold**(commit 04e57eb):`flightdeck/showcase/booyah-clone/` 包建好——`oracle.go` 只读打开原工程当取值神谕、`main.go` 跑通列出 12 comp、`render.jsx`/`verify.jsx` 起点、`INDEX.md` 覆盖账本(✅/🔧/⛔/⬜ 逐 comp + 逐未知特征)。
- ⬜ **Task 0.2 oracle 读侧审计**(轻):确认 mask bezier / keyframe 值 / effect params 全可读;读不出的(预期 Curves)登记账本 `🔧待RE(read)`。读侧已证基本可读。
- ⬜ **Task 0.3 表达式 spike(头号 de-risk)**:Go 写 Evolution=`time*N` + `wiggle(34,0.29)` → AE 双版本 ship-gate 验是否真求值(`expression-enable-byte-pair` 高危)。go/no-go 决定 ③⑩⑪ 怎么建;FAIL 则降级 keyframe 近似(用户已批准)。
- ⬜ **Task 0.4 many-mask(21)spike** · **Task 0.5 Curves spike** → 然后 Phase 1-4 按 DAG 逐 comp 建。

依据:plan + spec(`specs/2026-06-19-booyah-glitch-full-replication.md`)+ `tmp/booyah_dissect.txt`(完整画像,真相源)+ `cmd/aepdissect`。

---

**上游「理解一个工程」研究 arc(暂让位,本复刻即其「复现判据=理解金标准」检验项)**(契约 `specs/2026-06-18-technique-ontology.md`)。Day1(schema 冻结)+ Day2(技法对齐 + 理解自动化)已完成。

已完成:
- ✅ **effects-dict 基础设施**:5 版本英文字典 `data/effects-dict/` + dumper `scripts/dump_effects_dict.{jsx,ps1}` + 种子。流程/坑全记 `checklists/effects-dict.md`(切语言、脚本写权限只能人工开、LUT skip)。
- ✅ **technique-ontology spec 冻结**(graduate):三轴本体 + schema v2 + 工程画像 schema + 受控词表 v0。
- ✅ **fx-techniques 三轴重构**:14 技法对齐 schema;emissive-glow 跨 fire+lightning 实证(首个跨现象复用)。
- ✅ **aepdissect 消费字典**:自动产配方信号(译参数 + `✎`标非默认 + `→ tuned:`摘要)。

剩(Day2/3,投工程前的增强):
- ✅ **时间轴维度**(done,commit 59f8d83):aepdissect 读关键帧出活跃区间 act=[in→out] + 动画属性(transform 树遍历)+ 缓动分类(linear/ease/bezier/hold)+ 错位提示。揭示 bondbond ED:Line 层 3.9s 入场描边生长、Cam 运镜。⚠ 已知:kf span 是层/源相对时间,未归一化到合成时间。
- ⬜ **aepdissect 余下画像维度**:依赖图边(源/父/表达式/matte 整合)→ 结构化画像输出(当前人读文本,未出 yaml/json profile)。
- ⬜ **复现判据验证**(理解金标准):挑简单工程从画像反向生成,渲染对照测理解度。
- ⬜ 投工程时校准受控词表(角色/技法粒度、词表生长)。

依据:`specs/2026-06-18-technique-ontology.md`(冻结契约)· `specs/2026-06-18-fx-technique-internalization.md`(上游流水线)· `docs/fx-techniques.md`(技法库)· `cmd/aepdissect`(PARSE+配方信号)· `checklists/effects-dict.md`(字典 dump/用)。

**次要(需求驱动)**:程序化闪电(T14 未验证)· 残值 backlog(剩 53 roundtrip,`specs/2026-06-18-roundtrip-ae-accept-residual.md`)· 能力真相源 `go run ./cmd/capindex -q <词>`。

## Backlog

- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
