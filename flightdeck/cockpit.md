# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (Phase 5C.1 cross-Project InsertLayer 已合并 main，HEAD 156ee6a；vet 0 + `go test ./internal/aep/...` ok 复核通过)
**Active focus**: 无 active 实现线。Phase 5C.1 cross-Project InsertLayer 已 ship 并合并到 main（branch `v3-phase5c1-cross-project-insertlayer` 已清理）：抬 R7 折入现有 `InsertLayer`，跨 Project 时导入源层 item 闭包（footage+precomp，`locateItemBlockByID` 递归进 folder Sfdr）到 dest 根级取新 ID、footage 按 Path 去重、remap SourceID/AltSourceID；共享 `spliceLayerClone` 核（sourceRemap=identity 同 Project / itemIDMap 跨 Project）。Go 实现 + 单测 + 12/12 ship-gate + godoc Stable + coverage 收口全部落地。Phase 5 三选一已全落（#1 DuplicateComposition、#2 cross-Project InsertLayer、#3 解锁）。

## Next session

**剩余候选**（需用户拍板再起 brainstorming → design → plan）：
1. **V2.2.1 ShapeLayer 拓展** — Ellipse/Path/Stroke embed bytes / Fill Color 编码 RE / keyframe 持久化。**最重**，核心是字节 RE，不确定性高；需先 AE 建 fixture。
2. **泛型 `Project.DuplicateItem(Item, name)`** — 给 Composition/Footage/Folder 加 `Item` interface 当伞；footage/folder 复制无 scripting API（UI 自动化 RE 或跳过）。低优先。
3. **`Project.ImportComposition(from *Project, src)`** — public comp 级跨 Project 导入，复用 5C.1 闭包内部件（薄 wrapper + 自己的 ship-gate）。需求出现再做。

**5C.1 RE 收获**（写进 spec §0）：AEP 把 item 嵌在 folder 的 Sfdr 子容器里（solids 默认在 "Solids" folder）；闭包定位必须递归，否则 solid-source 跨 Project 插入会误判 dangling。跨 Project 在 AE 无 scripting 对应 → ship-gate 改 assert-based（无 byte-diff 基线）。

**自验留痕**（无需人工，已全绿，仅供复核）：
- `go vet ./... && go test -count=1 ./internal/aep/...` 全绿
- 5C.1 cross-Project ship-gate：`test_data/verify_ge_cross_project_insert_ae20{20,25}_{footage,precomp,dedup}.done` 末行 `PASS`（6/6）
- 同 Project InsertLayer 回归：`verify_ge_insert_layer_ae20{20,25}_{basic,footage,precomp}.done` 全 `PASS`（6/6，确认 spliceLayerClone 抽取无回归）
- 重生：`go run ./tmp_debug/ge_cross_project_insert` + `re_cross_project_insert.jsx`（env `RE_XPROJ_MODE`）；verify 用 `verify_ge_cross_project_insert.jsx`（env `GE_XPROJ_MODE` + `GE_XPROJ_TAG`）
- **已知 flake**：AE 2020/2025 cold-start 撞 splash/About 屏 → ship-gate 偶发 exit 2 假阴性；warm retry（先跑一个已知-good fixture 热身）即 exit 0。本次 AE 2025 footage mode 撞 3 次才热（非数据问题）

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

## Hanging tasks

无。
