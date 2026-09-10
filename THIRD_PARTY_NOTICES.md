# Third-party notices / 第三方声明

The project's PolyForm Noncommercial 1.0.0 license does not replace the licenses or copyrights of third-party components. Preserve their notices when redistributing them.

项目的 PolyForm Noncommercial 1.0.0 不替代第三方组件原有的许可与版权。再分发时应保留相关声明。

## Go runtime and standard library

CLI binaries include components of the Go runtime and standard library, copyright The Go Authors, under the BSD 3-Clause license. The license is reproduced in [licenses/Go-LICENSE](licenses/Go-LICENSE).

Source and toolchain releases: <https://go.dev/dl/>. The required toolchain is recorded in `go.mod`; release build provenance can also be inspected with `go version -m <binary>`.

## golang.org/x/text

The direct Go module dependency is `golang.org/x/text v0.38.0`, copyright The Go Authors, under the BSD 3-Clause license. Preserve [licenses/x-text-LICENSE](licenses/x-text-LICENSE) and the additional patent grant in [licenses/x-text-PATENTS](licenses/x-text-PATENTS).

Source: <https://go.googlesource.com/text/>. Exact module versions and checksums are recorded in `go.mod` and `go.sum`; obtain build dependencies with `go mod download`.

These notices describe the inspected direct dependency and Go runtime. They are not a provenance certification for every fixture, embedded binary template, dictionary, or reference artifact. See [the publication review](docs/open-source-audit.md) for outstanding asset checks. When dependencies or distributed materials change, review and update the notices before release.

这些声明覆盖本次检查的直接依赖与 Go runtime，不代表所有 fixture、嵌入二进制模板、字典与参考资料的来源已获得确认。待确认项见[发布审核记录](docs/open-source-audit.md)；依赖或分发材料变更后应更新声明。
