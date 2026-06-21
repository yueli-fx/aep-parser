# Cockpit — aep-parser

Updated: 2026-06-22 · claude · Stage: Booyah **Phase 1 叶子 comp 全完成**（① ② ③ 用户验收 · ④ カクッ 双版本 AE-accept，🔶待 review）；下一步 = Phase 2 中层 comp ⑤ プリコンポジション 1（首遇 NewPrecompLayer 嵌套）

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

**comp ③ マップ用フラクタルノイズ ✅ complete（用户真机验收 2026-06-22；a0ce40f）**：`gen_fractal_map.go` 2 黑 solid 各 1 Fractal Noise（L0 Overlay / L1 Normal）。**双版本 AE-accept + DOM 对账 PASS**——2 层不 drop、Fractal Noise 32 props 全在、Evolution `expr="time*1200"/"time*3000" on`（AE DOM 确认表达式真启用，红线1 清，非假绿）、Offset Turbulence 2kf→DOM 960,540（parser fraction[0.5,0.5]→AE 像素中心映射对）、UniformScaling off、blend 对。值全 oracle 取（连 L1 `time*3000\r` 尾 CR）。**footage-share 实测结论**：footage(67)=黑 solid，无 from-scratch 共享 footage-item API（SetSource 留孤儿 / DuplicateLayer 继承 effect 需脆弱二次 reopen）→ 各自 1 黑 solid（Fractal Noise 自生成像素，render-neutral delta）。终帧 render-pixel 归 ⑫。verify.jsx 加 `dumpEffects`（服务后续 ④⑦⑧⑨⑩⑪⑫）。无新 API（组合既有 gated 能力），capindex 不变。

**comp ④ カクッ ✅（本会话，🔶待用户 review）**：`gen_kakuh.go` 2 shape 层=**stroked rect(1635×810,无 fill,0.8px)+ Trim Paths(Start=97/Offset=5.3)** + 层级 `SetLayerTransform` Scale(2kf 64%→100%)/Opacity(20kf)，L0 加 RotateZ=180。双版本 AE-accept（4 comp 全在、④ 2 层不 drop、2×Stroke+2×Trim 入 DOM、OK）。**三个坑修掉**：①shape 几何是参数化(Rect+Stroke+Trim)非 freeform path（读侧实测）；②`SetLayerTransform.Scale` 单位=**percent 非 fraction**，需 ×100（初版 scale 100× 偏小）；③**层位置**（用户真机反馈）：原工程 Position 是 separated dimensions，parser 读 0,0 但真机居中——照抄 0,0 摆到左上角；修法=显式居中 comp 中心，render 复验 rect 居中。详 [[shape-layer-position-default-offscreen]] 第二 Case（separated-position 读侧陷阱）。Delta 记账：描边色取 AddStroke 默认黑（原 elide 了=AE 建时默认不可恢复，0.8px 细线）·shape wrapper 多套一层(同①)·Scale Z=1 vs 原 0.853(2D 无关)。无新 API。

**下一 = Phase 2 中层 comp**（[plan](plans/2026-06-19-booyah-glitch-replication.md)）：**Task 2.1 ⑤ プリコンポジション 1**（3 层均 src=① シェイイイイプ，**首遇 `NewPrecompLayer` 嵌套** + L0 Position(2kf)/L1 Opacity(13kf)）→ ⑥(7 层 src=⑤)→ ⑦(3 层+Glow)→ ⑧(3 层 src=② Fill 色差)→ ⑨(3 层+Trim,2 层 src=④)。

**Booyah 进度**：Phase 1 叶子 ① ✅ ② ✅ ③ ✅(均 complete) ④ ✅(🔶待 review) → Phase 2 中层 ⑤⑥⑦⑧⑨ 待建 → Phase 3 ⑩ 怪物 → Phase 4 ⑪⑫ 顶层+终帧。⚠AE 装 `E:\adobe\`。

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
