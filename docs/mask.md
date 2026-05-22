# Mask object

`layer.Masks[index]`

## Description

图层的矢量蒙版。每个 mask 有一条 Bezier 路径（封闭或开放）+ 模式（Add / Subtract / 等）+ Feather / Opacity / Expansion 等可选属性。

路径可以是**静态**（`Vertices` 字段，单次快照）或**动画**（`PathKeyframes`，每个时间点一份完整路径快照 + scalar ease）。动画 mask 的 `Vertices` 镜像 `PathKeyframes[0].Vertices`，方便不关心动画的调用方。

## Example

```go
for _, m := range layer.Masks {
    fmt.Printf("Mask %q  mode=%s  inverted=%v  closed=%v  feather=%v  opacity=%g  exp=%g\n",
        m.Name, m.Mode, m.Inverted, m.Closed, m.Feather, m.Opacity, m.Expansion)
    if len(m.PathKeyframes) > 0 {
        fmt.Printf("  animated path: %d snapshots\n", len(m.PathKeyframes))
    } else {
        fmt.Printf("  static path: %d vertices\n", len(m.Vertices))
    }
    for i, v := range m.Vertices {
        fmt.Printf("    [%d] anchor=%v in=%v out=%v\n", i, v.Anchor, v.InTangent, v.OutTangent)
    }
}
```

---

## Attributes

### Mask.Name

```go
Name string
```

mask 名（AE timeline 显示）。来自 `omtn` chunk；未命名 mask 为空。read-only。

---

### Mask.Closed

```go
Closed bool
```

是否封闭路径。来自 `shph` chunk `@0x14 == 0x01`。read / write via [`SetClosed`](#masksetclosed)。

---

### Mask.Mode

```go
Mode MaskMode
```

#### Description

合成模式。值：

| 常量 | 值 |
|---|---|
| `MaskModeNone` | 0 |
| `MaskModeAdd` | 1 |
| `MaskModeSubtract` | 2 |
| `MaskModeIntersect` | 3 |
| `MaskModeLighten` | 4 |
| `MaskModeDarken` | 5 |
| `MaskModeDifference` | 6 |

来自 mkif `@0x04` (uint32 BE)。带 `String()` 方法。

#### Type

`MaskMode`（`uint32` 别名）；read / write via [`SetMode`](#masksetmode)。

---

### Mask.Inverted

```go
Inverted bool
```

是否反转。来自 mkif `@0x00`。read / write via [`SetInverted`](#masksetinverted)。

---

### Mask.Index

```go
Index uint32
```

AE 内部 mask 编号（1-based，不一定连续 —— AE 删除 mask 后保留 ID）。来自 mkif `@0x08`。read-only。

---

### Mask.Color

```go
Color [3]uint8
```

#### Description

mask 在 AE timeline 旁的色卡 RGB。来自 mkif `@0x2D-0x2F`（alpha `@0x2C` 恒为 `0xFF`）。

#### Type

`[3]uint8`；read / write via [`SetColor`](#masksetcolor)。

---

### Mask.Vertices

```go
Vertices []MaskVertex
```

#### Description

路径顶点列表。

- 静态 mask：唯一的路径快照
- 动画 mask：镜像 `PathKeyframes[0].Vertices`

每个 `MaskVertex` 是 Bezier 控制点三元组（详见下面）。

#### Type

`[]MaskVertex`；read-only。

---

### Mask.PathKeyframes

```go
PathKeyframes []MaskPathKeyframe
```

#### Description

动画 mask 的路径关键帧序列。静态 mask 为 nil。

每个 keyframe 持有该时间点的**完整路径快照** + scalar ease（一条 ease，不是 per-vertex）。

#### Type

`[]MaskPathKeyframe`；read-only。

---

### Mask.Feather

```go
Feather [2]float64
```

mask 边缘羽化半径 (X, Y)，像素。当 mask atom 的 `"ADBE Mask Feather"` 子属性有 cdat 时填充；keyframed feather 时为零（动画值在 `Properties` 里）。read-only（间接可改：找到 Property 调 `SetStaticValue`）。

---

### Mask.Opacity

```go
Opacity float64
```

mask 不透明度，0..1。默认 1.0。read-only（间接可改）。

---

### Mask.Expansion

```go
Expansion float64
```

mask 扩展（AE UI "Mask Expansion"，内部 `"ADBE Mask Offset"`）。像素，正/负值都有效。read-only（间接可改）。

---

### Mask.Properties

```go
Properties []*Property
```

#### Description

mask atom 内的所有其他 tdmn+tdbs 叶子属性。在以下场景特别有用：

- keyframed Feather / Opacity / Expansion（动画时 `Mask.Feather` 等字段不填充）
- 未识别的其他 mask 子属性

```go
for _, p := range mask.Properties {
    fmt.Println(p.MatchName, len(p.Keyframes), p.StaticValue)
}
```

#### Type

`[]*Property`；read（Property 内部可写）。

---

### Mask.MkifRaw

```go
MkifRaw []byte
```

48 字节 mkif chunk 原始 payload。round-trip 写回时直接复用 —— 我们只解出了部分字段（mode/inverted/index/color），剩余字节（0x10 / 0x18 / 0x20-0x27 等）保留以避免破坏。read (raw)。

---

### Mask.ShphRaw

```go
ShphRaw []byte
```

24 字节 shph chunk 原始 payload。`Closed` 字段已解出，其余保留。read (raw)。

---

# MaskVertex

```go
type MaskVertex struct {
    Anchor     [2]float64
    InTangent  [2]float64
    OutTangent [2]float64
}
```

## Description

一个 Bezier 控制点。**坐标都是绝对位置**，不是相对偏移：

| 字段 | 含义 |
|---|---|
| `Anchor` | 顶点本身的位置 |
| `InTangent` | 入射切线控制点 |
| `OutTangent` | 出射切线控制点 |

### 直线段的约定

- **直线入射**：`InTangent == Anchor`（控制点和顶点重合 → 退化 Bezier = 直线）
- **直线出射**：`OutTangent == nextVertex.Anchor`（控制点落在下一顶点 → 直线）

所以一个全直线的封闭多边形：每个顶点 `InTangent == Anchor`，且 `OutTangent == anchorOf(next vertex)`。曲线段则切线偏离 anchor。

read-only。

---

# MaskPathKeyframe

```go
type MaskPathKeyframe struct {
    Time            float64
    Vertices        []MaskVertex
    InInterp        InterpType
    OutInterp       InterpType
    InTemporalEase  TemporalEase   // 标量 ease（不是 per-vertex）
    OutTemporalEase TemporalEase
}
```

## Description

动画 mask 的一个路径快照。

- `Time` —— 秒
- `Vertices` —— 该时刻的完整路径（顶点数可与其他 snapshot 不同）
- 插值方式按端独立
- Ease 是**标量**（每端一条），不像普通 keyframe 可以 per-component

read-only。

---

# MaskMode

```go
type MaskMode uint32

const (
    MaskModeNone       MaskMode = 0
    MaskModeAdd        MaskMode = 1
    MaskModeSubtract   MaskMode = 2
    MaskModeIntersect  MaskMode = 3
    MaskModeLighten    MaskMode = 4
    MaskModeDarken     MaskMode = 5
    MaskModeDifference MaskMode = 6
)

func (m MaskMode) String() string  // "add" / "subtract" / 等
```

---

---

## Methods

length-preserving 字节写回。每个 setter 同步更新 Go 字段，下一次 `Project.WriteAEP` 持久化。

### Mask.SetMode

```go
func (m *Mask) SetMode(mode MaskMode) error
```

写入新模式枚举（mkif `@0x04`，uint32 BE）。

```go
mask.SetMode(aep.MaskModeSubtract)
```

### Mask.SetInverted

```go
func (m *Mask) SetInverted(v bool) error
```

切换 Inverted 标记（mkif `@0x00`，1 字节）。

### Mask.SetColor

```go
func (m *Mask) SetColor(rgb [3]uint8) error
```

写入时间线色卡 RGB（mkif `@0x2D/@0x2E/@0x2F`，3 字节）。alpha 字段（`@0x2C`）AE 始终保持 `0xFF`，本 setter 不动它。

### Mask.SetClosed

```go
func (m *Mask) SetClosed(v bool) error
```

切换路径是否封闭（shph `@0x14`，1 字节）。对动画 mask 只影响第一个快照。

### Mask.SetLocked

```go
func (m *Mask) SetLocked(v bool) error
```

切换 mask 锁定标志（mkif `@0x01`，0=unlocked，1=locked）。锁定后 AE UI 拒绝编辑该 mask；本库写回不受影响。读取通过 `Mask.Locked bool` 字段；setter 会同步该字段。

### Mask.SetMaskMotionBlur

```go
func (m *Mask) SetMaskMotionBlur(mode MaskMotionBlurMode) error
```

写 per-mask 运动模糊 override（mkif `@0x02`，1 字节枚举）：

| 常量 | 值 | 含义 |
|---|---|---|
| `MaskMotionBlurSameAsLayer` | 0 | 跟随 layer.MotionBlur（默认） |
| `MaskMotionBlurOn` | 2 | 强制开启 |
| `MaskMotionBlurOff` | 3 | 强制关闭 |

读取通过 `Mask.MotionBlur MaskMotionBlurMode` 字段；setter 会同步该字段。

---

## 当前限制

- 修改顶点 / 路径几何尚未实现（需要重写变长 shap kfl 流）
- 增删 mask 不支持（结构性修改）
- mkif 余下未解字节（0x10 / 0x18 / 0x20-0x27）：看起来不是用户可配置的常量字段，但未完整 RE
- 动画 mask 的 per-keyframe Closed 标记不能逐帧改
