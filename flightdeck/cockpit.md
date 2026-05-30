# Cockpit — aep-parser

**Last updated**: 2026-05-30 by claude（注释纪律清理 pass **landed on main**（merge of `chore/comment-discipline-cleanup`，cleanup commit 4fcdf0c）：51 文件，`internal/aep` ~234 处 process-meta 注释债（Phase/Inv/iter/RE-S/spec §/doc 路径/日期戳/TODO）清空，保留 Mirrors py-aep·AE-script 参照 + deprecated 域事实 + 每条 RE finding 事实。手工 19 处定 rubric 后 6 并行 subagent 分文件清 + 中央复核补漏 1 处。`go vet` + `go test ./...` 全绿，零行为/零 API 变更。残留 14 处坐标在 string literal 留作可选 follow-up。详 logbook。）
**Active focus**: 无 active 实现线。

## Next session

1. **（可选，小）string-literal 内 process-坐标清理** — 清 14 处藏在 `fmt.Errorf` 报错串 + test 诊断串里的 Phase/RE-S/Inv 坐标（`insert_layer_test.go:138` 断言 `strings.Contains(…, "deferred to Phase 5C.1")` 与产线串耦合，须两侧同改）→ 让 §6 grep 含代码串也零命中。非注释、出 comments.md scope，纯洁癖。
2. **否则从下方长线 backlog 选下一条实现线**（用户定方向）。

> **长线 backlog**（大 arc 或缺 runtime setter）：

1. **Path keyframe**（逐帧 bezier）— V2.3+ 级大 arc。
2. **Layr Transform 3D 通道**（Orientation / Rotate X/Y / Position_Z）— 需先有 3D layer 支持（runtime 无 3D switch，V2.3）。
3. **Stroke Line Cap/Join/Miter**（enum/scalar）— 中等价值。groundwork：stroke body **不含**这些 slot（默认值被 elide），matchName **不是** `ADBE Vector Stroke Line Cap`（JSX property-not-found）。需先查真实 matchName + 富化 stroke body + 加 runtime model 字段/enum/setter。
4. **Gradient W**（SetGradient）— 大 arc。groundwork：默认 gradient 被 AE elide（须自定义 stops 强制 emit）；无现成 in-repo fixture；存储 `GCst > GCky > Utf8(prop.map XML)`（parse_properties.go）。需 gradient-fill fixture + XML RE + 序列化器 + API + 双版本 gate。
5. **Fill/Stroke BlendMode·CompositeOrder、Rect/Ellipse Direction** — enum，低价值，缺 runtime setter。

**其它候选**：泛型 `DuplicateItem`（无 scripting API）、`ImportComposition`（需求驱动）。其余 deferred R-only（DisplayColorSpace / ValueText 等）见 logbook § Deferred。

## Hanging tasks

无。
