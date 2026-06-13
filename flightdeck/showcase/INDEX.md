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

> 后续大阶段（新 layer 类型组 / expressions / precomp / gradient / text …）落地时按 `checklists/showcase.md` 逐方向补。
