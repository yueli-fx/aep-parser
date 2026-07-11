# Index — 项目方向与代码质量评估

## 状态

初评已完成并获用户授权执行，控制计划见 `plan.md`。Slice 0–4 已完成；当前进入 Slice 5：开源发布面（许可证选择除外）。

## 下一步

1. 补齐 SECURITY、CONTRIBUTING、CHANGELOG、CI 和 release workflow。
2. 整理 README 的安装、五分钟示例、支持范围和 WIP 表述。
3. LICENSE 必须由用户确认许可证后单独落地，不替用户作授权决定。

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
- Slice 2：HTTP server 增加默认 32 并发、读写/idle timeout、panic recovery；path input 强制 allowed roots，并拒绝目录与 symlink 逃逸。
- Slice 2 验证：`go test ./...`、`go vet ./...` 和三平台构建全部通过。
- Slice 3：根 package 新增可被模块外导入的 `Document` interface，提供 `Open/Parse`、`Inspect`、`ProfileJSON` 和 `Write`；内部 scene/serializer/backref 不外泄。
- Slice 3 验证：外部测试 package 通过公开 import path 完成 inspect/profile/roundtrip；全仓测试、vet、三平台构建通过。
- Slice 4：新增 `cmd/aep` 与 `internal/toolkitcli`，统一 inspect/profile/diff/migrate/capabilities、JSON envelope 和退出码；旧命令保留兼容。
- Slice 4 验证：真实 fixture CLI smoke、统一 runner tests、全仓测试、vet、capindex 和三平台构建全部通过。

当前：

- Slice 5：开源发布面（LICENSE 等待用户确认）。

## 未决问题

- 最合适的核心用户是谁：开发者、AE 自动化团队、模板/素材平台，还是数字资产分析与迁移工具作者？
- 是否接受“Headless AEP Toolkit”作为核心承诺，而不是继续以技法学习/视觉复刻为主线？
- 哪些实验性研究应继续，哪些应冻结为案例或证据？
- 开源前采用哪种许可证，以及哪些企业能力保留为商业层？
