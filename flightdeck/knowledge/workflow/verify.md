# 验证流程 — checklist

SUMMARY: 验证流程
READ WHEN: preparing to commit; verifying tests + vet pass; reconciling PASS count against cockpit.md; writing a new test (conventions for fixture / corruption / AE 24 fields); needing a debug tool to inspect chunks / layers / properties; deciding where temporary outputs belong

---

## 提交前自查

```bash
go vet ./...       # 必须干净（pre-existing modernize 警告除外）
go test ./...      # 必须全绿
```

跑出 FAIL → **不要** `--no-verify` 跳过 hooks。**不要** amend 已 commit。先查根因。

## 当前基线

PASS count 见 `../cockpit.md` `Last updated` 行（**唯一权威**）。

校对命令：

```bash
go test -count=1 ./internal/aep/... -v | grep -c '^--- PASS'    # 应等于 cockpit.md 写的数字
go test -count=1 ./internal/aep/... -v | grep -c '^--- FAIL'    # 应 = 0
go vet ./...                                                       # 应 clean
```

如果 PASS 数对不上 cockpit.md，先查最近改动是否漏了同步。

## 全量 AE gate 回归 / fixture 完整性

```powershell
pwsh -File scripts/run_ship_gates.ps1                  # 全量 ship-gate sweep + 台账 (test_data/gate_ledger.json)
pwsh -File scripts/run_ship_gates.ps1 -Run 'SolidNull' # 只跑子集
pwsh -File scripts/regen_fixtures.ps1 -CheckOnly       # fixture 盘点（缺失 + 无主清单）
pwsh -File scripts/regen_fixtures.ps1                  # 只补缺失 fixture（不碰已有，防 byte-diff 基线漂移）
```

跨切面改动（ID 分配 / write 路径 / ae_run wrapper / 规则表）→ 跑全量 sweep；新增生成 JSX → 同步 `scripts/fixtures_manifest.json`。详 `re-fixture.md` § 全量 gate sweep。

## 跑单个测试

```bash
go test ./internal/aep/ -run 'TestSetManualKerning' -v
go test ./internal/aep/ -run 'TestMaterialOptionsTypedSettersRoundtrip' -v
```

跑 real-file fixture 测试时，对应 `test_data/re_*.aep` 缺失 → `t.Skipf(...)`，不阻塞 CI。

## 用 -aep 标志跑外部文件

```bash
go test ./internal/aep -run TestManualFile -aep "C:/path/to/your.aep" -v
```

会 dump 整个 Project 结构供肉眼核对。

## 调试工具 `tools/debug/`（tracked，入 git）

> **2026-06-18 整顿**：可复用调试工具从 gitignored `tmp_debug/` **提进 tracked `tools/debug/`**（在 git、跨 clone、不丢）。一次性 RE scaffolding(findings 已进 incidents)+ 渲染输出全清。**约定见下「工件该放哪」。**

`tools/debug/` 下小程序按需用：

| 工具 | 用法 | 用途 |
| --- | --- | --- |
| `parse_btdk` | `go run ./tools/debug/parse_btdk <file.aep> <layer_filter>` | dump 文本图层 btdk PostScript 树 |
| `dump_kern` | `go run ./tools/debug/dump_kern <file.aep> <layer>` | 打 ManualKerning / Kerning |
| `dump_cdta` | `go run ./tools/debug/dump_cdta <file.aep> [name_prefix]` | dump comp cdta hex（RE cdta 字节用） |
| `dump_ldta_trackmatte` | `go run ./tools/debug/dump_ldta_trackmatte <file.aep> [comp] [prefix]` | dump layer ldta hex |
| `list_items` | `go run ./tools/debug/list_items <file.aep>` | 列所有 items + id |
| `list_props` | `go run ./tools/debug/list_props <file.aep> [comp]` | 列每个 layer 的所有 Property |
| `list_item_chunks` | `go run ./tools/debug/list_item_chunks <file.aep> [prefix]` | dump comp Item LIST 完整 chunk 树（找 PRin / prda 这种 sibling chunk） |
| `probe_effects` | `go run ./tools/debug/probe_effects <file.aep>` | 列每个 layer 的 effects + params |
| `dump_layers` | `go run ./tools/debug/dump_layers <file.aep> [<file2> ...]` | 每 comp parsed layers (ID/Type/Name/ParentID/TrackMatteLayerID/SourceID) + raw Item LIST 子 chunk 顺序 + Layr/Ewst pairing 检测 — DeleteLayer / InsertLayer 结构性 mutation 的 fixture diff 主力 |
| `parade_dump` / `dump_parade` | `go run ./tools/debug/parade_dump <file.aep>` | 每 Layr 的 outer tdgp 子 chunk 顺序（tdmn 名）+ Effect Parade 体内 chunk hex（parade 位置/头字节 RE 用） |
| `effect_id_scan` | `go run ./tools/debug/effect_id_scan <file.aep\|.bin> <decimal>` | 全树扫 32-bit BE 值命中（chunk 路径 + 偏移）+ 列 Layr ldta 层 ID — 追「无法找到图层 ID=N」类悬空引用（tdpi 等） |
| `dump_tdmn` / `dump_chunks` / `dump_root` / `dump_idta` / `dump_fdta` / `dump_comp` / `dump_text` / `count_layers` | 类似 | 全树 matchName / chunk 树 / 各类字节 RE 核对 |

另：`go run ./cmd/aepdissect <file.aep>` = 结构化解析报告(效果用量+原生/Cycore/第三方+预合成嵌套+逐层效果链),模版分析首选。

### 工件该放哪（避免再堆 junk drawer）

| 类型 | 家 | 在 git? |
| --- | --- | --- |
| 可复用调试工具 | `tools/debug/<name>/` | ✅ tracked |
| 用户面工具 | `cmd/<name>/` | ✅ tracked |
| vfx / showcase 生成器 | `flightdeck/showcase/<方向>/gen.go` | ✅ tracked |
| AE / 渲染 / gate 输出 | `tmp_debug/` | ❌ gitignored,**不放 `.go`, 用完即删** |
| 一次性 Go probe | `_tmp_debug/<name>/` | ❌ gitignored,`go test ./...` 跳过,用完即删 |
| 下载缓存 / 普通杂输出 | `tmp/` | ❌ gitignored,用完即删 |

原则:**有用→进 git(提到上面某个 tracked 家);没用→删;只 RE 一次但会改变未来行动的 findings 进 `flightdeck/knowledge/<domain>/` 后删探针。** 不放进 `internal/aep`,避免污染 public API。

## 重构脚本

`scripts/` 下：

- `split_tests.py` — 按 test name 映射拆 `aep_test.go`（Phase 1）
- `split_parse_text.py` — 按行号 slice 拆 `parse_text.go`（Phase 2）
- `split_write.go` — 按 func name + append 模式拆 `write.go`（Phase 3）

教训：纯机械批量移动写脚本（比手动 ~10× 省 token）。import 检测 regex 用 `\bpkg\.\w` 避免注释里 "ldat bytes. Used" 误中。

## 测试惯例

- **合成 RIFX fixture 优于真实文件**。`rifxBuilder` + `build*` helpers 让 corruption 类测试 reproducible。新增解析路径时优先合成 fixture。
- **真实文件走 `TestManualFile`**：`go test -aep "C:/path/to/your.aep"`，dump 整个 Project 结构供肉眼核对。不入 CI baseline。
- **每个解析改动都应该有对应 fixture 测试 + roundtrip 测试**。`TestKeyframeRoundtrip` 是好模板。
- **警告路径用 `buildCorruptKeyframedLeaf` 注入异常 lhd3**。不要改 `leafKeyframed` 的健康 helper，**分开 corruption 与正常 case**。
- **AE 24 字段 fixture 走 `re_*_ae24.jsx`**。详 `re-fixture.md` § RE fixture 双轨。
- **`*_test.go` 全部 `package aep_test`**（公开测试，强制走 exported API）。
- 新加 fixture 用 `t.Skipf` 缺文件跳过，不阻塞 CI。
