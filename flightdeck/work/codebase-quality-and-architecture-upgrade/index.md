# Index — 代码质量与架构综合升级

## 状态

综合升级已完成：CQ-001–009 verified，CQ-010 accepted P2；无开放 P0/P1。最终发布级门禁全部通过，准备归档。

## 下一步

1. 归档本 work，并从 cockpit 的进行中列表移除。
2. 商业化与 Toolkit 产品化 work 保持暂停，等待后续单独决策。

## 立即读取

- `design.md`
- `plan.md`
- `review-ledger.md`
- `quality-baseline.md`
- `../../knowledge/architecture/internal-codebase-map.md`
- `../../knowledge/workflow/project-operating-rules.md`
- `../../knowledge/workflow/verify.md`
- `../../knowledge/security/untrusted-aep-input.md`

## 按需读取

- `../../knowledge/workflow/delivery-contract.md` — 评价测试证据是否足以支撑公开承诺
- `../../knowledge/workflow/gofmt-baseline-scope.md` — 修改格式门禁或全仓格式化策略
- `../../knowledge/architecture/public-sdk-interface.md` — 审查根 SDK interface
- `../../knowledge/architecture/unified-cli-interface.md` — 审查 CLI contract
- `../../knowledge/shape/shape-layer-transform-runtime-sync.md` — 遇到 chunk/runtime 双重状态问题

## 进展

已完成：

- 定义综合升级总目标和七个分支目标。
- 定义 P0–P3 严重度、review 维度、证据要求和分阶段顺序。
- 决定优先 correctness、安全和状态不变量，不以全仓格式化制造虚假进展。
- 记录核心规模基线：RIFX 3 个 Go 文件 / 638 行，serializer 96 / 22,868，scene 53 / 15,167，internal facade 4 / 3,277。
- CQ-001：公开 `FromReader` 可被 `bpk=1` 的结构合法 keyframe table 触发 slice-bounds panic；`count*bpk` 还可整数溢出并绕过长度检查。
- 新增 `checkedTableLayout`，以除法形式统一验证 count、record size、payload 和最小记录尺寸；接入 keyframe、marker、mask path、shape path 与 mutation 后重解析。
- 新增 serializer 结构化 fuzz surface，种子覆盖最小工程和带关键帧工程，而不只 fuzz 外层 RIFX。
- CQ-002：RenderQueue AddItem 可被短 lhd3 触发 panic；Add/Remove 未验证 header count、stride 与 payload 的一致性。现在两者在 commit 前共用完整 table 验证，失败无 mutation。
- CQ-003：RenderQueue AddItem 曾接受另一 Project 的 Composition；backref 现在携带 owner Project 并按对象身份拒绝跨项目引用。
- CQ-004：RIFX Write 曾把 short write 当成功；exact writer 现在返回 `io.ErrShortWrite`，保护根 `Document.Write` 输出完整性。
- CQ-005：scene raw-byte debug accessor 曾暴露 live chunk slice；现在返回 detached snapshot，字节 mutation 留在 serializer seam。
- 评估并撤回 PropertyStream keyframe 浅拷贝方案：它无法深拷贝 BezierPath 内部 slice，却会给 lowering 热路径增加分配；留待完整 immutable value/interface 设计，不作为当前修复。
- CQ-006：migration 默认 capability ledger 的嵌套 slice/map 曾共享全局状态；默认值和 RuleByID 现在返回深拷贝。
- CQ-007：根 SDK 新增聚合 `Limits` 与受限 Open/Parse，外部服务无需绕过公开 interface 即可收紧不可信输入预算。
- 建立 coverage、race、DAG 和 parse/profile benchmark 基线，见 `quality-baseline.md`。
- CQ-008：Go 1.25.1 命中 18 个可达标准库漏洞；升级 1.25.12 后为 0，CI 固定 govulncheck 并扩展 serializer fuzz/race。
- CQ-009：迁移 gate 的旧 recipe 数量基线造成假红；脚本与知识真相源已更新为 906/900/0/0/6，完整矩阵与 standalone/explicit-matte gate 通过。
- CQ-010：PropertyStream mutable keyframe view 作为 P2 接受；root SDK 不暴露，未来公开 authoring 前必须先设计完整 immutable/mutation interface。

最终结论：

- 2 个 P0、4 个 P1、3 个 verified P2 已关闭；1 个 P2 明确接受并记录公开 authoring 前的重审条件。
- 综合升级覆盖 parser/write correctness、所有权与原子 mutation、公开输入预算、安全工具链、迁移真相源、性能基线、CI race/fuzz/vulnerability gates 和知识整合。
- 本 work 不包含商业化执行，也不恢复已冻结的复刻、technique research 或 ontology 扩张。

## 验证

- 每个修复至少运行目标 package 测试。
- 每个检查点运行 `go test ./...`、`go vet ./...` 和受影响的专用 gate。
- parser 安全问题增加合成畸形 fixture；高风险结构写回按证据等级决定是否需要 AE gate。
- `TestParseEmitsWarningOnInconsistentKeyframeStream` 覆盖短记录和整数溢出 header。
- `FuzzFromReaderNeverPanics` 10 秒约 42.5 万次执行，无 panic。
- `FuzzFromReaderNeverPanics` 延长 30 秒约 217 万次执行，无 panic。
- RenderQueue Add/Remove 正常 round-trip、短 header、count mismatch 和跨 Project 拒绝测试通过。
- recipe profile 151/151、迁移矩阵 906/900/0/0/6、standalone verify、explicit matte 1/1 通过。
- Go 1.25.12 `govulncheck`：No vulnerabilities found。
- 最终 `go test ./...`、`go vet ./...`、`go run ./cmd/capindex -check` 与 `git diff --check` 全部通过。
- Windows amd64、Darwin arm64、Linux amd64 交叉构建通过；core race gate 与 RIFX/serializer 双 fuzz smoke 通过。

## 未决问题

- 大文件与批量峰值内存 benchmark 需要未来取得代表性生产语料后补充；当前不阻塞核心升级关闭。
