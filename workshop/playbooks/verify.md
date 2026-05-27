---
when_to_read: preparing to commit; verifying tests + vet pass; reconciling PASS count against board.md; writing a new test (conventions for fixture / corruption / AE 24 fields)
applies_to: [verify, test, vet, pass-count, pre-commit, ship-gate, test-conventions, fixture, t-skipf]
last_updated: 2026-05-27
---

# 验证流程

## 提交前自查

```bash
go vet ./...       # 必须干净（pre-existing modernize 警告除外）
go test ./...      # 必须全绿
```

跑出 FAIL → **不要** `--no-verify` 跳过 hooks。**不要** amend 已 commit。先查根因。

## 当前基线

PASS count 见 `../board.md` `Last updated` 行（**唯一权威**）。

校对命令：

```bash
go test -count=1 ./internal/aep/... -v | grep -c '^--- PASS'    # 应等于 board.md 写的数字
go test -count=1 ./internal/aep/... -v | grep -c '^--- FAIL'    # 应 = 0
go vet ./...                                                       # 应 clean
```

如果 PASS 数对不上 board.md，先查最近改动是否漏了同步。

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

## tmp_debug 工具

`tmp_debug/` 下小程序按需用：

| 工具 | 用法 | 用途 |
| --- | --- | --- |
| `parse_btdk` | `go run ./tmp_debug/parse_btdk <file.aep> <layer_filter>` | dump 文本图层 btdk PostScript 树 |
| `dump_kern` | `go run ./tmp_debug/dump_kern <file.aep> <layer>` | 打 ManualKerning / Kerning |
| `dump_cdta` | `go run ./tmp_debug/dump_cdta <file.aep> [name_prefix]` | dump comp cdta hex（RE cdta 字节用） |
| `dump_ldta_trackmatte` | `go run ./tmp_debug/dump_ldta_trackmatte <file.aep> [comp] [prefix]` | dump layer ldta hex |
| `list_items` | `go run ./tmp_debug/list_items <file.aep>` | 列所有 items + id |
| `list_props` | `go run ./tmp_debug/list_props <file.aep> [comp]` | 列每个 layer 的所有 Property |
| `list_item_chunks` | `go run ./tmp_debug/list_item_chunks <file.aep> [prefix]` | dump comp Item LIST 完整 chunk 树（找 PRin / prda 这种 sibling chunk） |
| `probe_effects` | `go run ./tmp_debug/probe_effects <file.aep>` | 列每个 layer 的 effects + params |
| `dump_comp` / `dump_text` / `demo_*` | 类似 | RE 时核对字节 |

不放进 `internal/aep`，避免污染 public API。

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
