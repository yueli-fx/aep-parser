# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (V2.2.1 子项① Ellipse embed bytes 已 ship + 发现并修复 AE 2020 ldta 地基 bug；branch `v3-phase5d-ellipse-embed`，**待合并 main**)
**Active focus**: V2.2.1 ShapeLayer 拓展 — 子项① **Ellipse embed bytes 已落地**（AE 2020+2025 双版本 ship-gate PASS）：embed `v2_2_shape_ellipse_body.bin`（从 AE 2025 tolerance fixture 提取的 730B body，经双版本 gate 证 cross-version 兼容）+ overwrite Size/Position cdat，镜像已 ship 的 Rect 路径。**意外重大收获**：跑 AE 2020 ship-gate 时发现 `buildLdtaBytes` 硬编码 164B ldta，导致 **所有** from-scratch shape 图层（含已"ship"的 Rect+Fill）被 AE 2020 判损坏并跳过——长期未发现因 AE-2020 shape gate 一直 t.Skip()。修复：ldta 大小进 capability matrix（`LdtaSize` 160 AE2020/22 / 164 AE25）。Ellipse 实现 + ldta 修复 + 单测 + 双版本 ship-gate + incident report + docs/coverage 收口全部落地。

## Next session

**先合并** `v3-phase5d-ellipse-embed` 到 main（Ellipse + ldta 修复）。

**V2.2.1 剩余子项**（独立，逐个走 spec→plan→双版本 ship-gate）：
1. **Path embed bytes** — bezier 路径持久化。复用 Ellipse 同流程（AE-native fixture → extract → embed+overwrite → 聚焦 ship-gate）。
2. **Stroke embed bytes** — 最复杂（Dashes/Taper/Wave 嵌套组，见 lower_shape_node.go:239-272）。
3. **Fill Color 编码 RE** — 正交项，cdat scalar 跟 0-1 input 不对齐。
4. **Keyframe 持久化** — 正交项，跨 Rect/Ellipse Size / Fill Color / Layr Position，需 RE lhd3/ldat 注入。

**其它候选**（需用户拍板）：泛型 `DuplicateItem`（低优先，无 scripting API）、`ImportComposition`（需求驱动）。

**V2.2.1 子项① RE 收获**（写进 spec §0 / incident report）：
- AE 2020 ldta 必须 160B（AE 2025 才 164B）；164B 被 AE 2020 判损坏跳层。任何 NewX 写路径必须**真跑** AE 2020+2025 双版本 gate，skip 的 gate = 没验证。
- AE ExtendScript 读 shape 空间属性 `.value`/`.valueAtTime` 抛"除以零"（连 AE-native fixture 都中）→ ship-gate 改 **re-save + 解析器读 cdat** 验证。
- 调 AE 2020 "损坏/跳过"：取 AE-2020-native 参考 diff chunk 尺寸；对话框被会话窗口遮挡时用 **PrintWindow**（遮挡免疫）抓位图。

**自验留痕**（无需人工，已全绿，仅供复核）：
- `go vet ./... && go test -count=1 ./internal/aep/...` 全绿（含 `TestLowerShapeLayer_LdtaSizeByTarget` + `TestLowerEllipseNode_OverwritesSizeAndPosition`）
- V2.2.1 Ellipse 双版本 ship-gate：`AE_SHIP_GATE=1 go test -run TestV2_2_Ellipse_AEShipGate_AE20{20,25}` 均 `--- PASS`（AE 2020 25s / AE 2025 14s）
- 重生：fixture `go run ... AE 2020`-存 `tmp_debug/gen_shape_ellipse_ae2020.jsx`（AE-2020-native 参考，提取 embed）+ `tmp_debug/gen_shape_ellipse_tolerance.jsx`（AE 2025）；提取 `go run ./tmp_debug/extract_shape_bodies`；verify `test_data/verify_v2_2_ellipse.jsx`（re-save + 解析器读 cdat）
- **AE 2020 ldta 修复验证**：AE-2020-native shape ldta=160B、AE-2025=164B（`go run ./tmp_debug/dump_chunks <aep>` 对比）；修前 AE 2020 报"损坏跳过 1"，修后 PASS
- **已知 flake**：AE 2020/2025 cold-start 撞 splash/About 屏 → ship-gate 偶发 exit 2 假阴性；warm retry（先 kill AfterFX 进程清状态 + 跑已知-good fixture 热身）即 exit 0

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

## Hanging tasks

无。
