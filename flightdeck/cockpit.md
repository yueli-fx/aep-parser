# Cockpit — aep-parser

**Last updated**: 2026-06-15 by claude（**优先级3 形状剩余推进**：**Trim Type**（Individually/Simultaneously）经 synthesis-insert 第 2 次落地——enum leaf〔4-child tdbs〕同 scalar-with-range 走通用 `spliceShapeLeafBeforeGroupEnd`。`TestMGTrimType_AEShipGate` 双版本 PASS〔Individually=路径按序 trim→左满右空，与「Simultaneously=各半弧」相反，render 实锤纠直觉；left-ring 4/4 + right-empty 4/4 + resave Type=2，png 20470b 两版一致〕。前情同日：**Offset Copies** 经 synthesis-insert 落地——把 `SetMaterialOption`（material leaves）那套**首次推广到矢量滤镜 body**：默认 Amount-only body 不变，仅当 `SetCopies≠1` 时 splice `ADBE Vector Offset Copies` leaf〔`spliceShapeLeafBeforeGroupEnd` + 覆写 cdat〕。`TestMGOffsetCopies_AEShipGate` 双版本 PASS〔even-odd 同心环签名：外白环 d≈300 + gap 暗 d≈260 + 白心，单副本不可能；png 10777b 两版一致〕。蓝本可推 Trim Type / Offset 其余子流。详 `trim-paths-vector-filter-re.md` § Offset Copies。前情：优先级2（3D）全部收官。逐 commit 见 git log。）

**Active focus**: **剩余能力 roadmap**（`specs/2026-06-14-remaining-capability-roadmap.md`）。优先级顺序（用户 2026-06-14 定）：动画关键帧（基本收口）> **3D（✅ 全收官 2026-06-15）** > **形状图层剩余〔← 现在这里〕** > mask > 表达式 > 文字。模板/真实 .aep 起点（非 from-scratch，规避 silent-drop）。机制库：parse-the-clone + synthesis-insert（camera iris / light color / material leaves）。每渲染/可见类双版本 ship-gate（红线4）；ship-gate 自助（`scripts/ae_run.ps1`）。基本图形搁置。

## 进行中

<!-- AUTO:inprogress -->
- [2026-06-12-from-scratch-mg-roadmap.md](specs/2026-06-12-from-scratch-mg-roadmap.md) — 终极目标：不开 AE、纯 Go 从零生成完整 MG 动画工程（AI 直接产出 .aep）。按交付准则逐 slice 确权（每 slice 渲染像素级双版本 gate）：S1 ease 关键帧+规模 gate → S2 表达式激活 RE → S3 Trim Paths → S4 precomp 嵌套 → S5 Repeater/gradient 方向/圆角
- [2026-06-14-remaining-capability-roadmap.md](specs/2026-06-14-remaining-capability-roadmap.md) — 模板/真实 .aep 起点的未做能力清单，按优先级排序：动画关键帧 > 3D 图层 > 形状图层剩余 > mask > 表达式 > 文字图层。非 from-scratch（已有基础模板规避 silent-drop）。基本图形搁置。
- [coverage-detail.md](plans/coverage-detail.md) — 字段覆盖矩阵（详细参考 + 暂搁/不可达/negative findings）
- [coverage.md](plans/coverage.md) — 字段覆盖概览（精简入口）
<!-- /AUTO -->

## 下一步

**主线 = `specs/2026-06-14-remaining-capability-roadmap.md`**（模板起点，非 from-scratch）。优先级顺序（用户 2026-06-14 定）：**动画关键帧 > 3D 图层 > 形状剩余 > mask > 表达式 > 文字图层**。

**优先级1 动画关键帧 ✅ 基本收口**（lhd3 容量分页 / temporal ease / animated Trim / animated gradient 色标 全双版本，2026-06-14~15）。遗留：path 几何 lhd3 >4 顶点分页 + gradient STROKE 色标动画（均需求驱动）。

**优先级2 3D 图层 ✅ 全部收官（2026-06-15）**：enable / 相机推拉 / Z 视差 / RotateY 透视 / 相机 DoF / **光照** / **阴影** 七项全渲染 gate 双版本 PASS；showcase `3d-camera` 用户真机验收 complete。整个 3D transform group 从零零新 serializer 代码（transform 模板本就是 6-axis 3D schema）；光照零新代码；阴影靠新 `aep.SetMaterialOption` synthesis-insert（material leaves）。RotateX/Orientation/RotateZ 同路径按需补 gate（低优先）。详 `incidents/layer-3d-enable-bit-materializes.md`。

**➡ 优先级3 形状图层剩余〔进行中〕**（shape 矢量滤镜主体已收齐；剩 elided 子流 + 未单独 gate 的模式，多为 synthesis-insert / enum 补值小活）：
- ~~**Offset Copies**~~ ✅ **2026-06-15**（synthesis-insert 首发矢量滤镜；scalar-with-range leaf）。剩 Offset Line Join / Miter / Copy Offset 同 splice 路径，按需。
- ~~**Trim Type**（Simultaneously/Individually）~~ ✅ **2026-06-15**（synthesis-insert 第 2 次，enum leaf；Individually=左满右空 render 签名）。
- **Merge Add/Intersect/Exclude 模式**（已能写 enum，未单独 gate）← 推荐下一项（纯 gate 工作，无新代码；3 模式渲染验证）。
- **shape 次要子属性**（Fill/Stroke Opacity·BlendMode·CompositeOrder、Shape Direction；部分 runtime-only 逐个甄别）。
- **Gradient stroke 嵌套组 Dashes/Taper/Wave**（实心描边的三组已 ship 子项⑫⑬；gradient stroke 缺）。

**蓝本（synthesis-insert 推广到矢量滤镜，2026-06-15 验透 2 类 leaf）**：clone elided leaf 模板 → `spliceShapeLeafBeforeGroupEnd`（GroupEnd 前插 (tdmn,tdbs) pair）→ 覆写 cdat，仅当值≠默认。已验 scalar-with-range（Offset Copies 6-child）+ enum（Trim Type 4-child）两类。不污染默认 body、不需 hydration（filter 靠 opaque chunk 穿越 Reopen）。剩 ZigZag Points / Twist Center / Repeater Order / Offset Line Join·Miter·Copy Offset 同路径按需。详 `trim-paths-vector-filter-re.md` § Offset Copies / Trim Type。

完整清单（每层细项 + 优先级4-6 mask/表达式/文字 + 不可达附录）见 roadmap spec。

**搁置（用户决定）**：Essential Graphics 进阶 + EG 面板崩溃未修 RE。
**独立线（按需）**：Render Queue Set* slice-5~8（Alpha）。

## Backlog

- **Layr Transform 3D 通道**（Orientation / Rotate X·Y / Position_Z）— 需先有 3D layer 支持。
- 泛型 `DuplicateItem` · `ImportComposition` · **Property synthesis**（暂搁大 feature，`incidents/transform-group-default-omission.md`）。
- fixture/RE-gated + deferred R-only（DisplayColorSpace / ValueText / environmentLayer / ligature 等）详 `specs/deferred-backlog.md` + `plans/coverage.md` § 暂搁/不可达。

## Hanging tasks

无。
