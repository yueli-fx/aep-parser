---
status: active
when_to_read: 造任何程序化视觉(火/烟/风/雨/雷电/能量/转场)选效果/技法时;想知道某技法跨哪些现象复用;建新现象配方 build-X 时查可复用技法原子
applies_to: [fx-techniques, procedural-fx, technique-library, fire, wind, rain, lightning, transition, noise, displacement, glow, blend-modes, cross-domain]
when_to_update: 新增/验证一个技法原子;某技法发现新实现效果或新适用现象;某技法被 AE gate 验证
last_updated: 2026-06-18
---

# 程序化视觉技法库（跨域可复用原子）

> **这是什么**：`specs/2026-06-18-fx-technique-internalization.md` 的两层知识结构之**技法层**——按「角色/功能」组织的**跨域可复用技法原子**。现象配方(`checklists/build-<现象>.md`)由这些原子组合而成。每喂一个参考模版,新技法进这里、已有技法标新适用现象(交叉链越用越强)。
>
> **解析工具**：`go run ./cmd/aepdissect <file.aep>` 出结构化报告(效果用量+原生/Cycore/第三方分级+预合成嵌套+逐层效果链),是内化流水线的 PARSE 步。
>
> **验证状态约定**：✓=该现象已实证用到 · ○=推断适用待验 · 标「validated」=经 AE gate。当前原子源自 Colorful Fire Ball 解析(实例#1),火焰栏多为 ✓,其它现象为 ○(待喂对应模版验证)。

## 技法原子

### T1 噪声造质料 (noise-substance)
- **做什么**：生成湍流噪声纹理,当作自然现象的"基本物质"(不是画形状)。
- **实现**：`ADBE Fractal Noise`(native)。
- **关键参数**：Contrast(高→锐利边界/分明)、Brightness、Scale Width/Height(非等比→方向拉伸)、Complexity(细节层数)。
- **跨现象**：火✓ · 烟○ · 云○ · 风(气流)○ · 水焦散○。
- 详 `checklists/build-good-fire.md` 参数表。

### T2 位移扭曲 (displacement-distortion)
- **做什么**：用一个源(通常是噪声)推像素,把平滑的东西推成有机的弯曲/舔动/起伏。**形态"活"的关键**。
- **实现**：`ADBE Displacement Map`(native,**用另一层当位移源**,经 `SetEffectLayerParam` 指定)· `ADBE Turbulent Displace`(native,自带噪声、更简单)· `ADBE MESH WARP`/`WRPMESH`(大尺度网格弯)· `PEDX`=Displacer Pro(⚠第三方)。
- **关键参数**：Max Horizontal/Vertical Displacement(方向+幅度;火焰=大垂直量拉火舌)、Size、Evolution(随时间变=动)。
- **跨现象**：火舌✓ · 风吹弯/热浪○ · 水波/雨幕扰动○ · 扭曲转场○。

### T3 多层 Add 叠深度 (additive-depth) ⭐
- **做什么**：把同一套"质料+扭曲+上色"图层**复制 N 份**(不同位移/偏移/色),用混合模式叠 → **层次感/深度**。**这是结构技法,不是单效果**(单层无论调参到顶都没深度)。
- **实现**：`NewSolidLayer`×N + `Layer.SetBlendingMode`。`Add(相加)`重叠处变亮=自动长白热芯;`Difference(差值)`做"负相"层=暗筋/内部纹理;`Divide`压外层;`Screen`叠亮不溢。
- **关键**：层数、各层位移/偏移/色温差异、混合模式选择。
- **跨现象**：火层次✓(样本 Fire 4 层) · 闪电辉光叠○ · 能量○。**库可做,全 gated。**

### T4 亮度→调色板 (luminance-color)
- **做什么**：把灰度(噪声)按亮度映射到颜色,得到色温渐变。
- **实现**：`ADBE Tritone`(native,3 档:阴影/中间调/高光,本库验证可用)· 多层 `ADBE Ramp`(native,生成渐变后被位移)· `ADBE Hue/Saturation`微调。**注**:经典的 Colorama **不在本库 216 效果集**。
- **关键参数**：三/两档颜色端点(ARGB 0-255);火=阴影深红→中橙→高光黄白热芯。
- **跨现象**：火色温✓ · 闪电电色○ · 能量○ · 卡通上色○。

### T5 辉光/泛光 (emissive-glow)
- **做什么**：让亮处向外发光(bloom),给发光体 emissive 质感。
- **实现**：`ADBE Glo2`(Glow,native)。
- **关键参数**:`-0002` Threshold(阈值,只对亮处)· `-0003` Radius(半径)· `-0004` Intensity(强度)。常**在多个图层上分别用**。
- **跨现象**：火✓ · 闪电/电弧○ · 霓虹○ · 能量○。

### T6 时间驱动参数 (time-evolution)
- **做什么**：让某参数随时间变=现象"活"(翻腾/流动/下落/闪烁)。
- **实现**：`AnimateEffectParam`(标量)/`AnimateEffectParamVec`(矢量,点)打关键帧。**用关键帧不用表达式**——本库写的表达式 AE 端不求值(见 `incidents/expression-enable-byte-pair.md`)。
- **关键**：动哪个参数。火=Fractal Noise 的 Evolution(翻腾)+ Offset(上滚);风=方向位移流动;雨=下落;雷=闪烁。
- **跨现象**：火翻腾/上升✓ · 风流动○ · 雨下落○ · 雷闪烁○。

### T7 轮廓塑形 (silhouette-shape)
- **做什么**：把满屏的质料裁成现象的整体外形。
- **实现**：羽化 `AddMask`+`SetFeather`(火苗/任意形)· `CC Sphere`(⚠Cycore,球/火球)· 径向 `Ramp`/遮罩(球形近似,plugin-free)。
- **跨现象**：火苗形✓ · 火球(样本用 CC Sphere)✓ · 能量球○。

### T8 粒子发射 (particle-emit)
- **做什么**：发射大量小粒子(火星/雨滴/雪/火花/碎屑)。
- **实现**：`CC Particle World`(⚠Cycore 自带,人人能渲但非 native、未 gate)。**plugin-free 无原生粒子替代**(可用噪声+阈值近似,效果弱)。
- **跨现象**：火星✓ · 雨○ · 雪○ · 火花○ · 碎屑转场○。

### T9 收尾调色 (final-grade)
- **做什么**:统一整体氛围(对比/色调/暗角)。
- **实现**：`ADBE Brightness & Contrast 2` · `ADBE PhotoFilterPS` · `CS Vignette`(⚠第三方,可用 native 径向遮罩替代)· `ADBE Hue/Saturation`。
- **跨现象**：通用收尾。

## 现象配方索引（技法的组合）

| 现象 | 配方 | 用到的技法 |
|---|---|---|
| 火焰 | `checklists/build-good-fire.md` | T1+T2+T3+T4+T5+T6+T7(+T8/T9 可选) |
| 风 | (待建) | T1+T2(方向)+T6+运动模糊 |
| 雨 | (待建) | T8(条状)+T6(下落)+模糊 |
| 雷电 | (待建) | 分形分支(Advanced Lightning)+T5+T6(闪烁)+T3 |
| 转场 | (待建) | T2/擦除+T6(时间扫过) |

> 新增现象:跑 `aepdissect` 解析模版 → 拆角色 → 在此登记新技法/标已有技法新适用现象 → 写 `checklists/build-<现象>.md` 配方 → AE gate 验证。
