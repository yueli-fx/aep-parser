# Cockpit — aep-parser

Focus: `codebase-hygiene-survey` 已完成第二批审计与低风险 provenance 修复：2 条 stale-codepath knowledge 已修/重分类，26 个 generator delete-candidate 已拆为 keep-active 10、historical-evidence-review 1、archive/delete-candidate 15、direct-delete-now 0，internal 大包 watch 已确认不在本 survey 内拆包。下一步是决定历史 `verify_ge_duplicate_layer_explicit_matte.jsx` evidence 和 15 个候选的产物/基线处置策略。Booyah 复刻仍待用户真机验收；technique research 暂停在 glitch recipe/showcase gate 后。

## In flight

- **booyah-glitch-replication** (`work/booyah-glitch-replication/index.md`) — 12-comp from-scratch 复刻已建成并通过 agent AE gate；按 showcase review-gate，只有用户真机确认后才算 complete。①②③⑧ 已 complete；④⑤⑥⑦⑨⑩⑪⑫ 仍待用户 review，下一批优先 ⑩⑪⑫。
- **fx-technique-internalization** (`work/fx-technique-internalization/index.md`) — active research arc：参考 `.aep` → 拆角色/技法/机制 → 技法库/现象配方；effect-field understanding v1、report surface v1、pseudo/controller rebuild proof v1、pseudo behavior wiring proof v1、sample-fact behavior extraction v1、pseudo behavior payload application proof v1、multi-family pseudo application proof v1、vector pseudo behavior application v1、expression pseudo behavior boundary v1、glitch reference phenomenon v1、glitch showcase gate v1 已完成。下一步推荐另一个 reference phenomenon 包或基于 glitch recipe 的生成器升级。
- **technique-ontology** (`work/technique-ontology/index.md`) — parked research companion：角色/技法/机制三轴 ontology + schema v2。火焰/闪电/控制器样本已压测；大规模词表稳定性未验证。
- **codebase-hygiene-survey** (`work/codebase-hygiene-survey/index.md`) — active cleanup survey：四张初始 ledger 已生成；knowledge header repair + README sync + tmp safe-delete batch 已完成并通过 gates；2 条 stale-codepath knowledge 已修；26 个 generator delete-candidate 已细分并修复 4 个 manifest/provenance gap；internal watch 已复核。下一步决定 1 个 historical evidence review 和 15 个 archive/delete candidate 的 fixture/baseline 策略。

## Next

- **Codebase hygiene survey**：决定 `work/codebase-hygiene-survey/ledgers/generator-provenance-review.md` 里的 `verify_ge_duplicate_layer_explicit_matte.jsx` historical evidence policy，再决定 15 个 `archive-or-delete-candidate` 是否连同 tracked fixture / split-roundtrip baseline hash 一起删、迁移为 durable evidence，或保留为历史基线。
- **Booyah 用户真机验收 ⑩⑪⑫**：用户确认后更新 `flightdeck/showcase/booyah-clone/INDEX.md`，再按 review-gate 翻 complete；未确认前不要关闭 Booyah topic。
- **Technique research**：glitch showcase gate v1 已完成，Go/parse gate + AE2025/AE2020 render gate 均通过；下一步推荐另一个 reference phenomenon 包或基于 glitch recipe 的生成器升级。不要按单个字段继续。

## Open questions

- Booyah documented deltas 是否接受为限制，还是后续开 polish 包：Noise2、Curves arbitrary data、eased Position builder、颜色/撕裂强度细调。
- 15 个 generator archive/delete candidates 的 tracked fixture 产物和 split-roundtrip baseline hash：删除、迁移为 durable evidence，还是作为历史基线保留。
- technique ontology 的大规模 N>>3 词表稳定性何时验证。
- 能力真相源仍是 `go run ./cmd/capindex -q <词>`；不要恢复单独 coverage 表。
