# Cockpit — aep-parser

**Last updated**: 2026-06-18 by claude（**「理解一个工程」研究 arc：效果字典 + 三轴本体冻结 + 理解自动化**。① `effects-dict` 基础设施:5 版本英文效果字典(matchName→名+类型+**默认值**)+ 可重生成 dumper(`scripts/dump_effects_dict.*`)+ checklist。② **`technique-ontology` spec 冻结(graduate)**:角色/技法/机制三轴本体 + schema v2,投大规模工程前的数据契约,经火焰/闪电/控制器三类压测。③ `fx-techniques` 14 技法重构到三轴(glow 跨现象复用实证)。④ **aepdissect 消费字典**:自动译参数 + 标非默认 = 配方信号(`✎ Glow Threshold=139 (default 153)` + `→ tuned:` 摘要)。核心洞察:.aep 用 **elision**(只存非默认)→ 字节即配方信号、免费降噪。火焰 v3 早先已用户验收 + 双版本 gate 过。）

**Active focus**: **技法内化框架**(`specs/2026-06-18-fx-technique-internalization.md`,procedural-fx-generator 的配方提取引擎通用化)。目标:用户给参考 .aep → 我系统内化成可复用、可迁移到风/雨/雷电/转场的**技法**(非现象专属)。核心洞察:**资产是「技法」不是「现象」**——每现象 = 跨域技法的组合;技法库越长新现象越便宜。机制 = 5 步流水线 + flightdeck 两层结构(技法原子 + 现象配方),preflight 自动载入 = "内化"。**火焰是首个实证**:流水线跑通(解析真实多插件工程→角色词典→存),v2 坐实「层次感=多层合成不是参数」。**下一步**:把火焰专属升级成通用(① `cmd/aepdissect` 工具化解析 ② 抽技法原子建 fx-techniques 首批 ③ 火焰 v3 多层合成验证 additive-depth 技法 ④ 喂第二现象验复用率)。**不变量**:知识单一家 = flightdeck + CLAUDE.md;能力真相源 = capindex(`go run ./cmd/capindex -q <词>`);每渲染类双版本 AE ship-gate(红线4);火焰=番外(`incidents/procedural-fx-over-vector.md` Case2)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-18-fx-technique-internalization.md](specs/2026-06-18-fx-technique-internalization.md) — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techniques + 现象配方 checklists/build-X)→ AE gate 验证。火焰=实例#1已验证。适配风/雨/雷电/转场=复用技法+增量。是 procedural-fx-generator 的「配方提取」引擎的通用化。
- [2026-06-18-procedural-fx-generator.md](specs/2026-06-18-procedural-fx-generator.md) — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
- [2026-06-18-roundtrip-ae-accept-residual.md](specs/2026-06-18-roundtrip-ae-accept-residual.md) — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
- [2026-06-18-technique-ontology.md](specs/2026-06-18-technique-ontology.md) — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized)+ 工程画像 schema + 受控词表 v0。schema 闭 instance 开,约束所有工程理解的产出形式,使跨工程可聚合出通用技巧。经火焰(生成型)/闪电(素材装配型)/控制器(lib-blocked)三类压测。
<!-- /AUTO -->

## 下一步

**「理解一个工程」研究 arc**(契约 `specs/2026-06-18-technique-ontology.md`)。用户计划:3 天研究此课题 → 之后投大量参考工程。Day1(schema 冻结)+ Day2(技法对齐 + 理解自动化)已完成。

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
