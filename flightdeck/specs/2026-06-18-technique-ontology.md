---
status: active
graduate: true
summary: 投大规模工程前冻结的数据骨架:角色/技法/机制三轴本体 + 技法条目 schema v2(等价效果集·可复刻性两轴·迁移分proven/hypothesized)+ 工程画像 schema + 受控词表 v0。schema 闭 instance 开,约束所有工程理解的产出形式,使跨工程可聚合出通用技巧。经火焰(生成型)/闪电(素材装配型)/控制器(lib-blocked)三类压测。
last_updated: 2026-06-18
---

# 技法本体 + 工程画像 schema(三轴受控词表)

> **为什么要这份(冻结契约的理由)**:目标是喂大量参考工程 → 聚合出**通用技巧**。通用技巧只在**跨工程对齐**时浮现(单工程理解到顶只是"这一个怎么做的")。所以每个工程的理解产出**必须用同一套受控结构**——不能自由文本,否则 N 份理解汇不成通用技巧、只是 N 份散记。**这份 schema 必须在投大量工程之前冻结**:schema 错了,已积累的数据全要返工。
>
> **冻结的边界(schema 闭 / instance 开)**:**冻结**=三轴结构 + 字段格式 + 受控取值的"轴";**放开**=具体有哪些角色/技法/机制条目(随工程增长,按固定格式追加,不返工)。"先用着"= v2 即刻启用,取值集合持续生长。
>
> 上游:`specs/2026-06-18-fx-technique-internalization.md`(内化流水线;本 spec 是其"理解产出"的数据契约)。下游被重构对象:`docs/fx-techniques.md`(现单层 T1–T14,需按三轴重组)。

## 1. 三轴本体(核心)

理解不是一张表,是三个**不同抽象层** + 它们之间的**多对多映射**:

| 轴 | 回答 | 粒度 | 谁填 |
|---|---|---|---|
| **角色 Role** | 解决**什么问题** | 最粗(几十个) | 人/AI 归纳 |
| **技法 Technique** | 用**什么手法**解决 | 中(几百个) | 人/AI 归纳 |
| **机制 Mechanism** | 落到**哪些效果/操作** | 最细(matchName 级) | **自动提取** |

关系链 **角色 ← 技法 ← 机制**,多对多(一角色多技法、一技法多角色、一技法多机制)。

**粒度规则**:聚合通用技巧发生在**技法层**(机制太细聚不出模式,角色太粗都一样)。技法粒度判据 = **"一个可迁移的因果决策"**——能从一个现象搬到另一个现象的最小知识单元。
- ✓ "竖拉+高对比噪声造瘦高料子"(可迁移到烟/能量)
- ✗ "对比度设 185"(是参数 = 机制层的 signal 值)
- ✗ "辉光"(是角色)

## 2. 技法条目 schema v2(冻结)

```yaml
technique:
  id: <kebab-stable-id>                 # 稳定 ID,合并/拆分是一等操作
  role: [<受控:见 §4>]                  # 多值可;视觉角色与结构角色分组
  mechanism:                            # 列表,每项 effect 或 composition_op
    - kind: effect | composition_op
      any_of: [<matchName>, ...]        # 等价效果集(同技法的不同实现)
      op: <text>                        # 仅 composition_op:图层级操作描述
      signal:                           # 非默认参数 = 配方信号
        - param: <name>
          direction: up | down | set | toggle
          observed_range: [lo, hi]      # 机制层证据,跨工程累积
  reproducibility:                      # 两条正交轴
    mechanism: native | cycore | third-party | lib-blocked
    requires_asset: none | footage | image-seq | audio
  proven_transfers: [<现象>, ...]       # 真在该现象渲过的(实证)
  hypothesized_transfers: [<现象>, ...] # 推断待验
  not_this: "<反例,防粒度漂移>"
  evidence: [<showcase/sample/incident 路径>]
  confidence: validated | observed | hypothesized
```

**字段为什么这么定(压测得出的设计决策,= 约束理由)**:
- **`mechanism.any_of`(等价效果集)**:同一技法不同作者用不同效果(Fractal Noise / Turbulent Noise / CC Noise)。若绑具体效果,同技法被拆成 N 条 → 跨工程聚合直接失败。**这是对齐的命门。**
- **`kind: effect | composition_op`**:有些技法没有"效果",是图层结构(复制层+Add+不同 scale)。机制层不止效果,还有工程操作。
- **`signal.direction` 含 set/toggle**:颜色是"设成某值"、开关是 toggle,不只标量↑↓。
- **`signal` 拆 direction(技法层定性)+ observed_range(机制层证据)**:方向跨工程稳定,具体值是参数空间、每工程不同,分开存。
- **`reproducibility` 两正交轴**:机制可复刻性(能不能做)与素材依赖(要不要外部资产)**独立**——Fill 机制全可复刻,却依赖你造不出的闪电素材。漏了这维度,②素材+装配型工程无法诚实分类。`lib-blocked` 是本库特有状态(AE native 但本库表达式不求值,做不出活联动),区别于"AE 能不能做"。
  - **`third-party` = 可支持,不是不可做**(2026-06-18 用户定调):读已 opaque round-trip(红线5);从零写靠 embed-template(从真实样本采该效果 chunk → splice),**装了插件才渲对**——尚未实现但可行。招牌插件常**就是技法本身**(Twitch=glitch),不该跳过。`plugin-free` 仅是 procedural-fx-generator 产品的优先项,非学习库的准入门槛。真正做不了的只有 `lib-blocked`。流程细节见 `checklists/techniques/understand-a-project.md` §第三方插件 + §归属判据。
- **迁移拆 `proven` / `hypothesized`**:火焰没做过烟,`transfers:[smoke]` 是猜测不是事实;混在一起会让"通用技巧"虚高。诚实是地基(红线4 精神)。
- **顺序/依赖不在技法里**:技法是**无序原子**,技法间的编排顺序(displacement 必在 noise 之后)归**现象配方层**(`checklists/build-<现象>.md`)。职责分离。

## 3. 工程画像 schema(每工程一份,理解的产出形态)

理解的产出不是 dump 信息,是一份**可跨工程对齐**的画像(给学习引擎抽技法 / 给生成引擎当参考 / 给检索做指纹):

```yaml
project_profile:
  meta:      { type: [reel|logo|transition|...], resolution, duration, fps }
  fingerprint: { layers, comps, nest_depth, reuse_ratio }
  signal_layers:        # 降噪后剩的 ~5% 层(见 §5)
    - { name, role, non_default_params, expressions }
  timeline:  { active_intervals, motion_techniques: [ease|stagger|overshoot|hold] }
  graph:     { edges: [source|parent|expression|matte] }   # 核心依赖边
  reproducibility: { native_pct, cycore_pct, third_party_pct, asset_deps }
  techniques: [<technique.id>, ...]    # 用到的技法(对齐单位)
  verification: { reproduce_render_diff }   # 做了复现才有
```

## 4. 受控词表 v0(先用着,会生长)

**角色 Role**(现象无关,判据:必须跨现象成立;否则它是现象配方不是角色):
- **视觉角色**:`form`(形态) `contour`(轮廓/裁形) `distort`(扭曲) `motion`(运动) `color`(上色) `temperature`(色温) `depth`(深度) `glow`(辉光) `matte`(遮罩) `texture`(质感) `particle`(粒子) `grade`(收尾调色) `transition`(转场)
- **结构角色**(不产生像素,组织手法):`control`(控制器/参数化) `parametrize` `organize`(元件/复用)

**现象 Phenomenon**(transfers 用):`fire` `smoke` `lightning` `rain` `wind` `energy` `clouds` `transition` `water` `magic` `light-fx`…(开放)

**机制可复刻性**:`native` `cycore` `third-party` `lib-blocked`
**素材依赖**:`none` `footage` `image-seq` `audio`

## 5. 理解的核心动作 = 降噪(配套原则)

一个工程 95% 是默认/样板/不可见。理解 = 找那 5% 信号层/信号参数,**不是读全**:
- **AE 的 elision 免费降噪**:.aep 只存非默认参数(实测:Glo2 14 参数只存作者调的 3 个 + 结构槽;Lumetri 122 存 45)。字节本身≈作者改过的(~95% 强信号,例外:结构槽恒在、synthesis 物化、改回默认;用字典 default 交叉核对滤掉)。
- 其余降噪:丢不可见/0 不透明/被 matte 全遮层、重复元件看一次、静态属性=背景(有关键帧的=主角)。

## 6. 理解度判据(金标准 = 能复现)

能翻译(语义)< 能说角色(角色)< **能从画像反向生成等价工程、渲染对照接近原图**(复现)。复现 render-diff 是**可量化的理解度指标**。推论:**理解引擎与生成引擎是一体两面**(对工程的理解 = 生成它的程序)——`internalization` 与 `procedural-fx-generator` 两 spec 由此焊死。

## 7. 已填范例(怎么填的范本,真实数据)

```yaml
- id: noise-as-material            # 生成型
  role: [form]
  mechanism: [{kind: effect, any_of: [ADBE Fractal Noise, ADBE Turbulent Noise],
              signal: [{param: Contrast, direction: up, observed_range: [128,185]},
                       {param: ScaleHeight, direction: up}]}]
  reproducibility: {mechanism: native, requires_asset: none}
  proven_transfers: [fire]
  hypothesized_transfers: [smoke, energy, clouds]
  not_this: "贴静态噪声纹理当背景"
  evidence: [showcase/procedural-fx]
  confidence: validated

- id: additive-multilayer-depth    # 结构型(无效果,是图层操作)
  role: [depth]
  mechanism: [{kind: composition_op,
              op: "同套效果链复制 N 份 + 各层不同 noise scale + blend=Add"}]
  reproducibility: {mechanism: native, requires_asset: none}
  proven_transfers: [fire]
  not_this: "单层调高对比假装有层次(v2 被否)"
  evidence: [showcase/procedural-fx]
  confidence: validated

- id: footage-recolor              # 素材+装配型(机制可复刻,但依赖素材)
  role: [color]
  mechanism: [{kind: effect, any_of: [ADBE Fill],
              signal: [{param: Color, direction: set}]}]
  reproducibility: {mechanism: native, requires_asset: footage}
  proven_transfers: [lightning]
  confidence: observed

- id: customizer-controller-rig    # 结构角色 + 本库 blocked
  role: [control]
  mechanism: [{kind: composition_op,
              op: "null + Color/Slider/Checkbox Control + 各效果参数挂表达式指向它"}]
  reproducibility: {mechanism: lib-blocked, requires_asset: none}
  proven_transfers: [lightning]
  not_this: "把值烤死进各效果(那是放弃联动,不是控制器)"
  confidence: observed

- id: emissive-glow                # 跨现象复用实证:fire + lightning 都 proven
  role: [glow]
  mechanism: [{kind: effect, any_of: [ADBE Glo2],
              signal: [{param: Glow Threshold, direction: down},
                       {param: Glow Radius, direction: up}]}]
  reproducibility: {mechanism: native, requires_asset: none}
  proven_transfers: [fire, lightning]
  confidence: validated
```

## 8. 待定 / 会生长(非冻结部分)

- 角色词表取值会随工程增长(尤其结构角色,样本少);粒度判据需持续校准。
- 技法粒度漂移防治:新增前检索去重、合并/拆分常态化、`not_this` 反例必填。
- `docs/fx-techniques.md` 按本三轴重构(当前单层混了角色/技法)= 第一个落地动作。
- 工程画像的自动生成(aepdissect 演进:消费字典标非默认 → 时间轴 → 依赖图边 → 复现验证)。

## 9. 验证状态

- schema v2 经三类极端压测扛住:生成型(火焰)、素材+装配型(闪电)、lib-blocked(控制器)。
- 机制层地基实测可靠(elision 免费降噪,§5)。
- **未验证**:大规模(N≫3)下角色/技法词表是否仍可对齐、复现判据(§6)是否可操作——投工程时校准。
