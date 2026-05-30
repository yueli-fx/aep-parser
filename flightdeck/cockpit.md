# Cockpit — aep-parser

**Last updated**: 2026-05-30 by claude（`internal/aep` package 重组 **landed**：`<stage>_<domain>` 7 前缀命名轴（scene_/codec_/parse_/lower_/write_/back_/mutate_）+ 拆 `types_core`(→scene_{project,composition,layer,property}) + 拆 `scene_layer_accessors` + 测试按 feature 拆 + 测试 helper 归并 + AST 边界守卫 `arch_boundary_test.go`（scene_ 禁 import rifx / codec_ 禁 scene 类型；scene→rifx 残留 5 项入白名单，V3 M8 清零）。分支 `refactor/aep-package-reorg`，~16 commits；全程 byte-identical round-trip(115 fixture) + API-set 不变(sorted go-doc) + AE 双版本 ship-gate 24/24 PASS。CLAUDE.md 硬约束#3 已更正（Go 方法同包+Stable API 理由 + 7 前缀轴）。前序里程碑 V2.2.1 子项⑤-⑨（Layr Transform 全通道 + Rect/Stroke/Fill keyframe 持久化）见 git log / `coverage.md`。）
**Active focus**: 无 active 实现线。aep package 重组刚 landed（分支 `refactor/aep-package-reorg`，**待 merge/PR** —— finishing-a-development-branch 处理）。

## Next session

1. **merge/land aep 重组分支**（finishing-a-development-branch：merge `refactor/aep-package-reorg` → main 或开 PR）。
2. **注释纪律清理 pass**（comments.md §6，**单独分支**）：重组期发现 `internal/aep` 源文件 **233 处** comments.md §3 违规（`spec §`、`Phase N`、`Inv-N`、`iter N`、日期戳、`Mirrors`/历史考古等），是**预存债非重组引入**（重组只原样搬运）。跨 ~30 文件，需逐条删/改写/搬 commit-msg。建议另开分支按 comments.md §6 grep 收口。

> **以下为长线 backlog**（shape/transform keyframe+子属性已 drain，子项⑤-⑨）。剩下都是**大 arc 或缺 runtime setter**：

1. **Path keyframe**（逐帧 bezier shap）— V2.3+ 级，大 arc。
2. **Layr Transform 3D 通道** — Orientation / Rotate X/Y / Position_Z；**需先有 3D layer 支持**（runtime 无 3D switch，V2.3）。
3. **Stroke Line Cap/Join/Miter**（enum/scalar）— 中等价值。**子项⑩ groundwork 更正**：stroke body **不含** 这些 slot（template 只有 Color/Opacity/Width/Dashes/Taper/Wave；旧注释说"含 full child set"是错的，默认值被 elide）；且 matchName **不是** `ADBE Vector Stroke Line Cap`（JSX 报 property-not-found）。需先查真实 matchName + 富化 stroke body（设非默认）+ 加 runtime model 字段/enum/setter。
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

- 分支 `refactor/aep-package-reorg` 已完成全部 13 任务 + 终验，**待 merge/PR**（见 Next session #1）。
- 临时文件 `flightdeck/safety-reviews/{ds,claude,gpt}`（外审记录，未跟踪；disposition 已并入 spec §13）—— 用完可删。
