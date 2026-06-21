# Cockpit — aep-parser

Updated: 2026-06-22 · claude · Stage: Booyah comp ② 文字动画器 gap #2 Tracking + #3 Character Offset **已 ship**（双版本 render-gate PASS，commit 2add514，working tree 净）；下一步 = 拼 comp ② 工程

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

**A = 拼 Booyah comp ② テキスト変えるならココ！**（Task 1.2，[plan](plans/2026-06-19-booyah-glitch-replication.md)）。两个文字动画器 gap **已 ship**（见下），现可拼整 comp ②。剩下：
1. `gen_text_komako.go`：`NewTextLayer`+`SetText("GLITCH")` → 层级 **Position(28kf)+Opacity(24kf)**（`SetLayerTransform`/animate，oracle 取值）→ `AddTextTrackingAnimator`+`AnimateTextTracking`（2kf ease）→ `AddTextCharacterOffsetAnimator`+`AnimateTextCharacterOffset`（2kf ease）。kf times/values 由 oracle 从原工程 L0 "GLITCH" 提取。
2. 验收四关 + 单独 render 一帧 Read png 自查 + 账本 ② = 待review。Commit。

**已 ship（commit 2add514，本会话）**：`AddTextTrackingAnimator` / `AddTextCharacterOffsetAnimator` + `AnimateTextTracking` / `AnimateTextCharacterOffset`（alpha，text），**双版本 render-gate PASS**（AE2020≡AE2025：Tracking ink 宽 612→354；CharOffset frameDiff 0-2=2157）。RE 直接从真实工程读（companion 自动 materialize，零新字节代码），详 [[text-animator-create-re]] 末 Case。capindex 486→490。⚠AE 装 `E:\adobe\`。

**B. @tag schema flip(c)（可选遗留清理，非阻塞）** → plan [2026-06-21-apidoc-tag-schema-impl.md](plans/2026-06-21-apidoc-tag-schema-impl.md)。删 `extract.go` `parseCapTag` 旧读路径 + `tag.go` 旧枚举 map + 守卫测试 + `--validate` strict required CI + 改文档真相源注脚 → regen → commit flip → plan done → landing。dual-read 现仍工作、aep:cap=0，可随时做。

**旁支**：补 comp ① AE 2020 render；修 `NewComposition` 分数 fps cdta 时基（[[ntsc-tickrate-derive-3x-off]] 第二发现）。

## In Progress

<!-- AUTO:inprogress -->
- [2026-06-18-fx-technique-internalization.md](specs/2026-06-18-fx-technique-internalization.md) — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techni…
- [2026-06-18-technique-ontology.md](specs/2026-06-18-technique-ontology.md) — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized…
- [2026-06-19-booyah-glitch-full-replication.md](specs/2026-06-19-booyah-glitch-full-replication.md) — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle 表达式/Curves)用咱们的 Go API 从零完整…
- [2026-06-21-api-doc-tag-schema.md](specs/2026-06-21-api-doc-tag-schema.md) — Replace free-prose doc comments + the separate aep:cap directive with ONE swaggo…
- [2026-06-19-booyah-glitch-replication.md](plans/2026-06-19-booyah-glitch-replication.md) — 实现 Booyah Glitch 全工程复刻:Phase0 前置 spike(表达式 Evolution=time*N + wiggle / 单层 ~21 ma…
- [2026-06-21-apidoc-tag-schema-impl.md](plans/2026-06-21-apidoc-tag-schema-impl.md) — Step 1 of api-doc-tag-schema: build internal/apidoc (parse+schema+validate+jargo…
<!-- /AUTO -->

## Key Context

- comp ①：✅✅ **AE 2025 render 与原工程像素级一致**（中间帧 t=0.3/0.5/0.8 diff=0/0/32px，32px=sub-pixel 帧snap）。三个真 bug 全修：层 position→中心（4f579d1）·层时长（6091016）·**中间帧错位根因=`deriveTickRate` 把 NTSC kf 时间读大 3×**（cdta @0x08 才是真 tickrate，`×1000/scale` 是 py-aep 伪修正；AE valueAtTime 实证，d03101c）。根因档 `incidents/ntsc-tickrate-derive-3x-off.md`。⚠AE 2020 侧 render 待补；clone 用 30fps 绕开 NewComposition 分数 fps cdta 时基 bug（同档第二发现）。
- AE 自动化基建（arc 全程关键）：`clear_ae_crashstate.ps1` tracked 工具（`tools/debug/`，67023b6）+ **ae_run `Invoke-SendKeysSafe` 改 PostMessage 根治 foreground-lock**（b411d09）→ 所有 in-run modal 在前台游戏锁下都能无人值守消化（实证原工程 Resolve Fonts 框）。
- 复刻=理解压测定位（用户校准 2026-06-21）：复刻不为产物，为**逼出读侧漏解属性 / 写侧缺 API / 未知结构**;终极=任给一个 .aep 能说清构成+内容+cover 边界。每 comp 先读侧全审计→列 gap→建可建的→诚实标缺口。
- 新能力 `SetLayerTransform`（layer-set·stable·双版本 gate）：给**模板层**（text/precomp/footage/solid/null）物化+动画 transform（绕 AE default-omission：模板层 default 通道被省略，无 cdat 可转）。解锁 comp ②⑤⑥⑦⑧ 层级 Position/Opacity 动画。
- @tag schema spec 关键拍板（实现时照搬,别重推）：统一 @tag 吞 aep:cap（**用户拒了 hybrid,长期可维护性优先**）;字段规则机器校验（summary≤80/`@param` 对账签名/domain 16 项冻结枚举/`@since AE<年>`）;词表单一源 = `internal/apidoc/schema.go`;docgen 生成期跑 `apidoc.Validate`;单格式不变量（绝不同时带 aep:cap+@tag）;收口闸 = 残留 aep:cap=0 且无 @param TODO。
- ⚠ 预存 RED 单测（非本 arc）：`TestSynthControlEntries_PardLayout/label`（pseudo-effect label pard @0x04=0x0 want 0x20）——pseudo spec 已归档收工但此 unit test 红，待单独修。
- operator 在另一台机器、前台游戏窗口 ≠ 用户在用 → 不问用户让机器（`incidents/ae-automation-occlusion-crashstate.md` Case 2c）。
- 上游「理解工程」研究 arc（暂让位）进度归 `specs/2026-06-18-fx-technique-internalization.md` § 现状与下一步 + `technique-ontology` § 9。

## Pending Review

- (none)

## Hanging Tasks

无。
