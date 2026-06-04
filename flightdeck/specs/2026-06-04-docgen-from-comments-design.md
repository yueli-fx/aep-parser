---
status: active
summary: docs/*.md 改为从 Go doc comment 自动生成（Swagger 式，AST 推导结构 + 注释只写 prose + Example 函数）；先 pilot property.md
---

# docgen — 从 Go doc comment 自动生成 API 文档

**Status**: design（已 brainstorm 定稿，待写 plan）。**Created**: 2026-06-04。

## 动机

`docs/*.md` 现为 ~5600 行手工策展中文文档（仿 docsforadobe，13 文件，按类型分），与代码**双重维护**——改一个 setter 要同步改 .md，易漂移、且烦。改为 **Swagger 式：唯一源放 `.go` doc comment，脚本自动生成 `docs/*.md`**。

## 已定决策（brainstorm 三问）

1. **唯一源 = `.go` doc comment**（纯生成，不再手写 .md）。
2. **AST 推导结构 + 注释只写 prose**：签名 / 参数 / 返回类型 / JSON 字段 / R·RW 分类全从 AST + struct tag + setter 探测自动推；doc comment 只写描述 prose；示例用 `Example*` 测试函数（编译校验）。**不**用 Swagger `@param` 标签，**不**把整块 markdown 塞注释。
3. **范围 = 先 pilot `property.md`** 跑通闭环验证风格/保真，再增量推广其余 12 文件。

## 架构

```
cmd/docgen/
  main.go     : 读 manifest → go/doc 解析 internal/aep → 渲染 → 写 docs/*.md
  extract.go  : go/doc.NewFromFiles + go/ast 抽取（符号树 / 签名 / struct tag / setter 探测 / Example 关联）
  render.go   : 符号树 → markdown（复刻现有 Attributes / Methods / 示例 结构）
docs/docgen.toml : manifest — 输出文件 → 根类型映射（pilot 只填 property）
docs/_includes/  : 逃生舱 — 非符号 prose（类型间叙事 / 概念表 / README 类层级）手写 head/tail，生成时拼接
```

引擎选 **标准库 `go/doc`+`go/ast` 自研**（非 `go doc` 文本 scrape，非 gomarkdoc 三方——后者输出通用，匹配不了 R/RW + JSON 字段 + 按类型分文件的定制）。`go/doc.NewFromFiles` 一次拿全符号 + 注释 + Example 关联。

## 数据来源映射（注释零重复）

| 输出元素 | 来源 |
|---|---|
| 签名 / 参数 / 返回类型 | AST |
| JSON 字段名 | struct tag `json:"..."` |
| **R vs RW** | 字段有无对应 `Set<Field>` 方法；方法是否 `Set*` |
| 描述 prose | doc comment（**唯一手写源**） |
| 示例 | `Example<Type>_<Method>` 测试函数（godoc 惯例关联，编译校验） |
| 按类型分文件 | manifest |

## 渲染规则

| 现状 md 结构 | 生成规则 |
|---|---|
| `# <Type> object` + intro | 类型 doc comment |
| `## Attributes` 字段 | 导出 struct 字段 → `### Type.Field` + 类型代码块 + prose + R/RW marker |
| `## Attributes` 只读方法（ControlType/MinValue…） | 无参返回值、非 `Set*` 的方法 → 归 Attributes（匹配现状把这类方法当属性的惯例） |
| `## Methods` | `Set*` / 动作方法（Insert/Delete）→ `### Type.Method` + 签名 + prose + Example |
| `## X constants` | 导出常量块 |

**格式取舍**：现状 `#### Returns` 小节写「何种情况返回 error」是 AST 推不出的 prose——作者把错误条件**自然写进 doc comment prose**，生成器渲染 `签名 → prose → 示例`，不再硬分 Returns 小节（比现在略松，符合「注释只写 prose」）。

### 示例：SetDimensionsSeparated 前后

**源（doc comment）**：

```go
// SetDimensionsSeparated 切换 AE 的 "Separate Dimensions"。结构性写（改 chunk 树
// 结构，非 length-preserving），走 V2.1 atomic invariants：失败无副作用。仅作用于
// 3 维 Position leader（MatchName=="ADBE Position"）；其它属性返回 error。
//
// separate：leader 重置默认 [w/2,h/2,0]，真值迁入按轴 follower（3D→Position_0/1/2，
// 2D→仅 X/Y）。merge：leader 取回 [X,Y,Z]，删全部 follower。动画 Position 自动走
// keyframe 流迁移。双版本 ship-gate：静态 6/6 + 动画 8/8。
func (p *Property) SetDimensionsSeparated(separated bool) error {
```

**生成（docs/property.md 片段）**：`### Property.SetDimensionsSeparated` 标题 + 签名代码块 + prose 原样 + 来自 `ExampleProperty_SetDimensionsSeparated` 的示例代码块。作者只维护 prose + Example 函数。

## 非符号 prose 归宿

- 类型级 intro → 放该类型的 doc comment。
- 跨类型叙事（README 类层级图、概念表、导航表）→ 走 `docs/_includes/<file>.head.md` / `.tail.md` 手写，生成时拼接。不强塞进某个符号。
- `README.md` 不在生成器管辖（纯手写导航）。

## 铁律调和

CLAUDE.md「不写注释，除非 WHY 不明显」改述为：**「内部实现无注释；导出 API 的 doc comment = 文档源（docgen 生成 docs/*.md）。」** 写进 `flightdeck/rules.md` House rule。区分点：行内实现注释仍禁；导出符号上方的 doc comment 是文档载体，不算违反。

## Pilot 范围 + 验证闭环

1. 写 `cmd/docgen` + `docs/docgen.toml`（manifest 只填 property → [Property, Keyframe, TemporalEase, InterpType]）。
2. 把现有 `property.md` 的 prose 迁进对应符号的 doc comment（分布在 scene_property*.go / mutate_property*.go / codec keyframe 等），建若干 `Example*` 函数。
3. `go generate ./...`（或直接 `go run ./cmd/docgen`）→ 生成 `docs/property.md`。
4. **验证**：`git diff` 旧 vs 生成版，人工核对覆盖无丢失 + 风格可接受；`go test ./...`（Example 编译/跑通）。
5. 可接受 → manifest 推广其余 12 文件（增量，逐文件迁移 + 验证）；不可接受 → 调 render 规则重生成。

> 复用素材：当前工作树未提交的 property.md DimSep 增补（reader + SetDimensionsSeparated）正好是 pilot 要迁进注释的 prose 源；不浪费。

## 风险 / 开放问题

- **生成保真**：AST 推导的结构比手工策展更统一、可能丢失某些 bespoke 排版（如 Components 值表、keyframe byte-layout 概念表）。缓解：这类概念表走 `_includes/` 或类型 doc comment 内联 markdown 表（go/doc 注释支持有限 markdown）。pilot 阶段核对。
- **go/doc 注释 markdown 能力**：Go 1.19+ doc comment 支持有限 markdown（列表、代码块、链接），但不支持复杂表格。复杂表格 → `_includes/` 或渲染器特殊处理。pilot 验证够不够用。
- **迁移工作量**：5600 行一次性搬进注释是真成本，但 pilot 先验证再分批摊。
- **Example 关联粒度**：godoc `Example<Type>_<Method>` 命名约定必须严格，否则关联不上。生成器对缺失 Example 的符号容忍（无示例段）。

## 不做（YAGNI）

- 不做 Swagger UI / HTML / OpenAPI JSON——只生成现有 markdown。
- 不做 README.md 生成（纯手写导航）。
- 不一次性迁全 13 文件（pilot 先行）。
- 不引入 Swagger `@tag` 注解语法。
