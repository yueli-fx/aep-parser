# Cockpit — aep-parser

Focus: `codebase-hygiene-survey` 已完成并归档到 cold archive；当前活跃面回到 Booyah 真机验收、technique research、technique ontology。剩余 47 个 generated fixture missing outputs 属于独立 fixture regeneration follow-up，不再挂在 hygiene survey。

## In flight

- **booyah-glitch-replication** (`work/booyah-glitch-replication/index.md`) — 12-comp from-scratch 复刻已建成并通过 agent AE gate；按 showcase review-gate，只有用户真机确认后才算 complete。①②③⑧ 已 complete；④⑤⑥⑦⑨⑩⑪⑫ 仍待用户 review，下一批优先 ⑩⑪⑫。
- **fx-technique-internalization** (`work/fx-technique-internalization/index.md`) — active research arc：参考 `.aep` → 拆角色/技法/机制 → 技法库/现象配方；effect-field understanding v1、report surface v1、pseudo/controller rebuild proof v1、pseudo behavior wiring proof v1、sample-fact behavior extraction v1、pseudo behavior payload application proof v1、multi-family pseudo application proof v1、vector pseudo behavior application v1、expression pseudo behavior boundary v1、glitch reference phenomenon v1、glitch showcase gate v1、rain reference phenomenon v1、rain showcase gate v1 已完成。下一步推荐另一个 reference phenomenon 包、richer rain generator，或基于 glitch recipe 的生成器升级。
- **technique-ontology** (`work/technique-ontology/index.md`) — parked research companion：角色/技法/机制三轴 ontology + schema v2。火焰/闪电/控制器样本已压测；大规模词表稳定性未验证。

## Next

- **Arbitrary AEP version migration**
  (`work/arbitrary-aep-version-migration/index.md`) — parked future task for
  upgrading/downgrading arbitrary AE projects. When other work finds
  high-version-only fields, version-specific field layouts, or downgrade
  fallback needs, record them in this package instead of scattering them through
  temporary notes.
- **Fixture regeneration follow-up**：如后续需要恢复 `test_data/generated/fixtures`，为 `scripts/fixtures/regen_fixtures.ps1 -CheckOnly` 的 47 个 missing outputs 单独开包；不要混入已完成的 hygiene survey cleanup。
- **Booyah 用户真机验收 ⑩⑪⑫**：用户确认后更新 `flightdeck/showcase/booyah-clone/INDEX.md`，再按 review-gate 翻 complete；未确认前不要关闭 Booyah topic。
- **Technique research**：rain reference phenomenon v1 已完成并记录 `build-good-rain.md`；rain showcase gate v1 已完成，Go/parse gate + AE2020 render gate 通过；glitch showcase gate v1 已完成，Go/parse gate + AE2025/AE2020 render gate 均通过；下一步推荐另一个 reference phenomenon 包、richer rain generator，或基于 glitch recipe 的生成器升级。不要按单个字段继续。

## Open questions

- Booyah documented deltas 是否接受为限制，还是后续开 polish 包：Noise2、Curves arbitrary data、eased Position builder、颜色/撕裂强度细调。
- technique ontology 的大规模 N>>3 词表稳定性何时验证。
- 能力真相源仍是 `go run ./cmd/capindex -q <词>`；不要恢复单独 coverage 表。
