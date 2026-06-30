---
showcase: animated-path
direction: 动画形状路径 — 一条闭合 Path 两线性关键帧，几何本身 morph（横条→竖条），渲中间帧 t=2s 为正方形（顶点逐个插值的单帧证据）+ .done 三时刻 extent readback
capabilities: [path-keyframes, animated-shape-path, bezier-path-interpolation, multi-frame-oms]
gates: [TestV2_2_PathKf_AEShipGate_AE2020, TestV2_2_PathKf_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-14
regenerate: "go run ./flightdeck/showcase/animated-path  +  scripts/ae-worker/ae_run.ps1 render.jsx (renders t=2s)"
---

# animated-path — 动画形状路径 showcase

## 这个方向测什么

关键帧作用在**路径几何本身**（不是 transform）：一条**闭合 Path** 两个线性关键帧，把**宽横条**（t=0，440×120）morph 成**高竖条**（t=4，120×440）。两端共 4 个顶点 → AE 逐顶点插值；**渲染正中 t=2s**，每个顶点恰在中点，路径渲成**正方形**（280×280）——这就是「路径关键帧插值的是几何」的单帧证据。**在 AE 里拖时间轴**可看完整 横条→正方→竖条 morph。`.done` 把 t=0/2/4 三时刻路径 extent 读回数值佐证。

> 边界：用**闭合**路径 + 实心 fill（轮廓无歧义，规避 open-path 描边塌缩的假绿，详 `stroke-detail`）；仅 **2 关键帧**，远低于 lhd3 >4-kf 容量分页坑（`lhd3-keyframe-capacity-pages.md` / `encodePathTimeTable` 同病——未来加多帧路径前必修）。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `animated_path.aep`（2 层 = Morph 路径 + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(2.0)` + t=0/2/4 路径 extent readback → png/.done |
| `animated_path.aep` | 产出工程 (gitignored) | 1920×1080，30fps×4s，AE2020 target |
| `animated_path.png` | 渲染帧 (gitignored) | **t=2s（中间帧）= 正方形** |

## 布局 / morph 时间线

| 时刻 | 路径形状 | extent（readback） |
|---|---|---|
| t=0 | 宽横条 | 440×120 |
| **t=2（渲染帧）** | **正方形** | **280×280** |
| t=4 | 高竖条 | 120×440 |

> 审核要点：① png 是**青色正方形**（横条与竖条的几何中点）；② AE 里拖时间轴 0→4s 看到 横条平滑变竖条；③ `.done` extent 三值 = `440x120 / 280x280 / 120x440`，证顶点逐个线性插值。若中帧仍是横条或塌成细条 = 路径插值未生效。

## 溯源

Phase 2 animated path lower（om-s = N shaps + tdbs 时间表）：`incidents/path-keyframe-write-re.md`；ship-gate `shape_pathkf_shipgate_test.go`（3 闭合关键帧，验 AE 接受 + om-s 结构存活，**未验渲染像素** → 本档补像素眼验）；容量坑 `lhd3-keyframe-capacity-pages.md`。
