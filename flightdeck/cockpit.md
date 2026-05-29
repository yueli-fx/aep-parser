# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (Phase 5C InsertLayer impl + JSX shipped Alpha; awaiting AE ship-gate)
**Active focus**: V3 Phase 5C InsertLayer Alpha — Go-side complete (refuse R1-R11 + happy-path × 3 splice positions + round-trip + concurrent-mutate). Waiting on user JSX run → AE ship-gate (3 modes × 2 versions = 6 PASS) for Stable promotion.

## Next session

1. **User runs `re_insert_layer.jsx`** under AE 2020 + AE 2025 — 12 invocations total:
   `for mode in basic footage precomp; for state in before after: $env:RE_INSERT_MODE=$mode; $env:RE_INSERT_STATE=$state; afterfx.exe -r test_data/re_insert_layer.jsx`
   Produces `test_data/re_insert_layer_{basic,footage,precomp}_{before,after}.aep` (6 files).
2. **Re-run Go tests** to lift fixture skips: `go test -count=1 ./internal/aep/ -run TestInsertLayer -v` — expect 11 refuse + 6 happy/structural PASS.
3. **AE ship-gate** — `scripts/ae_run.ps1` opens each `ge_insert_layer_<mode>.aep` (Go-emitted post-InsertLayer) in both AE versions and byte-diffs against the `_after` baseline. 6/6 PASS → promote to Stable.
4. **Promote godoc tag** Alpha → Stable in `internal/aep/insert_layer.go` + add coverage row.

**Phase 5 后续候选**（5C 完后回到三选一）：
- **`Project.DuplicateItem(item Item, name string)`** — comp / footage / folder 通用
- **V2.2.1 ShapeLayer 拓展**
- Phase 5C.1 cross-Project InsertLayer

## Hanging tasks

无。
