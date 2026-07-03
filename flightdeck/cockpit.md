# Cockpit — aep-parser

Focus: `codebase-hygiene-survey` 执行完成后补做 knowledge body validity pass：internal map、knowledge 分类/header repair、tmp safe-delete、generator provenance 修复/删除和 post-cleanup registry/capindex gates 均已落账；新增正文有效性审计，修 stale path 并标注缺失 generated evidence，不做简单相似文本合并。剩余 47 个 generated fixture missing outputs 属于独立 fixture regeneration 包。Booyah 复刻仍待用户真机验收；technique research 暂停在 glitch recipe/showcase gate 后。

## In flight

- **booyah-glitch-replication** (`work/booyah-glitch-replication/index.md`) — 12-comp from-scratch 复刻已建成并通过 agent AE gate；按 showcase review-gate，只有用户真机确认后才算 complete。①②③⑧ 已 complete；④⑤⑥⑦⑨⑩⑪⑫ 仍待用户 review，下一批优先 ⑩⑪⑫。
- **fx-technique-internalization** (`work/fx-technique-internalization/index.md`) — active research arc：参考 `.aep` → 拆角色/技法/机制 → 技法库/现象配方；effect-field understanding v1、report surface v1、pseudo/controller rebuild proof v1、pseudo behavior wiring proof v1、sample-fact behavior extraction v1、pseudo behavior payload application proof v1、multi-family pseudo application proof v1、vector pseudo behavior application v1、expression pseudo behavior boundary v1、glitch reference phenomenon v1、glitch showcase gate v1 已完成。下一步推荐另一个 reference phenomenon 包或基于 glitch recipe 的生成器升级。
- **technique-ontology** (`work/technique-ontology/index.md`) — parked research companion：角色/技法/机制三轴 ontology + schema v2。火焰/闪电/控制器样本已压测；大规模词表稳定性未验证。
- **codebase-hygiene-survey** (`work/codebase-hygiene-survey/index.md`) — completed cleanup survey：四张初始 ledger 已生成；knowledge header repair + README sync + tmp safe-delete batch 已完成；2 条 stale-codepath knowledge 已修；knowledge body validity pass 已补，修 stale path/缺失 generated evidence 注记，不做简单文本合并；26 个 generator delete-candidate 已完成 provenance 修复/删除；internal watch 已复核；post-cleanup registry/capindex gates 通过。

## Next

- **Fixture regeneration follow-up**：如后续需要恢复 `test_data/generated/fixtures`，为 `scripts/fixtures/regen_fixtures.ps1 -CheckOnly` 的 47 个 missing outputs 单独开包；不要混入已完成的 hygiene survey cleanup。
- **Booyah 用户真机验收 ⑩⑪⑫**：用户确认后更新 `flightdeck/showcase/booyah-clone/INDEX.md`，再按 review-gate 翻 complete；未确认前不要关闭 Booyah topic。
- **Technique research**：glitch showcase gate v1 已完成，Go/parse gate + AE2025/AE2020 render gate 均通过；下一步推荐另一个 reference phenomenon 包或基于 glitch recipe 的生成器升级。不要按单个字段继续。

## Open questions

- Booyah documented deltas 是否接受为限制，还是后续开 polish 包：Noise2、Curves arbitrary data、eased Position builder、颜色/撕裂强度细调。
- technique ontology 的大规模 N>>3 词表稳定性何时验证。
- 能力真相源仍是 `go run ./cmd/capindex -q <词>`；不要恢复单独 coverage 表。
