---
showcase: glitch
direction: 纯 native（plugin-free）glitch — 基于 Booyah/GlitchText 抽出的原生技法，从零拼 RGB 色差 + 噪声驱动位移撕裂 + 辉光，AE 实渲验证 glitch 技法原子（docs/fx-techniques.md T16 + T2/T5 的 glitch 迁移）
capabilities: [rgb-channel-split, displacement-glitch, emissive-glow, add-blend, fill-effect, displacement-map, fractal-noise, glo2-glow, set-effect-layer-param, animate-effect-param, shape-fill-position, adjustment-layer]
gates: []
status: 待review
last_updated: 2026-07-03
regenerate: "go run ./showcase/glitch  +  scripts/ae-worker/ae_run.ps1 render.jsx"
---

# glitch — 纯 native glitch showcase（VALIDATE：glitch 技法实渲）

## 这个方向测什么

证 motionbox glitch 研究抽出的技法原子**能纯 native 渲出来**。这是一个技法门禁/烟测，不是 `showcase/booyah-clone` 的 1:1 复刻；Booyah 复刻仍由 `booyah-clone` 目录承载。一帧里同时呈现 3 个空间 glitch 技法：

主体 = 可读的「GLITCH」文字（在 `TXT` 预合成里，主合成实例化 3 份；点文字不能直接改色/位移，但预合成实例是普通 AV 层，用 Transform 效果缩放+定位、Fill 上色）。

- **T16 rgb-channel-split**：3 个文字实例分别 `Fill` 纯 R/G/B，Transform 微错位，`Add` 混合 → 字母左缘**青**、右缘**橙红**色差（chromatic aberration）。
- **T2 displacement-distortion（glitch 味）**：`Displacement Map` 调整层，源 = 高对比大横块 `Fractal Noise`（**眼睛关闭，仅当位移图源 → 背景保持干净**）→ 文字横向切片撕裂。Max H 位移关键帧抖动。
- **T5 emissive-glow**：`Glo2` 调整层给文字 neon bloom。

> **扫描线（T17 Venetian Blinds）已移除**：用户实测 CRT 行栅在压暗/添糊，去掉后字更鲜亮通透。T17 仍由 Booyah 样本 observed，本 showcase 不演示。

（T18 temporal-glitch 是时间域，单帧 render 验不了，未纳入。）

覆盖 `AddEffect`(Fill/Displacement Map/Fractal Noise/Glo2) + `SetEffectParam` + `SetEffectLayerParam`(Displacement Map 源指向 Noise 层) + `AnimateEffectParam` + `Layer.SetBlendingMode(Add)` + 形状 `Fill.SetColor`/`Position.SetStaticValue` + `NewAdjustmentLayer`。底层每个能力均已双版本 ship-gate；本 showcase 是 **glitch 组合** 的实渲。

## 产物

| 文件 | 类型 | 说明 |
|---|---|---|
| `gen.go` | 生成器(Go, tracked) | 构建 `glitch.aep`（1920×1080，AE2020 target，2 comp：`TXT` 文字源 + `GLITCH` 主合成 7 层） |
| `render.jsx` | 渲染脚本(tracked) | AE `saveFrameToPng(1.5)` → `glitch.png`（8bpc + purge 缓存） |
| `glitch.aep` / `glitch.png` | 产出 (gitignored) | |

## 应看到

居中的「**GLITCH**」文字：字母**左缘青、右缘橙红**（RGB 色差），有横向切片撕裂错位，neon 辉光，**干净深色背景**（噪声仅作位移图源、不可见）。全 native，stock AE 即可渲（无需任何插件）。

## 验证状态（2026-07-03）

- **agent 实渲 + 眼验：✅**（AE 2025，`scripts/ae-worker/ae_run.ps1`，exit 0）。4 个技法肉眼全部可见、零第三方依赖（aepdissect: RENDER DEPENDENCIES = none）。
- **本次 gate 重跑：✅**。`go run ./showcase/glitch` 重新生成 `glitch.aep`；`aeptechnique -mode explain` 读出 `analysis_ready`；`aepdissect` 仍为 no third-party。AE2025 与 AE2020 均用绝对 `-Jsx` / `-Done` 路径完成 render，`glitch.png` 已更新。初次相对路径调用曾表现为 startup timeout，已记录到 `tmp/technique_showcase_glitch/package_summary.json`。
- **⚠ 待用户真机验收**：agent 眼验 ≠ 用户真机验收（showcase review-gate）。未建独立 `TestGlitch_AEShipGate`（底层 caps 已各自 gated；组合走 showcase 眼验）。
- 升级路：若要把 T16 confidence 升 `validated`，补一个 `TestGlitch_AEShipGate_AE2020/AE2025`（双版本 + 像素采样）。

## 来源 / 配方

基于 `data/samples/motionbox/glitch/booyah-glitch` 与 GlitchText 样本抽出的角色/技法拆解。技法原子定义见 `docs/fx-techniques.md` T16/T2/T5；理解流程 `checklists/techniques/understand-a-project.md`（扫描线 T17 见样本，本配方已弃）。完整 Booyah 复刻看 `showcase/booyah-clone`。
