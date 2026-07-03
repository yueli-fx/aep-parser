# Cockpit — aep-parser

Focus: 新增 `codebase-hygiene-survey` 作为下一阶段重点调查包：先建立 `internal` 代码基座地图、knowledge 生命周期审计、JSX generator 引用矩阵、`tmp` JSON 清理判定，再分批执行清理。Booyah 复刻仍待用户真机验收；technique research 暂停在 glitch recipe/showcase gate 后。

## In flight

- **booyah-glitch-replication** (`work/booyah-glitch-replication/index.md`) — 12-comp from-scratch 复刻已建成并通过 agent AE gate；按 showcase review-gate，只有用户真机确认后才算 complete。①②③⑧ 已 complete；④⑤⑥⑦⑨⑩⑪⑫ 仍待用户 review，下一批优先 ⑩⑪⑫。
- **fx-technique-internalization** (`work/fx-technique-internalization/index.md`) — active research arc：参考 `.aep` → 拆角色/技法/机制 → 技法库/现象配方；effect-field understanding v1、report surface v1、pseudo/controller rebuild proof v1、pseudo behavior wiring proof v1、sample-fact behavior extraction v1、pseudo behavior payload application proof v1、multi-family pseudo application proof v1、vector pseudo behavior application v1、expression pseudo behavior boundary v1、glitch reference phenomenon v1、glitch showcase gate v1 已完成。下一步推荐另一个 reference phenomenon 包或基于 glitch recipe 的生成器升级。
- **technique-ontology** (`work/technique-ontology/index.md`) — parked research companion：角色/技法/机制三轴 ontology + schema v2。火焰/闪电/控制器样本已压测；大规模词表稳定性未验证。
- **codebase-hygiene-survey** (`work/codebase-hygiene-survey/index.md`) — new priority survey：已完成初扫并写入 spec；下一步是用户 review 后按四张 ledger 执行（internal map / knowledge audit / generator audit / tmp cleanup）。

## Next

- **Codebase hygiene survey**：review `work/codebase-hygiene-survey/design.md`，确认后先生成四张 ledger，不直接删除文件；`verify_v2_2_*` 仍属 active shape gate，不按名称清理。
- **Booyah 用户真机验收 ⑩⑪⑫**：用户确认后更新 `flightdeck/showcase/booyah-clone/INDEX.md`，再按 review-gate 翻 complete；未确认前不要关闭 Booyah topic。
- **Technique research**：glitch showcase gate v1 已完成，Go/parse gate + AE2025/AE2020 render gate 均通过；下一步推荐另一个 reference phenomenon 包或基于 glitch recipe 的生成器升级。不要按单个字段继续。

## Open questions

- Booyah documented deltas 是否接受为限制，还是后续开 polish 包：Noise2、Curves arbitrary data、eased Position builder、颜色/撕裂强度细调。
- codebase-hygiene-survey 的 README 架构更新是否作为第一批 deliverable，还是等清理完成后再同步。
- generator 清理候选是先 cold archive 还是经 gate 后直接删除。
- technique ontology 的大规模 N>>3 词表稳定性何时验证。
- 能力真相源仍是 `go run ./cmd/capindex -q <词>`；不要恢复单独 coverage 表。
