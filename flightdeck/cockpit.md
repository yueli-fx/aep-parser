# Cockpit — aep-parser

Updated: 2026-06-20 · claude · Stage: Booyah 复刻 · Phase 1（comp ① 待 AE 验收）

Focus: Booyah Glitch 全工程复刻 = 理解金标准检验 → [spec](specs/2026-06-19-booyah-glitch-full-replication.md)

Pointers: config → rules.md · 通用铁律/风格 → CLAUDE.md · 能力真相源 → `go run ./cmd/capindex -q <词>` · artifacts → 各 folder INDEX · history → archive/

## Next

重建 `clear_ae_crashstate.ps1` 为 tracked 工具（`tools/debug/`）清崩溃标志 → warm-retry verify.jsx 验 AE 接受 comp ① + shape 不 silent-drop → render 对照原工程 → [plan](plans/2026-06-19-booyah-glitch-replication.md)（进度细节见其 `## Progress`）

## In Progress

<!-- AUTO:inprogress -->
- [2026-06-18-fx-technique-internalization.md](specs/2026-06-18-fx-technique-internalization.md) — 通用机制:任何参考 .aep → 解析 → 拆角色 → 抽跨域技法原子 → 存 flightdeck 两层结构(技法库 references/fx-techni…
- [2026-06-18-technique-ontology.md](specs/2026-06-18-technique-ontology.md) — 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized…
- [2026-06-19-booyah-glitch-full-replication.md](specs/2026-06-19-booyah-glitch-full-replication.md) — 把整个 Booyah Glitch 真实工程(12 comp/61 层/~100 mask/wiggle 表达式/Curves)用咱们的 Go API 从零完整…
- [2026-06-19-booyah-glitch-replication.md](plans/2026-06-19-booyah-glitch-replication.md) — 实现 Booyah Glitch 全工程复刻:Phase0 前置 spike(表达式 Evolution=time*N + wiggle / 单层 ~21 ma…
<!-- /AUTO -->

## Key Context

- comp ① シェイイイイプ：Go 建好（e4d266c），4 rect kf + fill 色逐值对账原工程 ✓；AE-accept/render 待验。
- AE 验证前置：`clear_ae_crashstate.ps1` 已丢，需 tracked 重建（`tools/debug/`，PostMessage `VK_RETURN` 到 safe-mode hwnd，勿 SetForegroundWindow）；verify.jsx 已修（纯字符串、末尾一次写）。详 `incidents/ae-automation-occlusion-crashstate.md` Case 2c。
- operator 在另一台机器、前台游戏窗口 ≠ 用户在用 → 不问用户让机器。
- 上游「理解工程」研究 arc（暂让位）进度归 `specs/2026-06-18-fx-technique-internalization.md` § 现状与下一步 + `technique-ontology` § 9。

## Pending Review

- (none)

## Hanging Tasks

无。
