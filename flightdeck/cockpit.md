# Cockpit — aep-parser

Updated: 2026-06-22 · claude · Stage: Booyah **Phase 1 全 complete（①②③④ 用户验收）+ Phase 2 ⑤ プリコンポジション 1 成**（首遇 NewPrecompLayer 嵌套，3 层→comp ①，双版本 AE-accept + DOM source 绑定确认，🔶待 review）；下一步 = ⑥ シェイプの塊（7 层 src=⑤）

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

**Phase 1 叶子 ①②③④ 全 complete（用户验收）**，详见 [showcase/INDEX](showcase/INDEX.md) 账本。沉淀的坑：separated-position 读 0,0 陷阱（[[shape-layer-position-default-offscreen]] Case 2，④⑤ 均撞，显式居中）· `SetLayerTransform.Scale` 单位 percent 需 ×100 · footage-share=各自 solid（无共享 API,render-neutral）· precomp/solid 等 embed-template clone 的 start 须 post-reopen `SetStartTime`（pre-reopen scene 字段无效）。

**comp ⑤ プリコンポジション 1 ✅（本会话，🔶待 review；居中 bug 已修+render 眼验）**：`gen_precomp1.go` **首遇 precomp 嵌套**——3 个 `NewPrecompLayer` 全引用 comp ①（staggered 三份 glitch）。L0 Position 2kf[778→1204]/L1 Opacity 13kf+居中/L2 静态居中；start 错峰经 `SetStartTime`。**揪出库 bug**（用户真机反馈"右下角"→bisect）：`SetLayerTransform` 对 **AV/precomp 层 anchor 按「源尺寸分数」编码**（0.5=中心），非 shape/text 像素——写 960,540 被 AE 读成 ×源尺寸(1.84M)→渲染飞出屏全黑（anchor 0,0 不暴露故前版渲右下角）。修法=写分数 `anchor(0.5,0.5)`，AE DOM 实测读回 960,540，render 复验居中。详 [[setlayertransform-av-anchor-fraction]]（库 bug，gen 暂 workaround）。**双版本 AE-accept**（5 comp、3 层不 drop、source 绑 ①）+ AE2025 render 居中眼验。

**下一 = Task 2.2 ⑥ シェイプの塊**（[plan](plans/2026-06-19-booyah-glitch-replication.md)）：7 层均 src=⑤（复用 ⑤ 的 `NewPrecompLayer` 建法）+ 各层 Position(2kf)+Opacity(34/41/39/39/39/44 kf,L6 无)，值 oracle 取。→ ⑦(3 层+Glow)→ ⑧(3 层 src=② Fill 色差)→ ⑨(3 层+Trim,2 层 src=④)→ Phase 3 ⑩ 怪物 → Phase 4 ⑪⑫ 顶层+终帧。⚠AE 装 `E:\adobe\`。

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
