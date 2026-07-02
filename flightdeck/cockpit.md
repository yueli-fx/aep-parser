# Cockpit — aep-parser

Focus: 当前 active execution 是 `work/versioned-aep-migration/index.md`：AE2020-AE2025 versioned migration matrix、coverage JSON、batch inventory 和 registry gates。旧 `aep-understanding-generation` 大包已拆分；已完成的 umbrella/profile/recipe 历史阶段已归档到 cold archive。Booyah 复刻仍是压力测试与用户验收基准。2026-07-01 看板清理已把完成 arc 移入 cold archive：`apidoc-tag-schema`、`cross-platform-architecture`、`project-index`、`recipe-docgen`、`scene-architecture`。

## In flight

- **versioned-aep-migration** (`work/versioned-aep-migration/index.md`) — active mainline：AE2020-AE2025 migration matrix、coverage/current JSON、domain batch inventory、host-open planning、registry checkpoint。继续前读 `goal.md` / `current.md`，下一批是 `text_domain_residuals`；不要把一次 package cycle 当成整条主线完成。
- **asset-registry-cleanup** (`work/asset-registry-cleanup/index.md`) — parked governance arc：data/examples/test_data/tmp/registry ownership、generated-output cleanup policy、asset ledger normalization。
- **booyah-glitch-replication** (`work/booyah-glitch-replication/index.md`) — 12-comp from-scratch 复刻已建成并通过 agent AE gate；按 showcase review-gate，只有用户真机确认后才算 complete。①②③⑧ 已 complete；④⑤⑥⑦⑨⑩⑪⑫ 仍待用户 review，下一批优先 ⑩⑪⑫。
- **fx-technique-internalization** (`work/fx-technique-internalization/index.md`) — parked research arc：参考 `.aep` → 拆角色/技法/机制 → 技法库/现象配方；现在也拥有 technique facts/portrait/corpus/report 文件。火焰样本已验证；下次恢复需要先选新的参考现象。
- **technique-ontology** (`work/technique-ontology/index.md`) — parked research companion：角色/技法/机制三轴 ontology + schema v2。火焰/闪电/控制器样本已压测；大规模词表稳定性未验证。

## Next

- **versioned AEP migration**：从 `work/versioned-aep-migration/goal.md` 执行连续 package queue；下一批 `text_domain_residuals`，每个 package 完成后 commit/checkpoint 并自动进入下一包，直到无 candidate 或明确 blocked。不要默认跑全量 `-ae-open`，除非显式扩大/关闭 `-max-ae-open-cases`。
- **Booyah 用户真机验收 ⑩⑪⑫**：用户确认后更新 `flightdeck/showcase/booyah-clone/INDEX.md`，再按 review-gate 翻 complete；未确认前不要关闭 Booyah topic。
- **Technique research**：只有用户明确要继续学习/内化参考工程时才恢复；先选样本，不在主线自动推进。

## Open questions

- Booyah documented deltas 是否接受为限制，还是后续开 polish 包：Noise2、Curves arbitrary data、eased Position builder、颜色/撕裂强度细调。
- technique ontology 的大规模 N>>3 词表稳定性何时验证。
- 能力真相源仍是 `go run ./cmd/capindex -q <词>`；不要恢复单独 coverage 表。
