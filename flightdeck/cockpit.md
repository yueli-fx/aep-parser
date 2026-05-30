# Cockpit — aep-parser

**Last updated**: 2026-05-31 by claude（string-literal process-坐标清理 **landed on main**（commit d4338d4）：16 处坐标（production fmt.Errorf 4 + test 诊断/skip 串 12）清空，`insert_layer_test:138` 守卫改锚语义子串 `"different Projects"`。`go vet` clean + `go test ./internal/aep/` green + §6 grep 含代码串零命中，零行为/零 API 变更。注释 + string-literal 两轮 process-meta 清理至此收口。详 logbook。）
**Active focus**: 无 active 实现线。

## Next session

1. **从下方长线 backlog 选下一条实现线**（用户定方向）。

> **长线 backlog**（大 arc 或缺 runtime setter）：

1. **Path keyframe**（逐帧 bezier）— V2.3+ 级大 arc。
2. **Layr Transform 3D 通道**（Orientation / Rotate X/Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。
3. **Stroke Line Cap/Join/Miter**（enum/scalar）— 中等价值。**RE 完**（详 `incident-reports/stroke-line-cap-join-miter-re.md`）：matchName 就是文档的 `ADBE Vector Stroke Line {Cap,Join} / Miter Limit`（旧 "property-not-found" 说法是错的）；OneD float64-BE @ cdat[0:8]（Cap 1=Butt/2=Round/3=Proj，Join 1=Miter/2=Round/3=Bevel，Miter scalar 默认 4）；AE 三者绑定一起写 + Miter 在 Join≠Miter 时 hidden。剩工：富化 stroke body 模板（含三 slot）+ runtime model 字段/enum/setter + AE 双版本 ship-gate。
4. **Gradient W**（SetGradient）— 大 arc。groundwork：默认 gradient 被 AE elide（须自定义 stops 强制 emit）；无现成 in-repo fixture；存储 `GCst > GCky > Utf8(prop.map XML)`（parse_properties.go）。需 gradient-fill fixture + XML RE + 序列化器 + API + 双版本 gate。
5. **Fill/Stroke BlendMode·CompositeOrder、Rect/Ellipse Direction** — enum，低价值，缺 runtime setter。

**其它候选**：泛型 `DuplicateItem`（无 scripting API）、`ImportComposition`（需求驱动）。其余 deferred R-only（DisplayColorSpace / ValueText 等）见 logbook § Deferred。

## Hanging tasks

无。
