# 首次发布与设计伙伴验证

## Goal

完成首次真实发布验证，并用六周外部设计伙伴证据判断 aep-parser 应继续商业化、保持开源或私有 Toolkit，还是停止产品化扩张。

## Status

Open

## Current

公开资料已迁至 `showcase/` 和 `docs/knowledge/`；4 篇内部协作笔记留在工作站。新增 5 个公开 SDK 示例和 3 个组合工程配方。公开源码不依赖工作站文件，归档排除由 `.gitattributes` 控制。不含 Git/工作站/本地展示产物的干净源码已在 Go 1.25.12 下通过全仓测试；8 个新增示例、capindex、资源目录 gate、vet 和跨平台构建已通过。

技术预发布基线已经完成，根 Go SDK、统一 CLI、CI/release workflow、安全与贡献文档、CHANGELOG 均已落地；2026-09-11 按用户要求将许可证调整为 GPL-3.0-only 并补齐双语 README。当前分支为 `main`，仓库已配置 GitHub origin；本次未执行远端 CI、创建版本/tag 或开展外部团队验证；本地概览审核的发现见下方记录。

## Next

用户已确认准备公开开源；先处理 [开源审核记录](../../../docs/open-source-audit.md) 中实际复现的测试基线问题及具体参考材料说明，再从 [阶段 A](plan.md) 开始，按 [验证流程](../../../docs/knowledge/workflow/verify.md) 核对本地基线，再依 [交付契约](../../../docs/knowledge/workflow/delivery-contract.md) 运行首次真实 CI。除非用户明确授权，不要 push、创建 tag 或发布。

## Execution pointer

`plan.md` → 阶段 A「首次发布准备」→ 核对 Go 补丁版本对文档生成的影响，并运行首次真实 CI。

## Progress

- 技术预发布基线及 P0/P1 质量升级已经完成。
- OSS core 与 managed/on-prem 商业层已有初步边界。
- 原 `commercialization-and-toolkit-productization` 的产品承诺、交付、采用、商业验证与生态决策已合并到本 Work。

## References

- [稳定背景与决策边界](context.md)
- [完整执行计划](plan.md)
- [开源许可边界](../../../docs/knowledge/workflow/open-source-license.md)
- [不受信任 AEP 输入](../../../docs/knowledge/security/untrusted-aep-input.md)
- [公共 SDK 接口](../../../docs/knowledge/architecture/public-sdk-interface.md)
- [统一 CLI 接口](../../../docs/knowledge/architecture/unified-cli-interface.md)

## Open questions

- 首个版本采用 `v0.x` 预发布还是更保守的内部版本标识？
- 哪类设计伙伴最容易先提供合法 corpus 与真实失败标签？
