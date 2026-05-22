# Footage object

`project.Footage[index]`

## Description

外部素材文件、AE Solid、或 Placeholder。`Path` 字段可通过 `SetPath` 修改 —— 这是当前唯一**不受 length-preserving 限制**的写回 API（alas JSON 整 chunk 替换）。

## Example

```go
for _, f := range proj.Footage {
    fmt.Printf("%s  %dx%d  still=%v  solid=%v  path=%s\n",
        f.Name, f.Width, f.Height, f.IsStill, f.IsSolid, f.Path)
}

// 批量重定向素材
for _, f := range proj.Footage {
    if strings.HasPrefix(f.Path, `D:\old\`) {
        f.SetPath(strings.Replace(f.Path, `D:\old\`, `E:\new\`, 1))
    }
}
```

---

## Attributes

### Footage.ID

```go
ID uint32
```

AE 内部 item ID。read-only。

---

### Footage.Name

```go
Name string
```

显示名。`SetPath` 成功后会自动同步为新路径的 basename。read-only（仅 `SetPath` 间接更新）。

---

### Footage.Path

```go
Path string
```

#### Description

源文件磁盘路径。Solid / Placeholder 没有路径（为空）。

#### Type

`string`；read / write via [`SetPath`](#footagesetpath)。

---

### Footage.Width

```go
Width uint16
```

像素宽度。read-only。

---

### Footage.Height

```go
Height uint16
```

像素高度。read-only。

---

### Footage.FrameRate

```go
FrameRate float64
```

帧率（视频素材）。read-only。

---

### Footage.Duration

```go
Duration float64
```

时长（秒）。静帧素材为 0 或被 AE 当 still。read-only。

---

### Footage.IsStill

```go
IsStill bool
```

是否被 AE 标记为静帧素材。read-only。

---

### Footage.IsSolid

```go
IsSolid bool
```

是否是 AE Solid（合成内生成的固色层）。Solid 没有磁盘路径。read-only。

---

## Methods

### Footage.SetPath

```go
func (f *Footage) SetPath(newPath string) error
```

#### Description

改素材源文件路径。具体动作：

- 重写 alas / Als2 / Pin chunk 里的 JSON `fullpath` 字段（chunk 整体替换 —— 长度可变）。
- 兼容旧版：若同时存在 `Cpth` chunk（旧 AE 版本的明文路径），也一并替换。
- 同步更新 `Footage.Path` 字段值和 `Footage.Name`（设为新路径的 basename）。

下一次 `Project.WriteAEP` 会持久化变更。

> Solid / Placeholder 没有路径 chunk，调用返回 error。

#### Returns

`error`；若该素材没有可写路径 chunk（如 Solid），返回 `"no path chunks present"` 错误。

#### Example

```go
err := proj.Footage[0].SetPath(`D:\new\location\logo.png`)
if err != nil { log.Fatal(err) }
```
