---
showcase: keyframe-channels
direction: 四个 transform 通道关键帧 — Position/Scale/Rotation/Opacity 各两线性关键帧，渲中间帧 t=2s 让每通道落在插值（非端点）值，像素 + .done readback 双证
capabilities: [position-keyframes, scale-keyframes, rotation-keyframes, opacity-keyframes, linear-interpolation]
gates: [TestMGEase_AEShipGate_AE2020, TestMGEase_AEShipGate_AE2025]
status: complete
last_updated: 2026-06-14
regenerate: "go run ./showcase/keyframe-channels  +  scripts/ae-worker/ae_run.ps1 render.jsx (renders t=2s)"
---

# keyframe-channels — 四通道 transform 关键帧 showcase

## 这个方向测什么

把关键帧覆盖到**全部四个 transform 通道**（不止位置）：Position / Scale / Rotation / Opacity，每通道一行、两个线性关键帧（start≠end，4 秒）。**渲染正中 t=2s**，每个形状落在**插值中点值**（不是任一端点），这就是单帧内「逐通道关键帧插值确实渲染」的证据。`.done` 同时用 `valueAtTime(2.0)` 把每通道 t=2s 值读回，像素 + readback 双重佐证。

> 与 `keyframes-ease` 的区别：那个方向变的是**位置的缓动曲线**（linear/ease-out/ease-in-out）；这个方向插值都是平凡线性，但**横扫四个通道**，证明各通道关键帧流都能写+渲。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `keyframe_channels.aep`（5 层 = 4 通道动画 + BG） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(2.0)` + 每层 `valueAtTime(2.0)` 四通道 readback → png/.done |
| `keyframe_channels.aep` | 产出工程 (gitignored) | 1920×1080，30fps×4s，AE2020 target |
| `keyframe_channels.png` | 渲染帧 (gitignored) | **t=2s（中间帧）** |

## 布局（t=2s 时每行的插值状态 = 关键帧证据）

| 行 | 图层 | 颜色/形状 | 通道动画 | t=2s 期望 | 视觉读法 |
|---|---|---|---|---|---|
| 1 | 1_Position | 橙 圆点 | Position x 300→1620 | x≈960（正中） | 点居水平中心 |
| 2 | 2_Scale | 青 方块 | Scale 25%→150% | ≈87.5% | 中等大小方块 |
| 3 | 3_Rotation | 琥珀 竖条(56×200) | Rotation 0°→180° | 90° | 竖条**已转成横条** |
| 4 | 4_Opacity | 粉 方块 | Opacity 100%→10% | ≈55% | 明显**变暗/半透**的粉 |

> 审核要点：① 橙点在水平正中；② 琥珀条是**横的**（转了 90°，初始是竖的）；③ 粉方块**发暗**（半透明，非满亮粉）；④ `.done` readback 四值 = `pos x≈960 / scale≈87.5 / rot=90 / opac=55`，与像素一致。任一形状若停在端点（如琥珀条仍竖直、粉块满亮）= 插值未生效。

## 溯源

S1 ease 关键帧 + 多关键帧 gate：`mg_ease_shipgate_test.go`；关键帧字节布局 `incidents/keyframe-byte-layout-dispatcher.md`；>4 关键帧容量分页坑（本档每通道仅 2 kf，规避）`lhd3-keyframe-capacity-pages.md`。
