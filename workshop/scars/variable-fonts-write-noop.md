# Variable fonts axes 写 = ScriptingAPI 不存在该字段

## 症状

JSX 里 `td.fontVariation = {wght: 700}` 没报错，但 btdk 字节零变化。一开始以为是 read-only field，实际是 *根本没这个 key*。

## 根因

ExtendScript 允许对任意 key 赋值不报错（即使 key 不存在）。`TextDocument` 上没有 `fontVariation` 这个属性，赋值就是给 JS object 加一个无关联的 attribute。

实测 `TextDocument` 真实 keys（probe `for (var k in td)` 得）里没有 `fontVariation`，也没有 `setFontVariationAxisValue`。

## 教训

- **API 不抛错 ≠ API 真生效**。每次怀疑 ScriptingAPI 行为不正常，先 probe `for (var k in obj)` dump 真实 key 列表。
- 修字段必须 *跑 RE → 字节 diff* 才算确认，不能只看 JSX 不抛错就以为通了。
- Variable fonts axes 在 btdk 里 *是* 持久化的（`/0/1/0[i]/0/0/4` 16.16 fixed-point），只是写不通过 ScriptingAPI；唯一写途径 = `td.font = "Bahnschrift-Bold"` 之类切已加载的 named-instance，AE 自动重写 `/4` 数组。
- 这种 finding 进 `coverage.md` 的 ⚠ Negative findings 段（不是 ❌ 不可达，也不是 🗑️ 暂搁 —— 是 "AE 限制" 这种特殊类别）。
