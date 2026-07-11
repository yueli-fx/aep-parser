# Index — 首次发布与设计伙伴验证

## 状态

已建档，尚未执行。技术预发布基线已经完成；本 topic 只负责需要远端状态、版本决策和真实外部团队参与的工作。

## 下一步

1. 建立或确认远端仓库，运行首次真实 CI；除非用户明确要求，永远不自动 push。
2. 根据 CI 结果和兼容性承诺确定首个版本号、发布日期与 tag。
3. 按 `validation-plan.md` 招募至少 3 个外部团队并运行六周验证。
4. 用采用、节省成本和付费意愿证据决定继续商业化、保持开源 Toolkit，或停止产品化扩张。

## 立即读取

- `validation-plan.md`
- `../../knowledge/workflow/delivery-contract.md`
- `../../knowledge/workflow/open-source-license.md`
- `../../knowledge/workflow/verify.md`

## 按需读取

- `../../knowledge/security/untrusted-aep-input.md` — 设计公开服务或托管输入边界时
- `../../knowledge/architecture/public-sdk-interface.md` — 调整首发 SDK 承诺时
- `../../knowledge/architecture/unified-cli-interface.md` — 调整 CLI JSON contract 或集成入口时

## 已具备前提

- 根 Go SDK、统一 CLI、CI/release workflow、SECURITY、CONTRIBUTING、CHANGELOG 和 Apache-2.0 已落地。
- 代码质量综合升级已关闭全部 P0/P1，并建立 race、fuzz、漏洞与迁移矩阵门禁。
- OSS core、managed/on-prem 商业层边界已有初步方案。
- 原 `commercialization-and-toolkit-productization` 的产品承诺、Toolkit contract、交付、采用、商业验证与生态决策已合并到 `validation-plan.md`，旧 topic 不再单独恢复。

## 完成标准

- 远端首次 CI 通过，失败项均有关闭或明确接受记录。
- 首个版本号、发布日期和 tag 已确定并执行。
- 至少 3 个外部团队真实试用，至少 1 个进入日常 preflight，或明确记录未达到此门槛。
- 形成可量化的 adoption/成本/付费结论，并据此记录继续或停止决策。
- 明确首批核心用户、公开承诺、交付模型以及开源核心与商业层边界。
- 完成后压缩结果并归档；在外部参与尚未发生前保持活跃，不伪装为完成。

## 未决问题

- 首发是公开开源发布、私有 Toolkit 交付，还是两者并行？
- 首个版本采用 `v0.x` 预发布还是更保守的内部版本标识？
- 哪类设计伙伴最容易先提供合法 corpus 与真实失败标签？
