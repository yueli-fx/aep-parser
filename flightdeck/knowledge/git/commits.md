# commits checklist

Conventions for commit messages, atomic commits, and staging.

写 commit / 整理提交时**前置**读这份.

本仓库直接在当前 branch/mainline 上工作，除非用户要求，不例行创建 feature branch；永远不要 push 到远端。暂存、提交、tag、发布或准备 PR 都需要用户明确授权。

依据:

- Conventional Commits 1.0.0 ([https://www.conventionalcommits.org/](https://www.conventionalcommits.org/))
- Chris Beams《How to Write a Git Commit Message》([https://cbea.ms/git-commit/](https://cbea.ms/git-commit/))

## aep-parser 补充约定

- `re`：纯逆向发现或 fixture 证据，尚未形成已交付 setter/API。
- 常见 scope：`layer`、`comp`、`text`、`mask`、`keyframe`、`shape`、`property`、`marker`、`footage`、`project`、`aep`、`flightdeck`。
- `feat`：新的已交付字段读写、setter、结构 API，或 fixture 支撑的能力。
- `fix`：错误读写字节、setter 行为错误、round-trip 或 AE-acceptance 回归。
- `docs`：公共 API 文档或 Flightdeck 文档。

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

### 4.1 文档 / spec / plan 不要碎提交

- **探索期、讨论期、未定稿的 md / knowledge / spec / plan / Flightdeck Work 文案不要每改一次就提交**. 这些文件经常反复推敲，碎提交会污染 git 历史。
- 先把草案留在工作区。文档-only 提交必须更严格：用户明确说“提交”、明确说这份文档已经定稿/落账，或文档随已完成的代码/测试/fixture 作为同一个 landed unit 一起提交。
- 用户说“记录一下”“先写着”“停车/后期任务”“开始调研/开始执行”不等于允许提交 Work/spec/plan/knowledge 草案；这些仍然留在工作区，直到用户明确要求提交或明确确认定稿。
- 代码、测试、fixture、生成物、迁移结果、可运行能力落地后仍应及时提交；文档随代码作为同一 landed unit 的配套内容一起提交。
- 目标: commit 记录稳定决策和可交付进展，不记录每一次思路摆动。

### 5. 不带 AI 署名

不写 `Co-Authored-By: <AI>`, 不写 `🤖 Generated with ...`. 提交前 `git log` 扫一眼历史风格对齐.

### 6. 多行 message：认清 shell 再传

当环境**同时挂 Bash 和 PowerShell 工具**时, 给原生命令(`git commit` 等)传多行串前先认清当前工具是哪个 shell:

- **PowerShell 工具** → here-string `@'...'@`(结束 `'@` 必须顶列零缩进).
- **Bash 工具** → 真 heredoc(`git commit -F - <<'EOF' … EOF`)或 `-F <file>`; **别用 `@'...'@`** —— bash 没有 here-string, `@` 会当字面量混进 subject(`@ chore: …`).
- 最稳, 跨 shell 通用: 把信息写进文件, `git commit -F <file>`.

### 7. 暂存前扫 `RM`/`MM`（重命名+内容改动）

`git mv` 重命名文件后再编辑内容，`git status --short` 会显示 `RM`（index 已暂存重命名、工作区内容未暂存）；`R100` = 内容改动**未暂存**（只提交了纯重命名）。**提交前扫一遍 `RM`/`MM` 行，对命中的文件再 `git add <file>` 暂存内容**，直到 `git status --short` 干净（或只剩预期的未跟踪文件）再 commit。
