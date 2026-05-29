# Cockpit — aep-parser

**Last updated**: 2026-05-30 by claude (V2.2.1 子项⑤ **Layr Transform Position keyframe 持久化** ship：combined `ADBE Position` bpk-128 spatial dim-3，Z=0；新 helper `injectAnimatedLayerPosition` + 新 transform-group 模板（含 combined Position cdat，从 `v2_2_shape_transform_pos.aep` 提取）；`TestV2_2_LayrPosKf_*` 双版本 PASS + 全部既有 shape ship-gate 重跑双版本仍 PASS（模板 shared）；vet 0 + test ok)
**Active focus**: 无 active 实现线。**Layr Position keyframe（Path B / combined）收口**：runtime API 仍 2D，磁盘 dim-3 spatial（value@0x38 X/Y/Z，motion-path 标志@0x08），`encodeKeyframes` spatial 分支已泛化 dim3。**关键 RE**：AE 仅在 Position set/animated 时才写 combined `ADBE Position` cdat；默认/未触碰 → 只有分离维 Position_0/_1（旧模板即此态，无 combined slot），故换新模板。详 `coverage.md` 子项⑤。

## Next session

**V2.2.1 剩余 keyframe / 子属性**（自主推进，勿停下问）：
1. **Path keyframe**（逐帧 bezier shap）— V2.3+ 级。
2. **Layr Transform 其余 keyframe** — Anchor / Scale / Rotation / Opacity（镜像 Position：combined stream，多数非 spatial）。Position 套路（`injectAnimatedLayerPosition` + 新模板已含这些 stream 的 cdat slot）可直接参照。
3. **各 shape 次要子属性** — Fill/Stroke Opacity·BlendMode·CompositeOrder、Stroke Line Cap/Join/Miter/Dashes/Taper/Wave、Rect/Ellipse Direction。多数 runtime-only；逐个 RE。

**其它候选**：泛型 `DuplicateItem`（低优先，无 scripting API）、`ImportComposition`（需求驱动）。

**V2.2.1 子项① RE 收获**（写进 spec §0 / incident report）：
- AE 2020 ldta 必须 160B（AE 2025 才 164B）；164B 被 AE 2020 判损坏跳层。任何 NewX 写路径必须**真跑** AE 2020+2025 双版本 gate，skip 的 gate = 没验证。
- AE ExtendScript 读 shape 空间属性 `.value`/`.valueAtTime` 抛"除以零"（连 AE-native fixture 都中）→ ship-gate 改 **re-save + 解析器读 cdat** 验证。
- 调 AE 2020 "损坏/跳过"：取 AE-2020-native 参考 diff chunk 尺寸；对话框被会话窗口遮挡时用 **PrintWindow**（遮挡免疫）抓位图。

**自验留痕**（无需人工，已全绿，仅供复核）：
- `go vet ./... && go test -count=1 ./internal/aep/...` 全绿（含 `TestLowerShapeLayer_LdtaSizeByTarget` + `TestLowerEllipseNode_OverwritesSizeAndPosition`）
- V2.2.1 Ellipse 双版本 ship-gate：`AE_SHIP_GATE=1 go test -run TestV2_2_Ellipse_AEShipGate_AE20{20,25}` 均 `--- PASS`（AE 2020 25s / AE 2025 14s）
- V2.2.1 Path 双版本 ship-gate：`AE_SHIP_GATE=1 go test -run TestV2_2_Path_AEShipGate_AE20{20,25}` 均 `--- PASS`（AE 2025 13s / AE 2020 20s；distinct 三角形 re-save 反归一化验证 anchor）。embed body 与 AE-native 方块字节一致（`go run ./tmp_debug/diff_path_geom`）
- V2.2.1 Stroke 双版本 ship-gate：`go test -run TestV2_2_Stroke_AEShipGate_AE20{20,25}` 均 PASS（AE 2025 15s / AE 2020 18s；distinct Color[1,0,0,1]/Width4/Opacity60 re-save 解 ARGB×255+f64）。Ellipse gate 现含 Fill 颜色 ARGB×255 校验
- shape 颜色编码 RE：`go run ./tmp_debug/dump_stroke_cdat`（fixture Color[0,0,1,1]→磁盘[255,0,0,255] 即 [A,R,G,B]×255）
- **AE flake 应对（本会话教训）**：from-scratch shape 喂 AE 2020 会**崩溃**（非仅 silent-drop）→ 反复 force-kill 触发"崩溃修复选项"对话框，OCR 遮挡下 ae_run 无法消除 → 先跑 `tmp_debug/clear_ae_crashstate.ps1`（Win32 前台+Enter 清崩溃态）再跑 gate；冷启 splash >15s grace 偶发 exit 2，warm retry
- 重生：fixture `go run ... AE 2020`-存 `tmp_debug/gen_shape_ellipse_ae2020.jsx`（AE-2020-native 参考，提取 embed）+ `tmp_debug/gen_shape_ellipse_tolerance.jsx`（AE 2025）；提取 `go run ./tmp_debug/extract_shape_bodies`；verify `test_data/verify_v2_2_ellipse.jsx`（re-save + 解析器读 cdat）
- **AE 2020 ldta 修复验证**：AE-2020-native shape ldta=160B、AE-2025=164B（`go run ./tmp_debug/dump_chunks <aep>` 对比）；修前 AE 2020 报"损坏跳过 1"，修后 PASS
- **已知 flake**：AE 2020/2025 cold-start 撞 splash/About 屏 → ship-gate 偶发 exit 2 假阴性；warm retry（先 kill AfterFX 进程清状态 + 跑已知-good fixture 热身）即 exit 0

**并行 R-only 仍 deferred**（不阻塞 V3）：
- **Gradient W**: XML 重序列化 / SetGradient / per-keyframe gradients — 需 fixture
- **DisplayColorSpace R**: separate chunk 位置未 RE
- **ValueText**: per-type formatter — P3

## Hanging tasks

无。
