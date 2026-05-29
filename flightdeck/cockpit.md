# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (Phase 5C InsertLayer + Phase 5D DuplicateComposition 双 Stable，ship-gate 全 PASS)
**Active focus**: 无 active 实现线。一夜连做两个里程碑：Phase 5C InsertLayer（6/6 ship-gate）+ Phase 5D DuplicateComposition（2/2 ship-gate），均 Go-side + ship-gate + godoc Stable + coverage 收口。Phase 5 三选一已落 2 个（#1 DuplicateItem→DuplicateComposition、#3 现已解锁）。

## Next session

**剩余候选**（需用户拍板再起 brainstorming → design → plan）：
1. **V2.2.1 ShapeLayer 拓展** — Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化。**最重**，核心是字节 RE，不确定性高；需先 AE 建 fixture。
2. **Phase 5C.1 cross-Project InsertLayer** — **现已解锁**（DuplicateComposition 是 building block）。设计见 5C design §8 lean(a)：跨 Project = 先 DuplicateComposition 把源 comp 拷进 dest Project（footage item 复制无 scripting API，需另 RE 或限 comp-source），再 same-Project InsertLayer。比 ShapeLayer 轻，但 footage-item 跨 Project 复制仍有 RE 缺口。
3. **泛型 `Project.DuplicateItem(Item, name)`** — 给 Composition/Footage/Folder 加 `Item` interface 当伞；footage/folder 复制无 scripting API（UI 自动化 RE 或跳过）。低优先。

**自验留痕**（无需人工，已全绿，仅供复核）：
- `go vet ./... && go test -count=1 ./internal/aep/...` 全绿
- InsertLayer ship-gate：`test_data/verify_ge_insert_layer_ae20{20,25}_{basic,footage,precomp}.done` 末行 `PASS`
- DuplicateComposition ship-gate：`test_data/verify_ge_duplicate_composition_ae20{20,25}.done` 末行 `PASS`
- 所有 test_data baseline/ge 文件 gitignored；重生：`go run ./tmp_debug/ge_insert_layer`、`go run ./tmp_debug/ge_duplicate_composition` + 对应 `re_*.jsx`（见 logbook 2026-05-29 两条）
- **已知 flake**：AE 2020 cold-start 撞 splash/About 屏 → ship-gate 偶发 exit 2 假阴性；warm retry 即 exit 0（非数据问题）

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

## Hanging tasks

无。
