# Cockpit — aep-parser

Updated: 2026-06-20 · claude · Stage: Booyah 复刻 · Phase 1（comp ① render 主体对齐，残留中间帧）

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

RE comp ① 中间帧关键帧保真残留（t=0.3/t=0.8 与原工程差；lhd3 容量已对+全 linear，需逐 rect valueAtTime 细diff）+ 补 AE 2020 侧 render，然后 Task 1.2 ② テキスト → [plan](plans/2026-06-19-booyah-glitch-replication.md)（进度细节见其 `## Progress`）

## In Progress

<!-- AUTO:inprogress -->
- [2026-06-18-fx-technique-internalization.md](specs/2026-06-18-fx-technique-internalization.md) — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techni…
- [2026-06-18-technique-ontology.md](specs/2026-06-18-technique-ontology.md) — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized…
- [2026-06-19-booyah-glitch-full-replication.md](specs/2026-06-19-booyah-glitch-full-replication.md) — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle 表达式/Curves)用咱们的 Go API 从零完整…
- [2026-06-19-booyah-glitch-replication.md](plans/2026-06-19-booyah-glitch-replication.md) — 实现 Booyah Glitch 全工程复刻:Phase0 前置 spike(表达式 Evolution=time*N + wiggle / 单层 ~21 ma…
<!-- /AUTO -->

## Key Context

- comp ①：✅ AE 2025 接受 + shape 不 drop（verify 实证）；✅ render 主 bug 修复（层 position (0,0)→合成中心，4f579d1，t=0 帧与原工程字节一致）；🔶残留 >4kf rect 关键帧（见 Next）。根因档 `incidents/shape-layer-position-default-offscreen.md`。
- AE 自动化基建（arc 全程关键）：`clear_ae_crashstate.ps1` tracked 工具（`tools/debug/`，67023b6）+ **ae_run `Invoke-SendKeysSafe` 改 PostMessage 根治 foreground-lock**（b411d09）→ 所有 in-run modal 在前台游戏锁下都能无人值守消化（实证原工程 Resolve Fonts 框）。
- ⚠ 预存 RED 单测（非本 arc）：`TestSynthControlEntries_PardLayout/label`（pseudo-effect label pard @0x04=0x0 want 0x20）——pseudo spec 已归档收工但此 unit test 红，待单独修。
- operator 在另一台机器、前台游戏窗口 ≠ 用户在用 → 不问用户让机器（`incidents/ae-automation-occlusion-crashstate.md` Case 2c）。
- 上游「理解工程」研究 arc（暂让位）进度归 `specs/2026-06-18-fx-technique-internalization.md` § 现状与下一步 + `technique-ontology` § 9。

## Pending Review

- (none)

## Hanging Tasks

无。
