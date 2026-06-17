# Cockpit — aep-parser

**Last updated**: 2026-06-18 by claude（**火焰 v3 通过 + 技法内化框架地基 + 2 样本实证**。火焰 v3 多层合成**用户真机「过了」+ 双版本 gate 绿(逐像素一致)**——经 v1 被否→v2 单层→v3 多层出层次,showcase 翻 complete、配方升「验证配方」、技法 T3 validated。框架地基:`cmd/aepdissect`(PARSE 工具)+ `docs/fx-techniques.md`(两层技法库,14 技法)。样本 #1 火焰=①程序化、#2 闪电=②素材+装配(aepdissect 即判出 footage+rig),催生**「两类元素」框架概念**。）

**Active focus**: **技法内化框架**(`specs/2026-06-18-fx-technique-internalization.md`,procedural-fx-generator 的配方提取引擎通用化)。目标:用户给参考 .aep → 我系统内化成可复用、可迁移到风/雨/雷电/转场的**技法**(非现象专属)。核心洞察:**资产是「技法」不是「现象」**——每现象 = 跨域技法的组合;技法库越长新现象越便宜。机制 = 5 步流水线 + flightdeck 两层结构(技法原子 + 现象配方),preflight 自动载入 = "内化"。**火焰是首个实证**:流水线跑通(解析真实多插件工程→角色词典→存),v2 坐实「层次感=多层合成不是参数」。**下一步**:把火焰专属升级成通用(① `cmd/aepdissect` 工具化解析 ② 抽技法原子建 fx-techniques 首批 ③ 火焰 v3 多层合成验证 additive-depth 技法 ④ 喂第二现象验复用率)。**不变量**:知识单一家 = flightdeck + CLAUDE.md;能力真相源 = capindex(`go run ./cmd/capindex -q <词>`);每渲染类双版本 AE ship-gate(红线4);火焰=番外(`incidents/procedural-fx-over-vector.md` Case2)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-18-fx-technique-internalization.md](specs/2026-06-18-fx-technique-internalization.md) — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techniques + 现象配方 checklists/build-X)→ AE gate 验证。火焰=实例#1已验证。适配风/雨/雷电/转场=复用技法+增量。是 procedural-fx-generator 的「配方提取」引擎的通用化。
- [2026-06-18-procedural-fx-generator.md](specs/2026-06-18-procedural-fx-generator.md) — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执行引擎。v1=火焰,Phase 0(确定性造一个好火焰)为 make-or-break 门槛。
- [2026-06-18-roundtrip-ae-accept-residual.md](specs/2026-06-18-roundtrip-ae-accept-residual.md) — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项分类(N/A read/helper · 实勘负结论 · false-green · 硬尾 · 可做但低 ROI),有需要时按本表挑。ae-accept 35→261,roundtrip 279→53。
<!-- /AUTO -->

## 下一步

**技法内化框架**(spec `2026-06-18-fx-technique-internalization`)。地基已搭(步骤 1+2 done),剩 3+4:

- ✅ **步骤1 `cmd/aepdissect`**(done):正经 repo 工具,任意 .aep 一键出结构化报告(效果用量+原生/Cycore/第三方分级+预合成嵌套+逐层效果链)。= PARSE 步工具化。已在 Colorful Fire Ball 验证。
- ✅ **步骤2 两层知识结构**(done):`docs/fx-techniques.md`=跨域技法库(T1 噪声造质料…T9 收尾,从火焰抽,标跨现象复用)+ `checklists/build-<现象>.md`=现象配方;交叉链已通。
- ✅ **步骤4 喂第二现象**(done,闪电 Lightning Pack):框架在非程序化样本上验证通过——aepdissect 判出②素材+装配型,诚实结论「教不了生成闪电」,复用 T5 + 新增 T10–T13。两类元素概念催生。
- ✅ **步骤3 火焰 v3 多层合成**(**完成 + 用户验收过**):5 层(黑底+3 Fractal Noise→Tritone→TurbDisp 火层不同 scale + 顶 Glo2 调整层)+ **同心 mask** 温度分区(外深红→中橙→内黄白热芯柱)+ Add 叠 + 运动共相。**`TestFlameDemo_AEShipGate_AE2020/AE2025` 双版本绿(逐像素一致)+ 用户真机「过了」**。配方提进 ship-gate test + showcase gen.go(complete)+ checklist 升「验证配方」。验证技法 T3 additive-depth。踩坑(团块→同心 mask / 抖动→运动共相)记入 fx-techniques T3。
- ⬜ **可选:程序化闪电**(T14 `ADBE Lightning 2` 已在库,未验证)= 不靠素材纯生成电弧 + T5 Glow + T6 闪烁。
- ⬜ 喂更多模版,继续长技法库。

依据:`specs/2026-06-18-fx-technique-internalization.md` · `docs/fx-techniques.md` · `checklists/build-good-fire.md` · v2 builder `tmp_debug/flame2/`(local)。

**次要(需求驱动)**:残值 backlog(剩 53 roundtrip)→ `specs/2026-06-18-roundtrip-ae-accept-residual.md`;能力真相源 `go run ./cmd/capindex -q <词>`。

## Backlog

- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
