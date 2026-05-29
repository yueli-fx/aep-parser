# Path embed bytes — RE findings（V2.2.1 子项②预备）

**日期**: 2026-05-29　**状态**: ✅ RESOLVED — 子项② Path 已 ship（embed+splice，AE 2020+2025 双版本 ship-gate PASS，HEAD 8a85fb5）。ldat 编码已修（见下"待办 1"），from-scratch 崩溃 AE 2020 故走 embed（待办 2）。本文件保留 RE 过程供 Stroke 子项参考。

## 背景

V2.2.1 子项② = Path embed bytes，原以为是 Ellipse 同款 mirror（embed body + overwrite）。RE 后发现实质更重。

## 已做 RE

- AE-native 参考 fixture：`test_data/v2_2_shape_path_tolerance.aep`（`tmp_debug/gen_shape_path_tolerance.jsx`，AE 2025 存，4 顶点闭合方块 [[0,0],[100,0],[100,100],[0,100]] + Fill）。
- 我们 from-scratch 输出：`go run ./tmp_debug/gen_ellipse_input path` → `tmp_debug/ellipse_input_path.aep`。
- 结构对比工具：`tmp_debug/dump_chunks`（树）+ `tmp_debug/diff_path_geom`（几何字节 diff）。

### 发现 1：scaffolding 结构一致，仅 tdsn 显示名大小不同
Path body 树（`ADBE Vector Shape - Group` → tdsb/tdsn → tdmn `ADBE Vector Shape` → LIST(om-s)(om-s header: tdsb/tdsn/tdb4 124B/cdat 4B) → LIST(omks) → LIST(shap)(shph 24B + LIST(list)(lhd3 52B + ldat 96B) + omtn 0B))与 AE-native **逐 chunk 一致**，唯 tdsn（display name）大小不同（AE 14/16B vs 我们 8/12B）——与 Ellipse 当年同类差异，embed 可解。

### 发现 2（关键）：`encodeBezier` 的 ldat 顶点编码与 AE 不符
`tmp_debug/diff_path_geom` 结果：
- **shph (24B): 一致** ✓
- **lhd3 (52B): 一致** ✓
- **ldat (96B): 不同** ✗（同长度，24 个 f32，第 4/16/23 个 f32 位置不同）

ldat = 24 f32（4 顶点 × 6 f32，疑似 vertex.xy / inTan.xy / outTan.xy，且坐标**归一化到 0/1**，实际尺度在 shph 的 0x42c80000=100）。AE 与我们的 0/1 值**集合相同但排布不同** → 顶点/切线的 ordering 或 interleave 规则我们搞错了。

AE  f32[0..23]: 0,0,0,0, 1,0,1,0, 1,0,1,0, 1,1,1,1, 0,1,0,1, 0,1,0,0
OURS f32[0..23]: 0,0,0,0, 0,0,1,0, 1,0,1,0, 1,1,1,1, 1,1,0,1, 0,1,0,1
diff @ f32 idx 4 / 16 / 23。

`encodeBezier` 在 `internal/aep/lower_property_stream.go`（`LowerPathStream` 调用）。该编码从未过 AE ship-gate（V2.2 path 层一直 skip），故 bug 潜伏。

## 待办（子项② 真实范围）
1. **RE ldat 编码**：解出 AE 的顶点/切线/归一化/ordering 精确布局（对照更多 fixture：不同顶点数、非零切线、开放 vs 闭合、非方形），修 `encodeBezier`。**核心难点、纯字节 RE、不确定性高。**
2. **embed scaffolding**：提取 AE-native path body 作 `templates/v2_2_shape_path_body.bin`，clone + 重生几何（shph/lhd3/ldat by 修好的 encodeBezier）注入，防 silent-drop。变长几何 → 不是 overwrite-in-place，是 splice。
3. **聚焦双版本 ship-gate**：单 Path 层，re-save + 解析器读回 ldat 比对（沿用 Ellipse 范式；注意 ExtendScript shape `.value` 除零坑）。

## 复用资产
- ldta 已 target-conditional（160/164），shape 层地基已修，子项② 不再撞 AE 2020 corrupt。
- Ellipse 流程范本（fixture→extract→embed+inject→聚焦 gate）。
- AE flake 应对：先 `clear_ae_crashstate.ps1` 清崩溃态 + warm retry；对话框遮挡用 PrintWindow（`capture_dialog.ps1`）。
