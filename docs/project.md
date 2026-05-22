# Project object

`aep.Open(path)` / `aep.FromReader(r)`

## Description

`Project` 是解析结果的根。持有所有合成、素材、文件夹，以及解析时遇到的非致命异常（`Warnings`）。

修改后调 `WriteAEP` 序列化回 `.aep`；调 `ToJSON` / `WriteJSON` 导出 JSON 视图（**单向**，无 `ReadJSON`）。

## Example

```go
proj, err := aep.Open("template.aep")
if err != nil { panic(err) }

fmt.Println(len(proj.Compositions), "comps,", len(proj.Footage), "footage items")
for _, w := range proj.Warnings {
    fmt.Fprintln(os.Stderr, "warn:", w)
}

// 修改 → 写回
proj.Footage[0].SetPath(`D:\new\file.png`)
out, _ := os.Create("out.aep")
defer out.Close()
proj.WriteAEP(out)
```

---

## Entry functions

### aep.Open

```go
func Open(path string) (*Project, error)
```

#### Description

从磁盘路径打开 `.aep` 文件并解析。失败返回 error。

---

### aep.FromReader

```go
func FromReader(r io.ReadSeeker) (*Project, error)
```

#### Description

从任意 `io.ReadSeeker` 解析。用于测试 / 嵌入二进制 / 网络流场景。

```go
data := []byte{ /* aep bytes */ }
proj, err := aep.FromReader(bytes.NewReader(data))
```

---

### aep.NewProject

```go
func NewProject(target ...AETarget) *Project
```

#### Description

返回一个全新空 `Project`，可继续调 `NewComposition` 等添加内容。零参数 = `TargetAE2020`（最大兼容）。支持显式 `TargetAE2020 / TargetAE2022 / TargetAE2025`。

`target` 决定**输出文件的版本标签**（svap / nhed 等头字段）—— 不影响 comp items 本身的字节结构（详见 V3 planning 段：builder 用单一 canonical seed，items 永远 AE 2020 兼容，靠 AE 的向后读取能力跨版本工作）。

**Never returns error**。模板是 build-time trusted；如果 panic 出 "build bug" 信息，那是库自身 bug，不是用户输入问题。`AETarget` 不向前兼容 —— 升库时旧二进制传 unknown target 会 panic。

```go
proj := aep.NewProject()                      // AE 2020 兼容（默认）
proj25 := aep.NewProject(aep.TargetAE2025)    // 写出 AE 25 格式标签
```

#### AETarget enum

```go
const (
    TargetAE2020 AETarget = 2020 // 默认；任何 AE 2020+ 可开
    TargetAE2022 AETarget = 2022
    TargetAE2025 AETarget = 2025
)
```

---

## Attributes

### Project.Compositions

```go
Compositions []*Composition
```

#### Description

工程中所有合成的列表。详见 [composition.md](composition.md)。

#### Type

`[]*Composition`；read-only。

---

### Project.Footage

```go
Footage []*Footage
```

#### Description

工程中所有素材项（外部文件、Solid、Placeholder）。详见 [footage.md](footage.md)。

#### Type

`[]*Footage`；read-only。Footage 项内部可通过 `SetPath` 修改路径。

---

### Project.Folders

```go
Folders []*Folder
```

#### Description

项目面板的文件夹。

#### Type

`[]*Folder`；read-only。

---

### Project.BitsPerChannel

```go
BitsPerChannel BitsPerChannel
```

#### Description

工程色深。可能值：`8`、`16`、`32`。

#### Type

`BitsPerChannel`（`uint8` 别名，带 `String()`）；read-only。

---

### Project.Warnings

```go
Warnings []string
```

#### Description

解析时收集的非致命异常 —— chunk header 看起来 sane 但 payload 不符合预期 layout（长度不一致、count 不可能等）。每条是人类可读字符串，解析继续，产出 best-effort 的 `Project`。

clean 解析时为 nil（空切片）。

```go
if len(proj.Warnings) > 0 {
    for _, w := range proj.Warnings {
        log.Println("warn:", w)
    }
}
```

#### Type

`[]string`；read-only。

---

## Methods

### Project.CompositionByID

```go
func (p *Project) CompositionByID(id uint32) *Composition
```

#### Description

按 AE 内部 ID 查合成。`id == 0` 视为 no-match（real AE 工程的 item ID 从 1 开始）。Footage / Folder 共用同一 ID 命名空间但分属不同 slice —— 非合成项的 ID 在此返回 nil。

#### Returns

`*Composition`；找不到返回 nil。

---

### Project.CompositionByName

```go
func (p *Project) CompositionByName(name string) *Composition
```

按名称查合成（精确匹配，区分大小写）。合成名在 AE 里**不保证唯一**，需要精确身份时用 `CompositionByID`。

```go
if comp := proj.CompositionByName("Main Comp"); comp != nil {
    fmt.Println(len(comp.Layers), "layers")
}
```

---

### Project.FootageByName

```go
func (p *Project) FootageByName(name string) *Footage
```

按名称查素材（精确匹配，区分大小写）。

```go
if f := proj.FootageByName("logo.png"); f != nil {
    f.SetPath(`D:\new\logo.png`)
}
```

---

### Project.WriteAEP

```go
func (p *Project) WriteAEP(w io.Writer) error
```

#### Description

把（可能已修改的）项目序列化回 RIFX 二进制。chunk 大小从当前 chunk data 重算，所以即使是 `Footage.SetPath` 这种改长度的修改也能正确序列化。

> **Best-effort** —— 本库只懂部分 `.aep` 格式；不认识的 chunk 按字节透传。如果 AE 拒绝你写出的文件，请提交一个最小复现样本。

```go
f, _ := os.Create("out.aep")
defer f.Close()
err := proj.WriteAEP(f)
```

---

### Project.ToJSON

```go
func (p *Project) ToJSON() *JSONProject
```

#### Description

把 `Project` 转成可 JSON 序列化的镜像类型树。详见 [json.md](json.md)。

---

### Project.MarshalJSON

```go
func (p *Project) MarshalJSON() ([]byte, error)
```

#### Description

标准 `json.Marshaler` 实现。等价于 `json.Marshal(p.ToJSON())`。

```go
b, _ := json.MarshalIndent(proj, "", "  ")
os.Stdout.Write(b)
```

---

### Project.WriteJSON

```go
func (p *Project) WriteJSON(w io.Writer) error
```

#### Description

把 JSON 视图直接写到 writer（带 2-space indent）。

```go
proj.WriteJSON(os.Stdout)
```

---

# Folder object

## Description

项目面板里的文件夹（不参与渲染，只是 UI 组织单位）。

## Attributes

### Folder.ID

```go
ID uint32
```

AE 内部 item ID。read-only。

### Folder.Name

```go
Name string
```

文件夹显示名。read-only。
