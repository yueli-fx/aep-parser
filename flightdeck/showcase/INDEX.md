# showcase/ — 从零产出示例总览

每个子文件夹 = 一个**能力方向**的纯 Go 从零生成示例 + AE 实渲产物，供用户逐项审核。规则见 `rules.md` § Showcase + `checklists/showcase.md`。

**产物 gitignored**：`*.aep` / `*.png` 不进仓库；`INDEX.md` + `gen.go` + `render.jsx` tracked。干净 clone 后重生成：

```
go run ./flightdeck/showcase/<方向>            # 构建 <方向>.aep
# 再用 scripts/ae_run.ps1 跑 <方向>/render.jsx 出 <方向>.png
```

## 状态说明（review-gate）

**🔍 待review** = agent 已建 + AE 实渲 + 自己眼验，**等用户在真机打开 .aep 复核**；**✅ complete** = **用户真机验收过**。agent 不自标 complete（详 `rules.md` § Showcase）。

下面 15 个 🖼 可视方向**全部 ✅ complete**（2026-06-14 用户真机逐个验收通过）。📋 B 类读值档：4 个 from-scratch 可行已补（🔍 待review），3 个 from-scratch 不可表达（见末节）。

## 🖼 A 类 — 可视方向（看图档）

| 方向 | 测什么 | 状态 |
|---|---|---|
| [shape-filters](shape-filters/INDEX.md) | 形状矢量滤镜家族 11 件 + PolyStar + 双滤镜叠加 | ✅ complete |
| [shape-primitives](shape-primitives/INDEX.md) | 四种参数图元 + Fill/Stroke/Gradient 描绘变体 | ✅ complete |
| [keyframes-ease](keyframes-ease/INDEX.md) | 时间缓动 linear/ease-out/ease-in-out（渲中间帧看位置差） | ✅ complete |
| [expressions](expressions/INDEX.md) | 表达式激活（time\*N 旋转，渲 t=1s 看角度） | ✅ complete |
| [precomp-nesting](precomp-nesting/INDEX.md) | 预合成嵌套（parent 嵌套 child 组合场景） | ✅ complete |
| [gradient](gradient/INDEX.md) | 渐变填充 + ramp 方向（横/纵/对角/三停） | ✅ complete |
| [text](text/INDEX.md) | 文字层 from-scratch（NewTextLayer + SetText 多行） | ✅ complete |
| [layers](layers/INDEX.md) | Solid 同心色框 + Null/Adjustment 建层 | ✅ complete |
| [effects](effects/INDEX.md) | 加效果 AddEffect + 改参数 SetEffectParam（12 效果网格） | ✅ complete |
| [masks](masks/INDEX.md) | AddMask 遮罩裁切（圆/三角/星/inverted） | ✅ complete |
| [structural-ops](structural-ops/INDEX.md) | Duplicate/Move/Delete/维度分离（同心环+readback） | ✅ complete |
| [transform-values](transform-values/INDEX.md) | 改 Position/Scale/Rotation/Opacity（before/after） | ✅ complete |
| [stroke-detail](stroke-detail/INDEX.md) | 描边细节 Dashes/Line Join/Miter/Wave（闭合星，Cap·Taper 诚实暂缺） | ✅ complete |
| [keyframe-channels](keyframe-channels/INDEX.md) | 四通道关键帧 Position/Scale/Rotation/Opacity（渲 t=2s 插值 + readback） | ✅ complete |
| [animated-path](animated-path/INDEX.md) | 动画路径几何 morph 横条→正方→竖条（渲 t=2s + extent readback） | ✅ complete |

## 📋 B 类 — 读值档（dump 值核对，不看图）

无渲染视觉的能力 = `gen.go` 建工程 + `verify.jsx` dump DOM 值到 `.done`,**用户读日志核值**（非看图）。

| 方向 | 测什么 | 状态 |
|---|---|---|
| [comp-settings](comp-settings/INDEX.md) | 合成设置 motionBlur/workArea/bgColor/hideShy/nestedFrameRate（7 项 DOM 一致） | ✅ complete |
| [project-settings](project-settings/INDEX.md) | 工程设置 bitsPerChannel/linearBlending/expressionEngine/footageTimecode（4 项一致） | ✅ complete |
| [camera-light](camera-light/INDEX.md) | NewCameraLayer/NewLightLayer 建层（**灯光=环境光** + 相机=双节点,值皆模板默认；选项 setter elide） | ✅ complete |
| [essential-graphics](essential-graphics/INDEX.md) | AddEssentialProperty + 模板命名 | 🛑 **BLOCKED**（面板崩溃，连单 slider 也崩；RE 候选，暂不修） |

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
- **camera/light 选项 setter**：from-scratch 全 elide（"property not present"）,只在 parsed 层生效；**NewLightLayer 默认=环境光**(无从零 SetLightType)。

详各方向 INDEX「已发现边界/崩溃发现」。**建议独立 RE/修**。

---

> **15 个 🖼 可视方向 ✅ complete** · **3 个 📋 读值档 ✅ complete**(comp-settings / project-settings / camera-light) · **EG 🛑 BLOCKED**(面板崩溃,RE 候选) · **3 个 from-scratch 不可表达（markers/render-queue/media-replace，已注明）**。showcase 覆盖收官。全量清单 + 验证档位详 `specs/2026-06-13-full-showcase-coverage.md`。
