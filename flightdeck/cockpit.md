# Cockpit — aep-parser

Focus: AEP 理解/复刻诊断/生成路线已开新 spec → `work/aep-understanding-generation/index.md`。Booyah Glitch 全工程复刻仍是当前压力测试基准 → `work/booyah-glitch-replication/index.md`。

## In flight

- **booyah-glitch-replication** (`work/booyah-glitch-replication/index.md`) — 全 12-comp from-scratch 复刻 ✅ 全建成 + 双版本 AE gated。①②③⑧ 已用户真机验收 complete；④⑤⑥⑦⑨⑩⑪⑫ 🔶待用户真机 review（agent 已 render 自验，按 showcase review-gate 用户验后才翻 complete + 关 plan）。逐 comp 账本（建法/值/delta/commit）= `showcase/INDEX.md`，进度 = `work/booyah-glitch-replication/index.md` → `plan.md`。
- **aep-understanding-generation** (`work/aep-understanding-generation/index.md`) — 新方向 spec：把 `aepdissect -json` / `Project.WriteJSON` / Booyah 渲染脚本 / 能力表收敛成 profile → diff → render oracle → gap ledger → recipe generation 的长期流水线。Phase 1 已抽 `internal/profile`，`aepdissect -json` 已切到 stable profile builder；自审补齐 track matte refs + parsed in/out points（2051e04）。Phase 2 已实现 `internal/profilediff` / `cmd/aepdiff` / JSON ignore rules（adfb7da）。Phase 3 已实现 `internal/aeoracle` / `cmd/aeoracle` / generic JSX render harness（053b3fc）。Phase 4 已实现 `internal/gapledger` / `cmd/aepgaps`；Phase 5 已实现 `internal/sliceworkflow` / `cmd/aepslices`，fixture + real Booyah source + non-Booyah Motionbox sample 验收通过（见 `phase5-booyah-acceptance.md` / `phase5-non-booyah-acceptance.md`），非 Booyah 样本已过 AE 2025 hard render gate。Phase 6 frame-set render compare 已实现并在 Booyah source-vs-clone 跑通；minimal recipe IR 已实现并通过 AE 2025 render gate；recipe capability report 已接 `internal/capindex`；built-in effect + static effect params + embedded expected-profile contract 已过 profile + AE 2025 render gate；shape stroke/detail + Trim Paths filter + text style slices（font size/tracking/justification/fill/stroke/faux）+ `expected_profile.keyframes[]` Position/Anchor/Scale/Rotation/Opacity keyframe contract 已过 `go test ./...`、`go vet ./...` 和 AE 2025 render gate，并修正 `SetLayerTransform` stale boundary（见 `phase6-recipe-ir-acceptance.md`）。
- **apidoc-tag-schema** (`work/apidoc-tag-schema/index.md`) — @tag 文档 schema 已落地（facade 全转 + dual-read + jargon scrub COMPLETE）。仅剩**可选遗留** flip(c)：删 `extract.go` `parseCapTag` 旧读路径 + `tag.go` 旧枚举 map + 守卫测试 + `--validate` strict CI + 改文档真相源注脚 → regen → commit flip → done。dual-read 现仍工作、aep:cap=0，非阻塞，可随时做。
- **fx-technique-internalization** (`work/fx-technique-internalization/index.md`) — 上游「理解工程」研究 arc（暂让位）。通用机制：任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存两层结构（技法库 `knowledge/techniques/fx-techniques.md` + 现象配方）。火焰=实例#1已验证。下一步见 topic index。
- **technique-ontology** (`work/technique-ontology/index.md`) — 同上游 arc 的数据骨架：角色/技法/机制三轴本体 + schema v2。已经火焰/闪电/控制器三类压测；未验证大规模（N≫3）词表对齐。

## Next

- **aep-understanding-generation 下一步**：recipe IR 已能 one comp + solid/text/shape + transform + supported built-in effect + static effect params + shape stroke/detail + Trim Paths + profile-visible text style + Position/Anchor/Scale/Rotation/Opacity keyframe profile checks 编译 AEP，且 report 会列 used capabilities / downgrades / refusals / profile_checks。下一刀继续补更多 shape filters、表达式或 transform keyframe/ease（以 writer 支持为边界）；更多 text style 只补 profile 可读字段。不要开 automated correction loops。
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
- **pseudo label 红测已清**：`TestSynthControlEntries_PardLayout/label` 已按当前知识结论修正（普通 label @0x04=0，Dimmed label @0x04=0x20）；`go test ./...` 绿。
- **能力真相源** = `go run ./cmd/capindex -q <词>`（不再维护单独 coverage 表）。
