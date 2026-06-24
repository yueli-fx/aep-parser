# Cockpit — aep-parser

Focus: Booyah Glitch 全工程从零复刻 = 理解金标准检验 → `work/booyah-glitch-replication/`。全 12-comp 已建成 + 双版本 AE gated；待用户真机验收 ⑩⑪⑫。

## In flight

- **booyah-glitch-replication** (`work/booyah-glitch-replication/`) — 全 12-comp from-scratch 复刻 ✅ 全建成 + 双版本 AE gated。①②③⑧ 已用户真机验收 complete；④⑤⑥⑦⑨⑩⑪⑫ 🔶待用户真机 review（agent 已 render 自验，按 showcase review-gate 用户验后才翻 complete + 关 plan）。逐 comp 账本（建法/值/delta/commit）= `showcase/INDEX.md`，进度 = `work/booyah-glitch-replication/plan.md` `## Progress`。
- **apidoc-tag-schema** (`work/apidoc-tag-schema/`) — @tag 文档 schema 已落地（facade 全转 + dual-read + jargon scrub COMPLETE）。仅剩**可选遗留** flip(c)：删 `extract.go` `parseCapTag` 旧读路径 + `tag.go` 旧枚举 map + 守卫测试 + `--validate` strict CI + 改文档真相源注脚 → regen → commit flip → done。dual-read 现仍工作、aep:cap=0，非阻塞，可随时做。
- **fx-technique-internalization** (`work/fx-technique-internalization.md`) — 上游「理解工程」研究 arc（暂让位）。通用机制：任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存两层结构（技法库 `knowledge/techniques/fx-techniques.md` + 现象配方）。火焰=实例#1已验证。下一步见文件 § 现状与下一步。
- **technique-ontology** (`work/technique-ontology.md`) — 同上游 arc 的数据骨架：角色/技法/机制三轴本体 + schema v2。已经火焰/闪电/控制器三类压测；未验证大规模（N≫3）词表对齐。

## Next

- **用户真机验收 ⑩⑪⑫**（修后再验；按 review-gate 用户验后翻 complete + 关 booyah plan Task 4.3）。⚠AE 装 `E:\adobe\`，agent 自跑 `scripts/ae_run.ps1`。
- **booyah polish（非阻塞 documented delta）**：撕裂强度/文字色细调 · **eased Position kf**（LayerTransform 无 eased builder）· **Noise2 能力补**（不在 embed set）· **Curves 曲线数据**（arbitrary-data blocked）。
- **apidoc flip(c)** 可选遗留清理（见 In flight）。
- **旁支**：补 comp ① AE 2020 render；「gen 漏复刻层属性/标志」横切回扫——已撞两类活跳坑：①静态 Scale/Rotation/Anchor（`knowledge/layer/layer-replication-drops-static-transform-channels.md`）②motion blur 层标志（2026-06-22 frame26 揪出，已修 `SetMotionBlur`+`SetCompMotionBlur`）。

## Open questions

- **comp ①**：✅✅ AE 2025 render 与原工程像素级一致（中间帧 diff=0/0/32px，32px=sub-pixel 帧snap）。三个真 bug 全修：层 position→中心（4f579d1）· 层时长（6091016）· 中间帧错位根因=`deriveTickRate` 把 NTSC kf 时间读大 3×（根因档 `knowledge/composition/ntsc-tickrate-derive-3x-off.md`）。⚠AE 2020 侧 render 待补；clone 用 30fps 绕开 NewComposition 分数 fps cdta 时基 bug。
- **终帧保真审计（教训）**：复刻保真审计「运动曲线」必须同时查 kf interp + 表达式两个维度。kf ease 线程已查证证伪关闭（⑤⑥⑦⑧ Position/Opacity 全 394 kf 都 linear，唯一带 bezier ease 的是文字动画器 Tracking/CharOffset，② 早已拷）；但表达式维度揪出 ⑤L2/⑧ wiggle 漏复刻，已全扫 ⑤⑥⑦⑧ transform 表达式补齐。
- **复刻 = 理解压测定位**（用户校准 2026-06-21）：复刻不为产物，为逼出读侧漏解属性 / 写侧缺 API / 未知结构；终极=任给一个 .aep 能说清构成+内容+cover 边界。每 comp 先读侧全审计→列 gap→建可建的→诚实标缺口。
- **本会话沉淀**：库 fix `setLdtaFrac` 粗 divisor 截断（commit 2886c5e + 回归测试）· `knowledge/keyframe-expression/expression-enable-byte-pair.md`（kf+expr 坑扩到 effect param + 级联 drop，booyah ⑪ 实证）· `knowledge/shape/mask-shph-bbox-read-gap.md`（mask 真几何在 shph bbox，`Mask.Vertices` 只 surface 归一化 ldat）· Displacement Map tdpi = self-ref（`AddEffect` 默认绑宿主）。
- **AE 自动化基建**：`clear_ae_crashstate.ps1` tracked 工具（`tools/debug/`，67023b6）+ ae_run `Invoke-SendKeysSafe` 改 PostMessage 根治 foreground-lock（b411d09）。operator 在另一台机器、前台游戏窗口 ≠ 用户在用 → 不问用户让机器（`knowledge/workflow/ae-automation-occlusion-crashstate.md` Case 2c）。
- **⚠ 预存 RED 单测（非本 arc）**：`TestSynthControlEntries_PardLayout/label`（pseudo-effect label pard @0x04=0x0 want 0x20）——pseudo spec 已收工但此 unit test 红，待单独修。
- **能力真相源** = `go run ./cmd/capindex -q <词>`（不再维护单独 coverage 表）。
