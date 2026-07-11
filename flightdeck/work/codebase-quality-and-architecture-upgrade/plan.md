# 计划 — 代码质量与架构综合升级

## Total → Slices → Total

先建立全局控制台账，再按风险切片修复，最后回到全局复审残余风险、module interface 和质量门禁。

## Slice 0：基线与审计工具

- 记录包规模、依赖方向、测试面、复杂函数、panic/忽略错误/不受控分配模式。
- 为 review 台账建立稳定 ID、严重度、证据和关闭条件。
- 确认现有测试、vet、capindex 和跨平台 gate 基线。

## Slice 1：RIFX 与输入安全

- 复审 chunk 边界、容器声明、整数运算、递归、分配预算和序列化读取。
- 对每个可利用或可触发的发现增加合成 fixture、fuzz seed 或 property test。
- 确保错误可分类且不会泄漏内部路径或产生模糊成功。

## Slice 2：Serializer 与写回原子性

- 审查 parse/lower/backref/mutation/template seam。
- 检查未知 chunk 保留、parent size reflow、失败回滚、ID/reference 完整性和重复字节副本。
- 把跨调用方重复的不变量收进深 module。

## Slice 3：Scene 状态与 facade

- 盘点 chunk-side 与 runtime-side 双重状态，定义唯一 mutation 路径。
- 评价 `internal/aep` 与根 SDK interface 的深度、稳定性等级和泄漏类型。
- 分离稳定 inspection contract 与实验性 authoring surface，若证据表明确有必要。

## Slice 4：Profile、Recipe 与 Migration

- 追踪 profile 真相、缺失字段、默认值和版本 fallback。
- 审查 recipe/migration 是否绕过统一 mutation seam。
- 建立 rebuild 与 round-trip 的差异契约和证据等级。

## Slice 5：服务、CLI 与批处理

- 审查错误 envelope、退出码、取消、并发、超时、临时文件和大批量隔离。
- 用统一 toolkit module 降低 CLI/server 重复逻辑。
- 加入端到端容量和故障注入测试。

## Slice 6：知识、规范与自动化

- 建立能力 → owner → implementation → test → evidence → docs 的可追踪关系。
- 合并重复知识，保留 routing header 和 freshness 触发器。
- 在基线允许范围内逐步扩大 gofmt/lint/静态分析门禁。

## Final Total：复审与发布判断

- 复跑完整 P0–P3 台账，确认无开放 P0，P1 有明确关闭或接受理由。
- 重新绘制 module/interface/seam 地图和公开稳定性承诺。
- 输出私有使用、开放核心和完全开源都可采用的生产 readiness 结论。
