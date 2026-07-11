# Index — 项目方向与代码质量评估

## 状态

Slice 0–6 与 Apache-2.0 许可证均已完成，Headless AEP Toolkit 的技术基线、公开 interface、统一 CLI、发布基础设施和采用/商业验证方案均已落地。仓库已达到技术预发布状态；首次公开发布仍需建立远端、运行真实 CI 并确定版本号，外部设计伙伴验证需要真实团队参与。

## 下一步

1. 建立远端仓库并触发首次 CI，不自动 push。
2. 确定首个版本号和发布日期，再创建 tag。
3. 按 `adoption-and-commercial-validation.md` 招募设计伙伴并运行六周验证。

## 立即读取

- `../../knowledge/architecture/internal-codebase-map.md`
- `../../knowledge/workflow/project-operating-rules.md`
- `assessment.md`
- `plan.md`
- `adoption-and-commercial-validation.md`

## 按需读取

- `../../knowledge/workflow/verify.md` — 运行或评价完整验证面时
- `../../knowledge/workflow/delivery-contract.md` — 判断哪些能力可以公开宣称可用时
- `../../knowledge/security/untrusted-aep-input.md` — 审查 parser/server 的不可信输入边界时
- `../../knowledge/shape/shape-layer-transform-runtime-sync.md` — 修改 recipe/migration 的 ShapeLayer transform 时
- `../../knowledge/workflow/gofmt-baseline-scope.md` — 修改 CI 格式门禁或计划全仓 gofmt 时
- `../../knowledge/workflow/open-source-license.md` — 准备发布、调整贡献条款或开源/商业边界时
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
- Slice 5：新增 SECURITY、CONTRIBUTING、CHANGELOG、CI 和 tag release workflow；README 补充 pre-release、安装和许可证状态；跨平台 gate 覆盖根 SDK/统一 CLI。
- Slice 5 验证：CI 等价命令全部通过，包括测试、vet、capindex、三平台构建、10 秒 fuzz（约 184 万输入）和 race tests。
- Slice 6：定义模板市场、批量渲染、工作室/DAM 三类设计伙伴，六周验证节奏、OSS/商业边界、定价实验和继续/停止条件。
- Slice 6：新增 Nexrender/aerender preflight 示例，明确本工具位于 render worker 之前而不替代渲染编排。
- 许可证：采用 Apache License 2.0，README、CHANGELOG 与长期贡献/商业边界规则已同步。

当前：

- 技术预发布基线已完成；等待远端首次 CI、版本/tag 决策与真实外部设计伙伴验证。

## 未决问题

- 最合适的核心用户是谁：开发者、AE 自动化团队、模板/素材平台，还是数字资产分析与迁移工具作者？
- 是否接受“Headless AEP Toolkit”作为核心承诺，而不是继续以技法学习/视觉复刻为主线？
- 哪些实验性研究应继续，哪些应冻结为案例或证据？
- 哪些企业能力应在设计伙伴验证后保留为商业层？
