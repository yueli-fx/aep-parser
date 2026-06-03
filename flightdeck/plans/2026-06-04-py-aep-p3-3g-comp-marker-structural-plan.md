---
status: active
implements: specs/2026-05-26-py-aep-parity-design.md
summary: P3 §3G comp marker 增删（结构性：ldat block splice + lhd3 count + Nmrd splice，clone-template 规避 opaque RE，双版本 ship-gate）
---

# P3 §3G — Composition Marker 增删（结构性）

**Spec**: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md) §3 Phase 3 §3G
**前置已落**: Marker R/W setter 全套（`SetTime/Duration/Label/Comment/...`，见 `write_marker.go`）；comp marker 解码（`findCompMarkers` → `parseMarkers`，见 `parse_composition.go:240`）。
**性质**: length-variable 结构性写路径 → 走 V2.1 atomic invariants（snapshot → mutate → warnings-as-failure rollback）+ **AE 2020 + 2025 双版本 ship-gate**（CLAUDE.md 硬约束 #6）。落地前 = Alpha。

## 1. 现状与缺口

comp marker 存储链（已解码，`parse_marker.go` 头注释）：

```
SecL("Markers" 伪层) → Transform Group → tdmn"ADBE Marker" + mrst LIST
  mrst
  ├── tdbs LIST → kfl LIST → { lhd3 (count@0x08, bpk@0x10=16), ldat (count×16) }
  │     per-block: 0x00-0x03 time(1/tickRate); 0x04-0x0F opaque metadata
  └── mrky LIST → Nmrd LIST ×count { NmHd(20B) + 5×Utf8(comment/chapter/url/frame/cue) }
```

**缺口**：`Marker` 只持有 `ldat / ldatOffset / nmrd / nmHd`；增删还需 `kfl 的 lhd3`（改 count）、`ldat`（增删 16B block，已有）、`mrky`（增删 Nmrd）、以及"同 list 其余 marker"（删除后其 `ldatOffset` 需下移）。Composition 不持 mrst/kfl/mrky 引用。

## 2. API 面（Go 惯例，不镜像 py-aep snake_case）

```go
// 增：在 comp marker 列表尾部按 time 排序插入，返回新 Marker（可继续 Set*）
func (c *Composition) AddMarker(seconds float64) (*Marker, error)
// 删：从所属 comp marker 列表移除该 marker
func (m *Marker) Remove() error
```

- `AddMarker` 返回 `*Marker` 与 3C `Duplicate()` 返回 clone 节点一致，便于链式 `Set*`。
- `Remove()` 挂在 `*Marker`（接收者即被删对象），与 `AEPropertyGroup.Remove()` 风格一致。
- **本 slice 只做 comp-level**；layer marker 增删（2H 延伸）不在范围，但 back-ref 捕获设计要能复用（layer/comp marker 共享 `Marker` 类型与 writer）。

## 3. 实现机制

### 3.1 back-ref 捕获（parser 改动，前置）

`parseMarkers` 已有 mrst/tdbs/kfl/lhd3/ldat/mrky/nmrd 全部局部变量。新增一个**列表持有者**让每个 Marker 能回到容器：

- 方案 A（轻）：给 `Marker` 加 `markerList *markerList` 反指；`markerList{ mrst, lhd3, ldat, mrky *rifx.Chunk; markers []*Marker }`。Composition 持 `*markerList`（nil = 无 marker 伪层）。
- `findCompMarkers` 返回 `*markerList`（含 `.markers`），`Composition.Markers` 由其投影 / 或直接换成 method。**倾向保留 `Markers []*Marker` 字段不破坏 stable API**，markerList 作内部 sidecar（map comp→markerList 或 Composition 私有字段）。

> 决策点：是否动 `Composition.Markers` 的公开形状。倾向**不动**（stable），sidecar 私有字段挂 back-ref。

### 3.2 Remove（先做 — 零 RE）

纯 splice，无字节 RE：
1. 定位 `m` 在 markerList 的 index `i`。
2. ldat：删除 `[i*16, (i+1)*16)` 这 16B；其后所有 marker 的 `ldatOffset -= 16`。
3. lhd3：count `@0x08 -= 1`（length-preserving 字段写）。
4. mrky：从 Children 移除 `m.nmrd`。
5. scene：从 `markerList.markers` / `Composition.Markers` 摘除 `m`。
6. WriteAEP 重算父 LIST size（ldat/mrky/kfl/tdbs/mrst/SecL 链）—— 复用既有 length-variable 重算路径（验证 mrst 链在重算白名单内，否则补）。

### 3.3 Add（clone-template 规避 opaque RE）

opaque 12B（ldat 0x04-0x0F）+ NmHd 20B 的 canonical 默认值未 RE。**规避**：comp 已有 ≥1 marker 时，clone 末个 marker 的 ldat block + NmHd + 空 5×Utf8（opaque preservation，CLAUDE.md #5，拷真实 AE 字节），再 `SetTime`。
1. ldat：append 16B（拷模板 block，仅改 0x00-0x03 time）。
2. lhd3：count `+= 1`。
3. mrky：append 新 Nmrd（NmHd 拷模板 + 5 空 Utf8，与 `setNmrdUtf8` 一致）。
4. scene：append 新 Marker（back-ref 指向新 block/nmrd/nmHd）。
5. **排序**：AE comp marker 是否要求 ldat 按 time 升序？待 RE 确认（§4）。先实现"尾插"，排序约束按 ship-gate 反馈补。

### 3.4 Add-to-empty（延后 slice）

comp 无 marker 伪层时新建整条 SecL→...→mrst 链 = canonical seed RE，单独 slice，依赖 AE fixture。本 plan 先不做（`AddMarker` 在无伪层时返回明确 error）。

## 4. RE 未知 / fixture 计划

- **opaque 12B / NmHd 默认值**：clone-template 规避，无需 RE（除 add-to-empty）。
- **ldat time 排序约束**：AE 是否依赖 marker 按 time 升序？→ fixture：comp 加 3 个乱序 time marker，AE 存盘看 ldat 顺序。
- **ship-gate fixture**：① comp 有 N marker → 我们 AddMarker → AE 2020+2025 接受且显示 N+1、time/comment 正确；② Remove 中间一个 → 接受且剩余正确。走 `checklists/re-fixture.md` § GDI 自动化 + `scripts/ae_run.ps1` 无人值守。

## 5. 测试计划（TDD）

1. `marker_structural_test.go`（新）：
   - Remove：parse re_tickrate.aep（2 comp marker）→ `markers[0].Remove()` → WriteAEP → re-parse → 断言 1 marker、留存者 time/comment 正确、未触区 byte-identical。
   - Add：parse → `AddMarker(t)` → Set* → round-trip → 断言 N+1、新 marker 字段正确。
   - 空/边界：单 marker 全删→0；Remove 已脱离 list 的 marker 报错；AddMarker 无伪层报错。
2. 原子性：注入 warning 路径验证 rollback（参照 `mutate_property_structural.go` 测试）。
3. `go vet ./... && go test ./...` 全绿 + PASS count 对账（`checklists/verify.md`）。

## 6. 落地切片顺序

1. **back-ref 捕获**（parser sidecar）+ Remove + 测试 → commit（Alpha，标 BREAKING 若动签名）。
2. **Add（clone-template）** + 测试 → commit（Alpha）。
3. **AE 双版本 ship-gate**（Remove + Add）→ 过则升 stable，更新 coverage.md / spec §3G / cockpit。
4. （可选）time 排序约束 RE + add-to-empty seed slice。

## 7. 风险

- mrst 链可能不在 WriteAEP 的 length-variable 父 size 重算白名单 → 需补（硬约束 #1）。先验证 re_tickrate.aep 改 marker comment 长度后 round-trip 是否已正确重算（SetComment 已是 length-variable，应已覆盖 mrst 链 → 低风险）。
- bisection-over-stacking（incidents 教训）：ship-gate reject 时按最小失败 bisect，勿堆结构猜测。
