# Cockpit — aep-parser

**Last updated**: 2026-05-30 by claude (V2.2.1 子项⑨ **Fill Opacity 持久化**(static+kf) ship：1D non-spatial bpk-48 原始 %（同 Stroke Opacity）；富化 fill body 模板 7 children（源 `v2_2_shape_fill_full.aep`）；`TestV2_2_FillOpKf_*` 双版本 PASS + FillKf 重跑 PASS。前序：子项⑧ **Stroke Opacity+Width keyframe** ship：均 1D non-spatial bpk-48 原值（Stroke Opacity 存原始 %，**不**÷100，区别于 Layr Opacity）；stroke body 已含 slot→无需富化模板，animated 路径改真 inject；`TestV2_2_StrokeKf_*` 双版本 PASS。前序：子项⑦ **Rect Position+Roundness 持久化** ship：Position=spatial Vec2 motion-path bpk-104（同 Ellipse Position）、Roundness=1D non-spatial bpk-48；helper `lowerShapeVec2/lowerShapeScalar`；富化 rect body 模板 9 children（源 `v2_2_shape_rect_full.aep`）；`TestV2_2_RectSubKf_*` 双版本 PASS + 全 shape ship-gate 重跑双版本 PASS（V2.1 AE2025 一次 OCR-occlusion flake，clean 重试 PASS）。前序：子项⑤+⑥ **Layr Transform 全通道持久化** ship：⑤ Position（combined `ADBE Position` bpk-128 spatial dim-3）；⑥ Anchor/Scale/Rotation/Opacity（Anchor=spatial 同 Position；Scale=3D non-spatial ÷100 Z=1.0；Rotation=1D degrees；Opacity=1D ÷100）；helper `lowerTransformVec2Spatial/Scale/Scalar`；transform 模板扩到 25 children（源 `v2_2_shape_transform_full.aep`）；`TestV2_2_LayrPosKf_*`+`TestV2_2_XfKf_*` 双版本 PASS + 全部既有 shape ship-gate 重跑双版本仍 PASS（模板 shared）；新 API `ShapeLayer.AnchorPoint()`；vet 0 + test ok)
**Active focus**: 无 active 实现线。**Layr Transform 5 通道（Anchor/Position/Scale/Rotation/Opacity）static + keyframe 持久化全收口**。runtime API 2D，磁盘各通道独立编码（详 `coverage.md` 子项⑤⑥）。**关键坑**：① AE 仅在通道 set/animated 时才写其 cdat，默认值 elide → 模板须从「全通道设静态非默认」fixture 提取；② Scale/Opacity 写盘 ÷100 但 parser 读原始盘值（不反归一），ship-gate 用 AE keyValue 验 user 单位。

## Next session

> **shape/transform keyframe+子属性 backlog 已基本 drain**（子项⑤-⑨：Layr Transform 全 5 通道 + Rect Position/Roundness + Stroke Opacity/Width + Fill Opacity）。剩下都是**大 arc 或缺 runtime setter**：

1. **Path keyframe**（逐帧 bezier shap）— V2.3+ 级，大 arc。
2. **Layr Transform 3D 通道** — Orientation / Rotate X/Y / Position_Z；**需先有 3D layer 支持**（runtime 无 3D switch，V2.3）。
3. **Stroke Line Cap/Join/Miter**（enum/scalar）— stroke body **已含 slot**（15 children），套路同子项⑧但**需先加 runtime model 字段+setter**（StrokeNode 暂无）。中等价值、可做。
4. **Gradient W**（SetGradient）— **大 arc**。子项⑩ groundwork findings：① 默认 gradient 被 AE elide（连 G-Fill 默认都不写 prop.map Utf8）→ 须设自定义 stops 强制 emit；② re_gradient.aep 是 gradient **effect**（无 GCst/GCky）不是 shape gradient-fill，**无现成 in-repo fixture**；③ 存储 = `GCst > GCky > Utf8(prop.map XML)`（parse_properties.go:85），写=序列化 XML + 替换 Utf8 + length-variable（镜像 SetExpression）。需：gradient-fill fixture(自定义 stops) + XML 格式 RE + 序列化器 + API（Property.SetGradient 或新 G-Fill shape kind）+ 双版本 gate。
5. **Fill/Stroke BlendMode·CompositeOrder、Rect/Ellipse Direction** — enum，低价值，缺 runtime setter。

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
