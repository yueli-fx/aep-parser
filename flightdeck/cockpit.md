# Cockpit — aep-parser

Updated: 2026-06-22 · claude · Stage: Booyah comp ② 拼成 + 真机 review 追修中（样式/居中已对；**UI 编辑/删层报错根因=跨-comp 层 ID 撞号**，已修用户层 ID 全局化 commit 383bcd2，待真机验）；下一步 = 验通过→comp ③

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

**⚠ 待用户真机 review comp ②**（重生成 `go run ./flightdeck/showcase/booyah-clone` 后开 `booyah-clone.aep`）：本会话据真机反馈连修 4 处——①字体/字号/斜体/居中对齐原 text doc（f40cbe3）②竖直居中=anchor 按自身文字框中心（8e7f527）③动画器补 companion leaf（e870062，非 UI 报错主因）④**UI 编辑/删层报错 `{unexpected match name searched for in group}` 根因=跨-comp 用户层 ID 撞号**（comp① shape 与 comp② text 都拿 ID 13；`allocLayerID` 改全局单调，383bcd2，单-comp gate 字节不变+双版本 PASS，守卫 `TestNewLayer_CrossCompIDsDistinct`）。**复现卡点**：该报错**真机 UI 专属**，render+round-trip+ExtendScript 全过都复现不出（UI 走 ID-keyed 查找、脚本走直接引用）→ 只能用户真机验。**残留待定**：service 层（DLay/SLay…2..12）跨-comp 也撞号，但理论上 AE 视其 comp-内部而容忍（文件能在严格的 AE2025 open）；若 UI 报错仍在=service 撞号也需修（NewComposition re-ID，更大改动）。详 [[nextitemid-must-include-layer-ids]] 第三回。

**A = 续建 Booyah Phase 1**（Task 1.3/1.4，[plan](plans/2026-06-19-booyah-glitch-replication.md)）：
1. **③ マップ用フラクタルノイズ**：2 Fractal Noise 层（footage 源共享建法首遇——实测确定）+ Evolution `SetExpression("time*1200"/"time*3000")`（0.3 已 GO）+ Offset Turbulence `AnimateEffectParam` + L0 blend=Overlay。
2. **④ カクッ**：2 shape 层，层级 Scale(2kf)+Opacity(20kf) + oracle 提取静态 path。
3. 各走验收四关 + 账本。

**comp ② 已成（本会话，commit 2add514+377006c）**：两文字动画器 ship（`AddTextTrackingAnimator`/`AddTextCharacterOffsetAnimator`+`AnimateText*`，双版本 render-gate PASS，RE 直接从真实工程读，详 [[text-animator-create-re]] 末 Case）+ `gen_text_komako.go` 拼成 comp ②（`SetLayerTransform` 28kf Position/24kf Opacity + 2 动画器）。capindex 486→490。**fidelity delta**（非阻塞）：kf linear 近似原 ease；字体/字色用默认（原色经 ⑧ RGBズレ 覆写）。⚠AE 装 `E:\adobe\`。

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
