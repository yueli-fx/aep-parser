# workshop-workflow skill — design

**Date**: 2026-05-23
**Author**: claude (with 月离)
**Status**: draft — pending user review

## 1. Goal

把"AI 跟用户协作时怎么用 `workshop/` 目录"这套约定从 aep-parser 的 CLAUDE.md
里抽出来，做成**项目无关的可复用 skill**，让任何用 workshop/ 模式的项目
都能直接 inherit 工作流。

## 2. Scope

**In scope**:
- workshop/ 子目录命名 + 语义 + 文件命名约定
- 3 阶段工作流（entry / during / exit）+ 触发器
- 模板（scar / sketch / critique）+ 收敛规则
- 反垃圾门禁 + 反模式清单
- 跟其它 superpowers skill 的关系（brainstorming / writing-plans 的输出归宿）
- 跟项目自身 CLAUDE.md 的分工

**Out of scope**:
- aep-parser 项目特有规则（仍归 CLAUDE.md）
- 任何"内容应该怎么写" — 我们只管"写到哪 / 何时读"
- 完整 ADR / experiment 规范 — 这两个标记"未来扩展位"留白

## 3. Skill artifact

**位置**: `~/.claude/skills/workshop-workflow/` （user-level，全局 auto-discoverable）

**文件树**:
```
workshop-workflow/
├── SKILL.md                # 主入口 — lean (<500 词)
├── folder-semantics.md     # 10 个文件夹的详细语义（按需读）
├── templates.md            # scar / sketch / critique / board / INDEX 模板
└── exit-ritual.md          # 退出 brainstorm 决策树（按需读）
```

**为什么拆 4 个文件**:
- SKILL.md 每会话 auto-load → 必须 lean
- folder-semantics / templates / exit-ritual 是详情，按需读 — 不烧 context
- 跟 writing-skills 推荐的"分离 heavy reference"一致

## 4. SKILL.md 内容大纲

### frontmatter
```yaml
---
name: workshop-workflow
description: Use when working in a project with a workshop/ directory (or starting one) — guides session entry/work/exit using folder semantics, templates, and triggers
---
```

CSO 决策：description **只描写触发条件**，不写 workflow 摘要（writing-skills
明确警告：description 写 workflow 会让 Claude 跳过 skill 全文）。

### 主体段落（按读者扫描顺序）

1. **Core principle**（1-2 句）
   > workshop/ 是 AI 协作的工作台 — 按"何时读它"组织的命名空间。
   > 严格守门：**只写会改变未来行为 / 影响决策 / 反复引用**的内容。

2. **Entry checklist**（≤30 词）
   - 读 `workshop/board.md` "Next session 进来先做"
   - 跟 `git status` 对账，不一致 → 先核对
   - 一致 → 按 Next session 第一句执行

3. **During — 场景触发器表**
   | 场景 | 先看哪 |
   |---|---|
   | 找下一个任务 | board.md |
   | 不确定架构 | specs/ |
   | 跑测试 / 提交 | playbooks/ |
   | 奇怪行为 / 似曾相识的 bug | scars/ |
   | 拿不准方向 | reference/ + critiques/ |
   | 设计新功能 | invoke brainstorming → specs/ |
   | 拆任务 | invoke writing-plans → plans/ |

4. **Exit ritual**（精简，不强制 brainstorm）
   - 明显 case → 直接归类 + 写 + commit
   - **只在真模糊时**调 brainstorming（跨多文件夹 / 不知归哪）
   - 必更新 board.md `Last updated` + `Next session 进来先做`
   - 详细决策树：`exit-ritual.md`

5. **Folder map**（一句话一个 + 链接详细）
   见 `folder-semantics.md` for templates + edge cases

6. **写入约束 — 守门**（≤50 词）
   > workshop 只记会改变未来协作行为 / 影响决策 / 反复引用的内容。
   > 流水账 / 会话 byproduct / 一次性 debug 不算。严格守门。

7. **跟 CLAUDE.md 的分工**
   - **CLAUDE.md** = 项目铁律 + 场景触发器 + 项目特有规则（auto-load）
   - **workshop/** = 状态 + 知识 + 历史 + playbook（按需读）
   - 重叠时 **CLAUDE.md 优先**

8. **跨引用 superpowers skills**
   - `superpowers:brainstorming` → 输出 → `workshop/specs/`
   - `superpowers:writing-plans` → 输出 → `workshop/plans/`
   - `superpowers:executing-plans` 读 `workshop/plans/`

### 不放 SKILL.md 主体的内容（拆到子文件）

- scar 收敛规则 + 模板 + 升级路径 → `templates.md`
- sketch / critique 模板 → `templates.md`
- 文件夹未来扩展位（decisions / experiments / questions） → `folder-semantics.md`

## 5. Folder map（10 个 + 嵌套 finish/）

```
workshop/
├── INDEX.md            # 速查 + 子目录用途 + 模板 + 写入约束
├── board.md            # 当前看板
│
├── specs/              # 设计 doc（brainstorming 产出）
│   └── finish/         # ship 后归档
├── plans/              # 实施 plan（writing-plans 产出）
│   └── finish/
│
├── playbooks/          # 场景操作手册（commands + checklist + conventions）
├── scars/              # 错题集（含强收敛规则）
├── reference/          # 外部参考（竞品源码 / RFC / 博文剪藏）
│
├── sketches/           # 长期点子（无日期前缀，躺尸或升 spec）
├── critiques/          # 外部 AI 反馈（原文 + 处置）
└── wip/                # 会话临时 scratch（短命，用完 prune）

# 未来扩展位 — 不创建，需求出现再加：
# decisions/  experiments/
```

### 命名约定（folder-semantics.md 详细）

| 子目录 | 文件命名 |
|---|---|
| `specs/` | `YYYY-MM-DD-<feature>-design.md` |
| `plans/` | `YYYY-MM-DD-<feature>-plan.md` |
| `scars/` | `<topic>.md`（无日期前缀 — 错题是常翻资产） |
| `sketches/` | `<topic>.md`（无日期前缀 — 灵感不维护时序） |
| `critiques/` | `YYYY-MM-DD-<spec>-<reviewer>.md` |
| `playbooks/` | `<topic>.md`（verify / commit / re-fixture / ...） |
| `reference/` | `<source>-<topic>.md`（boltframe-shape-layer.md / rfc-6749.md） |
| `wip/` | 自由，短命 |

## 6. 三阶段工作流

### Entry（≤3 步）

```
1. cat workshop/board.md       # 读 Next session / Active focus / Last updated
2. git status + git log -3     # 对账
3. 一致 → 按 Next session 第一句执行
   不一致 → 跟用户核对再动
```

### During — 场景触发器（同 SKILL.md 表）

略，见 SKILL.md。

### Exit decision tree（详细在 exit-ritual.md）

```
对话快结束 / 用户说"收尾"等信号
↓
本会话有新知识 / 发现 bug 吗？
├─ 无 → 仅更新 board.md（如果 Active focus 推进了）→ commit
└─ 有 → 每条找归宿:
    ├─ 已机制化的教训 → scars/
    ├─ 一次性 / 流水账 → 不写（守门）
    ├─ 重复流程被发现 → playbooks/
    ├─ 跨多文件夹 / 真模糊 → invoke brainstorming 跟用户对话
    └─ 明确归一个文件夹 → 直接写
↓
更新 board.md `Last updated` + `Next session 进来先做`
↓
commit
```

**关键反摩擦原则**：90% case 是显然的，不需要 brainstorm。
默认 brainstorm = 高摩擦 = 跳过 = 知识丢失。

## 7. 模板（详细见 templates.md）

### scar 模板（采纳 YHFish 强收敛规则）

```markdown
# <一句话主题>

**症状**: 用户实际怎么观察到 / 报错文本
**根因** (**禁止** 写 "忘了 / 粗心 / 没注意" — 必须是 错误假设 / 错误模型 / 错误流程):
我假设 X 但实际 Y
**教训**: 下次具体怎么做（具体 action，非"小心一点"）

## [Case 2]  ← 第二次重犯时追加，不开新文件
...
```

**收敛规则**:
- 同主题重犯 → **不开新文件**，追加 `## [Case N]`
- 重犯阈值（3-5 次或单次极痛）→ **升级 CLAUDE.md 触发器表**
- 升级后顶部标 `状态: 已升级 → CLAUDE.md §X`，不删

### sketch 模板

```markdown
# <想法标题>

一句话灵感。

**触发**: 为啥这会儿冒出来 / 解决啥痛
**关联**: 跟哪个 spec/plan/scar 相关（可空）
**再看条件**: 啥情况下值得拿出来评估（可空）
```

灵感不维护 status — 要么变 spec（移 `specs/`），要么躺尸。

### critique 模板

```markdown
# <spec> — <reviewer> review

## 原文
<reviewer 完整反馈 — 黏贴一字不动，或顶上标"(用户转述)">

## 处置（跟用户一起填，可延后）
- **采纳**: 哪几条接受，改进 spec 的什么部分
- **拒绝**: 哪几条不采纳，简短为啥
- **延后**: 哪几条有道理但不是这轮 scope
```

长 review（>1000 词）用户可直接编辑文件粘贴原文，AI 只填"处置"段。

### board.md 模板

最小骨架（aep-parser 的 board.md 是 mature 版）：
```markdown
# Board — <project> 项目看板

**Last updated**: YYYY-MM-DD by <who> (<one-line state summary>)
**Active focus**: <current main thread>

## Next session 进来先做
1. <first concrete action>
2. ...

## 在飞 / 最近归档（≤ 2 周）
...
```

### INDEX.md 模板

参考 YHFish — 见 templates.md。**自动 / 手动维护**问题：默认手动（exit
时 AI 提醒用户加新条目；旧条目人审）。半自动选项 future work。

## 8. 反模式（写进 SKILL.md "Common Mistakes" 段）

| 反模式 | 后果 | 正解 |
|---|---|---|
| 同 fact 在 board + scar + spec 写 3 遍 | 漂移 / 信任崩塌 | 一个权威源，其它说"见 X" |
| sketches/ 当 wip 用（堆半成品） | 永远不收 | wip/ 用完 prune；sketches 只放未启动点子 |
| scars/ 写"忘了 / 粗心" | 教训空洞 | 强制 错误假设 / 模型 / 流程 |
| critiques/ 只贴原文不写处置 | 变垃圾堆 | 强制带处置段（采纳 / 拒绝 / 延后） |
| 每次 exit 全 brainstorm | 高摩擦 → 跳过 | 90% case 直接归类，模糊才 brainstorm |
| 把 tmp/ 当 workshop/ 子目录 | 垃圾混进 git track | tmp 放项目根 + gitignore |
| 流水账 / 会话 byproduct 进 workshop | 信噪比降 | 严格守门 — 只记影响未来的 |

## 9. 跟 CLAUDE.md 的分工

| | CLAUDE.md | workshop/ |
|---|---|---|
| Load 时机 | auto-load 每会话 | 按需读 |
| 内容 | 项目铁律 + 场景触发器表 + 工作风格 | 状态 + 知识 + 历史 + playbook |
| 例子（aep-parser） | "所有写都是 length-preserving"；"workshop/ 文档分工触发表" | board.md / scars/v2-2-aelayer-structure.md |
| 重叠时 | **优先** | 落地工具 |

**升级路径**：scar 重犯 → CLAUDE.md 触发器表加一行"做 X 前必读 scars/Y.md"

## 10. 跟其它 superpowers skills 的关系

- **`superpowers:brainstorming`**: 产出 design doc → 输出位置 default 是
  `docs/superpowers/specs/`，**workshop-workflow skill override 为
  `workshop/specs/YYYY-MM-DD-<topic>-design.md`**。需要在 SKILL.md
  里明确这条 override。
- **`superpowers:writing-plans`**: 输出 → `workshop/plans/`（同样 override）
- **`superpowers:executing-plans`**: 读 `workshop/plans/<topic>-plan.md`
- **`superpowers:finishing-a-development-branch`**: 推 spec/plan → `finish/`

## 11. Testing strategy

**Skill 类型**: Process skill（discipline-enforcing — 守门 + 触发器纪律）

**Baseline (RED)** — agent without skill：
1. 给 agent 一个有 workshop/ 但乱了的项目，问"该做什么"
   - 预期不读 board.md，自己猜 / 问用户
2. 给 agent 一个刚发现的 bug + 修法，问"该写哪"
   - 预期写进随便一个文件 / 不写
3. 让 agent 收尾，看会不会 brainstorm 一切 / 完全跳过

**With skill (GREEN)**:
- 1 → 读 board.md 第一
- 2 → 写 scars/ 用模板
- 3 → 明显归类直接写，只有模糊调 brainstorming

**Refactor**：testing 中发现 agent 找到的新借口 → 补 anti-pattern 表 + red flags

## 12. Open questions / future work

- **decisions/ 何时升级**：当前 ADR 散在 board + spec + scar + commit。是
  "什么时候开始痛得想统一"的问题。约定：单项目出现 ≥3 个跨 spec 决策需要回溯时
  开 decisions/。
- **experiments/ 何时升级**：aep-parser 现用 `tmp_debug/<probe>/main.go`。
  约定：当 probe 数据值得跨会话引用时升级（典型："这是我们验证过 AE 拒哪个
  byte 的全套数据"）。
- **INDEX.md 半自动**：当前手动；将来加 hook 自动 ToC 是优化项。
- **多语言项目** workshop/ 在 monorepo 怎么放：未触及，需要时再设计。

## 13. 实施清单（→ writing-plans 接手）

接下来调 `superpowers:writing-plans` 把上面拆成可执行 task 序列：
1. Scaffold `~/.claude/skills/workshop-workflow/` 目录
2. 写 SKILL.md（≤500 词）+ frontmatter
3. 写 folder-semantics.md（10 文件夹详细 + 命名约定）
4. 写 templates.md（scar / sketch / critique / board / INDEX）
5. 写 exit-ritual.md（决策树）
6. Test：跑 baseline + with-skill 验证 3 个场景
7. （可选）改 aep-parser 的 CLAUDE.md，引用 workshop-workflow skill 替代
   重复的触发器段（保留项目特有规则）

---

## Spec self-review（writing 后立即过）

- ✅ 无 TBD / placeholder（除明确标"未来扩展位 / future work"）
- ✅ 内部一致：folder list × 3 处出现（§5 / §7 / §8），命名一致
- ✅ Scope 单一：就是 workshop-workflow skill，不夹带其它
- ✅ 无歧义：每个 folder 唯一语义，每条 trigger 唯一目标

Pending：用户审。
