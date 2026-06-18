---
status: active
when_to_read: 拿到一个参考 .aep 要「理解/内化」它(抽可复用技法,而非只读懂这一个);要把某现象工程拆成三轴本体(角色/技法/机制);纠结理解的产出该长什么样、存哪;给学习引擎/生成引擎喂新样本前
applies_to: [internalization, understand-project, pipeline, aepdissect, three-axis-ontology, role-decompose, technique-extract, project-profile, reproducibility, cross-domain, recipe-signal, elision]
last_updated: 2026-06-18
---

# 内化一个参考工程 — 5 步流水线（每喂一个 .aep 跑一遍）

> **这是干嘛的**：把任意参考 `.aep` 系统地「内化」成**可复用、可迁移**的技法，而不是只读懂这一个怎么做的。每喂一个样本跑一遍，技法库单调增长 → 新现象越来越便宜。
> **契约**：产出形态（三轴 schema / project_profile）由 `specs/2026-06-18-technique-ontology.md` 冻结；本流水线由 `specs/2026-06-18-fx-technique-internalization.md` 定义——**那两份是「读懂型」契约（spec→docs），本文件是「执行型」流程（checklist）**。技法库实例 = `docs/fx-techniques.md`。已跑过的实例 = `build-good-fire.md`（火焰，本区）。

## 核心心法（先记住，再走步骤）

- **资产是「技法」不是「现象」**：每现象 = 一组跨域技法的组合。抽的是能搬到别的现象去的**因果决策**（「竖拉+高对比噪声造瘦高料子」），不是参数值（「对比度 169」=机制层 signal）。
- **理解 = 降噪**：一个工程 95% 是默认/样板/不可见。AE 的 elision 只存非默认参数 → 字节本身≈作者改过的（aepdissect 用 `✎` 标）。找那 5% 信号层/信号参数，不读全。
- **诚实分级**：`proven`（真渲过）vs `hypothesized`（推断）；`native`/`cycore`/`third-party`/`lib-blocked`。混淆会让「通用技巧」虚高（红线4 精神）。

## 步骤

1. **PARSE（机械）** — `go run ./cmd/aepdissect "<file.aep>"`
   - 读：效果用量 + 原生/Cycore/第三方分级、预合成嵌套、**依赖图边**（source/parent/matte/expr）、逐层效果链（`✎`=非默认=配方信号，`~`=有关键帧=主角，`act=[in→out]`=活跃区间）。
   - `-json` 出结构 dump（注：当前**还没**按 spec §3 project_profile schema 输出 meta/fingerprint/timeline/graph/techniques——那部分仍人工综合，是 aepdissect 待补的工具缺口）。
   - 辅助：`go run ./tools/debug/dump_tdmn`（全树 matchName）。
2. **判两类元素**（决定能否程序化复刻）：
   - **①程序化**（solid 源 + 生成类效果 + 关键帧）→ 能学会复刻。
   - **②素材+装配**（footage 源 + 0 关键帧 + 只有重上色/辉光/调色）→ 装配能复刻、元素得用户提供。诚实说清。
3. **DECOMPOSE 到角色**（判断，不可机械化）：每个信号效果/结构 → 它解决**什么问题**（form/distort/color/depth/glow/motion/control…受控词表见 ontology spec §4）。产物 = 角色词典。
4. **EXTRACT 技法**（提炼跨域原子）：每个「可迁移因果决策」抽成命名条目，对齐 schema v2：
   - 新技法 → 在 `docs/fx-techniques.md` 按 schema 追加（`mechanism.any_of` 等价效果集是对齐命门、`reproducibility` 两正交轴、`not_this` 反例必填）。
   - 已有技法 → 标新 `proven_transfers`（交叉链越用越强）；新等价实现 → 加进 `any_of`。
   - **先检索去重**（防粒度漂移）。
5. **STORE（两层结构，flightdeck 自动跨会话）**：
   - 技法原子 → `docs/fx-techniques.md`（跨域、读懂型）。
   - 现象配方 → `checklists/techniques/build-<现象>.md`（技法的**有序**组合，顺序/依赖归这层，不归技法）。
   - project_profile（每工程一份，可跨工程对齐）→ 暂存进该现象的 `build-<现象>.md`（火焰范例见 build-good-fire.md § 源工程真实结构）。
   - preflight 每次完整载入 INDEX = 下个会话自动浮现 = 「内化成我的」的确切机制。
6. **VALIDATE（金标准 = 能复现）**：按配方 build（`NewSolidLayer`×N + `AddEffect` + `SetEffectParam`/`AnimateEffectParam` + `SetBlendingMode` + `SetEffectLayerParam`，全 gated）→ AE 渲 → render-diff 对照原图。只有渲过的技法/配方升 `confidence: validated`。**理解引擎 = 生成引擎**（能复现才算真懂）。

## 易错（红线）

- **Go round-trip 绿 ≠ AE 接受 ≠ 渲染对**：渲染类技法必须验到像素（红线1/4）。
- **本库表达式 AE 不求值**：样本用 `time*[0,-500]` 这类表达式驱动的，复刻一律换关键帧（见 [[expression-enable-byte-pair]]）。
- **别把现象专属当通用**：技法判据是「能搬到另一现象」；搬不动的（顺序/某现象特定参数）归配方层。

## 交叉链接

- 契约：`specs/2026-06-18-technique-ontology.md`（三轴 schema/受控词表）· `specs/2026-06-18-fx-technique-internalization.md`（本流水线上游 + 跨现象复用表）。
- 技法库：`docs/fx-techniques.md`。工具：`cmd/aepdissect`。字典（译参数/标非默认）：[[effects-dict]]。
- 已跑实例：`build-good-fire.md`（火焰，①程序化，5 级嵌套）· 闪电 Lightning Pack（②素材+装配，见 internalization spec § 样本实证 #2）。
