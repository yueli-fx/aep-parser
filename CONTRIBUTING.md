# Contributing to aep-parser

感谢参与。这个项目处理未公开的二进制格式；高质量贡献必须同时说明行为、证据和兼容边界，而不只是让 Go round-trip 通过。

## 开始之前

1. 先阅读 `README.md`、`internal/README.md` 和最接近修改目录的 README。
2. 新用户能力优先进入根 `aep` package 或统一 `cmd/aep`，不要默认增加新的顶层命令。
3. 新的 AEP 字段或结构 mutation 应先附最小 fixture/生成步骤和 AE 版本信息。
4. 不要提交无权公开的客户工程、商业模板、字体、插件或素材。

## 本地验证

最低要求：

```powershell
gofmt -w <changed-go-files>
go test ./...
go vet ./...
go run ./cmd/capindex -check
go run ./cmd/aepverify cross-platform
```

修改 `internal/rifx` 时另外运行：

```powershell
go test ./internal/rifx -run '^$' -fuzz '^FuzzParseNeverPanics$' -fuzztime=10s
```

AE-sensitive 写路径还需要对应 AE2020/AE2025 acceptance 或 render gate。无法运行 AE 时，应明确说明未执行的 gate，不要把纯 Go 通过描述为可交付证明。

## Pull Request 内容

- 一次 PR 只解决一个逻辑问题。
- 说明用户可观察行为、为什么需要修改，以及不在范围内的内容。
- 列出实际运行的验证命令。
- breaking change 必须显式标记，并同步公开 interface、README、生成文档和 capability index。
- 二进制格式发现如果会改变未来实现方式，应同时补充可复用测试或项目知识；不要只留下临时 dump。

## Fixture 与安全

- 优先使用最小合成 fixture；真实工程必须确认再分发权利并删除敏感路径/素材信息。
- 畸形输入测试应尽量直接构造字节，不提交超大文件。
- 安全问题按 `SECURITY.md` 私下报告。

## Commit 格式

使用 Conventional Commits：`type(scope): subject`，例如：

```text
fix(rifx): reject chunks outside parent payload
feat(cli): add unified inspect command
```

不要添加 AI 署名或无关格式化改动。

## 贡献许可 / Contribution license

新贡献须按 **PolyForm Noncommercial 1.0.0** 提供。贡献者保留版权，确认自己有权提交，保留第三方声明并说明来源。完整条款见 [LICENSE](LICENSE)，授权方式见 [LICENSING.md](LICENSING.md)。

本项目还可能提供单独的商业授权。**仅提交或合并 PR，不视为贡献者已授权维护者将其贡献用于商业再授权。** 拟纳入商业授权版本的外部贡献，维护者须在纳入前另行取得明确的书面许可，覆盖所需的商业使用、修改、分发及再许可权；保留授权记录，必要时使用单独的贡献者协议。没有相应权利的贡献，不得宣称已被项目商业许可证覆盖。

Contributions must be offered under **PolyForm Noncommercial 1.0.0**. Contributors retain copyright, must have the necessary rights, and must preserve third-party notices and disclose provenance. See [LICENSE](LICENSE) and [LICENSING.md](LICENSING.md).

The project may also offer separate commercial licenses. **Submitting or merging a PR does not by itself grant maintainers commercial sublicensing rights to that contribution.** Before incorporating external contributions into commercially licensed versions, maintainers must obtain and retain explicit written permission covering the necessary commercial use, modification, distribution, and sublicensing rights, using a separate contributor agreement where appropriate. Contributions lacking those rights must not be represented as covered by the project's commercial license.
