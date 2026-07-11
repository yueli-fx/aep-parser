# Index — 代码质量与架构综合升级

## 状态

当前优先 work。总目标、分支目标和 review 台账已建立；核心数据链第一轮排查已修复 serializer 不可信固定记录 table 的整数溢出与短记录 panic（CQ-001 / P0）。

## 下一步

1. 继续扫描 serializer 中其余外部 count/offset/size → 分配或切片路径，确认是否仍有可触发 panic 或过量分配。
2. 审查 serializer mutation 的失败回滚、unknown chunk 保留、parent size reflow 和 ID/reference 原子性。
3. 再进入 scene/runtime 双重状态与 facade interface 深度审查。

## 立即读取

- `design.md`
- `plan.md`
- `review-ledger.md`
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

当前：

- 继续 serializer 安全与 mutation 原子性 review。

## 验证

- 每个修复至少运行目标 package 测试。
- 每个检查点运行 `go test ./...`、`go vet ./...` 和受影响的专用 gate。
- parser 安全问题增加合成畸形 fixture；高风险结构写回按证据等级决定是否需要 AE gate。
- `TestParseEmitsWarningOnInconsistentKeyframeStream` 覆盖短记录和整数溢出 header。
- `FuzzFromReaderNeverPanics` 10 秒约 42.5 万次执行，无 panic。

## 未决问题

- `scene` 与 `serializer` 的双重状态能否通过更深的 module interface 收敛，而不破坏现有稳定能力？
- inspection 与 authoring 是否应形成不同稳定性等级或不同公开 interface？
- 当前测试数量中有多少只验证 reparse，而没有验证 AE 接受或视觉语义？
