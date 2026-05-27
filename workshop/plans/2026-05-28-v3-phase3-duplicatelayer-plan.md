# V3 Phase 3 — Composition.DuplicateLayer Implementation Plan

> **For agentic workers:** mirror Phase 2 workflow — `plans/finish/2026-05-28-v3-phase2-deletelayer-plan.md` is the reference template. Tasks reuse Phase 2's tooling (ae_run.ps1 self-driven, dump_layers, ship-gate verifier pattern).

**Spec**: extends [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 Phase 3 candidate. No separate brainstorm spec — too small to need one; strategy doc (Task 2) substitutes.
**Predecessor**: V3 Phase 2 DeleteLayer (2026-05-28, commits `bc976ff…00dd136`). PASS 267.

**Goal**: `(c *Composition) DuplicateLayer(index int) (*Layer, error)` — second V3 structural mutation. Returns the cloned layer for caller convenience (Phase 2 DeleteLayer returns nothing because the deleted layer is gone). **Must pass AE 2020 + AE 2025 ship-gate** (CLAUDE.md hard constraint #6).

**Architecture reuse**:
- V2.1 atomic invariants (snapshot / mutate / warnings-as-failure / rollback) — same pattern as DeleteLayer
- Adaptive 16-chunk insight: each AE-saved layer occupies `Layr + Ewst + N leaf followers`. DuplicateLayer must clone the entire trailing block, NOT just Layr+Ewst (that would lose per-layer follower state)
- `findLayrIndexInItemList` helper from `delete_layer.go` — reusable

**Non-goals (Phase 3)**:
- Cross-composition duplicate (`InsertLayer(src, atIdx)` — Phase 4+)
- Non-AV layer dup (camera/light/audio) — same refuse pattern as Phase 2
- Dup a layer's source footage / pre-comp (`Project.DuplicateItem` — Phase 4+)
- Dup multiple layers as a batch — Phase 3+ wrapper if demand surfaces

---

## 关键 RE 未知

写代码前必须用 AE-saved fixture 回答（Task 1 解决）：

| Q | 问题 | 影响代码路径 |
|---|---|---|
| RE-Q1 | AE duplicate (layer.duplicate() JSX) — 新 layer 落在哪个 index？immediately below source (idx+1) / top (idx=1) / bottom？ | Item LIST insertion position |
| RE-Q2 | 新 layer 的 ID 怎么分？monotonic next (head counter +1) 还是其他算法？ | Layer ID allocation; head counter bump |
| RE-Q3 | 14 follower chunks (fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2) verbatim 复制还是有字段需要 mutate？ | Follower chunk clone strategy |
| RE-Q4 | source layer 是别人的 parent (e.g. L3.parent=source) — children 的 ParentID 跟着新 dup 走，还是仍指原 source？ | 引用 fixup pass |
| RE-Q5 | source layer 自己有 parent (e.g. source.parent=other) — 新 layer 的 ParentID 复制原值，还是 reset to 0？ | ldta @0x84 复制 vs 清零 |
| RE-Q6 | source layer 是 track matte target (有 TrackMatteLayerID 指别处) — 新 layer 的 matte 关系怎么走？同 Q5 但 @0xA0 | ldta @0xA0 复制 vs 清零 |
| RE-Q7 | source layer 名 = "L2_mid"，新 layer 名 = "L2_mid 2"？AE 自动加 " 2" 后缀？保持原名？ | Layer.Name + Utf8 chunk bytes 处理 |
| RE-Q8 | source 是 shape layer / text layer — embed bytes (btds / tdgp) verbatim clone 还是需要重生成 ID？ | byte-level clone vs lower-then-rebuild |

**RE 不到的话**：Phase 3 保守拒接复杂 source（shape/text）或带 children/matte 的 source layer，逼调用方先 unparent。MVP scope = simple AV solid layer dup。

---

## Pre-flight

```powershell
go vet ./...                                                # clean
go test -count=1 ./internal/aep/... 2>&1 | Select-String '^ok'   # expect ok
git log --oneline -1                                         # expect 00dd136 V3 Phase 2 ship-gate green
```

---

## Task 1: RE harness — JSX + 4 AE-saved fixtures

**Goal**: 用 AE 自己跑 `layer.duplicate()` 产出 4 个 fixture 答 RE-Q1..Q7。Agent 自己跑 ae_run.ps1 (per playbook update)。

**Files (create)**:
- `test_data/re_duplicate_layer.jsx` — 4 mode harness via `$.getenv("RE_DUP_MODE")`:
  - `solo` — 3 layers, dup L2 (middle, no refs)
  - `dup_parent` — L3.parent=L2, dup L2 (L2 is parent of L3)
  - `dup_child` — L1.parent=L2, dup L1 (L1 has outgoing parent ref)
  - `dup_matted` — L1.trackMatteType=ALPHA (implicit src=L2 above? careful with layout), dup L1
- AE 2020 saves: `re_duplicate_layer_{solo,dup_parent,dup_child,dup_matted}.aep`
- (no AE 2025 variant needed for Task 1 — focus on Q1-Q7 which are AE 2020 共有 behavior)

**Steps**:
- [ ] **Step 1.1**: 写 JSX (~120 lines, base on `re_delete_layer.jsx`)
- [ ] **Step 1.2**: Agent 跑 ae_run.ps1 × 4 (AE 2020 17.7x45)
- [ ] **Step 1.3**: `go run ./tmp_debug/dump_layers re_duplicate_layer_*.aep` — diff baseline vs dup'd states
- [ ] **Step 1.4**: 写 `workshop/scars/ae-duplicatelayer-re.md` 落 6-8 Findings (镜像 DeleteLayer scar 风格)

---

## Task 2: Strategy spec

**File**: `specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md`

Decisions table:
- Insertion index (Q1)
- ID allocation (Q2)
- Follower chunk clone (Q3) — verbatim byte copy vs regenerate
- ParentID inheritance from source (Q5) — copy vs zero
- TrackMatteLayerID inheritance (Q6) — copy vs zero
- Name suffix policy (Q7) — auto " 2" suffix vs require user pass new name
- Refuse-cases (mirror Phase 2 + shape/text if Q8 reveals embed bytes are non-trivial)

Signature decision:
```go
func (c *Composition) DuplicateLayer(index int) (*Layer, error)
```
or with explicit name param:
```go
func (c *Composition) DuplicateLayer(index int, name string) (*Layer, error)
```

Latter decoupled from AE's auto-suffix; cleaner. Pick after Q7 RE.

---

## Task 3: Implementation

**Files**:
- Create: `internal/aep/duplicate_layer.go`
- Modify: `internal/aep/delete_layer.go` — extract `findLayrIndexInItemList` to a shared file (e.g. `layer_itemlist.go`) if used by both
- Tests: `internal/aep/duplicate_layer_test.go` — refuse-cases + happy path round-trip + cloned-state verification (cloned layer's ParentID/TrackMatteLayerID match Task 2 decisions; cloned bytes don't share Data slice with source — concurrent-mutate safety per `scars/concurrency-unsafe-shared-chunk-bytes.md`)

Atomic invariants (Inv-10/-11):
- Snapshot itemList children + c.Layers + proj.Warnings + proj.nextItemID
- On warning rollback ALL including nextItemID bump (otherwise next NewShapeLayer/Duplicate gets a wrong ID)

---

## Task 4: Structural equivalence + round-trip

- [ ] **Step 4.1**: `TestDuplicateLayer_RoundTrip` — Open → Dup(1) → WriteAEP → Reopen → 4 layers w/ expected IDs
- [ ] **Step 4.2**: `TestDuplicateLayer_StructuralEquivalence_Solo` — Go's Open(baseline)→Dup(1) ≡ AE's solo.aep itemList.Children (chunk-shape compare)
- [ ] **Step 4.3**: PASS count: 267 + N (N ≥ 5)；FAIL=0；vet clean

---

## Task 5: Ship-gate (agent-side)

- [ ] **Step 5.1**: `tmp_debug/ge_duplicate_layer/main.go` produces ge_*.aep for 4 modes
- [ ] **Step 5.2**: AE 2025 × 4 modes via `verify_ge_duplicate_layer.jsx`
- [ ] **Step 5.3**: AE 2020 × 4 modes (or 3 if matte mode needs AE 25 input)
- [ ] **Step 5.4**: 全 PASS → Task 6; 任 FAIL → scar update

完成判定: 4 × 2 = 8 次开盘全 PASS (或文档化 mode/version 例外 as DeleteLayer matte_ae20/ae25 split)。

---

## Task 6: Docs sync + Stable promotion

- [ ] godoc: alpha → Stable
- [ ] board.md: V3 Phase 3 complete; Next session = Phase 4 候选
- [ ] coverage.md: + DuplicateLayer row
- [ ] Move plan + spec to finish/
- [ ] Commit: `feat(aep): V3 Phase 3 ship-gate green — DuplicateLayer Stable`

---

## Acceptance

1. `go vet ./...` clean + tests PASS 267 + N, FAIL=0
2. AE 2020 + AE 2025 各 4 mode PASS (per ship-gate split policy from Phase 2)
3. Public API: `Composition.DuplicateLayer(int) (*Layer, error)` godoc complete
4. board.md + Stable + finish/ migration

---

## Risk register

| Risk | Mitigation |
|---|---|
| Follower chunks (14 leaves) carry per-layer state we don't understand (e.g. timeline cache / opacity baseline) — verbatim clone breaks AE | Phase 2 RE only proved followers can be DELETED safely; clone safety needs separate RE. Stage 4 bisection if AE rejects |
| Layer ID collision — if AE allocates ID differently than head-counter monotonic, cloned ID collides with existing layer | Phase 2 RE confirmed head counter monotonic-no-reuse. Use `proj.nextItemID` and bump (same as NewShapeLayer line 139-141) |
| Shape/Text layer embed bytes (btds, tdgp ID refs) share state — naive byte-clone produces 2 layers pointing at same embed | Refuse non-AV in Phase 3; lift after embed-byte RE in Phase 4+ |
| Name collision — AE auto-suffixes " 2"; Go users may pass duplicate names | Take explicit `name string` parameter, validate non-empty + non-collision. Reject silent duplicate-name like NewShapeLayer does |
| Dup'd layer's ParentID still points to deleted layer in subsequent DeleteLayer call | Out of scope — callers should use SetParent to clean up before delete |

---

## Related

- Spec: [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 Phase 3 candidate
- Phase 2 reference: [`finish/2026-05-28-v3-phase2-deletelayer-plan.md`](finish/2026-05-28-v3-phase2-deletelayer-plan.md) + [`../specs/finish/2026-05-28-v3-phase2-deletelayer-strategy.md`](../specs/finish/2026-05-28-v3-phase2-deletelayer-strategy.md)
- DeleteLayer scar: [`../scars/ae-deletelayer-re.md`](../scars/ae-deletelayer-re.md) — F4 (16-chunk per-layer block) is the foundation for "what to clone"
- AE acceptance gate: [`../scars/ae25-acceptance-gate.md`](../scars/ae25-acceptance-gate.md)
- Playbook: [`../playbooks/re-fixture.md`](../playbooks/re-fixture.md) (agent runs ae_run.ps1 directly)
