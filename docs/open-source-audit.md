# 公开发布前概览审核

日期：2026-09-11。范围：当前 Git 工作树的公开发布准备；在修改 README 前盘点代码、文档、项目知识、依赖、素材目录与发布配置，再以实际命令核对入门流程。

这是一轮概览审核，不是逐行代码审计、完整 Git 历史密钥扫描、所有素材权属鉴定或全版本 AE 宿主验收。本次没有发现已证实的侵权素材；下文区分已经找到生成来源的测试资源、未逐项核实的参考材料，以及实际复现的测试问题。

## 项目实际内容

以下为修改前 `git ls-files` 的盘点快照，不作为长期自动更新的数量承诺：

| 内容 | 数量 | 位置 |
| --- | ---: | --- |
| 跟踪文件 | 2,176 | 整个仓库 |
| Go 文件（含测试） | 883 | 根 SDK、`internal/`、`cmd/`、工具与 showcase |
| Markdown | 301 | 文档、知识、证据与工作记录 |
| 知识笔记 | 155 | `docs/knowledge/` |
| Recipe JSON 示例 | 151 | `examples/recipes/` |
| Showcase 子目录索引 | 25 | `showcase/`，不含总索引 |
| Fixture 目录文件 | 151 | `test_data/fixtures/`，含说明与其他测试材料 |
| 嵌入模板目录文件 | 306 | `internal/serializer/templates/` |

根 package 提供 `Open`、`Parse`、资源预算、`Document` 检查/profile/属性快照/写回/Export，以及 `Compile`。内部 facade 和模型比根 SDK 广，生成文档中的许多符号不能被模块外调用方直接导入。CLI、Recipe、HTTP 与宿主 worker 也是不同入口，不能合并承诺为同一公开协议。

知识并非零散的临时笔记：已经按解析、图层、形状、文字、效果、关键帧、合成、安全、验证等主题组织。`registry/evidence/` 保存机器证据，showcase 保留生成器及验收说明。README 应引导用户访问这些资料，而不是把所有研究日志放进首页。

## 本次已完成

- 最终采用 PolyForm Noncommercial 1.0.0，根 `LICENSE` 为官方标准全文；初期 GPL 方案已在公开发布前被替换。
- 同步双语 README、贡献许可、CHANGELOG、许可证决策笔记与相关发布计划，统一当前 PolyForm Noncommercial 许可与商业授权说明。
- 增加第三方声明，保留 Go runtime/标准库、直接依赖 `golang.org/x/text v0.38.0` 的 BSD 许可及 x/text 专利授权文本。实际构建依赖仍由 `go.mod`/`go.sum` 固定。
- 重写中英文 README，提供可运行 SDK 示例、CLI/Recipe/HTTP 入门、状态报告处理、兼容边界、架构及知识导航。
- 将原 README 的详细自托管说明移至 [self-hosted-reports.md](self-hosted-reports.md)，保留已有工作流信息。
- 修改 release workflow：三个平台的二进制压缩包携带 PolyForm 许可、使用与商业授权说明及第三方声明；Release 同时上传对应 tag 的项目源码归档并说明构建方式。这里是本地配置修改，没有触发发布。

## 发布前待处理的发现

### 1. 测试资源与参考材料：来源核对范围

这里的「素材」原先用词过泛。仓库主要包含逆向测试工程、构建工程所用的二进制结构模板和从 AE 导出的效果元数据；本次没有发现已证实侵权的商业素材包。

进一步查阅项目知识和脚本后，已经找到这些来源链：

| 材料 | 已找到的生成/提取依据 |
| --- | --- |
| 效果参数字典 `data/effects-dict/` | `scripts/effects-dict/dump_effects_dict.jsx` 和 `.ps1` 在 AE 中添加效果并导出名称、类型、默认值；[字典说明](knowledge/workflow/effects-dict.md)记录了过程 |
| 效果/形状二进制模板 | [嵌入模板架构说明](knowledge/effects/embed-template-architecture.md)记录「JSX 在 AE 中创建最小工程 → 提取结构 → Go 克隆并覆写值」；仓库保留 `test_data/generators/` 与 `tools/debug/extract_shape_bodies/` 等工具 |
| 测试工程 fixture | `scripts/fixtures/fixtures_manifest.json` 包含 62 个生成配置条目，关联 JSX、AE 版本与输出工程；例如 `re_effect_lib12.jsx` 专门创建一组效果用于参数结构研究 |

这说明它们有大量自建测试工程和自动提取数据的来源证据，不能仅因为后缀是 `.aep` 或 `.bin` 就列为版权问题。该结论不等于已逐文件核实全部来源。

`data/reference/` 还保留参考项目的分析报告，例如 `booyah-glitch/booyah_dissect.txt`。本次只确认它是提取的工程结构报告，没有核实原始工程取得方式与报告公开范围；应针对这种具体参考材料确认，而不是把整个测试目录视为待授权的商业素材。

`.gitignore` 已排除 `data/samples/`、`learning/`、`tmp/`，当前 Git 索引未跟踪这些本地语料目录。Git 作者标签也不足以单独证明版权归属；本次没有确认外部贡献授权缺失。当前许可变更不追溯撤销过去已授予的许可，也不替换第三方组件原许可。当前条款见 [PolyForm 标准许可](https://polyformproject.org/licenses/noncommercial/1.0.0)。

### 2. 文档漂移检查依赖换行与 Go 补丁版本

原始工作树在 Windows / Go 1.25.13 下运行 `go test ./...` 失败于 `cmd/capindex` 和 `cmd/docgen`，其他报告的 package 未失败。

- `cmd/capindex/main.go` 的 `fileEqual` 与 `generated_test.go` 的 `mustMatch` 直接比较字节。当前 checkout 为 CRLF，而生成器输出 LF。临时归一化两个能力索引文件到 LF 后，`capindex -check` 通过；验证后已恢复原始工作树字节。
- `cmd/docgen/main.go` 和 `index.go` 使用 `runtime.Version()` 写入生成文档/索引，提交产物及 golden 使用 Go 1.25.12，本机为 1.25.13。golden 测试还直接比较换行。临时将 golden 转为 LF，并用 `GOTOOLCHAIN=go1.25.12 go test ./cmd/docgen` 验证通过，随后恢复原文件。

建议后续对文档生成与比较统一 LF，并消除 Go 补丁版本对内容漂移检查的影响或明确固定生成环境。本次不以批量重写生成文档来掩盖这个跨环境问题，也没有修改生成器实现。README 的安全补丁升级说明与现有生成器约束之间仍有这个已知落差。

### 3. 宿主验收与发布通路仍需实际验证

本次未启动 AE、未跑全版本宿主/渲染验收、未触发 GitHub Actions 或 Release。配置中已经有 CI、漏洞扫描、fuzz、race 和多平台发布步骤，但配置存在不代表它们在目标远端运行通过。

正式发布前应处理已复现的基线问题，并按需要核对具体参考材料的公开范围，确认目标 AE 版本的交付证据、真实 CI、安全私报入口、首发 tag 和附件。当前 `SECURITY.md` 依赖 GitHub Private Vulnerability Reporting，公开前须确认仓库设置已启用该入口。

## 本次验证记录

| 检查 | 结果 |
| --- | --- |
| 初始工作树 | 干净；本次未提交、push、创建 tag 或修改仓库可见性 |
| 常见凭证文件名与高置信密钥特征扫描 | 当前跟踪文件未命中；只检查若干常见私钥、GitHub/AWS/API token 模式，不等于无秘密证明 |
| 私有语料/临时目录是否被跟踪 | `data/samples/`、`learning/`、`tmp/` 未跟踪 |
| `go test ./...`，原始 Windows checkout，Go 1.25.13 | 失败：上述 capindex/docgen 环境敏感检查；未发现其他失败 package |
| `go vet ./...` | 通过 |
| `go run ./cmd/capindex -check` | 原始 CRLF checkout 失败；仅临时转 LF 后通过 |
| `go test ./cmd/docgen`，Go 1.25.12 + LF golden | 通过；临时调整已恢复 |
| `go run ./cmd/aepverify cross-platform` | Windows amd64、macOS arm64、Linux amd64 构建通过 |
| `go run ./cmd/aepverify recipe-profiles` | 151 / 151 通过；validate、compile、profile checks 均无失败 |
| README 完整 Go 示例 | 从 README 提取并运行，成功检查 `v2_smoke.aep` |
| README Recipe 示例 | validate、compile、再次 inspect 成功；产物 1 合成、2 图层 |
| 发布配置 | YAML 可解析；本地 Windows 二进制与许可附件的 tar.gz 打包检查通过；远端三个 runner 未执行 |
| 新增/修改阅读文档的相对链接 | 检查文件/目录目标存在；不验证深层原有文档的全部链接 |
| 当前许可全文与第三方许可 | 使用原始许可全文，未改写条款 |

临时日志和 smoke 产物放在忽略的 `tmp/oss-*`，本页保存可长期引用的结论。

## 许可边界

最终方案为 **PolyForm Noncommercial 1.0.0 + 单独书面商业授权**。这是 source-available 许可，限制免费许可范围以外的商业使用；标准许可中的机构用途例外仍然有效。详见 [使用与商业授权](../LICENSING.md)。本次修改不设定价格，不授予通用商业许可，也不追溯撤销旧版本已授予的权利。

## 公开目录与示例迁移更新

按用户要求将 25 个展示方向迁至 `showcase/`，公开的 151 篇专题笔记迁至 `docs/knowledge/`；4 篇协作/工作规划笔记留在内部工作站。新增 5 个公开 SDK 程序和 3 个组合工程 Recipe。

- 公开 README、代码、脚本、能力验证与资料链接同步使用新目录。
- `.gitattributes` 将工作站排除于源码归档，同时规定生成文档使用 LF；它不删除 Git 历史，也不使已有跟踪记录在 GitHub 仓库中不可见。
- 基于项目指定的 Go 1.25.12，全仓测试通过。原 CRLF 检查问题已由 LF 属性和工作树换行修正；docgen 对运行时补丁版本的依赖仍在。
- 5 个 SDK 示例和 3 个组合配方已实际运行，含错误路径、拒绝覆盖和源文件保持检查。新组合配方仅完成纯 Go 验证，未新增 AE 画面验收。
- 干净源码检查发现 profile、technique、slice 与自托管测试原先依赖被忽略的 showcase 输出。现将 3 个由仓库 Go 生成器自行创建的 AEP 固定到 `test_data/fixtures/showcase/`，附生成命令及 SHA-256；测试与自托管入口直接使用这些 fixture。
- 资源目录 gate（8 步）通过，新增公开文档/示例的归属已登记；同时补齐了此前 4 个 README 的目录归属。

迁移最终验证：不含工作站、Git 元数据和本地展示产物的 2,194 文件源码副本在 Go 1.25.12 下通过 `go test ./...`。`go vet ./...`、能力索引检查、8 步资源目录 gate 以及 Windows amd64 / macOS arm64 / Linux amd64 构建均通过。公开 Markdown 的相对文件链接已检查。没有新增 AE 宿主或画面验收，也没有提交或发布。

## 发布前最终许可更新（2026-09-11）

用户最终选择 PolyForm Noncommercial 1.0.0。已同步 LICENSE、双语 README、LICENSING.md、贡献指南、第三方声明、CHANGELOG、许可决策与 release workflow。根 LICENSE 与官方纯文本逐字节一致；授权说明不修改标准条款。发布文档压缩包已核对包含完整许可和双语使用/商业授权说明，YAML、链接、资源索引 JSON 与 diff 空白检查通过。未设定商业价格、签署商业合同、提交或发布。

同日新增的 AE2022 参数名称测试工程已由用户反馈本机测试通过，详见 [验收说明](../examples/authoring/ae2022-named-effects/README.md)；这是人工反馈，不扩展为全版本宿主验收。
