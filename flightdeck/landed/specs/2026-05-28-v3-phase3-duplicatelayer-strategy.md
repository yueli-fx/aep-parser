# V3 Phase 3 DuplicateLayer — Implementation Strategy

**Status**: pending — awaiting commit acceptance before Task 3 implementation
**Source of truth**: [`scars/ae-duplicatelayer-re.md`](../scars/ae-duplicatelayer-re.md) (10 Findings from AE 2020 4-mode fixtures + byte-level verification)
**Plan**: [`plans/2026-05-28-v3-phase3-duplicatelayer-plan.md`](../plans/2026-05-28-v3-phase3-duplicatelayer-plan.md) Task 2

This doc closes Plan Task 2. After commit, Task 3 (implementation) executes from this matrix without re-deriving decisions. Mirror of [`finish/2026-05-28-v3-phase2-deletelayer-strategy.md`](finish/2026-05-28-v3-phase2-deletelayer-strategy.md) section structure for predictable reading.

---

## 1. Public API signature

```go
// DuplicateLayer clones the layer at the given 0-based index in
// c.Layers (Go slice convention) and inserts the clone at that
// same position, pushing the source layer and everything below
// down by one (mirroring AE ScriptingAPI's layer.duplicate()).
// Returns the cloned *Layer on success, or an error if the index
// is out of range, the comp lacks a parsed itemList back-reference,
// the target layer is not a regular AV layer (camera/light/audio
// refused in Phase 3), the target has track-matte intent set
// (TrackMatte != TrackMatteNone — refused in Phase 3 due to AE's
// matte-preservation quirk, see scars/ae-duplicatelayer-re.md F2),
// or `name` is empty.
//
// Clone semantics (RE'd via 4 AE-saved fixtures + byte-diff,
// scars/ae-duplicatelayer-re.md F1/F3/F4/F5/F7/F8/F10):
//   - new layer ID = proj.allocItemID() (head counter +1, monotonic)
//   - new layer's 16-chunk block (Layr + Ewst + 14 follower leaves)
//     is a deep byte-clone of source's block, with ldta @0x00..0x03
//     overwritten with the new ID
//   - new layer's Layer.SourceID = source.SourceID (footage NOT
//     duplicated; both layers reference the same footage item)
//   - new layer's Layer.ParentID = source.ParentID (verbatim)
//   - new layer's Layer.TrackMatte byte (ldta @0x6B) = source value
//     (verbatim — but Phase 3 refuses non-None per refuse-cases §5)
//   - new layer's Name = caller-supplied `name` (AE's
//     layer.duplicate() keeps source's name verbatim; we require
//     explicit name to avoid silent duplicate-name confusion)
//   - children's ParentID is NOT updated to clone (F6 — clone is a
//     fresh shadow; source remains the canonical parent for any
//     incoming refs)
//
// Atomic mutation (Inv-10/Inv-11): on warning during clone, all
// state mutated by this call is rolled back to the pre-call
// snapshot, INCLUDING the proj.nextItemID bump.
func (c *Composition) DuplicateLayer(index int, name string) (*Layer, error)
```

**Index convention**: 0-based to match `c.Layers` Go slice (same as DeleteLayer §1).

**Why explicit `name` param** (F9 + plan Q7 conclusion):
- AE's `layer.duplicate()` JSX path produces silent duplicate names — bad caller UX
- Requiring `name` forces caller to think about naming, matches NewShapeLayer signature shape
- We do NOT validate name uniqueness against existing layers — AE accepts duplicates, so do we

**Alpha tag in godoc** until AE 2020 + AE 2025 ship-gate green (CLAUDE.md #2 + #6).

---

## 2. Decision matrix (RE Finding → impl rule)

| Concern | RE Finding | Decision |
|---|---|---|
| Insertion index in `c.Layers` | F1 source pushed down | **insert at `index`** — same position as source's pre-call index; clone takes the slot, source moves to `index+1`. Matches AE behavior for solo/dup_parent/dup_child (3/4 modes) |
| Insertion position in itemList.Children | F1 + F5 | **insert deep-cloned 16-chunk block immediately BEFORE source's Layr** — locate source's Layr via `findLayrIndexInItemList`, splice clone block in at that index |
| Block size to clone | F5 (+16 per dup) + F10 (byte-identical) | **adaptive clone**: from source's Layr index, copy Layr + Ewst + all leaf followers until next LIST/EOF (mirrors DeleteLayer §3 adaptive splice — handles AE-saved 16-block and Go-built 2-block alike) |
| Per-chunk Data slice | concurrent-mutate scar | **fresh copy**: every cloned chunk's `Data = append([]byte(nil), src.Data...)`; LIST chunks recurse on Children. NO slice sharing |
| New layer ID | F3 (head counter +1) | **`proj.allocItemID()`** — same as NewShapeLayer line 139-141. Bump `proj.nextItemID` |
| ldta @0x00..0x03 (layer ID) | F10 (only mutation) | **overwrite with new ID** (big-endian uint32) — `writeUint32BE(clone.back.ldta.Data, 0x00, newID)` |
| ldta @0x04+ (all other body) | F10 byte-identical | **leave verbatim** — body carries SourceID @0x28, ParentID @0x84, TrackMatte @0x6B unchanged from source per F4/F7/F8 |
| Layer.SourceID (struct field) | F4 shared footage | **copy from source** — `clone.SourceID = source.SourceID`. No footage clone |
| Layer.ParentID (struct field) | F7 verbatim | **copy from source** — `clone.ParentID = source.ParentID`. Struct field matches ldta bytes (which are already correct via verbatim body clone) |
| Layer.TrackMatteLayerID (struct field) | F7 inverse (out-ref copy) | **copy from source** — `clone.TrackMatteLayerID = source.TrackMatteLayerID`. (Refuse-cases §5 currently blocks non-zero; field set for parity) |
| Layer.TrackMatte (struct field, byte @0x6B) | F8 verbatim | **copy from source** — refuse-cases §5 currently filters non-None paths, but field is set for parity with future lift |
| Children's outgoing refs | F6 source remains canonical | **leave alone** — DO NOT walk c.Layers to rewrite anyone's ParentID/TrackMatteLayerID. AE doesn't; we don't |
| Layer name (Utf8 chunk inside clone's Layr LIST) | F9 + signature decision | **rewrite with caller-supplied `name`** — length-variable path (CLAUDE.md #1 exception). Use existing name-setter helper or write new minimal one targeting the cloned Layr LIST's Utf8 child |
| head chunk `nextItemID` counter | F3 monotonic | **bump via `proj.allocItemID()`** (already covered above). Rollback must un-bump (§6) |

---

## 3. Deep-clone algorithm

```
algorithm: cloneLayrBlock(itemList.Children, srcLayrIdx) → []*rifx.Chunk:
  // Mirror of DeleteLayer's adaptive splice (§3 of DeleteLayer strategy).
  // Determines block extent [srcLayrIdx, j) by scanning leaf followers.

  i := srcLayrIdx
  if Children[i].FormType != IDLayr:        return ErrCorruptBackref
  if i+1 >= len(Children) || Children[i+1].FormType != IDEwst:
                                            return ErrMissingEwstSibling

  // Find end of trailing leaf block (same scan as DeleteLayer).
  j := i + 2
  for j < len(Children) && !Children[j].IsList():
    j++

  // Deep-clone block [i, j).
  out := make([]*rifx.Chunk, 0, j-i)
  for k := i; k < j; k++:
    out = append(out, deepCloneChunk(Children[k]))
  return out

algorithm: deepCloneChunk(src) → *rifx.Chunk:
  // Recursive deep copy: every Data slice is freshly allocated;
  // LIST children are cloned recursively. NO slice sharing with src.
  c := &rifx.Chunk{
    ID:       src.ID,
    Size:     src.Size,
    FormType: src.FormType,
    Data:     append([]byte(nil), src.Data...),
  }
  if src.Children != nil:
    c.Children = make([]*rifx.Chunk, len(src.Children))
    for k, child := range src.Children:
      c.Children[k] = deepCloneChunk(child)
  return c
```

**Why this is RE-safe**:
- Matches F5 (Δ+16 itemList children) for AE-saved files via the adaptive leaf scan
- Matches F10 (byte-identical block) by construction (deep copy with no slice sharing)
- Handles Go-built source layers (Layr + Ewst only) without consuming downstream service chunks — same adaptive logic as DeleteLayer §3

**Algorithm:**

```
algorithm: duplicateLayer(c, index, name) → (*Layer, error):
  // 1. Validate (§5 refuse-cases) — return early without mutation if any fail.

  // 2. Locate source's Layr in itemList.
  source := c.Layers[index]
  srcLayrIdx := findLayrIndexInItemList(c.back.itemList, source.back.layrList)
  if srcLayrIdx < 0:  return nil, ErrLayerChunkNotFound

  // 3. Snapshot (§6) — capture pre-mutation state for rollback.

  // 4. Deep-clone source's 16-chunk block.
  cloneBlock := cloneLayrBlock(c.back.itemList.Children, srcLayrIdx)
  // cloneBlock[0] is the cloned Layr LIST; cloneBlock[1] is cloned Ewst; rest are leaf followers.

  // 5. Mutate clone's ldta @0x00..0x03 with new ID.
  newID := c.proj.allocItemID()
  clonedLayrList := cloneBlock[0]
  clonedLdta := findChildByID(clonedLayrList, IDldta)   // ldta is leaf inside Layr LIST
  writeUint32BE(clonedLdta.Data, 0x00, newID)

  // 6. Rewrite clone's name (Utf8 chunk inside clone's Layr LIST).
  setLayerNameInPlace(clonedLayrList, name)  // length-variable; updates Layr LIST size + Utf8 chunk

  // 7. Splice clone block into itemList.Children at srcLayrIdx (BEFORE source).
  c.back.itemList.Children = sliceInsert(c.back.itemList.Children, srcLayrIdx, cloneBlock...)

  // 8. Build clone *Layer struct and insert into c.Layers at index.
  cloneLayer := buildLayerFromChunks(clonedLayrList, cloneBlock[1:], c)  // reuses parse helpers
  cloneLayer.ID                 = newID
  cloneLayer.Name               = name
  cloneLayer.SourceID           = source.SourceID
  cloneLayer.ParentID           = source.ParentID
  cloneLayer.TrackMatteLayerID  = source.TrackMatteLayerID
  cloneLayer.TrackMatte         = source.TrackMatte
  c.Layers = sliceInsert(c.Layers, index, cloneLayer)

  // 9. Warnings-as-failure (Inv-11) → rollback (§6) if any new warnings.

  return cloneLayer, nil
```

**Note on `buildLayerFromChunks`**: implementation may either (a) re-run the parse path on the cloned chunks to build a fresh `*Layer` with proper backrefs, or (b) struct-clone source's `*Layer` and rebind backrefs to the cloned chunks. Task 3 picks based on what's cleaner — (a) is more robust; (b) is fewer lines. The decision matrix above defines the OUTPUT state either way.

---

## 4. Reference cleanup pass — NONE

Unlike DeleteLayer, **DuplicateLayer does NOT walk c.Layers to rewrite neighbor refs**:

- F6 confirms AE does NOT update children's ParentID to point at the clone (clone is a sibling shadow, not a promoted parent)
- F7 confirms clone's own outgoing refs (ParentID, TrackMatteLayerID) are byte-copied from source — handled by §3's verbatim block clone, not by a separate pass
- No orphan-reset logic needed (no IDs disappearing)

This section exists explicitly so future-reader knows we considered and rejected a cleanup pass — match-AE-behavior dictates "leave neighbors alone."

---

## 5. Refuse-cases (Phase 3 conservative)

| Case | Detection | Reason for refuse |
|---|---|---|
| Index out of range | `index < 0 \|\| index >= len(c.Layers)` | Standard Go-API bounds check |
| Comp not from parse | `c.back == nil \|\| c.back.itemList == nil` | No itemList to splice; refuse with clear error rather than partial mutation |
| Layer not AV | `c.Layers[index].Type != LayerTypeAV` | Camera/Light/Audio ldta layout differs. AE's duplicate behavior NOT RE'd for these. Future RE can lift this restriction once fixtures exist |
| Source has TrackMatte != None | `c.Layers[index].TrackMatte != TrackMatteNone` | F2 quirk: AE relocates clone to "above matte source" position (not source's old index) to preserve original's implicit matte. Implementing this special case adds ~30 LOC for a niche scenario — Phase 3 refuses, caller must clear matte intent first. Lift in Phase 3.1 if demand surfaces |
| Empty name | `name == ""` | Forces caller intent; AE accepts empty names but we don't (consistency with NewShapeLayer name validation) |
| Backref corruption | `Children[i].FormType != IDLayr \|\| Children[i+1].FormType != IDEwst` | Defensive — should never trigger from a freshly-parsed comp. Indicates state corruption upstream; refuse rather than over-clone |
| Source is shape/text layer (deferred) | `source` is shape layer or text layer | Deferred per plan Q8 — embed bytes (btds / tdgp ID refs) may share state; naive byte-clone could produce 2 layers pointing at same embed. AV-only refuse case covers this transitively (shape/text are AV-typed; need a dedicated detection if we want a clearer error msg — Task 3 decides) |

Error messages should name the case ("DuplicateLayer: index 5 out of range (have 3 layers)", "DuplicateLayer: refuse layer with TrackMatte set (Type=Alpha); clear matte intent first or wait for Phase 3.1", etc.) so callers can branch on the error string in Phase 3.1+ when we lift restrictions.

---

## 6. Atomic invariants reuse

Standard V2.1 pattern (`scars/ae25-acceptance-gate.md`, NewShapeLayer + DeleteLayer reference impl):

```
1. Snapshot pre-mutation state:
   - oldItemChildren := slices.Clone(c.back.itemList.Children)
   - oldLayers       := slices.Clone(c.Layers)
   - oldWarningsLen  := len(c.proj.Warnings)
   - oldNextItemID   := c.proj.nextItemID          ← critical: clone allocates new ID

2. Validate (refuse-cases §5) — return early without mutation if any fail

3. Mutate:
   - allocItemID (bumps nextItemID)
   - deep-clone block + mutate ldta @0x00..0x03 + setLayerNameInPlace
   - splice into itemList.Children at srcLayrIdx
   - sliceInsert into c.Layers at index

4. Warnings-as-failure (Inv-11): if len(c.proj.Warnings) > oldWarningsLen → rollback:
   - c.back.itemList.Children = oldItemChildren
   - c.Layers = oldLayers
   - c.proj.nextItemID = oldNextItemID             ← un-bump so next alloc is correct
   - c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
   - return nil, error with all new warnings appended

5. Return cloneLayer, nil
```

**Difference vs DeleteLayer**: we do NOT need to snapshot neighbor ldta bytes (no reference cleanup pass — §4). We DO need to snapshot `proj.nextItemID` (DeleteLayer doesn't bump it; we do).

**Snapshotting cost**: 2 slice clones + 2 scalars. Negligible. The deep-cloned chunks themselves are not in the snapshot — they're discarded on rollback by virtue of `itemList.Children` being restored (the orphaned clone chunks become garbage).

---

## 7. Open / deferred to Phase 3.1+

Recorded so future-Phase-N can pick them up without re-RE:

- **Source with TrackMatte != None** (F2 quirk) — implement AE's "above matte source" relocation. ~30 LOC special-case in §3 step 7 (compute insertion position differently when matte detected). Needs fixture re-RE only if matte semantics shifted in AE 25+ (F2 was AE 2020).
- **AE 23+ explicit TrackMatteLayerID (ldta @0xA0)** — scar Open notes this needs AE 2025 RE with explicit setTrackMatte before duplicate. Phase 3 covers implicit (AE 2020) path; default assumption (byte-copy via verbatim block clone) likely holds but unverified. Lift when fixture exists.
- **Camera/Light/Audio duplicate** — RE needed (fixture: variant of `re_duplicate_layer.jsx` using comp.layers.addCamera/addLight). Lift the Type≠AV refuse case after.
- **Shape/Text layer duplicate** — RE needed for embed bytes (btds/tdgp ID refs); naive byte-clone may produce 2 layers pointing at same embed dict. Lift after embed-byte RE in Phase 4+.
- **Cross-composition duplicate** (`InsertLayer(src *Layer, atIdx int)`) — Phase 4+.
- **Multi-select batch duplicate** — `DuplicateLayers([]int)` wrapper. Trivial but call-order matters (sort descending or insert-then-shift). Phase 3+ if demand surfaces.
- **Auto-suffix " 2" naming policy** — we explicitly chose caller-supplied name (§1). If demand surfaces for AE-matching auto-suffix wrapper, add `DuplicateLayerAutoName(index int) (*Layer, error)` as convenience over the explicit-name API. Phase 3.1+.
- **String-level ref cloning** — expressions referencing source layer's ID don't auto-rewrite to clone (out of scope; AE itself doesn't do this either).

---

## 8. Ship-gate criteria (Plan Task 5)

Spec is "done" when Task 3 implementation + Task 4 Go round-trip + Task 5 AE ship-gate all pass (mirrors Phase 2 §8):

1. `go vet ./...` clean + `go test -count=1 ./internal/aep/...` PASS = 267 + N (N ≥ 5 new DuplicateLayer tests)
2. AE 2020 + AE 2025 each open the Go-emitted `test_data/ge_duplicate_layer_<mode>.aep` fixtures without "file data missing" / silent-drop. Modes: `solo`, `dup_parent`, `dup_child` (matte refused per §5; `dup_matted` Go-build will produce a refuse error in test, not a ge_*.aep fixture — document the absence as expected, not a coverage gap)
3. Byte-equivalence (best-effort): Go's `Open(re_duplicate_layer_solo.aep before dup) → DuplicateLayer(idx, name) → WriteAEP` produces a file whose `itemList.Children` matches AE's `re_duplicate_layer_solo.aep` post-dup state (chunk-shape compare; AE timestamp/UUID variance trimmed). F10 guarantees byte-identical at the clone-block level so this should hold modulo AE metadata jitter

If ship-gate FAIL on any mode: scars/ae25-acceptance-gate.md Stage 4 bisection. Most likely failure modes:
- Follower chunk we assumed verbatim actually carries layer-ID-derived state → bisect by cloning fewer followers, see which one AE rejects when missing or wrong
- ldta @0x00 mutation missed a co-located ID elsewhere (e.g. ldta has a self-ID byte we missed) → byte-diff Go output vs AE output

Phase 2 set the playbook: 8/8 PASS earns Stable promotion. For Phase 3: 3 modes × 2 versions = 6 PASS target (matte excluded by design).

---

## 9. Implementation file map

Create:
- `internal/aep/duplicate_layer.go` — `DuplicateLayer` + helpers (`cloneLayrBlock`, `deepCloneChunk`, `setLayerNameInPlace` if not already extractable from existing name-write path)
- `internal/aep/duplicate_layer_test.go` — unit tests (refuse-cases happy path + clone identity + round-trip + concurrent-mutate safety + structural equivalence vs AE's solo/dup_parent/dup_child fixtures)

Modify:
- `internal/aep/delete_layer.go` — extract `findLayrIndexInItemList` into shared file `layer_itemlist.go` (or `scene_layer.go`) if it isn't already — DuplicateLayer needs the same locator. (Plan Task 3 explicitly lists this.)
- godoc / doc.go — add DuplicateLayer to public-API list

No changes to:
- types (Layer / Composition / layerBackrefs already have everything we need from Phase 1)
- write paths (length-preserving WriteAEP doesn't care about insert; itemList.Children walks what's there. Length-variable name write IS exercised — but that path already exists for SetName)
- parser (no new chunk semantics — clone uses existing parse helpers)
