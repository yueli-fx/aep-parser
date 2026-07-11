# 执行方案 — 首次发布与六周外部验证

## 目标

验证无 AE 的 inspect、diff 和 migration preflight 是否能减少真实团队的失败渲染、人工排查与版本兼容成本。不要以继续扩展 AEP 字段数量代替采用验证。

## 产品承诺

给定一个 AEP，Toolkit 应在明确资源边界内回答：工程包含什么、依赖哪些素材/字体/插件/版本能力、是否适合目标环境、两个版本有何变化、哪些迁移可自动完成、哪些必须人工确认，以及结论由哪一级测试或 AE 证据支持。

首发不承诺完整 authoring。公开 contract 优先稳定 inspect、profile、diff、保守 migration、结构化错误和证据等级，避免实验性 setter 污染稳定 interface。

## 阶段 A：首次发布准备

- 建立远端并运行真实 CI，不自动 push。
- 核对 release workflow、安装说明、安全报告入口和 Apache-2.0 元数据。
- 确定版本号、发布日期、兼容范围与已知限制后再创建 tag。
- 若选择私有交付，同样保留可复现构建、版本标识、变更记录与安全响应流程。
- 核对 Go SDK、CLI JSON envelope、错误分类和兼容承诺；每项公开能力必须能关联验证证据。

## 阶段 B：设计伙伴招募

优先覆盖三类伙伴：

1. 模板市场/素材平台：20–50 个有人工审核结论的合法模板。
2. 批量视频/渲染团队：约 100 个带成功/失败标签的 render jobs。
3. 工作室 Pipeline/DAM：一个授权历史项目目录和 2–3 个真实升级任务。

记录现有失败率、平均排查时间、worker 浪费和敏感数据边界。只承诺可解释报告，不承诺覆盖全部 AEP。

## 阶段 C：六周验证

- Week 1：每类联系 3–5 个团队，建立当前流程和成本基线。
- Week 2–3：离线运行统一 CLI，把失败分为 parser defect、unsupported、dependency issue、version boundary 或 bad input。
- Week 4：以 shadow mode 接入 render/DAM pipeline，不自动阻断生产。
- Week 5：只让无法解析、明确版本边界、确定缺失关键依赖等高置信规则有限阻断。
- Week 6：复盘节省的人工时、worker 时间、避免的失败任务和继续付费意愿。

只修复至少命中两个真实样本或阻塞主 workflow 的问题；其他扩展进入独立 topic，不污染验证周期。

## 衡量指标

- 已知失败原因召回率和误报率；
- 单工程审核耗时与减少的人工打开次数；
- 提前拦截失败比例、减少的无效 worker 分钟和 preflight P95；
- corpus 可解析比例、升级人工时节省与 unsupported 解释准确度；
- 是否进入日常 workflow，以及是否愿意为托管、私有部署、版本保障或集成付费。

## 决策门槛

继续产品化的最低信号：至少 3 个外部团队真实试用、至少 1 个接入日常流程、至少 1 个表达明确付费意愿。

若六周后只有“技术很酷”而没有 workflow adoption，停止 SaaS 扩张，保留大型开源或私有 SDK/CLI Toolkit，并只由真实 issue 驱动能力扩展。

## 交付与商业决策

验证期间比较以下交付组合，不预设开源一定优于私有，也不先建 SaaS：

- 可下载 CLI/SDK 与容器；
- 企业 on-prem、升级保障和长期支持；
- 托管 API、队列、鉴权、配额、审计与 webhook；
- DAM、nexrender/aerender、render farm 和私有插件连接器；
- 历史 corpus 盘点、迁移与 pipeline 集成服务。

商业价值按节省的人工检查时间、避免的失败渲染、资产治理规模和采购意愿衡量，不用下载量代替生产采用。只有出现至少两个真实 adapter 后才抽象新的 extension port。

## 边界

- 开源与否不影响技术价值；私有 Toolkit、on-prem 和集成服务都是有效交付路径。
- 基础 parser correctness、安全与互操作能力不应因商业包装而弱化。
- 托管队列、鉴权、审计、SSO、policy packs、私有插件适配、版本保障和专业服务可作为商业层候选，但必须由真实验证触发。
