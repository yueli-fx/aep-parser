# 首次发布与设计伙伴验证

## Goal

完成首次真实发布验证，并用六周外部设计伙伴证据判断 aep-parser 应继续商业化、保持开源或私有 Toolkit，还是停止产品化扩张。

## Status

Open

## Current

技术预发布基线已经完成，根 Go SDK、统一 CLI、CI/release workflow、安全与贡献文档、CHANGELOG 和 Apache-2.0 均已落地。当前分支为 `mainline/aep-understanding`，仓库尚未配置 remote；真实 CI、版本/tag 与外部团队验证尚未开始。

## Next

与用户确认首发采用公开开源、私有 Toolkit 或并行路径，并取得要绑定的 remote 信息；随后从 [阶段 A](plan.md) 开始，按 [验证流程](../../knowledge/workflow/verify.md) 核对本地基线，再依 [交付契约](../../knowledge/workflow/delivery-contract.md) 运行首次真实 CI。除非用户明确授权，不要 push、创建 tag 或发布。

## Execution pointer

`plan.md` → 阶段 A「首次发布准备」→ 确认交付路径与 remote。

## Progress

- 技术预发布基线及 P0/P1 质量升级已经完成。
- OSS core 与 managed/on-prem 商业层已有初步边界。
- 原 `commercialization-and-toolkit-productization` 的产品承诺、交付、采用、商业验证与生态决策已合并到本 Work。

## References

- [稳定背景与决策边界](context.md)
- [完整执行计划](plan.md)
- [开源许可边界](../../knowledge/workflow/open-source-license.md)
- [不受信任 AEP 输入](../../knowledge/security/untrusted-aep-input.md)
- [公共 SDK 接口](../../knowledge/architecture/public-sdk-interface.md)
- [统一 CLI 接口](../../knowledge/architecture/unified-cli-interface.md)

## Open questions

- 首发是公开开源发布、私有 Toolkit 交付，还是两者并行？
- 首个版本采用 `v0.x` 预发布还是更保守的内部版本标识？
- 哪类设计伙伴最容易先提供合法 corpus 与真实失败标签？
