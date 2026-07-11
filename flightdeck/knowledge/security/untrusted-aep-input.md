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

`aepserver` 已提供服务层防护：

1. 上传 body 上限与 parser 默认预算共同生效；畸形 chunk 不能借声明尺寸越过实际输入。
2. CLI 配置 read-header/read/write/idle timeout，handler 有并发上限和 panic recovery。
3. path input 默认关闭；启用时 CLI 强制 `-allowed-path-roots`，解析 real path 后拒绝目录和 symlink 逃逸。

更高风险或多租户部署仍应使用独立 worker 进程、操作系统级资源限制，并按业务文件规模收紧 HTTP body 和 parser Limits；进程内 recovery 不是内存耗尽的隔离替代品。
