# Cockpit — aep-parser

Updated: 2026-06-22 · claude · Stage: Booyah **Phase 1（①②③④）✅ + Phase 2 ⑤⑥⑦ ✅（用户真机验收）+ ⑧ RGBズレ ✅（🔶待 review，3 层 src=② + Fill RGB 色差，4 关全过 + render 眼验红绿蓝三色错位）**；下一步 = Task 2.5 ⑨ なんか周りのやつ（L0/L1 src=④ + L2 shape 层带 Trim Paths 动画）。旁支：本会话另实证了降级器 bug #6（非相邻 track-matte 降级丢绑，incidents/2026-06-09-ae-version-downgrader-re.md）

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

**Phase 1 叶子 ①②③④ 全 complete（用户验收）**，详见 [showcase/INDEX](showcase/INDEX.md) 账本。沉淀的坑：separated-position 读 0,0 陷阱（[[shape-layer-position-default-offscreen]] Case 2，④⑤ 均撞，显式居中）· `SetLayerTransform.Scale` 单位 percent 需 ×100 · footage-share=各自 solid（无共享 API,render-neutral）· precomp/solid 等 embed-template clone 的 start 须 post-reopen `SetStartTime`（pre-reopen scene 字段无效）。

**comp ⑤ プリコンポジション 1 ✅ complete（用户真机验收）**：`gen_precomp1.go`，3 个 `NewPrecompLayer`→comp ①（staggered）。L0 Position 2kf/L1 Opacity 13kf/L2 静态，全居中 + start 错峰。复刻保真逼出**两个独立 bug**（用户逐帧对比 frame18 揪出）：① **tdb4 @0x0C 关键帧时基**（库 bug，commit 39d3c16）——AE 按 tdb4@0x0C 而非 cdta 求值关键帧 tick，硬编码 30720 使 29.97 被压 0.781；修法 makeTdb4/injectAnimatedStream 改用 comp.TickRate，详 [[ntsc-tickrate-derive-3x-off]] § tdb4。② **gen 漏复刻静态 Scale/Rotation**——只搬了 Position/Opacity，原版 L0=44%/L1=45%+180° 被默认成 100%/0°，合成飞散；修法读全 transform 通道，详 [[layer-replication-drops-static-transform-channels]]。两 bug 正交（前者管关键帧 WHEN，后者管静态 transform WHAT）。逐帧 orig-vs-clone 对比工作流 → `checklists/showcase.md`。anchor 分数坑见 [[setlayertransform-av-anchor-fraction]]。

**comp ⑥ シェイプの塊 ✅（🔶待 review，commit 6c695b9）**：`gen_shape_katamari.go` 7 层全 src=⑤，复用 ⑤ 的 `NewPrecompLayer` 建法 + **复刻全 transform 通道**（Position 2kf 绝对坐标 541→1340 横滑 + Opacity 34/41/39/39/39/44 kf + 静态 Scale 24/33/75% + RotateZ 90° on L0/L3/L4，anchor 源中心分数 0.5,0.5）。**4 关全过**：Go 结构对账（7 层 srcID 全=⑤、kf 数 + 静态 Scale/Rotation/start 全等）+ AE2020≡AE2025 接受（6 comp、⑥ 7 层不 drop、DOM source 全绑 ⑤）+ render 眼验（t=1.0 渲出缩放 glitch 切片簇，无 blank/飞散/off-screen）。带上 ⑤ 两教训：复刻全 transform 通道（[[layer-replication-drops-static-transform-channels]]）+ AV-anchor 用分数（[[setlayertransform-av-anchor-fraction]]）；本次 Position 是绝对坐标无 separated 0,0 陷阱。无新 API，终帧 render-pixel 归 ⑫。

**comp ⑧ RGBズレ ✅（🔶待 review，commit 见下）**：`gen_rgbzure.go` 3 层全 src=② テキスト + **各层 `AddEffect("ADBE Fill")` 染单一 RGB 通道**（L0"B"=[A,R,G,B][255,0,131,255] 蓝 / L1"R" Fill-0002 elided=AE 默认红 / L2"G"=[255,0,255,86] 绿）+ start 错峰 → 文字 chromatic aberration。全 Position 静态[960,540]居中,L2 start=**-0.1335 负值**正确 round-trip。4 关全过：Go 结构对账（3 层 srcID=②、Opacity 23·24·25kf、Fill 颜色 L0L2 精确·L1 elided、start 含负全等）+ AE2020≡AE2025 接受（3×Fill 入 DOM、颜色 AE 读回对 [A,R,G,B]→AE[R,G,B,A]）+ **render 眼验 t=1.0 红绿蓝三份文字重叠+相位错峰**。**坑**：`SetEffectParam` 颜色要 `[]float64` len=Components（传 [4]float64 报 unsupported value type）、编码 [A,R,G,B] 0-255=parser 读出格式；L1 默认红靠 AddEffect 模板默认（AE DOM 确认）。无新 API，终帧 render-pixel 归 ⑫。

**下一 = Task 2.5 ⑨ なんか周りのやつ**（[plan](plans/2026-06-19-booyah-glitch-replication.md)）：L0/L1 src=④ カクッ（无动画）+ **L2 = shape 层 "シェイプレイヤー 1" 带 Trim Paths 动画**（`ADBE Vector Trim Start` 2kf ease + `ADBE Vector Trim Offset` 2kf ease）。**新维度 = 混合源（precomp + 新建 shape 层）+ shape Trim 动画**（`AddTrim`，[[trim-paths-vector-filter-re]]）。→ Phase 3 ⑩ 怪物（27 层/~100 mask）→ Phase 4 ⑪⑫ 顶层+终帧。⚠AE 装 `E:\adobe\`。

**B. @tag schema flip(c)（可选遗留清理，非阻塞）** → plan [2026-06-21-apidoc-tag-schema-impl.md](plans/2026-06-21-apidoc-tag-schema-impl.md)。删 `extract.go` `parseCapTag` 旧读路径 + `tag.go` 旧枚举 map + 守卫测试 + `--validate` strict required CI + 改文档真相源注脚 → regen → commit flip → plan done → landing。dual-read 现仍工作、aep:cap=0，可随时做。

**旁支**：补 comp ① AE 2020 render；**扫 comp ②③④ 图层有无非默认 Scale/Rotation/Anchor 被 gen 漏复刻**（与 ⑤ 同类隐患 [[layer-replication-drops-static-transform-channels]]，用户建议回扫）。（注：原「NewComposition 分数 fps cdta 时基」第二发现已根因化为 tdb4 @0x0C 并修复，见 [[ntsc-tickrate-derive-3x-off]]。）

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
