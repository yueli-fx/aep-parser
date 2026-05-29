# V2.2.1 子项① — Ellipse embed bytes（设计）

**日期**: 2026-05-29
**状态**: design（待实现）
**前置**: V2.2 alpha（Rect+Fill embed+overwrite 已双版本 ship-gate PASS，iter-8 sealed）
**分支建议**: `v3-phase5d-ellipse-embed`

## §0 背景与定位

V2.2.1 是一捆**相互独立**的 ShapeLayer 写路径子线（Ellipse / Path / Stroke embed bytes、Fill Color 编码 RE、keyframe 持久化）。塞不进一个 spec，逐个拆开走完整 spec→plan→实现→双版本 ship-gate 循环。

本 spec 只覆盖**第一个子项：Ellipse embed bytes**。选它作头炮的理由：结构上最接近已 ship 的 Rect（Size + Position 两条 Vec2 流），复用 embed+overwrite 已证明的路径，风险最低，能最快把"每-kind fixture → 提取 boilerplate → embed → 解 ship-gate skip"流程跑通，给 Path/Stroke 立范本。

**已证明的事实**（iter-5~8 RE 结论，写死不再质疑）：shape 的 from-scratch 字节发射会触发 AE **semantic-level silent-drop**（解析不报错但 comp.layers 里丢层），byte 级 RE 不足以修复；唯一可行路径是 embed 一份 AE-saved 的 canonical body + 原位 overwrite cdat。详 `2026-05-22-v2-2-tdb4-tdbs-re-findings.md`。

## §1 目标 / 范围

**In scope**：
- `AddEllipse()` 写出的 Ellipse 被 AE 2020 + AE 2025 接受（不再 silent-drop）。
- Ellipse **Size + Position** 静态值正确持久化（经 embed body cdat overwrite）。
- AE 2020 + 2025 双版本 ship-gate PASS。
- godoc / `docs/shape.md` / `flightdeck/flight-plans/coverage.md` 收口。

**Out of scope（仍 deferred，本子项不碰）**：
- Path、Stroke embed bytes（各自后续子项）。
- Fill Color 编码 RE（正交子项）。
- keyframe 持久化（Ellipse Size/Position 动画仍走 first-kf static fallback，与 Rect 一致）。
- per-group transform、Ellipse 的其它子属性（Direction 仍 placeholder / AE 默认）。

## §2 现状（待替换的代码）

`internal/aep/lower_shape_node.go:188-200` `lowerEllipseNode` 当前 from-scratch：
```go
size := LowerVec2Stream(e.size, "ADBE Vector Ellipse Size", ...)
pos  := LowerVec2Stream(e.position, "ADBE Vector Ellipse Position", ...)
direction := emptySubPropPlaceholder("ADBE Vector Shape Direction", "Direction")
body := nodeBodyTdgp("Ellipse Path", []*rifx.Chunk{direction, size, pos})
```
→ AE silent-drop。ship-gate `shape_layer_shipgate_test.go:108-124` 因 B/C 层含 Ellipse/Stroke/Path 整体 `t.Skip()`。

参照样板：`lowerRectNode`（同文件 175-186）= `cloneShapeRectBody()` + `overwriteShapeStreamCdat(..., "ADBE Vector Rect Size", ...)`。embed 资源 `templates/v2_2_shape_rect_body.bin`，由 `tmp_debug/extract_shape_bodies` 从 `test_data/v2_2_shape_tolerance.aep` 提取。该 tolerance fixture 只含 Rect+Fill（`gen_shape_tolerance.jsx`），**无 Ellipse**。

## §3 设计

### 3.1 建 Ellipse fixture
新写 `tmp_debug/gen_shape_ellipse_tolerance.jsx`（照抄 `gen_shape_tolerance.jsx` 骨架）：
- 1 个 comp + 1 ShapeLayer + 1 Ellipse + 1 Fill（Fill 保证层可见、与 Rect fixture 对齐）。
- Ellipse 的 **Size 和 Position 都 `setValue` 成非默认值**（如 Size=[200,100]、Position=[50,30]），强制 AE 把两个 cdat 槽都写进盘——否则 AE 会 elide 默认值的槽，overwrite 就没目标（Rect 当年 Position 就是被 elide 才只能 runtime-only）。
- 存成 `test_data/v2_2_shape_ellipse_tolerance.aep` + `.done`（首行 PASS/FAIL）。
- 经 `scripts/ae_run.ps1` 自助跑（AE 2025），不动现有 `v2_2_shape_tolerance.aep`（preservation 测试钉死其字节）。

### 3.2 提取 boilerplate
扩展 `tmp_debug/extract_shape_bodies`：当前 `extractions` 表硬编码读 `v2_2_shape_tolerance.aep`。改成支持**每条 extraction 带自己的源 fixture**（加 `srcPath` 字段），新增一条：
```
{src: "test_data/v2_2_shape_ellipse_tolerance.aep",
 matchName: "ADBE Vector Shape - Ellipse",
 outPath: "internal/aep/templates/v2_2_shape_ellipse_body.bin"}
```
跑出 `templates/v2_2_shape_ellipse_body.bin`。

### 3.3 改 lower
`lower_shape_node.go`：
- `//go:embed templates/v2_2_shape_ellipse_body.bin` → `v22ShapeEllipseBodyBytes`。
- 加 `v22ShapeEllipseOnce/Cache/Err` + `cloneShapeEllipseBody()`（照抄 `cloneShapeRectBody`）。
- `lowerEllipseNode` 重写为：
  ```go
  body := cloneShapeEllipseBody()
  sz := e.size.static (animated → keyframes[0].Value)   // 与 Rect 一致的 static fallback
  overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Size", encodeF64sBE(sz[0], sz[1]))
  ps := e.position.static (同上 fallback)
  overwriteShapeStreamCdat(body, "ADBE Vector Ellipse Position", encodeF64sBE(ps[0], ps[1]))
  return body
  ```
- 删掉旧 from-scratch 段。godoc 仿 `lowerRectNode`，标注 Direction 仍 AE 默认、动画仍 first-kf fallback。

### 3.4 ship-gate（聚焦 Ellipse，避开 Path/Stroke）
现有 3 层 canonical 同时含 Ellipse/Stroke/Path，无法只验 Ellipse。新增**聚焦 ship-gate**：
- 单 ShapeLayer = Ellipse + Fill（Fill 已 ship，不引入新风险）。
- 复用 `runAeRunShipGate` harness；新 verify JSX（`test_data/verify_v2_2_ellipse.jsx`）assert：层存在于 comp.layers、含 `ADBE Vector Shape - Ellipse`、Size/Position 值匹配输入。assert-based（无 AE-native byte 基线，跟 cross-Project ship-gate 同范式）。
- 新增 `TestV2_2_Ellipse_AEShipGate_AE2020 / _AE2025`，**不 skip**。原 3 层 `TestV2_2_AEShipGate_AE20xx` 保持 skip（注释更新为"Path/Stroke 仍 deferred"）。

### 3.5 收口
- godoc：`lowerEllipseNode` 从"alpha 无 embed → silent-drop"改为已 ship。
- `docs/shape.md`：Ellipse 从 V2.2.1 deferred 表移到已支持，标注 Size+Position 持久化、Direction/动画仍受限。
- `flightdeck/flight-plans/coverage.md`：V2.2 ShapeLayer 段更新 Ellipse 状态。

## §4 风险与 fallback
- **Position 槽被 elide**：若即便 fixture setValue 了 Position，AE 仍 elide 该 cdat（或带 Position 的 body 触发 silent-drop）→ 回退 **Size-only**（Position 转 runtime-only，与 Rect 完全一致），ship-gate 只验 Size。先按 Size+Position 做，ship-gate 暴露问题再降级。
- **AE cold-start flake**：2020/2025 冷启撞 splash/About → ship-gate 偶发 exit 2 假阴性；warm retry（先跑一个已知-good fixture 热身）。已知问题，非数据。
- **extract 工具改动**：给 `extractions` 加 `srcPath` 时保持原 Rect/Fill 两条入口（读旧 tolerance.aep）不变，避免回归既有 embed bytes。

## §5 验收（DoD）
1. `go vet ./... && go test -count=1 ./internal/aep/...` 全绿。
2. `templates/v2_2_shape_ellipse_body.bin` 已生成并 embed。
3. `TestV2_2_Ellipse_AEShipGate_AE2020` + `_AE2025` 在 `AE_SHIP_GATE=1` 下 PASS（`.done` 首行 PASS）。
4. 现有 Rect/Fill ship-gate + preservation 测试无回归。
5. godoc / docs/shape.md / coverage.md 收口。
6. cockpit + manifest 更新。

## §6 后续子项（本 spec 不做，立此存照）
Ellipse 跑通后，Path / Stroke 复用同流程（各自 fixture + extract 入口 + clone+overwrite + 聚焦 ship-gate）。Stroke 更复杂（Dashes/Taper/Wave 嵌套组，见 lower_shape_node.go:239-272 现状）。Fill Color 编码 RE 与 keyframe 持久化是正交子项，独立排期。
