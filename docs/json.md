# JSON view

`project.ToJSON()` / `project.MarshalJSON()` / `project.WriteJSON(w)`

## Description

`Project` 提供一组 snake_case 的 JSON 镜像类型 (`JSON*`)，方便导出到外部消费者（前端 inspector / CI 报告 / diff 工具）。

> ⚠️ **单向**：没有 `ReadJSON` 反序列化。JSON 视图不保留底层 RIFX 字节，所以无法从 JSON 重建可写回的 `Project`。**任何修改必须走二进制路径**（`Open` → `Set*` → `WriteAEP`）。

## Example

```go
// 标准 json 包接口
b, _ := json.MarshalIndent(proj, "", "  ")
os.Stdout.Write(b)

// 直接写到 writer
proj.WriteJSON(os.Stdout)

// 拿到中间结构体自己处理
jp := proj.ToJSON()
fmt.Println(len(jp.Compositions), "comps in JSON view")
```

---

## Methods

### Project.ToJSON

```go
func (p *Project) ToJSON() *JSONProject
```

把 `Project` 转成可序列化的镜像类型树。read-only。

---

### Project.MarshalJSON

```go
func (p *Project) MarshalJSON() ([]byte, error)
```

标准 `json.Marshaler` 实现。等价 `json.Marshal(p.ToJSON())`。

---

### Project.WriteJSON

```go
func (p *Project) WriteJSON(w io.Writer) error
```

把 JSON 视图（含 2-space indent）直接写到 writer。

---

## JSON 类型映射

| Go 核心类型 | JSON 镜像类型 |
|---|---|
| `Project` | `JSONProject` |
| `Composition` | `JSONComposition`（含 `markers`） |
| `Layer` | `JSONLayer` |
| `Property` | `JSONProperty` |
| `Keyframe` | `JSONKeyframe` |
| `TemporalEase` | `JSONTemporalEase` |
| `Effect` | `JSONEffect` |
| `Marker` | `JSONMarker` |
| `Mask` | `JSONMask` |
| `MaskVertex` | `JSONMaskVertex` |
| `MaskPathKeyframe` | `JSONMaskPathKeyframe` |
| `ShapePath` | `JSONShapePath` |
| `TextSource` | `JSONTextSource` |
| `TextStyleRun` | `JSONTextStyleRun` |
| `TextParagraph` | `JSONTextParagraph` |
| `Footage` | `JSONFootage` |
| `Folder` | `JSONFolder` |

---

## 命名与编码约定

- 所有字段都是 snake_case（如 `frame_rate`、`work_area_start_seconds`、`is_box_text`）
- 时间值用 `_seconds` 后缀
- 颜色用 `#RRGGBB` 字符串（Composition.BGColor / Mask.Color）
- 浮点数大多 round 到 3 / 4 位小数（消除浮点噪声便于 diff）
- 默认值（boolean false / 空 slice）用 `omitempty` 忽略

---

## JSONProject 字段示例

```json
{
  "compositions": [
    {
      "id": 17,
      "name": "Main",
      "width": 1920,
      "height": 1080,
      "frame_rate": 30,
      "duration_seconds": 5,
      "tick_rate": 30720,
      "bg_color": "#000000",
      "work_area_start_seconds": 0,
      "work_area_end_seconds": 5,
      "shutter_angle_degrees": 180,
      "shutter_phase": 0,
      "motion_blur_adaptive_sample_limit": 128,
      "motion_blur_samples_per_frame": 16,
      "layers": [
        {
          "index": 0,
          "name": "Logo",
          "type": "av",
          "visible": true,
          "properties": [...],
          "effects": [...],
          "markers": [...],
          "text_source": {
            "text": "Hello",
            "fonts": ["Myriad-Roman"],
            "runs": [
              {
                "font_index": 0,
                "font_name": "Myriad-Roman",
                "font_size": 50,
                "fill_color": [1, 1, 1, 1]
              }
            ],
            "paragraphs": [{"justification": "Left"}],
            "justification": "Left"
          }
        }
      ],
      "markers": []
    }
  ],
  "footage": [...],
  "folders": [...]
}
```

---

## 与 `Project.WriteAEP` 的关系

| 操作 | 走哪条路径 |
|---|---|
| 只读检查 / 报告 | `WriteJSON` |
| 修改并保存 | `Open` → 各种 `Set*` → `WriteAEP` |
| 从 JSON 重建 .aep | ❌ 不支持 |

如果你需要从外部数据驱动 .aep 生成，目前的可行路径是：

1. 在 AE 里手动准备一个模板 .aep
2. 用本库 `Open` 模板
3. 按外部数据调 `SetPath` / `SetStaticValue` / `SetTime` / `SetValue` / `SetText`
4. `WriteAEP` 写新文件

length-preserving 限制意味着外部数据只能驱动现有 keyframe 的值变化、文本的等字节替换等 —— 不能新增 / 删除结构。
