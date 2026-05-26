# aep-parser — AI 协作规范

Go 实现的 Adobe After Effects `.aep` 二进制解析器，对照 boltframe/aftereffects-aep-parser 重写 + 增量。**读取下限 AE 2020**，写回是 **length-preserving splice**。

## 核心铁律

1. **所有写都是 length-preserving**（V1 阶段；V2 引入结构性写，仍保持 length-preserving 默认路径）。改字段不准动 chunk 大小。少数 length-variable 例外（name / comment / expression / 字体名 / 文本字符串），但 `WriteAEP` 会重算父 LIST size。
2. **public API 不动**。重构 / 拆文件 / 移函数都不能改 exported 类型 / 方法签名 / JSON 输出字段。
3. **`internal/aep` 单 package**。不引子包（会强制 API 重排）。
4. **改任何 public API 必须同步**：
   - `workshop/plans/coverage.md` — 状态行（✅ / 🟢 / 🗑️ / ❌）
   - `workshop/plans/coverage-detail.md` — AE attr 详细交叉表
   - `workshop/board.md` — "最近归档" + PASS count（**唯一权威**，其它文档不写 PASS 数）
   - `docs/{layer,property,text,…}.md` — **大批量 parity / 多 phase 工作期间可滞后**，完工后专项一次"读源码遍历重生 docs/"。中途零散小改照常同步。源码 + coverage 是权威；docs 是给外部用户读，落后于源码也不阻塞。
5. **所有 superpower 产物落 `workshop/`**（不进 `docs/superpowers/` 默认位置）：
   - brainstorming 产出的 design spec → `workshop/specs/YYYY-MM-DD-<topic>-design.md`（workshop v0.8.0 命名约定，时间戳前缀）
   - writing-plans 产出的 implementation plan → `workshop/plans/YYYY-MM-DD-<topic>-plan.md`
   - 跟既有 `workshop/specs/architecture.md` + `workshop/plans/coverage*.md` 同目录，统一文档入口。
   - 完工后 `git mv` 到对应 `finish/` 子目录（specs/finish/ 或 plans/finish/）归档。
6. **嵌入资源（templates / fixtures）的目录命名复数**：`internal/aep/templates/` / 不是 `template/`。Go `//go:embed` 限制资源必须在 package 同目录或子目录，所以不能放项目根的 `templates/`。

## 项目入口命令

- 验证: `go vet ./... && go test ./...`
- 跑单测: `go test ./internal/aep/ -run 'TestX' -v`
- 详细操作 / tmp_debug 工具表 / fixture 验证: `workshop/playbooks/verify.md`

## 场景触发器（**唯一权威文档地图**）

| 场景 | 必读 |
| --- | --- |
| 新会话 | `workshop/board.md`（最先读，看 "Next session 进来先做"） |
| 找下一个能动的字段 | `workshop/plans/coverage.md` |
| 查 "AE 的 X 字段对应到哪个 Go API" | `workshop/plans/coverage-detail.md` |
| 动代码前 / 不确定架构 | `workshop/specs/architecture.md` |
| 跑测试 / 准备 commit | `workshop/playbooks/verify.md` |
| 写新 fixture / 重跑 RE | `workshop/playbooks/re-fixture.md` |
| 写 commit | `workshop/playbooks/commits.md` |
| 遇到奇怪行为 / 怀疑结论 | `workshop/scars/`（高代价 negative findings 错题集） |

`workshop/{sketches,wip}/` 给临时草稿，目前都空。

## 工作风格

- 代码优先，设计讨论精简
- 重构 / 迁移**先读源码再动**，不瞎猜
- 不写注释，除非 WHY 不明显
- 高代价 negative findings → 单独建 `workshop/scars/<topic>.md`

## 维护 board.md 的规则

只记 "用户视角可感知" 的语义推进，**不是 activity log**。纯探索 / grep / 改 typo 都不更新。
- 修改在飞 / 归档段时同步更新 `Last updated`
- `Active focus` 反映**当前**主线，不是历史
- PASS count 只在 board.md `Last updated` 一处记，其它文档不重复（防漂移）

## 进来第一件事

读 `workshop/board.md` 的 "Next session"：
- 空 / 跟 `Last updated` 对不上 / 跟当前 `git status` 矛盾 → **先核对再动**
- 一致 → 按第一句执行
