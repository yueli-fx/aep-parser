---
status: active
when_to_read: 被要求从零造火焰/烟/能量类程序化 FX；要重做火焰 v2；调火焰参数想知道某个旋钮的视觉影响；纠结哪些效果是核心、哪些靠插件
applies_to: [fire, flame, procedural-fx, fractal-noise, displacement-map, ramp, glow, recipe, parameters, plugin-free, sample-analysis, AnimateEffectParam, blend-modes]
last_updated: 2026-06-18
---

# 造一个好火焰 — plugin-free 原生配方 + 参数影响对照

> **这份文档的用途**：跨会话固化「怎么用本库从零造一个**好**火焰」的知识。下一个会话开场 preflight 会把 `checklists/INDEX.md` 读进上下文，所以这条会被自动看到。配合 cockpit Active focus（火焰 Phase 2）即可接力，不必重新解析样本。

## ⚠ 成熟度（诚实前提，别当已验证流程用）

- **参数→效果对照表 = 事实**：从真实样本 `samples/Colorful Fire Ball`（by Plugin Everything）用 parser 解析读出（`go run ./cmd/aepdissect <file>` + `tools/debug/dump_tdmn`）。详 `specs/2026-06-18-procedural-fx-generator.md` § 样本解析 #1。
- **✅ 已升级为「验证配方」（2026-06-18）**：**v3 多层合成**（黑底+3 火层+同心 mask 温度分区+Add+Glo2，运动共相）经 **`TestFlameDemo_AEShipGate_AE2020/AE2025` 双版本 gate 绿（逐像素一致）+ 用户真机验收过**。这条多层配方现在是验证过的,可照搬。
- 历程：v1 被否（单层 Tint）→ v2（单层 Tritone+Glo2,修色温/Glow 但无层次）→ **v3 多层合成出层次（通过）**。下面参数表/步骤为 v3 实际值。

## 一句话原理

**好火焰 = 噪声生成形态 + 位移驱动「舔动」+ 色温渐变上色 + 辉光 + 多层合成出深度。** 不是"画个水滴形糊上橙色"（那是 v1，被否）。也**不是靠单层调参**——单层无论怎么调，天花板就是"一团有色温的火"（v2 实证，见下）；**层次感来自多层合成结构，不是参数**。真实专业火焰是 7 层预合成的深度合成（见样本解析），核心火焰是纯原生、可复刻。

## 源工程真实结构 — aepdissect 三轴重解析（2026-06-18）

> `go run ./cmd/aepdissect "samples/fx/Fire/Colorful Fire Ball/Colorful fire AE 2023.aep"`。比早先读法**深一层**：深度不是「单 comp 平铺 3–4 个 Add 层」，而是 **5 级预合成嵌套**，每级一个 displaced/recolored pass，再用 Add/Difference/Divide 重组。

```yaml
project_profile:   # Colorful Fire Ball (by Plugin Everything)
  meta: {type: [reel], resolution: 1200x1200, duration: 55s, fps: 30}
  fingerprint: {comps: 7, nest_depth: 5, top: "Comp 1"(10层)}
  graph:           # ← 深度的真相：嵌套式 pass 累积
    Noise 1 → Noise 1 looped → Fire → Whole fire animation → Comp 1
    旁支: Outer fire(CC Sphere 球形包裹) · Particles(CC Particle World 火星)
  signal_comps:
    - Noise 1:              noise-as-material   # Fractal Noise: Contrast169, 竖拉 ScaleW79/H210, Invert; expr 上滚 time*[0,-500] + 翻腾 time*90
    - Noise 1 looped:       seamless-loop       # 2×time-offset copy + opacity crossfade + SilhouetteAlpha matte → 无缝循环
    - Fire:                 displace+color      # Distortion adj(DispMap MaxV280) + 3×[DispMap + Ramp 各自色温] Add
    - Whole fire animation: depth+glow          # "Negative Fire"=Difference + Fire=Add + Particles=Divide + Glo2(thr175 r125 / thr108 r68)
    - Comp 1:               assemble+grade      # 2×Outer fire(Divide+TimeRemap) + Whole fire(+BoxBlur+Transform) + 背景(Checkerboard/Grid/Solid Composite) + B&C/Photo Filter/Vignette
  techniques: [noise-as-material, seamless-loop, displacement-distortion, luminance-color,
               additive-multilayer-depth, emissive-glow, time-evolution,
               silhouette-shape, particle-emit, final-grade]
  reproducibility:
    native:       核心火焰全原生 — Fractal Noise / Displacement Map(×6) / Ramp(×4) / Glo2(×3) / Add·Difference·Divide 嵌套
    cycore:       CC Sphere(球形) · CC Particle World(火星) — bundled 能渲、未 gate
    third_party:  CS Vignette(×2) · PEDX=Displacer Pro(×1) — 可 native 替代(径向遮罩 / Displacement Map)
  verification: v3 简化 5 层合成已双版本 gate + 用户验收；此 5 级全嵌套未逐级复刻(render-diff 未做)
```

**两条比旧读法更新的认知**：
1. **深度 = 嵌套式 pass，不是平铺。** `additive-multilayer-depth` 在这工程的真实形态是「每个预合成贡献一个 displaced+recolored pass，外层用 Add/Difference/Divide 累积」——见 `docs/fx-techniques.md` T3（已据此 refine）。
2. **"Negative Fire" 是实证不是假设。** `Whole fire animation` 里真有一层名为 "Negative Fire"、blend=Difference、加 Warp(Bend −100) + Hue/Sat —— Difference 叠暗筋这条从「猜测」升为 proven。
3. **新技法 `seamless-loop`**：`Noise 1 looped` = 把演化中的噪声源复制 2 份、时间错位 + opacity 交叉淡入 + SilhouetteAlpha matte → 无缝循环。跨域（烟/云/能量/水任何演化噪声都可循环）。已登记 `docs/fx-techniques.md` T15。

## ⭐ 效果用途词典 —— 什么效果干什么用（按角色，比参数表重要）

这是生成器真正需要的知识：**按「角色/功能」理解每个效果在火焰里的作用**（AI 层是按角色拼装，不是抄参数）。源自 Colorful Fire Ball 解析。

> **跨域技法原子在 `docs/fx-techniques.md`**（技法库）。本表是火焰这个**现象**怎么组合那些技法；技法本身(噪声造质料/位移扭曲/多层Add叠深度…)跨火/烟/风/雨/雷电/转场复用。解析任意模版用 `go run ./cmd/aepdissect <file.aep>`。

| 角色（要解决的问题） | 用什么效果/技术 | 怎么起作用 |
|---|---|---|
| ① **形态/质料**（火的"料子"） | **Fractal Noise** | 生成湍流噪声纹理 = 火的基本物质；高对比→锐火舌，竖拉→瘦高 |
| ② **轮廓**（整体外形） | 羽化 **Mask** /（球形）**CC Sphere** | 把噪声裁成火苗/火球形 |
| ③ **舔动/扭曲**（边缘有机不死板） | **Displacement Map**（噪声驱动、**大垂直量**）· Turbulent Displace · Mesh Warp · PEDX | 用噪声推像素，把平滑渐变推成弯曲火舌。**样本主力 = Displacement Map** |
| ④ **颜色/色温** | 多层 **Ramp**（不同色温）· Tritone · Hue/Sat | 给灰度噪声上火色，垂直色温渐变 |
| ⑤ **层次感/深度** ⭐ | **③④ 的图层重复 N 份 + 混合模式叠加** | 见下节，**是合成结构不是单个效果** |
| ⑥ **运动**（活/翻腾） | Fractal Noise 的 **Evolution**（翻腾）+ **Offset**（上滚）+ 无缝循环层 | 参数随时间动（用关键帧，非表达式） |
| ⑦ **辉光/泛光** | **Glo2 (Glow)**（且在多个图层上分别用） | 亮处 bloom，emissive 感 |
| ⑧ **火星/迸射** | **CC Particle World**（自带插件） | 粒子喷发 |
| ⑨ **收尾氛围** | Brightness/Contrast · Photo Filter · Vignette | 整体调色 |

## ⭐ 层次感 = 合成结构，不是参数

示例的"深度"是堆出来的，**没有一个叫"层次感"的效果**：

1. **同一套「噪声→Displacement Map→Ramp」图层复制 3–4 份**，每份给**不同的位移量/偏移/色温**（样本 Fire 合成 = 4 层）。
2. **混合模式叠**：`Add(相加)` 重叠处变亮 = **自动长出白热芯**（亮叠亮）；`Difference(差值)` 做一层"**负火 Negative Fire**" = 叠暗筋/内部纹理；`Divide` 压外层。
3. **内焰/外焰/粒子拆成各自预合成**，各贡献一层再合。
4. 全局 `Glow` + 收尾调色统一氛围。

**= 重复的位移彩色层 + 混合模式 + 正火/负火 + 内外分层。** 本库能做：多 `NewSolidLayer` + `AddEffect` + `Layer.SetBlendingMode` + `Displacement Map` 经 `SetEffectLayerParam` 指噪声层（全已 gate）。**做火焰 v3 = 搭这个结构，不是继续调单层参数。**

### v2 实证（2026-06-18，单层天花板）

单层栈 `Fractal Noise → Tritone(三档色温) → Turbulent Displace → Glo2` + 羽化 mask + 黑底，AE2025 自渲：**色温/辉光/翻腾都对了**（修掉 v1 三个缺口），但**层次感不足**（用户：「调再多参数没意义了，示例是多层混合的」）。→ 坐实「单层调参到顶 = 一团有色温的火」，深度必须上多层合成。（v2 builder 为迭代 scratch，已清理；最终 v3 builder=`flightdeck/showcase/procedural-fx/gen.go`，tracked。）

### v3 实证（2026-06-18，多层合成 → 层次感出来了）✓ 自渲验证 T3

5 层结构(builder=`flightdeck/showcase/procedural-fx/gen.go`，tracked)：黑底 + **3 个 Fractal Noise→Tritone→Turbulent Displace 火层**(不同噪声 scale=不同细节频率，越内层 Tritone 越热) + 顶部 Glo2 调整层。双版本 AE 实渲 + 用户验收：清晰的**径向色温分层**——外深红 wispy → 中橙 → **内黄白热芯柱**，有舔动有翻腾，明显比 v2 有深度。

**两个关键踩坑 → 解法**(都记进 `docs/fx-techniques.md` T3):
1. **团块**：3 层 Add **共用同一 mask** → streak 并集填满成"发光团块"。**解法=同心 mask**(`scalePath` 缩 f=0.5/0.76/1.0,核层小/外层全)+ 各层 Tritone 越内越热 → 温度分区(外红中橙内白)而非实心。
2. **抖动**(用户在 ~3s 处发现):3 层各用**不同动画速率** → 逐渐失相,Add+Glow 出拍频闪烁,越后越抖。**解法=层运动共相**(shared evolution/offset 速率),只让静态属性(scale/色/mask)分层。修复后 t=3.2/3.6 平稳。

⚠ 仍是 agent 自渲，待用户真机验收。

## 核心配方（plugin-free，全 ADBE 原生，本库已 gate）

按本库已验证的 from-scratch render harness（镜像 `orbit_demo`/v1 flame：`NewSolidLayer` + `AddEffect` + `SetEffectParam` + `AnimateEffectParam` + `AddMask`/`SetFeather` → `WriteAEP` → AE 渲一帧 PNG）：

1. **底噪 = `ADBE Fractal Noise`**：高对比（样本 Contrast≈169）、竖向拉伸（Scale Width 小 / Height 大 → 瘦高火苗）。
2. **动画**：给 Fractal Noise 的 **Offset Turbulence 沿 −Y 随时间推**（火往上滚）+ **Evolution 随时间增**（内部翻腾），用 **`AnimateEffectParam` 打关键帧**（**不要用表达式**——样本用 `time*[0,-500]`/`time*90` 表达式，但本库写的表达式 AE 端不求值，见 [[expression-enable-byte-pair]]；v1 已证 Evolution 关键帧可行、没撞 elision 缺口）。
3. **火舌「舔动」= `ADBE Displacement Map`**：用上面的噪声层当位移源，**Max Vertical Displacement 给大值（样本 ~150–280）**把图像向上拽成火舌。**这是 v1 缺的关键**——v1 用了 Turbulent Displace（自带噪声、控制弱），样本用的是独立高对比噪声驱动的 Displacement Map。
4. **颜色 = `ADBE Ramp`（垂直色温渐变）**：多层不同色温的线性 Ramp（红/橙/黄端点），**层间 `Add(4)` 混合叠出明亮热芯**。这给用户要的「白→黄→橙→红→暗尖」色温（v1 用 Tint 2 色，正是被否的主因）。
5. **泛光 = `ADBE Glo2`（Glow）**：阈值只对亮处发光 + 半径/强度（样本 Intensity≈68）。补用户要的「Glow 泛光」。
6. **外形**：火苗轮廓用羽化 mask（`AddMask`+`SetFeather`）；若要"火球"形，用径向遮罩/径向 Ramp 近似（样本的球形靠 `CC Sphere` 插件，plugin-free 只能近似）。

## 参数 → 视觉效果 对照表

| 效果 | 参数 | 调大/调小的视觉影响 | 样本值 |
|---|---|---|---|
| **Fractal Noise** | Contrast | 大 → 火舌边缘锐利、黑间隙分明；小 → 糊成一团 | **169（高）** |
| | Brightness | 整体明暗 | — |
| | Scale Width / Height | W 小 + H 大 → 瘦高竖向火苗；等比 → 团状 | 竖拉 |
| | Complexity / Sub-Influence | 大 → 细碎火丝多 | — |
| | **Offset Turbulence**(point) | 沿 −Y 随时间推 → **火向上滚动** | expr `time*[0,-500]` |
| | **Evolution**(angle) | 随时间增 → **内部翻腾/boiling** | expr `time*90` |
| **Displacement Map** | **Max Vertical Displacement** | **大 → 火舌向上拉长舔动**；0 → 死板不动 | **~150–280** |
| | Max Horizontal Displacement | 小幅 → 左右轻摆；大 → 撕裂 | -22 / 81 |
| | Displacement Map Layer / Use For | 指向噪声层 + 用其亮度通道 | 噪声层 |
| **Ramp** | Start/End Color | 火色温端点；多层不同色温 + Add → 白热芯+渐变 | 红→橙→品 ARGB |
| | Ramp Shape | Linear（垂直）；Radial 可做球形光 | Linear |
| **Glo2 (Glow)** | Glow Threshold | 高 → 只最亮处发光（芯）；低 → 整体泛白 | — |
| | Glow Radius / Intensity | 大 → 光晕宽/强 | Intensity≈68 |
| **Blend mode** | Add(4) | 叠亮、出热芯 | Fire 层 |
| | Difference(26) | 出暗筋/负向细节（"Negative Fire"） | — |
| | Screen / Lighten | 也可叠火不变暗 | — |

> 注：上表 Fractal Noise/Displacement Map 的具体 param **index→名** 以 v1 `flame/gen.go` 里已用对的映射为准（v1 设值成功）；表里写"待确认名"的，build v2 时按本库 API 的 matchName/index 现取现核，别照搬 `-NNNN`。

## 插件分级（决定能复刻到哪）

| 层 | 用什么 | plugin-free 能否复刻 |
|---|---|---|
| **核心火焰** | Fractal Noise + Displacement Map + Ramp + Glo2 **全原生·已 gate** | ✅ 直接做 |
| 火球外形 | `CC Sphere`(Cycore 自带) + `PEDX`=Displacer Pro(**第三方**·需装+GPU) | ⚠ 第三方违反「人人可开」；native 径向近似 or 砍 |
| 火星/迸射 | `CC Particle World`(Cycore 自带) | ⚠ 无原生粒子替代；自带能渲但未 gate |

**第三方插件二进制**（`.aex`/`.plugin`）留 `samples/`（gitignore），**不进 `internal/serializer/templates/`**——产出 .aep 依赖插件就违反交付准则。

## v1 被否对照（别再犯）

| 用户抱怨 | v1 错法 | 正确法 |
|---|---|---|
| 无白→黄→橙→红色温 | `Tint`（2 色） | 多层 `Ramp` 色温渐变 + Add |
| 无 Glow 泛光 | 没加 | `Glo2` |
| 只有形态、没深度 | 单层 | Add 叠多层位移彩渐变 |
| 不像在"舔" | `Turbulent Displace`（自带噪声） | `Displacement Map` 驱动独立高对比 Fractal Noise，大垂直量 |

## 来源 / 交叉链接

- 样本解析全文（渲染图 + 效果用量 + blend 解码）：`specs/2026-06-18-procedural-fx-generator.md` § 样本解析 #1。
- 教训（手搓只到「可辨认」、需真实样本）：[[procedural-fx-over-vector]] Case 2。
- v1 实现（被否，保留作对照）：`flightdeck/showcase/procedural-fx/`（status ❌质量未过）。
- 动画走关键帧不走表达式的原因：[[expression-enable-byte-pair]]。
- **新增样本时**：在此表追加"样本解析 #N"的参数范围，并更新对照表（多样本对齐 = Phase 2 通过判据）。