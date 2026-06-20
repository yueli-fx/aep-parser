---
status: active
last_updated: 2026-06-20
when_to_read: before deciding to use subagent
applies_to: [subagent, api, claude-in-claude, artifact, autonomous, context, skill, system-prompt]
synced: true
---
# Subagent 使用指南

## 核心原则

Subagent 是**无状态的**。每次 API 调用都是全新会话，不携带任何来自主对话的信息。

**你显式传入的 = subagent 知道的。其他一概不知。**

---

## 什么时候用 Subagent ✅

适合**自包含**任务，即任务本身提供了全部所需上下文：

| 场景                 | 原因                       |
| -------------------- | -------------------------- |
| 总结 / 翻译一段文本  | 文本本身就是全部上下文     |
| 数据提取 / 格式转换  | 输入输出明确，无需额外规范 |
| 评分 / 分类          | 给数据 + 标准，一次完成    |
| 生成结构化 JSON      | prompt 里写清格式即可      |
| 单次问答（无需约束） | 不需要记忆，不需要规范     |

---

## 什么时候不用 Subagent ❌

| 场景                     | 原因                             |
| ------------------------ | -------------------------------- |
| 写代码                   | 需要代码风格、架构约束、项目背景 |
| 需要遵守复杂规则 / skill | 规则不会自动传入                 |
| 多轮迭代任务             | 每次失忆，无法累积               |
| 需要了解 codebase 结构   | 没有文件系统访问                 |

---

## Subagent 不会自动携带的内容

| 内容                     | 是否自动带入  |
| ------------------------ | ------------- |
| 当前对话历史             | ❌            |
| 系统提示 / skill 文件    | ❌            |
| claude.md                | ❌            |
| Memory / 记忆            | ❌            |
| 你显式写进代码的任何东西 | ✅            |
| MCP 连接                 | ✅ 需显式声明 |

---

## 必须显式注入的内容

### 注入约束 / skill（system 参数）

```javascript
const response = await fetch("https://api.anthropic.com/v1/messages", {
  method: "POST",
  body: JSON.stringify({
    model: "claude-sonnet-4-6",
    max_tokens: 1000,
    system: `你的代码规范、约束、或 skill 内容写在这里`,
    messages: [{ role: "user", content: userInput }]
  })
});
```

### 注入对话历史（多轮场景）

```javascript
messages: [
  { role: "user", content: "第一轮" },
  { role: "assistant", content: "回复" },
  { role: "user", content: currentInput }
]
```

### 注入 MCP 连接

```javascript
body: JSON.stringify({
  model: "claude-sonnet-4-6",
  max_tokens: 1000,
  messages: [...],
  mcp_servers: [
    { type: "url", url: "https://mcp.example.com/sse", name: "my-mcp" }
  ]
})
```

---

## 最佳实践

- **写代码、遵守规范 → 直接在主对话做**，不要用 subagent
- **非用不可时**，把所有约束塞进 `system` 参数
- **结构化输出**，让 subagent 只返回 JSON，避免解析歧义
- **token 意识**：每次注入大量 system prompt 成本较高，权衡是否值得
