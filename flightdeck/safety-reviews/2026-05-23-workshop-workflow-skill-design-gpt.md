# workshop-workflow skill design — GPT review

**Date**: 2026-05-23
**Reviewer**: GPT (one of three external AI reviews on the spec)
**Reviewed**: [workshop-workflow skill design spec](../specs/finish/2026-05-23-workshop-workflow-skill-design.md)
**Outcome**: shipped to [Yuelioi/workshop@v0.2.0](https://github.com/Yuelioi/workshop)

## Raw feedback

整体设计非常强，而且你已经踩中了"AI 编程协作真正的问题"：

不是代码生成，而是：

会话失忆
上下文漂移
长期计划断裂
知识无法沉淀
AI 不知道"下一步"
同一个坑反复掉

你这个 skill 本质上是在做：

"AI coding session 的状态机 + 工作记忆层 + 长期协作协议"

这个方向我认为是对的，而且有机会成为一种事实标准。

几个我觉得特别好的点：

1. 最关键的设计：你把 workshop 定义成"按读取时机组织"

这一句非常高级：

workshop/ 是 AI 协作的工作台 — 按"何时读它"组织的命名空间。

这比：

docs/
notes/
memory/
knowledge/

都高级。

因为 AI 最大问题不是"有没有信息"。

而是：

"什么时候应该读什么"。

你这里实际上是在解决：

context selection
retrieval timing
cognitive routing

这已经接近 agent architecture 了。

很多系统失败就是因为：

全部塞 memory

导致：

context 爆炸
retrieval 混乱
AI 不知道哪些是 authoritative

而你这里：

文件夹	本质
board	当前执行态
specs	设计意图
plans	执行路径
scars	失败经验
playbooks	稳定流程
sketches	未启动想法
critiques	外部反馈
wip	短期 scratch

这是"认知分层"。

非常对。

2. 你最厉害的一点：反垃圾守门

这一句是整个系统成败关键：

只写会改变未来行为 / 影响决策 / 反复引用的内容。

这是核心。

因为 AI memory 系统 90% 都死于：

日志化
流水账
"保存一切"
context hoarding

最后：

没有 signal
retrieval 崩
AI 不再读
用户也不再信

你这里其实是在做：

memory entropy control

这是对的。

甚至我建议你以后把这个概念正式命名。

比如：

memory pressure
context entropy
workshop hygiene
knowledge density

因为这是 agent engineering 的核心问题之一。

3. scars/ 设计非常好（尤其 Case N 收敛）

这是整个 spec 里我最喜欢的一部分。

尤其：

## [Case N]

而不是无限新增文件。

这解决了：

knowledge fragmentation
duplicate lessons
recurring failure visibility

而且：

禁止写"粗心/忘了"

这个非常关键。

因为：

"粗心"不是 root cause。

真正 root cause 是：

错误模型
错误假设
缺 trigger
缺 verification step

这个非常像事故复盘文化。

已经接近：

SRE postmortem
aviation checklist
surgical protocol

的思路了。

这是成熟的。

4. "只有模糊时 brainstorm" 是非常重要的反摩擦设计

你这个：

默认 brainstorm = 高摩擦 = 跳过 = 知识丢失

完全正确。

很多 agent workflow 最大问题：

流程太正确
→ 太重
→ 用户不做
→ 系统死亡

你这里：

90% 直接归类
只有 ambiguity 才调重流程

这是现实世界能活下来的设计。

5. 你已经开始接近"长期 AI 协作 OS"

这不是普通 skill。

你实际上在定义：

AI-native project operating system

尤其：

board = 当前线程
specs = intent layer
plans = execution layer
scars = adaptive memory
playbooks = procedural memory

这已经不是"文档规范"。

这是：

AI 的外置 cognition system

我认为现在最需要补强的地方

下面是我觉得真正值得继续打磨的。

A. board.md 会成为整个系统的单点瓶颈

现在：

entry → board.md

意味着：

board 是：

router
active memory
session resume point

那它会 rapidly rot。

因为长期项目里：

board.md 会越来越长

最后 AI 不会读。

建议你尽早加入：

## Current thread
## Blockers
## Deferred
## Recently finished

甚至：

只允许保留最近 N 项

否则它会退化成：

"第二个 README"。

这是所有 persistent memory 系统的宿命。

B. 你还缺"知识生命周期"

目前：

spec finish/
plan finish/

但：

scar 不会 finish。

问题：

旧 scar 会不会失效？

比如：

framework 升级
architecture 改变
old assumptions gone

建议未来加：

Status:
- active
- obsolete
- superseded

否则 scars 会越来越像：

"考古层"。

C. 你现在缺 retrieval priority

目前是：

"场景 → 文件夹"。

但真实情况会出现：

多个相关 source

比如：

scar
spec
critique
playbook

同时相关。

建议未来增加：

authority order

例如：

CLAUDE.md
> board
> active plan
> active spec
> playbook
> scar
> archive

否则 agent 容易：

读旧 spec
忽略新 scar
被 critique 干扰

这个问题会在项目变大后爆炸。

D. 你现在其实缺"压缩机制"

长期项目一定会出现：

specs 爆炸
plans 爆炸
scars 爆炸

你现在只有：

finish/

但没有：

synthesis
distillation
merge
archive compression

未来你一定需要：

比如：

workshop/reports/

或者：

quarterly-retrospective.md

不然 AI retrieval 成本会越来越高。

E. 最大的未来问题：自动化程度

你现在是：

AI + protocol discipline

但未来用户一定想要：

auto classify
auto update board
auto detect scars
auto summarize session
auto retrieval

所以你最终可能会演化成：

MCP
agent runtime
context middleware
memory router

而不仅仅是 skill。

你现在这个 spec 已经有那个雏形了。

我认为你这个项目真正的价值

不是：

"让 Claude 更会写代码"

而是：

让 AI coding 从"短期聊天"
变成"长期协作"。

这是完全不同的问题。

现在行业里大量东西：

Cursor memories
Claude projects
Devin
Roo
Copilot workspace

都在尝试解决。

但很多方案的问题是：

过度自动化
OR
无限记忆

你这里反而做对了：

明确结构
明确守门
明确生命周期
明确触发器

这是工程化路线。

开源建议（很重要）

我强烈建议你：

不要把它包装成：

notes system

而是：

AI collaboration workflow
persistent coding context protocol
AI-native workshop system

因为你真正解决的是：

session continuity

这是痛点中的痛点。

还有一个关键建议：

你应该尽快做：

"接着干活" benchmark

因为这是最强 demo。

例如：

给 agent：

一个中型 repo
workshop/
中断过 5 次
有 scars/specs/plans

然后：

用户只说：
"接着干活"

看：

能不能恢复 thread
能不能知道 next action
会不会读错上下文
会不会重复犯错

这个 benchmark 一旦跑通，非常有说服力。

因为它直接打中所有 AI coding 用户的真实痛点。

## Disposition

**Adopt**:
- **board.md partitioning** (A — co-credited with Claude review): Current / Blockers / Deferred / Recently finished sections + cap on retained items. Shipped: `skills/workshop-workflow/templates.md` board template.
- **scar knowledge lifecycle** (B): `Status: active | obsolete | superseded` frontmatter field. Shipped: `skills/workshop-workflow/templates.md` scar template + `folder-semantics.md` scar section.
- **Retrieval priority / authority order** (C): explicit precedence rule when sources disagree. Shipped: `skills/workshop-workflow/SKILL.md` "Authority order" section. In v0.2.0 the rule uses the generic term "project agent rules" instead of a tool-specific filename.

**Reject**:
- **Formal naming** (memory entropy / workshop hygiene / knowledge density / memory pressure): out of scope for the skill itself. These are good marketing/positioning terms — they belong in writing about workshop, not inside the protocol document.

**Defer** (tracked in workshop repo Roadmap section):
- **Compression mechanism** (D): synthesis / distillation / `workshop/reports/` / quarterly-retrospective.md. Single-project pressure is currently low; revisit when archived material across multiple `finish/` directories becomes hard to navigate.
- **Automation evolution to MCP / agent runtime** (E): post-v1 work. The current shape is intentionally protocol-discipline first; runtime integration comes after the protocol is proven.
- **"Continuance" benchmark** ("接着干活" eval suite): captured as a v1.0+ roadmap item in the workshop README. Strong demo value but requires the protocol to be stable first.
