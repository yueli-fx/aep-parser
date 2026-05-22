# ShapePath object

`layer.ShapePaths[index]`（仅当 `layer.IsShapeLayer == true`）

## Description

Shape Layer 上用 Pen 工具自由绘制的 Bezier 路径。**只有自由路径会产生 `ShapePath` 条目**。

参数化基元（Rect / Star / Ellipse 等带 Size、Rotation、Inner Radius 等参数的内置形状）出现在 `Layer.ShapePrimitives` —— 见下面 [ShapePrimitive object](#shapeprimitive-object)。

底层与 Mask 共享 `shap` chunk 结构，区分方式：是否在 `"ADBE Mask Parade"` tdmn 子树下 —— 在 = mask 路径，不在 = shape 路径。

## Example

```go
if !layer.IsShapeLayer {
    return
}
for _, sp := range layer.ShapePaths {
    fmt.Printf("%q  closed=%v  vertices=%d\n", sp.Name, sp.Closed, len(sp.Vertices))
    for i, v := range sp.Vertices {
        fmt.Printf("  [%d] anchor=%v in=%v out=%v\n", i, v.Anchor, v.InTangent, v.OutTangent)
    }
}

// 参数化基元仍在 Properties 里
for _, p := range layer.Properties {
    if p.MatchName == "ADBE Vector Rect Size" {
        fmt.Println("rect size:", p.StaticValue)
    }
}
```

---

## Attributes

### ShapePath.Name

```go
Name string
```

路径名（"Path 1" 等 AE 默认或用户命名）。来自 `omtn`；未命名为空。read-only。

---

### ShapePath.Closed

```go
Closed bool
```

是否封闭路径。来自 `shph` chunk `@0x14 == 0x01`。read-only。

---

### ShapePath.Vertices

```go
Vertices []MaskVertex
```

路径顶点列表。每个 `MaskVertex` 是 Bezier 三元组（Anchor + InTangent + OutTangent，绝对坐标）。完整规则见 [mask.md#maskvertex](mask.md#maskvertex)。

read-only。

---

### ShapePath.ShphRaw

```go
ShphRaw []byte
```

24 字节 shph chunk 原始 payload。`Closed` 已解出，其余保留以利 round-trip 写回。read (raw)。

---

## 当前限制

- **无 `Path Keyframes`**：动画化的 Shape 路径解码尚未实现（mask 已有 `PathKeyframes`，shape 同结构理论可加）。
- **无路径顶点写回**：修改 Pen-drawn 路径顶点不支持。

---

# ShapePrimitive object

`layer.ShapePrimitives[index]`

## Description

Shape Layer 里的**参数化基元** —— Rectangle / Ellipse / Star（Polygon 也算 Star，由 `StarType` 区分）。每个 primitive 占用 Shape Layer 内一个 Vector Group。多个 primitive 可以共存在同一图层中。

每个字段都是 `*Property`（可能 nil）—— `*Property` 与 `Layer.Properties` 里同 match-name 的扁平条目共享底层 chunk 引用，调 `SetStaticValue` / Keyframe setter 任何一边都改同一份字节。

> AE 在某些 property 等于默认值时不写 cdat chunk（例如 `Rect.Position == [0,0]`），此时对应字段为 nil。

## Example

```go
for _, comp := range proj.Compositions {
    for _, layer := range comp.Layers {
        if !layer.IsShapeLayer {
            continue
        }
        for i, prim := range layer.ShapePrimitives {
            fmt.Printf("[%d] kind=%s group=%q\n", i, prim.Kind, prim.GroupName)
            switch prim.Kind {
            case aep.ShapePrimitiveRect:
                if prim.Size != nil { fmt.Println("  size:", prim.Size.StaticValue) }
                if prim.Roundness != nil { fmt.Println("  roundness:", prim.Roundness.StaticValue) }
            case aep.ShapePrimitiveEllipse:
                if prim.Size != nil { fmt.Println("  size:", prim.Size.StaticValue) }
            case aep.ShapePrimitiveStar:
                if prim.Points != nil { fmt.Println("  points:", prim.Points.StaticValue) }
                if prim.OuterRadius != nil { fmt.Println("  outer R:", prim.OuterRadius.StaticValue) }
            }
        }
    }
}
```

## Attributes

### ShapePrimitive.Kind

```go
Kind ShapePrimitiveKind
```

`"rect"` / `"ellipse"` / `"star"`。常量：

```go
const (
    ShapePrimitiveRect    ShapePrimitiveKind = "rect"
    ShapePrimitiveEllipse ShapePrimitiveKind = "ellipse"
    ShapePrimitiveStar    ShapePrimitiveKind = "star"
)
```

read-only。

---

### ShapePrimitive.GroupName

```go
GroupName string
```

owning Vector Group 的显示名（AE timeline 里看到的那个）。从 group 的 `tdsn` 字段（内嵌一个 `Utf8` 子记录）解出。read-only。

---

### Common-shape fields

| 字段 | 类型 | 适用 kind | 含义 |
|---|---|---|---|
| `Size` | `*Property` (2D `[w, h]`) | Rect / Ellipse | 基元尺寸 |
| `Position` | `*Property` (2D `[x, y]`) | Rect / Ellipse / Star | 在 owning group 内的局部偏移 |
| `Roundness` | `*Property` (1D) | Rect | 圆角半径 |

### Star-only fields

| 字段 | 类型 | 含义 |
|---|---|---|
| `StarType` | `*Property` (1D enum) | 1 = Star，2 = Polygon |
| `Points` | `*Property` (1D, 整数) | 边数 |
| `Rotation` | `*Property` (1D, degrees) | 旋转 |
| `InnerRadius` | `*Property` (1D) | 内半径（Polygon 模式忽略） |
| `OuterRadius` | `*Property` (1D) | 外半径 |
| `InnerRoundness` | `*Property` (1D, %) | 内圆角（AE 内部 match-name 拼错为 `"Inner Roundess"`；本字段名修正） |
| `OuterRoundness` | `*Property` (1D, %) | 外圆角（同上拼错原因） |

---

## 写回

每个非 nil 字段都是普通 `*Property`，可调 `SetStaticValue` / Keyframe API：

```go
for _, prim := range layer.ShapePrimitives {
    if prim.Kind == aep.ShapePrimitiveRect && prim.Size != nil {
        prim.Size.SetStaticValue([]float64{500, 250}) // 改 Rect 尺寸到 500×250
    }
}
```

---

## 当前限制

- 增删 primitive 不支持（需要重排 Vector Group 子节点）
- AE 默认值（Rect Position [0,0]、Star StarType 1 等）不会写 cdat，对应字段为 nil — 调用方需要 nil-check
- AE 的 match-name 拼写 `"ADBE Vector Star Inner Roundess"` / `"Outer Roundess"` 是 AE 自己的错字，本库匹配时按字面值，但 Go 字段名修正为 `InnerRoundness` / `OuterRoundness`
