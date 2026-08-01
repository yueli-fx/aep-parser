# 首次发布与六周外部验证计划

## 阶段 A：首次发布准备

- [ ] 与用户确认公开开源、私有 Toolkit 或并行交付路径。
- [ ] 配置或确认 remote，运行首次真实 CI；不自动 push。
- [ ] 核对 release workflow、安装说明、安全报告入口和 Apache-2.0 元数据。
- [ ] 核对 Go SDK、CLI JSON envelope、错误分类和兼容承诺；每项公开能力可关联验证证据。
- [ ] 确定版本号、发布日期、兼容范围与已知限制；获得授权后才创建 tag 或发布。

## 阶段 B：设计伙伴招募与基线

- [ ] 每类候选伙伴联系 3–5 个团队，最终至少 3 个团队进入真实试用。
- [ ] 确认 corpus 授权、敏感数据边界和离线/托管处理约束。
- [ ] 记录现有失败率、平均排查时间、worker 浪费和人工打开工程次数。
- [ ] 只承诺可解释报告，不承诺覆盖全部 AEP。

## 阶段 C：六周验证

- [ ] Week 1：建立伙伴当前流程与成本基线。
- [ ] Week 2–3：离线运行统一 CLI，将失败分为 parser defect、unsupported、dependency issue、version boundary 或 bad input。
- [ ] Week 4：以 shadow mode 接入 render/DAM pipeline，不自动阻断生产。
- [ ] Week 5：仅让无法解析、明确版本边界、确定缺失关键依赖等高置信规则有限阻断。
- [ ] Week 6：复盘节省的人工时、worker 时间、避免的失败任务和继续付费意愿。

## 阶段 D：路线决策与收尾

- [ ] 汇总采用、成本、可靠性和付费证据。
- [ ] 决定继续商业化、保持开源或私有 Toolkit，或停止产品化扩张。
- [ ] 记录首批核心用户、公开承诺、交付模型以及开源核心与商业层边界。
- [ ] 将完成项、剩余风险和最终结论同步回 `index.md`，满足 Goal 后将 Status 设为 `Finished`、Next 设为 `None`，并从 deck 的 Open Work 移除。
