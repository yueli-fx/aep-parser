# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (Phase 5C InsertLayer plan committed, ready for impl execution)
**Active focus**: V3 Phase 5C InsertLayer — design + plan 都已 committed。下一步选执行路径 → Phase A 起 R1-R11 refuse (Go-side, 无 fixture 依赖) → Phase B happy-path impl → Phase C 用户 JSX → Phase D ship-gate 6/6 PASS → Stable。

## Next session

1. **选执行路径**: subagent-driven (per-task A.1→D.1) vs inline (`executing-plans` w/ checkpoints after B.2 + D.1)
2. **Phase A R1-R11 refuse** — A.1 (R1 nil) / A.2 (R2-R7 + fixture helper) / A.3 (R8-R11 corruption defense)；Go-side only，无 fixture 依赖直跑
3. **Phase B happy-path impl** — B.2 完整 InsertLayer (clone block + 4 ldta deltas @0x00/0x6B/0x84/0xA0 + 3-branch splice + parseLayer + warnings-as-failure rollback)；B.3 splice positions；B.4 round-trip + concurrent-mutate
4. **Phase C 用户 JSX** — `re_insert_layer.jsx` × 3 modes × 2 states under AE 2020 + 2025 → 6 `.aep` fixtures populate `test_data/`
5. **Phase D ship-gate + 收尾** — `scripts/ae_run.ps1` 双版本 6/6 PASS → coverage doc + godoc Alpha→Stable

**Phase 5 后续候选**（5C 完后回到三选一）：
- **`Project.DuplicateItem(item Item, name string)`** — comp / footage / folder 通用；扩展 V2.1 NewComposition + V3 DuplicateLayer 到 item 级
- **V2.2.1 ShapeLayer 拓展**（Ellipse/Path/Stroke embed bytes / Fill Color 编码 / keyframe 持久化）
- Phase 5C.1 cross-Project InsertLayer（couples with DuplicateItem）

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

## Hanging tasks

无。
