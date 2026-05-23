# workshop-workflow skill design — Claude review

**Date**: 2026-05-23
**Reviewer**: Claude (one of three external AI reviews on the spec)
**Reviewed**: [workshop-workflow skill design spec](../specs/finish/2026-05-23-workshop-workflow-skill-design.md)
**Outcome**: shipped to [Yuelioi/workshop@v0.2.0](https://github.com/Yuelioi/workshop)

## Raw feedback

20:04
这个设计相当扎实，能看出来是在真实项目里磨出来的，不是凭空想的。几个具体看法：

设计里做对的事

"Entry checklist ≤3步" 的约束好。AI经常在session开头乱逛，强制读 board.md 再 git status 再执行，这个顺序对。
scar 模板禁止写"忘了/粗心"这条非常关键。这是大多数错题本的致命伤——根因全是废话，下次照样犯。
Exit 的"90% case直接归类，模糊才brainstorm"是反摩擦设计，符合实际使用节奏。高摩擦=跳过=知识丢失，这个链条说得准。
4文件拆分的逻辑清晰：SKILL.md lean保证auto-load不烧context，重的内容按需读。
几个值得考虑的问题

board.md的信任危机：整个entry流程建立在board.md是最新的这个假设上。但实际上最容易烂的就是这个文件——上次session因为各种原因没更新，进来读到的是过期状态，然后按错误的"Next session"执行。你有 git status 对账，但对账的是代码状态，不是任务状态。或许需要一个"board.md可能过期"的fallback：比较 Last updated 跟 git log 最近commit时间，如果超过N天就警告。

wip/ 的生命周期没有机制保障：设计里说"短命，用完prune"，但靠自觉prune的东西基本都会堆着。可以考虑加个约定：每次exit时wip/里超过1个session的文件要么归类要么删，不能带着过夜。

critiques/ 的处置段"可延后"是个漏洞：延后容易变成永远不填。或许可以规定：没有处置段的critique文件在board.md里要有一条悬挂任务，不能关闭session直到处置。

跨项目使用的现实挑战：skill放 ~/.claude/skills/ 是user-level全局，但不同项目的workshop/深度差异很大。有些项目可能只有board.md+scars/，没有specs/plans/。trigger表里写"设计新功能 → invoke brainstorming → specs/"，但如果这个项目没有specs/约定怎么办？或许SKILL.md里需要一段"minimal setup"：只有board.md也能用，其它目录需要时再建。

关于开源

这类"AI协作工作流"的skill，开源的核心价值不是代码，是设计文档本身——就是你这份spec。建议：

README直接放这份设计的精简版，让人一眼明白why，不只是how
附一个真实项目的workshop/快照（脱敏），比任何说明都有说服力
考虑加个"如何验证skill在你项目里生效"的quickcheck，就是你§11 testing strategy的用户版
整体设计是可以直接实施的，open questions里的decisions/和experiments/留白处理得克制，没有过度设计。

## Disposition

**Adopt**:
- **board.md rot mitigation** (co-credited with GPT review): partition into `Current` / `Blockers` / `Deferred` / `Recently finished` sections + stale check (compare `Last updated` vs most-recent commit time). Shipped: `skills/workshop-workflow/SKILL.md` entry checklist + `templates.md` board template.
- **`wip/` lifecycle mechanism**: every session exit must classify or delete any `wip/` file older than the current session, no carry-over. Shipped: `skills/workshop-workflow/exit-ritual.md` "Hanging tasks — block session exit".
- **Critique disposition non-deferrable**: if a critique exists without a disposition, add a hanging task to `board.md` and the session can't close cleanly. Shipped: `skills/workshop-workflow/templates.md` critique rules + `exit-ritual.md`.
- **Minimal setup**: only `board.md` required; other folders added on demand. Shipped: `skills/workshop-workflow/folder-semantics.md` "Minimal vs full setup" + `scaffolds/minimal/workshop/board.md`.
- **OSS bootstrap (README + LICENSE)**: shipped in v0.1.0. `README.zh.md` added v0.2.0. Real-project workshop snapshot deferred to v1.0+ release prep (in roadmap).

**Reject**: none.

**Defer**: none from this review specifically — the bootstrap-screenshot item is tracked in the workshop repo roadmap.
