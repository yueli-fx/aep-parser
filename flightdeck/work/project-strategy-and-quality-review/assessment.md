# aep-parser 项目方向与代码质量初评

## 结论

这个项目不是“方向太少”，而是同时长出了过多方向，却没有选定一个对外承诺。

底层能力和验证资产已经显著超过普通实验项目：纯 Go 解析/写回、场景模型、结构性修改、从零生成、语义 profile/diff、版本迁移、AE 双版本验收、render-pixel gate、recipe、HTTP 服务和能力索引都已存在。真正的问题是产品边界、外部消费面和不可信输入安全性仍处于原型阶段。

建议把主线收束为：

> **Headless AEP Toolkit：无需启动 After Effects，对 AEP 做 inspect、lint、diff、safe edit、generate 和 migrate，并对数据损失给出可机器消费的证据。**

复刻与技法学习不再作为主线。它们保留为压力测试、案例和未来高级产品能力，但不继续按“再学一个现象”扩张。

## 代码质量画像

### 强项

- 包依赖 DAG 清楚，`rifx/codec -> scene/serializer -> aep -> tools` 的方向检查通过。
- 外部依赖极少；当前仅直接依赖 `golang.org/x/text`。
- `go vet ./...`、capindex drift check 和 Windows/macOS/Linux 构建门禁通过。
- 497 个已登记能力中，435 个标为 stable；验证分类包括 266 个 AE acceptance、163 个 render-pixel、62 个 roundtrip。
- 大量公开黑盒测试集中在 `internal/aep_test`，并且对 AE2020/AE2025、结构 mutation、render/readback 有强验收意识。
- 未知 chunk 的保留、length-variable 重排、能力边界与 incident 记录体现出较成熟的逆向工程治理。

### 主要风险

1. **当前基线为红。** 未提交的 shape transform 编译改动令多项 `internal/aepmigrate` 用例把 Position 从预期值迁移成 `[0,0,0]`。vet、capindex 和跨平台构建仍通过，但 `go test ./...` 不通过。
2. **所谓公共 Go API 实际不可被外部项目导入。** facade 位于 `internal/aep`，仓库根目录没有 Go 包；Go 的 `internal` 规则会阻止模块外调用者导入它。
3. **不可信输入安全性不足。** `internal/rifx` 直接信任 chunk 声明尺寸、无递归深度/节点数预算；异常 LIST size 可导致负长度或超大分配，服务层也没有 parser panic 隔离。
4. **HTTP 服务仍是本地实验面。** 虽有限制请求体大小，但 `http.Server` 未设置读写/idle timeout；启用 path input 后没有 configured-root 限制。
5. **核心局部复杂度过高。** `internal/aep/facade.go` 约 2431 行，多个 scene/serializer 文件超过 800–1200 行；recipe 两个测试文件分别约 5408/3137 行。分层合理不等于模块内部已经易维护。
6. **低层防御性测试不足。** `rifx` 和 `codec` 没有本包测试，仓库没有 fuzz test。高层 fixture/ship-gate 很强，但不能替代畸形二进制、资源预算和 parser invariant 测试。
7. **对外产品面过散。** 现有 17 个 `cmd` 入口覆盖生成、迁移、注册表、研究、自托管、diff、search、oracle 等，用户难以判断主入口。
8. **尚不具备开源发布基本面。** 当前没有 LICENSE、CONTRIBUTING、SECURITY、CI、tag、remote、安装文档，也没有外部可导入的 Go package。

## 外部生态判断

- `py-aep` 仍在快速发展，已经提供读取、写入、从零项目、属性/关键帧、图层、render queue 等广覆盖能力。不要把路线定义为“逐字段追平 Python 项目”。
- 老的 Go `boltframe/aftereffects-aep-parser` 证明了静态分析的需求，但覆盖面远小于本项目。
- `nexrender` 已占据 AE 渲染编排/模板视频自动化心智，并以开源核心、云服务和定制基础设施变现。与它正面竞争 render farm 没有优势。
- 本项目真正稀缺的组合是：**不启动 AE 的二进制修改 + 未知数据保留 + 版本迁移 + 可验证的损失报告 + Go/服务部署**。这应成为差异化。

## 推荐产品结构

### 开源核心

- 对外 Go SDK：`aep` 或 `pkg/aep`，提供稳定的 parse/profile/diff/edit/write API。
- 单一 CLI：`aep inspect|lint|diff|edit|build|migrate|capabilities`。
- 规范化 IR/profile、兼容性报告、recipe schema 和确定性输出。
- 安全解析预算、fuzz corpus、公开 compatibility matrix。

### 可变现层

- 托管 API：批量 inspect/lint/diff/migrate，按文件或计算量收费。
- 企业/on-prem：素材平台、模板市场、工作室资产库的合规检查、版本迁移和批量修复。
- 高价值支持：私有 chunk/effect 适配、版本升级保障、迁移报告和定制规则包。
- 与现有 render pipeline 集成：在 nexrender/aerender 之前完成预检、迁移和风险报告，而不是自己重做调度与渲染农场。

### 保留但降级的研究线

- Booyah/雨/火焰等 showcase：作为能力证明和回归样本，不作为路线图主轴。
- technique ontology/selfhost：冻结 schema 扩张；只有明确客户用例（模板检索、相似度、自动拆解）出现时再产品化。
- 单字段逆向：仅由主产品缺口、真实 issue 或版本迁移需求驱动。

## 分阶段路线

### Phase 0：恢复可信基线（2–4 周）

- 修复或回退当前 shape transform 迁移回归，让 `go test ./...` 恢复全绿。
- 为 RIFX 增加尺寸、深度、节点数、总分配预算和 panic-free malformed-input tests/fuzzing。
- 明确支持的输入信任模型；服务增加 timeout、recover 和 path root 限制。
- 补齐外部 Go package，先做最小稳定 API，不直接暴露 1000+ 内部符号。

### Phase 1：做出一个可传播产品（4–8 周）

- 合并用户入口为单一 `aep` CLI；其余命令保留为 internal/advanced tooling。
- 只打磨一个 killer workflow：`inspect + lint + diff + migrate report`。
- 提供 5–10 个真实案例、性能对比、JSON schema、Docker image 和可复制 demo。
- 添加 LICENSE、SECURITY、CONTRIBUTING、CI、release/tag 和安装说明。

### Phase 2：验证采用与付费（2–4 月）

- 找 3 类设计伙伴：AE 模板/素材平台、批量视频团队、工作室 pipeline/资产管理团队。
- 以真实文件测试“无 AE 预检能减少多少失败”“跨版本迁移能节省多少人工”。
- 开源核心获取 issue/corpus；收费提供托管 API、私有部署、规则包和版本保障。

### Phase 3：选择规模化方向

- 若开发者采用强：做大型开源 AEP SDK/CLI 生态。
- 若企业迁移/质检需求强：做 on-prem + managed API。
- 若 recipe 生成需求强：再建设 headless AEP compiler/template platform。
- technique learning 只有在出现明确付费场景后才升级为独立产品。

## 建议立即停止的工作方式

- 不再以“还能加哪个字段/效果”为默认下一步。
- 不再把一次视觉复刻完成当作项目战略进展。
- 不在没有外部用户入口、安全边界和发布面的情况下继续扩大内部工具数量。
- 不同时推进开源 SDK、SaaS、渲染农场和技法 AI；先用一个 killer workflow 验证用户。

## 下一轮评审建议

对底层做正式 review 时，不要一次审 700 个文件。按风险切片：

1. `rifx + codec`：畸形输入、安全预算、roundtrip invariant。
2. `scene + serializer`：状态一致性、事务/回滚、opaque preservation。
3. `aep facade`：外部 API 深度、稳定性和最小公开面。
4. `recipe + migrate`：IR 重复、编译路径和当前 Position 回归。
5. `server + CLI`：产品接口、安全和可运维性。

每个切片产出 findings、测试缺口和可独立提交的修复包，最后汇总成质量基线。
