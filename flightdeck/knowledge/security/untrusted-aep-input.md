# ⚠ 不可信 AEP 输入必须受解析预算与进程边界保护

SUMMARY: 当前 RIFX 解析器按文件声明直接分配和递归，适合受信本地文件；公开上传服务必须先增加尺寸、深度、节点数、总分配预算与 panic 隔离。
READ WHEN: 对外开放 AEP 上传、把 parser/server 部署到不可信网络、增加 fuzzing，或修改 `internal/rifx` chunk 读取逻辑时
RECHECK WHEN: `internal/rifx` 引入统一解析预算、fuzz gate 或 server 进程隔离后

---

`internal/rifx.readChunk` 当前直接使用 chunk header 的 `uint32 size`：

- 普通 chunk 按声明长度分配 `[]byte`，未先与剩余输入或全局预算核对。
- LIST/RIFX 默认假设 `size >= 4`；更小的声明值会进入负长度/负结束位置路径。
- 容器递归没有最大深度、节点数或累计 payload 上限。
- `Chunk.U8/U16/U32` 只检查上界，不接受负 offset；如果未来调用方把不可信偏移传入，切片会 panic。

HTTP 层限制上传 body 大小只能限制原始文件字节数，不能阻止伪造的 chunk 声明触发超大内存分配或深递归。公开服务至少需要：

1. 解析前验证每个 size 不超过当前容器剩余字节。
2. 为单 chunk、总 payload、节点数和递归深度设置显式预算。
3. 所有畸形输入返回 typed error，不 panic；增加 fuzz 和最小恶意样本测试。
4. 服务设置请求 timeout、并发/内存限制和 panic recovery；更高风险部署使用独立 worker 进程。
5. path input 若启用，必须限制在配置的 storage roots 内，不能接受任意宿主机路径。

在这些条件满足前，对外文档应把 `aepserver` 定义为受信环境/本地实验入口，不宣称为可直接暴露到公网的服务。
