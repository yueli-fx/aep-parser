# showcase/ — 从零产出示例总览

每个子文件夹 = 一个**能力方向**的纯 Go 从零生成示例 + AE 实渲产物，供用户逐项审核。规则见 `rules.md` § Showcase + `checklists/showcase.md`。

**产物 gitignored**：`*.aep` / `*.png` 不进仓库；`INDEX.md` + `gen.go` + `render.jsx` tracked。干净 clone 后重生成：

```
go run ./flightdeck/showcase/<方向>            # 构建 <方向>.aep
# 再用 scripts/ae_run.ps1 跑 <方向>/render.jsx 出 <方向>.png
```

## 方向一览

| 方向 | 测什么 | 状态 |
|---|---|---|
| [shape-filters](shape-filters/INDEX.md) | 形状矢量滤镜家族 11 件 + PolyStar + 双滤镜叠加 | ✅ complete |
| [shape-primitives](shape-primitives/INDEX.md) | 四种参数图元 + Fill/Stroke/Gradient 描绘变体 | ✅ complete |
| [keyframes-ease](keyframes-ease/INDEX.md) | 时间缓动 linear/ease-out/ease-in-out（渲中间帧看位置差） | ✅ complete |
| [expressions](expressions/INDEX.md) | 表达式激活（time\*N 旋转，渲 t=1s 看角度） | ✅ complete |
| [precomp-nesting](precomp-nesting/INDEX.md) | 预合成嵌套（parent 嵌套 child 组合场景） | ✅ complete |
| gradient | 渐变填充/描边 + ramp 方向 | ⏳ 待建 |
| text | 文字层 from-scratch | ⏳ 待建 |
| layers | Solid/Null/Adjustment/Camera/Light/Precomp 建层 | ⏳ 待建 |

> 后续大阶段落地时按 `checklists/showcase.md` 逐方向补。⏳ = 本轮回填中。
