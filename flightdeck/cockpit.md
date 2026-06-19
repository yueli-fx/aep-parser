# Cockpit — aep-parser

**Last updated**: 2026-06-18 by claude（**「理解一个工程」研究 arc：效果字典 + 三轴本体冻结 + 理解自动化**。① `effects-dict` 基础设施:5 版本英文效果字典(matchName→名+类型+**默认值**)+ 可重生成 dumper(`scripts/dump_effects_dict.*`)+ checklist。② **`technique-ontology` spec 冻结(graduate)**:角色/技法/机制三轴本体 + schema v2,投大规模工程前的数据契约,经火焰/闪电/控制器三类压测。③ `fx-techniques` 14 技法重构到三轴(glow 跨现象复用实证)。④ **aepdissect 消费字典**:自动译参数 + 标非默认 = 配方信号(`✎ Glow Threshold=139 (default 153)` + `→ tuned:` 摘要)。核心洞察:.aep 用 **elision**(只存非默认)→ 字节即配方信号、免费降噪。火焰 v3 早先已用户验收 + 双版本 gate 过。）

**Active focus**: **Booyah Glitch 全工程复刻 = 理解金标准检验**(spec `2026-06-19-booyah-glitch-full-replication.md` + plan `2026-06-19-booyah-glitch-replication.md`)。把整个真实工程(12 comp/61 层/~100 mask/wiggle+Evolution 表达式/Curves)用咱们 Go API 从零重建——round-trip 只证读得回字节,**从零重建证真懂每个 chunk 怎么来的**。用户铁律:**不逃避未知字段**(撞墙 RE,物理不可写给实证负结论)。判据 = 结构保真(AE 接受 + DOM 读回值对账)+ 终帧渲染像素对照;范围 = 全 12 comp,DAG 叶→根逐 comp 验;方法 = 手写 per-comp 生成器、**原工程只读当取值神谕、chunk 全由 API 重建(绝不 copy 字节)**。是上游「技法内化框架」(`specs/2026-06-18-fx-technique-internalization.md`)的「复现判据 = 理解金标准」检验项,bar 拉到字节/工程级。**不变量**:知识单一家 = flightdeck + CLAUDE.md;能力真相源 = capindex(`go run ./cmd/capindex -q <词>`);每渲染类双版本 AE ship-gate(红线4)。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-18-fx-technique-internalization.md](specs/2026-06-18-fx-technique-internalization.md) — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techni…
- [2026-06-18-procedural-fx-generator.md](specs/2026-06-18-procedural-fx-generator.md) — 用户 NL 描述 → 网站一键产出可在 AE 打开的 .aep。架构=离线配方提取 + 运行时(NL→参数→库出字节)。AI 永不碰字节,Go 库是保证合法的执…
- [2026-06-18-roundtrip-ae-accept-residual.md](specs/2026-06-18-roundtrip-ae-accept-residual.md) — 补验 arc(批1-28,2026-06-17/18)完成记录 + 剩 53 个 verify=roundtrip 残值的 someday-backlog:逐项…
- [2026-06-18-technique-ontology.md](specs/2026-06-18-technique-ontology.md) — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized…
- [2026-06-19-booyah-glitch-full-replication.md](specs/2026-06-19-booyah-glitch-full-replication.md) — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle 表达式/Curves)用咱们的 Go API 从零完整…
- [2026-06-20-pseudo-effect-support.md](specs/2026-06-20-pseudo-effect-support.md) — 让本库纯 Go 离线把 AE Pseudo Effect(.ffx 自定义伪效果)splice 进层 Effect Parade,免开 AE。.ffx=RIFX…
- [2026-06-19-booyah-glitch-replication.md](plans/2026-06-19-booyah-glitch-replication.md) — 实现 Booyah Glitch 全工程复刻:Phase0 前置 spike(表达式 Evolution=time*N + wiggle / 单层 ~21 ma…
<!-- /AUTO -->

## 下一步

**Booyah Glitch 全工程复刻 arc — Phase 0 完成 ✅,Phase 1 起**(plan `2026-06-19-booyah-glitch-replication.md`)。

Phase 0(前置 de-risk)收口:
- ✅ **0.1 scaffold**(04e57eb):`flightdeck/showcase/booyah-clone/` 包 + oracle(只读神谕)+ 覆盖账本。
- ✅ **0.2 读侧审计**(c3bff18):mask 顶点/keyframe/params/expr 全可读;Curves 曲线数据=读侧 gap。
- ✅ **0.3 表达式 = GO**(29b4db6,头号 de-risk 解除):`SetExpression` 早已 stable+双版本 AE gated;Evolution=`time*N`(`ADBE Fractal Noise-0023`)+ `wiggle(34,0.29)`(`ADBE Exposure2-0003`)round-trip 干净,落在已 gated 字节路径+idiom。**噪声动画+wiggle 能忠实复刻,无需降级。**
- 🔶 **0.4 many-mask(21)** 并入 ⑩ 在位 AE 验(masks 集中在 ⑩;Go round-trip 对 silent-drop 假绿)。
- ⛔ **0.5 Curves**:曲线数据物理 blocked(arbitrary-data 无 scripting),实例可加 → ⑫ 降级 + 在位验。

**Phase 1 进行中**:
- 🔶 **Task 1.1 comp ① シェイイイイプ**(commit e4d266c,Go 建):4 rect 精确 kf + fill 色 round-trip 逐值对账原工程 ✓;结构 delta(shape 组嵌套 + transform 默认物化,render-neutral)已记账。**AE-accept/render 待验**。
- ⚠ **首次 AE 验证(基座检查)撞 crash-state cascade,受阻**(详 `incidents/ae-automation-occlusion-crashstate.md` Case 2c):run1 的 verify.jsx 用 JSON.stringify 在 catch 外抛 → 0 字节 done → ae_run 假 PASS(exit 0)→ force-kill AE → 置崩溃标志;run2 弹 safe-mode 框,但 ae_run 的 `SetForegroundWindow`+SendKeys 关框被**前台游戏 foreground-lock 挡住**(focus-mismatch)→ 超时。已修 verify.jsx(纯字符串、末尾一次写)。**注:游戏窗口≠用户在用(用户在另一台机器,operator-context 立规),不问用户让机器。**
- ⬜ **下一步 = 重建 `clear_ae_crashstate.ps1` 为 tracked 工具**(`tools/debug/`,免再丢;**PostMessage** `VK_RETURN`=继续 到 safe-mode hwnd,不用 SetForegroundWindow)→ 清崩溃标志 → warm-retry verify.jsx(确认 AE 接受 + shape 不 drop)→ render comp ① 对照原工程。这是 arc 全程 AE 验证的前置基建。
- ⬜ 然后 Task 1.2-1.4:② テキスト(text+animator)③ マップ用ノイズ(fractal+expr,0.3 已 GO)④ カクッ(shape),按 DAG 上行。

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
