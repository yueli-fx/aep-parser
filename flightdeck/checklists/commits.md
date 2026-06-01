---
status: active
last_updated: 2026-05-29
when_to_read: before writing a commit message / staging files / preparing a PR
applies_to: [commit, git, staging, message, push, pr, conventional-commits, type-scope]
portable: partial   # §通用 项目无关, 可整段拷到别的仓库; §项目覆盖 是本仓库 (aep-parser) 专属, 移植时整段替换
---
# Commits Playbook

写 commit / 整理提交时**前置**读这份.

依据:

- Conventional Commits 1.0.0 (<https://www.conventionalcommits.org/>)
- Chris Beams《How to Write a Git Commit Message》(<https://cbea.ms/git-commit/>)

---

## 通用 (项目无关)

### 1. 格式: `type(scope): subject`

```
feat(auth): add refresh-token rotation

老 access token 过期后客户端被强登出. 引入 refresh token
轮换, server 端单次使用 + 失效旧 token.

BREAKING CHANGE: /login 响应去掉 token 字段, 改 accessToken + refreshToken
```

**type**(必填):

| type         | 用于                  |
| ------------ | --------------------- |
| `feat`     | 新功能                |
| `fix`      | 修 bug                |
| `refactor` | 不改行为的重构        |
| `perf`     | 性能优化              |
| `docs`     | 仅文档                |
| `test`     | 仅测试                |
| `build`    | 构建系统 / 依赖       |
| `ci`       | CI 配置               |
| `chore`    | 杂项 (不进上述任何类) |
| `revert`   | 回滚某 commit         |

- **scope**(可选): 受影响的模块/包, e.g. `fix(parser):`. 没有明确单一模块就省略.
- **BREAKING CHANGE**: 破坏性变更在 body 起一段 `BREAKING CHANGE: ...`, 或 type 后加 `!` (`feat!:`).

### 2. Subject 行

- **祈使句现在时**: "add X" / "fix Y", 不是 "added" / "adds" / "fixing". (判据: 补全成 "If applied, this commit will ___".)
- **`type:` 后小写开头**, 句尾**不加句号**.
- **≤50 字符** 为佳 (硬上限 72). 一句话说不完 → 改动可能不原子, 见 §4.

### 3. Body (需要时才写)

- 跟 subject **空一行**隔开.
- **72 字符折行**.
- 讲 **what & why**, 不讲 **how** —— how 看 diff 就知道, 但"为什么这么改 / 解决什么问题 / 取舍了什么"是 diff 表达不出的.
- subject 已经说清的小改动 (typo / 显然的 fix) 不必硬写 body.

### 4. 原子提交

- **一个 commit = 一个逻辑改动**. 能用 "and" 描述 → 多半该拆.
- 重构和功能改动分开提交 (review 时一眼看清哪些是行为变更).
- 不把"顺手"的无关改动 (格式化 / 重命名 / 清理) 卷进功能 commit.

### 5. 不带 AI 署名

不写 `Co-Authored-By: <AI>`, 不写 `🤖 Generated with ...`. 提交前 `git log` 扫一眼历史风格对齐.

---

## 项目覆盖 (本仓库专属 — aep-parser; 移植到别的仓库时整段替换)

### type 扩展 + 细化

通用 type 表之外, 本仓库增加一个:

| type | 用于 |
| ---- | ---- |
| `re` | 纯 RE 探索 (新 fixture + 找到新字段位置, 但 setter 尚未写) |

并细化几个通用 type 在本仓库的判定:

| type | 本仓库语义 |
| ---- | ---------- |
| `feat`  | 新字段 R/W、新 setter、新结构性写 API、新 fixture |
| `fix`   | 错读字段 / setter 写错字节 / round-trip 不一致 |
| `docs`  | `docs/` public API 文档 / flightdeck 文档同步 |

### scope (常用)

`layer` / `comp` / `text` / `mask` / `keyframe` / `shape` / `property` / `marker` / `footage` / `project` / `aep` / `flightdeck`

### 示例

```
feat(text): add SetManualKerning + FontAxes reader
feat(comp): add Composition.SetSize (cdta @0x8C/0x8E)
feat(aep): DuplicateComposition Alpha→Stable — 2/2 ship-gate PASS
fix(mask): persist Locked/MotionBlur on setter call
refactor(test): extract testutil helpers into 5 files
re(text): variable-font axes located at btdk /0/1/0[i]/0/0/4
chore(flightdeck): archive shipped designs/plans to landed/
```

- 命中 negative finding 时用 `re` type, 并在 body 写明结论 (runtime-only / structural / API limitation).

### 命令一致性 (改 public API 必读)

新加 / 修改 / 删除任何 public API 时, 必须同步:

1. `flightdeck/plans/coverage.md` —— 对应行从 🟢 R / 🗑️ 暂搁 升 ✅ R/W (或反向降级)
2. `flightdeck/plans/coverage-detail.md` —— AE attr 详细交叉表对应行
3. `docs/{layer,property,text,…}.md` —— public API 文档 (API 文档优先级最高)
4. `flightdeck/cockpit.md` —— "最近归档" 一条 + ship-gate PASS count

只改行为不改 API 表面: 仅更 `cockpit.md`.
</content>
