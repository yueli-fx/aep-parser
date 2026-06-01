---
status: active
---

# P3 §3A slice-1 — Render Queue R-only reader

**Created**: 2026-06-01
**Spec**: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md) §2.6 + §3 Phase 3A
**Scope**: 只读暴露 render queue 顶层结构。写区（全 settings enum / format options × 7 / OutputModule 写）**全部 defer 后续 slice**。

## 背景 — RE 已由 py-aep 解完

P3 spec §4.1 把 RQHF/RQMF 标"未解"是**我们的**状态；参照源 [`../charts/py-aep`](../charts/py-aep) 已完整 RE，且自带 80+ render queue fixture（`samples/models/renderqueue/*.aep`）+ 对应 golden JSON。本 slice 把 RE 降级成"读 py-aep 源 + golden 对账"。

参照源关键文件：
- `src/py_aep/binary/render_chunks.py` — chunk 字节布局（RenderSettingsItem 2246B / OutputModuleSettingsItem 128B / Roou / Ropt / Rout）
- `src/py_aep/parsers/render_queue.py` + `output_module.py` — 挂载逻辑
- `src/py_aep/models/renderqueue/*` — 字段语义

## chunk family 结构

```
root
└ LIST:LRdr                       ← render queue 容器（root 直接子）
  ├ LIST:list → lhd3(count) + ldat(RenderSettingsItem × N, 2246B/项)   ← comp_id/status/timespan/settings
  └ LIST:LItm                     ← RQItemCollection
     └ per item（顺序）:
        [RCom → Utf8(comment)]    ← 可选
        LIST:list → ldat(OutputModuleSettingsItem × M, 128B)
        LIST:'LOm '               ← output modules：每个 Roou 起新 OM；split_on_type(children, "Roou")
           per OM: Roou + [Ropt] + [hdrm + Utf8] + LIST:Als2(alas JSON) + Utf8(name) + Utf8(file_name_template)
```

item 配对：遍历 LItm children，`list` 暂存→遇 `LOm ` 闭合一项，`item_index++` 取 settings ldat[item_index]。

## RenderSettingsItem 字节偏移（BE，slice-1 用到的）

| 字段 | offset | 类型 |
|---|---|---|
| comp_id | 0x08 | u4 |
| status | 0x0C | u4 |
| time_span_start_dividend | 0x14 | u4 |
| time_span_start_divisor | 0x18 | u4 |
| time_span_duration_dividend | 0x1C | u4 |
| time_span_duration_divisor | 0x20 | u4 |
| time_span_source | @0x... (see render_chunks.py) | u2 |
| template_name | 0x5A | str(64) win-1252 |

time_span 解析（py-aep `_resolved_time_span`）：source==LENGTH_OF_COMP→(0, comp.Duration)；WORK_AREA→(comp.WorkAreaStart, comp.WorkAreaDuration)；else→(start_div/divisor, dur_div/divisor)。

## 落地步骤（TDD，每步先红后绿）

1. **rifx ID 常量** — 补 `IDLRdr / IDLItm / IDLOm（'L','O','m',' '） / IDRCom / IDRoou / IDRopt / IDRout`。`IDAls2 / IDAlas / IDLhd3 / IDLdat / IDkfl("list")` 已有。
2. **scene 类型** — `scene_render_queue.go`：`RenderQueue{ Items []*RenderQueueItem }`、`RenderQueueItem{ Comp *Composition; Status uint32; Name string; Comment string; ... TimeSpanStart/Duration float64; OutputModules []*OutputModule }`、`OutputModule{ Name, FileTemplate, FullPath string }`。`Project.RenderQueue *RenderQueue` 字段 + accessor。
3. **codec decoder** — `codec_render_settings.go`：纯字节 → RenderSettingsItem 值（无 scene 耦合，守 `codec_` 边界）。
4. **parser** — `parse_render_queue.go`：walk LRdr → settings ldat + LItm 配对 → 链接 comp_id → Composition。在 `parseProject` 收尾（initDerived 后，comp 已就位）调用。
5. **JSON 导出** — `write_json.go` 加 `renderQueue` 节点（对齐 golden 字段名）。
6. **测试** — `parse_render_queue_test.go`：用 `numItems_1.aep` + `numItems_2.aep` + `comment_aaaaa.aep` + `empty.aep` fixture，对 golden JSON 关键字段断言（NumItems / compName / timeSpan / OM name+file_template / comment）。fixture 从 py-aep samples 复制进 `test_data/`（非敏感，见 memory）。

## 验证

- `go vet ./... && go test ./...`
- golden 对账：我们的字段 vs py-aep `samples/models/renderqueue/*.json`
- **无 AE ship-gate**（纯 R，不写回）。round-trip byte-identical 由现有 opaque preservation 保证（不解的 chunk 原位透传）。

## 非目标（defer 后续 slice）

- 任何写：status/comment/name/timeSpan setter、OutputModule.file 写
- 全 RenderSettings enum 映射（quality/effects/motion_blur/... 26 个）+ get_settings
- Format options × 7（Cineon/Jpeg/OpenExr/Png/Targa/Tiff/Xml）
- OutputModule settings（128B OutputModuleSettingsItem 全字段）
- `file` 模板变量解析（[compName]/[fileextension]/[width]...）

## 风险

- template_name / time_span_source 精确 offset 需对着 `render_chunks.py` 数 reserved 块累加，易差 1 — 用 golden 对账兜底。
- 旧 AE（<2024）OM 无 hdrm，Als2 前无 Utf8——Name/FileTemplate 取 Als2 *之后*的 Utf8，与 hdrm 无关，逻辑一致。
- empty render queue：lhd3.count==0 → Items 空，早返回。
