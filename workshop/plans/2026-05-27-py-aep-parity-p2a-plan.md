# py-aep parity — Phase 2a implementation plan

**Spec**: [`../specs/2026-05-26-py-aep-parity-design.md`](../specs/2026-05-26-py-aep-parity-design.md) § 2.1-2.5
**Predecessor**: [P1 plan](finish/2026-05-26-py-aep-parity-p1-plan.md) (archived 2026-05-26)
**Goal**: ship P2 第一批低悬果 — 5 个 task / ~10-15 新 API，全 length-preserving 或现有 chunk slot 复用，零 V3-blocking 依赖。

**Out of scope for P2a**: pard chunk reader / specs.py schema table 端口 (推 P2b)、nnhd 字节布局 RE (P2b)、CMS JSON (P2b)、Gradient XML (P2b)、PropertyGroup 链式访问 (P2c)、TimeRemap enable (P2d, 等 V3 capability framework)。

**完工标准**: PASS count +10 以上、`go vet ./...` clean、`docs/{layer,property,project}.md` + `coverage.md` + `coverage-detail.md` + `board.md` 同步、2E ImportPlaceholder 跑 AE 2020+2025 ship gate (NewComposition pattern)。

**Pre-flight check**:
```powershell
go vet ./...                                                      # clean
go test -count=1 ./internal/aep/... -v 2>&1 | grep -c '^--- PASS' # 202 (P1 baseline)
git status --short                                                # clean
```

---

## Task 1 — 2J ThreeDModelLayer R only（最简，先 prove pattern）

**目标**: 给 Layer 加 3D Model 类型判别 + `IsThreeDModelLayer()` typed accessor。

**实现位置**: `internal/aep/layer_convenience.go`（已有，append）

**API surface**:
```go
func (l *Layer) IsThreeDModelLayer() bool
```

**RE prereq**: ldta subtype enum @0x83 已有 parse；查 py-aep `LayerType.THREE_D_MODEL` 数值，加 const 到 `parse_layer.go`。

**步骤**:
1. py-aep `parsers/layer.py` 找 `LayerType.THREE_D_MODEL` 数值
2. 在 `parse_layer.go` ldta subtype enum 加常量
3. `layer_convenience.go` 加 `IsThreeDModelLayer()` 返回 `l.subtype == ThreeDModelLayerSubtype`
4. 写 fixture-based test（如缺 3D Model fixture 用 `t.Skipf`）
5. 同步 `docs/layer.md` + `coverage.md`

**估算**: 30 min

---

## Task 2 — 2I LightSource R/W (AE 24+，无新 chunk)

**目标**: Light layer 的 `LightSource` accessor — Environment-type light 指向另一个 layer 作光源。

**实现位置**: `internal/aep/layer_accessors.go`（已有，append 到 Light 段）

**API surface**:
```go
func (l *Layer) LightSource() *Layer            // nil if none / non-light layer
func (l *Layer) SetLightSource(target *Layer) error
```

**实现要点**（参考 py-aep `light_layer.py:41-79`）:
- **复用既有 ldta `source_id` 字段**（不是新 chunk）— light layer 的 source_id 在 AE 24+ 语义改为 light source ref
- Read: 读 ldta source_id；用 `Composition.LayerByID()`（P1 1B 已 ship）查 layer
- Write 验证：target 必须 in same comp、AVLayer、`!Is3D`（py-aep 同等约束）；source_id == 0 → 清除关联
- `LightSource()` 对非-light layer 返回 nil（不报错）

**步骤**:
1. parse_layer.go 暴露 source_id getter 给 light 用（如未暴露）
2. layer_accessors.go 加 R/W pair
3. write_layer.go 加 source_id setter（ldta length-preserving 单字段改）
4. test: 用 `re_lights_ae24.jsx` 或合成 fixture 测 R/W roundtrip + validation 错误路径
5. 同步 `docs/layer.md` § Light + `coverage.md`

**估算**: 1 hr

---

## Task 3 — 2D 部分: LockedRatio R/W (tdsb bit)

**目标**: Property tdsb chunk 的 `LockedRatio` bit reader/writer。

**实现位置**: `internal/aep/property_flags.go`（已有，append）

**API surface**:
```go
func (p *Property) LockedRatio() bool
func (p *Property) SetLockedRatio(v bool) error
```

**RE prereq**: tdsb 是 4 字节 chunk。py-aep `binary/property_chunks.py:30-52`:
- byte 0: `roto_bezier` (bit 0)
- byte 1: pad
- byte 2 (`_lock_flags`): `locked_ratio` = bit 4
- byte 3 (`_enable_flags`): `dimensions_separated` = bit 1, `enabled` = bit 0

我们 parser 目前是否解 tdsb？grep `tdsb` 看现状；若 parser 不解，parse_properties.go 加 tdsb reader。

**步骤**:
1. grep `tdsb` 查现状（rifx 是否声明常量、parser 是否提取 tdsb chunk ref）
2. 若无：rifx.go 加 IDTdsb 常量；parse_properties.go 给 Property 加 `tdsb` 私有 ref（类似 P1 1G 的 `tdb4` 注入）
3. property_flags.go: `LockedRatio()` 读 byte 2 bit 4
4. write 路径: bit set/clear, length-preserving
5. test: roundtrip + 缺 tdsb fallback（standalone Property 返 false）
6. 同步 `docs/property.md` + `coverage.md`

**估算**: 1.5 hr

**注**: `DimensionsSeparated` (byte 3 bit 1) **read 在此 batch 加**（参考 LockedRatio 一行 cost），write 是 structural toggle（不在 P2a 范围，推 P3）。

---

## Task 4 — 2F ReplaceSource（Layer 级 splice）

**目标**: 公开 `Layer.ReplaceSource(target Item, fixExpressions bool)` 镜像 py-aep API；底层走既有 SetSource 路径。

**实现位置**: `internal/aep/layer_accessors.go`（append）

**API surface**:
```go
func (l *Layer) ReplaceSource(target Item, fixExpressions bool) error
```

**实现要点**:
- `target Item` 接口（Composition / Footage 共用），取 ID → splice ldta source_id
- `fixExpressions` 参数我们**接受但不实现**（symbolic execution 不在 scope）；记 warning to Project.Warnings 如果 fixExpressions=true
- 内部直接调既有 `Layer.SetSource` (已 ship per coverage.md)

**步骤**:
1. layer_accessors.go: 加 wrapper method
2. test: 创建 2 个 Footage，对一个 Layer 调 ReplaceSource，验 source_id 变 + reopen 后跟过
3. test: fixExpressions=true 路径产生 warning
4. 同步 `docs/layer.md` + `coverage.md`

**估算**: 30 min

---

## Task 5 — 2E ImportPlaceholder（NewComposition pattern，atomic）— 🗑️ 暂搁 2026-05-27

**状态**: deferred — 合成 opti tag AE 拒收，删 builder + tests。重起需要真 AE-saved placeholder fixture 把 opti / sspc / idta 字段 RE 出来。

**目标**: `Project.ImportPlaceholder(name, width, height, frameRate, duration)` 在 root folder 新建 placeholder Footage item。

**实现位置**: `internal/aep/new_project.go` 或新建 `internal/aep/import_placeholder.go`（与 new_composition.go 同模式）

**API surface**:
```go
func (p *Project) ImportPlaceholder(name string, width, height int, frameRate, duration float64) (*Footage, error)
```

**实现要点**（参考 py-aep `models/project.py:477-519`）:
- name 规范化: `""` → `"Placeholder"`, special-case
- 验证 width/height [4, 30000], frameRate [1.0, 99.0], duration > 0 且 ≤ 10800
- Item Fold-level siblings：Fold + cdta(footage variant) + sspc + Pin + opti 等（参考 parse_footage.go 知道 placeholder 走 opti `Plac` tag — P1 1H 已 ship discriminator）
- canonical seed：抽 AE 自存的 placeholder fixture bytes（如 `test_data/re_placeholder_ae24.aep` — 缺则提示用户生成）→ embed bytes 路线（类似 V2.2 iter-8 法）
- 原子 (warning/error rollback)

**RE prereq**: 需 AE-saved placeholder fixture; 若缺，本 task 标 deferred 等 fixture。

**ship-gate**: 跑 V2.1 ship-gate 同等 path — 用 `runAeRunShipGate` helper（已 ship）调 AE 2020+2025 双开验。

**步骤**:
1. 检查 `test_data/` 是否已有 placeholder fixture；缺则先 RE 提示用户操作 AE 生成（File > Import > Placeholder + Save）
2. 抽 placeholder Item Fold-level siblings 字节（tmp_debug 工具复用 V2.2 extract_shape_bodies pattern）
3. import_placeholder.go 实现 + canonical seed embed
4. roundtrip test (Go-only)
5. ship-gate test (AE 2020+2025)
6. 同步 `docs/footage.md` + `coverage.md` + V2 segment

**估算**: 半天（如 fixture 在手）；fixture 缺 → 1 天 + 等 fixture

---

## 执行顺序

1 → 2 → 3 → 4 → 5。Task 5 是唯一需要 ship-gate 的，放最后；前 4 个全 length-preserving / 无新 chunk，roundtrip Go test 够。

每个 task 完工 commit `feat(<scope>): py-aep parity P2a Task N — <desc>`；最后 docs/board 闭环 commit `docs(workshop): py-aep parity P2a 全闭环 / archive plan`。

---

## Self-review checklist

| Spec section | Plan coverage |
|---|---|
| 2.1 Project — `ImportPlaceholder` | Task 5 |
| 2.3 Layer — `ReplaceSource` | Task 4 |
| 2.3 Layer — `LightSource` (AE 24+) | Task 2 |
| 2.3 Layer — `ThreeDModelLayer` R | Task 1 |
| 2.4 Property — `LockedRatio` (tdsb) | Task 3 |
| 2.4 Property — `DimensionsSeparated` R | Task 3（W 推 P3 structural） |

**Deferred to P2b / 后续**：
- 2D Property metadata 剩余: `MinValue / MaxValue / UnitsText / NbOptions / PropertyControlType / PropertyValueType / DefaultValue / LastValue` — 需 pard chunk reader + specs.py 2353 行 schema table 端口
- 2A nnhd 字节布局 RE
- 2B CMS JSON
- 2C Gradient XML
- 2G PropertyGroup 链式访问（推 P2c）
- 2K TimeRemap enable（推 P2d，等 V3 capability framework）
