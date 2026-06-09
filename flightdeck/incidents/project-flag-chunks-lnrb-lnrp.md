---
status: active
since: 2026-05-26
last_updated: 2026-05-26
when_to_read: implementing or debugging project-level flag chunks (lnrb / lnrp / similar presence-encoded settings); AE rejects writer output with "文件数据丢失"
applies_to: [project-settings, lnrb, lnrp, flag-chunk, py-aep-parity, ship-gate, ae-acceptance, scripting-api-quirk]
---

# lnrb / lnrp flag chunks — 不是空 chunk，位置敏感；lnrp ScriptingAPI readback quirk

## Symptom

P1 Task 1D (py-aep parity Project settings) 第一版实现把 `lnrb` / `lnrp` toggle 当成 "presence = true" — 添加时 `&rifx.Chunk{ID: id, Data: nil}` 直接 append 到 root.Children 末尾。AE 2025 / 2020 打开 modified .aep 报：

```
Error: After Effects错误: 文件数据丢失
```

bisect (`tmp_debug/ship_gate_1d_bisect/`) 隔出来真凶：8 个 setter 里只有 lnrb / lnrp 两个 add-chunk path 触发拒绝；纯字节改 (acer/adfr/dwga) 和 Utf8 splice (gpuG/ExEn) 全 PASS。

## Root cause

我假设了 py-aep 的 "presence-encoded toggle" 是空 chunk + 任意位置。AE 实际两条硬约束都不能违反：

1. **Chunk 有 1 byte payload**，不是空。py-aep 注释虽然说 toggle，但 `lnrb / lnrp` 注册在 `U1Chunk`（1 byte value）。AE 写出 `lnrb len=1 bytes=01`。
2. **位置固定**：紧跟 root 的 `cpid` 之后（color-profile id），`dwga` 之前。Append-to-end 跟错位都触发 "文件数据丢失"。

RE 验证：`tmp_debug/dump_root_compare/` 对 `re_linear_blending_on.aep`（AE 自己写）跟 `re_cameralight.aep`（无 lnrb）做 root children 序列 diff — AE 把 lnrb 插在 `[14]` 紧跟 `cpid`，1 byte `0x01`。lnrp 做同样实验（`re_linearize_workspace_on.aep`），相同位置/字节。

## Lesson

**Presence-encoded ≠ empty chunk**：py-aep 文档说 "toggle by add/remove" 不等于 "空 chunk 任意位置"。
- 加新 chunk 类的 setter 实现前，必须先 RE：让 AE 自己写出 enabled state，dump root chunk 序列 + chunk bytes，跟 baseline diff。
- AE 对 root chunk **位置敏感** — chunk ordering 是隐性约束，不只是 chunk 集合是否完整。
- Empty chunk 跟带 payload 的 chunk byte-for-byte 不同（8B header 一样，但缺 1B payload AE 当 malformed）。

**ScriptingAPI ≠ stored value (lnrp 子项)**：lnrp 写对了 (byte-identical with AE-saved fixture) AE 接受文件，但 `app.project.linearizeWorkingSpace` 重 open 后仍读 false。同 fixture (AE 自己 set true → save → reload → read false)。说明这字段是 OCIO/CMS profile 联动的 derived state，ScriptingAPI 不直接绑 lnrp chunk。属同 `runtime-only-fields.md` / `shutter-side-effect-divisors.md` 类的 quirk — chunk 写法正确不代表 ScriptingAPI 能 read back，必须用 byte-level diff 跟 AE-saved fixture 比，不能信 ScriptingAPI 单 round-trip。

**Lookup anchors**：固定位置的 chunk insertion 选 stable anchor。我们选 `cpid`（AE 必写 + 位置固定 root child 之一），fallback `dwga`（不在前置），fallback append。

## How to fix (locked in)

`internal/serializer/back_project.go::setRootFlagChunk`:

```go
if on && idx < 0 {
    insertAt := flagChunkInsertPosition(p.root)
    newChunk := &rifx.Chunk{ID: id, Data: []byte{0x01}}
    p.root.Children = append(p.root.Children, nil)
    copy(p.root.Children[insertAt+1:], p.root.Children[insertAt:])
    p.root.Children[insertAt] = newChunk
}

func flagChunkInsertPosition(root *rifx.Chunk) int {
    for i, c := range root.Children {
        if !c.IsList() && c.ID == chunkIDCpid { return i + 1 }
    }
    for i, c := range root.Children {
        if !c.IsList() && c.ID == rifx.IDDwga { return i }
    }
    return len(root.Children)
}
```

Ship gate validated: bisect 8/8 PASS, full modified.aep ship_gate_1d.done shows AE-visible `linearBlending=true / workingGamma=2.2 / expressionEngine="extendscript"` matching Go-side. `linearizeWorkingSpace` 单字段读 false 是 ScriptingAPI quirk (per above).

## 关联文件

| 文件 | 角色 |
|---|---|
| `internal/serializer/back_project.go` | `setRootFlagChunk` + `flagChunkInsertPosition` + `chunkIDCpid` 锚 |
| `tmp_debug/ship_gate_1d_bisect/` | per-setter 隔离 ship-gate runner (8 个 variant) |
| `tmp_debug/dump_root_compare/` | root chunk children 序列 byte-diff 工具 |
| `test_data/re_linear_blending.jsx` | AE 自己写 lnrb=true 的 RE fixture driver |
| `test_data/re_linearize_workspace.jsx` | 同 lnrp |
| `test_data/ship_gate_1d.jsx` | 全 8 setter combined ship-gate driver |
