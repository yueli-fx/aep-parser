# Cockpit — aep-parser

**Last updated**: 2026-05-29 by claude (V2.2.1 全 5 shape kind ship：Ellipse/Path/Stroke embed + Fill/Stroke Color ARGB×255 编码修复 + AE 2020 ldta 地基 + encodeBezier ldat 修复；全部双版本 ship-gate PASS，直接在 main HEAD 3aee822；vet 0 + test ok)
**Active focus**: 无 active 实现线。**V2.2.1 ShapeLayer 拓展 + shape keyframe 持久化（Size/Color）全部收口**：5 个 shape kind 全 embed；3 个潜伏 bug（AE 2020 ldta 地基、encodeBezier ldat 顶点编码、Fill/Stroke ARGB×255 颜色编码）修复；keyframe 持久化 Rect/Ellipse **Size** + Fill/Stroke **Color** 落地（`injectAnimatedStream`：cdat↔LIST(list) + tdb4 标志 patch）。全部双版本 ship-gate PASS。HEAD 0751962。（AE 2020+2025 双版本 ship-gate PASS）：embed `v2_2_shape_ellipse_body.bin`（从 AE 2025 tolerance fixture 提取的 730B body，经双版本 gate 证 cross-version 兼容）+ overwrite Size/Position cdat，镜像已 ship 的 Rect 路径。**意外重大收获**：跑 AE 2020 ship-gate 时发现 `buildLdtaBytes` 硬编码 164B ldta，导致 **所有** from-scratch shape 图层（含已"ship"的 Rect+Fill）被 AE 2020 判损坏并跳过——长期未发现因 AE-2020 shape gate 一直 t.Skip()。修复：ldta 大小进 capability matrix（`LdtaSize` 160 AE2020/22 / 164 AE25）。Ellipse 实现 + ldta 修复 + 单测 + 双版本 ship-gate + incident report + docs/coverage 收口全部落地。

## Next session

**先合并** `v3-phase5d-ellipse-embed` 到 main（Ellipse + ldta 修复）。

**V2.2.1 剩余 keyframe / 子属性**（自主推进，勿停下问）：
1. **Layr/shape Position keyframe**（spatial 真运动路径）— RE 已起：bpk=128（dim2，value@0x38 后 9 f64，≠ color 的 3·dim），含 spatial 切线，走 transform-group 路径（`lowerLayerTransform`）+ 分离维 Position_0/_1。dump 工具 `tmp_debug/dump_kf`、fixture `test_data/v2_2_shape_kf_re.aep` 已备。
2. **Ellipse Position keyframe** — 疑非 spatial Vec2（同 Size），可 `injectAnimatedVec2` 直接接，但需先建 animated Ellipse fixture 验证。
3. **Path keyframe**（逐帧 bezier shap）— V2.3+ 级。
4. **各 shape 次要子属性** — Fill/Stroke Opacity·BlendMode·CompositeOrder、Stroke Line Cap/Join/Miter/Dashes/Taper/Wave、Rect/Ellipse Direction、Layr Transform。多数 runtime-only；逐个 RE。

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
