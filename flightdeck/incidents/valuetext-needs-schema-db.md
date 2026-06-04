---
status: active
when_to_read: 想给 Property 加 ValueText / propertyParameters；以为 dropdown 选项 label 能从字节 RE；评估 py-aep parity P3 §3H 是否值得做
applies_to: [property, valuetext, propertyparameters, dropdown, enum, pard, pdnm, nbOptions, schema-database, scripting-api-gap, negative-finding, ae26, p3-3h]
last_updated: 2026-06-04
---

# Property.ValueText（§3H）—— 通用实现不可达，need external schema DB

## 结论（2026-06-04 决策：整个 3H defer / won't-implement）

`Property.valueText` 是 AE **26.0 才新加**的 ScriptingAPI（连同 `Property.propertyParameters`，见
`charts/after-effects-scripting-guide/docs/property/property.md` + changelog）。语义 = dropdown 当前选中项的 label，
等价 `propertyParameters[value]`（1-based 索引，分隔符 `"(-"` 占位）。

忠实的**通用** valueText 从 .aep 字节**做不到**，故整项 defer。

## 为什么不可达（核心 RE 发现）

label 数组（propertyParameters）按 dropdown 类型分两类，存储本质不同：

| 类型 | 例子 | label 存哪 | 可 RE? |
| --- | --- | --- | --- |
| **内置枚举** | Blend Mode（Normal/Multiply/…）、`dropShadow/mode2` | pard chunk 里**只有 `nbOptions` 计数**（我们 `parse_effect_pard.go` 已解析），label 字符串是 AE 插件内置、随 AE 版本变 | ❌ 需外部 schema DB |
| **自定义 Dropdown Menu Control** | 用户手打 "Sunday"/"Monday"… | label 用户定义，存在文件 `pdnm` chunk（我们现在 skip，连 chunk ID 都没在 rifx 定义） | ✅ 有界可 RE |

内置枚举占绝大多数，其 label 不在文件里 —— 要从 value=5 还原 "Multiply" 必须有一张
per-effect × per-AE-version 的 (matchName, index) → label 表。**Adobe 不公开此表**，需逐 effect RE，无界。

## 旁证

- **py-aep 自己也没实现 valueText** —— 在它 `docs/extendscript_coverage.md` 的 Property missing 🚧 列表里
  （`alternateSource, ..., valueText`）。它 `samples/versions/ae2026/complete.json` 里的 `"valueText": "Multiply"`
  是 AE 26 ScriptingAPI **导出的 golden**，不是 py-aep 的解析能力。所以 parity 目标本身不要求我们做。
- valueText 是 AE 26.0-only，远超本项目 AE 2020 读取下限。

## 如果将来真要做（有界子集）

唯一字节可恢复的子集 = **自定义 Dropdown Menu Control**：
1. 在 rifx 定义 `pdnm` chunk ID，RE 其 layout（option 字符串表，含分隔符编码）。
2. 在 enum pard（`PCTLEnum`）旁关联同 param 的 pdnm，解出 label 数组。
3. 暴露 `Property.PropertyParameters []string` + `Property.ValueText`（= `PropertyParameters[StaticValue]`，越界/空表返回 ""）。
4. 内置枚举显式返回 "" 或标 needs-DB，**不**尝试合成。

但 payoff 窄（自定义 dropdown 实际项目少见），当前决策是不做。
