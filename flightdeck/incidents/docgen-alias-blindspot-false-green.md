---
status: active
when_to_read: 改 cmd/docgen 的包扫描/类型抽取；新增任何 facade/别名层（type X = pkg.Y）后发现 docs 段变空；写「生成物 vs 已提交」对账型测试时想确认它能不能抓回归；排查「docgen test 绿但 docs 实际不对」
applies_to: [docgen, go-doc, type-alias, facade, generated-vs-committed-test, false-green, package-split, regression]
---

# docgen 单包扫描 + 类型别名 → 静默空文档（且对账测试假绿）

## 症状

M8 物理分包 stage-1（`ac62b25`）把运行时类型从 `internal/aep` 迁到 `internal/scene`，`internal/aep` 退化为别名 facade（`type Composition = scene.Composition`）。此后 `docs/*.md` 里**所有类型段全空**——`# Composition object` / `# Layer object` 等标题下零字段/零方法/零常量，只有 re-export 自由函数（`## Functions`）还在。直到 2026-06-09 才被发现，期间一直 committed 着降级文档。

## 根因（两层）

1. **docgen 扫单包 + `go/doc` 不解析别名**：`docgen.json` 的 `pkg` 指 `../internal/aep`；`go/doc.NewFromFiles` 对 `type X = scene.X` 别名只记一个壳，**不带出** `scene.X` 的字段/方法（它们属 scene 包，不在 aep AST 里）。`extractFields` 遇别名 TypeSpec（`*ast.SelectorExpr`，非 `StructType`）→ 0 字段；`withMethods` 在 aep 包找不到方法 → 0 方法。
2. **对账测试假绿**：`TestDocsUpToDate` 比「内存重生成」vs「已提交 docs」。stage-1 重生成时把 committed 也写成了同样的降级态 → `generated == committed` → **PASS**。测试只保证「生成器自洽」，**不保证「内容正确」**——双双降级时它发现不了。

## 修复（`1bc4ad4`）

docgen 改**多包扫描**：`manifest.pkgs = [aep facade, scene, codec]`。`extractTypesMulti` 同名类型偏好**真定义**（`isAlias` 检测：`ts.Assign.IsValid()`）而非别名壳；`withMethodsMulti` 跨包累加方法（真定义包在 facade 之后 → 其 setter 决定字段 R/RW）；`extractPackageFuncsMulti` 自由函数 facade-first。恢复 +11.8k 行类型文档（含降级期间新加却从没进文档的符号，如 `Property.OwnerLayer`）。

## 教训（可复用）

- **「生成物 vs 已提交」对账测试有结构性盲区**：当一次改动同时劣化「生成逻辑」和「重生成的产物」，gen==committed 仍 PASS。这类测试需配一个**内容锚**（如断言关键类型至少有 N 个字段/方法，或对账一份手写的 golden 期望），不能只靠自我对账。
- **别名 facade 对反射/AST 工具透明但对 `go/doc` 不透明**：任何「扫一个包、按 AST 抽符号」的工具，遇到 `type X = otherpkg.Y` 都会丢 Y 的成员。加 facade/别名层后，务必让这类工具多包扫描或解析别名目标。
- 排查 codegen「测试绿但产物不对」时，先看**测试是不是在跟自己对账**。
