# Layer object

`composition.Layers[index]` / `composition.LayerByID(id)`

## Description

合成内的图层。Layer 是所有图层类型的统一类型（无子类）—— camera / light / text / shape / null / adjustment 通过 `Layer.Type` 字段区分。

字段大致分四组：

1. **基础信息** —— ID / 名称 / 类型 / 时间 / 父子关系
2. **渲染设置** —— Quality / Blending Mode / Track Matte / 各类 flag bit
3. **子集合** —— Properties / Effects / Markers / Masks / ShapePaths / TextSource
4. **方法** —— 父查找 / source comp 解析 / property 访问器 / `SetText`

## Example

```go
layer := proj.Compositions[0].Layers[0]
fmt.Printf("[%d] %q  type=%s  3D=%v  visible=%v\n",
    layer.Index, layer.Name, layer.Type, layer.Is3D, layer.Visible)

if parent := layer.Parent(); parent != nil {
    fmt.Println("  parent:", parent.Name)
}
if src := layer.SourceComposition(); src != nil {
    fmt.Println("  pre-comp source:", src.Name)
}
if pos := layer.Position(); pos != nil {
    fmt.Printf("  position has %d keyframes, static=%v\n", len(pos.Keyframes), pos.StaticValue)
}
```

---

## Basic information

### Layer.Index

```go
Index int
```

合成内顺序，0-based。AE UI 显示是 1-based，所以打印时常 `Index+1`。read-only。

---

### Layer.Name

```go
Name string
```

图层名（AE timeline 显示的那个）。read / write via [`SetName`](#layersetname)。

---

### Layer.Type

```go
Type LayerType
```

#### Description

图层类型，字符串别名。可能值：

| 值 | 含义 |
|---|---|
| `"av"` | 普通 AV 图层（含 footage / pre-comp） |
| `"text"` | 文本图层 |
| `"shape"` | Shape 图层 |
| `"null"` | Null Object |
| `"light"` | 灯光 |
| `"camera"` | 摄像机 |
| `"adjustment"` | 调整图层 |

#### Type

`LayerType`（`string` 别名）；read-only。

---

### Layer.ID

```go
ID uint32
```

own layer ID（ldta `@0x00`），被子图层的 `ParentID` 引用。real AE 工程从 1 开始。read-only。

---

### Layer.ParentID

```go
ParentID uint32
```

父图层 ID（ldta `@0x84`）。0 = 无父。用 [`Parent()`](#layerparent) 直接拿到 `*Layer` 而不是 ID。read / write via [`SetParent`](#layersetparent)。

---

### Layer.SourceID

```go
SourceID uint32
```

源项目项 ID（ldta `@0x28`）—— 指向 footage / pre-comp / solid。用 [`SourceComposition()`](#layersourcecomposition) 直接拿到 `*Composition`。read / write via [`SetSource`](#layersetsource)。

---

### Layer.StartTime

```go
StartTime float64
```

图层在合成时间轴上的起始时间（秒）。read / write via [`SetStartTime`](#layersetstarttime)。

---

### Layer.Duration

```go
Duration float64
```

图层时长（秒）。read（间接可写：调 [`SetInPoint`](#layersetinpoint) / [`SetOutPoint`](#layersetoutpoint) 改 source 端点；Duration = OutPoint − InPoint 自动重算）。

---

### Layer.Stretch

```go
Stretch float64
```

时间伸缩系数。1.0 = 原速度，2.0 = 2× 慢放。read / write via [`SetStretch`](#layersetstretch)。

---

### Layer.Comment

```go
Comment string
```

图层备注（AE timeline 的 Comments 列 / Layer Settings dialog）。来自 sibling `cmta` chunk；CRLF 已规范化为 LF；无 `cmta` 时为空。read / write via [`SetComment`](#layersetcomment)。

---

## Render settings

### Layer.Quality

```go
Quality LayerQuality
```

渲染质量：`LayerQualityWireframe` (0) / `LayerQualityDraft` (1) / `LayerQualityBest` (2)。来自 ldta `@0x04` (uint16)。read / write via [`SetQuality`](#layersetquality)。

---

### Layer.Label

```go
Label uint8
```

时间线色卡索引，0..16。0 = 默认。来自 ldta `@0x3D`。read / write via [`SetLabel`](#layersetlabel)。

---

### Layer.BlendingMode

```go
BlendingMode BlendingMode
```

混合模式 enum。40 个值，从 `BlendingModeNormal` (2) 到 `BlendingModeDivide` (38)。完整列表见 [constants.md](constants.md#blendingmode)。来自 ldta `@0x63`。read / write via [`SetBlendingMode`](#layersetblendingmode)。

---

### Layer.TrackMatte

```go
TrackMatte TrackMatteType
```

轨道蒙版模式：`None` / `Alpha` / `AlphaInverse` / `Luma` / `LumaInverse`。来自 ldta `@0x6B`。read / write via [`SetTrackMatte`](#layersettrackmatte)（仅改模式）或 [`SetTrackMatteLayer`](#layersettrackmattelayer)（模式 + 显式 source ID 一起写）。

---

### Layer.TrackMatteLayerID

```go
TrackMatteLayerID uint32
```

显式 track matte 源 layer ID（ldta `@0xA0`，AE 23+ 引入）。`0` = 无显式 source（AE ≤ 22 用隐式"上一层"约定，AE 23+ 也支持但 ScriptingAPI 默认走显式 ID）。read。

要 resolve 成 `*Layer` 用 [`TrackMatteLayer()`](#layertrackmattelayer)。要写改用 [`SetTrackMatteLayer`](#layersettrackmattelayer)。

> **AE 版本要求**：写入要求 ldta 长度 ≥ 0xA4（AE 22 写的 ldta 是 160 字节，AE 23+ 写的是 164 字节）。AE 22 文件强写会返回 error；先在 AE 23+ 打开重存一次 ldta 才会加长。

---

### Layer.TrackMatteLayer

```go
func (l *Layer) TrackMatteLayer() *Layer
```

把 `TrackMatteLayerID` resolve 成同 comp 内的 `*Layer`，找不到返回 `nil`（包括 ID=0、未挂回 comp、ID 不存在三种情况）。AE 23+ 显式 matte source。

```go
if src := layer.TrackMatteLayer(); src != nil {
    fmt.Printf("matte = layer %q mode=%v\n", src.Name, layer.TrackMatte)
}
```

---

### Layer.PreserveTransparency

```go
PreserveTransparency bool
```

"Preserve Underlying Transparency" 开关。来自 ldta `@0x67`。read / write via [`SetPreserveTransparency`](#layersetpreservetransparency)。

---

### Layer.AutoOrient

```go
AutoOrient AutoOrientType
```

自动朝向：`None` / `AlongPath` / `CameraOrPointOfInterest` / `CharactersTowardCamera`。来自 ldta `@0x25` / `@0x26` 中 3 个互斥位。read / write via [`SetAutoOrient`](#layersetautoorient)。

---

### Layer flag bits

每个都是 `bool`，来自 ldta `@0x25-0x27` 的位运算解码，全部可 read / write（writer 见下方 [Flag-bit setters](#flag-bit-setters)）：

| 字段 | 对应 AE UI | Setter |
|---|---|---|
| `Is3D` | 3D Layer 开关 | `SetIs3D` |
| `Solo` | Solo (●) | `SetSolo` |
| `Shy` | Shy (●) | `SetShy` |
| `Locked` | Lock (🔒) | `SetLocked` |
| `Visible` | Video (👁️) —— 位 0x27 bit0 | `SetVisible` |
| `IsAdjust` | Adjustment Layer | `SetIsAdjust` |
| `IsNull` | Null Object | `SetIsNull` |
| `IsGuide` | Guide Layer | `SetIsGuide` |
| `MarkersLocked` | Lock markers | `SetMarkersLocked` |
| `MotionBlur` | Motion Blur switch | `SetMotionBlur` |
| `EffectsEnabled` | Effects switch (fx) | `SetEffectsEnabled` |
| `AudioEnabled` | Audio switch | `SetAudioEnabled` |
| `FrameBlendEnabled` | Frame Blend switch | `SetFrameBlendEnabled` |
| `CollapseTransform` | Collapse Transformations / Continuously Rasterize | `SetCollapseTransform` |
| `SamplingBicubic` | Sampling mode — false = Bilinear, true = Bicubic | `SetSamplingBicubic` |
| `FrameBlendPixelMotion` | Frame Blend type — false = Frame Mix, true = Pixel Motion | `SetFrameBlendPixelMotion` |

---

### Layer.IsShapeLayer

```go
IsShapeLayer bool
```

`true` 当图层含 `"ADBE Root Vectors Group"` 属性树（= Shape Layer）。`Type` 也会被设为 `"shape"`。read-only。

---

## Sub-collections

### Layer.Properties

```go
Properties []*Property
```

#### Description

图层的所有动画属性 —— Transform（Position / Scale / Rotation 等）+ Effects 的内部参数（如果不走 `Effects` 集合）+ Mask / Shape 树等。

便捷查找见下面的 [Transform 访问器](#transform-accessors)；通用查找用 [`PropertyByMatchName`](#layerpropertybymatchname)。

#### Type

`[]*Property`；read-only。Property 内部的 keyframe / static value 可写（见 [property.md](property.md)）。

---

### Layer.Effects

```go
Effects []*Effect
```

应用到图层上的特效列表。详见 [effect.md](effect.md)。read-only（每个 Effect 的 Parameters 是 Property，可通过 Property 路径写）。

---

### Layer.Markers

```go
Markers []*Marker
```

图层 marker。详见 [marker.md](marker.md)。read-only。

---

### Layer.Masks

```go
Masks []*Mask
```

图层蒙版。详见 [mask.md](mask.md)。read-only（Mask 内部 Feather/Opacity/Expansion 可通过其 Properties 写）。

---

### Layer.ShapePaths

```go
ShapePaths []*ShapePath
```

Shape Layer 的 Pen 工具自由 Bezier 路径。只在 Shape Layer 有内容，否则为空。详见 [shape.md](shape.md)。read-only。

---

### Layer.TextSource

```go
TextSource *TextSource
```

文本图层解码视图。非文本图层为 nil。可通过 [`Layer.SetText`](#layersettext) 改文字字符串。详见 [text.md](text.md)。

---

### Layer.TextSourceRaw

```go
TextSourceRaw []byte
```

文本图层的原始 btds 字节 payload。decode 出 `TextSource` 后此字段仍保留，用于：

- round-trip 写回
- `Layer.SetText` 在线 splice
- 下游工具读取本库尚未结构化暴露的字段（baseline shift / scale / 等）

非文本图层为 nil。read (raw)。

---

## Methods

### Layer.Parent

```go
func (l *Layer) Parent() *Layer
```

#### Description

解析后的父图层（同 comp 内）。任何下列条件都返回 nil：

- `ParentID == 0`（无父）
- 同 comp 内没有匹配 ID 的图层
- Layer 不是经 parser 创建（无 owning comp 反向引用）

AE 强制 same-comp parenting；本方法不做跨 comp 查找。只返回 immediate parent；要追溯链需要自己递归调用。

#### Returns

`*Layer`；找不到返回 nil。

```go
chain := []string{layer.Name}
for p := layer.Parent(); p != nil; p = p.Parent() {
    chain = append(chain, p.Name)
}
```

---

### Layer.SourceComposition

```go
func (l *Layer) SourceComposition() *Composition
```

#### Description

如果该图层是 pre-comp（源是另一个合成），返回那个合成。其他情况都返回 nil：

- 源是 footage / solid / placeholder
- `SourceID` 在工程中找不到匹配项
- Layer 不是经 parser 创建

#### Returns

`*Composition`；找不到返回 nil。

```go
if src := layer.SourceComposition(); src != nil {
    fmt.Println("pre-comp:", src.Name)
}
```

---

### Layer.PropertyByMatchName

```go
func (l *Layer) PropertyByMatchName(name string) *Property
```

#### Description

按 ADBE matchname 找 property。第一个匹配返回，找不到返回 nil。

```go
if p := layer.PropertyByMatchName("ADBE Rotate Y"); p != nil { /* 3D 图层的 Y 轴旋转 */ }
```

#### Returns

`*Property`；找不到返回 nil。

---

### Transform accessors

便捷访问器，找不到都返回 nil（不 panic）。

```go
func (l *Layer) AnchorPoint() *Property  // "ADBE Anchor Point" — 3D
func (l *Layer) Position()    *Property  // "ADBE Position"     — 3D
func (l *Layer) Scale()       *Property  // "ADBE Scale"        — 3D
func (l *Layer) Rotation()    *Property  // "ADBE Rotate Z"     — 1D 度
func (l *Layer) Opacity()     *Property  // "ADBE Opacity"      — 1D 0..1
```

---

### Other AV-layer accessors

```go
func (l *Layer) TimeRemap()   *Property  // "ADBE Time Remapping" — 1D 秒；仅启用时间重映射的图层有
func (l *Layer) AudioLevels() *Property  // "ADBE Audio Levels"   — 2D [L, R] dB；仅有音频内容的图层有
```

```go
if tr := layer.TimeRemap(); tr != nil {
    for _, kf := range tr.Keyframes {
        fmt.Printf("source-time=%vs @ comp-time=%vs\n", kf.Value, kf.Time)
    }
}
```

> 3D 图层的 X / Y 旋转 matchname 是 `"ADBE Rotate X"` / `"ADBE Rotate Y"`，用 [`PropertyByMatchName`](#layerpropertybymatchname) 取。

预定义常量也可单独使用：

```go
const (
    MatchNameAnchorPoint = "ADBE Anchor Point"
    MatchNamePosition    = "ADBE Position"
    MatchNameScale       = "ADBE Scale"
    MatchNameRotateZ     = "ADBE Rotate Z"
    MatchNameOpacity     = "ADBE Opacity"
)
```

---

### Layer.SetText

```go
func (l *Layer) SetText(newText string) error
```

#### Description

文本图层用户字符串 length-preserving 替换。新文字编码后字节数**必须等于**原字节数，否则返回 error 不动数据。

编码规则与 AE 一致：

- `\n` → `\r`，段末自动补 `\r`
- 字符以 `\xFE\xFF` BOM + UTF-16BE 编码
- `(` `)` `\` 字节级转义

成功后自动重新 decode `Layer.TextSource`，下次读取反映新值。下一次 `Project.WriteAEP` 持久化。

用 [`aep.TextEncodedByteLen`](text.md#aeptextencodedbytelen) 预测候选字符串编码字节数。

#### Returns

`error`；不是文本图层、原字符串偏移没记录、或新旧字节数不一致都返回错误。

#### Example

```go
layer := proj.Compositions[0].Layers[1]
if layer.TextSource == nil {
    log.Fatal("not a text layer")
}

candidate := "FINAL"
if aep.TextEncodedByteLen(candidate) != aep.TextEncodedByteLen(layer.TextSource.Text) {
    log.Fatalf("won't fit: candidate=%d bytes, original=%d bytes",
        aep.TextEncodedByteLen(candidate),
        aep.TextEncodedByteLen(layer.TextSource.Text))
}
layer.SetText(candidate)
```

---

## Flag-bit setters

下面这一组都是 length-preserving 的单 bit 翻转（mutates ldta 字节）。每个 setter 同时更新对应 Go 字段，下一次 `Project.WriteAEP` 持久化。

如果 Layer 是在 parser 之外手工构造（没有 owning ldta chunk），所有 setter 返回 `error` 而不是 panic。

### Layer.SetVisible

```go
func (l *Layer) SetVisible(v bool) error
```

切换 layer 的视频开关（👁️）。

```go
layer.SetVisible(false) // 隐藏图层
```

### Layer.SetSolo

```go
func (l *Layer) SetSolo(v bool) error
```

切换 Solo 开关。

### Layer.SetShy

```go
func (l *Layer) SetShy(v bool) error
```

切换 Shy 开关（在 Shy 过滤视图下隐藏）。

### Layer.SetLocked

```go
func (l *Layer) SetLocked(v bool) error
```

切换 Lock 开关（🔒）。锁定后 AE UI 拒绝该图层的编辑；AEP 文件本身仍可写。

### Layer.SetEffectsEnabled

```go
func (l *Layer) SetEffectsEnabled(v bool) error
```

切换特效渲染开关（fx）。

### Layer.SetMotionBlur

```go
func (l *Layer) SetMotionBlur(v bool) error
```

切换该图层的运动模糊开关。

### Layer.SetAudioEnabled

```go
func (l *Layer) SetAudioEnabled(v bool) error
```

切换音频开关。

### Layer.SetFrameBlendEnabled

```go
func (l *Layer) SetFrameBlendEnabled(v bool) error
```

切换帧混合开关。

### Layer.SetCollapseTransform

```go
func (l *Layer) SetCollapseTransform(v bool) error
```

切换"Collapse Transformations"（pre-comp 用）/"Continuously Rasterize"（矢量 / 形状图层用）。

### Layer.SetIs3D

```go
func (l *Layer) SetIs3D(v bool) error
```

切换 3D Layer 开关。

### Layer.SetIsAdjust

```go
func (l *Layer) SetIsAdjust(v bool) error
```

切换 Adjustment Layer 标记。

### Layer.SetIsGuide

```go
func (l *Layer) SetIsGuide(v bool) error
```

切换 Guide Layer 标记（在 comp viewer 渲染但不进 output）。

### Layer.SetIsNull

```go
func (l *Layer) SetIsNull(v bool) error
```

切换 Null Object 标记位。AE UI 没有直接开关；通常 Null 在创建时就是 null。post-hoc 翻转该位可能产生不常见的 AE 行为，谨慎使用。

### Layer.SetMarkersLocked

```go
func (l *Layer) SetMarkersLocked(v bool) error
```

切换"Lock markers"。

### Layer.SetSamplingBicubic

```go
func (l *Layer) SetSamplingBicubic(v bool) error
```

切换采样模式。`false` = Bilinear，`true` = Bicubic。

### Layer.SetFrameBlendPixelMotion

```go
func (l *Layer) SetFrameBlendPixelMotion(v bool) error
```

切换帧混合类型。`false` = Frame Mix，`true` = Pixel Motion。仅当 `FrameBlendEnabled` 也为 true 时生效。

---

## Byte-field setters

更宽的单字节 / 双字节 enum 写回，也都 length-preserving。

### Layer.SetBlendingMode

```go
func (l *Layer) SetBlendingMode(m BlendingMode) error
```

写入新的 BlendingMode 枚举字节（ldta `@0x63`）。完整 enum 列表见 [constants.md](constants.md#blendingmode)。

```go
layer.SetBlendingMode(aep.BlendingModeMultiply)
```

### Layer.SetTrackMatte

```go
func (l *Layer) SetTrackMatte(t TrackMatteType) error
```

写入新的 TrackMatte 类型字节（ldta `@0x6B`）。可选值见 [constants.md](constants.md#trackmattetype)。**只改模式**，不动 source 指针 —— 旧文件用隐式"上一层"约定时合适。要同时指定显式 source 用 [`SetTrackMatteLayer`](#layersettrackmattelayer)。

```go
layer.SetTrackMatte(aep.TrackMatteAlphaInverse)
```

### Layer.SetTrackMatteLayer

```go
func (l *Layer) SetTrackMatteLayer(sourceID uint32, mode TrackMatteType) error
func (l *Layer) ClearTrackMatteLayer() error
```

一次写完 AE 23+ 显式 matte 关系：`sourceID` → ldta `@0xA0`（4 字节 BE），`mode` → ldta `@0x6B`（1 字节）。length-preserving。

- `sourceID` 必须解到同 comp 内的层（AE 不允许跨 comp matte）。`sourceID == l.ID` 拒绝（self-matte）。
- 传 `sourceID = 0`、`mode = TrackMatteNone` 清除关系；或直接调 `ClearTrackMatteLayer()` 是这俩参数的语法糖。
- 也支持 "保留 source 指针但暂时关闭 matte 渲染"：`sourceID != 0`, `mode = TrackMatteNone` —— AE 接受。
- **AE 22 / 2020 文件不可写**：那些 ldta 长度只有 160 字节，`@0xA0` 不存在 —— 本 setter 返回 error。先在 AE 23+ 打开重存一次让 ldta 加长到 164。

```go
// 把 mt_layer 设置为用 src_layer 作 ALPHA matte
mtLayer.SetTrackMatteLayer(srcLayer.ID, aep.TrackMatteAlpha)

// 清除
mtLayer.ClearTrackMatteLayer()
```

### Layer.SetAlternateSource

```go
func (l *Layer) SetAlternateSource(item AVItem) error
func (l *Layer) ClearAlternateSource() error
```

把 layer 的 media-replacement override 改为指向 `item`（写 blsi 4 字节 BE = `item.ItemID()`）。length-preserving。

- 传 `nil`（或 id 为 0 的 item）= 清除 override，等价 `ClearAlternateSource()`。
- 要求 layer 已经有 EGP slot（`HasAlternateSourceSlot() == true`），否则报错 —— AE 仅当 source-side layer 被 promote 后才写这些 chunks，缺 slot 时插入会破坏 length-preserving 约束。
- 当 layer 挂在 parsed project 上时，`item.ItemID()` 会按 project AV item 校验；找不到返回错误。
- 实现 `AVItem` interface 的类型：`*Composition`、`*Footage`。Folder 不是 AV item，不支持。

```go
src := proj.CompositionByName("BackupSource")
if err := layer.SetAlternateSource(src); err != nil {
    log.Fatal(err)
}
// 或清除：
layer.ClearAlternateSource()
```

> **AE 端注意**：脚本 `setAlternateSource(item)` 会自动包一层 wrapper precomp（见上方 `AlternateSource` 说明）。我们的 setter 不会包装 —— 直接把 `item.ItemID()` 写进 blsi。如果你要复现 AE 的 wrapper 行为，自己先建 wrapper comp。

### Layer.SetLightKind

```go
func (l *Layer) SetLightKind(k LightKind) error
```

写 light 层的 type（ldta `@0x88`，4 字节 BE）。length-preserving。仅 `Layer.Type == LayerTypeLight` 才有意义；其他层会接受写入但视觉上无影响。

```go
spot := comp.Layers[1] // 假设是 light layer
spot.SetLightKind(aep.LightKindAmbient)
```

> AE 22 / 2020 文件 ldta 也是 ≥ 0x8C 长度，所以此 setter 在旧 fixture 上也工作。

### Layer.AlternateSourceID / AlternateSource / HasAlternateSourceSlot

```go
AlternateSourceID uint32                       // AVItem id, 0 = no override
func (l *Layer) AlternateSource() AVItem        // resolved *Composition / *Footage
func (l *Layer) HasAlternateSourceSlot() bool   // true 当 layer 已被 promoted 到 Essential Properties
```

AE 18+ 的 Media Replacement / Essential Properties 覆盖源。`AlternateSourceID` 是覆盖目标 AVItem 的 id；0 = 未覆盖（要么 layer 没 EGP slot，要么有但没 set）。`AlternateSource()` 把 id resolve 成 `*Composition` 或 `*Footage`（接口 `AVItem`），找不到 / 无 project / id=0 都返回 `nil`。

`HasAlternateSourceSlot()` 报告 layer 是否有 EGP media-replacement 槽位 —— AE 仅在 source-side layer 被 `AVLayer.addToMotionGraphicsTemplateAs()` promote 后才会写出这个 chunk pattern。读 / 写 via [`SetAlternateSource`](#layersetalternatesource).

```go
for _, l := range comp.Layers {
    if alt := l.AlternateSource(); alt != nil {
        fmt.Printf("%s -> alt source: %s (id=%d)\n", l.Name, alt.ItemName(), alt.ItemID())
    }
}
```

> **AE 包装注意**：通过 AE 脚本 `setAlternateSource(item)` 设置时，AE 会**自动把目标包一层 wrapper precomp**（名字像 `<slotName>_<originalName> 2`，归到自动建的 "媒体替换合成" 文件夹）。我们的 parser 暴露的就是 AE 持久化的 id —— 通常指向 wrapper，而不是脚本原始传入的 item。

### Layer.TimeRemap / TimeRemapEnabled

```go
func (l *Layer) TimeRemap() *Property
func (l *Layer) TimeRemapEnabled() bool
```

`TimeRemap()` 返回 "ADBE Time Remapping" property（AE 总是写这个槽位，即使 disabled，所以 nil 一般只在非 AV 层上出现）。

`TimeRemapEnabled()` 判断是否启用：等价于 `len(TimeRemap().Keyframes) > 0`。AE 启用 timeRemap 时会自动加 2 个 identity keyframe。

```go
if layer.TimeRemapEnabled() {
    fmt.Println("kf 数:", len(layer.TimeRemap().Keyframes))
}
```

> **写**：未实现。toggle timeRemap on/off 是结构性变更（AE 在属性树里加 / 删 2 个 keyframe），破坏 length-preserving 约束，标 `❌ structural`。

### Layer.SetLabel

```go
func (l *Layer) SetLabel(index uint8) error
```

写入新的时间线色卡索引（ldta `@0x3D`）。AE 默认色卡 palette 0..16；超出范围的索引会被原样写入但 AE 显示 0 号色。

```go
layer.SetLabel(11) // Orange
```

### Layer.SetQuality

```go
func (l *Layer) SetQuality(q LayerQuality) error
```

写入新的渲染质量枚举（ldta `@0x04`，uint16 BE）。可选值：`LayerQualityWireframe` / `Draft` / `Best`。

```go
layer.SetQuality(aep.LayerQualityBest)
```

### Layer.SetPreserveTransparency

```go
func (l *Layer) SetPreserveTransparency(v bool) error
```

切换 "Preserve Underlying Transparency"（ldta `@0x67`，单字节 0/1）。

### Layer.SetParent

```go
func (l *Layer) SetParent(parentID uint32) error
```

改父图层 ID（ldta `@0x84`，4 字节 length-preserving）。

- `parentID == 0` 清除父（AE 显示 "None"）
- `parentID == l.ID` 拒绝（自-引用）
- 当 Layer 经 parser 创建且 owning comp 已知，`parentID` 必须能在**同 comp** 内解析到一个 layer，否则返回 error；AE 不允许跨 comp 父级

```go
layer.SetParent(targetLayer.ID)
layer.SetParent(0) // 清除父
```

### Layer.SetSource

```go
func (l *Layer) SetSource(sourceID uint32) error
```

改源项目项 ID（ldta `@0x28`，4 字节 length-preserving）。

- `sourceID == 0` 罕见 — 等价 "无源"（AE 通常用 Null layer）
- 当 Layer 经 parser 创建且 owning project 已知，`sourceID` 必须能解析到一个 `Composition` 或 `Footage` item

```go
// 把图层重定向到另一个 pre-comp
layer.SetSource(otherComp.ID)
```

> 注意：AE 在渲染时从源派生 Width/Height 等数据，这些不在 ldta 里 — 改 source 不改 ldta 其余字节。

### Layer.SetAutoOrient

```go
func (l *Layer) SetAutoOrient(t AutoOrientType) error
```

写自动朝向枚举。底层是 ldta `@0x25` / `@0x26` 中 3 个互斥位 —— setter 先清空全 3 位再设置目标位：

| 枚举值 | 位 |
|---|---|
| `AutoOrientNone` | 全 0 |
| `AutoOrientAlongPath` | `@0x26` bit 0 |
| `AutoOrientCameraOrPointOfInterest` | `@0x26` bit 5 |
| `AutoOrientCharactersTowardCamera` | `@0x25` bit 4 |

length-preserving (2 bytes touched)。

```go
layer.SetAutoOrient(aep.AutoOrientCameraOrPointOfInterest) // 摄像机面向
```

---

## Time-field setters

ldta 里的时间字段都是 `dividend / divisor` int/uint pair（typically divisor = 600）。每个 setter 写 8 字节 length-preserving，复用现有 divisor（缺失时回退到 600 / 100）。

### Layer.SetStartTime

```go
func (l *Layer) SetStartTime(seconds float64) error
```

改图层在合成时间轴上的起始时间（ldta `@0x0C/@0x10`）。AE 支持负值（pre-roll）。

```go
layer.SetStartTime(2.5)
```

### Layer.SetInPoint

```go
func (l *Layer) SetInPoint(seconds float64) error
```

改 source-media 的 in-point（ldta `@0x14/@0x18`）。同时自动重算 `Layer.Duration = OutPoint − InPoint`。

### Layer.SetOutPoint

```go
func (l *Layer) SetOutPoint(seconds float64) error
```

改 source-media 的 out-point（ldta `@0x1C/@0x20`）。同时自动重算 `Layer.Duration`。

### Layer.SetStretch

```go
func (l *Layer) SetStretch(ratio float64) error
```

改时间伸缩系数（1.0 = 原速度，2.0 = 2× 慢放）。dividend 在 ldta `@0x08`，divisor 在 `@0x6C`（**split across ldta**，不是相邻 8 字节）。divisor 缺失时回退 100（AE 默认）。

---

## Text-field setters (length-variable)

下面两个 setter 改的是 Utf8 / cmta chunk 内容 —— **不**受 length-preserving 限制。`WriteAEP` 序列化时会重算父 LIST size。

### Layer.SetName

```go
func (l *Layer) SetName(newName string) error
```

改图层名（Utf8 chunk 的整体替换）。

```go
layer.SetName("Main BG")
```

如果原 layer 没有 Utf8 chunk（极少；通常 AE 都会写），返回 error。

### Layer.SetComment

```go
func (l *Layer) SetComment(comment string) error
```

改图层备注（cmta chunk）。如果原 layer 没有 cmta chunk，本方法会**自动插入**一个新的到 Layr LIST 的末尾。LF 自动转 CRLF + NUL 终止符（与 AE 写出格式一致）。

```go
layer.SetComment("第一行\n第二行")
```

---

## Camera Layer accessors

只对 `Layer.Type == LayerTypeCamera` 有意义；其他图层调用全部返回 nil。每个 accessor 都是 `PropertyByMatchName` 的便捷包装，返回的 `*Property` 可读 `StaticValue` / `Keyframes` 或调 `SetStaticValue` / Keyframe setters 写回。

| Accessor | MatchName | 含义 |
|---|---|---|
| `CameraZoom()` | `"ADBE Camera Zoom"` | Zoom（像素，等价 35mm 焦距换算） |
| `CameraDepthOfField()` | `"ADBE Camera Depth of Field"` | DoF 总开关（0/1） |
| `CameraFocusDistance()` | `"ADBE Camera Focus Distance"` | 焦距（像素） |
| `CameraAperture()` | `"ADBE Camera Aperture"` | 光圈大小（像素） |
| `CameraBlurLevel()` | `"ADBE Camera Blur Level"` | 模糊程度（%） |
| `IrisShape()` | `"ADBE Iris Shape"` | 光圈形状菜单（1=Fast Rect, 3..10=Triangle..Decagon） |
| `IrisRotation()` | `"ADBE Iris Rotation"` | 光圈旋转（度） |
| `IrisRoundness()` | `"ADBE Iris Roundness"` | 光圈圆度（%） |
| `IrisAspectRatio()` | `"ADBE Iris Aspect Ratio"` | 光圈宽高比 |
| `IrisDiffractionFringe()` | `"ADBE Iris Diffraction Fringe"` | 光圈衍射边缘（%） |
| `IrisHighlightGain()` | `"ADBE Iris Highlight Gain"` | 高光增益 |
| `IrisHighlightThreshold()` | `"ADBE Iris Highlight Threshold"` | 高光阈值（0..1 归一化亮度） |
| `IrisHighlightSaturation()` | `"ADBE Iris Highlight Saturation"` | 高光饱和度 |

```go
cam := comp.Layers[0]
if cam.Type == aep.LayerTypeCamera {
    if z := cam.CameraZoom(); z != nil {
        fmt.Println("zoom:", z.StaticValue)
    }
    if fd := cam.CameraFocusDistance(); fd != nil && len(fd.Keyframes) == 0 {
        fd.SetStaticValue(2000.0) // 改焦距
    }
}
```

每个 getter 都有配对的 **typed setter**（除 LightType — 多数情况 property 不存在）：`SetCameraZoom(v float64)` / `SetCameraDepthOfField(enabled bool)` / `SetCameraFocusDistance(v) / SetCameraAperture(v) / SetCameraBlurLevel(v)` / `SetIrisShape(v) / SetIrisRotation(v) / SetIrisRoundness(v) / SetIrisAspectRatio(v) / SetIrisDiffractionFringe(v) / SetIrisHighlightGain(v) / SetIrisHighlightThreshold(v) / SetIrisHighlightSaturation(v)`。每个 setter 内部 nil-check + 委托给 `Property.SetStaticValue`，property 缺失（非 camera 层 / Iris 子树未启用）时返回 `"layer %q: %s property not present"` 错误。

```go
if err := cam.SetCameraZoom(2000); err != nil { /* not a camera */ }
if err := cam.SetCameraDepthOfField(true); err != nil { /* not a camera */ }
```

---

## Light Layer accessors

只对 `Layer.Type == LayerTypeLight` 有意义；其他图层调用全部返回 nil。

| Accessor | MatchName | 含义 |
|---|---|---|
| `LightType()` | `"ADBE Light Type"` | ⚠ 多数情况返回 nil — AE 把灯光 type 存在 layer 级 metadata 里，不是 property 树。本方法保留以备未来 / 旧 AE 版本 |
| `LightColor()` | `"ADBE Light Color"` | 灯光颜色（4D，AE 2020 存 `[A, R, G, B]` 0..255） |
| `LightIntensity()` | `"ADBE Light Intensity"` | 强度（%） |
| `LightConeAngle()` | `"ADBE Light Cone Angle"` | 聚光灯锥角（度） |
| `LightConeFeather()` | `"ADBE Light Cone Feather 2"` | 聚光灯锥羽化（%） |
| `LightFalloffType()` | `"ADBE Light Falloff Type"` | 衰减类型菜单 |
| `LightFalloffStart()` | `"ADBE Light Falloff Start"` | 衰减起始距离（像素） |
| `LightFalloffDistance()` | `"ADBE Light Falloff Distance"` | 衰减距离（像素） |
| `LightCastsShadows()` | `"ADBE Casts Shadows"` | 可读到属性（property 树里存在；通过 `StaticValue == 1` 判断 on）。AE 24+ 确认非 layer-level flag |
| `LightShadowDarkness()` | `"ADBE Light Shadow Darkness"` | 阴影暗度（%） |
| `LightShadowDiffusion()` | `"ADBE Light Shadow Diffusion"` | 阴影扩散（像素） |

**光类型走 layer-level `Layer.LightKind` 字段**（ldta `@0x88`，AE 23+ RE'd via `re_wave2_ae24.aep`）：

```go
type LightKind uint8

const (
    LightKindParallel LightKind = 0
    LightKindSpot     LightKind = 1
    LightKindPoint    LightKind = 2
    LightKindAmbient  LightKind = 3
)

func (k LightKind) String() string  // "parallel" / "spot" / "point" / "ambient"
```

读：`layer.LightKind`（仅 `layer.Type == LayerTypeLight` 时有意义；非 light 层都是 0）。写：[`Layer.SetLightKind(kind)`](#layersetlightkind)（4 字节 BE，length-preserving）。

> 旧 `LightTypeID` 常量（4101..4104）已 deprecated —— 那些值是占位符，AE 实际不写。新代码请用 `LightKind`。

```go
spot := comp.Layers[1]
if spot.Type == aep.LayerTypeLight {
    if intensity := spot.LightIntensity(); intensity != nil {
        fmt.Println("intensity:", intensity.StaticValue)
    }
    if cone := spot.LightConeAngle(); cone != nil && len(cone.Keyframes) == 0 {
        cone.SetStaticValue(45.0) // 收紧聚光锥
    }
}
```

**Typed setter** 配对：`SetLightColor([]float64)` (3 或 4 分量，Component 数随 AE 版本) / `SetLightIntensity(v)` / `SetLightConeAngle(v) / SetLightConeFeather(v)` / `SetLightFalloffType(v) / SetLightFalloffStart(v) / SetLightFalloffDistance(v)` / `SetLightCastsShadows(enabled bool)` / `SetLightShadowDarkness(v) / SetLightShadowDiffusion(v)`。

```go
spot.SetLightColor([]float64{200, 100, 50, 255})   // RGBA 0..255
spot.SetLightIntensity(120)
spot.SetLightCastsShadows(true)
```

---

## Transform group accessors

每个 layer 共有 Transform 组：Anchor Point / Position / Scale / Rotation (Z-axis) / Opacity 是 2D + 3D 共通；Rotate X / Rotate Y / Orientation 仅 3D 层。

| Getter | Setter | 含义 |
|---|---|---|
| `AnchorPoint()` | `SetAnchorPoint([]float64)` | 锚点（像素，2D 或 3D 由 Is3D 决定） |
| `Position()` | `SetPosition([]float64)` | 位置（像素） |
| `Scale()` | `SetScale([]float64)` | 缩放（归一化 1.0 = 100%） |
| `Rotation()` | `SetRotation(deg float64)` | Z-轴旋转（度） |
| `RotateX()` | `SetRotateX(deg float64)` | X-轴旋转（度，仅 3D） |
| `RotateY()` | `SetRotateY(deg float64)` | Y-轴旋转（度，仅 3D） |
| `Orientation()` | `SetOrientation([]float64)` | 3D 朝向（3 分量度数，仅 3D） |
| `Opacity()` | `SetOpacity(v float64)` | 不透明度（归一化 0..1） |
| `AudioLevels()` | `SetAudioLevels([]float64)` | 音频电平 `[L, R]`（dB；仅含音频图层） |

```go
cam.SetPosition([]float64{1000, 500, -800})  // 3D 镜头
cam.SetRotation(45)                           // 滚转 45°
cam.SetOpacity(0.8)
```

非 3D 层调 `SetRotateX` / `SetRotateY` / `SetOrientation` 返回 `"property not present"` 错误。

---

## Material Options accessors (3D AV layer)

3D-enabled 的 AV layer（solid / footage / precomp / shape / text + Is3D=true）有完整的 "Material Options" 子树，控制阴影 / 光照 / 反射 / 折射等物理光照参数。17 个 typed getter + 17 个 setter（部分 setter 用 `bool` / 三态 enum 增强类型安全）：

| Getter | Setter | 默认 | 含义 |
|---|---|---|---|
| `MaterialCastsShadows()` | `SetMaterialCastsShadows(MaterialCastsShadowsMode)` | Off | 投射阴影（**tri-state**: `MaterialCastsOff / On / Only`；Only = 阴影显示但 layer 自身隐藏） |
| `MaterialLightTransmission()` | `SetMaterialLightTransmission(v)` | 0 | 透光（0..1） |
| `MaterialAcceptsShadows()` | `SetMaterialAcceptsShadows(bool)` | true | 接受阴影 |
| `MaterialAcceptsLights()` | `SetMaterialAcceptsLights(bool)` | true | 接受光照 |
| `MaterialShadowColor()` | `SetMaterialShadowColor([]float64)` | 黑 | 阴影色（RGBA） |
| `MaterialAppearsInReflections()` | `SetMaterialAppearsInReflections(bool)` | true | 出现在反射中 |
| `MaterialAmbient / Diffuse / Specular / Shininess / Metal / Reflection / Glossiness / Fresnel / Transparency / TranspRolloff / IndexOfRefraction` | 同名 `Set*(v float64)` | 见 AE | 物理光照系数 |

```go
if !cube.Is3D { return }
cube.SetMaterialCastsShadows(aep.MaterialCastsOn)
cube.SetMaterialReflection(0.4)        // 反光度
cube.SetMaterialIndexOfRefraction(1.5) // 玻璃 IoR
cube.SetMaterialAcceptsShadows(true)
```

**注意**：`ADBE Casts Shadows` 同 match name 在 light layer 上是双态 (on/off)。`Layer.LightCastsShadows()` 跟 `Layer.MaterialCastsShadows()` 返回同一 `*Property` — getter 等价，只是命名表达上下文意图。Setter `SetLightCastsShadows(bool)` 只支持 on/off；要 "Only" 模式必须用 `SetMaterialCastsShadows(MaterialCastsOnly)`。

**AE 持久化陷阱**：material option 设回默认值时 AE 会**裁掉**这条 property 不序列化（精简文件大小）。所以 fresh-from-AE 的 fixture 上，default-valued material option 可能 getter 返回 nil。Setter 会报 `"property not present"` —— 这种情况说明用户得先在 AE 里手动改一次让 AE emit slot。
