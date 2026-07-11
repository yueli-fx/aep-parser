# 执行计划 — 从研究型仓库收束为 Headless AEP Toolkit

## 目标

在不继续扩大功能面的前提下，把项目推进到三个可验证状态：

1. 当前代码和迁移基线全绿。
2. 底层 parser 对畸形/不可信输入有明确预算且不 panic。
3. 外部用户拥有一个小而深的 Go interface 和单一 CLI 入口。

## 原则

- total -> slices -> total：本文件是控制台账；每个切片独立验证、独立提交，最后重新评估产品化缺口。
- 新 interface 必须隐藏内部复杂度，不把 `internal/aep` 的大表面机械搬到公开包。
- 先保证 preserve/edit/migrate 正确，再优化传播和商业包装。
- 不把复刻、technique ontology 或新字段扩张混入本计划。

## Slice 0：恢复可信基线

交付：

- 保留当前 shape layer transform 的直接 runtime 写法。
- 修复 `aepmigrate` 重建 shape layer 时漏写 layer transform 的历史假绿。
- 增加 recipe -> profile -> migrate 的回归覆盖。
- `go test ./...`、`go vet ./...`、capindex check、跨平台构建全绿。

## Slice 1：RIFX 安全解析 module

交付：

- 在统一 parse seam 引入默认安全预算：输入/单 chunk、累计 payload、节点数、递归深度。
- 校验容器 `size >= 4`、child 不越过 parent、声明 payload 不超过可读范围。
- 负 offset accessor 返回错误而不是 panic。
- malformed-input 单元测试和 Go fuzz target。
- 现有受信 AEP roundtrip 行为不变。

## Slice 2：服务安全与运维基线

交付：

- HTTP server 配置 read-header/read/write/idle timeout。
- handler panic recovery，parser 异常返回稳定错误而不是终止进程。
- path input 增加显式 allowed roots；默认继续关闭。
- 并发和输入限制形成可配置 interface 与测试。

## Slice 3：公开 Go interface

交付：

- 在模块外可导入的位置建立最小 `aep` package。
- 第一版只公开 `Open/Parse/Profile/Write` 及少量稳定模型/错误，不导出内部 backref、serializer 和全部 setter。
- 用外部测试 package 验证真实下游可以导入和完成 inspect/roundtrip。
- 明确版本兼容和稳定性策略。

## Slice 4：单一 CLI 产品面

交付：

- 新增统一 `cmd/aep`：`inspect`、`lint`、`diff`、`migrate`、`capabilities`。
- 复用现有深层 module，不复制各 CLI 实现。
- 原有研究/治理命令标为 advanced/internal，不立即删除。
- 统一 JSON envelope、退出码和错误格式。

## Slice 5：开源发布面

交付：

- README 改成用户价值、安装、五分钟示例和支持矩阵。
- LICENSE 选择等待用户确认；在确认前不擅自授权发布。
- 补齐 SECURITY、CONTRIBUTING、CHANGELOG、CI 和 release workflow。
- 提供 benchmark、真实案例和 Docker/local service 示例。

## Slice 6：采用与商业验证准备

交付：

- 定义三类设计伙伴的访谈/试用包。
- 输出 OSS core 与 managed/on-prem 商业层边界。
- 准备与 aerender/nexrender pipeline 的 preflight 集成示例。
- 只有真实需求支持时才恢复 recipe compiler 或 technique learning 的扩张。

## 完成标准

- 每个已执行切片拥有代码、测试、必要文档和原子提交。
- 任一时刻 `index.md` 能说明当前切片、验证结果和下一步。
- 最终回到本计划汇总已完成、剩余风险、公开发布阻塞和下一季度路线。
