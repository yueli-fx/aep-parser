# Showcase 验证记录

[展示与运行入口](README.md) · [返回项目首页](../README.md)

本页保留各展示方向的验证状态与已知边界。生成器、脚本和说明随源码分发，生成的 AEP 与图片在本地重建。

## 状态说明（review-gate）

**🔍 待review** = agent 已建 + AE 实渲 + 自己眼验，**等用户在真机打开 .aep 复核**；**✅ complete** = **用户真机验收过**。agent 不自标 complete（详 `docs/knowledge/showcase/showcase.md`）。

16 个 🖼 可视能力方向 ✅ complete（2026-06-14 用户真机逐个验收通过；gradient +radial 类型、expressions +四语汇 idiom 的增量同日复核通过；[3d-camera] 2026-06-15 通过；[shape-filters] 2026-07-03 复核通过）。番外 [procedural-fx] 也已 complete。📋 B 类读值档：4 个 from-scratch 可行已补并 complete，3 个 from-scratch 不可表达（见末节）。

## 🖼 A 类 — 可视方向（看图档）

| 方向 | 测什么 | 状态 |
|---|---|---|
| [shape-filters](shape-filters/INDEX.md) | 形状矢量滤镜家族 11 件 + PolyStar + 双滤镜叠加 + **Wiggle 调制对比行**（Points/Correlation） | ✅ complete（2026-07-03 用户确认通过） |
| [shape-primitives](shape-primitives/INDEX.md) | 四种参数图元 + Fill/Stroke/Gradient 描绘变体 | ✅ complete |
| [keyframes-ease](keyframes-ease/INDEX.md) | 时间缓动 linear/ease-out/ease-in-out（渲中间帧看位置差） | ✅ complete |
| [expressions](expressions/INDEX.md) | 表达式激活（time\*N 旋转）+ 四语汇 idiom（跨层/loopOut/wiggle/slider） | ✅ complete |
| [precomp-nesting](precomp-nesting/INDEX.md) | 预合成嵌套（parent 嵌套 child 组合场景） | ✅ complete |
| [gradient](gradient/INDEX.md) | 渐变填充 + ramp 方向 + **类型**（4 线性 + 2 径向同心） | ✅ complete |
| [text](text/INDEX.md) | 文字层 from-scratch（NewTextLayer + SetText 多行） | ✅ complete |
| [layers](layers/INDEX.md) | Solid 同心色框 + Null/Adjustment 建层 | ✅ complete |
| [effects](effects/INDEX.md) | 加效果 AddEffect + 改参数 SetEffectParam（12 效果网格） | ✅ complete |
| [masks](masks/INDEX.md) | AddMask 遮罩裁切（圆/三角/星/inverted） | ✅ complete |
| [structural-ops](structural-ops/INDEX.md) | Duplicate/Move/Delete/维度分离（同心环+readback） | ✅ complete |
| [transform-values](transform-values/INDEX.md) | 改 Position/Scale/Rotation/Opacity（before/after） | ✅ complete |
| [stroke-detail](stroke-detail/INDEX.md) | 描边细节 Dashes/Line Join/Miter/Wave（闭合星，Cap·Taper 诚实暂缺） | ✅ complete |
| [keyframe-channels](keyframe-channels/INDEX.md) | 四通道关键帧 Position/Scale/Rotation/Opacity（渲 t=2s 插值 + readback） | ✅ complete |
| [animated-path](animated-path/INDEX.md) | 动画路径几何 morph 横条→正方→竖条（渲 t=2s + extent readback） | ✅ complete |
| [3d-camera](3d-camera/INDEX.md) | 3D 图层 + 相机：Z 视差（同尺寸卡按深浅渲大小不同）+ Rotate Y 透视 tumble（梯形） | ✅ complete |
| [procedural-fx](procedural-fx/INDEX.md) | 程序化火焰 v3（多层合成：黑底+3 火层 FractalNoise→Tritone→TurbulentDisplace+同心 mask 温度分区+Add 叠热芯+Glo2 Glow；技法 T3 additive-depth） | ✅ complete（用户真机验收过 2026-06-18；双版本 gate 绿、逐像素一致） |
| [glitch](glitch/INDEX.md) | 纯 native glitch 技法门禁（RGB 色差 split + 噪声驱动位移撕裂 + 辉光；不是 Booyah 1:1 复刻；零插件） | 🔍 待review（2026-07-03 Go/parse gate + AE2025/AE2020 render 重跑通过；待真机复核） |
| [rain](rain/INDEX.md) | 纯 native rain 技法门禁（shape/repeater 雨线 + 下落关键帧 + Fractal Noise 雾气 + Echo 拖尾 + trim 水花；不是 Particular/Unmult 精确复刻；零插件） | 🔍 待review（2026-07-04 Go/parse gate + AE2020 render 通过；待真机复核） |

## 📋 B 类 — 读值档（dump 值核对，不看图）

无渲染视觉的能力 = `gen.go` 建工程 + `verify.jsx` dump DOM 值到 `.done`,**用户读日志核值**（非看图）。

| 方向 | 测什么 | 状态 |
|---|---|---|
| [comp-settings](comp-settings/INDEX.md) | 合成设置 motionBlur/workArea/bgColor/hideShy/nestedFrameRate（7 项 DOM 一致） | ✅ complete |
| [project-settings](project-settings/INDEX.md) | 工程设置 bitsPerChannel/linearBlending/expressionEngine/footageTimecode（4 项一致） | ✅ complete |
| [camera-light](camera-light/INDEX.md) | NewCameraLayer/NewLightLayer 建层（**灯光=环境光** + 相机=双节点,值皆模板默认；选项 setter elide） | ✅ complete |
| [essential-graphics](essential-graphics/INDEX.md) | AddEssentialProperty + 模板命名 | 🛑 **BLOCKED**（面板崩溃，连单 slider 也崩；RE 候选，暂不修） |
| [pseudo-effect](pseudo-effect/INDEX.md) | BuildPseudoEffect 从零合成伪效果，单效果铺满全部 11 种控件（en 纯 ASCII + zh GBK 中文标签）；控件容器无渲染面 | ✅ complete（2026-07-03 用户确认；label 默认亮态，`pseudo_max.aep` 含一条故意 dimmed label） |

### ⚠ from-scratch 不可表达（3 个 — 非缺陷，能力本质是 fixture-mutation）

这三个能力**无法纯 Go 从零生成**,只能在已有 .aep（fixture）上改,故不符 showcase 的「纯 Go 从零」前提。它们各有 **fixture-based ship-gate** 验证(非 from-scratch showcase)：

- **markers** — `AddMarker` 报「empty comp marker set unsupported（需 canonical seed 克隆）」：空 marker 集无模板可克隆,只能往已有 ≥1 marker 的 comp 加。
- **render-queue** — `AddItem` 报「empty queue has no template item to clone」：同理需已有 RQ item 作种。
- **media-replace** — `SetAlternateSource` 需真实 footage 导入,from-scratch 无素材源。

> 这三项的写能力本身经各自 ship-gate（对 fixture）验证过；只是不适合做「从零 showcase」。需要时可改成「fixture-mutation 演示」（破坏「干净 clone 重生成」原则,故未做）。

### 发现的边界（readback + 用户真机核出，红线4a/4b 活样本 → RE 候选）

- **EG 面板崩溃 🛑 BLOCKED（红线4b，用户两次真机确认）**：从零 EG 工程展开「基本图形」面板崩溃 AE——3 混合控件崩,**退回单 slider 也崩**。DOM readback 全过=假绿,根源是 EG ship-gate（load+DOM+resave）**从不打开面板**。整个 from-scratch EG 不可交付,留作 RE repro,用户决定暂不修(EG 用得少)。
- **comp setter（红线4a）**：**SetShutterAngle/Phase** 读回 ×≈1.2、**SetResolutionFactor** 让 AE `resolutionFactor` 除零 → 已排除。
- **project setter（红线4a）**：**SetTimeDisplayType/FeetFramesFilmType**(共用 nnhd byte8 疑位打包)+**SetFramesCountType** AE DOM 不反映 → 已排除。
- **camera/light/material 选项 setter**：选项流在 embed 模板里 elide；解法 = synthesis-insert（camera iris / light color leaves 在 `newTemplatedLayer` 自动 splice；material leaves 经 `aep.SetMaterialOption` 按需 splice）或 reopen 后设值。**已退役的旧限制**（2026-06-15 更新）：~~NewLightLayer 无从零 SetLightType~~ → `SetLightKind` 从零生效（ldta @0x88，读回 LightType.POINT）；~~material 从零不可用~~ → `SetMaterialOption` 解锁从零 3D 阴影。剩 elide 项主要是 camera DoF/light intensity 等需 reopen-后-设值。

详各方向 INDEX「已发现边界/崩溃发现」。**建议独立 RE/修**。

---

> **16 个 🖼 可视能力方向 ✅ complete** · **4 个 📋 读值档 ✅ complete**(comp-settings / project-settings / camera-light / pseudo-effect) · **glitch 技法门禁 🔍 待review** · **EG 🛑 BLOCKED**(面板崩溃,RE 候选) · **3 个 from-scratch 不可表达（markers/render-queue/media-replace，已注明）**。全量清单 + 验证档位详 `specs/2026-06-13-full-showcase-coverage.md`。
>
> **番外 [procedural-fx] ✅ complete**（2026-06-18,非能力域,是 `specs/procedural-fx-generator` 产品 demo）:火焰 v3 多层合成,用户真机验收过 + 双版本 gate 绿。经 v1(被否)→v2(单层修色温/Glow)→v3(多层出层次)三轮。详 procedural-fx/INDEX.md + `checklists/techniques/build-good-fire.md`。
