# 不可信 AEP 输入安全 checklist

SUMMARY: 对外处理任何不可信 AEP 前，始终启用 parser 资源预算、fuzz gate、HTTP timeout、panic 隔离和受限 path roots。
READ WHEN: before exposing any AEP upload/parser service to untrusted input, or changing `internal/rifx` / `internal/server` safety controls
RECHECK WHEN: parser limits、server isolation 或 path-input 模型发生变化时

---

`internal/rifx` 已通过统一 parse seam 提供默认预算和可配置的 `Limits`：

- 默认输入 1 GiB、单 chunk 512 MiB、累计 payload 分配 1 GiB、100 万节点、256 层深度。
- `ParseWithLimits` / `ReadChunkWithLimits` 允许服务层进一步收紧。
- 容器 `size >= 4`、child 不越过 parent、声明 payload 不越过输入都在分配前验证。
- `Chunk.U8/U16/U32` 对负 offset 和上界统一返回错误。
- `FuzzParseNeverPanics` 是畸形输入的持续 gate。

公开服务还必须完成：

1. 为公网场景传入比 parser 默认值更小的 Limits，而不是只依赖 HTTP body 大小。
2. 设置请求 timeout、并发/内存限制和 panic recovery；更高风险部署使用独立 worker 进程。
3. path input 若启用，必须限制在配置的 storage roots 内，不能接受任意宿主机路径。

在服务层条件满足前，`aepserver` 仍只定义为受信环境/本地实验入口。
