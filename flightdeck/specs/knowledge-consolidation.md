---
status: active
summary: 退役 auto-memory 系统(迁入 flightdeck/CLAUDE.md)+ incidents 过时清理与合并减量 + coverage.md/coverage-detail.md 退役(待 capindex 完成)+ CLAUDE.md 瘦身 + 根目录 stray exe 清理;目标=单一知识家 flightdeck + 精简 CLAUDE.md + 干净仓库
graduate: true
last_updated: 2026-06-16
---

# 知识库整合与退役 (memory/coverage/incidents/CLAUDE.md → flightdeck 单一家)

> **进度(2026-06-16,capindex P2 done 后启动)**:**E ✅**(12 个根目录 stray exe 删除)· **D ✅**(coverage 退役→残值,见下)· **C ✅**(8-agent 只读审计 grep 核实:0 过时、4 并 3 锚、46 keep → 51→47,见下)· B 部分(CLAUDE.md 能力源指针已改指 capindex)。**剩 A(记忆退役,repo 外删除需用户确认)· B 全量瘦身。**

## 背景 / 动机

项目知识现散在四处:① harness auto-memory(`~/.claude[-accounts/max.config]/projects/E--projects-tools-aep-parser/memory/`,~20 条 + MEMORY.md);② `flightdeck/`(specs/plans/incidents/checklists/references);③ `CLAUDE.md`(68 行密集长段,硬约束 #1-7 尤其臃肿);④ `coverage.md`/`coverage-detail.md`(863 行,手写易过期)。多处=重复 + 漂移 + 不知道去哪查。

用户决策:**单一知识家 = flightdeck;CLAUDE.md 只留精简精华;退役 auto-memory 系统**(可犯的错→incidents、精华→CLAUDE.md、指针→references)。coverage 在 capindex(`specs/2026-06-16-capability-index.md`)完成后退役。本 spec 排在 capindex 之后(coverage 部分硬依赖;其余可在 capindex P1 后并行)。

## 目标 / 非目标

**目标**
- auto-memory 系统退役:~20 条记忆全部迁入 flightdeck/CLAUDE.md,memory 目录清空,加 CLAUDE.md 规则禁用该系统。
- incidents 全量审核,过时内容归档/删除,存活项与 capindex `incident=` 反向链接对齐。
- `coverage.md` + `coverage-detail.md` 退役(内容已被 capindex tag 吸收后)。
- CLAUDE.md 瘦身:留"精简精华 + 指针",细节下沉 flightdeck。
- 立约:**新知识只进 flightdeck(错误→incident、流程→checklist、指针→reference、设计→spec)或 CLAUDE.md(跨切面精华);不再用 auto-memory。**

**非目标**
- 不在本 spec 重做 capindex(它是前置依赖)。
- 不改运行时代码。
- 不动 showcase 升级(另起 spec)。

## 原则

1. **单一来源**:每条知识只在一处权威;重复即删一份。
2. **就近归类**:错误/陷阱→`incidents/`;可复用流程→`checklists/`;外部指针→`references/`;能力状态→capindex tag;跨切面铁律/工作风格→CLAUDE.md。
3. **CLAUDE.md 是索引不是仓库**:铁律保留为"短规则 + 指针",长 RE 解释移出。
4. **删前先验迁移**:任何删除(memory 文件 / coverage / incident)前,目标位置内容必须已落盘且可读(迁移→核对→才删,三步;memory 在 repo 外,尤其慎)。

## 工作流 A — auto-memory 系统退役

**A1 分类迁移**(~20 条,按内容路由;重复项直接删):

| 桶 | 去向 | 代表记忆 |
|---|---|---|
| 错误/陷阱 | `incidents/`(合并进相关现有 incident 或新建) | ae_scripted_quit_grace · ship_gate_exit2 · subagent_lsp_stale · gate_fixture_id_coincidence · aep_crlf_gofmt |
| 工作风格精华 | CLAUDE.md `## 工作风格` | collaboration_style(terse autonomous)· finish_the_direction · look_at_artifact · bisection_over_stacking |
| 可复用流程 | `checklists/`(合并) | ae_ship_gate_self_serve→re-fixture · self_verify_showcase_render→showcase · procedural_over_vector_blur→(火焰番外,留 checklist 或 reference) |
| 外部指针 | `references/` | docs_index_json · git_commit_multiline_shell · test_data_clearance |
| 已重复 → 直接删 | (已在 flightdeck) | delivery_contract(=checklists/delivery-contract.md)· docs_destructive_updates(=rules.md Autonomy)· ae_version_downgrader_cep(=incidents/2026-06-09-…)· aep_parser_reference(=CLAUDE.md 抬头)· board_status_drift(capindex 落地后基本 moot,留一句于 CLAUDE.md 或删) |

执行时产出**逐条映射表**(20/20 都有去向或删除理由),迁移完核对后再清空目录。

**A2 退役机制**(auto-memory 是 harness 级,不能"关闭"功能,只能让其空转 + 用更高优先级指令覆盖):
- 迁移核对后,删除 memory 目录下所有 `*.md`(含 MEMORY.md),或把 MEMORY.md 置空。
- CLAUDE.md 加规则(最高优先级,覆盖默认 Memory 行为):**"本项目不使用 auto-memory 系统。新知识写入 flightdeck(错误→incident/流程→checklist/指针→reference/设计→spec)或 CLAUDE.md(跨切面精华)。不要写 memory 文件。"**
- 注:memory 目录在 repo 外、且双路径疑似同一库(`.claude` 与 `.claude-accounts/max.config` 指向同一文件)——删除是 repo 外破坏性操作,执行时须显式向用户确认。

## 工作流 B — CLAUDE.md 瘦身

**留**(精华 + 指针):数据流图 · 硬约束 #1-7 的**一句话规则 + 指针**(详情指向 incident/checklist)· 工作风格 · 文档地图 · 入口命令。
**移出**:#2/#3 里的长篇 RE/历程解释 → `references/` 或对应 incident;能力分级细节 → capindex tag(coverage 退役后)。
**criteria**:一条内容若 (a) 是"为什么这么设计"的长解释、或 (b) 只在特定任务才需要 → 移出留指针;若是"每次动代码都要守"的铁律 → 留短规则。目标:CLAUDE.md 读一遍即知边界与去哪查,不在其中堆细节。

## 工作流 C — incidents 审核(51 个):过时清理 + 合并减量 — ✅ DONE(2026-06-16)

**做法**:8-agent 只读 workflow,每篇 grep 代码核实结论是否仍成立(铁律:代码>正文)。**结论:0 篇过时**(incidents 维护良好、多为各自独立 RE,非冗余——"太多"更多是知识量大而非重复)。**4 篇并入 3 现有锚**(复用最全锚以免断引用):cdta-duration→cdta-0xB0(@0xB0=duration 前提已证伪,保留 RQ 断言指导)· keyframe-byte-layout-dispatcher→lhd3-keyframe-capacity-pages · pwsh-7-no-winrt + windows-media-ocr-cjk-glyph-spacing→ae-automation-occlusion-crashstate(脚本已引此 slug)。**51→47**。被并 4 篇 rename 入 `archive/incidents/`(git 保留)。C3 对齐:capindex `-check` OK,incident= 反链全解析无悬链(被并 4 篇均未被任何 tag 引用)。原下文为执行前的候选/判据记录。

---

## 工作流 C(原始)— incidents 审核(51 个):过时清理 + 合并减量

**C1 过时清理**——过时判定(任一即标候选):被后续代码改动推翻(记忆 board-status-drift:正文滞后于代码)· 已 resolved 且不再可达 · 内容已被 capindex tag / 其他 incident 吸收。处置:过时→归档(`archive/` 或删,执行时定)+ 在相关存活 incident 注明。

**C2 合并减量**——51 个偏多,同子系统 / 同根因 / 互相交叉引用的应折叠成一个**多 Case incident**(用 `## Case N` 约定,findings 全保留、不丢信息)。合并判据:同一 chunk/字段族 + 同一根机制 + 高频共读。候选簇(执行时核实):AE 自动化类(`ae-automation-occlusion-crashstate` + `pwsh-7-no-winrt` + `windows-media-ocr-cjk-glyph-spacing`)· cdta 时长/shutter 类(`cdta-0xB0-shutter-ref-not-duration` + `cdta-duration-two-representations` + `shutter-side-effect-divisors`)· ldta 长度类(`ae2020-shape-ldta-164-corrupt` + `ldta-length-third-variant-not-found` + `camera-filmsize-ldta-write-blocked`)· keyframe 容量/布局类(`lhd3-keyframe-capacity-pages` + `keyframe-byte-layout-dispatcher`)· negative-finding 类小条目可归一篇"negative findings 合集"。目标:51 → 显著更少、更好导航。

**C3 对齐**:存活/合并后 incident 确保被 capindex 某 tag 的 `incident=` 引用(反向链接核:capindex 跑出的 incident 引用集 vs incidents 目录;孤儿 incident 人工判去留;合并后旧 slug 在 tag 里的引用需同步改指新 slug)。

**贯穿**:**真相源=代码 > incident 正文**(对齐 board-status-drift),审核/合并须 grep 代码/ship-gate 核实,不照信正文。

## 工作流 D — coverage 退役(硬依赖 capindex)— ✅ DONE(2026-06-16)

capindex P2 完成后核对发现:coverage 含 capindex **不覆盖**的内容——故**退役到残值**而非整删(合原则 #4):
- `coverage.md`(415→75 行):`## ✅ 已 ship` 写能力快照(347 行)退役(被 capindex 写能力 + docgen read/API + incidents 机制吸收且更准)→ 换成 capindex 指针;**保留** capindex 不覆盖的 暂搁/不可达/negative findings/可探方向(残值)。status→`superseded`。
- `coverage-detail.md`:AE-attribute→Go-field 矩阵是**正交另一维**(capindex 是符号级,不含 attribute 级映射)→ **保留为参考**,status→`reference` + 顶部 banner 指向 capindex 为能力状态源。
- CLAUDE.md 文档地图 + 分级条目已改指 capindex(`-q` / `docs/capabilities.{json,md}`)。
- 原文全在 git history,可追。

## 工作流 E — 根目录 stray exe 清理(无依赖)— ✅ DONE(2026-06-16)

实测根目录有 **12 个 stray `.exe`**(spec 当初记 11,实为 12:含 `capindex.exe`),全 untracked + `.gitignore` 已含 `*.exe` → 删除零风险。已 `rm` 全部 12 个,根目录 0 stray exe。防复发约定(`go run ./path` / `go build -o tmp_debug/bin/`)待写入 CLAUDE.md 入口命令(随 B)。

---

原始现状记录:项目根有 **11 个 `.exe`**(aepdemo / animated-path / bisect_v2_2 / docgen / dump_fdta / ge_cross_project_insert / gradient_roundtrip / keyframe-channels / stroke-detail / structural-ops / transform-values),共 ~44MB。**均未被 git 跟踪**(`git ls-files '*.exe'` 空)且 `.gitignore` 已含 `*.exe` 规则 —— 纯构建产物垃圾,删除零风险、不影响 git。

**根因**:`go build`(showcase 生成器 / tmp_debug 工具)不带 `-o` 时把二进制落在 CWD=根目录。
**处置**:① 直接删根目录全部 `.exe`(可 git 回滚?否——本就未 tracked,但可随时 `go build` 重生成,无损失)。② 防复发约定(写入 CLAUDE.md `## 入口命令` 或 checklist):跑生成器/工具优先 `go run ./path`;必须 build 时 `go build -o tmp_debug/bin/`,不在根目录裸 build。`.gitignore` 已覆盖,无需改。
**独立**:不依赖 capindex,随时可做(可作为本 spec 第一个落地的小项)。

## 依赖 / 排序

```
E 根目录 exe 清理 ── 无依赖,随时可做(本 spec 第一个可落地小项)
capindex P1(地基) ──┬─► A 记忆退役(可并行起)
                     ├─► B CLAUDE.md 瘦身(可并行起)
                     └─► C incidents 审核(需 capindex 的 incident= 反链,P1 后)
capindex P2(全量标注) ──► D coverage 退役(硬 gate)
```
E 独立、随时可做;A/B/C 可在 capindex P1 落地后开始;D 必须等 capindex P2。本 spec 现为 `idea`,capindex 推进到位后翻 `active` 并起实现计划。

## 验收标准

- [ ] memory 目录清空 + CLAUDE.md 含"不用 auto-memory"规则;20 条记忆有逐条去向记录。
- [ ] incidents 审核完成,无过时项,存活项与 capindex `incident=` 对齐(无孤儿/无悬链)。
- [ ] coverage.md + coverage-detail.md 删除,无悬空引用。
- [ ] CLAUDE.md 瘦身后仍覆盖全部铁律(以短规则+指针形式),`go vet ./... && go test ./...` 不受影响。
- [ ] 根目录无 stray `.exe`;防复发约定已写入 CLAUDE.md/checklist。
- [ ] 全库搜索:同一知识无两处权威副本。

## 风险

- **memory 删除在 repo 外、双路径同库**:破坏性 + 不可 git 回滚,执行时务必先迁移核对 + 用户确认。
- **CLAUDE.md 瘦身过度**:把铁律当细节删掉会丢约束。criteria=铁律留短规则,只移"解释/历程"。
- **incidents 误删**:正文看似过时但代码仍依赖其结论 → 必须 grep 代码核实(board-status-drift 教训),存疑保留。
