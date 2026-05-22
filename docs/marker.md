# Marker object

`layer.Markers[index]` 或 `composition.Markers[index]`

## Description

时间线 marker。**同一类型用于两种位置**：

1. **`Layer.Markers`** —— 图层 marker（鼠标右键 layer → Markers → Add Marker）
2. **`Composition.Markers`** —— 合成级章节 marker（在 timeline panel 顶端的标尺上加，比 layer marker 更"全局"）

两者底层存储不同但结构完全一致：

- Layer marker：图层属性树里的 `"ADBE Marker"` 属性
- Comp marker：合成内一个 `LIST SecL`（Utf8 名固定为 "Markers"）伪图层里的 `"ADBE Marker"` 属性

解码器是同一份代码（`parseMarkers`）。

## Example

```go
// 合成级 markers
for _, m := range proj.Compositions[0].Markers {
    fmt.Printf("[comp] @%.2fs (%.2fs) label=%d %q\n",
        m.Time, m.Duration, m.Label, m.Comment)
}

// 图层 markers
for _, m := range proj.Compositions[0].Layers[0].Markers {
    fmt.Printf("[layer] @%.2fs %q\n", m.Time, m.Comment)
    if m.URL != "" { fmt.Println("  URL:", m.URL) }
}
```

---

## Attributes

### Marker.Time

```go
Time float64
```

marker 时间（秒），按 owning comp 的 `TickRate` 换算。read / write via [`SetTime`](#markersettime)。

---

### Marker.Duration

```go
Duration float64
```

#### Description

marker 持续时长（秒）。

- `Duration == 0` → 点 marker（▼ 三角形）
- `Duration > 0` → 区间 marker（带尾巴的 ▼）

来自 NmHd `@0x08` / 600（600ths-of-a-second 单位）。

#### Type

`float64`；read / write via [`SetDuration`](#markersetduration)。

---

### Marker.Label

```go
Label uint8
```

时间线色卡索引，0..16。0 = AE 默认色。来自 NmHd `@0x10`。read / write via [`SetLabel`](#markersetlabel)。

---

### Marker.Comment

```go
Comment string
```

主备注（AE Marker dialog 的 "Comment" 字段）。来自 Nmrd block 的第 1 个 Utf8 chunk。read / write via [`SetComment`](#markersetcomment)。

---

### Marker.Chapter

```go
Chapter string
```

章节链接（AE Marker dialog 的 "Chapter" 字段）。来自 Nmrd 第 2 个 Utf8。read / write via [`SetChapter`](#markersetchapter)。

---

### Marker.URL

```go
URL string
```

Web target URL（AE Marker dialog 的 "URL" 字段）。来自 Nmrd 第 3 个 Utf8。read / write via [`SetURL`](#markerseturl)。

---

### Marker.FrameTarget

```go
FrameTarget string
```

Frame Target ID（用于交互式视频跳转）。来自 Nmrd 第 4 个 Utf8。read / write via [`SetFrameTarget`](#markersetframetarget)。

---

### Marker.CuePointName

```go
CuePointName string
```

Cue Point Name（Flash 时代遗留功能；现代 AE 项目通常为空）。来自 Nmrd 第 5 个 Utf8。read / write via [`SetCuePointName`](#markersetcuepointname)。

---

---

## Setters

下面这一组在 parser 创建的 Marker 上有效。手工构造的 Marker（不通过 `Open` / `FromReader`）没有 owning chunk 引用，setter 返回 error。

length-preserving setters（固定偏移字节写）：

### Marker.SetTime

```go
func (m *Marker) SetTime(seconds float64) error
```

按 owning composition 的 `TickRate` 把秒换算成 uint32 ticks 写入 ldat。

```go
m.SetTime(2.5)
```

### Marker.SetDuration

```go
func (m *Marker) SetDuration(seconds float64) error
```

写入 NmHd `@0x08`（uint32 BE，600ths-of-a-second）。`seconds == 0` 产生点 marker。负值 clamp 到 0。

### Marker.SetLabel

```go
func (m *Marker) SetLabel(index uint8) error
```

写入 NmHd `@0x10` 时间线色卡索引（0..16）。

---

length-variable setters（Utf8 chunk 整体替换；WriteAEP 自动重算父 LIST size）：

### Marker.SetComment

```go
func (m *Marker) SetComment(s string) error
```

改主备注（第 1 个 Utf8）。

### Marker.SetChapter

```go
func (m *Marker) SetChapter(s string) error
```

改章节链接（第 2 个 Utf8）。

### Marker.SetURL

```go
func (m *Marker) SetURL(s string) error
```

改 web target URL（第 3 个 Utf8）。

### Marker.SetFrameTarget

```go
func (m *Marker) SetFrameTarget(s string) error
```

改 Frame Target ID（第 4 个 Utf8）。

### Marker.SetCuePointName

```go
func (m *Marker) SetCuePointName(s string) error
```

改 Cue Point Name（第 5 个 Utf8）。

> 文本 setter 按"声明顺序"写入。如果你要写第 N 个 slot 但前面的 Utf8 槽尚未存在（罕见），本库会自动用空 Utf8 chunk 补齐前导槽。

---

## 当前限制

- 增删 marker（在已有 mrst 里加 / 删条目）不支持 —— 需要同步重排 ldat 和 Nmrd 列表，破坏 length-preserving。
