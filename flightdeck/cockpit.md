# Cockpit — aep-parser

Focus: 当前主线是 `work/aep-understanding-generation/index.md` 下的 AEP 理解/诊断/生成流水线，当前工作树正在推进 versioned AEP migration matrix。Booyah 复刻仍是压力测试与用户验收基准。2026-07-01 看板清理已把完成 arc 移入 cold archive：`apidoc-tag-schema`、`cross-platform-architecture`、`project-index`、`recipe-docgen`、`scene-architecture`。

## In flight

- **aep-understanding-generation** (`work/aep-understanding-generation/index.md`) — 长期流水线：profile → diff → render oracle → gap ledger → slice workflow → recipe IR → technique facts/portrait/report → versioned AEP migration。当前恢复点是 versioned AEP migration：`cmd/aepmigrate assess/convert` 已有保守转换路径，convert 现已覆盖 supported non-shape layer metadata/switches/refs、dynamic layer transforms、single-run text style、solid-backed null-controller、camera/light options/light-source refs、supported built-in effects、effect layer-ref params、effect param expressions、scalar keyframes、vector keyframes；writer target 现覆盖 AE2020-AE2025，text baseline / text animator / text style / static effects / auto-orient / layer mask / shape gradient stroke / project / camera / light / structural representatives / dynamic transform / dynamic effect-param 均已有 PD-6x6 writer evidence，classic track matte 已记录 31/36 all-writer boundary（AE2025 source 降 W2020-W2024 blocked），最新无 AE 全量矩阵为 846 total / 840 pass / 0 blocked / 0 failed / 6 skipped，`cmd/aepmigrate matrix` 支持 `-ae-versions` 将 writer target 与 AE2020-AE2025 打开验证 host 维度分离，并带 AE-open case 上限保护；继续前先读 topic index、history 最新段和当前 diff。
- **booyah-glitch-replication** (`work/booyah-glitch-replication/index.md`) — 12-comp from-scratch 复刻已建成并通过 agent AE gate；按 showcase review-gate，只有用户真机确认后才算 complete。①②③⑧ 已 complete；④⑤⑥⑦⑨⑩⑪⑫ 仍待用户 review，下一批优先 ⑩⑪⑫。
- **fx-technique-internalization** (`work/fx-technique-internalization/index.md`) — parked research arc：参考 `.aep` → 拆角色/技法/机制 → 技法库/现象配方。火焰样本已验证；下次恢复需要先选新的参考现象。
- **technique-ontology** (`work/technique-ontology/index.md`) — parked research companion：角色/技法/机制三轴 ontology + schema v2。火焰/闪电/控制器样本已压测；大规模词表稳定性未验证。

## Next

- **versioned AEP migration**：下一刀继续扩 supported convert surface，或把有用的矩阵边界固化成 recurring verification；不要默认跑全量 `-ae-open`，除非显式扩大/关闭 `-max-ae-open-cases`。
- **Booyah 用户真机验收 ⑩⑪⑫**：用户确认后更新 `flightdeck/showcase/booyah-clone/INDEX.md`，再按 review-gate 翻 complete；未确认前不要关闭 Booyah topic。
- **Technique research**：只有用户明确要继续学习/内化参考工程时才恢复；先选样本，不在主线自动推进。

## Open questions

- Booyah documented deltas 是否接受为限制，还是后续开 polish 包：Noise2、Curves arbitrary data、eased Position builder、颜色/撕裂强度细调。
- technique ontology 的大规模 N>>3 词表稳定性何时验证。
- 能力真相源仍是 `go run ./cmd/capindex -q <词>`；不要恢复单独 coverage 表。
