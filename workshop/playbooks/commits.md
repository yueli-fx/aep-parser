---
when_to_read: preparing a commit message; deciding type/scope/subject format
applies_to: [commit-format, conventional-commits, workshop, type-scope, commit-style]
last_updated: 2026-05-22
---

# Commit 规范

## 格式

```
<type>(<scope>): <subject>
```

## Type

| type | 场景 |
|------|------|
| `feat` | 新字段 R/W、新 setter、新 fixture |
| `fix` | bug 修复（错读字段 / setter 写错字节） |
| `refactor` | 重构（拆文件 / 移函数），不改行为 |
| `chore` | 依赖、配置、构建 |
| `docs` | workshop / docs/ 文档同步 |
| `test` | 新测试 / 测试重构 |
| `re` | 纯 RE 探索（新 fixture + 找新字段位置但 setter 未写） |

## Scope（常用）

`layer` / `comp` / `text` / `mask` / `keyframe` / `shape` / `property` / `marker` / `footage` / `project` / `workshop`

## 示例

```
feat(text): add SetManualKerning + FontAxes reader
feat(comp): add Composition.SetSize (cdta @0x8C/0x8E)
fix(mask): persist Locked/MotionBlur on setter call
refactor(test): extract testutil helpers into 5 files
re(text): variable-font axes located at btdk /0/1/0[i]/0/0/4
docs(workshop): reorganize CLAUDE/board/specs structure
```

## 原则

- subject 用动词原形开头，英文
- 不超过 72 字符
- 不写 "fix bug" / "update code" 这类无意义描述
- 命中 negative finding 时用 `re` type 并在 body 写明结论（runtime-only / structural / API limitation）

## 命令一致性

新加 / 修改 / 删除任何 public API 时，必须同步：

1. `workshop/plans/coverage.md` —— 对应行从 🟢 R / 🗑️ 暂搁 升 ✅ R/W（或反向降级）
2. `workshop/plans/coverage-detail.md` —— AE attr 详细交叉表对应行
3. `docs/{layer,property,text,…}.md` —— public API 文档（API 文档优先级最高）
4. `workshop/board.md` —— "最近归档" 一条 + PASS count

只改行为不改 API 表面：仅更 board.md。
