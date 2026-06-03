---
version: 3.0
disabled_folders: []
---

# flightdeck rules — aep-parser

## House rules

### Project conventions

- **完成即 commit**：每做完一个可交付单元（feature slice / bugfix / 一批文档同步）就 commit，不要攒在工作树里。**本项目的任务进度靠 commit 历史追溯**——未提交的工作不算 landed，会话结束即丢失上下文锚点。
- **直推 main**：本项目无 PR 流，所有工作直接 commit 到 main（见 git log 既有惯例）。不要为常规改动单开分支。
- commit 拆分按逻辑单元：代码 + 其配套测试/fixture/文档同步算一个 commit；无关的 meta/文档改动单独 commit。commit body 可英文按既有惯例。

### Autonomy overrides
<!-- migrated from commit_mode:confirm — 旧 toggle 取默认值 confirm，但上方 House rule「完成即 commit·无需用户每次点名」是既定的真实意图，故按 auto 编码 -->
commit without asking

- **大计划完成即自动 landing**：当一个代表已 ship 功能 arc 的 plan 翻到 `status: done`，主动跑 `/flightdeck:landing` 把它归档到 `landed/plans/` 并同步 `landed/HISTORY.md` + INDEX + cockpit，无需用户点名。判定「大计划」= 有独立 plan 文件、对应一个可交付 feature/结构性 arc（如 P3 §3G）；滚动 reference 矩阵（coverage*.md）等长期文档不适用，保持 active。
