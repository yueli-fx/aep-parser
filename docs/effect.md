# Effect object

`layer.Effects[index]`

## Description

应用到图层的特效实例（如 Gaussian Blur、Tritone、Curves）。每个 Effect 有 MatchName（标识类型）和 Parameters（特效内部参数，每个都是 `Property`）。

特效参数 matchname 是 `<EffectMatchName>-NNNN` 形式（如 `"ADBE Gaussian Blur 2-0001"` = Blurriness），NNNN 是 effect 模板里参数的内部编号。

## Example

```go
for _, fx := range layer.Effects {
    fmt.Printf("Effect: %s\n", fx.MatchName)
    for _, p := range fx.Parameters {
        if len(p.Keyframes) > 0 {
            fmt.Printf("  %s = animated (%d keyframes)\n", p.MatchName, len(p.Keyframes))
        } else {
            fmt.Printf("  %s = %v\n", p.MatchName, p.StaticValue)
        }
    }
}
```

---

## Attributes

### Effect.MatchName

```go
MatchName string
```

特效类型标识 —— 如 `"ADBE Gaussian Blur 2"`、`"ADBE Gradient Wipe"`、`"ADBE Tritone"`。AE 内部使用，跨 AE 版本稳定。read-only。

---

### Effect.Name

```go
Name string
```

特效显示名。当前镜像 `MatchName`（用户在 AE 里手动改的 effect-instance display name 未独立解出来）。read-only。

---

### Effect.Parameters

```go
Parameters []*Property
```

#### Description

特效内部参数列表。每个参数本身是个 `Property`，所以可以：

- 读 `StaticValue` 或 `Keyframes`
- 用 `SetStaticValue` 改静态值
- 用 `Keyframe.SetValue` / `SetTime` 改单帧

```go
// 修改 Gaussian Blur 的 Blurriness 到 20
for _, fx := range layer.Effects {
    if fx.MatchName == "ADBE Gaussian Blur 2" {
        for _, p := range fx.Parameters {
            if p.MatchName == "ADBE Gaussian Blur 2-0001" && len(p.Keyframes) == 0 {
                p.SetStaticValue(20.0)
            }
        }
    }
}
```

#### Type

`[]*Property`；read（每个 Property 内部可写）。详见 [property.md](property.md)。

---

## 当前已测覆盖

| 参数类型 | 状态 |
|---|---|
| 1D 静态 / 缓动（如 Gaussian Blur Blurriness） | ✓ 测试覆盖 |
| 4D 颜色缓动（如 Tritone Highlights） | ✓ 测试覆盖 |
| 其他参数（菜单 enum / 复选框 / 点选项等） | 走通用 Property 路径，理论 OK，未系统验证 |

如果遇到奇怪的 effect 参数解码异常，请提供 .aep 样本和期望值。
