# Flightdeck

`flightdeck/` 使用普通 Markdown 保存可恢复的长期工作与项目知识；Git 负责历史，不维护隐藏状态、版本化 schema 或平行工作流图。

## 结构

- `deck.md`：极小的导航页，只列 Open Work、唯一 Focus 和少量稳定项目链接。
- `work/<work-id>/index.md`：Work 的恢复权威，保存 Goal、Status、Current、Next 和当前执行指针。
- `work/<work-id>/context.md`：稳定的目标背景、约束、术语和决策。
- `work/<work-id>/plan.md`：可选的阶段顺序与完成汇总；只有需要跨会话保留局部执行细节时才建立 `slices/`。
- `work/<work-id>/references/`：可选的 Work 自有支撑材料。
- `knowledge/<subject>/<topic>.md`：按自然主题组织、可独立应用的项目实践。
- `showcase/`：项目资产，不属于 Flightdeck 恢复模型，但继续随仓库维护。

## 恢复与维护

恢复时先读 `deck.md`，再完整读取 Focus Work 的 `index.md` 与 `context.md`；若存在 `plan.md`，只读取其低分辨率顺序和完成状态，再跟随 Next 中最多三个立即所需的本地链接。随后核对 Git 状态与实际仓库内容。

Knowledge 按路径和标题做有界发现，只读取与 Goal、Current、Next 明确相关的主题。知识文件不要求 frontmatter、路由字段或总索引；新知识应是经验证、自洽且可复用的一条项目实践。

Work 完成或停止时，在原目录保留其页面与材料，将 Status 改为 `Finished` 或 `Stopped`、Next 改为 `None`，并从 `deck.md` 的 Open Work 列表移除；不要为了生命周期仪式归档或改名。
