# Index — 代码质量与架构综合升级

## 状态

当前优先 work。已完成总目标、分支目标和 review 台账设计，即将从 RIFX → serializer → scene → facade 的核心数据链开始第一轮 P0/P1 排查。

## 下一步

1. 建立包级规模、依赖、复杂度、错误与测试基线。
2. 审查核心 parse/write seam 的边界、资源所有权、未知 chunk 保留和双重状态风险。
3. 把首批发现写入 `review-ledger.md`，为可复现问题补失败测试并修复。
4. 完成首个修复检查点后，再扩展到 profile、recipe、migration 和公开 facade。

## 立即读取

- `design.md`
- `plan.md`
- `review-ledger.md`
- `../../knowledge/architecture/internal-codebase-map.md`
- `../../knowledge/workflow/project-operating-rules.md`
- `../../knowledge/workflow/verify.md`

## 按需读取

- `../../knowledge/security/untrusted-aep-input.md` — parser/server 不可信输入或资源限制问题
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

当前：

- 核心数据链第一轮 review。

## 验证

- 每个修复至少运行目标 package 测试。
- 每个检查点运行 `go test ./...`、`go vet ./...` 和受影响的专用 gate。
- parser 安全问题增加合成畸形 fixture；高风险结构写回按证据等级决定是否需要 AE gate。

## 未决问题

- `scene` 与 `serializer` 的双重状态能否通过更深的 module interface 收敛，而不破坏现有稳定能力？
- inspection 与 authoring 是否应形成不同稳定性等级或不同公开 interface？
- 当前测试数量中有多少只验证 reparse，而没有验证 AE 接受或视觉语义？
