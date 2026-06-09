# specs/ — INDEX

<!-- AUTO:specs -->
### 待启动（idea）
- [deferred-backlog.md](deferred-backlog.md) — idea — Deferred backlog 细项（从退役 logbook §Deferred 迁入）

### 进行中·完成（active·done）
- [2026-06-10-gradient-stroke-write.md](2026-06-10-gradient-stroke-write.md) — active — GradientStroke R/W（最小·对称已 ship 的 GradientFill）：reader 补 G-Stroke 节点解 gradient（局部降级，对齐 G-Fill），writer 加 GradientStrokeNode 只 model gradient color/alpha stops；fill 的 lower/hydrate 本体抽共享 helper、stroke 复用；双版本 ship-gate（color+alpha，解 stops 比值，对齐 GradientFill）。stroke 几何 + ramp geometry deferred
- [2026-05-22-v3-direction.md](2026-05-22-v3-direction.md) — active — V3 direction：scene-graph IR + capability matrix + serializer split。M1-M8 框架已实现（结构性 mutation Phase 1-5 + 包重组① + M8 真·物理分包② 全落）；可达字段 ~99% ship，剩余前沿 fixture/RE-gated（详 plans/coverage.md）。
<!-- /AUTO -->
