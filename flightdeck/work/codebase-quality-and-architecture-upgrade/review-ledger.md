# Review 台账 — 代码质量与架构综合升级

## 严重度

- **P0**：可造成工程损坏、任意访问、失控资源消耗、稳定 panic，或无安全绕过方式的数据错误。
- **P1**：常见输入产生错误结果、mutation 状态不同步、稳定 interface 无法兑现，或架构使高风险缺陷反复出现。
- **P2**：局部耦合、重复实现、浅 module、错误上下文不足、测试脆弱或明显性能浪费。
- **P3**：命名、格式、注释、局部可读性和低风险维护问题。

## 状态

- `candidate`：静态观察，尚未复现。
- `confirmed`：已有最小复现、失败测试或明确代码路径证据。
- `fixing`：修复进行中。
- `verified`：修复及相关回归通过。
- `accepted`：风险被明确接受，并记录原因和重新检查条件。

## 条目格式

| ID | 严重度 | Module / seam | 状态 | 发现与影响 | 证据 | 关闭条件 |
|---|---|---|---|---|---|---|
| CQ-001 | P0 | serializer fixed-record table seam | verified | keyframe 等 table 使用 `count*bpk` 检查，可整数溢出；`bpk` 小于固定读取尺寸时会通过总长度检查并 slice panic。公开 `FromReader` 可由结构合法的畸形 AEP 触发。 | `TestParseEmitsWarningOnInconsistentKeyframeStream/bytes_per_keyframe_too_small` 修复前稳定 panic；overflow case、目标包测试和 10 秒 serializer fuzz 修复后通过。 | keyframe、marker、mask/shape path 和重解析共用溢出安全、含最小记录尺寸的检查；结构化 parser fuzz 持续运行。 |
| CQ-002 | P0 | RenderQueue mutation seam | verified | `AddItem` 只检查 lhd3 存在，随后直接切 `Data[0x08:]`；结构合法但短 header 可从公开 mutation 触发 panic。header count/stride 与 scene item 数不一致时 Add/Remove 还会写出损坏 table。 | `TestRenderQueueAddItem_RefusesShortSettingsHeader`、`TestRenderQueueMutationsRefuseInconsistentSettingsCounts`。 | Add/Remove 在 commit 前统一验证双 count、固定 stride、payload 长度和 scene item 数，失败无 observable mutation。 |
| CQ-003 | P1 | RenderQueue ownership interface | verified | `AddItem` 接受属于另一 Project 的 Composition，只写入碰巧相同命名空间的数值 ID，可能生成错误或悬空 comp reference。 | `TestRenderQueueAddItem_Refuse` 增加双 Project 用例并确认 queue size 不变。 | RenderQueue backref 记录 owner Project，AddItem 使用对象身份拒绝跨 Project composition。 |
| CQ-004 | P1 | RIFX write interface | verified | `Chunk.Write` 忽略 `io.Writer` 的 short write，`Document.Write` 可能在输出被截断时返回 nil。 | `TestChunkWriteRejectsShortWrites` 使用 `(n < len, nil)` writer，要求 `io.ErrShortWrite`。 | RIFX write seam 统一包装 exact writer，所有 header、payload、padding 和 trailing write 都拒绝静默短写。 |
| CQ-005 | P2 | scene debug byte views | verified | `CdtaRawBytes`、`PrdaRawBytes`、`LdtaRawBytes`、`SspcData` 声称只读却返回 live chunk slice，调用方可绕过 serializer 改字节。 | `TestRawByteAccessorsReturnDetachedSnapshots`。 | 所有 raw accessor 返回 detached copy；内部读取行为和现有长度测试保持不变。 |
| CQ-006 | P2 | migration default configuration | verified | `DefaultCapabilityLedger` 只复制顶层 struct，返回的 Rules slice、SourceVersions 和 TargetSupport map 与全局默认值共享；调用方修改会污染后续迁移判断。 | `TestDefaultCapabilityLedgerReturnsDetachedState`。 | Default ledger 和 RuleByID 都深拷贝可变 slice/map。 |
| CQ-007 | P1 | root SDK parse seam | verified | 外部 SDK 只能使用 1 GiB 本地默认预算，处理不可信上传时无法收紧 chunk、分配、节点和深度限制，只能绕过公开 interface。 | 根外部测试 `TestExternalPackageCanTightenParseLimits`；docgen/capindex 500 项对齐。 | 根 `Limits` 值对象与 `OpenWithLimits` / `ParseWithLimits` 进入同一 serializer/RIFX seam，零字段沿用默认。 |
| CQ-008 | P1 | Go toolchain / CI | verified | Go 1.25.1 的标准库存在 18 个项目可达漏洞，涉及 net、crypto/tls、x509、url、os 等 server/parser 可达路径。 | 升级前 `govulncheck` 报 18；Go 1.25.12 下扫描为 `No vulnerabilities found`。 | `go.mod`、setup-go CI/release 使用 1.25.12；CI 固定 govulncheck v1.6.0，并扩展 serializer fuzz/race。 |
| CQ-009 | P2 | migration verification truth source | verified | 新增 recipe 后矩阵已是 151×6=906 case，脚本仍断言旧 858/852，导致 900 个真实 pass 后 gate 假红。知识文档更旧为 846/840。 | 更新后 `verify_matrix.ps1 -OutRoot tmp/code_upgrade_matrix`：906/900/0/0/6、standalone verify pass、explicit matte 1/1。 | 脚本与 verify knowledge 同步 906/900；新增/删除 recipe 必须审查两处基线。 |
| CQ-010 | P2 | PropertyStream keyframe view | accepted | `Keyframes()` 返回内部 slice，可绕过排序/重复时间约束；浅拷贝又无法隔离 BezierPath 内嵌 slice，并增加 lowering 分配。 | 全仓调用审计显示根 SDK 不暴露 PropertyStream，当前调用集中在 internal lowering/tests。 | 本轮不做半深拷贝；若 authoring 进入根 SDK，先设计 immutable value 或受控 iterator/mutation interface。 |

## Review 维度

每个核心 module 至少检查：

- interface 需要调用方知道多少实现细节；
- 输入、资源、生命周期和并发不变量；
- 错误是否有类型、上下文和稳定分类；
- mutation 是否原子，失败是否回滚；
- 未知或未来版本数据是否保留；
- 测试是否穿过正确 seam，并覆盖错误路径；
- 性能是否存在无界分配、重复全量复制或不必要解析；
- durable knowledge 与代码真相源是否一致。
