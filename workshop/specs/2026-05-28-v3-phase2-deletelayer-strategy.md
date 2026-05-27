# V3 Phase 2 DeleteLayer — Implementation Strategy

**Status**: pending — awaiting commit acceptance before Task 3 implementation
**Source of truth**: [`scars/ae-deletelayer-re.md`](../scars/ae-deletelayer-re.md) (6 Findings from AE 2020 + AE 2025 fixtures)
**Plan**: [`plans/2026-05-28-v3-phase2-deletelayer-plan.md`](../plans/2026-05-28-v3-phase2-deletelayer-plan.md) Task 2

This doc closes Plan Task 2. After commit, Task 3 (implementation) executes from this matrix without re-deriving decisions.

---

## 1. Public API signature

```go
// DeleteLayer removes the layer at the given 0-based index in
// c.Layers (Go slice convention). Returns nil on success, or an
// error if the index is out of range, the comp lacks a parsed
// itemList back-reference, the target layer is not a regular AV
// layer (camera/light/audio refused in Phase 2), or removing the
// layer would leave the composition empty.
//
// Reference cleanup: per AE's own delete behavior (RE'd via 4
// AE-saved fixtures, see scars/ae-deletelayer-re.md):
//   - any other layer's Layer.ParentID == deleted.ID → reset to 0
//   - any other layer's Layer.TrackMatteLayerID == deleted.ID → reset to 0
//   - Layer.TrackMatte byte (ldta @0x6B) on neighbors: left untouched
//   - Project.nextItemID counter: left untouched (IDs not reused; Inv-9)
//
// String-level references to the deleted layer's ID (expressions,
// render queue refs, essential graphics) are out of scope for Phase
// 2 — callers must scrub these manually if needed.
//
// Atomic mutation (Inv-10/Inv-11): on warning during delete, all
// state mutated by this call is rolled back to the pre-call snapshot.
func (c *Composition) DeleteLayer(index int) error
```

**Index convention**: 0-based to match `c.Layers` Go slice. Rejected 1-based (AE Scripting style) — it adds ambiguity without symmetry payoff (NewShapeLayer doesn't return an index either).

**Alpha tag in godoc** until AE 2020 + AE 2025 ship-gate green (CLAUDE.md #2 + #6).

---

## 2. Decision matrix (RE Finding → impl rule)

| Concern | RE Finding | Decision |
|---|---|---|
| `itemList.Children` removal mode | F1 splice (Δ-16 chunks) | **splice via `append(s[:i], s[j:]...)`** — j computed adaptively (see §3) |
| Ewst sibling | F4 co-deleted | **co-delete** — assert `Children[i+1].FormType == IDEwst`, fail-rollback if not |
| Per-Layer follower chunks (fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2 = 14) | F4 co-deleted in AE-saved files | **adaptive splice** — see §3. NewShapeLayer-built layers carry 0 followers; AE-saved layers carry 14. Don't hardcode 16 |
| `Layer.ParentID` orphan | F2 reset to 0 | **reset to 0** — for every `c.Layers[k]` (k ≠ deletedIdx), if `Layer.ParentID == deleted.ID` set both struct field and ldta @0x84..0x87 bytes to 0 |
| `Layer.TrackMatteLayerID` orphan | F3 reset to 0 (`@0xA0..0xA3`) | **reset to 0** — same pattern as ParentID, at ldta @0xA0..0xA3. Skip if back-ref nil (layer built without parse) |
| `Layer.TrackMatte` byte (`@0x6B`) | F3 untouched by AE | **leave alone** — match AE; user must clear via UI/SetTrackMatte API |
| head chunk `nextItemID` counter | F6 IDs not reused | **untouched** — Invariant #9 monotonic still holds |
| `layerBackrefs.opaque` (Phase 1 placeholder) | not needed in Phase 2 | **don't enable** — 14 followers are at Item LIST level (not Layer LIST level), and adaptive splice consumes them atomically. Defer to Phase 3 if InsertLayer/DuplicateLayer needs per-layer cloning |
| Pseudo-layer LISTs (DLay/SLay/CLay/SecL) | F5 same 16-chunk pattern but not in `c.Layers` | **refuse** — filtered automatically by index-based API (these aren't in `c.Layers`), but assert `Children[i].FormType == IDLayr` defensively in case `back.layrList` was misset |

---

## 3. Adaptive splice algorithm

NewShapeLayer inserts only Layr + Ewst (2 chunks; no followers) and ship-gates green; AE-saved files have Layr + Ewst + 14 follower leaves (16 chunks total). DeleteLayer must handle both.

```
algorithm: splice(itemList.Children, deletedLayer):
  // 1. Locate the Layr LIST in the Item LIST children.
  i := index of deletedLayer.back.layrList in itemList.Children
  if i < 0:           return ErrLayerChunkNotFound
  if Children[i].FormType != IDLayr:  return ErrCorruptBackref  // defensive

  // 2. Assert Ewst follows.
  if i+1 >= len(Children) || Children[i+1].FormType != IDEwst:
    return ErrMissingEwstSibling  // AE-required boilerplate per F4

  // 3. Find end of trailing leaf block: consume leaves until next LIST or EOF.
  j := i + 2  // start past Layr + Ewst
  for j < len(Children) && !Children[j].IsList():
    j++
  // Now Children[j] is either a LIST chunk (next layer / service / cifs / Gide)
  // or j == len(Children). Either way, [i, j) is the block to remove.

  // 4. Splice.
  Children = append(Children[:i], Children[j:]...)
```

**Why this is RE-safe**:
- Matches F1 (splice) by structural definition
- Matches F4 (16-chunk removal) for AE-saved files because trailing 14 leaves get consumed by the leaf-scan loop
- Handles Go-built layers (2-chunk removal) without over-eating service block chunks
- Handles partial-state files (e.g. round-trip-then-add layers) by reading actual byte structure rather than assuming a count

**Implementation note**: the chunks consumed in step 3 are NOT moved into any backref shard — they're discarded entirely. If a future feature needs to preserve "soft delete" or "undo" we'll capture them then. For Phase 2: gone is gone.

---

## 4. Reference cleanup pass

After locating the layer and BEFORE splicing, walk neighbors to clear references. The cleanup is per-comp (deletedLayer's ParentID/TrackMatteLayerID don't matter since the layer is going away).

```go
deletedID := deleted.ID
for k, neighbor := range c.Layers {
    if k == deletedIdx { continue }
    if neighbor.ParentID == deletedID {
        neighbor.ParentID = 0
        if neighbor.back != nil && neighbor.back.ldta != nil {
            writeUint32BE(neighbor.back.ldta.Data, 0x84, 0)  // ldta @0x84..0x87
        }
    }
    if neighbor.TrackMatteLayerID == deletedID {
        neighbor.TrackMatteLayerID = 0
        if neighbor.back != nil && neighbor.back.ldta != nil &&
           len(neighbor.back.ldta.Data) >= 0xA4 {
            writeUint32BE(neighbor.back.ldta.Data, 0xA0, 0)  // ldta @0xA0..0xA3
        }
        // Deliberately DO NOT touch @0x6B (TrackMatte byte) — F3.
    }
}
```

Skip ldta write if `neighbor.back.ldta == nil` (layer was built outside the parser; struct field is the source of truth in that case).

Bounds check on ldta length for `@0xA0..0xA3`: AE 2020 files may not have the `TrackMatteLayerID` field (introduced AE 23+) — `len(ldta.Data) >= 0xA4` guards this. ParentID at @0x84 is present in all versions.

---

## 5. Refuse-cases (Phase 2 conservative)

| Case | Detection | Reason for refuse |
|---|---|---|
| Index out of range | `index < 0 \|\| index >= len(c.Layers)` | Standard Go-API bounds check |
| Comp not from parse | `c.back == nil \|\| c.back.itemList == nil` | No itemList to splice; refuse with clear error rather than partial mutation |
| Layer not AV | `c.Layers[index].Type != LayerTypeAV` | Camera/Light/Audio ldta layout differs (LightKind @0x88, camera-specific fields). AE's delete behavior NOT RE'd for these. Future RE can lift this restriction once fixtures exist |
| Single-layer comp | `len(c.Layers) == 1` | AE behavior on "empty comp" delete NOT RE'd. AE may treat 0-layer comps specially; refuse pre-emptively. Future RE can lift |
| Backref corruption | `Children[i].FormType != IDLayr \|\| Children[i+1].FormType != IDEwst` | Defensive — should never trigger from a freshly-parsed comp. Indicates state corruption upstream; refuse rather than over-delete |

Error messages should name the case ("DeleteLayer: index 5 out of range (have 3 layers)", "DeleteLayer: refuse non-AV layer (Type=camera); not yet supported", etc.) so callers can branch on the error string in Phase 3 when we lift restrictions.

---

## 6. Atomic invariants reuse

Standard V2.1 pattern (`scars/ae25-acceptance-gate.md`, NewShapeLayer reference impl):

```
1. Snapshot pre-mutation state:
   - oldItemChildren := slices.Clone(c.back.itemList.Children)
   - oldLayers       := slices.Clone(c.Layers)
   - oldWarningsLen  := len(c.proj.Warnings)
   - oldLdtaBytes    := map[*rifx.Chunk][]byte{} — for every neighbor we plan to mutate ldta on, save Data slice copy

2. Validate (refuse-cases §5) — return early without mutation if any fail

3. Mutate:
   - Reference cleanup pass (§4) — writes ldta bytes on neighbors
   - Adaptive splice (§3) — modifies itemList.Children
   - c.Layers = append(c.Layers[:index], c.Layers[index+1:]...)

4. Warnings-as-failure (Inv-11): if len(c.proj.Warnings) > oldWarningsLen → rollback:
   - c.back.itemList.Children = oldItemChildren
   - c.Layers = oldLayers
   - for chunk, data := range oldLdtaBytes: chunk.Data = data
   - c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
   - return error with all new warnings appended

5. Return nil
```

**Snapshotting cost**: 3 slice clones (Children typically 50-300 entries; Layers typically 1-30) + N ldta byte saves (N = neighbors with matching ParentID/TrackMatteLayerID, typically 0-3). Negligible.

---

## 7. Open / deferred to Phase 3+

Recorded so future-Phase-N can pick them up without re-RE:

- **Camera/Light/Audio delete** — RE needed (fixture: `re_delete_layer_nonav.jsx` with comp.layers.addCamera/addLight). Lift the Type≠AV refuse case after.
- **Single-layer comp delete** — RE needed (variant of existing JSX: pre-delete state has only 1 layer). Lift the `len==1` refuse case after.
- **String-level ref cleanup** — expressions / render queue / essential graphics carrying the deleted layer's ID. Out of scope for Phase 2 (per plan risk register). Phase 4+ if user demand surfaces.
- **Per-layer follower chunk tracking** — for InsertLayer/DuplicateLayer cloning. Phase 3. The 14 followers per layer need to be modeled as `layerBackrefs.followers []*rifx.Chunk` (or similar) so they can be cloned. Phase 2's adaptive-splice doesn't need this.
- **Multi-layer batch delete** — `DeleteLayers([]int)` convenience API. Trivial wrapper but call-order matters (sort descending to keep indices stable). Phase 3+ if demand exists.

---

## 8. Ship-gate criteria (Plan Task 5)

Spec is "done" when Task 3 implementation + Task 4 Go round-trip + Task 5 AE ship-gate all pass:

1. `go vet ./...` clean + `go test -count=1 ./internal/aep/...` PASS = 259 + N (N ≥ 4 new DeleteLayer tests)
2. AE 2020 + AE 2025 each open the 4 Go-emitted `test_data/ge_delete_layer_<mode>.aep` fixtures without "file data missing" / silent-drop
3. Byte-equivalence: Go's `Open(re_delete_layer_baseline.aep) → DeleteLayer(1) → WriteAEP` produces a file that, after trimming AE timestamp/UUID variance, matches `re_delete_layer_middle.aep` byte-for-byte (or, if AE writes extra metadata we don't, matches the round-trip-stable subset)

If byte-equivalence fails: use scars/ae25-acceptance-gate.md Stage 4 bisection (cut back changes until AE accepts, then add forward).

---

## 9. Implementation file map

Create:
- `internal/aep/delete_layer.go` — `DeleteLayer` + helpers (splice, refcheck, snapshot)
- `internal/aep/delete_layer_test.go` — unit tests (refuse-cases happy path + neighbor cleanup + round-trip + byte-equivalence)

Modify:
- `internal/aep/new_layer.go` — extract `findLayrIndexInItemList(itemList, layrChunk) int` as shared helper (DeleteLayer needs to locate by chunk pointer; NewShapeLayer already does similar)
- godoc / doc.go — add DeleteLayer to public-API list

No changes to:
- types (Layer / Composition / layerBackrefs already have everything we need from Phase 1)
- write paths (length-preserving WriteAEP doesn't care about splice; itemList.Children walks what's there)
- parser (no new chunk semantics)
