---
status: active
when_to_read: 造任何程序化视觉(火/烟/风/雨/雷电/能量/转场)选效果/技法时;想知道某技法跨哪些现象复用;建新现象配方 build-X 时查可复用技法原子;新增/重构技法条目需对齐三轴 schema 时
applies_to: [fx-techniques, procedural-fx, technique-library, three-axis-ontology, role, technique, mechanism, fire, wind, rain, lightning, transition, noise, displacement, glow, blend-modes, cross-domain, reproducibility]
when_to_update: 新增/验证一个技法原子;某技法发现新实现效果(等价集)或新适用现象;某技法被 AE gate 验证(confidence 升级);schema 字段变更需对齐
last_updated: 2026-06-18
---

# 程序化视觉技法库(三轴本体 · 技法层)

> **这是什么**:`specs/2026-06-18-technique-ontology.md` 三轴本体的**技法层**实例库——按冻结 schema v2 组织的跨域可复用技法原子。现象配方(`checklists/build-<现象>.md`)由这些原子**有序组合**而成。每喂一个参考模版,新技法按 schema 追加、已有技法标新 `proven_transfers`(交叉链越用越强)。
>
> **schema/字段定义见 spec**;本文件是 instance 库(schema 闭、instance 开)。**解析工具** `go run ./cmd/aepdissect <file.aep>` = 内化流水线 PARSE 步。
>
> **confidence 约定**:`validated`=经 AE gate(火焰双版本 gate 链内的技法)· `observed`=样本实证用到、未单独渲验 · `hypothesized`=推断/库里有但没渲过。**reproducibility.mechanism 取 any_of 里最可达的**(只要有一个 native 实现技法即可复刻;第三方/cycore 是增强选项,记在备注)。

## 两类元素:程序化 vs 素材+装配(aepdissect 自动判别)

- **① 程序化**(火焰 Colorful Fire Ball):.aep 用效果**生成**视觉 → **能学会复刻**(T1–T9,多 `requires_asset: none`)。
- **② 素材+装配**(闪电 Lightning Pack):重上色/辉光/控制器包在**外部素材**上,真正视觉在素材里 → **装配能复刻(T10–T13),元素得用户提供**(`requires_asset: footage`)。
- **判别信号**:layer `src=footage(..)` + 0 关键帧 + 效果只有重上色/辉光/调色 → ②型;反之(solid 源 + 生成类效果 + 关键帧)→ ①型。

---

## 技法原子(三轴 schema v2)

### T1 noise-as-material(噪声造质料)
生成湍流噪声当"基本物质"(非画形状)。高对比→锐边,竖拉→瘦高。
```yaml
id: noise-as-material
role: [form, texture]
mechanism:
  - kind: effect
    any_of: [ADBE Fractal Noise, ADBE Turbulent Noise]
    signal:
      - {param: Contrast, direction: up, observed_range: [128, 185]}
      - {param: ScaleWidth, direction: down}
      - {param: ScaleHeight, direction: up}
      - {param: Brightness, direction: down}
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire]
hypothesized_transfers: [smoke, clouds, wind, water]
not_this: "贴静态噪声纹理当背景(无形态生成意图)"
evidence: [showcase/procedural-fx]
confidence: validated
```

### T2 displacement-distortion(位移扭曲)
用一个源(常是噪声)推像素,把平滑的推成有机弯曲/舔动。形态"活"的关键。
```yaml
id: displacement-distortion
role: [distort, motion]
mechanism:
  - kind: effect
    any_of: [ADBE Turbulent Displace, ADBE Displacement Map, ADBE MESH WARP, ADBE WRPMESH]
    signal:
      - {param: MaxVerticalDisplacement, direction: up, observed_range: [102, 280]}  # 火舌=大垂直量
      - {param: Size, direction: set}
      - {param: Evolution, direction: up}                 # 随时间=动
      - {param: Bend, direction: set}                     # 仅 WRPMESH(Warp):整体弯曲
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire, glitch]
hypothesized_transfers: [wind, water, transition]
not_this: "整层平移(那是位置动画,不是逐像素扭曲)"
# glitch 用法:高对比/块状噪声驱动 → 横向块状撕裂(Booyah Displacement Map ×12 / GlitchText ×20);火焰用法=MaxV 拉火舌。
evidence: [showcase/procedural-fx, samples/ColorfulFireBall]
confidence: validated
# 备注:Displacement Map=主力(ColorfulFireBall 用 ×6,MaxV 最大 280),需用另一层当源(SetEffectLayerParam);
#       WRPMESH(Warp/Bend)在 "Negative Fire" 层作整体弯曲;PEDX=Displacer Pro 更强但⚠第三方(native DispMap 可替)
```

### T3 additive-multilayer-depth(多层 Add 叠深度)⭐
同套"质料+扭曲+上色"层复制 N 份叠加 → 层次/深度。**结构技法非单效果**(单层调参到顶也没深度)。
```yaml
id: additive-multilayer-depth
role: [depth]
mechanism:
  - kind: composition_op
    op: "同套效果链复制 N 份 + 各层不同 noise scale + blend 重组。两种形态:(a)单 comp 平铺多 Add 层(showcase v3);(b)多级预合成嵌套,每级一个 displaced/recolored pass(ColorfulFireBall:Noise→Loop→Fire→Whole→Comp1 五级)"
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire]
hypothesized_transfers: [smoke, energy, magic]
not_this: "单层调高对比假装有层次(v2 被否)"
evidence: [showcase/procedural-fx, samples/ColorfulFireBall]
confidence: validated
# 三种 blend 各司其职(ColorfulFireBall 实证):Add=亮叠出白热芯 / Difference=负相暗筋("Negative Fire"层,proven) / Divide=压外层(particles)。Screen=叠亮不溢(备选)。
# ⚠防团块:N 层共用同一 mask→并集填满成团块。解法=同心 mask(内小外大)+各层越内越热=温度分区。
# ⚠防抖动:各层动画速率不一→渐失相→Add+Glow 拍频闪烁。铁律=运动共相,只静态属性(scale/色/mask)分层。
```

### T4 luminance-color(亮度→调色板)
把灰度噪声按亮度映射到颜色,得色温渐变。
```yaml
id: luminance-color
role: [color, temperature]
mechanism:
  - kind: effect
    any_of: [ADBE Tritone, ADBE Ramp]
    signal:
      - {param: Highlights, direction: set}   # 火=黄白热芯
      - {param: Midtones, direction: set}     # 火=橙
      - {param: Shadows, direction: set}      # 火=深红
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire]
hypothesized_transfers: [lightning, energy, lava]
not_this: "整体 Hue 偏移(那是调色,不是按亮度分档上色)"
evidence: [showcase/procedural-fx]
confidence: validated
# 备注:多层 Ramp 不同色温+位移亦可;经典 Colorama 不在本库效果集。
```

### T5 emissive-glow(辉光/泛光)⭐跨现象复用实证
亮处向外 bloom,给发光体 emissive 质感。**fire + lightning 都 proven**。
```yaml
id: emissive-glow
role: [glow]
mechanism:
  - kind: effect
    any_of: [ADBE Glo2]
    signal:
      - {param: Glow Threshold, direction: down}   # 越低越整体泛白;高=只芯
      - {param: Glow Radius, direction: up}
      - {param: Glow Intensity, direction: up}
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire, lightning, glitch]
hypothesized_transfers: [neon, energy, magic]
not_this: "整体提亮(无阈值=不是选择性辉光)"
evidence: [showcase/procedural-fx, samples/LightningPack]
confidence: validated
# 常在多个图层分别用。
```

### T6 time-evolution(时间驱动参数)
让某参数随时间变=现象"活"(翻腾/流动/下落/闪烁)。
```yaml
id: time-evolution
role: [motion]
mechanism:
  - kind: composition_op
    op: "AnimateEffectParam(标量)/AnimateEffectParamVec(矢量)打关键帧。用关键帧不用表达式(本库表达式 AE 不求值)"
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire]
hypothesized_transfers: [wind, rain, lightning]
not_this: "静态值(无关键帧=不动)"
evidence: [showcase/procedural-fx]
confidence: validated
# 动哪个:火=Evolution(翻腾)+Offset(上滚);风=方向位移;雨=下落;雷=闪烁。
```

### T7 silhouette-shape(轮廓塑形)
把满屏质料裁成现象整体外形。
```yaml
id: silhouette-shape
role: [contour]
mechanism:
  - kind: composition_op
    op: "羽化 AddMask + SetFeather(任意形);径向 mask/Ramp 近似球形"
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire]
hypothesized_transfers: [energy, smoke]
not_this: "硬边裁剪(无羽化=边缘死板)"
evidence: [showcase/procedural-fx]
confidence: validated
# 备注:火球用 CC Sphere(⚠Cycore);plugin-free 用径向近似。
```

### T8 particle-emit(粒子发射)
发射大量小粒子(火星/雨滴/雪/火花)。
```yaml
id: particle-emit
role: [particle]
mechanism:
  - kind: effect
    any_of: [CC Particle World]
    signal: [{param: Birth Rate, direction: up}, {param: Velocity, direction: set}]
reproducibility: {mechanism: cycore, requires_asset: none}
proven_transfers: [fire]
hypothesized_transfers: [rain, snow, transition]
not_this: "噪声+阈值近似(弱替代,非真粒子)"
evidence: [samples/ColorfulFireBall]
confidence: observed
# ⚠Cycore 自带(人人能渲)但非 native、未 gate;plugin-free 无原生粒子替代。
```

### T9 final-grade(收尾调色)
统一整体氛围(对比/色调/暗角)。
```yaml
id: final-grade
role: [grade]
mechanism:
  - kind: effect
    any_of: [ADBE Brightness & Contrast 2, ADBE PhotoFilterPS, ADBE HUE SATURATION, ADBE Lumetri]
    signal: [{param: Contrast, direction: up}, {param: Saturation, direction: set}]
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: []
hypothesized_transfers: [fire, lightning, transition]
not_this: "单层局部调色(收尾=全局,常在调整层)"
evidence: [samples/bondbond]
confidence: observed
# 备注:Vignette 用 native 径向遮罩替代(CS Vignette=第三方)。
```

### T10 customizer-controller-rig(控制器装配)〔结构角色 · 本库 blocked〕
建控制器空层挂用户旋钮,表达式把各效果参数连到旋钮 → 一处调全联动。几乎所有商业模版的套路(=.mogrt 本质)。
```yaml
id: customizer-controller-rig
role: [control]
mechanism:
  - kind: composition_op
    op: "NewNullLayer + Color/Slider/Checkbox Control + 各效果参数挂表达式指向控制器"
reproducibility: {mechanism: lib-blocked, requires_asset: none}
proven_transfers: [lightning]
hypothesized_transfers: [transition, energy]
not_this: "把值烤死进各效果(那是放弃联动,不是控制器)"
evidence: [samples/LightningPack, incidents/expression-enable-byte-pair.md]
confidence: observed
# ⚠本库表达式 AE 不求值→做不出活联动 Customizer;只能烤死值。这是 lib-blocked 的典型。
```

### T11 drop-shadow-as-glow(投影当辉光)〔技巧〕
Drop Shadow 设 0 距离 + 亮色 + 大柔和 = 廉价方向/颜色可控的辉光 halo。
```yaml
id: drop-shadow-as-glow
role: [glow]
mechanism:
  - kind: effect
    any_of: [ADBE Drop Shadow]
    signal:
      - {param: Distance, direction: down}     # 0=纯 halo
      - {param: Softness, direction: up}
      - {param: Color, direction: set}
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [lightning]
hypothesized_transfers: [neon, text-fx]
not_this: "真投影(有距离=阴影不是辉光)"
evidence: [samples/LightningPack]
confidence: observed
# 常叠多个;比 Glow 更可控方向/颜色。
```

### T12 footage-recolor(素材重上色)〔素材+装配型〕
把白色/带 alpha 的素材重上色成任意颜色。
```yaml
id: footage-recolor
role: [color]
mechanism:
  - kind: effect
    any_of: [ADBE Fill, ADBE Tint]
    signal: [{param: Color, direction: set}]
reproducibility: {mechanism: native, requires_asset: footage}
proven_transfers: [lightning]
hypothesized_transfers: [light-fx, smoke-footage, particle-seq]
not_this: "给生成的 solid 上色(那走 luminance-color;此条专指素材重上色)"
evidence: [samples/LightningPack]
confidence: observed
```

### T13 mosaic-stylize(像素化风格)〔可选风格〕
把元素马赛克/像素化做"复古/数字"变体(常配开关切换)。
```yaml
id: mosaic-stylize
role: [texture]
mechanism:
  - kind: effect
    any_of: [ADBE Mosaic]
    signal: [{param: Horizontal Blocks, direction: set}, {param: Vertical Blocks, direction: set}]
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [lightning]
hypothesized_transfers: [transition, glitch]
not_this: "降分辨率(Mosaic 是块化风格,不是渲染质量)"
evidence: [samples/LightningPack]
confidence: observed
# 常配 Checkbox Control 开关切换变体。
```

### T14 fractal-branch(分形分支/电弧生成)〔程序化生成本体〕
**程序化生成**分叉电弧/闪电本体(不靠素材)。Lightning Pack 素材路线之外的纯生成路。
```yaml
id: fractal-branch
role: [form]
mechanism:
  - kind: effect
    any_of: [ADBE Lightning 2, ADBE Lightning]
    signal: [{param: Conductivity, direction: set}, {param: Core Radius, direction: set}, {param: Branching, direction: up}]
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: []
hypothesized_transfers: [lightning, electricity, cracks, neural-tree]
not_this: "用闪电素材(那是 footage-recolor 路线;此条是纯生成)"
evidence: [docs/capabilities (ADBE Lightning 2 在库)]
confidence: hypothesized
# ⚠可用但未验证(未 build/render 过)。配 T5 辉光 + T6 闪烁。
```

### T15 seamless-loop(时间无缝循环)〔结构技法 · ColorfulFireBall 催生〕
把一个**随时间演化**的源(噪声/粒子)做成可无限循环,无跳帧。火焰/烟/能量这类"持续翻腾"现象的隐形刚需。
```yaml
id: seamless-loop
role: [motion, organize]
mechanism:
  - kind: composition_op
    op: "演化源复制 2 份、时间错位半个周期 + 尾段 opacity 交叉淡入 + (可选)SilhouetteAlpha matte → 接缝隐形(ColorfulFireBall 'Noise 1 looped')"
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [fire]
hypothesized_transfers: [smoke, clouds, energy, water]
not_this: "单层直接循环(演化参数在端点不连续=可见跳帧)"
evidence: [samples/ColorfulFireBall]
confidence: observed
# 本库:time-remap 需关键帧(见 incidents/layer-settimeremapenabled-needs-keyframes)+ opacity 关键帧交叉淡入。未单独 build/render 验过。
```

### T16 rgb-channel-split(RGB 通道分离/色差)⭐glitch 催生
把图像拆成 R/G/B 三份各自偏移/闪烁 → 色差/通道错位。glitch 的招牌信号,也用于赛博朋克/复古 CRT/转场。
```yaml
id: rgb-channel-split
role: [color, distort]
mechanism:
  - kind: composition_op
    op: "源复制 3 份,各 ADBE Fill 成纯 R/G/B + 相加类 blend,三层各自位移/opacity 关键帧错位闪烁(Booyah 'RGBズレ':R/G/B 三层各 23-25kf opacity)"
  - kind: effect
    any_of: [ADBE Set Channels, ADBE Channel Blur]   # 单效果做通道操作/逐通道软化(GlitchText:Red Blurriness=80=软色差)
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [glitch]
hypothesized_transfers: [cyberpunk, retro-crt, transition, text-fx]
not_this: "整体 Hue 偏移(那是调色,不是把 R/G/B 拆开各自位移)"
evidence: [samples/motionbox/glitch/booyah-glitch, samples/motionbox/glitch/glitchtext]
confidence: observed
# 三层 Fill 路最可控(纯 native);Set Channels/Channel Blur 是单效果近似。配 time-evolution 让错位闪烁。
```

### T17 scanlines-crt(扫描线 / CRT 行)〔质感技法〕
横向行栅叠加 → CRT / HUD / 监视器质感。glitch、复古、全息常配。
```yaml
id: scanlines-crt
role: [texture]
mechanism:
  - kind: effect
    any_of: [ADBE Venetian Blinds, ADBE Grid]   # 百叶窗/网格生成等距横线,低 opacity 叠
    signal: [{param: Transition Completion, direction: set}, {param: Width, direction: down}]
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [glitch]
hypothesized_transfers: [retro-crt, hologram, hud, surveillance]
not_this: "整体降噪/模糊(扫描线是规律横纹叠加,不是噪点)"
evidence: [samples/motionbox/glitch/booyah-glitch]
confidence: observed
# Booyah 用 Venetian Blinds ×3 做 HUD 行栅。
```

### T18 temporal-glitch(时间 glitch / 抽帧卡顿 + datamosh)⭐glitch 催生
让时间轴本身 glitch:抽帧卡顿 + 帧间涂抹。区别于空间位移,作用在**时间**上。
```yaml
id: temporal-glitch
role: [motion, distort]
mechanism:
  - kind: effect
    any_of: [ADBE Posterize Time, ADBE Time Displacement]
    signal: [{param: Frame Rate, direction: down}, {param: Max Displacement Time [sec], direction: up}]
reproducibility: {mechanism: native, requires_asset: none}
proven_transfers: [glitch]
hypothesized_transfers: [datamosh, music-video, transition]
not_this: "空间位移(那是 displacement-distortion;此条是时间轴抽帧/涂抹)"
evidence: [samples/motionbox/glitch/glitchtext]
confidence: observed
# Posterize Time=降帧率出卡顿;Time Displacement=按亮度图错时间出涂抹(GlitchText Max Displacement Time=2s)。
# ⚠ GlitchText 招牌跳变靠第三方 Videocopilot Twitch;纯 native 用 Posterize Time + Displacement Map(块状噪声驱动)近似。
```

---

## 现象配方索引(技法的有序组合)

| 现象 | 配方 | 技法(id) |
|---|---|---|
| 火焰 | `checklists/techniques/build-good-fire.md` | ①程序化:noise-as-material → seamless-loop → displacement-distortion → luminance-color → silhouette-shape → additive-multilayer-depth(嵌套式 pass) → emissive-glow → time-evolution (+particle-emit/final-grade) |
| 闪电(素材包) | 实证#2 Lightning Pack | ②素材+装配:footage-recolor + drop-shadow-as-glow + emissive-glow + mosaic-stylize + customizer-controller-rig。**电弧本体=外部素材** |
| 闪电(程序化) | (待建,可行) | ①程序化:fractal-branch(ADBE Lightning 2) + emissive-glow + time-evolution。纯生成不靠素材 |
| 风 | (待建) | noise-as-material + displacement-distortion(方向) + time-evolution + 运动模糊 |
| 雨 | (待建) | particle-emit(条状) + time-evolution(下落) + 模糊 |
| 转场 | (待建) | displacement-distortion/擦除 + time-evolution(时间扫过) |
| glitch(纯 native·可复刻) | (待建,Booyah 实证可行) | ①程序化:rgb-channel-split + displacement-distortion(块状噪声驱动) + scanlines-crt + temporal-glitch(Posterize Time) + emissive-glow。Booyah Glitch 全 native |
| glitch(重度/datamosh) | 实证 GlitchText | ②插件依赖:招牌跳变=**Videocopilot Twitch**(第三方)+ PEDG/Colorama 等;native 部分=temporal-glitch + rgb-channel-split + displacement。纯 native 只能近似 |

> 新增现象:`aepdissect` 解析 → 拆角色 → 按 schema 在此登记新技法/标已有技法新 `proven_transfers` → 写 `checklists/build-<现象>.md` 配方(负责技法间顺序)→ AE gate 验证升 confidence。

---

## 重构发现(反馈给 spec,2026-06-18)

迁 T1–T14 时撞到一个 schema v2 没覆盖的张力,已按"先用着"原则定规(2026-06-18 fire 重解析又增 T15 seamless-loop):
- **any_of 内可复刻性不齐**(T2 = Turbulent Displace[native] + PEDX[third-party];T9 含 native+Lumetri):**技法级 `reproducibility.mechanism` 取 any_of 里最可达的**(有一个 native 即技法可复刻),更强但需插件的实现记备注。语义=回答"这技法能否被复刻"而非"每种实现各自如何"。若将来需要 per-effect 可复刻性再下沉到 mechanism 项(spec §8 候选)。
