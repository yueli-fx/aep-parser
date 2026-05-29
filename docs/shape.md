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

---

# V2.2 alpha — Builder API（从零构造 ShapeLayer）

新写路径：用 `aep.NewProject` + `Composition.NewShapeLayer` 构造可被 AE 2025 接受的 .aep（不需要在 AE 里手动建图层）。已经通过 6-variant ship gate，AE 打开后 `comp.layers.length=1, layer[1].class=ShapeLayer`。

## Example — 一个红色 200×200 矩形

```go
proj := aep.NewProject(aep.TargetAE2025)
comp, _ := proj.NewComposition("Main", 1920, 1080, 30, 5) // 5 seconds @ 30 fps

shape, _ := comp.NewShapeLayer("MyRect")
rect, _ := shape.RootGroup().AddRect()
_ = rect.SetSize([2]float64{200, 200})

fill, _ := shape.RootGroup().AddFill()
_ = fill.SetColor([4]float64{1, 0, 0, 1}) // RGBA 0..1

out, _ := os.Create("out.aep")
_ = proj.WriteAEP(out)
out.Close()
```

打开 `out.aep`：AE 显示 1 个 ShapeLayer，名字 "MyRect"，含一个 Rect + Fill 子节点。

## API surface

| API | 描述 |
|---|---|
| `aep.NewProject(target ...AETarget)` | 新空 project（V2.1）。`AETarget` = `TargetAE2020/2022/2025` |
| `proj.NewComposition(name, w, h, fps, duration)` | 新建空 comp（V2.1）|
| `(c *Composition) NewShapeLayer(name) (*ShapeLayer, error)` | 在 comp 里加新空 ShapeLayer，原子（warning/error → rollback）|
| `(s *ShapeLayer) RootGroup() *VectorGroup` | 取顶层 Contents 容器 |
| `(g *VectorGroup) AddRect() (*RectNode, error)` | 加 Rect 子节点 |
| `(g *VectorGroup) AddFill() (*FillNode, error)` | 加 Fill 子节点 |
| `(r *RectNode) SetSize([w, h] float64)` | 设矩形尺寸（static）|
| `(f *FillNode) SetColor([r, g, b, a] float64)` | 设填充色（RGBA 0..1，static）|

ShapeLayer 跟 V1 parse 出来的 `Layer` 同构 — `comp.Layers[i]` 既是 V1 `*Layer` 也能 `WrapShapeLayer(layer)` 拿到 V2.2 视图。

## V2.2 alpha 限制

V2.2 ship gate 走的是 **embed boilerplate** 路线（详 `flightdeck/incident-reports/v2-2-aelayer-structure.md` iter-7/8 实施记）：3 处 "complex multi-stream container" 字节直接从 tolerance.aep 拷出来作 `//go:embed` 资源，runtime 只覆盖 cdat 数值。这意味着：

### 不持久化（runtime-only）

调用 setter API 不报错，runtime 内存中能读到改后值，但 `WriteAEP` 后磁盘字节不变；再读回来是默认值。

| 字段 | 状态 |
|---|---|
| `ShapeLayer.Transform().AnchorPoint / Scale / Rotation / Opacity` | runtime-only |
| `ShapeLayer.Transform().Position` keyframes | 仅 first kf 作 static fallback |
| `RectNode.Position / Roundness / Direction` | runtime-only |
| `RectNode.Size` keyframes | 仅 first kf 作 static fallback |
| `FillNode.Opacity / BlendMode / CompositeOrder / FillRule` | runtime-only |
| `FillNode.Color` keyframes | 仅 first kf 作 static fallback |

### shape kind 支持状态

| 字段 | 状态 |
|---|---|
| `VectorGroup.AddEllipse` | ✅ **V2.2.1 已 ship**（AE 2020+2025 双版本 ship-gate PASS）— embed `v2_2_shape_ellipse_body.bin` + overwrite Size/Position cdat。Direction 仍 AE 默认；动画仍 first-kf static fallback |
| `VectorGroup.AddPath` | ✅ **V2.2.1 已 ship**（AE 2020+2025 双版本 ship-gate PASS）— embed `v2_2_shape_path_body.bin` + splice `encodeBezier` 几何（shph/lhd3/ldat）。ldat 逐顶点布局 = `[anchor, anchor+outTangent_i, anchor_{i+1}+inTangent_{i+1}]`（bbox 归一化，wrap mod n；V2.2.1 RE 修正，曾错存本顶点 in/out）。`SetVertices` 仅线性段（切线置零）；动画仍 first-kf fallback。**注**：from-scratch path 曾 **崩溃 AE 2020**（0::42），故走 embed |
| `VectorGroup.AddStroke` | ✅ **V2.2.1 已 ship**（AE 2020+2025 双版本 ship-gate PASS）— embed `v2_2_shape_stroke_body.bin`（含 Blend Mode/Composite Order/Line Cap/Join/Miter + Dashes/Taper/Wave 嵌套组）+ overwrite Color/Opacity/Width cdat。其余子属性留 embed 默认；动画仍 first-kf fallback |

> **AE 2020 地基修复（V2.2.1）**：ShapeLayer 的 ldta 大小现按 target 分支（160B AE 2020/22，164B AE 2025）。此前 buildLdtaBytes 硬编码 164B，导致 **所有** from-scratch shape 图层（含已"ship"的 Rect+Fill）被 AE 2020 判为损坏并跳过——因 AE-2020 shape ship-gate 长期 skip 而未发现。详 `flightdeck/incident-reports/ae2020-shape-ldta-164-corrupt.md`。

### Fill / Stroke Color 编码（V2.2.1 已 RE 修正）

AE 存 shape 颜色为 **`[A,R,G,B] × 255` 的 f64 BE**（offset 0/8/16/24），不是原始 `[r,g,b,a] × 1.0`（RE 自 stroke tolerance fixture：JSX `[0,0,1,1]` → 磁盘 `[255,0,0,255]`，0x406fe0=255）。`encodeShapeColorBE` 统一编码，Fill + Stroke 共用。修复前 Fill 写原始 RGBA 导致可见色错；现经 AE round-trip 验证正确（Ellipse gate 解 re-saved Fill Color = ARGB×255）。

### Keyframes 不持久化

所有 shape 子流 + Layr Position 的 keyframe 调用（`AddKeyframeLinear`）在 V2.2 alpha 里都仅作 *runtime* tracking + **first keyframe value 作 static fallback** 写盘。完整 keyframe 持久化需要 RE `LIST(list) lhd3/ldat` 在 embed body 里的注入方式，V2.2.1 工作。

## RE 路线 — 为什么是 embed 而不是 from-scratch 构造

V2.2 Phase 5 ship gate 经历了 iter-1 到 iter-8 共 8 轮。iter-6a/b/c/d/e/f 走 byte-level RE 路线（猜 tdsb / tdb4 head bytes / tdum / placeholder flags / trailing chunks 等单字段）6 轮无果。GPT 看完 bisect 数据 pivot 到 semantic-level：

1. `verify_baseline` 排除 measurement bug（tolerance 跑同样 probe 报 layers=1）
2. `swap_propgroup` / `transplant_*` 系列工具用 chunk-level swap 测试 isolate silent-drop 的 chunk
3. iter-7: 锁定 trigger = Layr Transform Group body → `//go:embed templates/v2_2_transform_group_body.bin` + 覆 Position cdat
4. iter-8: 同思路缩到 shape body → embed Rect + Fill bodies 各自 → 覆 Size / Color cdat

**教训**: silent-drop 类问题（AE 接受文件但内部不实例化 layer）是 semantic-level，不是 byte-level corruption。byte 路线在 silent-drop 场景是 dead end；transplant + embed 是正确 tool。完整 RE 历史见 `flightdeck/incident-reports/v2-2-aelayer-structure.md`。
