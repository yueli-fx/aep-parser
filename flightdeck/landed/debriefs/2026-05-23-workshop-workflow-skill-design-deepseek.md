---
status: active
reviewed: landed/specs/2026-05-23-workshop-workflow-skill-design.md
last_updated: 2026-05-23
---

# workshop-workflow skill design — DeepSeek review

**Date**: 2026-05-23
**Reviewer**: DeepSeek (one of three external AI reviews on the spec)
**Reviewed**: [workshop-workflow skill design spec](../landed/specs/2026-05-23-workshop-workflow-skill-design.md)
**Outcome**: shipped to [Yuelioi/workshop@v0.2.0](https://github.com/Yuelioi/workshop)

## Raw feedback

你这个 workshop-workflow skill 的设计非常扎实，把"AI 协作工作台"的约定体系化、可复用化，思路很清晰。我仔细读完了整个 spec，下面从结构、实用性、可维护性、开源化四个维度给一些具体反馈。

整体评价：优秀，可以开干
优点说在前面：

拆分文件很克制：4 个文件，按加载场景分离，SKILL.md 保持极简，符合 writing-skills 的 heavy reference 分离原则。

守门规则强硬且正确："只写会改变未来行为/影响决策/反复引用的内容"这条原则是区分知识库和垃圾堆的核心，反模式表格也务实。

三阶段工作流 + 场景触发器表：这让你每次说"接着干活"时，AI 有明确的执行路径，而不是猜。

与 superpowers 其它 skill 的协作设计：specs/plans 作为 brainstorming/writing-plans 的输出归宿，finish/ 归档，闭环清晰。

scar 模板的强收敛规则：禁止"忘了/粗心"，强制写错误假设，重犯追加 case、阈值升级到 CLAUDE.md，这是这套系统最闪亮的部分。

下面是我认为可以再推敲的几个点，供你决定是否调整。

需要斟酌的几个设计细节
1. 入口清单的"对账"不够具体
读 workshop/board.md … 跟 git status 对账，不一致 → 先核对

"不一致"的定义对 AI 来说可能模糊。建议在 exit-ritual.md 或 SKILL.md 里给一个快速对账 checklist，例如：

board.md 的 Active focus 分支名与 git branch --show-current 是否一致？

board.md 的 Next session 进来先做 第一条是否在 git status 中有对应的未提交变更？

是否有 git stash 未恢复？

这能避免 AI 遇到"有一条 stash 但 board.md 没提"时不知所措。

2. Exit 决策树中"明显 case"与"模糊 case"的边界
你写了 90% case 是显然的，只有模糊时才调 brainstorming。但 AI 很容易把"跨两个文件夹"当成模糊，或者把本该写 scar 的错题当成流水账。建议在 exit-ritual.md 里加几条强制归类启发式：

如果信息是修复一个 bug 的方法 + 根因 → 必然归 scars/

如果信息是以后每次做 X 都要执行的步骤 → 必然归 playbooks/

如果信息是一次性的日志分析，不会再用 → 必然不写

只有同时满足"对多个文件夹都有影响"且"不确定哪个影响更持久"时，才算模糊，触发 brainstorming

另外，"跟用户对话"可以给一个 AI 提问模板："我发现 [信息]，可能适合放在 scars/ 或 playbooks/，你更倾向于放哪？还是都不放？"这样降低决策成本。

3. INDEX.md 的维护摩擦
设计写"默认手动（exit 时 AI 提醒用户加新条目）"。这可能在初期还好，但项目一大就容易忘。建议在 SKILL.md 或 templates.md 里提供一个AI 自动维护 INDEX.md 的轻量规则：

每次在 scars/、specs/、plans/ 下创建新文件后，自动在 INDEX.md 对应小节添加一行链接 + 一句话摘要。

如果担心 AI 乱加，可以规定 INDEX.md 每个小节有 <!-- AUTO-START --> ... <!-- AUTO-END --> 标记，AI 只在标记内插入，其余部分手工维护。

这样能保持"一眼看尽"的速查价值，又不会完全依赖人的纪律。

4. 反模式"同 fact 写 3 遍"的解法还可以更工具化
你说"一个权威源，其它说'见 X'"，但实践中 AI 容易忘记去引用。能不能在模板里就直接固化引用语法？比如在 board.md 里提到某个 scar，写成：

markdown
- 已知坑：[v2-aelayer 结构陷阱](scars/v2-aelayer-structure.md)
这样当 AI 读 board.md 时，引用本身就成了可点击的链接（对 IDE / Markdown 阅读器），AI 也能识别出这是一个必须跳转阅读的线索。建议在 templates.md 里给每个模板加上"如何引用其它文件夹"的说明。

5. 开源化需要补充的东西
目前这个设计是"内部 spec"，如果要作为开源项目发布，建议增加：

README.md：一句话讲清楚这个 skill 解决什么问题，30 秒看懂文件夹地图，给出安装命令（复制到 ~/.claude/skills/）。

LICENSE。

示例项目截图/录屏：一个真实的 workshop/ 目录 tree，展示几个关键文件内容片段，尤其是 board.md 和 scar 的 case 追加效果。

迁移指南：如果用户已有零散的 docs/notes/ 目录，怎么迁移到 workshop/ 结构。

与 Claude Code 的兼容说明：skill 的 auto-load 机制依赖 Claude Code 的 skills 目录配置，需要说明最低版本。

此外，SKILL.md 里的 "CSO 决策：description 只描写触发条件" 对开源读者是内部术语，可改成更通用的说明，比如"description 字段只写触发场景，不写工作流步骤，避免 AI 偷懒跳过正文"。

6. 与其它 AI 工具的对齐
你现在显然是针对 Claude Code / Claude 的 skills 机制做的。如果开源，可以顺便说明这个结构是 tool-agnostic 的约定层：即使不用 Claude，任何 AI 编码助手都可以通过项目级 instructions 引用这套目录规范。这样能扩大受众。

接下来可以做什么
你的实施清单已经列了 7 步，我建议在第 3 步和第 4 步之间加一个：

写一个 "5 分钟快速体验"脚本：生成一个最小 workshop/ 骨架，塞一个示例 scar、一个 board.md，让用户跑一遍 entry → 触发场景 → exit，直观感受价值。

然后再按你计划的 6 个 task 推进即可。

整体来说，这个设计已经可以直接进入实施，上面几点只是锦上添花。非常期待看到这个 skill 开源——它会解决很多 AI 编码用户"断片"的痛点。如果你需要，我可以帮你起草 SKILL.md 正文、templates.md 或者快速体验脚本的初稿。

## Disposition

**Adopt**:
- **Concrete entry-checklist commands**: `git branch --show-current` / `git stash list` / `Last updated` vs latest commit time. Shipped: `skills/workshop-workflow/SKILL.md` entry checklist (4 explicit reconciliation questions).
- **Exit-ritual classification heuristics + AI-asks-user template**: first-match-wins rules (bug+rootcause→scars, repeated procedure→playbooks, one-off→don't write) + structured ask template for ambiguous cases. Shipped: `skills/workshop-workflow/exit-ritual.md` "Classification heuristics" + "AI-asks-user template".
- **`INDEX.md` AUTO-START / AUTO-END markers**: AI maintains marked sections, human curates the rest. Shipped: `skills/workshop-workflow/templates.md` INDEX section + `scaffolds/full/workshop/INDEX.md`.
- **Cross-folder reference syntax in templates**: explicit example of `[name](scars/X.md)` links in board / INDEX templates. Shipped: `skills/workshop-workflow/templates.md` "Cross-folder reference syntax" section.
- **5-min experience quickstart**: README "Usage" section + Day-1 bootstrap example (`scaffolds/minimal/` + `install.sh --scaffold=minimal`). Shipped: `README.md` Usage section (v0.2.0).
- **Tool-agnostic clarification** (point 6): adopted progressively — v0.1.0 added a 1-line note; v0.2.0 made the skill content fully tool-neutral (all `CLAUDE.md` references replaced with generic "project agent rules") + added `.codex-plugin/`, `.cursor-plugin/`, `gemini-extension.json` + `GEMINI.md` manifests.

**Reject**: none.

**Defer**: none — all DeepSeek items shipped or scheduled.
