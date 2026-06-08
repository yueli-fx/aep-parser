# plans/ — INDEX

<!-- AUTO:plans -->
- [2026-06-07-v3-m8-physical-split-plan.md](2026-06-07-v3-m8-physical-split-plan.md) — active — V3 M8 方案② 物理分包实现计划（A 先行 + B′）：P0 基线+inventory → P1 抽 internal/codec → P2 单包内 back-ref 接口化（concrete→XWriter） → P3 git mv 物理分包 → P4 下游+收口+双版本 ship-gate
- [coverage-detail.md](coverage-detail.md) — active — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](coverage.md) — active — 字段覆盖概览（精简入口）
- [m8-setter-inventory.md](m8-setter-inventory.md) — active — V3 M8 Task 0.3 产出 — 全量 Set* 分类表（A 类 back-ref setter / B 类 pure-graph setter）+ 每 backref 结构的 XWriter 接口方法清单。P2（back-ref 接口倒置）的逐类输入。
<!-- /AUTO -->

<!-- 注：这两份是长期 reference 矩阵（非一次性实现 plan），随字段进展滚动更新。 -->
