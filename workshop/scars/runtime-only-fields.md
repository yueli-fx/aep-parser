---
when_to_read: tempted to add a setter for a "weird" field; ScriptingAPI doc says read-only or "from system registry"; diff baseline shows zero byte change
applies_to: [runtime-only, dropframe, fontlocation, variable-fonts, negative-finding, coverage, scripting-api-gap]
last_updated: 2026-05-22
---

# Runtime-only fields（AE 不持久化到 .aep）

## 现状名单

| 字段 | 证据 |
| --- | --- |
| `CompItem.dropFrame` | AE 25 实测：脚本 `comp.dropFrame=true` 不报错，但 .aep 字节零变化（cdta/idta/iide/idpc 全部 baseline === DROP 测试）。`dropFrame` 行为通过 FrameRate (NTSC 29.97) 自动推断。 |
| `TextDocument.fontLocation` | AE 13.1+ read-only，文档说 "Path of font file ... on disk"。AE 把 PostScript 名拿去查系统字体注册表得 disk path，**不持久化到 .aep**。btdk `/0/1/0[i]/0/0` 只有 `/0`/`/1`/`/2` 三个 key。 |
| Variable fonts axes 的**写**入口 | `TextDocument.fontVariation` 实测不存在，dict-form 赋值 no-op（详 `variable-fonts-write-noop.md`）。**读** 是可达的（`/0/1/0[i]/0/0/4`）。 |

## 识别 runtime-only 的征兆三连

1. AE ScriptingAPI 文档说 read-only，**且**
2. "only reflects first character" 或 "from system registry" 类描述，**且**
3. 跨多个 AE 版本（如 13.1+ 一直存在）

命中三个 → 大概率 runtime-only。优先做"证据三连"判断，省去 RE 工夫。

## 教训

- 怀疑某字段 runtime-only 时，先 probe：用 JSX 显式 set 一个值 → save → dump btdk → diff baseline。零字节变化 = 确认 runtime-only。
- Coverage 表里这种字段标 `❌ runtime-only`，不混进 ❌ structural / 🗑️ 暂搁。
- 这是 "AE 限制" 类 negative finding，写代码这边没办法绕，文档/coverage 表注明就行。
