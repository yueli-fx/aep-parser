# aep-parser API 文档

Go 原生 Adobe After Effects `.aep` 项目解析 + length-preserving 写回库。无需 AE 运行实例。

风格参考 [docsforadobe/after-effects-scripting-guide](https://github.com/docsforadobe/after-effects-scripting-guide) — 每个核心类型一个 md 文件，列出字段（read-only / read-write）+ 方法 + 示例。

---

## 类层级

```text
Project
├── Compositions []*Composition
│   ├── Layers []*Layer
│   │   ├── Properties []*Property       (含 Keyframes []*Keyframe)
│   │   ├── Effects    []*Effect
│   │   ├── Markers    []*Marker
│   │   ├── Masks      []*Mask
│   │   ├── ShapePaths []*ShapePath
│   │   └── TextSource *TextSource       (含 Runs / Paragraphs)
│   └── Markers []*Marker                (合成级 marker)
├── Footage []*Footage
├── Folders []*Folder
└── Warnings []string                    (非致命解析异常)
```

## 分类导航

| 文件 | 涵盖 |
|---|---|
| [project.md](project.md) | `Project`, `Folder`, 入口函数 `Open` / `FromReader` |
| [composition.md](composition.md) | `Composition` |
| [footage.md](footage.md) | `Footage` |
| [layer.md](layer.md) | `Layer`（基础 / 渲染 / 子集合 / Transform 访问器 / `SetText`） |
| [property.md](property.md) | `Property`, `Keyframe`, `TemporalEase`, `InterpType` |
| [effect.md](effect.md) | `Effect` |
| [marker.md](marker.md) | `Marker`（图层与合成共用） |
| [mask.md](mask.md) | `Mask`, `MaskVertex`, `MaskPathKeyframe`, `MaskMode` |
| [shape.md](shape.md) | `ShapePath` |
| [text.md](text.md) | `TextSource`, `TextStyleRun`, `TextParagraph`, `TextJustification`, `SetText`, `TextEncodedByteLen` |
| [json.md](json.md) | JSON 视图（`JSON*` 类型 + `ToJSON` / `MarshalJSON` / `WriteJSON`） |
| [constants.md](constants.md) | 枚举常量速查（`LayerType`, `BlendingMode`, `TrackMatteType`, `AutoOrientType`, `LayerQuality`, `MaskMode`, `InterpType`, `TextJustification`, `BitsPerChannel`） |
| [../workshop/plans/coverage.md](../workshop/plans/coverage.md) | **覆盖度概览（精简）** —— 已 ship / 暂搁 / 不可达 / 下一步候选 |
| [../workshop/plans/coverage-detail.md](../workshop/plans/coverage-detail.md) | **AE attr 详细交叉表** —— 对照 AE 脚本指南逐项标注 ✅R/W / 🟢R / ❌ 状态 |

---

## 五分钟上手

### 读取

```go
import aep "github.com/example/aep-parser/internal/aep"

proj, err := aep.Open("template.aep")
if err != nil { panic(err) }

for _, comp := range proj.Compositions {
    fmt.Printf("%s  %dx%d  %.2ffps  %.2fs\n",
        comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration)
    for _, layer := range comp.Layers {
        fmt.Printf("  [%d] %s (%s)\n", layer.Index+1, layer.Name, layer.Type)
    }
}

// 解析时的非致命异常
for _, w := range proj.Warnings {
    fmt.Fprintln(os.Stderr, "warn:", w)
}
```

### 写回

所有 `Set*` API 都是 length-preserving（不改字节数）。修改后调一次 `Project.WriteAEP(w)` 落盘。

```go
proj, _ := aep.Open("in.aep")

// 改素材路径
proj.Footage[0].SetPath(`D:\new\file.png`)

// 改关键帧值 / 时间
pos := proj.Compositions[0].Layers[0].Position()
pos.Keyframes[0].SetValue([]float64{960, 540, 0})
pos.Keyframes[0].SetTime(0.0)

// 改静态属性值（无关键帧时）
op := proj.Compositions[0].Layers[0].Opacity()
op.SetStaticValue(0.5)

// 改文本字符串（同字节数才接受）
tl := proj.Compositions[0].Layers[1]
if tl.TextSource != nil && aep.TextEncodedByteLen("FINAL") == aep.TextEncodedByteLen(tl.TextSource.Text) {
    tl.SetText("FINAL")
}

// 改图层开关 / 渲染设置（length-preserving 位 / 字节翻转）
layer := proj.Compositions[0].Layers[0]
layer.SetVisible(false)
layer.SetBlendingMode(aep.BlendingModeMultiply)
layer.SetLabel(11)

f, _ := os.Create("out.aep")
defer f.Close()
proj.WriteAEP(f)
```

### JSON 导出（单向）

```go
proj.WriteJSON(os.Stdout)
// 或：b, _ := json.MarshalIndent(proj, "", "  ")
```

> ⚠️ **没有 `ReadJSON`** —— JSON 仅作只读快照。任何修改都必须走二进制路径（`Open` → `Set*` → `WriteAEP`）。

---

## 全局约束

- **Length-preserving 是硬约束**：除 `Footage.SetPath` 外，所有 `Set*` 改写的字节数必须等于原 chunk 字节数。增删结构（图层 / 关键帧 / 顶点）目前不支持。
- **TickRate 是 per-composition 的**：不要假设统一 8000。每个 `Composition.TickRate` 由 cdta 解出，关键帧/marker 时间换算用它。
- **不安全的并发**：所有类型共享底层 RIFX chunk 字节。多 reader 可以；任何 `Set*` 写入需要调用方自己同步。
- **`Project.Warnings`**：non-nil 表示遇到非致命解析异常。空切片 = clean parse。

---

## 操作能力 / 边界一览

> 本表是高频能力的速查，**不是穷举清单**。AE 全字段覆盖矩阵以 [../workshop/plans/coverage-detail.md](../workshop/plans/coverage-detail.md) 为权威 —— 每个 AE 脚本字段都有 ✅R/W / 🟢R / ⚠ / ❌ 标注 + 引用代码位置。

| 类别 | 操作 | 状态 |
| --- | --- | --- |
| **结构性** | 增删 keyframe | ✅ `Property.InsertKeyframe(time, value)` / `DeleteKeyframe(i)` — 自动 ldat 重排 + lhd3 count 同步 |
| **结构性** | 增删图层 / 属性 / 特效 / mask vertex / shape primitive | ❌ 需重排多个父 chunk 字节，破坏 length-preserving 不变量 |
| **结构性** | `ReadJSON` 反序列化 | ❌ 设计如此 —— JSON 仅作只读快照，所有修改走 `Open` → `Set*` → `WriteAEP` |
| **结构性** | 修改 Mask 顶点 / Shape 路径顶点 | ❌ 未实现 |
| **静态值 / 表达式** | 改静态属性值 / 关键帧值 / 时间 / ease / 切线 | ✅ `Property.SetStaticValue` / `Keyframe.SetValue` / `SetTime` / `SetIn/OutInterp` / `SetIn/OutTemporalEase` / `SetIn/OutSpatialTangent` |
| **静态值 / 表达式** | 表达式写回（创建 / 替换 / 清空 / 启用切换） | ✅ `Property.SetExpression`（length-variable）+ `SetExpressionEnabled` —— [property.md](property.md#propertysetexpression) |
| **Layer 标志位** | Visible / Solo / Shy / Locked / 3D / Null / Adjust / Guide / MotionBlur / EffectsEnabled / AudioEnabled / FrameBlend / CollapseTransform / MarkersLocked / Sampling / FrameBlendMode | ✅ 共 16 个 `Layer.Set*` flag-bit setter —— [layer.md](layer.md#flag-bit-setters) |
| **Layer 字节字段** | BlendingMode / TrackMatte / Label / Quality / PreserveTransparency / AutoOrient | ✅ `Layer.SetBlendingMode` / `SetTrackMatte` / `SetLabel` / `SetQuality` / `SetPreserveTransparency` / `SetAutoOrient` |
| **Layer 字节字段** | TrackMatteLayer (AE 23+，ldta `@0xA0` 显式 source) | ✅ `Layer.TrackMatteLayerID` / `TrackMatteLayer()` / `SetTrackMatteLayer` / `ClearTrackMatteLayer` |
| **Layer 字节字段** | LightKind (AE 23+，ldta `@0x88`) | ✅ `Layer.LightKind` / `SetLightKind` —— 取代旧 `LightTypeID`（已 deprecated） |
| **Layer 字节字段** | Parent / Source ID / StartTime / InPoint / OutPoint / Stretch | ✅ `Layer.SetParent` / `SetSource` / `SetStartTime` / `SetInPoint` / `SetOutPoint` / `SetStretch` |
| **Layer 文本字段** | Name / Comment | ✅ `Layer.SetName` / `SetComment`（length-variable）—— [layer.md](layer.md#text-field-setters-length-variable) |
| **Layer Essential Properties** | Media Replacement 替换源（AE 18+） | ✅ `Layer.AlternateSourceID` / `AlternateSource()` / `HasAlternateSourceSlot()` / `SetAlternateSource(AVItem)` / `ClearAlternateSource()` —— 需 layer 已有 EGP slot；AE 脚本 quirk：`setAlternateSource(item)` 自动包 wrapper precomp |
| **Composition** | 元数据 BGColor / Shutter / MotionBlur / WorkArea / FrameRate / Duration / PixelAspect / DisplayStartTime | ✅ `Composition.SetBGColor` / `SetShutterAngle/Phase` / `SetMotionBlur*` / `SetWorkArea` / `SetFrameRate` / `SetDuration` / `SetPixelAspect` / `SetDisplayStartTime` / `SetDisplayStartFrame` —— [composition.md](composition.md) |
| **Composition** | Name / cdta 标志位（Draft3D / HideShyLayers / CompMotionBlur / FrameBlending / PreserveNestedFrameRate / PreserveNestedResolution） | ✅ `Composition.SetName` + 6 个 cdta flag setter |
| **Item-level** | Comment / Label（Composition / Footage 共用） | ✅ `Composition.SetComment` / `SetLabel` + `Footage.SetComment` / `SetLabel` |
| **Footage** | 素材路径 | ✅ `Footage.SetPath`（**唯一不 length-preserving 的 API** —— 用整 chunk 替换） |
| **Marker** | Time / Duration / Label / Comment / Chapter / URL / FrameTarget / CuePointName（图层 marker + 合成 marker 共用） | ✅ `Marker.SetTime` / `SetDuration` / `SetLabel` / `SetComment` / `SetChapter` / `SetURL` / `SetFrameTarget` / `SetCuePointName` —— [marker.md](marker.md#setters) |
| **Mask** | Mode / Inverted / Color / Closed / Locked / MaskMotionBlur | ✅ `Mask.SetMode` / `SetInverted` / `SetColor` / `SetClosed` / `SetLocked` / `SetMaskMotionBlur` |
| **Project** | BitsPerChannel (nhed + nnhd 双写) | ✅ `Project.SetBitsPerChannel` |
| **文本字符串** | 整段文本替换 | ⚠ `Layer.SetText` 只接受**同字节数**（length-preserving 硬约束）；用 `TextEncodedByteLen` 预判 |
| **文本 per-run** | FontSize / FillColor / Stroke{Color,Width,Apply} / Tracking / Leading / AutoLeading / BaselineShift / HorizontalScale / VerticalScale / Tsume / FauxBold / FauxItalic / FontIndex + 字体表 `AddFont` | ✅ 共 16 个 `Layer.SetRun*` setter（length-variable PostScript splice）+ `Layer.AddFont(name)` —— [text.md](text.md#per-run-setters) |
| **文本 per-run (AE 24+)** | CapsOption / BaselineOption (superscript/subscript 镜像) / StrokeOverFill / AutoKernType / NoBreak / LineJoinType / DigitSet | ✅ `Run.{CapsOption,BaselineOption,StrokeOverFill,AutoKernType,NoBreak,LineJoinType,DigitSet}` + 7 个 `Layer.SetRun*` setter |
| **文本 per-paragraph** | Justification / FirstLineIndent / StartIndent / EndIndent / SpaceBefore / SpaceAfter / AutoHyphenate / LeadingType / HangingRoman / Direction | ✅ `Layer.SetParagraph*` 共 9 个 setter |
| **文本 manual kerning** | per-char kerning array | ✅ R `TextSource.ManualKerning []int` + `Kerning int` (first-char mirror) + W `Layer.SetManualKerning(values []int)`。要求 slot 已 emit（首次启用是结构性，refused）；AutoKernType 由 `SetRunAutoKernType` 单独管 |
| **Camera 专属** | Zoom / DepthOfField / FocusDistance / Aperture / BlurLevel / IrisShape / IrisRotation / IrisRoundness / IrisAspectRatio / IrisDiffractionFringe / IrisHighlight{Gain,Threshold,Saturation} | ✅ 13 个 `Layer.Camera*` / `Iris*` accessor 返回 `*Property` —— [layer.md](layer.md#camera-layer-accessors) |
| **Light 专属** | Color / Intensity / ConeAngle / ConeFeather / FalloffType / FalloffStart / FalloffDistance / CastsShadows / ShadowDarkness / ShadowDiffusion | ✅ 11 个 `Layer.Light*` accessor 返回 `*Property` —— [layer.md](layer.md#light-layer-accessors) |
| **Shape 参数化** | Rect / Star / Ellipse 基元 → Size / Position / Roundness / Star{Type,Points,Rotation,Inner/OuterRadius,Inner/OuterRoundness} | ✅ `Layer.ShapePrimitives []*ShapePrimitive`，每个 primitive 暴露子属性为 `*Property`（与扁平 `Layer.Properties` 共享 chunk 引用） |
| **运行时字段** | `TextDocument.fontLocation` (字体磁盘路径) / `CompItem.dropFrame` | ❌ **runtime-only** —— AE 不持久化到 .aep（前者 runtime 查系统字体注册表得来；后者是 session UI 状态） |
