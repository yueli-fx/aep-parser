# Constants & enums

所有公开的 enum 和常量速查。

---

## LayerType

```go
type LayerType string
```

`Layer.Type` 的取值。字符串别名，便于打印。

| 常量 | 值 | 含义 |
|---|---|---|
| `LayerTypeAV` | `"av"` | 普通 AV 图层（含 footage / pre-comp） |
| `LayerTypeText` | `"text"` | 文本图层 |
| `LayerTypeShape` | `"shape"` | Shape 图层 |
| `LayerTypeNull` | `"null"` | Null Object |
| `LayerTypeLight` | `"light"` | 灯光 |
| `LayerTypeCamera` | `"camera"` | 摄像机 |
| `LayerTypeAdjust` | `"adjustment"` | 调整图层 |

---

## BlendingMode

```go
type BlendingMode uint8
```

`Layer.BlendingMode` 的取值。来自 ldta `@0x63`，与 AE 内部 enum / py-aep 编号一致。

| 常量 | 值 |
|---|---|
| `BlendingModeNormalCamera` | 0 |
| `BlendingModeNormal` | 2 |
| `BlendingModeDissolve` | 3 |
| `BlendingModeAdd` | 4 |
| `BlendingModeMultiply` | 5 |
| `BlendingModeScreen` | 6 |
| `BlendingModeOverlay` | 7 |
| `BlendingModeSoftLight` | 8 |
| `BlendingModeHardLight` | 9 |
| `BlendingModeDarken` | 10 |
| `BlendingModeLighten` | 11 |
| `BlendingModeClassicDifference` | 12 |
| `BlendingModeHue` | 13 |
| `BlendingModeSaturation` | 14 |
| `BlendingModeColor` | 15 |
| `BlendingModeLuminosity` | 16 |
| `BlendingModeStencilAlpha` | 17 |
| `BlendingModeStencilLuma` | 18 |
| `BlendingModeSilhouetteAlpha` | 19 |
| `BlendingModeSilhouetteLuma` | 20 |
| `BlendingModeLuminescentPremul` | 21 |
| `BlendingModeAlphaAdd` | 22 |
| `BlendingModeClassicColorDodge` | 23 |
| `BlendingModeClassicColorBurn` | 24 |
| `BlendingModeExclusion` | 25 |
| `BlendingModeDifference` | 26 |
| `BlendingModeColorDodge` | 27 |
| `BlendingModeColorBurn` | 28 |
| `BlendingModeLinearDodge` | 29 |
| `BlendingModeLinearBurn` | 30 |
| `BlendingModeLinearLight` | 31 |
| `BlendingModeVividLight` | 32 |
| `BlendingModePinLight` | 33 |
| `BlendingModeHardMix` | 34 |
| `BlendingModeLighterColor` | 35 |
| `BlendingModeDarkerColor` | 36 |
| `BlendingModeSubtract` | 37 |
| `BlendingModeDivide` | 38 |

> Null / Camera / Light 默认是 `BlendingModeNormalCamera` (0)，不是 `BlendingModeNormal` (2)。

---

## TrackMatteType

```go
type TrackMatteType uint8
```

`Layer.TrackMatte` 的取值。来自 ldta `@0x6B`。

| 常量 | 值 | 含义 |
|---|---|---|
| `TrackMatteNone` | 0 | 无轨道蒙版 |
| `TrackMatteAlpha` | 1 | Alpha Matte（上一层 alpha） |
| `TrackMatteAlphaInverse` | 2 | Alpha Inverted |
| `TrackMatteLuma` | 3 | Luma Matte |
| `TrackMatteLumaInverse` | 4 | Luma Inverted |

---

## AutoOrientType

```go
type AutoOrientType uint8
```

`Layer.AutoOrient` 的取值。来自 ldta `@0x25` / `@0x26` 中 3 个互斥位的解码。带 `String()` 方法。

| 常量 | 值 | 含义 |
|---|---|---|
| `AutoOrientNone` | 0 | 无自动朝向 |
| `AutoOrientAlongPath` | 1 | 沿 Position 路径朝向 |
| `AutoOrientCameraOrPointOfInterest` | 2 | 面向摄像机或其 POI |
| `AutoOrientCharactersTowardCamera` | 3 | 3D 文本字符级 billboard |

---

## LayerQuality

```go
type LayerQuality uint16
```

`Layer.Quality` 的取值。来自 ldta `@0x04`。

| 常量 | 值 |
|---|---|
| `LayerQualityWireframe` | 0 |
| `LayerQualityDraft` | 1 |
| `LayerQualityBest` | 2 |

---

## MaskMode

```go
type MaskMode uint32
```

`Mask.Mode` 的取值。来自 mkif `@0x04`。带 `String()` 方法。

| 常量 | 值 | String() |
|---|---|---|
| `MaskModeNone` | 0 | `"none"` |
| `MaskModeAdd` | 1 | `"add"` |
| `MaskModeSubtract` | 2 | `"subtract"` |
| `MaskModeIntersect` | 3 | `"intersect"` |
| `MaskModeLighten` | 4 | `"lighten"` |
| `MaskModeDarken` | 5 | `"darken"` |
| `MaskModeDifference` | 6 | `"difference"` |

---

## InterpType

```go
type InterpType uint8
```

`Keyframe.InInterp` / `OutInterp` 的取值。带 `String()` 方法。

| 常量 | 值 | String() |
|---|---|---|
| `InterpLinear` | 1 | `"linear"` |
| `InterpBezier` | 2 | `"bezier"` |
| `InterpHold` | 3 | `"hold"` |

---

## TextJustification

```go
type TextJustification int
```

`TextSource.Justification` / `TextParagraph.Justification` 的取值。带 `String()` 方法。

| 常量 | 值 | String() |
|---|---|---|
| `TextJustifyLeft` | 0 | `"Left"` |
| `TextJustifyRight` | 1 | `"Right"` |
| `TextJustifyCenter` | 2 | `"Center"` |

---

## BitsPerChannel

```go
type BitsPerChannel uint8
```

`Project.BitsPerChannel` 的取值。带 `String()` 方法。

| 值 | String() |
|---|---|
| 8 | `"8bpc"` |
| 16 | `"16bpc"` |
| 32 | `"32bpc"` |

---

## Property MatchName 常量

Transform 五件套的 ADBE matchname，便于直接传给 `Layer.PropertyByMatchName` 而不是手输字符串。

```go
const (
    MatchNameAnchorPoint = "ADBE Anchor Point"
    MatchNamePosition    = "ADBE Position"
    MatchNameScale       = "ADBE Scale"
    MatchNameRotateZ     = "ADBE Rotate Z"
    MatchNameOpacity     = "ADBE Opacity"
)
```

`Layer` 上同名访问器（`Position()` 等）更简洁，详见 [layer.md](layer.md#transform-accessors)。

> 3D 图层的 X / Y 旋转 matchname：`"ADBE Rotate X"` / `"ADBE Rotate Y"`（未提常量；用字面量）。

---

## Marker label 索引

`Marker.Label` 是 AE timeline 色卡索引，0..16：

| 索引 | 默认色名（AE 默认 palette） |
|---|---|
| 0 | None (Layer 默认色) |
| 1 | Red |
| 2 | Yellow |
| 3 | Aqua |
| 4 | Pink |
| 5 | Lavender |
| 6 | Peach |
| 7 | Sea Foam |
| 8 | Blue |
| 9 | Green |
| 10 | Purple |
| 11 | Orange |
| 12 | Brown |
| 13 | Fuchsia |
| 14 | Cyan |
| 15 | Sandstone |
| 16 | Dark Green |

> 用户可在 AE Preferences → Labels 修改色卡到任意 RGB；索引本身稳定，RGB 不稳定。
