# TextSource object

`layer.TextSource`（仅当 layer 是文本图层时非 nil）

## Description

文本图层的解码视图。从 `Layer.TextSourceRaw`（btds 原字节）解出 —— 内含一个 PostScript-style 字典（AE 内部叫 CoolType TextEngine document），描述文字内容、字体、样式、段落、box-text 边界等。

```text
Layer
└── TextSource
    ├── Text          string            (用户文字)
    ├── Fonts         []string          (字体名表)
    ├── Runs          []TextStyleRun    (每个字符样式 span)
    ├── Paragraphs    []TextParagraph   (每段样式)
    ├── Justification TextJustification (首段便捷字段)
    ├── IsBoxText     bool              (box / point)
    └── BoxBounds     [4]float64        (仅 box text)
```

文字字符串可通过 [`Layer.SetText`](layer.md#layersettext) length-preserving 写回。per-run / per-paragraph 字段（FontSize / FillColor / Tracking / Leading / BaselineShift / Stroke / Justification / Caps / Baseline / StrokeOverFill / 段缩进等）可通过下面的 [Per-run setters](#per-run-setters) / [Per-paragraph setters](#per-paragraph-setters) 一组 length-variable splice 方法写回。AE 24+ 新引入的 `fontCapsOption` / `fontBaselineOption` / `strokeOverFill` / 段缩进 / `autoHyphenate` 全部覆盖。

## Example

```go
layer := proj.Compositions[0].Layers[1]
if layer.TextSource == nil {
    return // 不是文本图层
}
ts := layer.TextSource

fmt.Printf("Text: %q\n", ts.Text)
fmt.Printf("Fonts: %v\n", ts.Fonts)
fmt.Printf("Justification: %s\n", ts.Justification)
fmt.Printf("Box text: %v %v\n", ts.IsBoxText, ts.BoxBounds)

for i, r := range ts.Runs {
    fmt.Printf("Run %d: font=%q size=%.0f color=%v leading=%g(auto=%v) tracking=%g\n",
        i, r.FontName, r.FontSize, r.FillColor,
        r.Leading, r.AutoLeading, r.Tracking)
    if r.ApplyStroke {
        fmt.Printf("  stroke: color=%v width=%g\n", r.StrokeColor, r.StrokeWidth)
    }
}

// 同字节数才能写回
candidate := "FINAL"
if aep.TextEncodedByteLen(candidate) == aep.TextEncodedByteLen(ts.Text) {
    layer.SetText(candidate)
}
```

---

## Attributes

### TextSource.Text

```go
Text string
```

#### Description

用户可见文字。**多段已合并为 `\n`**；AE 原本每段尾部的 `\r` 已经被剥除。

#### Type

`string`；read（写入用 [`Layer.SetText`](layer.md#layersettext)，length-preserving）。

---

### TextSource.Fonts

```go
Fonts []string
```

PostScript 字体名表（如 `["YouYuan", "Myriad-Roman", "AdobeInvisFont"]`）。AE 通常在文本图层里挂多个字体作为 fallback —— 实际使用哪个由 `Run.FontIndex` 决定。read-only。

---

### TextSource.FontAxes

```go
FontAxes [][]float64
```

跟 `Fonts` 平行：`FontAxes[i]` 是 `Fonts[i]` 的 OpenType variable-font axes 值数组（顺序跟字体注册的 named-axes 一致 —— Bahnschrift = `[wght, wdth]`、Inter Variable = `[wght, slnt]` 等）。非 variable font 时该项为 `nil`。值已从 btdk 里的 16.16 fixed-point 解码成可读浮点（如 `wght=700` 读出来就是 `700.0`）。

存在 `/0/1/0[i]/0/0/4` 数组。AE 把它跟 PostScript 字体名捆绑：你切到 `Bahnschrift-Bold` 这样的 named instance 时，AE 同时改写 `/0` 字符串 + `/4` 数组的对应 axis 值。

read-only。AE ScriptingAPI 没有独立写 axis 值的接口（`TextDocument.fontVariation` 不存在 —— 实测所有 dict-form 赋值无效），可写途径只能是 `SetRunFontIndex` 切到已加载的另一个 named instance。

```go
for i, fontName := range ts.Fonts {
    if axes := ts.FontAxes[i]; len(axes) > 0 {
        fmt.Printf("%s axes: %v\n", fontName, axes) // e.g. Bahnschrift-Bold axes: [700 100]
    }
}
```

---

### TextSource.Runs

```go
Runs []TextStyleRun
```

字符级样式 run 列表。多 style 文本（不同段不同字号 / 颜色）有多个 run；单一样式文本只有一个 run。详见下面 [TextStyleRun](#textstylerun)。read-only。

---

### TextSource.Paragraphs

```go
Paragraphs []TextParagraph
```

每个段落的样式（按文本中 `\n` 分隔）。当前只暴露 `Justification`。详见下面 [TextParagraph](#textparagraph)。read-only。

---

### TextSource.Justification

```go
Justification TextJustification
```

#### Description

首段对齐方式。等价于 `Paragraphs[0].Justification`（如果 Paragraphs 非空）。便捷字段。

#### Type

`TextJustification` (`int`)；read-only。

```go
const (
    TextJustifyLeft   TextJustification = 0
    TextJustifyRight  TextJustification = 1
    TextJustifyCenter TextJustification = 2
)
```

带 `String()` 方法（`"Left"` / `"Right"` / `"Center"`）。

---

### TextSource.IsBoxText

```go
IsBoxText bool
```

#### Description

- `true` —— Box (Paragraph) Text：固定矩形内的文字，超出会换行
- `false` —— Point Text：从插入点向四周扩展的单点文字

#### Type

`bool`；read-only。

---

### TextSource.BoxBounds

```go
BoxBounds [4]float64
```

#### Description

`[xmin, ymin, xmax, ymax]`，图层局部坐标系下的 box text 矩形。仅 `IsBoxText == true` 时有意义；否则为零值。

#### Type

`[4]float64`；read-only。

---

### TextSource.ManualKerning / Kerning

```go
ManualKerning []int    // per-character, length == utf-8 chars in Text
Kerning       int      // first-char mirror, matches AE TextDocument.kerning
```

#### Description

`ManualKerning` 是 per-character 手动 kerning（AE "字符" 面板里的 Kerning 字段，1/1000 em 单位），仅当对应 run 的 `AutoKernType == TextAutoKernNoAuto` 时生效。长度等于 `Text` 的字符数；如果 AutoKernType 全是 Metric / Optical，整个 sub-tree 不存在，本字段为 `nil`。

`Kerning` 是 AE 脚本 `TextDocument.kerning` 的等价物 —— "只反映第一个字符"。来自 btdk `/1/1[0]/0/7`（per-doc scalar，AE 写盘时跟 `/1/1[0]/0/8/0[0]/0/0` 保持一致）。

```go
ts := layer.TextSource
fmt.Printf("first-char kerning: %d\n", ts.Kerning)
for i, k := range ts.ManualKerning {
    fmt.Printf("  char[%d] = %d\n", i, k)
}
```

#### Type

`[]int` / `int`；read-only。写通过 `Layer.SetManualKerning(values []int)` —— 见下。

#### Layer.SetManualKerning

```go
func (l *Layer) SetManualKerning(values []int) error
```

写整段 per-char 手动 kerning 值。

约束：

- `values` 长度必须等于 `len(TextSource.ManualKerning)`（即既有 char 数）。
- 要求 `/1/1[0]/0/8` slot 已存在 —— AE 只在某 run 被设为 `AutoKernType=NoAuto` 且有非零 kerning 值时才会 emit 这棵子树。"首次启用" 是结构性添加，**不支持**：先在 AE 里设一次非零 kerning，再用这个 setter 改值。
- 不动 `AutoKernType`。要让 kerning 真正生效，对应 run 必须是 `TextAutoKernNoAuto`，用 `SetRunAutoKernType(runIdx, TextAutoKernNoAuto)` 单独切换。
- 自动联动写 `/1/1[0]/0/7`（first-char 镜像 = `values[0]`），跟 AE 脚本 `TextDocument.kerning` 行为一致。

变长 splice 安全：每次内部调用都重 parse btdk body，串行写 N+1 个位置时偏移会自动重定位。

```go
// 把 4-char 文本的 kerning 改成 [10, 20, 30, 40]
if err := layer.SetManualKerning([]int{10, 20, 30, 40}); err != nil { ... }
```

---

# TextStyleRun

```go
type TextStyleRun struct {
    FontIndex       int                // 索引到 TextSource.Fonts
    FontName        string             // 已 resolve 的字体名
    FontSize        float64            // em points
    FillColor       [4]float64         // [R, G, B, A] 每通道 0..1
    FauxBold        bool               // 合成粗体
    FauxItalic      bool               // 合成斜体
    AutoLeading     bool               // true → Leading 是 AE 自动值
    Leading         float64            // em points；!AutoLeading 时有意义
    Tracking        float64            // 1/1000 em
    BaselineShift   float64            // em points；正 = 上移，负 = 下移
    HorizontalScale float64            // raw 值（AE 默认 1，脚本 setter 范围 0..100；单位未完全 RE）
    VerticalScale   float64            // raw 值（同上）
    Tsume           float64            // CJK 字符压缩（0..100，AE 默认 0）
    ApplyStroke     bool               // 是否启用描边
    StrokeColor     [4]float64         // [R, G, B, A]
    StrokeWidth     float64            // em points
    CapsOption      TextCapsOption     // AE 24+ writeable；镜像 AE 的 allCaps / smallCaps 只读属性
    BaselineOption  TextBaselineOption // AE 24+ writeable；镜像 AE 的 subscript / superscript 只读属性
    StrokeOverFill  bool               // 描边在填充上 (默认 true)；AE 2020 ScriptingAPI 只读
    AutoKernType    TextAutoKernType   // AE 24+ writeable；0=NoAuto / 1=Metric / 2=Optical
    NoBreak         bool               // AE 24+ writeable；"不允许此 run 内换行"
    LineJoinType    TextLineJoinType   // AE 24+ writeable；描边角连接 0=Miter / 1=Round / 2=Bevel
    DigitSet        TextDigitSet       // AE 24+ writeable；0=Default / 1=Arabic / 2=Hindi / 3=Farsi / 4=ArabicRTL
}

// 便捷镜像（与 AE 的只读字段同义）
func (r TextStyleRun) AllCaps() bool      // CapsOption ∈ {All, AllSmall}
func (r TextStyleRun) SmallCaps() bool    // CapsOption ∈ {Small, AllSmall}
func (r TextStyleRun) Superscript() bool  // BaselineOption == Superscript
func (r TextStyleRun) Subscript() bool    // BaselineOption == Subscript
```

## Description

字符级样式 span。一段文本里如果所有字符样式一致，只有一个 run；如果某些字符有不同字号 / 颜色等，会拆成多个 run。**当前 run 与字符索引的映射未结构化**（runs 按 AE 文档顺序排列）。

## Attributes

### TextStyleRun.FontIndex

```go
FontIndex int
```

索引到 [`TextSource.Fonts`](#textsourcefonts)。`-1` 表示未识别。read-only。

---

### TextStyleRun.FontName

```go
FontName string
```

已 resolve 的字体名 = `Fonts[FontIndex]`（如果索引有效），否则为空。便捷字段。read-only。

---

### TextStyleRun.FontSize

```go
FontSize float64
```

字号（em points），= AE Character panel 的 "Font Size"。read / write via [`Layer.SetRunFontSize`](#layersetrunfontsize)。

---

### TextStyleRun.FillColor

```go
FillColor [4]float64
```

填充色 `[R, G, B, A]`，每通道 0..1。AE 内部存储顺序是 `[A, R, G, B]`，本库已转换。read / write via [`Layer.SetRunFillColor`](#layersetrunfillcolor)。

---

### TextStyleRun.FauxBold / FauxItalic

```go
FauxBold   bool
FauxItalic bool
```

合成粗体 / 斜体（字体本身缺粗 / 斜 cut 时由 AE 模拟）。read / write via [`Layer.SetRunFauxBold`](#layersetrunfauxbold) / [`SetRunFauxItalic`](#layersetrunfauxitalic)。

> AE 2020 的 ExtendScript 不允许通过 `td.fauxBold = true` 设置；本库直接改 btdk 字节，不受 AE 脚本 API 限制。

---

### TextStyleRun.AutoLeading / Leading

```go
AutoLeading bool
Leading     float64
```

#### Description

- `AutoLeading == true` —— `Leading` 是 AE 自动值（典型 `FontSize × 1.2`）
- `AutoLeading == false` —— `Leading` 是用户显式设置的 em points

```go
if !run.AutoLeading {
    fmt.Println("explicit leading:", run.Leading)
}
```

read-only。

---

### TextStyleRun.Tracking

```go
Tracking float64
```

字符间距，单位 1/1000 em（AE Character panel 显示的就是这个值）。`0` = 正常。read / write via [`Layer.SetRunTracking`](#layersetruntracking)。

---

### TextStyleRun.BaselineShift

```go
BaselineShift float64
```

基线偏移（em points）。正数 = 上移，负数 = 下移；AE 默认 `0`。read / write via [`Layer.SetRunBaselineShift`](#layersetrunbaselineshift)。

---

### TextStyleRun.HorizontalScale

```go
HorizontalScale float64
```

水平缩放倍率（raw 值，AE 默认 `1`）。AE 脚本 API setter 范围是 0..100，但底层存储单位与之关系未完全 RE — 直接以 PostScript dict 原值暴露。read / write via [`Layer.SetRunHorizontalScale`](#layersetrunhorizontalscale)。

---

### TextStyleRun.VerticalScale

```go
VerticalScale float64
```

垂直缩放倍率（raw 值，AE 默认 `1`）。同 HorizontalScale 的单位说明。read / write via [`Layer.SetRunVerticalScale`](#layersetrunverticalscale)。

---

### TextStyleRun.Tsume

```go
Tsume float64
```

CJK 字符压缩（"つめ" — 收紧 CJK 字符两侧的空白），0..100，AE 默认 `0`。仅在中文 / 日文 / 韩文字符上有视觉效果。read / write via [`Layer.SetRunTsume`](#layersetruntsume)。

---

### TextStyleRun.ApplyStroke

```go
ApplyStroke bool
```

是否启用描边。即使 `StrokeWidth > 0`，必须 `ApplyStroke == true` 才会渲染描边。read / write via [`Layer.SetRunApplyStroke`](#layersetrunapplystroke)。

---

### TextStyleRun.StrokeColor

```go
StrokeColor [4]float64
```

描边色 `[R, G, B, A]`。read / write via [`Layer.SetRunStrokeColor`](#layersetrunstrokecolor)。

---

### TextStyleRun.StrokeWidth

```go
StrokeWidth float64
```

描边宽度（em points）。read / write via [`Layer.SetRunStrokeWidth`](#layersetrunstrokewidth)。

---

### TextStyleRun.CapsOption

```go
CapsOption TextCapsOption  // 0=Normal, 1=Small, 2=All, 3=AllSmall
```

字符大小写选项。AE 24+ 引入的 `fontCapsOption` 写入路径 —— AE 2020 ScriptingAPI 把 `allCaps` / `smallCaps` 标 readonly，本库通过此字段提供写入能力（fixture 由 AE 24+ 生成）。read / write via [`Layer.SetRunCapsOption`](#layersetruncapsoption)。

便捷镜像：`r.AllCaps()` ↔ AE `TextDocument.allCaps`；`r.SmallCaps()` ↔ AE `TextDocument.smallCaps`。

---

### TextStyleRun.BaselineOption

```go
BaselineOption TextBaselineOption  // 0=Normal, 1=Superscript, 2=Subscript
```

上下标选项。AE 24+ 引入的 `fontBaselineOption` 写入路径。read / write via [`Layer.SetRunBaselineOption`](#layersetrunbaselineoption)。

便捷镜像：`r.Superscript()` / `r.Subscript()` ↔ AE 的同名只读属性。

---

### TextStyleRun.StrokeOverFill

```go
StrokeOverFill bool  // AE 默认 true
```

描边相对填充的渲染顺序。`true` = 描边盖在填充之上；`false` = 填充盖在描边之上。AE 2020 ScriptingAPI 标 readonly，本库通过直接改 btdk 字节绕过限制。read / write via [`Layer.SetRunStrokeOverFill`](#layersetrunstrokeoverfill)。

---

# TextParagraph

```go
type TextParagraph struct {
    Justification   TextJustification      // /0
    FirstLineIndent float64                // /1 — em points
    StartIndent     float64                // /2 — em points
    EndIndent       float64                // /3 — em points
    SpaceBefore     float64                // /4 — em points
    SpaceAfter      float64                // /5 — em points
    LeadingType     TextLeadingType        // /8 — AE 24+ (0=Roman, 1=Japanese)
    AutoHyphenate   bool                   // /9 — AE 默认 true
    HangingRoman    bool                   // /21 — AE 24+ (box-text-only)
    Direction       TextParagraphDirection // /33 — AE 24+ (0=LTR, 1=RTL)
}
```

## Description

一段的样式（按 `\n` 分隔）。`Justification` 自首发起就支持；其余字段是 AE 24+ ScriptingAPI 才能写入的（chunk slot 一直在），适用于 box-text 排版。read 全部可用 —— write 见下面的 [Per-paragraph setters](#per-paragraph-setters)。

---

# TextJustification

```go
type TextJustification int

const (
    TextJustifyLeft   TextJustification = 0
    TextJustifyRight  TextJustification = 1
    TextJustifyCenter TextJustification = 2
)

func (j TextJustification) String() string  // "Left" / "Right" / "Center"
```

---

# TextCapsOption

```go
type TextCapsOption int

const (
    TextCapsNormal   TextCapsOption = 0
    TextCapsSmall    TextCapsOption = 1
    TextCapsAll      TextCapsOption = 2
    TextCapsAllSmall TextCapsOption = 3
)

func (c TextCapsOption) String() string  // "Normal" / "SmallCaps" / "AllCaps" / "AllSmallCaps"
```

镜像 AE 24+ `FontCapsOption` enum。

---

# TextBaselineOption

```go
type TextBaselineOption int

const (
    TextBaselineNormal      TextBaselineOption = 0
    TextBaselineSuperscript TextBaselineOption = 1
    TextBaselineSubscript   TextBaselineOption = 2
)

func (b TextBaselineOption) String() string  // "Normal" / "Superscript" / "Subscript"
```

镜像 AE 24+ `FontBaselineOption` enum。

---

# TextAutoKernType / TextLineJoinType / TextDigitSet / TextLeadingType / TextParagraphDirection

AE 24+ 引入的 5 个 enum，全部 String() 实现。

```go
type TextAutoKernType int   // 0 NoAuto / 1 Metric / 2 Optical          (style-run /11)
type TextLineJoinType int   // 0 Miter / 1 Round / 2 Bevel              (style-run /62)
type TextDigitSet int       // 0 Default / 1 Arabic / 2 Hindi / 3 Farsi / 4 ArabicRTL (style-run /70)
type TextLeadingType int    // 0 Roman / 1 Japanese                     (paragraph /8)
type TextParagraphDirection int  // 0 LTR / 1 RTL                       (paragraph /33)
```

镜像 AE 24+ `AutoKernType` / `LineJoinType` / `DigitSet` / `LeadingType` / `ParagraphDirection`。Farsi / ArabicRTL 按文档顺序假定，未 fixture 验证。

---

## Functions

### aep.TextEncodedByteLen

```go
func TextEncodedByteLen(s string) int
```

#### Description

预测给定 Go 字符串经 AE 编码后的 PostScript 字符串字节数（含括号、`\xFE\xFF` BOM、UTF-16BE 码元、段末 `\r`、`( ) \` 转义）。用来在调 `Layer.SetText` 之前检查候选字符串是否符合 length-preserving 约束。

```go
candidate := "World"
if aep.TextEncodedByteLen(candidate) == aep.TextEncodedByteLen(layer.TextSource.Text) {
    layer.SetText(candidate) // safe
} else {
    log.Println("length mismatch — skip")
}
```

#### Returns

`int`，编码后的字节数。

---

## Layer.SetText

详见 [layer.md#layersettext](layer.md#layersettext)。length-preserving 文本字符串替换。

---

## Per-run setters

下面这一组 setter 走 **length-variable PostScript splice** 路径：每次调用都重新解析 btdk dict，定位目标节点的字节范围（`psValue.srcStart/srcEnd`），用新值字节替换，同时更新内嵌 `LIST btdk` 的 size header。`WriteAEP` 序列化时父 LIST size 自动重算。

调一次 setter ≈ 一次 btds extract + parse + splice + re-decode。re_text.aep（830 KB / 6.7 KB btds）上实测每个 setter < 1 ms。

`runIdx` 是 `TextSource.Runs` 的索引；`paraIdx` 是 `TextSource.Paragraphs` 的索引。索引越界返回 error。

### Layer.SetRunFontSize

```go
func (l *Layer) SetRunFontSize(runIdx int, sizePts float64) error
```

改字号（em points）。`88.0` 这样的整数会写成 `"88"`，小数写为 `g` 风格紧凑形式。

```go
layer.SetRunFontSize(0, 144)
```

### Layer.SetRunFontIndex

```go
func (l *Layer) SetRunFontIndex(runIdx, fontIdx int) error
```

让 style run 引用 `TextSource.Fonts[fontIdx]`。`fontIdx` 必须在 `Fonts` slice 范围内。

> 想换成 fonts 表里**没有**的字体：用 [`Layer.AddFont`](#layeraddfont) 先扩展字体表，再 `SetRunFontIndex` 指向新索引。

### Layer.AddFont

```go
func (l *Layer) AddFont(fontName string) (newIndex int, err error)
```

向 btdk 字体表（路径 `/0/1/0`）追加一个新字体条目，返回它在 `TextSource.Fonts` 里的索引。length-variable splice。

新条目按 AE 写非默认字体的格式序列化：

```text
<< /0 << /99 /CoolTypeFont /0 << /0 (FE FF utf16be fontName) /2 0 >> >> >>
```

```go
idx, err := layer.AddFont("Arial-BoldMT")
if err == nil {
    layer.SetRunFontIndex(0, idx) // 把 run 0 切到新字体
}
```

> 字体名要用 PostScript 名（如 `"Arial-BoldMT"` 而不是 `"Arial Bold"`）。AE 渲染时若该字体未安装，会用 fallback 显示。本库不校验字体存在 —— 字符串原样写入 btdk。

### Layer.SetRunFillColor

```go
func (l *Layer) SetRunFillColor(runIdx int, rgba [4]float64) error
```

改填充色 `[R, G, B, A]`（每通道 0..1）。内部转回 AE 存储顺序 `[A, R, G, B]`。

```go
layer.SetRunFillColor(0, [4]float64{1, 0, 0, 1}) // 纯红
```

### Layer.SetRunStrokeColor

```go
func (l *Layer) SetRunStrokeColor(runIdx int, rgba [4]float64) error
```

改描边色 `[R, G, B, A]`。同 SetRunFillColor 的颜色顺序处理。

### Layer.SetRunApplyStroke

```go
func (l *Layer) SetRunApplyStroke(runIdx int, apply bool) error
```

切换是否渲染描边。

### Layer.SetRunStrokeWidth

```go
func (l *Layer) SetRunStrokeWidth(runIdx int, width float64) error
```

改描边宽度（em points）。

### Layer.SetRunTracking

```go
func (l *Layer) SetRunTracking(runIdx int, tracking float64) error
```

改字符间距（1/1000 em）。

### Layer.SetRunLeading

```go
func (l *Layer) SetRunLeading(runIdx int, leading float64) error
```

改行距（em points）。**注意**：AE 默认 auto-leading 打开，会用 `FontSize × 1.2` 覆盖你的 Leading 值。要让 Leading 生效，先调 `SetRunAutoLeading(runIdx, false)`。

### Layer.SetRunAutoLeading

```go
func (l *Layer) SetRunAutoLeading(runIdx int, auto bool) error
```

切换 auto-leading 标记（`/4` 位）。

### Layer.SetRunBaselineShift

```go
func (l *Layer) SetRunBaselineShift(runIdx int, shift float64) error
```

改基线偏移（em points；正 = 上移，负 = 下移）。

### Layer.SetRunHorizontalScale

```go
func (l *Layer) SetRunHorizontalScale(runIdx int, scale float64) error
```

改水平缩放（raw 值，AE 默认 1）。

### Layer.SetRunVerticalScale

```go
func (l *Layer) SetRunVerticalScale(runIdx int, scale float64) error
```

改垂直缩放（同上）。

### Layer.SetRunTsume

```go
func (l *Layer) SetRunTsume(runIdx int, tsume float64) error
```

改 CJK 字符压缩（0..100）。

### Layer.SetRunFauxBold

```go
func (l *Layer) SetRunFauxBold(runIdx int, on bool) error
```

切换合成粗体。

### Layer.SetRunFauxItalic

```go
func (l *Layer) SetRunFauxItalic(runIdx int, on bool) error
```

切换合成斜体。

### Layer.SetRunCapsOption

```go
func (l *Layer) SetRunCapsOption(runIdx int, caps TextCapsOption) error
```

写字符大小写选项（AE 24+ 引入的 `fontCapsOption` 字段）。`caps` 取 `TextCapsNormal` / `TextCapsSmall` / `TextCapsAll` / `TextCapsAllSmall`。底层 btdk `/12` 字段单字节数字 —— 0..3 内切换 length-preserving 一致。

```go
layer.SetRunCapsOption(0, aep.TextCapsAll)
```

### Layer.SetRunBaselineOption

```go
func (l *Layer) SetRunBaselineOption(runIdx int, base TextBaselineOption) error
```

写上下标选项（AE 24+ 引入的 `fontBaselineOption`）。`base` 取 `TextBaselineNormal` / `TextBaselineSuperscript` / `TextBaselineSubscript`。

```go
layer.SetRunBaselineOption(0, aep.TextBaselineSubscript)
```

### Layer.SetRunStrokeOverFill

```go
func (l *Layer) SetRunStrokeOverFill(runIdx int, over bool) error
```

切换描边相对填充的渲染顺序。AE 默认 `true`。

### Layer.SetRunAutoKernType

```go
func (l *Layer) SetRunAutoKernType(runIdx int, kt TextAutoKernType) error
```

写自动 kerning 模式（AE 24+ 引入的 `autoKernType`，btdk style-run `/11`）。`NoAuto` 后的手动 kerning 值存在另一棵 sub-tree（`/1/1[0]/0/8`），本库暂未结构化暴露写入。

### Layer.SetRunNoBreak

```go
func (l *Layer) SetRunNoBreak(runIdx int, on bool) error
```

切换 "no break" 字符标志（AE 24+，btdk style-run `/52`）。`true` 时 AE 不在此 run 内拆行。

### Layer.SetRunLineJoinType

```go
func (l *Layer) SetRunLineJoinType(runIdx int, j TextLineJoinType) error
```

写描边角连接模式（AE 24+ `lineJoinType`，btdk style-run `/62`）。

### Layer.SetRunDigitSet

```go
func (l *Layer) SetRunDigitSet(runIdx int, d TextDigitSet) error
```

写数字字符集（AE 24+ `digitSet`，btdk style-run `/70`）。

---

## Per-paragraph setters

### Layer.SetParagraphJustification

```go
func (l *Layer) SetParagraphJustification(paraIdx int, j TextJustification) error
```

改第 paraIdx 段的对齐方式（`TextJustifyLeft` / `Right` / `Center`）。

```go
layer.SetParagraphJustification(0, aep.TextJustifyCenter)
```

### Layer.SetParagraphFirstLineIndent / StartIndent / EndIndent

```go
func (l *Layer) SetParagraphFirstLineIndent(paraIdx int, v float64) error
func (l *Layer) SetParagraphStartIndent(paraIdx int, v float64) error
func (l *Layer) SetParagraphEndIndent(paraIdx int, v float64) error
```

段缩进（em points）。AE 24+ ScriptingAPI 才能在 box-text 上写入；本库直接改 btdk 字节，point text 也可写但效果取决于 AE 重绘逻辑。

```go
layer.SetParagraphStartIndent(0, 24)   // 段首左缩进
layer.SetParagraphFirstLineIndent(0, 36) // 首行额外缩进
```

### Layer.SetParagraphSpaceBefore / SpaceAfter

```go
func (l *Layer) SetParagraphSpaceBefore(paraIdx int, v float64) error
func (l *Layer) SetParagraphSpaceAfter(paraIdx int, v float64) error
```

段前 / 段后间距（em points）。

### Layer.SetParagraphAutoHyphenate

```go
func (l *Layer) SetParagraphAutoHyphenate(paraIdx int, on bool) error
```

切换自动连字符。AE 默认 `true`。

### Layer.SetParagraphLeadingType

```go
func (l *Layer) SetParagraphLeadingType(paraIdx int, lt TextLeadingType) error
```

写 leading 类型（AE 24+ `leadingType`，btdk paragraph `/8`）。`TextLeadingRoman` / `TextLeadingJapanese`。

### Layer.SetParagraphHangingRoman

```go
func (l *Layer) SetParagraphHangingRoman(paraIdx int, on bool) error
```

切换 Roman 悬挂标点（AE 24+ `hangingRoman`，btdk paragraph `/21`）。仅对 box-text 有意义。

### Layer.SetParagraphDirection

```go
func (l *Layer) SetParagraphDirection(paraIdx int, d TextParagraphDirection) error
```

写段落阅读方向（AE 24+ `direction`，btdk paragraph `/33`）。`TextDirectionLeftToRight` / `TextDirectionRightToLeft`。

---

## 批量改字体颜色 + 字号示例

```go
proj, _ := aep.Open("template.aep")

for _, comp := range proj.Compositions {
    for _, layer := range comp.Layers {
        if layer.TextSource == nil {
            continue
        }
        for i := range layer.TextSource.Runs {
            layer.SetRunFontSize(i, 72)
            layer.SetRunFillColor(i, [4]float64{1, 0.4, 0, 1}) // 橘色
            layer.SetRunAutoLeading(i, false)
            layer.SetRunLeading(i, 90)
        }
        for i := range layer.TextSource.Paragraphs {
            layer.SetParagraphJustification(i, aep.TextJustifyCenter)
        }
    }
}

f, _ := os.Create("out.aep")
defer f.Close()
proj.WriteAEP(f)
```

---

## 当前未结构化暴露的字段

绝大多数 AE 24+ TextDocument 字段已 R/W —— 详见 [../flightdeck/flight-plans/coverage-detail.md](../flightdeck/flight-plans/coverage-detail.md) 文本字段表。仍未实现：

- `ligature`（位置未 RE；AE 默认 false 且 set false 无 diff，需 ligature-enabled 字体 fixture）
- character-level selection runs（字符级选区，运行时 UI 状态，不太可能持久化）
- `lineOrientation` 横/竖排切换 —— 结构性（layer-local 坐标重排），超 length-preserving 范围
- `composerEngine` / `everyLineComposer` —— AE 24+ 只接受 `UNIVERSAL_TYPE_ENGINE`，其它写入抛错；行为属性 read-only
- 删除 fonts 表条目（`AddFont` 只追加）
- 手动 kerning **首次启用**（`/8` slot 不存在时 `SetManualKerning` refused — 需先在 AE 里设非零 kerning 让 AE emit slot）

需要扩展时按 [parse_text.go](../internal/aep/parse_text.go) 里的 `psPath` 模式 + [write_text.go](../internal/aep/write_text.go) 的 `splicePSValue` 模式实现。RE 路径见 `test_data/re_text_caps_ae24.jsx` / `re_text_ae24_more.jsx` 模板。
