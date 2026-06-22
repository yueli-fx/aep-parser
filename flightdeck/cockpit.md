# Cockpit — aep-parser

Updated: 2026-06-22 · claude · Stage: **Booyah Glitch 全 12-comp from-scratch 复刻 ✅ 全建成 + 双版本 AE gated（🔶待用户真机 review ⑩⑪⑫）**。本会话扫完 Phase 3+4：⑩ グリッチテキスト（27 层/131 mask/27 effect，三 Step）+ ⑪ 背景（6 solid + Venetian Blinds/Ramp/Exposure）+ ⑫ メインコンプ（顶层 3 precomp + 终帧 render 对照）。**库 fix**：`setLdtaFrac` 粗 divisor 截断（commit 2886c5e + 回归测试）。**沉淀 incident**：[[expression-enable-byte-pair]]（kf+expr 坑扩到 effect param + 级联 drop，booyah ⑪ 实证）· [[mask-shph-bbox-read-gap]]（mask 真几何在 shph bbox，`Mask.Vertices` 只 surface 归一化 ldat）。下一 = 用户真机验收 ⑩⑪⑫ → 翻 complete + 关 plan；或 polish（mask feather/opacity 补全背景保真 · eased Position kf · Noise2 能力补）。

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

**Booyah Glitch 全 12-comp from-scratch 复刻 ✅ 全建成 + 双版本 AE gated**。逐 comp 账本（建法/值/delta/commit）= [showcase/INDEX](showcase/INDEX.md)，进度 = [plan](plans/2026-06-19-booyah-glitch-replication.md) `## Progress`。①②③⑧ 已用户真机验收 complete；④⑤⑥⑦⑨⑩⑪⑫ 🔶待用户真机 review（见 Pending Review）。

**下一**：① **用户真机验收 ⑩⑪⑫**（agent 已 render 自验，按 showcase review-gate 用户验后翻 complete + 关 plan Task 4.3）。② 可选 polish（非阻塞，均 documented delta）：**mask feather/opacity 复刻**（⑫ 终帧 ⑪ flare 渲成锐利亮矩形 vs 原柔和，[[mask-shph-bbox-read-gap]] 末节）· **eased Position kf**（LayerTransform 无 eased builder，⑫ L0 75px nudge 写 linear）· **Noise2 能力补**（不在 embed set，⑪ L0 grain skip）· **Curves 曲线数据**（arbitrary-data blocked）。⚠AE 装 `E:\adobe\`，agent 自跑 `scripts/ae_run.ps1`。

**沉淀（本会话）**：库 fix `setLdtaFrac` 粗 divisor 截断（commit 2886c5e + 回归测试）· incident [[expression-enable-byte-pair]]（kf+expr 坑扩到 effect param + 级联 drop）· incident [[mask-shph-bbox-read-gap]]（mask 真几何在 shph bbox）· Displacement Map tdpi = self-ref（`AddEffect` 默认绑宿主，无需 `SetEffectLayerParam`）。

**B. @tag schema flip(c)（可选遗留清理，非阻塞）** → plan [2026-06-21-apidoc-tag-schema-impl.md](plans/2026-06-21-apidoc-tag-schema-impl.md)。删 `extract.go` `parseCapTag` 旧读路径 + `tag.go` 旧枚举 map + 守卫测试 + `--validate` strict required CI + 改文档真相源注脚 → regen → commit flip → plan done → landing。dual-read 现仍工作、aep:cap=0，可随时做。

**旁支**：补 comp ① AE 2020 render；**「gen 漏复刻层属性/标志」横切回扫（终帧 ⑫ 前做）**——已撞两类活跳坑：①静态 Scale/Rotation/Anchor（[[layer-replication-drops-static-transform-channels]]）②**motion blur 层标志**（2026-06-22 用户 frame26 揪出 ②，全项目仅 ② "GLITCH" 开，gen 漏→无拖影；已修 `SetMotionBlur`+`SetCompMotionBlur`）。

  **③ 关键帧 bezier ease — ✅ 已查证证伪、收（2026-06-22）**：探针 dump 原版 ⑤⑥⑦⑧ 的 Position/Opacity **全部 394 个 kf 都是 `interp=linear/linear`、0 个 bezier** → gen 的 `AddKeyframeLinear` 本就忠实，**无 ease 可补**。假阴性已排除：同一 parser、同一 comp ② 同一层 L0，文字动画器 `ADBE Text Tracking Amount`/`ADBE Text Character Offset` 正确读出 **bezier/bezier**（influence 0.333/0.000、1.000/0.333 = 之前记录的「出点≈0/入点=1」），证明 parser 能区分 bezier 与 linear。**结论：本工程唯一带 bezier ease 的就是文字动画器（Tracking/CharOffset），② 早已拷 `In/OutTemporalEase`；层级 Position/Opacity 一律 linear，⑤⑥⑦⑧ 关键帧 interp 与原版无别。** 此（ease/interp）线程关闭。⚠**但 ease 探针只看了 kf interp、没看表达式**——用户随后截图原版 DOM 揪出 ⑤L2/⑧ 的 **wiggle 表达式漏复刻**（见上 ⑧ 行 + [[expression-enable-byte-pair]] 末节）；已全扫 ⑤⑥⑦⑧ transform 表达式补齐（⑤L2 opacity wiggle、⑧ position X-wiggle；⑥⑦ 无表达式真干净）。**教训：复刻保真审计「运动曲线」必须同时查 kf interp + 表达式两个维度。**（注：原「NewComposition 分数 fps cdta 时基」第二发现已根因化为 tdb4 @0x0C 并修复，见 [[ntsc-tickrate-derive-3x-off]]。）

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

- **showcase booyah-clone ④⑤⑥⑦⑨⑩⑪⑫ 待用户真机 review**（agent 已 render 自验；按 showcase review-gate，用户在真机打开 `booyah-clone.aep` 复核后才翻 complete）→ 逐 comp 状态见 [showcase/INDEX](showcase/INDEX.md)。⑩⑪⑫ 为本会话新建。

## Hanging Tasks

无。
