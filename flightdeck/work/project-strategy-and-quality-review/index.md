# Index — 项目方向与代码质量评估

## 状态

初评已完成并获用户授权执行，控制计划见 `plan.md`。Slice 0–1 已完成；当前进入 Slice 2：服务安全与运维基线。

## 下一步

1. 为 HTTP server 增加 timeout、panic recovery 和收紧的 parser limits。
2. 为 path input 增加 allowed roots，并覆盖目录逃逸测试。
3. 完成 Slice 2 全量门禁后设计公开 Go interface。

## 立即读取

- `../../knowledge/architecture/internal-codebase-map.md`
- `../../knowledge/workflow/project-operating-rules.md`
- `assessment.md`
- `plan.md`

## 按需读取

- `../../knowledge/workflow/verify.md` — 运行或评价完整验证面时
- `../../knowledge/workflow/delivery-contract.md` — 判断哪些能力可以公开宣称可用时
- `../../knowledge/security/untrusted-aep-input.md` — 审查 parser/server 的不可信输入边界时
- `../../knowledge/shape/shape-layer-transform-runtime-sync.md` — 修改 recipe/migration 的 ShapeLayer transform 时
- `../fx-technique-internalization/index.md` — 评价技法学习线是否值得继续时
- `../arbitrary-aep-version-migration/index.md` — 评价版本迁移方向时

## 进展

已完成：

- 读取项目入口 README、Flightdeck cockpit 和知识路由索引。
- 确认现有活跃面偏向复刻、技法内化与未来版本迁移，尚缺统一的产品/开源战略主线。
- 盘点约 700 个 Go 文件和主要包规模，抽样检查 RIFX、serializer、scene、recipe、server 与 CLI。
- `go vet ./...`、capindex check、三平台构建和依赖 DAG 通过。
- `go test ./...` 发现当前未提交 shape transform 改动导致 `aepmigrate` Position 回归。
- 确认外部 Go 项目无法导入 `internal/aep`，当前没有真正公开的 SDK package。
- 确认仓库尚无 LICENSE、CI、release/tag、remote 和开源贡献/安全文档。
- 完成外部生态对比与初步产品/变现路线，见 `assessment.md`。
- Slice 0：修复 shape migration 漏写 runtime transform；原有 recipe transform 在途改动由回归变为完整闭环。
- Slice 0 验证：`go test ./...`、`go vet ./...`、`go run ./cmd/capindex -check`、`go run ./cmd/aepverify cross-platform` 全部通过。
- Slice 1：RIFX parse seam 增加输入、chunk、累计分配、节点和深度预算，增加 typed `LimitError`/`FormatError`、父子边界校验、负 offset 防御和 bounded file read。
- Slice 1 验证：全仓测试、vet、三平台构建通过；5 秒 fuzz 执行约 103 万次输入无 panic。

当前：

- Slice 2：服务安全与运维基线。

## 未决问题

- 最合适的核心用户是谁：开发者、AE 自动化团队、模板/素材平台，还是数字资产分析与迁移工具作者？
- 是否接受“Headless AEP Toolkit”作为核心承诺，而不是继续以技法学习/视觉复刻为主线？
- 哪些实验性研究应继续，哪些应冻结为案例或证据？
- 开源前采用哪种许可证，以及哪些企业能力保留为商业层？
