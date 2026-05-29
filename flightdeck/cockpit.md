# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (Phase 5C InsertLayer **Stable** — 6/6 AE 2020+2025 ship-gate PASS)
**Active focus**: 无 active 实现线。Phase 5C InsertLayer 完整收口（Go-side + 6/6 ship-gate + godoc Stable + coverage）。下一步是 Phase 5 三选一，需用户定方向。

## Next session

**Phase 5 三选一**（需用户拍板再起 brainstorming → design → plan）：
1. **`Project.DuplicateItem(item Item, name string)`** — comp / footage / folder 通用 item 复制（project-level，区别于 layer-level DuplicateLayer）
2. **V2.2.1 ShapeLayer 拓展** — Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化（需 AE create fixture）
3. **Phase 5C.1 cross-Project InsertLayer** — 现 InsertLayer refuse 的 cross-Project 分支解封（src/dest 不同 Project，需复制 source item 进 dest Project）

**自验留痕**（无需人工，已全绿，仅供复核）：
- `go vet ./... && go test -count=1 ./internal/aep/...` 全绿（最近一次 48.9s ok）
- InsertLayer ship-gate done 文件：`test_data/verify_ge_insert_layer_ae20{20,25}_{basic,footage,precomp}.done` 末行均 `PASS`（gitignored，本地留存）
- baseline/ge 文件 gitignored；要重生：`go run ./tmp_debug/ge_insert_layer` + `re_insert_layer.jsx`（见 logbook 2026-05-29 条）

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

## Hanging tasks

无。
