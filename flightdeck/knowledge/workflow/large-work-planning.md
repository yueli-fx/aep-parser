# 大型工作的 Plan 与 Slices

大型工作使用 `Plan → Slices → Plan` 保持单一完成汇总：先在 Work 的 `plan.md` 定义低分辨率的阶段顺序和完成状态，只为必须跨会话保存局部执行细节的 Plan 项建立链接 Slice，完成 Slice 后再把结果汇总回 Plan。

Slice 是可交付、可验证的单位，只拥有本地步骤、证据、Current 和 Next；它不复制 Work 的 Goal、Status、Current 或总完成度。只有子目标具有独立 Goal、Context、Current 和 Next 时才新建另一个 Work。
