# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (Phase 5C InsertLayer design committed, awaiting user review → plan)
**Active focus**: V3 Phase 5C InsertLayer — cross-comp deep-clone (same-Project sibling)。Design spec `specs/2026-05-29-v3-phase5c-insertlayer-design.md` committed (`d5ba860`)；下一步 writing-plans → impl → ship-gate (3 modes × 2 versions = 6 PASS)。

## Next session

1. **User 复核 Phase 5C 设计 spec** (`specs/2026-05-29-v3-phase5c-insertlayer-design.md`)；如有修改先 patch 再起 plan
2. **Approved → writing-plans skill** → `flight-plans/2026-05-29-v3-phase5c-insertlayer-plan.md`
3. **Impl + Go unit tests** (refuse-cases + 3 happy modes + round-trip + concurrent-mutate safety)
4. **AE ship-gate** — user 跑 `re_insert_layer.jsx` 产 3 mode RE fixtures → `tmp_debug/ge_insert_layer` produce ge files → `scripts/ae_run.ps1` 双版本验 6/6 PASS → godoc alpha → Stable

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
