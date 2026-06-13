---
status: active
summary: 把所有已 ship 能力补齐 showcase 供用户逐个真机验收（代码过≠效果对）；已回填 8 个 from-scratch 可视方向，列出剩余缺口 + 非可视能力的验证方式，交接下个对话逐方向补
last_updated: 2026-06-13
---

# 全能力 showcase 覆盖 — 每个已 ship 能力都要有用户可真机验证的 showcase

## 目标（一句话）

**库里每个已 ship 的能力，都要有一个用户能在真机打开 .aep 逐个验证的 showcase**。理由：**「代码过 ≠ 效果对」**（交付准则红线4 + 用户原话）——ship-gate 像素验证是 agent 跑的，用户要亲自开 .aep 核一遍才算数。规约见 `rules.md` § Showcase + `checklists/showcase.md`（含 `待review`↔`complete` review-gate）。

## 现状：已回填 9 方向（全部 🔍 待review，等用户真机验）

`flightdeck/showcase/` 下：shape-filters · shape-primitives · keyframes-ease · expressions · precomp-nesting · gradient · text · layers · **effects**。每个有 `INDEX.md`（布局表）+ `gen.go` + `render.jsx`，产物 `.aep/.png` 已本地生成（gitignore）。**这 9 个等用户逐个真机验收后翻 complete。**

> 前 8 方向：2026-06-13 本对话已全部 agent 实渲眼验通过（含 keyframes-ease / shape-filters / layers 三个**此前缺 png = 从未实渲**的方向，已补齐 AE 渲染、AE 接受、眼验对）。

> **effects（2026-06-13 补，commit d193f38）**：4×3 网格 = 源 token（teal 方块+amber 边）× 12 效果，AE 2020 实渲 **12/12 可见正确**（GaussianBlur/DropShadow/Invert/Tint/Tritone/WaveWarp/Brightness/FractalNoise/GradientRamp/Mosaic/DirectionalBlur）。**红线4d 活体样本**：HueSaturation master hue（`-0004` angle）`SetEffectParam` 物化值后 Go round-trip 绿、**AE frame 0 渲染色相未变** = 疑似假绿（或通道控制前置/编码未对，待 RE 确认），已剔除换 Wave Warp；Mosaic 块数 control-type 1 无 generic 模板调不了。**含义**：coverage.md「SetEffectParam 任意效果任意参数即设即用」对**未单独 gate 的参数**需打折——只有经 ship-gate 的参数确证被 AE 引擎应用。

> **看图档批次（2026-06-13，并行 4 subagent 设计+编码、主控串行渲染眼验）**：
> - ✅ **masks**（commit）— AddMask 圆/三角/星/inverted 4/4 AE 渲染对（修了圆 bezier 切线 out/in 反向 → 尖角 bug）。
> - ✅ **transform-values**（commit）— SetPosition/Scale/Rotation/Opacity 6/6 渲染对（scale/opacity=百分比单位）。
> - ✅ **structural-ops**（commit）— Duplicate/Move/Delete/Separate；**solid 不可摆位硬限制**（Position 未物化、merged 无 leader → AE 堆 solid 到 comp 中心）→ 同心环方案：Move/Delete 像素可见、Dup/Separate 靠 .done readback。
> - ✅ **stroke-detail**（commit，闭合形状方案）— 第一版 open-path 全塌缩（假绿边界，见下），改**全闭合形状**后 8 zone AE 实渲眼验对：Dashes（闭合 rect 虚线）+ Line Join Miter/Round/Bevel（闭合五角星尖角三态）+ Miter Limit 高/低（极尖星，长刺 vs 削平）+ Wave Fine/Bold（星轮廓密/疏波纹）。**诚实暂缺 2 项**：**Line Cap**（只在 open path 端点出现，open path 塌缩 → 无可视）·**Taper**（闭合环无起止 → AE 渲等宽轮廓，值 round-trip 但无视觉，已实渲确认）。**关键发现**：open-path（`AddPath`+`SetClosed(false)`）+stroke 渲染塌缩——几何 round-trip 绿、AE 渲微小图形；三个 stroke ship-gate 全建闭合 rect、只验值不验渲染像素、从不用 open path → 唯一渲染被证实的 stroke 几何 = 闭合形状。
>
> - ✅ **keyframe-channels**（commit）— 四 transform 通道各两线性关键帧 Position/Scale/Rotation/Opacity，渲 t=2s 中间帧每通道落插值（x≈960 / scale87.5 / rot90 横条 / opac55 暗粉），像素 + `.done` `valueAtTime(2.0)` readback 双证。与 keyframes-ease 区别：那个变位置缓动曲线，这个横扫四通道。每通道仅 2 kf 规避 lhd3 >4-kf 容量坑。
>
> - ✅ **animated-path**（commit）— 闭合 Path 两线性关键帧，几何 morph 横条(440×120)→正方(280×280)→竖条(120×440)，渲 t=2s 正方形 + `.done` 三时刻 extent readback 数证顶点逐个插值。闭合+实心 fill 规避 open-path 塌缩；仅 2 kf 规避 lhd3 >4-kf 容量坑（多帧路径前必修 `encodePathTimeTable` 同病）。
>
> **🖼 看图档（A 类可视方向）已全部补齐**（共 15 个可视方向，用户 2026-06-14 真机验收全 complete）。
>
> **📋 B 类读值档（2026-06-14 用户「全补 7 个」）**：4 个 from-scratch 可行**已补 🔍 待review** + 3 个 from-scratch **不可表达**（已注明）：
> - ✅ **comp-settings** — motionBlur/samples/adaptiveLimit/bgColor/workArea/hideShy/nestedFrameRate 7 项 DOM 一致。**边界**：SetShutterAngle/Phase 读回 ×≈1.2、SetResolutionFactor AE 除零（排除）。
> - ✅ **project-settings** — bitsPerChannel/linearBlending/expressionEngine/footageTimecodeDisplayStartType 4 项一致。**边界**：SetTimeDisplayType/FeetFramesFilmType（nnhd byte8 疑位打包）/FramesCountType 不反映（排除）。
> - ✅ **camera-light** — NewCameraLayer/NewLightLayer 建层 + 类型确认（CameraLayer/LightLayer，POINT）。**边界**：选项 setter from-scratch 全 elide（"property not present"），只在 parsed 层生效。
> - 🛑 **essential-graphics — BLOCKED**（用户 2026-06-14 两次真机确认）：从零 EG 工程展开「基本图形」面板**崩溃 AE**——3 混合控件崩,**退回单 slider 也崩**。DOM readback 全过(模板名+控件名)=假绿,根源 EG ship-gate（load+DOM+resave）**从不打开面板**。整个 from-scratch EG 不可交付,留 RE repro,用户决定暂不修(EG 用得少)。
> - ❌ **markers / render-queue / media-replace** — from-scratch 不可表达：AddMarker/AddItem 需 canonical seed（空集无模板克隆）、SetAlternateSource 需真实 footage。各有 fixture-based ship-gate，非 from-scratch showcase。
>
> **B 类 readback 系统性发现（红线4a）**：多个「仅字节 round-trip、从未 AE-DOM 验证」的 settings setter，AE-DOM 核出字节写对但 AE 不反映（shutter/resolution/nnhd-byte8）→ 建议独立 RE/修。

## 缺口：还没 showcase 的已 ship 能力

> 权威清单以 `plans/coverage.md` 为准（看板可能漂移，建时用 grep/Explore 核实代码 + ship-gate test 真在）。下面是分组待办，**逐方向补、每个落 `待review`**。

> **验证档位（2026-06-13 用户澄清「有些是不是代码过就算过、不需要验证」）**：**没有「纯 Go 代码过就算过」**——最低门槛永远是 AE 接受 + AE 读回（红线4a：Go round-trip ≠ AE 接受）。但验证**深度**按能力**有无可见作用面**分两档：**A 类（🖼 看图档，渲染会变）必须像素验证 + 用户真机看图**，是假绿高发区（shape 颜色 / effects 的 HueSaturation hue 都栽在这）；**B 类（📋 读值档，渲染不变的纯数据字段）= AE 接受 + readback 读回值对，用户读值核对不看图**（无「值对但渲染错」陷阱，可信度本就高，ship-gate 已验 readback 的甚至可攒批/可信任）。下面 A/B 分组即此二分。

### A. 可像素验证（同既有 from-scratch + AE 渲染套路）— 🖼 看图档

| 方向 | 覆盖能力 | 渲染思路 |
|---|---|---|
| `masks` | AddMask（mask 形状裁切/显隐图层） | 一个填充层 + mask → 渲染只露 mask 内区域 |
| ~~`effects`~~ ✅ | AddEffect + SetEffectParam | **已补 2026-06-13（d193f38），🔍 待review**（12 效果网格；HueSaturation hue 假绿剔除） |
| `transform-values` | 改字段值：transform（位置/缩放/旋转）+ 颜色 before/after | 读/建工程 → 改值 → before/after 并排渲染 |
| `keyframe-channels` | 各属性关键帧（扩展现有 keyframes-ease） | 渲中间帧看各通道插值 |
| `structural-ops` | DeleteLayer / DuplicateLayer / MoveLayer / 属性 Remove·Duplicate·MoveTo / SetDimensionsSeparated | 建基线 → 应用 op → 渲染出结果（如 duplicate→两份、separate→XY 分离动画） |
| `stroke-detail` | 虚线 dashes / Line Cap / Line Join / Miter | 几条不同端点·连接·虚线的描边并排（可并入 shape-primitives 或单列） |
| `animated-path` | animated shape path / mask path 关键帧 | 渲中间帧看路径插值形态 |

### B. 非可视 / 结构性（不出像素，改用 readback 验证产物）

这些能力**没有可渲染的视觉**（EG 面板控件 / 标记 / 渲染队列 / 相机灯光参数 / 媒体替换 / 合成设置等）。其 showcase = `gen.go` 建工程 + `verify.jsx` 把相关 DOM 值 dump 到 `.done` 日志，**用户读日志核值**（而非看图）。INDEX 的「产物」表注明是 readback 验证、不是像素。

| 方向 | 覆盖能力 | readback 验证 |
|---|---|---|
| `essential-graphics` | AddEssentialProperty / SetMotionGraphicsTemplateName | dump EG 控件名/类型/值 |
| `markers` | layer / comp markers | dump marker time/comment |
| `render-queue` | RQ add / insert | dump RQ item 设置 |
| `camera-light` | NewCameraLayer / NewLightLayer + 选项 | dump 层类型 + camera/light 参数（需 3D 才有像素，暂 readback） |
| `comp-settings` | shutter angle/phase 等合成设置 | dump cdta 字段 |
| `media-replace` | SetAlternateSource | 需真实 footage；dump source 引用（或暂搁） |

> B 类是否值得逐个出，按需——用户点名要验哪个就补哪个；纯 round-trip 的可以攒批。

## 做法（交接下个对话）

1. 一次一个方向（A 类优先，视觉最直观）：照 `checklists/showcase.md` 6+1 步——`gen.go`→`go run`→`render.jsx`→AE 实渲眼验→写 INDEX(`待review`)→commit→**通知用户真机复核**。
2. **agent 不自标 complete**；用户真机验过才翻。
3. 每补一个，顶层 `showcase/INDEX.md` 加一行。
4. 全部补齐（A 类 + 用户点名的 B 类）+ 用户逐个验收 complete = 本 spec done。

## 验收口径

每个方向 done 的标准 = **用户在真机打开该方向 .aep、对照 INDEX 布局表确认效果对** → 翻 `complete`。不是 agent 渲染眼验，不是 ship-gate 绿。
