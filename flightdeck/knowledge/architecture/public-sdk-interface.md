# Public Go SDK interface

SUMMARY: 根 `aep` package 通过 `Document` 提供小而深的 inspect/profile/write interface；不要把 `internal/aep` 的大量 facade 符号机械 alias 到外部。
READ WHEN: adding any exported root-package SDK capability, changing `Document`, or deciding whether an internal scene/serializer type should become public
RECHECK WHEN: 首个外部 mutation workflow 要求扩大公开 interface，或 profile schema 进入版本升级时

---

外部消费 seam 位于模块根 package `github.com/yueli-fx/aep-parser`（package name `aep`）。第一版 interface：

- `Open(path)` / `Parse(io.ReadSeeker)` 使用本地文件默认预算并返回 `*Document`。
- `DefaultLimits` + `OpenWithLimits` / `ParseWithLimits` 允许不可信输入调用方收紧输入、chunk、累计分配、节点和深度预算；root `Limits` 是稳定值对象，不暴露 internal RIFX 类型。
- `Document.Inspect()` 返回 detached、轻量、稳定的 `Inspection`。
- `Document.ProfileJSON()` 返回内部 diff/migration 使用的规范化 profile JSON，不泄露内部 profile/scene 类型。
- `Document.Write(io.Writer)` 做 opaque-preserving round-trip；底层拒绝 short write，不能在输出截断时返回成功。

`Document` 隐藏 `internal/aep`、scene、serializer、RIFX chunk 和 back-reference。外部调用者不需要理解内部 497 项 capability surface，也不会因内部 package 重构被迫修改。

扩展规则：先有一个真实外部 workflow，再向 `Document` 增加能完整隐藏该 workflow 复杂度的操作。不要导出内部对象图或批量 type alias；那会把内部实现变成公开 interface，并使现有大 facade 的维护成本永久外溢。

测试以根目录外部测试 package `aep_test` 通过公开 import path 完成 inspect、profile 和 round-trip。interface 命名遵循 Go 标准约定；例如不使用签名不兼容 `io.WriterTo` 的 `WriteTo`。
