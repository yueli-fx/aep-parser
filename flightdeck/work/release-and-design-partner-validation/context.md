# 首次发布与设计伙伴验证：背景

## Why this Work exists

技术能力已经达到预发布基线，但产品价值尚未被真实团队采用证明。这个 Work 把剩余风险集中在远端 CI、版本承诺、外部 corpus、工作流接入和付费意愿上；不要用继续扩展 AEP 字段数量代替采用验证。

## Product promise

给定一个 AEP，Toolkit 应在明确资源边界内回答：工程包含什么、依赖哪些素材/字体/插件/版本能力、是否适合目标环境、两个版本有何变化、哪些迁移可自动完成、哪些必须人工确认，以及结论由哪一级测试或 AE 证据支持。

首发不承诺完整 authoring。公开 contract 优先稳定 inspect、profile、diff、保守 migration、结构化错误和证据等级，避免实验性 setter 污染稳定 interface。

## Validation population

优先覆盖三类设计伙伴：

1. 模板市场或素材平台：20–50 个有人工审核结论的合法模板。
2. 批量视频或渲染团队：约 100 个带成功/失败标签的 render jobs。
3. 工作室 Pipeline 或 DAM：一个授权历史项目目录和 2–3 个真实升级任务。

## Measures and decision threshold

核心指标包括已知失败原因召回率与误报率、单工程审核耗时、减少的人工打开次数、提前拦截的失败比例、节省的 worker 分钟、preflight P95、corpus 可解析比例、升级人工时、unsupported 解释准确度、日常 workflow 接入和付费意愿。

继续产品化的最低信号是：至少 3 个外部团队真实试用、至少 1 个进入日常流程、至少 1 个表达明确付费意愿。若六周后只有技术兴趣而没有 workflow adoption，则停止 SaaS 扩张，保留大型开源或私有 SDK/CLI Toolkit，并只由真实 issue 驱动能力扩展。

## Boundaries

- 只修复至少命中两个真实样本或阻塞主 workflow 的问题；其他扩展建立独立 Work。
- 验证先以 shadow mode 接入，不自动阻断生产；只有高置信规则才能有限阻断。
- 开源与否不影响技术价值；私有 Toolkit、on-prem 和集成服务都是有效交付路径。
- 基础 parser correctness、安全与互操作能力不因商业包装弱化。
- 只有出现至少两个真实 adapter 后才抽象新的 extension port。
- 任何 push、tag、发布或外部联络都需要用户明确授权。

## Delivery options under evaluation

- 可下载 CLI/SDK 与容器；
- 企业 on-prem、升级保障和长期支持；
- 托管 API、队列、鉴权、配额、审计与 webhook；
- DAM、nexrender/aerender、render farm 和私有插件连接器；
- 历史 corpus 盘点、迁移与 pipeline 集成服务。

商业价值按节省的人工检查时间、避免的失败渲染、资产治理规模和采购意愿衡量，不用下载量代替生产采用。
