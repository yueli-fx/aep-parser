---
showcase: shape-filters
direction: 形状矢量滤镜家族 — 纯 Go 从零生成 11 件 cdat-based vector filters（+ PolyStar 形状 + Wiggle 调制对比）并 AE 实渲
capabilities: [round-corners, offset-paths, trim-paths, zigzag, pucker-bloat, twist, wiggle-paths, repeater, merge-paths, polystar, wiggle-transform, roughen-points, correlation]
gates: [TestMGTrim_AEShipGate, TestMGRepeater_AEShipGate, TestMGRoundCorners_AEShipGate, TestMGOffset_AEShipGate, TestMGMerge_AEShipGate, TestMGZigZag_AEShipGate, TestMGPuckerBloat_AEShipGate, TestMGTwist_AEShipGate, TestMGWiggle_AEShipGate, TestMGWiggleTransform_AEShipGate, TestMGWiggleMod_AEShipGate, "+PolyStar star gate"]
status: 待review
last_updated: 2026-06-16
regenerate: "go run ./flightdeck/showcase/shape-filters  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# shape-filters — 形状矢量滤镜家族 showcase

## 这个方向测什么

不开 AE，纯 Go（`internal/aep` facade）从零拼一个 **4×4** 网格工程，每格一个形状层 + 一个矢量滤镜/形状。前三行覆盖**全部 11 件常用 shape 矢量滤镜**（cdat-based vein 已闭合）+ PolyStar 形状 + 一个双滤镜叠加；**第 4 行 = Wiggle 调制对比**（家族最后收口的 elided 子流：Roughen Points Corner↔Smooth、Correlation 低↔高）。AE 2020 打开零损坏弹窗、渲染 frame 0 → `shape_filters.png` 供逐格眼验（红线4：看图，不靠值 round-trip）。每件能力本身已过 AE 2020+2025 双版本渲染像素 ship-gate（见 `gates`），本工程验证它们**组合**也成立。

> Temporal/Spatial Phase + Wiggler Correlation 是噪声相位调制、无 categorical 像素（roundtrip-gated，非视觉），故不入本看图档——见 `TestMGWiggleModRT` + `incidents/trim-paths-vector-filter-re.md`。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 纯 facade 调用构建 `shape_filters.aep`（13 层） |
| `render.jsx` | 渲染脚本(tracked) | AE 打开 + `saveFrameToPng(0)` → `shape_filters.png` |
| `shape_filters.aep` | 产出工程 (gitignored) | 1920×1080，13 层（12 示例格 + BG），AE2020 target |
| `shape_filters.png` | 渲染帧 (gitignored) | frame 0，审核用 |

## 布局（4 列 × 3 行，逐格眼验）

| 格 | 图层 | 颜色 | 应看到 | 能力 |
|---|---|---|---|---|
| 行1·列1 | 01_PlainRect | 白 | 正方形（参照，无滤镜） | Rect + Fill |
| 行1·列2 | 02_RoundCorners | 琥珀 | 圆角方块 squircle | Round Corners R=45 |
| 行1·列3 | 03_OffsetPaths | 青 | 外扩的大方块 | Offset Amount=30 |
| 行1·列4 | 04_TrimPaths | 粉 | 不闭合弧（~62% 圆环） | Trim + Stroke End=62 |
| 行2·列1 | 05_ZigZag | 绿 | 锯齿星爆 | ZigZag Size=18/Detail=6 |
| 行2·列2 | 06_PuckerBloat | 紫 | 四叶草 | Pucker&Bloat Amount=90 |
| 行2·列3 | 07_Twist | 琥珀 | 风车螺旋 | Twist Angle=140 |
| 行2·列4 | 08_WigglePaths | 青 | 毛糙噪声边方块 | Wiggle Paths Size=22 |
| 行3·列1 | 09_Repeater | 粉 | 5 个点横排 | Repeater Copies=5 |
| 行3·列2 | 10_MergeSubtract | 绿 | 方块中央挖圆洞 | Merge Subtract（fill 在上） |
| 行3·列3 | 11_PolyStar | 琥珀 | 五角星 | PolyStar 5 points |
| 行3·列4 | 12_TwistPlusWiggle | 紫 | 既扭又毛糙的团块 | Twist + Wiggle 叠加 |
| 行4·列1 | 13_WiggleCorner | 青 | 毛糙方块·**尖角**刺（默认 Corner） | Wiggle Points=Corner |
| 行4·列2 | 14_WiggleSmooth | 青 | 同种子毛糙方块·**圆鼓**波浪（Smooth） | Wiggle Points=Smooth |
| 行4·列3 | 15_WiggleCorrLow | 粉 | 毛糙 jagged 方块（各点独立抖） | Correlation=0 |
| 行4·列4 | 16_WiggleCorrHigh | 粉 | **近乎干净方块**（相干→刚性平移） | Correlation=100 |

> 行4 读法：列1↔列2 同种子 Size34/Detail8，唯 Points 不同 → 尖角 vs 圆鼓；列3↔列4 同种子，唯 Correlation 不同 → jagged vs 几近平整（完全相干使位移退化为整体平移，边几乎不 roughen）。

## 溯源

矢量滤镜家族 RE + 蓝本（11 次复用全绿）：`incidents/trim-paths-vector-filter-re.md`。各滤镜 API：`plans/coverage.md` § MG roadmap S5。
