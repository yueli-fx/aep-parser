---
git: true
emit_agents_md: false
disabled_folders: []
disabled_gates: []
---

# flightdeck rules — aep-parser

## House rules

- **完成即 commit**：每做完一个可交付单元（feature slice / bugfix / 一批文档同步）就 commit，不要攒在工作树里。**本项目的任务进度靠 commit 历史追溯**——未提交的工作不算 landed，会话结束即丢失上下文锚点。无需用户每次点名，完成就提。
- **直推 main**：本项目无 PR 流，所有工作直接 commit 到 main（见 git log 既有惯例）。不要为常规改动单开分支。
- commit 拆分按逻辑单元：代码 + 其配套测试/fixture/文档同步算一个 commit；无关的 meta/文档改动单独 commit。commit body 可英文按既有惯例。
