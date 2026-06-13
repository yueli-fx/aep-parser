# showcase/ — 从零产出示例总览

每个子文件夹 = 一个**能力方向**的纯 Go 从零生成示例 + AE 实渲产物，供用户逐项审核。规则见 `rules.md` § Showcase + `checklists/showcase.md`。

**产物 gitignored**：`*.aep` / `*.png` 不进仓库；`INDEX.md` + `gen.go` + `render.jsx` tracked。干净 clone 后重生成：

```
go run ./flightdeck/showcase/<方向>            # 构建 <方向>.aep
# 再用 scripts/ae_run.ps1 跑 <方向>/render.jsx 出 <方向>.png
```

## 状态说明（review-gate）

**🔍 待review** = agent 已建 + AE 实渲 + 自己眼验，**等用户在真机打开 .aep 复核**；**✅ complete** = **用户真机验收过**。agent 不自标 complete（详 `rules.md` § Showcase）。

下面 15 个可视方向**全部 ✅ complete**（2026-06-14 用户真机逐个验收通过）。📋 B 类读值档（render-queue / essential-graphics / markers / comp-settings / project-settings / camera-light / media-replace）补全中。

## 方向一览

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

> **15 个 🖼 可视方向全部 ✅ complete**（用户真机验收）：8 from-scratch 可视 + effects + masks + structural-ops + transform-values + stroke-detail + keyframe-channels + animated-path。**📋 B 类读值档补全中**（render-queue / essential-graphics / markers / comp-settings / project-settings / camera-light / media-replace，dump 值 readback 不看图）。全量清单 + 验证档位（🖼看图 vs 📋读值）详 `specs/2026-06-13-full-showcase-coverage.md`。
