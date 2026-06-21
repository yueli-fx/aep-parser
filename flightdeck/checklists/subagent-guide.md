---
status: active
last_updated: 2026-06-20
when_to_read: before deciding to use subagent
applies_to: [subagent, api, claude-in-claude, artifact, autonomous, context, skill, system-prompt]
synced: true
---
# Subagent 使用指南

> 语境 = Claude Code 的 **Task / Agent 工具**(及 SDD 等多 agent 流程)。subagent 不是「调原始 API」,
> 是主对话派出去的一个**全新、隔离上下文**的 agent;它的最终消息**只回给你**,不直接给用户。

## 核心原则

Subagent **无状态、上下文隔离**:每次派发都是全新会话,不继承主对话的任何东西。

> **你在 prompt 里显式写进去的 = 它全部所知;其余一概不知。**

## 它不会自动带入

| 内容 | 自动带入? |
| --- | --- |
| 当前对话历史 | ❌ |
| `CLAUDE.md`(项目 / 全局) | ❌ |
| **flightdeck** 的 deck(cockpit / rules / checklists / spec / plan / 约定) | ❌ |
| 你已加载的 skill | ❌(它只有 workflow-subagent 自己的系统提示) |
| codebase 结构 | ⚠️ 它能读文件,但不知道读哪、为什么 —— 要指明 |
| 你在 prompt 里显式写的 | ✅ |

→ 所以**项目的「必带上下文」必须手动塞进 prompt**;靠记忆漏带,subagent 只会按通用默认习惯产出,
偏离项目规范而不自知。

## 什么时候用 ✅

上下文随任务一次给全的**自包含**活:
- 总结 / 翻译 / 分类 / 抽取 / 格式转换 / 生成结构化 JSON —— 数据本身即全部上下文
- **并行 fan-out**:多文件审查、多角度搜索、跨子系统调研 —— 各 agent 拿自己那份完整输入
- 读密集调研「读一堆只回结论」—— 用只读的 `Explore` / `general-purpose` agent,省主线 context

## 什么时候不用 ❌(除非把上下文全注入)

- **写代码** —— 要项目风格 / 架构约束 / 编码 / 注释 / 提交规范,这些都不会自动传
- 要遵守复杂 skill / 规则 —— 同上,规则不自动带
- 多轮迭代、要累积记忆 —— 每次失忆;改走主线,或拆成逐 task 派发

## 非派不可时:brief 必带清单

派出去前,把它做对所需的一切**显式**写进 prompt(SDD 里就是 implementer / reviewer 的 brief):
1. **任务本身** —— 那一个明确子任务(大计划抽成单 task 文件,别贴整份)
2. **项目规范** —— 该项目的编码约定 / 注释规范 / 提交规范(逐字给,别只丢文件名让它自己找)
3. **接口契约** —— 它要用到的、上游产出的签名 / 类型
4. **验证方式** —— 用项目自己的 build / test 当真相源,别让它信 IDE 诊断面板(常 stale)
5. **reviewer 带同一份规范** —— implementer 漏了,review 关还能拦

**大块产物走文件、不贴正文**:落成文件传**路径**,subagent 一个 Read 拿全,主线 context 不被占。

## 最佳实践

- 能主线做就主线做;**非派不可**才派,且**约束全注入**
- **模型分级**:机械活用便宜模型;判断 / 架构 / 终审用强模型 —— 每次显式指定
- **结构化输出**:让它只回 JSON / schema,避免解析歧义
- **一份 dispatch 只描述一个任务**:别把前几个任务的总结灌进去
- **token 意识**:注入大量上下文成本高,权衡值不值
