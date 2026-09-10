# Go SDK examples / Go SDK 示例

五个独立程序，只使用标准库和公开的 `github.com/yueli-fx/aep-parser` package，不导入项目内部实现。从仓库根目录运行，无需 AE。也可以复制任一 `main.go` 到自己的 Go module，添加 SDK 依赖后运行。

Five standalone programs using only the standard library and public SDK. Run from the repository root without AE, or copy a `main.go` into your own module and add the SDK dependency.

| 示例 / Example | 学到什么 / What it demonstrates |
| --- | --- |
| [inspect](inspect/main.go) | 使用资源预算读取工程，输出合成和图层清单 / Bounded parsing and inventory |
| [snapshot](snapshot/main.go) | 导出 ProjectJSON v2，查找属性记录与写回目标 / Property records and write targets |
| [roundtrip](roundtrip/main.go) | Parse → Write → Parse，比对 profile 与字节并输出 SHA-256 / Profile and byte comparison |
| [compile](compile/main.go) | 编译 Recipe，检查 error、状态和降级报告后写文件 / Recipe compilation and status handling |
| [edit-static](edit-static/main.go) | 从快照选择目标、Export 修改、核对 preservation 报告 / Verified static-property export |

```sh
go run ./examples/sdk/inspect
go run ./examples/sdk/snapshot
go run ./examples/sdk/roundtrip -out tmp/sdk-roundtrip.aep
go run ./examples/sdk/compile -out tmp/sdk-compiled.aep
go run ./examples/sdk/edit-static -value '[2.5]' -out tmp/sdk-edited.aep
```

默认输入都是仓库里的 fixture 或 Recipe。写文件的示例拒绝覆盖已有输出；重复运行时换一个 `-out` 路径。这让示例不会意外覆盖输入工程。I/O 失败仍可能留下不完整输出，示例并未实现生产级文件事务。

Defaults use included fixtures and recipes. Output files must not already exist; choose a new `-out` path when rerunning. I/O failures can still leave partial output; these examples do not implement a production file transaction.

`edit-static` 默认修改测试工程中的 `ADBE Time Remapping` 静态数值，以演示写回协议。它不是「启用时间重映射动画」操作。同名目标不唯一时会拒绝操作；在自己的工程中应通过 composition/layer/path 明确选择目标。

`edit-static` changes the fixture's static `ADBE Time Remapping` value to demonstrate export. It does not enable time-remapping animation. Ambiguous matches are rejected; real applications should select by composition, layer, and canonical path.

`roundtrip` 检查重新解析后的 profile 一致性，并独立报告整文件字节是否相同。容器序列化可能规范化字节；它不会把 profile 相同说成所有原字节不变。

`roundtrip` verifies profile equality after reparsing and separately reports whole-file byte equality. Serialization can normalize container bytes.

试着换用组合工程配方 / Compile a composed project:

```sh
go run ./examples/sdk/compile -recipe examples/projects/animated-title.json -out tmp/my-title.aep
go run ./examples/sdk/inspect -in tmp/my-title.aep
```

[完整中文指南](../../docs/usage.md) · [English guide](../../docs/usage.en.md) · [更多示例](../README.md)
