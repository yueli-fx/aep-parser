# SDK 与集成指南

[中文首页](../README.md) · [English README](../README.en.md) · [中文指南](usage.md) · [English guide](usage.en.md)

## Go SDK

仓库公开后，可在外部 Go 项目中添加依赖：

```sh
go get github.com/yueli-fx/aep-parser
```

外部导入路径是 `github.com/yueli-fx/aep-parser`，package 名称是 `aep`。

下面是完整的工程清单示例，将它放进自己的 Go 项目并传入工程路径：

```go
package main

import (
    "encoding/json"
    "log"
    "os"

    aep "github.com/yueli-fx/aep-parser"
)

func main() {
    if len(os.Args) != 2 {
        log.Fatal("usage: inspect project.aep")
    }
    doc, err := aep.Open(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }
    if err := json.NewEncoder(os.Stdout).Encode(doc.Inspect()); err != nil {
        log.Fatal(err)
    }
}
```

| API | 用途 |
| --- | --- |
| `Open(path)` / `Parse(io.ReadSeeker)` | 使用默认资源预算读取工程 |
| `OpenWithLimits` / `ParseWithLimits` / `DefaultLimits` | 控制输入大小、chunk 大小、累计分配、节点数和递归深度；零字段使用默认值 |
| `Document.Inspect()` | 返回可独立修改的轻量清单 |
| `Document.ProfileJSON()` | 返回规范化 profile JSON |
| `Document.ProjectJSON()` | 返回版本化属性快照，详见 [ProjectJSON v2](project-json-v2.md) |
| `Document.Write(io.Writer)` | 对已解析工程做保留未知 chunk 的 round-trip |
| `Document.Export(ctx, request, writer)` | 在私有副本上执行有边界的属性修改，重新解析并验证未修改字节 |
| `Compile(ctx, recipeJSON, writer)` | 从 Recipe 生成新工程，返回能力、降级、拒绝和重新解析结果 |

`Document` **不支持并发使用**，并发任务应使用独立文档实例。

`Export` 的目标必须来自当前文档快照的 `property_records[*].write_target`；`property_ref` 只用于快照关联，不能用作写回目标。Export v1 仅接受有真实 writer backing、无关键帧的数值静态属性，批次采用全有或全无语义。验证成功时 preservation mode 为 `byte-exact-outside-claimed-ranges`，表示声明修改范围外的 RIFX 数据及未知 chunk 字节保持一致。

`Compile` 和 `Export` 都应同时检查 Go `error` **及报告的 `Status`**：`rejected` 可以伴随空 error 返回。使用 `CompileStatusVerified` / `ExportStatusVerified` 判断成功，并检查降级信息。写入器发生 I/O 错误时仍可能留下部分输出；文件工作流应先写临时文件，成功关闭后再替换目标。

从 Recipe 生成不等于从任意 ProjectJSON 重建完整工程。未知数据可以随原工程保留，并不意味着其语义已经被解析或能被重新生成。

## HTTP 服务

```sh
go run ./cmd/aepserver -addr 127.0.0.1:8080
```

| 方法与路径 | 用途 |
| --- | --- |
| `GET /health` | 健康状态 |
| `GET /capabilities` | 服务能力，包含不可用的宿主能力 |
| `POST /parse` | 上传 `.aep` 并返回解析结果 |
| `POST /profile` | 上传 `.aep` 并返回规范化 profile |

使用 curl 上传（Windows PowerShell 中使用 `curl.exe`）：

```sh
curl -H "Content-Type: application/octet-stream" --data-binary @project.aep http://127.0.0.1:8080/profile
```

可通过 `X-AEP-Path` 提供来源路径标记。默认上传上限为 256 MiB，并发上限为 32，服务配置了 HTTP 超时。直接读取服务端路径默认关闭；本地批处理需显式启用 `-allow-path-input -allowed-path-roots <root1,root2>`，再提交 JSON `{"path":"..."}`。路径模式会检查目录和 symlink 逃逸。

HTTP 服务使用自己的响应结构，`/parse` 不是根 SDK 的 ProjectJSON v2 接口。服务不提供 AE 渲染或宿主读回；不可信上传和部署边界见 [SECURITY.md](../SECURITY.md)。

## 兼容范围与限制

- 格式为大端 RIFF 容器 **RIFX / Egg!**，通过已验证的 chunk 布局解析和写回。
- 兼容基线为 **After Effects 2020（17.0）**；版本化 writer/迁移工作覆盖 AE2020–AE2025。较新工程是否可用取决于具体字段和布局，不承诺所有新版本完全兼容。
- 字段的“可读取”“可保留”“可写回”“可从零创建”和“已通过宿主验证”是不同能力，详见[能力矩阵](capabilities.md)。
- 核心不执行表达式、不渲染画面，也不替代第三方插件、字体或外部素材。视觉结果须在目标 AE 环境确认。
- 迁移采用已支持语义的重建路径；保留未知 chunk 的承诺不能直接套用到迁移或 Recipe 生成。
- 部分测试依赖本地 fixture、AE 安装或私有语料，可能跳过。纯 Go 测试通过不等于完整宿主验收通过。
