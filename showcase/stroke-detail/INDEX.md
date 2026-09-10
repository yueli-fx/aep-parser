---
showcase: stroke-detail
direction: 描边细节 setter — Dashes / Line Join / Miter Limit / Wave，纯 Go 从零生成后 AE 实渲眼验（全用闭合形状，规避 open-path 渲染塌缩假绿）
capabilities: [stroke-dashes, stroke-line-join, stroke-miter-limit, stroke-wave]
gates: [TestV2_2_Stroke_AEShipGate, TestV2_2_StrokeTaperWave_AEShipGate, shape_dashes_shipgate_test]
status: complete
last_updated: 2026-06-14
regenerate: "go run ./showcase/stroke-detail  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# stroke-detail — 描边细节 setter showcase

## 这个方向测什么

`StrokeNode` 的细节 setter，纯 Go 从零建层后 AE 实渲：`Dashes().SetDash/SetGap` · `SetLineJoin` · `SetMiterLimit` · `Wave().SetAmount/SetWavelength`。每个 zone 一层、一个**闭合**形状 + 一条**粗描边**（18–26px），让几何一眼可读。

## ⚠ 真实边界（诚实标注 · 关键 — 红线4d 两个假绿）

**全用闭合形状（star / rect），不用 open path。** 第一版把每条线/V 建成 open path（`AddPath`+`SetClosed(false)`），**每个 open-path zone 都渲成 ~26px 小блоб**：open shape path 的几何在 Go round-trip 绿、但在 AE **塌缩**，粗描边只画出近零尺寸轮廓。这是确认的假绿边界——三个 stroke ship-gate（`shape_stroke_shipgate_test` / `shape_taperwave_shipgate_test` / `shape_dashes_shipgate_test`）**全建闭合 rect、只验值 round-trip，从不验渲染像素、从不用 open path**。故唯一**渲染**被证实的 stroke 几何 = 闭合形状。Join/Miter 改用**闭合五角星**（尖角正是 join 差异出现处）。

**两项故意省略**（无诚实的 from-scratch 视觉）：

- **Line Cap**（Butt/Round/Projecting）：cap 只出现在 **open path 的端点**，而 open shape path 塌缩。除非 open-path 描边渲染本身有像素级 ship-gate，否则 Line Cap 无诚实可视。
- **Taper**：闭合环无起点/终点，AE 渲成**等宽轮廓**——值 round-trip（`taperEndW` 读回对）但**无视觉**（已实渲确认：taper 星与普通星无差）。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| gen.go | 生成器(Go, tracked) | 构建 stroke_detail.aep（8 zone + BG，全闭合形状） |
| render.jsx | 渲染脚本(tracked) | `saveFrameToPng(0)` + 每层 dump stroke 细节值 → png/.done |
| stroke_detail.aep | 产出工程 (gitignored) | 1920×1080，AE2020 target，9 层 |
| stroke_detail.png | 渲染帧 (gitignored) | frame 0 |

## 布局（4 列 × 2 行）

| 位置 | zone | 形状 | 设定 | 期望渲染 |
|---|---|---|---|---|
| (0,0) | **Dashes** | 闭合 rect 300×220 | dash 36 / gap 22 @ W18 | cyan 虚线框（行进蚂蚁） |
| (1,0) | **Join Miter** | 闭合五角星 | LineJoin=Miter @ W24 | white 星 **尖角** |
| (2,0) | **Join Round** | 闭合五角星 | LineJoin=Round @ W24 | green 星 **圆角** |
| (3,0) | **Join Bevel** | 闭合五角星 | LineJoin=Bevel @ W24 | violet 星 **切角（平拐）** |
| (0,1) | **Miter High** | 极尖五角星(inner34) | Miter join, limit 28 @ W20 | amber 星 **长尖刺保留** |
| (1,1) | **Miter Low** | 同上 | Miter join, limit 1 @ W20 | cyan 星 **尖刺被削成切角** |
| (2,1) | **Wave Fine** | 闭合五角星 | Wave amt20 / wavelen22 | green 星轮廓 **密小波纹** |
| (3,1) | **Wave Bold** | 闭合五角星 | Wave amt48 / wavelen60 | pink 星轮廓 **疏大波纹** |

> 审核要点：① cyan 虚线框成立；② 第一行三星尖角/圆角/切角**三态可分**；③ 第二行 amber 长刺 vs cyan 削平（miter limit 高低差）；④ green/pink 两星轮廓**起伏**、明显不同于第一行的光滑星（Wave 生效），fine 波密、bold 波疏。Line Cap / Taper 见上「真实边界」诚实暂缺。

## 溯源

Stroke 细节 RE：`incidents/stroke-line-cap-join-miter-re.md`（含 Taper/Wave 增补、Miter Limit 仅 Join=Miter 时保留的经验行为）；Dashes 见 `shape_dashes_*`。open-path 塌缩假绿见本对话发现，待 open-path 描边像素 ship-gate 后再补 Line Cap。
