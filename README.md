# aep-parser

一个从头实现的 Adobe After Effects `.aep` 项目文件解析器，纯 Go 编写，无需 AE 运行实例。
支持**读取**项目结构（合成 / 素材 / 图层 / 属性 / 关键帧 / 蒙版 / 文本 / 标记）和**字节长度保持的写回**（修改素材路径、关键帧值/时间、静态属性值、文本字符串）。

**兼容下限：After Effects 2020（CC 17.0）**。新版本 AE 写的 .aep 也能读，但 AE 24+ 才引入的字段（如 `fontCapsOption` / `alternateSource`）本库不主动解码，对调用方返回 nil 而非错误。详见 [workshop/plans/coverage-detail.md](workshop/plans/coverage-detail.md) 的兼容性章节。

> **每个分类支持哪些字段、哪些 API 是只读还是可写、示例片段**：见 [docs/](docs/README.md) —— 每个核心类型一个 markdown，结构参考 [docsforadobe/after-effects-scripting-guide](https://github.com/docsforadobe/after-effects-scripting-guide)。

## 原理

`.aep` 文件使用 **RIFX** 格式（Big-Endian RIFF），魔数为 `RIFX` + `Egg!`。
内部是嵌套的 Chunk 树，本库通过逆向工程已知偏移量提取各类数据。

## 支持内容

### 读取

| 数据类型 | 字段 |
| --- | --- |
| **Project** | 合成 / 素材 / 文件夹列表、色深 (8/16/32 bpc)、**Warnings []string** (解析时遇到的非致命异常 — chunk 长度不符等；空切片 = 干净解析) |
| **Composition** | 名称、ID、分辨率、帧率、时长、**TickRate (per-comp keyframe ticks/s)**、**BGColor** (cdta @0x34-0x36, RGB 0..255；之前被声明但从未填充，已修复)、**WorkAreaStart/End** (s；end_dividend=0xFFFFFFFF 自动替换为 Duration)、**ShutterAngle** (°, u16)、**ShutterPhase** (i32 raw)、**MotionBlur 采样参数** (Adaptive Limit + Samples/Frame, AE 默认 128/16)、图层列表、**Markers []*Marker** (合成级 timeline 章节 marker；来自 `SecL "Markers"` 伪图层，与 layer marker 同一解码器) |
| **Footage** | 名称、ID、文件路径、分辨率、帧率、时长、是否 Still / Solid |
| **Folder** | 名称、ID |
| **Layer** | Name/Type (av/text/shape/null/light/camera/adjustment)、**ID / ParentID** (own + parent layer's ID)、**Parent()** (解析后的 `*Layer` 父引用，跨同 comp)、SourceID、**SourceComposition()** (pre-comp 源指向的 `*Composition`；footage/solid/缺失项返回 nil)、StartTime/Duration/Stretch (fractional dividend/divisor decoding)、**Quality** (Wireframe/Draft/Best)、**Label** (timeline color chip 0..16)、**BlendingMode** (40-value enum including Normal/Multiply/Screen/Overlay/Add/Subtract/Difference/...)、**TrackMatte** (None/Alpha/AlphaInv/Luma/LumaInv)、**PreserveTransparency**、**AutoOrient** (None/AlongPath/CameraOrPOI/CharactersTowardCamera)、**Comment** (cmta chunk, CRLF→LF normalized)、flag bits (3D/Solo/Shy/Locked/Visible/IsAdjust/**IsNull/IsGuide/MarkersLocked/MotionBlur/EffectsEnabled/AudioEnabled/FrameBlendEnabled/CollapseTransform/SamplingBicubic/FrameBlendPixelMotion**)、IsShapeLayer、属性 / 特效 / 标记 / 蒙版 / 文本源 |
| **Property** | MatchName、Name、维度 (1D / **2D** / 3D / **4D 颜色 RGBA**)、关键帧列表 **或** 静态值、表达式源码 |
| **Keyframe** | 时间（秒）、值（float64 或 [3]float64）、In/Out 插值类型 (linear/bezier/hold)、空间切线 (3D vec, spatial 属性才有)、**InTemporalEase/OutTemporalEase** ([]TemporalEase，含 Speed + Influence，按维度展开：spatial 和 1D 长度 1，非空间 N-D 长度 N) |
| **TemporalEase** | Speed (单位/秒，0=Easy Ease 在该点停顿) + Influence (0..1, 默认 1/3) |
| **Effect** | MatchName (如 "ADBE Gaussian Blur 2")、Name、参数列表（每个参数都是 Property） |
| **Marker** | 时间、**Duration** (s；NmHd@0x08/600，点 marker = 0)、**Label** (u8 时间线色卡 0..16)、Comment、Chapter、URL、FrameTarget、CuePointName |
| **Mask** | Name、Index、**Mode** (Add/Subtract/Intersect/Lighten/Darken/Difference/None)、**Inverted**、**Color** ([3]uint8 RGB)、Closed、Vertices ([Anchor, InTangent, OutTangent] 各 [2]float64，绝对坐标 Bezier 控制点)、**Feather** ([2]float64 X,Y)、**Opacity** (0..1)、**Expansion** (px)、**Properties** (mask 内层属性)、**PathKeyframes** (动画蒙版每个时间点的完整路径快照 + scalar TemporalEase)、MkifRaw、ShphRaw |
| **ShapePath** | 形状图层 Pen 工具自由 Bezier 路径（Name、Closed、Vertices、ShphRaw）—— 参数化的 Star/Rect/Ellipse 不会产生 ShapePath，它们以普通 Property 形式暴露 |
| **TextSource** | 文本图层解码视图：`Text` (UTF-8，多段已合并为 `\n`)、`Fonts []string` (PostScript 字体名表)、`Paragraphs []TextParagraph` (per-paragraph `Justification`)、`Runs []TextStyleRun`、`Justification` (首段便捷字段)、`IsBoxText`+`BoxBounds [xmin,ymin,xmax,ymax]` (point text vs box text 区分 + 包围盒) |
| **TextStyleRun** | per-run 字符样式：`FontIndex`/`FontName`/`FontSize`/`FillColor [R,G,B,A]`、`FauxBold`/`FauxItalic`、`AutoLeading`+`Leading`、`Tracking` (1/1000 em)、`ApplyStroke`+`StrokeColor [R,G,B,A]`+`StrokeWidth` |
| **TextSourceRaw** | 文本图层 btds 原始字节，与 `TextSource` 并存。round-trip 写回 + 解码器未覆盖字段（tracking/leading/stroke/box-text 等）需要直接读这里 |

### 写回（length-preserving）

修改后调用 `Project.WriteAEP(w)` 序列化回合法 .aep。所有写入都保持原 chunk 字节长度，AE 可重新打开。

| API | 作用 |
| --- | --- |
| `Footage.SetPath(newPath)` | 改素材路径（重写 alas JSON 的 `fullpath` + 兼容旧版 Cpth chunk） |
| `Property.SetStaticValue(v)` | 改没有关键帧的属性值（1D `float64` / 3D `[]float64`） |
| `Property.SetExpression(s)` | 改 / 清表达式 JS 源码（length-variable Utf8 chunk 替换 / 创建 / 删除） |
| `Keyframe.SetTime(seconds)` | 改关键帧时间 |
| `Keyframe.SetValue(v)` | 改关键帧值（1D / 3D，自动适配 spatial 和 non-spatial 两种字节布局） |
| `Keyframe.SetInInterp / SetOutInterp(InterpType)` | 改单端插值方式（Linear / Bezier / Hold） |
| `Keyframe.SetInTemporalEase / SetOutTemporalEase([]TemporalEase)` | 改时间 ease（length 自动匹配 spatial=1 / non-spatial=N） |
| `Keyframe.SetInSpatialTangent / SetOutSpatialTangent([]float64)` | 改 3D Bezier 切线（仅 spatial 属性） |
| `Property.InsertKeyframe(time, value)` / `DeleteKeyframe(i)` | **增删 keyframe**（length-variable ldat 重排 + 同步 lhd3 count；前置：属性需有 ≥1 现有 keyframe 来克隆 layout） |
| `Layer.SetStartTime / SetInPoint / SetOutPoint / SetStretch` | 改图层时间字段（dividend/divisor pair；SetIn/OutPoint 自动重算 Duration） |
| `Layer.SetAutoOrient(AutoOrientType)` | 改自动朝向（3 个互斥位翻转） |
| `Layer.SetParent(parentID)` / `SetSource(sourceID)` | 改父图层 / 源项目项（ldta 4 字节写 + ID 校验） |
| `Layer.AddFont(fontName string)` | 向文本图层 fonts 表追加新字体，返回索引；与 `SetRunFontIndex` 配合切换 run 字体 |
| `Composition.SetName / SetFrameRate / SetDuration` | 改合成元数据（Utf8 + cdta 字节；SetFrameRate 自动重算 Duration） |
| `Composition.SetDraft3D / SetHideShyLayers / SetCompMotionBlur / SetFrameBlending / SetPreserveNestedFrameRate / SetPreserveNestedResolution` | 合成 cdta 标志位翻转（`@0x8A`/`@0x8B`） |
| `Composition.SetPixelAspect(par float64)` | 改像素宽高比（cdta num/den 整数对） |
| `Composition.SetComment(s) / SetLabel(idx)` / `Footage.SetComment / SetLabel` | Item-level 备注 + 色卡（cmta + idta `@0x3A`） |
| `Mask.SetLocked(bool) / SetMaskMotionBlur(MaskMotionBlurMode)` | mask 锁定 + 运动模糊 override |
| `Property.SetExpressionEnabled(bool)` | 切换表达式启用状态（与 `SetExpression` 写源码独立） |
| `Project.SetBitsPerChannel(BPC8/16/32)` | 改工程色深（双 write nhed + nnhd 头部冗余存储） |
| `Layer.SetText(s string)` | 文本图层用户文字 length-preserving 替换（同字节数才接受）；`TextEncodedByteLen(s)` 预测候选字符串编码后字节数，方便调用者先检查 |
| `Layer.SetRunFontSize / SetRunFillColor / SetRunStrokeColor / SetRunApplyStroke / SetRunStrokeWidth / SetRunTracking / SetRunLeading / SetRunAutoLeading / SetRunBaselineShift / SetRunHorizontalScale / SetRunVerticalScale / SetRunTsume / SetRunFauxBold / SetRunFauxItalic / SetRunFontIndex` | 文本 per-run 字段写回（length-variable PostScript splice，自动更新内嵌 LIST btdk size header） |
| `Layer.SetParagraphJustification(i, j)` | 文本 per-paragraph 对齐方式写回 |
| `Layer.SetName(s)` / `SetComment(s)` | length-variable Utf8 / cmta chunk 替换（cmta 缺失时自动插入） |
| `Layer.SetVisible/SetSolo/SetShy/SetLocked/SetEffectsEnabled/...` | 16 个 flag bit setter（ldta 单 bit 翻转） |
| `Layer.SetBlendingMode/SetTrackMatte/SetLabel/SetQuality/SetPreserveTransparency` | 5 个 enum / 字节 setter |
| `Composition.SetBGColor/SetShutterAngle/SetShutterPhase/SetMotionBlur*/SetWorkArea` | 6 个合成元数据 setter（cdta 字节写） |
| `Marker.SetTime/SetDuration/SetLabel` | length-preserving 时间 / 时长 / 色卡 |
| `Marker.SetComment/SetChapter/SetURL/SetFrameTarget/SetCuePointName` | length-variable Utf8 chunk 文本 setter |

## 文件结构

```text
aep-parser/
├── internal/
│   ├── aep/
│   │   ├── parse.go        # 项目/合成/图层/属性/关键帧解析
│   │   ├── write.go        # WriteAEP + Set* 写回 API
│   │   ├── types.go        # 公开数据模型
│   │   ├── json.go         # JSON 序列化
│   │   └── aep_test.go     # 单元测试（含合成 RIFX 构造器）
│   └── rifx/
│       └── rifx.go         # 底层 RIFX/Chunk 读写器
└── main/
    └── main.go             # 示例 runner（读取 + round-trip 写回 demo）
```

## 快速开始

### 读取

```go
import aep "github.com/example/aep-parser/internal/aep"

project, err := aep.Open("my-project.aep")
if err != nil {
    panic(err)
}

for _, comp := range project.Compositions {
    fmt.Printf("合成: %s  %dx%d  %.2ffps  %.2fs\n",
        comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration)
    for _, layer := range comp.Layers {
        // 一级 Transform 访问器
        if pos := layer.Position(); pos != nil && len(pos.Keyframes) > 0 {
            fmt.Printf("  图层[%d] %s: Position 有 %d 个关键帧\n",
                layer.Index+1, layer.Name, len(pos.Keyframes))
        }
        // 特效、标记、文本、表达式
        for _, fx := range layer.Effects {
            fmt.Printf("    Effect: %s (%d 参数)\n", fx.MatchName, len(fx.Parameters))
        }
        for _, m := range layer.Markers {
            fmt.Printf("    Marker @ %.2fs: %s\n", m.Time, m.Comment)
        }
        if layer.TextSource != nil {
            ts := layer.TextSource
            font := ""
            size := 0.0
            if len(ts.Runs) > 0 {
                font = ts.Runs[0].FontName
                size = ts.Runs[0].FontSize
            }
            fmt.Printf("    文本: %q  font=%q size=%.1f  %s\n",
                ts.Text, font, size, ts.Justification)
        }
        for _, p := range layer.Properties {
            if p.Expression != "" {
                fmt.Printf("    %s 表达式: %s\n", p.MatchName, p.Expression)
            }
        }
    }
}

project.WriteJSON(os.Stdout)

// 解析过程中遇到的异常（chunk 长度不一致等）以 warning 形式收集，不影响其他字段。
for _, w := range project.Warnings {
    fmt.Fprintln(os.Stderr, "warn:", w)
}
```

### 写回

```go
project, _ := aep.Open("in.aep")

// 1. 改素材路径
project.Footage[0].SetPath(`D:\new\location\file.png`)

// 2. 改关键帧值
if pos := project.Compositions[0].Layers[0].Position(); pos != nil {
    pos.Keyframes[0].SetValue([]float64{960, 540, 0})  // 改值
    pos.Keyframes[1].SetTime(5.0)                       // 改时间
}

// 3. 改静态属性
if opa := project.Compositions[0].Layers[0].Opacity(); opa != nil && opa.StaticValue != nil {
    opa.SetStaticValue(0.5)
}

f, _ := os.Create("out.aep")
defer f.Close()
project.WriteAEP(f)
```

## 已知限制

- **JSON 是单向导出**：`Project.WriteJSON` / `Project.MarshalJSON` 只把解析后的 Project 序列化为 JSON 视图，**没有 `ReadJSON` 反序列化**。来回 round-trip 必须走二进制路径 (`aep.Open` → 修改 → `Project.WriteAEP`)，JSON 路径不保留底层 chunk 字节。
- **不支持结构性编辑**：增删图层 / 关键帧 / 属性 / 特效 / 标记 / 蒙版顶点 都需要重排 chunk 长度，当前只做 length-preserving 改写。
- **文本图层**：`Layer.TextSource` 解出 Text / Fonts / per-paragraph Justification / per-run Font*/FillColor/FauxBold/FauxItalic/AutoLeading/Leading/Tracking/Stroke / IsBoxText+BoxBounds。`Layer.TextSourceRaw` 同时保留 btds 原字节。**尚未结构化暴露**：baseline shift、Horizontal/Vertical Scale、subscript/superscript、character/paragraph indent、字符级 selection runs —— 这些值仍在 btdk PostScript dict 里，需要时按相同模式从 `TextSourceRaw` 解。
- **Mkif 已结构化**：`Mask.Mode` / `Mask.Inverted` / `Mask.Index` / `Mask.Color` 已解出；`Mask.MkifRaw` 原始 48 字节仍保留供 round-trip。剩余 mkif 字节（0x10/0x18/0x20-0x27 几个常量字段）未解但也不像是用户可配置的。
- **形状层 vector primitives**：`IsShapeLayer` 已标记，Pen 工具自由 Bezier 路径以 `Layer.ShapePaths` 暴露。**参数化的** Rect/Star/Ellipse（Size、Rotation、Inner Radius 等）仍以普通 `Properties` 形式出现，没有结构化的 `Layer.Shapes []*Shape` 分组 API。
- **颜色值范围**：4D 颜色属性 (e.g. Tritone Highlights) 的 RGBA 是 0..255 doubles（8bpc 项目）。32bpc 项目里可能是 0..1，未验证。
- **效果参数关键帧**：测试过 1D 静态/缓动（Gaussian Blur Blurriness）和 4D 颜色缓动（Tritone Highlights）。其他效果参数类型（菜单 enum 等）走通用 Property 路径，按字节读取应该可以。
- **per-comp 关键帧 tick rate** ✓ 已 RE：cdta 0x08-0x0B = ticks/s，cdta 0xA8-0xAB = legacy scale。modern comp 直接用 0x08；legacy NTSC comp 用 `0x08 × 1000 / 0xA8` (= 8000 for 29.97). `Composition.TickRate` 暴露解码后的值。
- **帧率**：cdta 偏移 0x9C 处的值 ÷ 100；部分旧版本 AE 格式可能不同。
- **测试文件**：Adobe 未公开 AEP 格式，建议用真实 .aep 文件验证后按需调整偏移量。

## 与上游 boltframe/aftereffects-aep-parser 的差异

本库以 boltframe 的 Go 实现为骨架重写。相对上游主要变更：

| 改进 | 说明 |
| --- | --- |
| **per-comp TickRate** | 从 cdta @0x08 / @0xA8 解码每个合成自己的 keyframe-tick-per-second，修复上游用统一 8000 导致非-29.97-fps 合成关键帧时间偏差的问题 |
| **BGColor / WorkArea 补完** | cdta @0x34-0x36 解 RGB；WorkArea dividend = 0xFFFFFFFF 自动替换为 Duration |
| **MotionBlur / Shutter 参数** | ShutterAngle、ShutterPhase、AdaptiveSampleLimit、SamplesPerFrame 全部解出 |
| **4D 颜色 + Non-spatial 关键帧** | 4D RGBA 关键帧（如 Tritone Highlights）、Mask Feather 2D 等 non-spatial N-D layout 完整支持 |
| **TemporalEase per-component** | spatial / non-spatial 两种字节布局分别处理，每条 ease 包含 Speed + Influence |
| **Mask 完整解码** | Mode / Inverted / Color / Feather / Opacity / Expansion / 顶点 + 动画路径快照 + scalar TemporalEase |
| **Shape Layer Pen 路径** | `Layer.ShapePaths` 暴露 Bezier 自由路径（参数化 Rect/Star/Ellipse 仍走 Property） |
| **文本图层结构化** | `Layer.TextSource` 把 btdk PostScript dict 解出来：Text / Fonts 表 / per-run Font*+FillColor+Stroke+Leading+Tracking / per-paragraph Justification / IsBoxText+BoxBounds（上游只有 raw 字节） |
| **`Layer.SetText` 写回** | length-preserving 文本字符串替换；`TextEncodedByteLen` 帮调用者预测候选长度，AE 仍可正常打开 |
| **Composition 级 Markers** | `Composition.Markers` 来自 SecL "Markers" 伪图层（上游未识别该 LIST） |
| **Marker 全字段** | NmHd Duration（点 marker = 0）+ Label + 五个 Utf8 字符串 |
| **Length-preserving 写回 API** | `Footage.SetPath` / `Property.SetStaticValue` / `Keyframe.SetTime` / `Keyframe.SetValue` —— 上游只读 |
| **Project.Warnings** | 解析时检测到 chunk 长度异常等非致命问题，会收集为 warning 而不是静默丢弃数据 |

## 参考

- RIFX 规范：[RIFF/RIFX on Wikipedia](https://en.wikipedia.org/wiki/Resource_Interchange_File_Format)
- Kaitai Struct IDE：用于可视化 AEP 二进制结构
- 原有 Go 实现参考：[boltframe/aftereffects-aep-parser](https://github.com/boltframe/aftereffects-aep-parser)
- Python 实现参考：[forticheprod/py-aep](https://github.com/forticheprod/py-aep)
