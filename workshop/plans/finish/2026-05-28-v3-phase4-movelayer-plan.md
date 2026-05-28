# V3 Phase 4 — Composition.MoveLayer Implementation Plan

> **For agentic workers:** mirror Phase 2/3 workflow — `plans/finish/2026-05-28-v3-phase3-duplicatelayer-plan.md` is the reference template. **No RE step** for this phase (AE behavior is known; layer order = order of Layr LISTs in itemList.Children; AE's `layer.moveAfter()/moveBefore()/moveToBeginning()/moveToEnd()` all reduce to "reorder Layr blocks"). Single plan doc combines strategy + tasks since the spec surface is small.

**Spec**: extends [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 Phase 4 candidate.
**Predecessor**: V3 Phase 3 DuplicateLayer (2026-05-28, commit `b2d3e18`). PASS 278.

**Goal**: `(c *Composition) MoveLayer(from, to int) error` — third V3 structural mutation. Pure reorder: no ID alloc, no new chunks, no follower mutation. Atomic block-splice of source's [Layr + Ewst + leaf followers] from `from`-th slot to `to`-th slot.

**Architecture reuse**:
- V2.1 atomic invariants (snapshot / mutate / rollback) — same pattern as DeleteLayer / DuplicateLayer
- Adaptive 16-chunk block-splice machinery from Phase 2 (`findLayrIndexInItemList` + leaf-follower scan)
- No new RE; no scar entry; no separate strategy spec

**Non-goals (Phase 4)**:
- Cross-composition move (`Composition.InsertLayer(src, atIdx)` — Phase 5+)
- Multi-select batch move (`MoveLayers([]int, int)` — wrapper deferred)
- Reorder via target-relative API (`MoveAfter / MoveBefore / MoveToBeginning / MoveToEnd`) — caller can compute target index from `c.Layers`

---

## 1. Public API signature

```go
// MoveLayer reorders the layer at `from` to position `to` in c.Layers
// (both 0-based). Source's entire chunk block (Layr + Ewst + leaf
// followers, adaptive scan to next LIST/EOF — same machinery as
// DeleteLayer / DuplicateLayer) is spliced out and re-inserted at the
// target slot. After the call, c.Layers[to] == the moved layer.
//
// Refuse-cases (Phase 4 conservative):
//   - from or to out of range
//   - comp lacks parsed itemList back-ref
//   - source layer lacks Layr back-ref / corrupted block (Layr formType
//     / Ewst sibling mismatch)
//
// from == to is a no-op (returns nil, no warnings, no state change).
//
// Atomic mutation (Inv-10 / Inv-11): snapshot pre-call itemList.Children
// + c.Layers + Warnings; no re-parse needed (no new chunks created), so
// the warnings-as-failure path mostly catches defensive errors.
//
// Layer.Index is updated for every layer in c.Layers to match the new
// slice positions (mirrors how parseLayer assigns Index from the parse-
// time slice iteration). Note: DeleteLayer / DuplicateLayer do NOT
// currently update Index on neighbors; this is a deliberate asymmetry —
// MoveLayer's whole purpose is index changes, so leaving them stale
// would be surprising.
func (c *Composition) MoveLayer(from, to int) error
```

**Index convention**: 0-based matches `c.Layers` Go slice (same as DeleteLayer / DuplicateLayer §1).

---

## 2. Algorithm

```
algorithm: moveLayer(c, from, to) → error:
  // 1. Validate refuse-cases. from == to → return nil immediately.

  source := c.Layers[from]

  // 2. Locate source's Layr in itemList.Children and find block extent
  //    [srcLayrIdx, endIdx) — adaptive scan (Layr + Ewst + leaf
  //    followers until next LIST/EOF). Same as DeleteLayer §3.

  // 3. Snapshot: oldChildren = clone(itemList.Children),
  //              oldLayers   = clone(c.Layers),
  //              oldWarnings = len(proj.Warnings),
  //              oldIndexes  = [l.Index for l in c.Layers]   // for rollback

  // 4. Build cut state:
  //      cutChildren = children[:srcLayrIdx] + children[endIdx:]
  //      cutLayers   = c.Layers[:from]       + c.Layers[from+1:]
  //      block       = children[srcLayrIdx:endIdx]   (the moved block)

  // 5. Determine insertion index in cutChildren:
  //    - if to < len(cutLayers):
  //        targetLayrChunk = cutLayers[to].back.layrList
  //        insertIdx = indexOf(cutChildren, targetLayrChunk)
  //        // insert BEFORE the target Layr → result has source at `to`
  //    - else (to == len(cutLayers), move to last slot):
  //        lastLayer = cutLayers[len-1]
  //        lastIdx   = indexOf(cutChildren, lastLayer.back.layrList)
  //        // scan to end of last block (skip Ewst + leaves until next LIST/EOF)
  //        insertIdx = lastIdx + 2
  //        while insertIdx < len(cutChildren) and !cutChildren[insertIdx].IsList():
  //            insertIdx++

  // 6. Splice block back into cutChildren at insertIdx:
  //      newChildren = cutChildren[:insertIdx] + block + cutChildren[insertIdx:]

  // 7. Splice source back into cutLayers at to:
  //      newLayers = cutLayers[:to] + [source] + cutLayers[to:]

  // 8. Apply:
  //      c.back.itemList.Children = newChildren
  //      c.Layers                 = newLayers
  //      for i, l := range c.Layers: l.Index = i

  // 9. Warnings-as-failure (Inv-11): if proj.Warnings grew (shouldn't
  //    happen in this path — no re-parse — but defensive), rollback ALL.

  return nil
```

**Why no ID alloc / no nextItemID bump**: source's ID is unchanged; we just move its block. Project.nextItemID is untouched. (Reverse of DuplicateLayer §6 snapshot difference.)

**Why no follower mutation**: each follower carries per-layer state (timeline cache / fvdv / etc) that AE indexes by Layr ID, not by position in itemList. Move preserves that binding.

**Why adaptive block scan vs hardcoded 16**: handles AE-saved (16-chunk) and Go-built (2-chunk, just Layr + Ewst) layers uniformly. Same justification as DeleteLayer F4 (scar `ae-deletelayer-re.md`).

---

## 3. Refuse-cases

| Case | Detection | Reason |
|---|---|---|
| `from` out of range | `from < 0 \|\| from >= len(c.Layers)` | Standard Go-API bounds check |
| `to` out of range | `to < 0 \|\| to >= len(c.Layers)` | Standard Go-API bounds check (note: `to == len(c.Layers)-1` is in-range and means move to last slot) |
| Comp not from parse | `c.back == nil \|\| c.back.itemList == nil` | No itemList to splice |
| Source lacks Layr backref | `source.back == nil \|\| source.back.layrList == nil` | Defensive — caller passed a layer not in this comp's parse tree |
| Backref corruption | `Children[srcLayrIdx].FormType != IDLayr \|\| Children[srcLayrIdx+1].FormType != IDEwst` | Defensive — same as DeleteLayer / DuplicateLayer |
| Target chunk not found post-cut | `indexOf(cutChildren, target) < 0` | Should never trigger from a clean parse; defensive guard against the rare case where two layers' back-refs alias the same Layr chunk (corruption) |

No non-AV / track-matte refuse: MoveLayer doesn't care about layer type or matte intent — it's pure reorder. Camera / Light / track-matte layers all reorder cleanly.

---

## 4. Atomic invariants reuse

Standard V2.1 pattern. **Difference vs DeleteLayer / DuplicateLayer**:
- No `nextItemID` bump → no rollback of that counter
- No re-parse → `localWarnings` path is theoretical (no parser invocation that could surface new warnings); kept for symmetry but unlikely to fire
- Snapshot Layer.Index values across all layers (for rollback) — extra per-call cost is O(n) ints, negligible

---

## 5. Test plan

`internal/aep/move_layer_test.go`:

| Test | Setup | Verifies |
|---|---|---|
| `RefuseFromOutOfRange` | baseline 3-solid | `MoveLayer(-1, 0)` / `MoveLayer(3, 0)` return error |
| `RefuseToOutOfRange` | baseline 3-solid | `MoveLayer(0, -1)` / `MoveLayer(0, 3)` return error |
| `RefuseMissingBackref` | hand-built comp with `back = nil` | refused |
| `NoOpSameIndex` | baseline 3-solid | `MoveLayer(1, 1)` returns nil, no state change |
| `HappyPath_FirstToLast` | baseline 3-solid `[L1, L2, L3]` | `MoveLayer(0, 2)` → `[L2, L3, L1]`, Index re-assigned 0/1/2 |
| `HappyPath_LastToFirst` | baseline 3-solid | `MoveLayer(2, 0)` → `[L3, L1, L2]` |
| `HappyPath_MidToMid` | baseline 3-solid | `MoveLayer(1, 1)` no-op; `MoveLayer(1, 0)` → `[L2, L1, L3]` |
| `RoundTrip` | baseline 3-solid | Open → Move → Write → Reopen → expected order, Layer.Names match |
| `IndexFieldUpdated` | baseline 3-solid | After `MoveLayer`, every `c.Layers[i].Index == i` |
| `ItemListChildrenPreserved` | baseline 3-solid | Total itemList.Children count unchanged (no chunks lost / added) |

Acceptance: PASS 278 → 278 + N (N ≥ 8); FAIL = 0; vet clean.

---

## 6. Ship-gate (agent-side)

Modes for `tmp_debug/ge_move_layer/main.go` (baseline = `re_delete_layer_baseline.aep`):
- `first_to_last`: `MoveLayer(0, 2)` — L1_top to end
- `last_to_first`: `MoveLayer(2, 0)` — L3_bot to start
- `mid_swap`: `MoveLayer(1, 0)` — L2_mid up one

`test_data/verify_ge_move_layer.jsx` (mirror verify_ge_duplicate_layer.jsx):
- Open ge_*.aep
- Check `comp.numLayers == 3` (preserved — no add/remove)
- Check layer name ordering matches expected per mode

**Target**: 3 modes × 2 versions (AE 2020 + 2025) = 6/6 PASS.

---

## 7. File map

Create:
- `internal/aep/move_layer.go` — ~120 LOC
- `internal/aep/move_layer_test.go` — ~250 LOC
- `tmp_debug/ge_move_layer/main.go` — ~80 LOC
- `test_data/verify_ge_move_layer.jsx` — ~120 LOC

Modify:
- `workshop/plans/coverage.md` — add MoveLayer row in V3 structural ops section

No changes to:
- types (`Layer.back.layrList` already populated by parser)
- write paths (move is length-preserving at chunk level; WriteAEP walks whatever's in itemList.Children, order respected)
- parser (no new chunk semantics)

---

## 8. Tasks

- [ ] **Task 1**: this plan doc (you're reading it)
- [ ] **Task 2**: implement `move_layer.go` + tests (~10 tests)
- [ ] **Task 3**: ship-gate — `tmp_debug/ge_move_layer` + `verify_ge_move_layer.jsx` + 6 AE runs
- [ ] **Task 4**: Stable promotion + plan → finish/ + commit

---

## 9. Acceptance

1. `go vet ./...` clean + tests PASS 278 + N (N ≥ 8), FAIL = 0
2. AE 2020 + AE 2025 × 3 modes = 6/6 PASS via `verify_ge_move_layer.jsx`
3. Public API: `Composition.MoveLayer(int, int) error` godoc complete + marked Stable from start (no Alpha gate since no RE uncertainty)
4. Plan migrated to finish/, board.md updated
5. coverage.md V3 structural ops section gains MoveLayer row

---

## 10. Risk register

| Risk | Mitigation |
|---|---|
| Adjacent DLay/SLay/CLay/SecL LISTs in itemList get reorder-disturbed | Adaptive scan stops at LIST boundaries, so move only touches [Layr + Ewst + leaves] — DLay etc untouched. Test: ship-gate verifies AE doesn't reject moved file |
| Layer.Index field updated for some layers but not others on partial failure | Snapshot oldIndexes + roll back in warnings-as-failure path |
| Followers carry layer-position-derived state (timeline cache keyed by slot, not ID) — moved layer renders wrong slot | Adaptive scan moves the WHOLE block atomically, including ALL followers. Followers stay attached to their Layr. If AE caches anything by position, it'll re-resolve from ldta@0x00 (layer ID) — same as DeleteLayer's reorganized state |
| Two layers' back-refs alias the same Layr chunk (parse corruption) | Defensive `indexOf` returning -1 surfaces this; refuse |

---

## Related

- Spec: [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 Phase 4 candidate
- Phase 3 reference: [`finish/2026-05-28-v3-phase3-duplicatelayer-plan.md`](finish/2026-05-28-v3-phase3-duplicatelayer-plan.md)
- Phase 2 reference (block-splice machinery): [`finish/2026-05-28-v3-phase2-deletelayer-plan.md`](finish/2026-05-28-v3-phase2-deletelayer-plan.md) + scar [`../scars/ae-deletelayer-re.md`](../scars/ae-deletelayer-re.md) F4
- AE acceptance gate: [`../scars/ae25-acceptance-gate.md`](../scars/ae25-acceptance-gate.md)
