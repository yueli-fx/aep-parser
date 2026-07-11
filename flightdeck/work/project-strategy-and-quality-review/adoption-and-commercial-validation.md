# 采用与商业验证方案

## 目标

不要先证明“还能做更多 AEP 字段”，而要证明：无 AE 的 inspect、diff、migration preflight 能否减少真实团队的失败渲染、人工排查和版本兼容成本。

## 三类设计伙伴

### 1. AE 模板市场 / 素材平台

典型问题：上传工程版本混乱、缺字体/插件/素材、模板结构不规范、买家打不开。

试用输入：20–50 个可合法共享的上架/退回模板及其人工审核结论。

验证输出：

- 无 AE inventory/profile；
- plugin/font/source/版本风险报告；
- 同模板版本之间的 semantic diff；
- AE2020–AE2025 migration 可行性和损失说明。

成功指标：自动发现已知退回原因的召回率、误报率、单工程审核耗时、减少的人工打开次数。

### 2. 批量视频 / 模板渲染团队

典型问题：任务进入 aerender/nexrender 后才发现工程损坏、版本错误、缺依赖或模板漂移，浪费 worker 时间。

试用输入：最近 100 个成功/失败 render jobs 的 AEP、失败分类和运行耗时。

验证输出：在提交 worker 前运行 inspect/profile/diff/migrate preflight，给出 pass/blocked JSON 和可追踪原因。

成功指标：提前拦截的失败比例、减少的无效 worker 分钟、preflight P95、重复问题自动归类率。

### 3. 工作室 Pipeline / DAM 团队

典型问题：历史 AEP 无法检索、项目依赖不透明、大版本升级需要人工逐个打开、资产合规难审计。

试用输入：一个已授权的历史项目目录和 2–3 个真实升级任务。

验证输出：跨平台 corpus inventory、版本分布、依赖/效果索引、迁移队列和差异报告。

成功指标：可解析比例、索引耗时、升级人工时节省、无法自动迁移项是否有准确解释。

## 六周验证节奏

### Week 1：招募与基线

- 每类联系 3–5 个团队，争取至少 1 个有效设计伙伴。
- 记录他们当前流程、失败率、平均排查时间和不能上传的敏感边界。
- 不承诺覆盖全部 AEP；只承诺产出可解释报告。

### Week 2–3：离线试跑

- 用统一 `aep` CLI 跑客户授权样本。
- 将所有失败归为 parser defect、unsupported、dependency issue、version boundary 或 bad input。
- 只修复至少命中两个真实样本或阻塞主 workflow 的问题。

### Week 4：接入 preflight

- 在现有 render/DAM pipeline 前插入 CLI 或本地 HTTP service。
- shadow mode 运行，不自动阻断生产。
- 对比工具判断与人工/真实 render 结果。

### Week 5：有限阻断

- 只让高置信度规则阻断：无法解析、明确版本 boundary、确定缺失关键依赖。
- 其他结果保持 warning，收集误报。

### Week 6：商业复盘

- 每类计算节省的人工时、worker 时间和避免的失败任务。
- 只有存在可量化价值且对方愿意继续时，才进入报价/产品化。

## OSS 与商业层边界

### 保持开源

- 安全 RIFX parser、根 Go SDK、统一 CLI；
- inspect/profile/diff 和保守 migration core；
- profile/recipe/schema、compatibility matrix 和基础 HTTP handler；
- fuzz、安全修复、通用格式兼容和可复现测试。

正确性、安全性和基础互操作不能放到付费墙后；它们是获得 corpus、issue 和社区信任的条件。

### 可收费

- 托管 API：队列、存储、鉴权、配额、审计、团队空间和 webhook；
- 企业 on-prem：部署、升级、监控、SLA、SSO、私有网络和数据保留策略；
- 行业 policy packs：模板市场上架规则、工作室命名/依赖规范、迁移审批；
- 私有适配：专有 effect/plugin、内部 chunk、资产系统和 render farm 集成；
- 版本保障：新 AE 发布后的优先兼容、迁移评估和受控 rollout；
- 专业服务：历史 corpus 盘点、迁移项目和 pipeline 改造。

## 定价实验（先验证，不写死）

- Managed API：免费小额度 + 按文件/处理量计费；验证客户是否愿意为避免一次失败渲染付费。
- On-prem：年度订阅，按团队/worker 或支持等级报价；核心价值是数据不出网与版本保障。
- 迁移/集成项目：固定范围服务费，交付报告、adapter 和后续维护选项。

暂不采用把核心 parser 改成闭源或过早 dual-license 的路线。先让开源 interface 成为事实标准，再根据企业采用决定商业包装。

## 传播材料

首批内容只讲三个可复现故事：

1. 不启动 AE，在 Linux/macOS/Windows 上秒级 inventory 一个 AEP。
2. 在 render worker 前发现版本/依赖/结构问题，避免失败任务。
3. 将 AE2020 工程迁移到目标版本，并明确列出 preserved/translated/blocked。

不把“AI 学会做火焰/故障效果”作为首发叙事；它会模糊基础设施定位。

## 继续与停止条件

继续投入的最低信号：

- 至少 3 个外部团队真实试用；
- 至少 1 个团队把 preflight 接入日常流程；
- 至少 1 个团队愿意为托管、私有部署、版本保障或集成付费。

若六周后只有“技术很酷”但没有 workflow adoption，停止 SaaS 建设，保持大型开源 SDK/CLI 路线，并只由真实 issue 驱动扩展。
