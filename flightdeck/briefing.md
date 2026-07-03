# flightdeck briefing — aep-parser

## Conventions

### Project conventions

- **完成即 commit**：每做完一个可交付单元（feature slice / bugfix / 一批文档同步）就 commit，不要攒在工作树里。**本项目的任务进度靠 commit 历史追溯**——未提交的工作不算 landed，会话结束即丢失上下文锚点。
- **直推 main**：本项目无 PR 流，所有工作直接 commit 到 main（见 git log 既有惯例）。不要为常规改动单开分支。
- commit 拆分按逻辑单元：代码 + 其配套测试/fixture/文档同步算一个 commit；无关的 meta/文档改动单独 commit。commit body 可英文按既有惯例。
- **`data/` 是合法测试数据,勿提议清理**：`data/项目/`（含中文名目录 `001_项目A/` `#ep资产/` 等）已由用户确认非敏感（2026-05-17 closed decision）；勿在代码审计中再提 git-purge / gitignore / 换合成 fixture。新出现的路径（如有人丢 `Downloads/xxx.aep`）是另案。
- **导出 API doc comment = 文档源（docgen）**：`docs/*.md` 由 `cmd/docgen` 从 `internal/aep` 导出符号的 doc comment 自动生成（Swagger 式，唯一源=注释）。**doc comment 用英文为源**（中文可后续一键翻译）；prose 写描述、`Example*` 测试函数写示例、概念表走 `docs/_includes/*.head.md`/`*.tail.md`。改导出符号的 doc comment 后须重生成（`go generate ./cmd/docgen` 或 `go run ./cmd/docgen -manifest docs/docgen.json`）。`docs/*.gen.md` 头有 `DO NOT EDIT` —— 不手改生成物。区分点：**内部实现行内注释仍禁**（详 `flightdeck/knowledge/workflow/project-operating-rules.md`），导出符号上方的 doc comment 是文档载体不算违反。
- **知识库必须自包含**：`flightdeck/knowledge/**` 是未来行动规则/陷阱/流程的稳定层，不把 `spec` / `plan` / 临时 work 文件当必读依赖。历史来源可一句话说明，但关键结论、触发条件、修法要写在知识正文里；若需要补充引用，优先指向同属知识库的条目，并保证不读链接也能执行。
- **知识库按需读取**：不要在 preflight 或普通任务开始时批量读取 `flightdeck/knowledge/**`。先读 `flightdeck/knowledge/INDEX.md`，再按 `SUMMARY` / `READ WHEN` 只载入当前任务需要的 1-2 个具体知识文件。
- **CLAUDE.md 已退役**：仓库根部不再保留 CLAUDE.md。原有长期规则已迁入 briefing、`flightdeck/knowledge/workflow/project-operating-rules.md`、`internal/README.md` 和相关 workflow/showcase 知识。

### Commit conventions（本仓库专属 — aep-parser；通用规范见订阅的 `knowledge/commits.md`）

通用 commit 规范走全局订阅（见 § Subscriptions 的 `knowledge/commits.md`）。本仓库在其上叠加：

- **type 扩展**：通用 type 表之外，本仓库增加 `re`（纯 RE 探索：新 fixture + 找到新字段位置，但 setter 尚未写）。
- **type 细化**：`feat` = 新字段 R/W、新 setter、新结构性写 API、新 fixture；`fix` = 错读字段 / setter 写错字节 / round-trip 不一致；`docs` = `docs/` public API 文档 / flightdeck 文档同步。
- **scope（常用）**：`layer` / `comp` / `text` / `mask` / `keyframe` / `shape` / `property` / `marker` / `footage` / `project` / `aep` / `flightdeck`。
- 命中 negative finding 时用 `re` type，并在 body 写明结论（runtime-only / structural / API limitation）。
- **示例**：

  ```
  feat(text): add SetManualKerning + FontAxes reader
  feat(comp): add Composition.SetSize (cdta @0x8C/0x8E)
  feat(aep): DuplicateComposition Alpha→Stable — 2/2 ship-gate PASS
  fix(mask): persist Locked/MotionBlur on setter call
  refactor(test): extract testutil helpers into 5 files
  re(text): variable-font axes located at btdk /0/1/0[i]/0/0/4
  ```

- **命令一致性（改 public API 必读）**：新加 / 修改 / 删除任何 public API 时必须同步——
  1. `docs/{layer,property,text,…}.md` —— public API 文档（API 文档优先级最高）。**新增 facade free function（New*/Add*/Remove* 等自由函数）必须同步登记 `docs/docgen.json` 对应文件的 `funcs` 数组，否则 docgen 跑绿但函数被静默漏文档化**（AddMask 自由函数因 mask.md 无 funcs 数组漏文档整整一天，2026-06-12 补；roots 里的类型方法不受此限，funcs 仅管自由函数）。改 doc comment 后重生成（`go run ./cmd/docgen -manifest docs/docgen.json`）。
  2. `flightdeck/cockpit.md` —— ship-gate PASS count + In flight / Next 状态。
  3. 能力矩阵真相源 = `go run ./cmd/capindex`（capindex tag），不再维护单独的 coverage 表。
- 只改行为不改 API 表面：仅更 `cockpit.md`。

### Showcase routing

Showcase 的完整规则不再放在 briefing。做视觉能力审核、showcase 生成器、review-gate、原版/clone 渲染对比时，按需读取 `flightdeck/knowledge/showcase/showcase.md`。

### Rules

- commit 直推本地 main，无需逐次点名（既定项目惯例：完成一个可交付单元即 commit；无 PR 流）。
- 永不 push 远端：所有工作 commit 到本地 main 即可，任何流程都不提议或执行 `git push`，本地 commit 历史就是交付记录（you, 2026-06-16）。
- 文档可自由更新（含破坏性重写）：flightdeck deck / docs / doc comment 无需逐次确认即可更新、重组、删改；分级（Alpha↔Stable）等契约表述变更照常在 commit body 说明；代码 API 契约本身按 `flightdeck/knowledge/workflow/project-operating-rules.md` 执行（you, 2026-06-11；2026-07-03 从 CLAUDE.md 迁入 Flightdeck）。
- **大计划先总分总**：执行范围较大的计划前，先判断是否需要建立「计划总表 / 覆盖总表 / 汇总结果」作为主控件；需要时按 **总 → 分 → 总** 执行：先定义总目标、维度、域划分、验收口径和结果登记位置；再逐域 / 逐 slice 执行并把证据写回总表；最后回到总表汇总状态、缺口、下一步。不得只靠散落的临时结果或零散 history 来证明大计划进度（you, 2026-07-01）。
- **`tmp/` 不作真相源**：`tmp/` 只放可删除的临时报表、scratch output、重跑缓存；任何 registry/current/coverage/evidence 会长期引用的生成证据必须落到 `registry/evidence/<topic>/` 并登记在 `registry/evidence.json` 或对应 coverage/current JSON。需要清理、迁移、批量验证前先按 `flightdeck/knowledge/workflow/folder-usage-policy.md` 判定目录归属。
- 大计划完成即归档：代表已 ship 功能 arc 的 work effort 完成时，把它的 design.md/plan.md 从 `flightdeck/work/` 移到冷存 `~/.flightdeck/projects/<slug>/archive/`（plain mv，非 git），并同步 cockpit 的 In flight / Next。判定「大计划」= 有独立 work effort、对应一个可交付 feature/结构性 arc；滚动 reference 矩阵不适用，保持在 work/。

## Subscriptions

<!-- 本项目订阅的全局知识，一行一个 ~/.flightdeck-相对路径 -->
knowledge/coding/comments.md
knowledge/git/commits.md
knowledge/agents/subagent-guide.md
