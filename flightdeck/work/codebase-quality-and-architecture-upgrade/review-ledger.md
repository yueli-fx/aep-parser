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
