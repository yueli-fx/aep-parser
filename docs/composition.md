# Composition object

`project.Compositions[index]` / `project.CompositionByID(id)`

## Description

合成（AE UI 里的 Comp）。持有分辨率 / 帧率 / 时长 / 工作区 / 运动模糊采样设置 / 背景色 / 图层列表 / 合成级 markers。

> **TickRate 是 per-composition 的** —— 不要假设统一 8000。`Composition.TickRate` 由 cdta `@0x08` / `@0xA8` 推导，关键帧 / marker 时间换算都用它。modern comp = ticks/sec；legacy 29.97 NTSC comp = 8000。

## Example

```go
comp := proj.Compositions[0]
fmt.Printf("%s  %dx%d  %.2ffps  %.2fs  TickRate=%g\n",
    comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration, comp.TickRate)
fmt.Printf("BGColor: #%02X%02X%02X\n", comp.BGColor[0], comp.BGColor[1], comp.BGColor[2])
fmt.Printf("Work area: %.2fs - %.2fs\n", comp.WorkAreaStart, comp.WorkAreaEnd)
for _, m := range comp.Markers {
    fmt.Printf("  comp marker @ %.2fs: %s\n", m.Time, m.Comment)
}
```

---

## Creation

### Project.NewComposition

```go
func (p *Project) NewComposition(name string, width, height uint16, frameRate, duration float64) (*Composition, error)
```

#### Description

在 project 根目录新建一个空 composition（0 layers）。

Required 参数（任一不合法返 error，project 不被修改）：
- `name` — 非空 string
- `width` / `height` — uint16, > 0
- `frameRate` — > 0 (Hz)；NTSC 容差内的 23.976 / 29.97 / 59.94 映射到 AE canonical 编码
- `duration` — > 0 (秒)；内部换算为整帧数

可选属性默认值（用 `Set*` 方法在 New 后修改）：BGColor=[0,0,0] / PixelAspect=1.0 / ResolutionFactor=[1,1] / ShutterAngle=180 / ShutterPhase=0 / MotionBlurAdaptive=128 / MotionBlurSamples=16。

ID 自动分配（monotonic，never reuse）。新 comp append 到 project root folder。与 `Open(...)` 出来的 comp 完全同构 —— 所有 `Set*` 方法立即可用。

```go
proj := aep.NewProject()
main, err := proj.NewComposition("Main", 1920, 1080, 29.97, 10)
if err != nil { log.Fatal(err) }
main.SetBGColor([3]uint8{20, 30, 40})
main.SetResolutionFactor(2, 2)
```

输出文件用 AE 跨版本通用结构 —— builder 永远写最小 canonical seed (AE 2020 兼容)，AE 25 / 22 / 20 都能开。详见 [project.md NewProject 段](project.md#aepnewproject)。

#### Atomicity

`NewComposition` 是原子操作：如果内部 chunk parse 失败或产生 parser warning，回滚 project 到 pre-call 状态并返 error。调用者不会看到半成品 comp。

---

## Attributes

### Composition.ID

```go
ID uint32
```

AE 内部 item ID。read-only。

---

### Composition.Name

```go
Name string
```

合成显示名。read / write via [`SetName`](#compositionsetname)。

---

### Composition.Width

```go
Width uint16
```

像素宽度。read-only。

---

### Composition.Height

```go
Height uint16
```

像素高度。read-only。

---

### Composition.FrameRate

```go
FrameRate float64
```

帧率（fps）。来自 cdta `@0x9C-0x9F`（whole + 1/65536 fractional 组合）。read / write via [`SetFrameRate`](#compositionsetframerate)。

---

### Composition.Duration

```go
Duration float64
```

合成时长（秒）= 帧总数 ÷ FrameRate。read / write via [`SetDuration`](#compositionsetduration)。

---

### Composition.TickRate

```go
TickRate float64
```

#### Description

per-composition 的关键帧 / marker 时间 base —— ticks per second。本库内部用它把 cdta / ldat 里的整数 tick 计数换算成秒。

解码逻辑（来自 cdta）：

```text
rate  := cdta_uint32_BE @ 0x08
scale := cdta_uint32_BE @ 0xA8
if scale <= 1:                          // modern comp
    TickRate = rate                     // 如 30720 for 30fps
else:                                   // legacy NTSC
    TickRate = rate * 1000 / scale      // 8000 for 29.97
```

cdta 缺失时回退到常量 `8000`（AE 的 legacy time base）。

#### Type

`float64`；read-only。

---

### Composition.BGColor

```go
BGColor [3]uint8
```

#### Description

合成背景色 RGB，每通道 0..255。来自 cdta `@0x34/@0x35/@0x36`。默认 `{0, 0, 0}`（AE 默认黑）。

#### Type

`[3]uint8`；read / write via [`SetBGColor`](#compositionsetbgcolor)。

---

### Composition.WorkAreaStart

```go
WorkAreaStart float64
```

#### Description

工作区起点（秒）。来自 cdta dividend/divisor pair `@0x1C` / `@0x20`。

#### Type

`float64`；read / write via [`SetWorkArea`](#compositionsetworkarea)（写法同时改 start + end）。

---

### Composition.WorkAreaEnd

```go
WorkAreaEnd float64
```

#### Description

工作区终点（秒）。来自 cdta dividend/divisor pair `@0x24` / `@0x28`。

> **Sentinel 处理**：当 AE 把工作区"留空"，会写 `0xFFFFFFFF` 为 end dividend。本库自动把这种情况替换为 `Duration`，所以调用方拿到的总是有效的秒数。

#### Type

`float64`；read / write via [`SetWorkArea`](#compositionsetworkarea)。

---

### Composition.DisplayStartTime

```go
DisplayStartTime float64
```

#### Description

合成的"显示起始时间"（秒）—— 对应 AE 脚本 `CompItem.displayStartTime` / `displayStartFrame`（= `DisplayStartTime × FrameRate`）。控制时间线显示的时间偏移，**不**改变实际渲染范围。来自 cdta dividend/divisor pair `@0xA4` / `@0xA8`（uint32 BE 对）。

RE 来源：`re_wave2_ae24.aep` (`RE_CDTA_DSF_120` 用 displayStartFrame=120 @ 29.97 fps，存储为 4.004 秒)。

#### Type

`float64`；read / write via [`SetDisplayStartTime`](#compositionsetdisplaystarttime) / [`SetDisplayStartFrame`](#compositionsetdisplaystartframe)。

---

### Composition.ShutterAngle

```go
ShutterAngle uint16
```

#### Description

运动模糊快门角度（度）。AE UI 范围 0..720，默认 180。来自 cdta `@0xAE`。

#### Type

`uint16`；read / write via [`SetShutterAngle`](#compositionsetshutterangle)。

---

### Composition.ShutterPhase

```go
ShutterPhase int32
```

#### Description

运动模糊快门相位偏移。来自 cdta `@0xB4`。UI 单位很可能是度（-90 fixture 与 AE UI 显示 -90° 一致），但未严格独立验证；以原始 int32 暴露。

#### Type

`int32`；read / write via [`SetShutterPhase`](#compositionsetshutterphase)。

---

### Composition.MotionBlurAdaptiveSampleLimit

```go
MotionBlurAdaptiveSampleLimit int32
```

#### Description

运动模糊自适应采样上限。AE 默认 128。来自 cdta `@0xC4`。

#### Type

`int32`；read / write via [`SetMotionBlurAdaptiveSampleLimit`](#compositionsetmotionbluradaptivesamplelimit)。

---

### Composition.MotionBlurSamplesPerFrame

```go
MotionBlurSamplesPerFrame int32
```

#### Description

每帧运动模糊采样数。AE 默认 16。来自 cdta `@0xC8`。

#### Type

`int32`；read / write via [`SetMotionBlurSamplesPerFrame`](#compositionsetmotionblursamplesperframe)。

---

### Composition.Layers

```go
Layers []*Layer
```

#### Description

合成内的图层列表，按 AE timeline 顺序排列（top of stack = index 0）。

#### Type

`[]*Layer`；read-only（图层内部字段是否可写见 [layer.md](layer.md)）。

---

### Composition.Markers

```go
Markers []*Marker
```

#### Description

**合成级** marker —— AE 时间线顶端的章节标记。与 `Layer.Markers` 是同一 `Marker` 类型，但作用域不同。底层存储在一个 `LIST SecL`（formType `SecL`）伪图层（Utf8 名固定 "Markers"）里。

```go
for _, m := range comp.Markers {
    fmt.Printf("@%.2fs (%.2fs) label=%d %q\n", m.Time, m.Duration, m.Label, m.Comment)
}
```

#### Type

`[]*Marker`；read-only。详见 [marker.md](marker.md)。

---

### Composition.Guides

```go
Guides []*Guide
```

#### Description

合成的**标尺参考线**（AE 从标尺拖出的对齐线）。纯 UI，不影响渲染，无 ExtendScript 等价物（py-aep parity）。底层存在 Item 级 `LIST:Gide → list → ldat`，每条 guide 16 字节。

```go
for _, g := range comp.Guides {
    fmt.Printf("%s @ %.0fpx\n", g.Orientation, g.Position) // horizontal @ 270px
}
```

`Guide` 字段：

- `Orientation GuideOrientation` —— `GuideHorizontal`（binary 2，Position = 距顶边像素）/ `GuideVertical`（binary 1，距左边像素）。`.String()` → `"horizontal"` / `"vertical"`。
- `Position float64` —— 像素偏移。

#### 写（length-preserving，Alpha）

```go
comp.Guides[0].SetPosition(540)
comp.Guides[0].SetOrientation(aep.GuideVertical)
```

原地 patch ldat 的 16 字节槽，不改 chunk 大小。共享 chunk bytes，调用方自己锁（见 [并发约束](#)）。增删 guide 是结构性操作，暂未实现。JSON 导出为 `guides[]{orientation, position}`。

---

### Composition.Renderer

```go
Renderer string
```

#### Description

合成的 3D 渲染引擎，存为 `PRin` LIST → `prin` chunk 里的 binary match_name（AE 内部画家代号）。AE Scripting 的 `CompItem.renderer` 暴露的是另一套 module name —— 只有 `ADBE Escher`（binary）↔ `ADBE Advanced 3d`（ExtendScript）不同名，其余三个两套相同：

| binary match_name | ExtendScript | UI |
|---|---|---|
| `ADBE Escher` | `ADBE Advanced 3d` | Advanced 3D（旧版 AE 为 Classic 3D） |
| `ADBE Calder` | `ADBE Calder` | Advanced 3D（AE 2025） |
| `ADBE Ernst` | `ADBE Ernst` | Cinema 4D |
| `ADBE Picasso` | `ADBE Picasso` | Ray-traced 3D |

各 AE 版本暴露的引擎不同（AE 2020 = Escher/Standard/Ernst；AE 2025 = Calder/Ernst/Picasso，且 load 时把废弃 Escher/Picasso 自动提升为 Advanced 3D）。空字符串表示该 comp 无 `PRin` LIST（罕见，仅见于跳过 AE 序列化的程序化 comp）。

#### Type

`string`；read / write via [`SetRenderer`](#compositionsetrenderer)。

---

## Methods

### Composition.LayerByID

```go
func (c *Composition) LayerByID(id uint32) *Layer
```

#### Description

按 layer ID 查图层。`id == 0` 视为 no-match（real AE 工程 layer ID 从 1 开始；0 是 `ParentID` 的"无父" sentinel）。

#### Returns

`*Layer`；找不到返回 nil。

---

### Composition.LayerByName

```go
func (c *Composition) LayerByName(name string) *Layer
```

按名称查图层（精确匹配，区分大小写）。同名图层在 AE 里**不保证唯一**；需要精确身份时用 `LayerByID`。

```go
if logo := comp.LayerByName("Logo"); logo != nil {
    logo.SetVisible(false)
}
```

---

## Setters (length-preserving)

下面这一组都是 length-preserving 的 cdta 字节写回。每个 setter 同步更新对应 Go 字段，下一次 `Project.WriteAEP` 持久化。

如果 Composition 在 parser 之外手工构造（没有 owning cdta chunk），所有 setter 返回 `error` 而不是 panic。

### Composition.SetName

```go
func (c *Composition) SetName(newName string) error
```

length-variable 改合成名（Utf8 chunk 整体替换）。

```go
comp.SetName("Final Output")
```

### Composition.SetFrameRate

```go
func (c *Composition) SetFrameRate(fps float64) error
```

写帧率到 cdta `@0x9C-0x9F`（uint16 whole + uint16/65536 fractional）。partial fps 如 29.97 / 23.976 round-trip 精确。同时按 (已有 frame count ÷ 新 fps) 重算 `Duration`，所以 SetFrameRate 后立刻读 `Duration` 也是一致的。

```go
comp.SetFrameRate(60.0)
comp.SetFrameRate(29.97)
```

### Composition.SetDuration

```go
func (c *Composition) SetDuration(seconds float64) error
```

写合成时长。底层是 cdta `@0xB0` 的 uint32 frame count = `round(seconds × FrameRate)`。**需要 `FrameRate > 0`**（必要时先 `SetFrameRate`）。

```go
comp.SetDuration(7.5)
```

### Composition.SetBGColor

```go
func (c *Composition) SetBGColor(rgb [3]uint8) error
```

写入新的合成背景色 RGB（每通道 0..255），cdta `@0x34/@0x35/@0x36`。

```go
comp.SetBGColor([3]uint8{0x33, 0x66, 0x99})
```

### Composition.SetSize

```go
func (c *Composition) SetSize(width, height uint16) error
```

写合成画布像素尺寸，cdta `@0x8C` (width) / `@0x8E` (height) uint16 BE 对，length-preserving 4 字节。不动 PixelAspect（`@0x90/@0x94`），需要时单独调 `SetPixelAspect`。`width` / `height` 任一为 0 报错。

```go
comp.SetSize(3840, 2160)  // → 4K
```

### Composition.ResolutionFactor / SetResolutionFactor

```go
ResolutionFactor [2]uint16
func (c *Composition) SetResolutionFactor(x, y uint16) error
```

合成的预览分辨率倍率（AE Scripting `CompItem.resolutionFactor`）。`[1,1]` = Full，`[2,2]` = Half，`[3,3]` = Third，`[4,4]` = Quarter；非方形比如 `[3,4]` 也合法。cdta `@0x00`（X uint16 BE）/ `@0x02`（Y uint16 BE），length-preserving 4 字节。任一为 0 拒绝。

```go
comp.SetResolutionFactor(2, 2)  // 半分辨率预览
comp.SetResolutionFactor(1, 1)  // 还原全分辨率
```

### Composition.SetShutterAngle

```go
func (c *Composition) SetShutterAngle(degrees uint16) error
```

写入运动模糊快门角度（uint16 BE，cdta `@0xAE`）。AE UI 范围 0..720。

```go
comp.SetShutterAngle(360)
```

### Composition.SetShutterPhase

```go
func (c *Composition) SetShutterPhase(phase int32) error
```

写入运动模糊快门相位（int32 BE，cdta `@0xB4`）。单位与解读以 AE 内部一致。

### Composition.SetMotionBlurAdaptiveSampleLimit

```go
func (c *Composition) SetMotionBlurAdaptiveSampleLimit(limit int32) error
```

写入运动模糊自适应采样上限（cdta `@0xC4`）。AE 默认 128。

### Composition.SetMotionBlurSamplesPerFrame

```go
func (c *Composition) SetMotionBlurSamplesPerFrame(n int32) error
```

写入每帧运动模糊采样数（cdta `@0xC8`）。AE 默认 16。

### Composition.SetWorkArea

```go
func (c *Composition) SetWorkArea(startSeconds, endSeconds float64) error
```

写入工作区起 / 终时间（秒）。底层是 cdta `@0x1C-0x2B` 的两组 dividend/divisor pair。复用已有 divisor 时（典型 600）保持原值，否则回退到 600。

```go
comp.SetWorkArea(1.5, 4.5)
```

> 同 Layer 的 setter 一样，`SetWorkArea(0, comp.Duration)` 等价于 AE UI 里的"Reset Work Area"。

### Composition.SetDisplayStartTime

```go
func (c *Composition) SetDisplayStartTime(seconds float64) error
```

写合成的显示起始时间（秒）—— cdta `@0xA4` (dividend) / `@0xA8` (divisor) uint32 BE 对，length-preserving (8 字节)。divisor 自动用 `TickRate`；`seconds == 0` 时把 dividend / divisor 都写 0 跟 AE "未设" 编码对齐。

```go
comp.SetDisplayStartTime(4.0) // 显示从 00:00:04:00 开始
comp.SetDisplayStartTime(0)   // 清回默认
```

### Composition.SetDisplayStartFrame

```go
func (c *Composition) SetDisplayStartFrame(frame int) error
```

帧数视角的便捷封装：`SetDisplayStartTime(frame / FrameRate)`。

```go
comp.SetDisplayStartFrame(120) // 从第 120 帧开始显示
```

---

## Boolean flag setters

cdta `@0x8A` / `@0x8B` 里的标志位，逆向于 [test_data/re_batch.aep](../test_data/re_batch.aep) fixture。每个 setter 翻 1 bit length-preserving。

> `dropFrame` 不在列表里 —— AE 通过 work-area divisor (600 vs 30720) 间接编码，不是单 bit；用 [`SetFrameRate`](#compositionsetframerate) 切到 NTSC 帧率即可让 AE 自动按 drop-frame 显示。

### Composition.SetDraft3D

```go
func (c *Composition) SetDraft3D(v bool) error
```

切换 Draft 3D 预览开关（cdta `@0x8A` bit 0）。开启后预览不渲染阴影 / 运动模糊 / DOF，加快交互速度。

### Composition.SetHideShyLayers

```go
func (c *Composition) SetHideShyLayers(v bool) error
```

切换 "Hide Shy Layers" 总开关（cdta `@0x8B` bit 0）。与每个 layer 的 `Shy` 标志（`SetShy`）配合：comp 开启 + layer.Shy=true 才隐藏。

### Composition.SetCompMotionBlur

```go
func (c *Composition) SetCompMotionBlur(v bool) error
```

切换合成级运动模糊总开关（cdta `@0x8B` bit 3）。**与 `Layer.MotionBlur` 独立** —— 两个都开才渲染。

### Composition.SetFrameBlending

```go
func (c *Composition) SetFrameBlending(v bool) error
```

切换合成级帧混合总开关（cdta `@0x8B` bit 4）。与每层的 `FrameBlendEnabled` 配合：两个都开才渲染。

### Composition.SetPreserveNestedFrameRate

```go
func (c *Composition) SetPreserveNestedFrameRate(v bool) error
```

切换 "Preserve frame rate when nested or in render queue"（cdta `@0x8B` bit 5）。

### Composition.SetPreserveNestedResolution

```go
func (c *Composition) SetPreserveNestedResolution(v bool) error
```

切换 "Preserve resolution when nested"（cdta `@0x8B` bit 7）。

### Composition.SetPixelAspect

```go
func (c *Composition) SetPixelAspect(par float64) error
```

写像素宽高比（PAR）。底层是 cdta `@0x90`(uint32 BE 分子) + `@0x94`(uint32 BE 分母)。

整数值 (1.0 / 2.0) 写为 `n/1`；其他写为 `round(par×100) / 100`，覆盖 AE 常见 PAR 预设（0.91 / 1.09 / 1.21 / 1.33 / 1.46 / 1.5）。

```go
comp.SetPixelAspect(2.0)  // anamorphic → 2/1
comp.SetPixelAspect(1.21) // D1 NTSC widescreen → 121/100
```

---

## Item-level setters

下面是合成作为 project 面板 Item 的元数据 setter（与 layer 级 cmta 概念相同，但作用在 comp 自身）。

### Composition.SetComment

```go
func (c *Composition) SetComment(s string) error
```

写 project 面板的注释（length-variable cmta chunk 替换；缺失时插入到 Item LIST）。

### Composition.SetLabel

```go
func (c *Composition) SetLabel(index uint8) error
```

写 project 面板的色卡索引（0..16；idta payload `@0x3A` 单字节，length-preserving）。

## Renderer setter (structural)

### Composition.SetRenderer

```go
func (c *Composition) SetRenderer(name string) error
```

切换合成的 3D 渲染引擎。`name` 可传 binary match_name（`ADBE Escher` / `ADBE Calder` / `ADBE Ernst` / `ADBE Picasso`）或 ExtendScript 名（`ADBE Advanced 3d` → 归一为 `ADBE Escher`），见 [`Renderer`](#compositionrenderer) 对照表。`prin` 的 match_name + display_name 原地改（length-preserving），`prda`（引擎专属选项）整块换成目标引擎默认模板（**结构性** —— 父 `PRin` LIST size 变，`WriteAEP` 重算）。切换会把引擎选项重置为默认（与 AE 在 Composition Settings 改 renderer 的行为一致）。

未知引擎 / comp 无 `prin·prda` back-ref（程序化 comp）/ `prin` 非 104B / 触发 parser warning（回滚）时返回 error。原子写：snapshot + warnings-as-failure + rollback。

Ship-gate：AE 2025（4/4，废弃 Escher/Picasso 被 AE load 时提升为 Advanced 3D，文件仍接受）+ AE 2020（Ernst 精确、Escher → `ADBE Advanced 3d`）双版本绿。

#### Composition.PrdaRawBytes

```go
func (c *Composition) PrdaRawBytes() []byte
```

只读返回 `prda` chunk 的 Data（引擎专属选项原始字节），无 `PRin` LIST 时返回 nil。调试 / RE 用，勿改返回的 slice（是 live chunk 数据）。
