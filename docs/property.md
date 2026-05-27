# Property object

`layer.Properties[index]` / `layer.PropertyByMatchName(name)` / `layer.Position()` 等访问器

## Description

可动画的图层属性 —— Transform 五件套（Position / Scale / Rotation / AnchorPoint / Opacity）、Effect 内部参数、Mask 内部 Feather / Opacity / Expansion 等都是 `Property`。

每个 Property **要么有 keyframes**（`Keyframes` 非空，`StaticValue` 为 nil），**要么有 static value**（反过来）。两者不会同时存在。

`Expression` 不为空时表示属性挂了 JS 表达式 —— 这与 keyframes / static value **不互斥**，AE 运行时用表达式覆盖。

## Example

```go
pos := layer.Position()
if pos != nil {
    fmt.Printf("%s  components=%d  keyframes=%d  static=%v  expr=%q\n",
        pos.MatchName, pos.Components, len(pos.Keyframes), pos.StaticValue, pos.Expression)
    for i, kf := range pos.Keyframes {
        fmt.Printf("  kf[%d] t=%.2fs value=%v in=%s out=%s\n",
            i, kf.Time, kf.Value, kf.InInterp, kf.OutInterp)
    }
}
```

---

## Attributes

### Property.MatchName

```go
MatchName string
```

ADBE 标识符 —— 如 `"ADBE Position"`、`"ADBE Opacity"`、`"ADBE Gaussian Blur 2-0001"`（效果参数有 `-NNNN` 后缀）。read-only。

---

### Property.Name

```go
Name string
```

用户可见名（通常为空，AE 用 matchname 自动生成显示名）。read-only。

---

### Property.Components

```go
Components int
```

#### Description

属性维度：

| 值 | 类型 |
|---|---|
| 1 | 标量 (Opacity / Rotation / Effect knob) |
| 2 | 2D 点 (Mask Feather X+Y) |
| 3 | 3D 点 (Position / Scale / Anchor Point) |
| 4 | 4D 颜色 RGBA (Tritone Highlights 等) |

`Keyframes[i].Value` 和 `StaticValue` 的 Go 类型由它决定：
- `Components == 1` → `float64`
- `Components >= 2` → `[]float64`（长度 = Components）

#### Type

`int`；read-only。

---

### Property.Keyframes

```go
Keyframes []*Keyframe
```

#### Description

关键帧列表（按时间升序）。每个 Keyframe 可用 `SetTime` / `SetValue` 等 length-preserving 改写；新增 / 删除 keyframe 用 [`InsertKeyframe`](#propertyinsertkeyframe) / [`DeleteKeyframe`](#propertydeletekeyframe)（length-variable，自动更新 lhd3 count）。详见下面 [Keyframe](#keyframe).

#### Type

`[]*Keyframe`；read（每个 keyframe 内部可写）。

---

### Property.StaticValue

```go
StaticValue any
```

#### Description

无关键帧时的常量值。Go 类型：

- `Components == 1`：`float64`
- `Components >= 2`：`[]float64`

有 keyframes 时为 nil。

#### Type

`any`；read / write via [`SetStaticValue`](#propertysetstaticvalue)。

---

### Property.Expression

```go
Expression string
```

#### Description

属性挂的 JavaScript 表达式源码。空字符串表示无表达式。

```go
if pos.Expression != "" {
    fmt.Println("expr:", pos.Expression) // 如 "time * 100"
}
```

不影响 `StaticValue` / `Keyframes` 的存储（AE 运行时表达式覆盖它们）。

#### Type

`string`；read / write via [`SetExpression`](#propertysetexpression)。

---

### Property.DefaultValue

```go
DefaultValue any
```

属性的默认值（AE 认为的"未修改"状态）。Transform 属性从硬编码表赋值（Scale→[100,100,100]，Opacity→100 等）；Effect 参数从 `pard` chunk 提取。非 transform、非 effect 属性为 nil。

---

### Property.LastValue

```go
LastValue any
```

Effect 参数的上次设置值（从 `pard` chunk 提取）。非 effect 属性为 nil。

---

### Property.NbOptions

```go
NbOptions int
```

Dropdown/Enum effect 参数的选项数量（从 `pard` chunk 的 `nb_options >> 16` 提取）。非 enum 属性为 0。

---

### Property.ControlType

```go
func (p *Property) ControlType() PropertyControlType
```

UI 控件类型（标量滑块、颜色选择器、角度盘、复选框、下拉菜单等）。从 tdb4 flags 推导。

| 常量 | 值 | 含义 |
|---|---|---|
| `PCTLLayer` | 0 | 图层引用 |
| `PCTLInteger` | 1 | 整数 |
| `PCTLScalar` | 2 | 标量滑块 |
| `PCTLAngle` | 3 | 角度盘 |
| `PCTLBoolean` | 4 | 复选框 |
| `PCTLColor` | 5 | 颜色选择器 |
| `PCTLTwoD` | 6 | 2D 点 |
| `PCTLEnum` | 7 | 下拉菜单 |
| `PCTLThreeD` | 18 | 3D 点 |
| `PCTLUnknown` | 15 | 未知 |

---

### Property.ValuePropertyType

```go
func (p *Property) ValuePropertyType() PropertyValueType
```

属性存储的值类型（1D / 2D / 3D / Color / NoValue 等）。从 tdb4 flags 推导。对应 ExtendScript 的 `Property.propertyValueType`。

| 常量 | 值 | 含义 |
|---|---|---|
| `PVTUnknown` | 0 | 未知 |
| `PVTNoValue` | 6412 | 无值（分隔符/按钮） |
| `PVTThreeDSpatial` | 6413 | 3D 空间（Position） |
| `PVTThreeD` | 6414 | 3D 非空间（Scale） |
| `PVTTwoDSpatial` | 6415 | 2D 空间 |
| `PVTTwoD` | 6416 | 2D 非空间 |
| `PVTOneD` | 6417 | 标量 |
| `PVTColor` | 6418 | RGBA 颜色 |

---

### Property.MinValue / MaxValue

```go
func (p *Property) MinValue() any
func (p *Property) MaxValue() any
```

属性的最小/最大允许值。从 tdbs LIST 内的 `tdum` / `tduM` sibling chunks 解码。返回类型取决于属性种类：

- 颜色属性 → `[]float64`（长度 4）
- 整数属性 → `float64`（从 uint32 解码）
- 标量 → `float64`
- 多维 → `[]float64`

无 tdum/tduM chunk 时返回 `nil`。

---

### Property.UnitsText

```go
func (p *Property) UnitsText() string
```

属性值的单位描述（`"pixels"` / `"degrees"` / `"percent"` / `"seconds"` / `"dB"` 等）。从静态 match-name 映射表查找。无已知单位时返回空字符串。

---

### Property.PropertyIndex

```go
func (p *Property) PropertyIndex() int
```

属性在其父 `AEPropertyGroup` 中的 0-based 位置。无父组时返回 -1（parser 外构建的属性）。

---

### Property.PropertyDepth

```go
func (p *Property) PropertyDepth() int
```

从该属性到包含图层之间的父组层数。顶层组（Transform / Effects 等）为 1，其直接子属性为 2，依此类推。无父组时返回 -1。

---

## Methods

### Property.SetStaticValue

```go
func (p *Property) SetStaticValue(v any) error
```

#### Description

length-preserving 改静态值。**仅在 `Keyframes` 为空时可用**（有 keyframes 的属性没有 static-value chunk，调用返回 error）。

参数类型：
- 1D 属性传 `float64`
- 多维属性传 `[]float64`（长度必须 `== Components`）

```go
// 改 Opacity 到 50%
if op := layer.Opacity(); op != nil && len(op.Keyframes) == 0 {
    op.SetStaticValue(0.5)
}

// 改 Position 到 (960, 540, 0)
if pos := layer.Position(); pos != nil && len(pos.Keyframes) == 0 {
    pos.SetStaticValue([]float64{960, 540, 0})
}
```

#### Returns

`error`；属性有 keyframes、类型不匹配、维度不一致都返回错误。

---

### Property.InsertKeyframe

```go
func (p *Property) InsertKeyframe(time float64, value any) (*Keyframe, int, error)
```

#### Description

构造一个 `bytesPerKF` 字节的新关键帧 block 插入到 `ldat` 流中，并把 `lhd3` 的 count header (`@0x08`) +1。返回新 Keyframe 和它在时间排序后的索引（同时刻 ties 在已有 keyframe 之**后**）。

**前置条件**：属性必须已有 ≥1 个关键帧（用来克隆 layout header byte `@0x07`）。无关键帧的属性插入还不支持 —— 这种情况要么用 `SetStaticValue`、要么先在 AE 里造一个再读。

`value` 类型规则同 `Keyframe.SetValue`：1D 传 `float64`，多维传 `[]float64`。

新 keyframe 的 `InInterp` / `OutInterp` 默认 `InterpLinear`，ease + tangents 全 0。调返回的 Keyframe 的 `SetInInterp` / `SetInTemporalEase` / `SetInSpatialTangent` 微调。

```go
pos := layer.Position()
kf, idx, err := pos.InsertKeyframe(2.5, []float64{960, 540, 0})
if err == nil {
    kf.SetInInterp(aep.InterpBezier)
    kf.SetOutInterp(aep.InterpBezier)
}
```

#### Returns

`(*Keyframe, int, error)`；属性无现有 keyframe / 类型不匹配 / 维度不一致都返回错误。

### Property.DeleteKeyframe

```go
func (p *Property) DeleteKeyframe(i int) error
```

#### Description

删除索引 `i` 处的 keyframe（按 ldat 流中字节顺序，跟 `Property.Keyframes` 索引一致）。把 ldat 数据缩短 `bpk` 字节，同时 lhd3 count -1。剩余 keyframe 的 `.offset` 全部重算。

```go
op := layer.Opacity()
op.DeleteKeyframe(2) // 删除第 3 个
```

#### Returns

`error`；索引越界或属性无 keyframe 流都返回错误。

---

### Property.SetExpressionEnabled

```go
func (p *Property) SetExpressionEnabled(enabled bool) error
```

#### Description

切换 AE 是否在渲染时执行表达式（与 `SetExpression` 写的 JS 源码独立 —— 表达式源保留但可被禁用）。

length-preserving 1 字节。底层位置：tdb4 payload `@0x78`，**反语义**编码：

- `enabled=true` 写 `0x00`
- `enabled=false` 写 `0x01`

```go
op := layer.Opacity()
op.SetExpression("time * 50") // 写源码
op.SetExpressionEnabled(false) // 暂时禁用（保留源码，不渲染）
op.SetExpressionEnabled(true)  // 恢复执行
```

读取通过 `Property.ExpressionEnabled bool` 字段（tdb4 `@0x78` 反转后的非反语义形 —— `true` = AE 评估）。无表达式 / tdb4 字节缺失时默认 `true`。setter 写入会同步该字段。

#### Returns

`error`；属性的 tdbs 缺失 / tdb4 chunk 缺失 / tdb4 短于 0x79 字节都返回错误。

---

### Property.SetExpression

```go
func (p *Property) SetExpression(source string) error
```

#### Description

写 / 改 / 清表达式 JS 源码。length-variable —— 底层 Utf8 chunk 的 data 整体替换：

- 已有表达式 + 新 source 非空 → 替换 Utf8 chunk data
- 已有表达式 + 新 source 为空 → 从 tdbs 移除 Utf8 chunk（"无表达式"状态）
- 无表达式 + 新 source 非空 → 创建新的 Utf8 chunk 并 append 到 tdbs
- 无表达式 + 新 source 为空 → no-op

```go
pos.SetExpression("wiggle(2, 30)")
pos.SetExpression("")        // 清空
```

> AE 还有独立的 "Enable / Disable Expression" 开关；本 setter 操作的是 source 本身。`SetExpression("")` 后属性不挂任何表达式，等价于 AE UI 里删除表达式的状态。

#### Returns

`error`；当 Property 在 parser 之外构造（无 owning tdbs）返回错误。

---

## Property accessor constants

```go
const (
    MatchNameAnchorPoint = "ADBE Anchor Point"
    MatchNamePosition    = "ADBE Position"
    MatchNameScale       = "ADBE Scale"
    MatchNameRotateZ     = "ADBE Rotate Z"
    MatchNameOpacity     = "ADBE Opacity"
)
```

`Layer` 上有对应的便捷访问器 `AnchorPoint()` / `Position()` / `Scale()` / `Rotation()` / `Opacity()`，详见 [layer.md](layer.md#transform-accessors)。

---

# Keyframe object

`property.Keyframes[index]`

## Description

单个关键帧 —— 时间 + 值 + 插值方式 + 切线 + ease。

时间用秒（解码时按 owning composition 的 `TickRate` 换算）。Value 类型同 Property：1D 是 `float64`，多维是 `[]float64`。

切线和 ease 字段的 length 规则：

- **Spatial 属性**（Position / Anchor Point）：用 `InSpatialTangent` / `OutSpatialTangent`（`[]float64` 长度 3）；`InTemporalEase` / `OutTemporalEase` 长度 1（一条 ease，沿 motion path 速度）。
- **Non-spatial N-D 属性**（Scale / Mask Feather / Opacity / 4D 颜色）：`In/OutSpatialTangent` 为 nil；`In/OutTemporalEase` 长度 N（每分量一条 ease）。

## Example

```go
for _, kf := range pos.Keyframes {
    fmt.Printf("t=%.2f value=%v in=%s out=%s\n", kf.Time, kf.Value, kf.InInterp, kf.OutInterp)
    if len(kf.InSpatialTangent) == 3 {
        fmt.Printf("  in tangent: %v\n", kf.InSpatialTangent)
    }
    for i, e := range kf.OutTemporalEase {
        fmt.Printf("  out ease[%d] speed=%.3f influence=%.3f\n", i, e.Speed, e.Influence)
    }
}

// 改时间 + 值
pos.Keyframes[0].SetTime(0.5)
pos.Keyframes[0].SetValue([]float64{500, 300, 0})
```

---

## Attributes

### Keyframe.Time

```go
Time float64
```

时间（秒），按 comp `TickRate` 换算。read / write via [`SetTime`](#keyframesettime)。

---

### Keyframe.Value

```go
Value any
```

关键帧值。1D 属性是 `float64`，多维是 `[]float64`（长度 = `Property.Components`）。read / write via [`SetValue`](#keyframesetvalue)。

---

### Keyframe.InInterp / OutInterp

```go
InInterp  InterpType
OutInterp InterpType
```

插值方式（每端独立）：`InterpLinear` (1) / `InterpBezier` (2) / `InterpHold` (3)。read / write via [`SetInInterp`](#keyframesetininterp) / [`SetOutInterp`](#keyframesetoutinterp)。

---

### Keyframe.InSpatialTangent / OutSpatialTangent

```go
InSpatialTangent  []float64
OutSpatialTangent []float64
```

空间切线（3D 向量）。只在 spatial 属性（Position / Anchor Point）有，长度恒为 3。非 spatial 属性为 nil。read / write via [`SetInSpatialTangent`](#keyframesetinspatialtangent) / [`SetOutSpatialTangent`](#keyframesetoutspatialtangent)。

---

### Keyframe.InTemporalEase / OutTemporalEase

```go
InTemporalEase  []TemporalEase
OutTemporalEase []TemporalEase
```

#### Description

时间 ease（每端独立，每分量一条或一共一条）：

| 属性类型 | 长度 |
|---|---|
| Spatial（Position / Anchor） | 1（沿 motion path 整体速度） |
| Non-spatial 1D（Opacity） | 1 |
| Non-spatial N-D（Scale 3D / Feather 2D / Color 4D） | N |

#### Type

`[]TemporalEase`；read-only。

---

## Methods

### Keyframe.SetTime

```go
func (k *Keyframe) SetTime(seconds float64) error
```

#### Description

length-preserving 改关键帧时间。内部按 owning composition 的 `TickRate` 编码为 uint32 ticks。

```go
pos.Keyframes[0].SetTime(2.5)
```

#### Returns

`error`；负数时间或 underlying ldat chunk 缺失返回错误。

---

### Keyframe.SetValue

```go
func (k *Keyframe) SetValue(v any) error
```

#### Description

length-preserving 改关键帧值。Spatial / non-spatial 两种字节布局自动适配。

参数类型：
- 1D 属性传 `float64`
- 多维属性传 `[]float64`（长度必须等于 `Property.Components`）

```go
opa.Keyframes[0].SetValue(0.5)
pos.Keyframes[0].SetValue([]float64{960, 540, 0})
```

#### Returns

`error`；类型不匹配或维度不一致返回错误。

---

### Keyframe.SetInInterp

```go
func (k *Keyframe) SetInInterp(t InterpType) error
```

写入新的 in-side 插值方式（block `@0x04`，1 字节）。

```go
kf.SetInInterp(aep.InterpBezier)
```

### Keyframe.SetOutInterp

```go
func (k *Keyframe) SetOutInterp(t InterpType) error
```

写入新的 out-side 插值方式（block `@0x05`，1 字节）。

---

### Keyframe.SetInTemporalEase

```go
func (k *Keyframe) SetInTemporalEase(eases []TemporalEase) error
```

#### Description

写入新的 in-side 时间 ease 列表。slice 长度必须匹配 keyframe 当前的 ease 形状：

- spatial 属性（Position / Anchor）或 1D non-spatial：长度 **1**
- non-spatial N-D（Scale 3D / Mask Feather 2D / 4D 颜色）：长度 **N**

按 layout 自动写到正确的字节偏移：spatial 写 block `@0x18`(speed) / `@0x20`(influence)；non-spatial 写 `@0x08 + (N+i)*8` / `@0x08 + (2N+i)*8`。

```go
// 1D：单条 ease
opa.Keyframes[0].SetInTemporalEase([]aep.TemporalEase{{Speed: 1.5, Influence: 0.25}})

// 非空间 3D：每分量一条
scale.Keyframes[0].SetInTemporalEase([]aep.TemporalEase{
    {Speed: 0, Influence: 1.0/3},
    {Speed: 0, Influence: 1.0/3},
    {Speed: 0, Influence: 1.0/3},
})
```

#### Returns

`error`；slice 长度不对、ldat 缺失、或 block bounds 不足都返回错误。

---

### Keyframe.SetOutTemporalEase

```go
func (k *Keyframe) SetOutTemporalEase(eases []TemporalEase) error
```

镜像 `SetInTemporalEase`，写 out-side：spatial 写 `@0x28/@0x30`；non-spatial 写 `@0x08 + (3N+i)*8` / `@0x08 + (4N+i)*8`。

---

### Keyframe.SetInSpatialTangent

```go
func (k *Keyframe) SetInSpatialTangent(v []float64) error
```

写入 in-side 3D Bezier 切线向量。**仅 spatial 属性**（Position / Anchor + motion path）有意义；非 spatial 属性调用返回 error。slice 长度必须等于 `Property.Components`（通常 3）。

底层偏移：`valueOff + dims*8`（in-tan 紧接 value 之后）。

```go
pos.Keyframes[0].SetInSpatialTangent([]float64{10, 20, 0})
```

### Keyframe.SetOutSpatialTangent

```go
func (k *Keyframe) SetOutSpatialTangent(v []float64) error
```

镜像 SetInSpatialTangent，写到 `valueOff + 2*dims*8`。

---

# TemporalEase

```go
type TemporalEase struct {
    Speed     float64 // 单位/秒；0 = Easy Ease 在该点停顿
    Influence float64 // 0..1；AE 默认 1/3
}
```

#### Description

时间 ease 的两个分量。

- **Speed** = 值在该关键帧时刻的瞬时变化率（单位/秒）。设为 0 = "Easy Ease" 该端停顿。
- **Influence** = ease 的"持续度"（影响曲线展开多远）。AE UI 用百分比，0..1 对应 0%..100%；默认 ≈ 0.333。

read-only。

---

# InterpType

```go
type InterpType uint8

const (
    InterpLinear InterpType = 1
    InterpBezier InterpType = 2
    InterpHold   InterpType = 3
)
```

关键帧插值方式。带 `String()` 方法（输出 `"linear"` / `"bezier"` / `"hold"`）。
