# 内化一个参考工程 — 5 步流水线（每喂一个 .aep 跑一遍） — checklist

SUMMARY: 内化一个参考工程 — 5 步流水线（每喂一个 .aep 跑一遍）
READ WHEN: 拿到一个参考 .aep 要「理解/内化」它(抽可复用技法,而非只读懂这一个);要把某现象工程拆成三轴本体(角色/技法/机制);纠结理解的产出该长什么样、存哪;给学习引擎/生成引擎喂新样本前

---

> **这是干嘛的**：把任意参考 `.aep` 系统地「内化」成**可复用、可迁移**的技法，而不是只读懂这一个怎么做的。每喂一个样本跑一遍，技法库单调增长 → 新现象越来越便宜。
> **契约**：产出形态是三轴 schema / project_profile；本文件是执行型流程。技法库实例 = `fx-techniques.md`。已跑过的实例 = `build-good-fire.md`（火焰，本区）。

## 核心心法（先记住，再走步骤）

- **资产是「技法」不是「现象」**：每现象 = 一组跨域技法的组合。抽的是能搬到别的现象去的**因果决策**（「竖拉+高对比噪声造瘦高料子」），不是参数值（「对比度 169」=机制层 signal）。
- **理解 = 降噪**：一个工程 95% 是默认/样板/不可见。AE 的 elision 只存非默认参数 → 字节本身≈作者改过的（aepdissect 用 `✎` 标）。找那 5% 信号层/信号参数，不读全。
- **诚实分级**：`proven`（真渲过）vs `hypothesized`（推断）；`native`/`cycore`/`third-party`/`lib-blocked`。混淆会让「通用技巧」虚高（红线4 精神）。

## 何时把发现总结进通用知识（归属判据）— 三层

每个发现落三层之一，判据 = **可迁移性**（一句话测试：「能搬到另一个现象吗？」）：

| 落点 | 判据 | 例 |
|---|---|---|
| **通用技法库** `fx-techniques.md`（T-atom） | 「能搬到另一现象」= **能** → 是可迁移的因果决策 | rgb-channel-split（glitch→赛博朋克/CRT/转场）；displacement-distortion（火→风→glitch） |
| **现象配方层** `build-<现象>.md` | **现象专属**：技法间的**顺序/编排**、为这个观感调的特定组合 | 火焰「噪声→位移→上色→Add 叠」的次序；某 look 的具体参数搭 |
| **不单列**（机制层 signal / 不收） | 是个**参数值**或一次性琐碎 | Contrast=169 → 记进某技法 `observed_range`，不立条 |

**confidence 怎么升（强化规则）**：首次见 → `observed`/`hypothesized`（低，一个工程里看到≠通用）；**跨工程复现** → 加 `proven_transfers`（同 flightdeck「第二次出现=pattern」，技法库越用越强靠这个）；**AE render-gate 过** → `validated`。立新条前**先检索去重**，能并进已有 `any_of`/`proven_transfers` 就别新立。

> **判不准时默认留配方层**（宁可窄）：把现象专属误当通用 = 污染通用知识，比漏收更坏（红线4 精神）。

## 步骤

1. **PARSE（机械）** — `go run ./cmd/aepdissect "<file.aep>"`
   - 读：效果用量 + 原生/Cycore/第三方分级、预合成嵌套、**依赖图边**（source/parent/matte/expr）、逐层效果链（`✎`=非默认=配方信号，`~`=有关键帧=主角，`act=[in→out]`=活跃区间）。
   - `-json` 出结构 dump（注：当前**还没**自动输出完整 project_profile：meta/fingerprint/timeline/graph/techniques 仍需人工综合，是 aepdissect 待补的工具缺口）。
   - 辅助：`go run ./tools/debug/dump_tdmn`（全树 matchName）。
2. **判两类元素**（决定能否程序化复刻）：
   - **①程序化**（solid 源 + 生成类效果 + 关键帧）→ 能学会复刻。
   - **②素材+装配**（footage 源 + 0 关键帧 + 只有重上色/辉光/调色）→ 装配能复刻、元素得用户提供。诚实说清。
   - **插件依赖是正交维度**（见下「第三方插件」节）：① 或 ② 都可能用插件；**插件 ≠ 跳过**，照样抽技法、标 `reproducibility: third-party`。
3. **DECOMPOSE 到角色**（判断，不可机械化）：每个信号效果/结构 → 它解决**什么问题**（form/distort/color/depth/glow/motion/control…受控词表写在本文件和 `fx-techniques.md` 的 schema 约定里）。产物 = 角色词典。
4. **EXTRACT 技法**（提炼跨域原子）：每个「可迁移因果决策」抽成命名条目，对齐 schema v2：
   - 新技法 → 在 `fx-techniques.md` 按 schema 追加（`mechanism.any_of` 等价效果集是对齐命门、`reproducibility` 两正交轴、`not_this` 反例必填）。
   - 已有技法 → 标新 `proven_transfers`（交叉链越用越强）；新等价实现 → 加进 `any_of`。
   - **先检索去重**（防粒度漂移）。
5. **STORE（两层结构，flightdeck 自动跨会话）**：
   - 技法原子 → `fx-techniques.md`（跨域、读懂型）。
   - 现象配方 → 自包含的 `build-<现象>.md`（技法的**有序**组合，顺序/依赖归这层，不归技法）。
   - project_profile（每工程一份，可跨工程对齐）→ 暂存进该现象的 `build-<现象>.md`（火焰范例见 build-good-fire.md § 源工程真实结构）。
   - preflight 扫 routing header 建知识地图；当 `READ WHEN` 命中时再读正文 = 下个会话能恢复这套知识，而不是靠聊天记忆。
6. **VALIDATE（金标准 = 能复现）**：按配方 build（`NewSolidLayer`×N + `AddEffect` + `SetEffectParam`/`AnimateEffectParam` + `SetBlendingMode` + `SetEffectLayerParam`，全 gated）→ AE 渲 → render-diff 对照原图。只有渲过的技法/配方升 `confidence: validated`。**理解引擎 = 生成引擎**（能复现才算真懂）。

## 第三方插件：支持，不是跳过（2026-06-18 用户定调）

招牌插件往往**就是那个技法本身**（Videocopilot Twitch = glitch 跳变；Colorama = 调色；Trapcode = 粒子）。写死/跳过 = 丢掉技法。立场改为**支持**：

- **读 / 保真**：✅ 已做——opaque preservation 字节级 round-trip 任何未知效果 chunk（红线5），aepdissect 连其参数槽都 dump（`PEQCAGL-0013=189`）。读插件工程 = 已解决。
- **从零写（AddEffect 插件效果）**：可行路 = **embed-template**（同 native 效果，关键机制是从用了它的**真实样本 .aep 采该效果 chunk**：sspc/tdmn/参数 → 嵌入 → splice 进目标，参数按 index 戳值）。产物 AE 能打开；**装了插件才渲对**（没装 = AE 显示 missing-effect 占位）。**尚未实现，是可行的下一能力。**
- **`reproducibility.mechanism` 语义重定**：`third-party` = **「可支持——读✓ / 写靠采模板 / 渲染需装插件」**，**不**等于不可做。真正做不了的是 `lib-blocked`（表达式 AE 不求值，见 [[expression-enable-byte-pair]]）。
- **与交付准则的关系**：`plugin-free` 仍是 **procedural-fx-generator 产品**的优先（网页用户没装 Twitch）；但**学习库 + 装了插件的用户 + 我们自己的分析**里，插件技法是一等公民、可建，**诚实标注渲染依赖**即可（别声称无插件能渲）。
- **采模板金矿** = `samples/motionbox/`（含真实 Twitch/Colorama/PEDG 用法），正好当 embed-template 采集源。

## 易错（红线）

- **Go round-trip 绿 ≠ AE 接受 ≠ 渲染对**：渲染类技法必须验到像素（红线1/4）。
- **本库表达式 AE 不求值**：样本用 `time*[0,-500]` 这类表达式驱动的，复刻一律换关键帧（见 [[expression-enable-byte-pair]]）。
- **别把现象专属当通用**：技法判据是「能搬到另一现象」；搬不动的（顺序/某现象特定参数）归配方层。

## 交叉链接

- 契约：三轴 schema/受控词表 + 本流水线 + 跨现象复用表，关键执行规则已写在本文件。
- 技法库：`fx-techniques.md`。工具：`cmd/aepdissect`。字典（译参数/标非默认）：[[effects-dict]]。
- 已跑实例：`build-good-fire.md`（火焰，①程序化，5 级嵌套）· 闪电 Lightning Pack（②素材+装配，素材提供真实电弧，本库只能复刻重上色/辉光/控制器装配）。
