# V3 Phase 5C InsertLayer — Design Strategy

**Status**: design draft — pending plan + implementation
**Builds on**: `landed/specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md` (DuplicateLayer same-comp clone, Stable post-5A/5B), `landed/specs/2026-05-28-v3-phase2-deletelayer-strategy.md` (adaptive splice pattern)
**Picklist origin**: `specs/2026-05-27-v3-deep-think.md` § 4 — Phase 5 残余候选
**Scope decision**: Phase 5C ships same-Project sibling-comp insert only. Cross-Project deferred to 5C.1.

This doc is the strategy matrix. Once approved → plan in `flight-plans/2026-05-29-v3-phase5c-insertlayer-plan.md`; implementation executes from this matrix without re-deriving decisions.

---

## 1. Public API

```go
// InsertLayer deep-clones the given source layer (from a sibling comp
// within the SAME Project as c) into c.Layers at atIdx (0-based, Go
// slice convention). The clone takes the slot; existing layers at
// atIdx.. shift down by one. Returns the inserted *Layer.
//
// Cross-comp clone semantics (Phase 5C; same-Project only):
//   - new layer ID = c.proj.allocItemID() (head counter +1)
//   - clone block = deep byte-copy of src's [Layr, Ewst, leaf-followers)
//     (adaptive scan to next LIST/EOF — same machinery as DuplicateLayer §4)
//   - clone.SourceID = src.SourceID (verbatim — the referenced
//     Footage/Comp item exists in the shared Project; no footage
//     duplication)
//   - clone.ParentID = 0 (RESET — src's ParentID named a layer ID
//     in src.comp which means nothing in c)
//   - clone.TrackMatteLayerID = 0 + clone.TrackMatte byte forced to
//     TrackMatteNone (RESET — same reason as ParentID)
//   - clone.Name = src.Name (verbatim — AE ScriptingAPI's
//     layer.copyToComp(comp) preserves source name)
//
// Refuse-cases (Phase 5C conservative; §2):
//   - src nil; dest comp missing itemList back-ref; atIdx out of range
//   - cross-Project (src.comp.proj != c.proj) — Phase 5C.1+
//   - same-comp (src.comp == c) — use DuplicateLayer instead
//   - src.Type != LayerTypeAV — non-AV deferred (mirrors DuplicateLayer
//     Phase 5B refuse)
//   - direct pre-comp loop: src.SourceID == c.ID
//   - backref corruption (Layr formType / Ewst sibling mismatch)
//
// Atomic mutation (Inv-10 / Inv-11): snapshot dest itemList.Children +
// c.Layers + c.proj.nextItemID + c.proj.Warnings count; on any new
// parser warning during the re-parse, roll all back including the
// nextItemID bump.
//
// Stable promotion: pending AE 2020 + AE 2025 ship-gate
// (3 modes × 2 versions = 6 PASS). Marked Alpha in godoc until then.
func (c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error)
```

**Index convention**: 0-based to match `c.Layers` Go slice. `atIdx == len(c.Layers)` is accepted as "append to bottom" — diverges from DuplicateLayer (which uses the source's existing index, so always strictly `< len`). MoveLayer (Phase 4) chose strict `< len` for the same reason DuplicateLayer did (source must exist at `from`); InsertLayer has no such constraint on the dest side, so accepting `== len` matches Go's `slice[:n]` mental model.

**No `name` param**: implicit clone of `src.Name` matches AE's `layer.copyToComp` and saves callers from manual lookup. Users wanting a rename do `cloned.SetName(...)` post-call (existing length-variable path).

**Alpha tag in godoc** until ship-gate green (CLAUDE.md #2 + #6).

---

## 2. Refuse-case matrix

| # | Detection | Reason for refuse |
|---|---|---|
| R1 | `src == nil` | nil-safe guard |
| R2 | `c.back == nil \|\| c.back.itemList == nil` | dest has no itemList to splice; refuse with clear error rather than partial mutation |
| R3 | `c.proj == nil` | need `allocItemID()`; dest must be project-owned |
| R4 | `atIdx < 0 \|\| atIdx > len(c.Layers)` | range check — `== len` is "append", `> len` rejected |
| R5 | `src.comp == nil` | src is detached from any comp; can't compute src's chunk-block extent |
| R6 | `src.comp == c` | same-comp redirect — "use DuplicateLayer" |
| R7 | `src.comp.proj != c.proj` | cross-Project deferred to Phase 5C.1 (needs Footage/Comp deep-clone strategy) |
| R8 | `src.Type != LayerTypeAV` | mirrors DuplicateLayer Phase 5B refuse — non-AV ldta layouts not yet ship-gated |
| R9 | `src.SourceID == c.ID && src.SourceID != 0` | direct pre-comp loop — inserting a layer whose source is the dest comp itself produces an immediate cycle; AE rejects, we refuse upfront |
| R10 | `src.back == nil \|\| src.back.layrList == nil` | src lacks chunk backrefs — built outside parser |
| R11 | structural assert: `Children[i].FormType != IDLayr \|\| Children[i+1].FormType != IDEwst` | corruption defense; same pattern as DuplicateLayer §5 |

**Not refuse-cases** (deliberately):
- `src` has `ParentID != 0` or `TrackMatteLayerID != 0` — we reset to 0 in the clone (§3 step 5). Refusing would force callers into a manual pre-step that adds no safety.
- Name collision with existing layer in `c` — AE allows duplicate layer names; we mirror.

Indirect pre-comp loops (clone refs compX which refs c via a chain) — out of scope for Phase 5C; AE itself catches these at load time. Document but don't pre-detect.

---

## 3. Algorithm

```
algorithm: insertLayer(c, src, atIdx) → (*Layer, error):

  // 1. Validate (§2 R1–R11). Return early without mutation if any fail.

  // 2. Locate src's Layr in src.comp.back.itemList.Children.
  srcChildren := src.comp.back.itemList.Children
  srcLayrIdx  := findLayrIndexInItemList(src.comp.back.itemList, src.back.layrList)
  if srcLayrIdx < 0:  return nil, ErrSrcLayrNotFound

  // 3. Adaptive block end (mirror DuplicateLayer §4 / MoveLayer §4):
  //    scan leaf followers until next LIST/EOF.
  endIdx := srcLayrIdx + 2
  for endIdx < len(srcChildren) && !srcChildren[endIdx].IsList():
    endIdx++

  // 4. Snapshot dest state for rollback.
  oldDestChildren := slices.Clone(c.back.itemList.Children)
  oldDestLayers   := slices.Clone(c.Layers)
  oldNextItemID   := c.proj.nextItemID
  oldWarningsLen  := len(c.proj.Warnings)

  // 5. Deep-clone src block (deepCloneChunk recursive; fresh Data slices
  //    per concurrency-unsafe-shared-chunk-bytes scar).
  cloneBlock := make([]*rifx.Chunk, endIdx-srcLayrIdx)
  for k := srcLayrIdx; k < endIdx; k++:
    cloneBlock[k-srcLayrIdx] = deepCloneChunk(srcChildren[k])

  // 6. Mutate clone's ldta bytes — the per-byte delta vs DuplicateLayer.
  newID := c.proj.allocItemID()
  clonedLayr := cloneBlock[0]
  clonedLdta := clonedLayr.FindFirst(rifx.IDLdta)
  if clonedLdta == nil || len(clonedLdta.Data) < 0xA4:
    c.proj.nextItemID = oldNextItemID
    return nil, ErrLdtaTooShort

  binary.BigEndian.PutUint32(clonedLdta.Data[0x00:0x04], newID) // layer ID
  binary.BigEndian.PutUint32(clonedLdta.Data[0x84:0x88], 0)     // ParentID reset
  clonedLdta.Data[0x6B] = byte(TrackMatteNone)                  // matte mode reset

  // Explicit matte ID byte exists only on AE 23+ ldta sizes (164 bytes).
  // Phase 5B confirmed Go-emitted layers may be either 160 or 164. Guard
  // with length check — when 160 there is no @0xA0 to write.
  if len(clonedLdta.Data) >= 0xA4:
    binary.BigEndian.PutUint32(clonedLdta.Data[0xA0:0xA4], 0)   // TrackMatteLayerID reset

  // (Name NOT rewritten — verbatim from src per implicit-name decision.)

  // 7. Compute dest splice index in c.back.itemList.Children.
  destChildren := c.back.itemList.Children
  var insertChunkIdx int
  switch {
  case len(c.Layers) == 0:
    // Empty dest comp — insert at end (matches NewComposition layout
    // where layers append to itemList after any leading non-Layr
    // siblings). Walk forward past any non-Layr/Ewst LISTs to find
    // the first available layer-block slot.
    insertChunkIdx = findFirstLayerSlot(destChildren)
  case atIdx < len(c.Layers):
    // Insert BEFORE the Layr block of c.Layers[atIdx].
    target := c.Layers[atIdx]
    insertChunkIdx = indexOfChunk(destChildren, target.back.layrList)
    if insertChunkIdx < 0:  return rollback, ErrTargetLayrNotFound
  default:
    // atIdx == len(c.Layers): append after the last layer's block.
    last := c.Layers[len(c.Layers)-1]
    lastLayrIdx := indexOfChunk(destChildren, last.back.layrList)
    if lastLayrIdx < 0:  return rollback, ErrLastLayrNotFound
    insertChunkIdx = lastLayrIdx + 2
    for insertChunkIdx < len(destChildren) && !destChildren[insertChunkIdx].IsList():
      insertChunkIdx++
  }

  // 8. Splice cloneBlock into dest itemList at insertChunkIdx.
  newDestChildren := make([]*rifx.Chunk, 0, len(destChildren)+len(cloneBlock))
  newDestChildren = append(newDestChildren, destChildren[:insertChunkIdx]...)
  newDestChildren = append(newDestChildren, cloneBlock...)
  newDestChildren = append(newDestChildren, destChildren[insertChunkIdx:]...)
  c.back.itemList.Children = newDestChildren

  // 9. Re-parse cloned Layr to build a fresh *Layer with backrefs into
  //    the cloned chunks (mirror DuplicateLayer step 10).
  var localWarnings []string
  ctx := newParseCtxFPS(c.TickRate, c.FrameRate, c.Name, &localWarnings)
  cloneLayer, parseErr := parseLayer(clonedLayr, atIdx, ctx)
  if parseErr != nil:
    c.back.itemList.Children = oldDestChildren
    c.proj.nextItemID = oldNextItemID
    return nil, ErrReparse(parseErr)
  cloneLayer.comp = c
  assignTransformDefaults(cloneLayer.Properties, c, cloneLayer.Type)

  // 10. Insert cloneLayer into c.Layers at atIdx.
  newLayers := make([]*Layer, 0, len(c.Layers)+1)
  newLayers = append(newLayers, c.Layers[:atIdx]...)
  newLayers = append(newLayers, cloneLayer)
  newLayers = append(newLayers, c.Layers[atIdx:]...)
  c.Layers = newLayers

  // 11. Warnings-as-failure (Inv-11): if c.proj.Warnings grew, restore
  //     all snapshots (incl. nextItemID) and return error.
  if len(localWarnings) > 0:
    c.proj.Warnings = append(c.proj.Warnings, localWarnings...)
  if len(c.proj.Warnings) > oldWarningsLen:
    c.back.itemList.Children = oldDestChildren
    c.Layers = oldDestLayers
    c.proj.nextItemID = oldNextItemID
    newWarnings := slices.Clone(c.proj.Warnings[oldWarningsLen:])
    c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
    return nil, ErrWarningsRollback(newWarnings)

  return cloneLayer, nil
```

**Per-byte mutation diff vs DuplicateLayer**:

| Byte | DuplicateLayer | InsertLayer |
|---|---|---|
| ldta @0x00..0x03 (layer ID) | overwrite with newID | overwrite with newID |
| ldta @0x6B (TrackMatte mode) | verbatim (or refuse implicit per Phase 5B) | force = TrackMatteNone |
| ldta @0x84..0x87 (ParentID) | verbatim | overwrite with 0 |
| ldta @0xA0..0xA3 (explicit matte ID, AE 23+) | verbatim | overwrite with 0 (guarded by len ≥ 0xA4) |
| ldta @0x28..0x2B (SourceID) | verbatim | verbatim |
| Utf8 name chunk | rewritten to caller's name | verbatim |
| all other ldta + follower body bytes | verbatim | verbatim |

The verbatim guarantee for non-mutated bytes preserves the F10 byte-identical clone-block invariant within the per-byte deltas above.

---

## 4. Reference / propagation pass — NONE outside the clone itself

InsertLayer does NOT walk `c.Layers` (dest) or `src.comp.Layers` (source) to rewrite anyone's refs:

- **Dest side**: existing layers in `c` don't know about the new clone; nothing to update.
- **Source side**: `src` itself stays put in `src.comp` — InsertLayer is a copy, not a move. No source-comp mutation.
- **Clone's own outgoing refs**: handled in-place via §3 step 6 byte mutations (ParentID/TrackMatteLayerID/TrackMatte mode reset; SourceID verbatim).

Mirrors DuplicateLayer's §4 ("no cleanup pass") — explicit so future-reader knows we considered and rejected propagation.

---

## 5. Atomic invariants reuse

Standard V2.1 pattern (NewShapeLayer / DeleteLayer / DuplicateLayer reference impl):

```
1. Snapshot pre-mutation dest state:
   - oldDestChildren := slices.Clone(c.back.itemList.Children)
   - oldDestLayers   := slices.Clone(c.Layers)
   - oldNextItemID   := c.proj.nextItemID                ← allocItemID bumps
   - oldWarningsLen  := len(c.proj.Warnings)

2. Validate (§2) — return early without mutation if any refuse-case fires

3. Mutate (§3 steps 5–10):
   - allocItemID (bumps nextItemID)
   - deep-clone block + ldta byte mutations
   - splice into c.back.itemList.Children at insertChunkIdx
   - parseLayer + insert into c.Layers

4. Warnings-as-failure (Inv-11): if len(c.proj.Warnings) > oldWarningsLen → rollback:
   - c.back.itemList.Children = oldDestChildren
   - c.Layers = oldDestLayers
   - c.proj.nextItemID = oldNextItemID
   - c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
   - return nil, error with new warnings appended

5. Return cloneLayer, nil
```

**Source comp**: never mutated (we only read its chunk bytes into `deepCloneChunk`). No source-side snapshot needed. Multi-comp rollback complexity is therefore zero — Phase 5C's "two comps involved" turns out to be one-sided write.

**Difference vs DuplicateLayer**: identical snapshot set; the byte-mutation step writes 3 extra ldta offsets (matte mode, ParentID, explicit matte ID).

---

## 6. Ship-gate plan (Plan Task ahead)

3 modes × 2 AE versions = **6 PASS target** to earn Stable promotion (mirrors DuplicateLayer Phase 3's 3-mode gate, MoveLayer Phase 4's 3-mode gate).

JSX fixture `re_insert_layer.jsx` produces three before/after pairs:

| Mode | Fixture | Setup | Insert call |
|---|---|---|---|
| `insert_basic` | `re_insert_layer_basic_{before,after}.aep` | compA: 1 solo AV layer "Src" (no parent/matte) over Solid; compB: 1 unrelated layer | `compB.InsertLayer(compA.Layers[0], 0)` — clone lands at top of compB |
| `insert_with_footage` | `re_insert_layer_footage_{before,after}.aep` | compA: 1 AV layer over Footage F1 (file solid); compB: unrelated layer + same Footage F1 used by other comp already | `compB.InsertLayer(compA.Layers[0], 0)` — assert clone also refs F1 (shared) |
| `insert_with_precomp` | `re_insert_layer_precomp_{before,after}.aep` | compA: 1 AV layer whose source is precomp compC; compB: unrelated layer + compC also exists | `compB.InsertLayer(compA.Layers[0], 0)` — assert clone refs compC |

**Refuse-tests (Go-only, no AE round-trip)**: cross-Project (synthesize 2 separate `*Project`), same-comp, src.SourceID == c.ID, atIdx out of range, src nil, src.comp == nil, non-AV (camera layer in src).

**Ship-gate flow** (per `flightdeck/incident-reports/ae25-acceptance-gate.md` + `checklists/re-fixture.md`):

1. User runs `re_insert_layer.jsx` against AE 2020 + AE 2025 → produces `re_*` baseline (both versions ideally byte-identical at chunk-shape level).
2. Go-side: open `re_*_before.aep`, call `compB.InsertLayer(srcLayer, 0)`, `WriteAEP` → `ge_insert_layer_<mode>.aep`.
3. `scripts/ae_run.ps1` opens each `ge_*` in both AE versions, asserts no "file data missing" / silent-drop dialog, saves as `verify_*.aep`, byte-diffs against `re_*_after.aep` baseline (chunk-shape compare; AE timestamp/UUID jitter trimmed).
4. 6/6 PASS → promote `InsertLayer` from Alpha to Stable godoc tag. Update `flight-plans/coverage.md` § structural mutation API.

**Likely failure modes** (Stage 4 bisection candidates, per ship-gate scar):
- Cross-comp follower chunks may carry layer-ID-derived state not zeroed by §3 step 6 → bisect by zeroing fewer fields, see which AE rejects when wrong.
- `src.comp.back.itemList` chunk-encoded references back at src's ID could mean cloned followers contain stale ID — RE check first follower chunk byte-by-byte if a mode fails.
- AE 2025 explicit matte gymnastics (per Phase 5A/5B incident) may interact unexpectedly with the @0xA0 reset on a 164-byte ldta; double-check with a `re_insert_layer_basic` variant where src has explicit matte cleared in source comp before duplicate.

---

## 7. Implementation file map

**Create**:
- `internal/aep/insert_layer.go` — `InsertLayer` + local helper `findFirstLayerSlot(children []*rifx.Chunk) int` (empty-dest case; walks to first slot where a Layr block can go).
- `internal/aep/insert_layer_test.go` — refuse-case matrix + happy-path × 3 modes + round-trip byte-stability + concurrent-mutate safety + post-call assertions (clone.ParentID == 0, clone.TrackMatteLayerID == 0, clone.SourceID == src.SourceID, clone.Name == src.Name, c.proj.nextItemID bumped).

**Modify**: none planned — `deepCloneChunk`, `findLayrIndexInItemList`, `indexOfChunk`, `chunkIDString`, `parseLayer`, `newParseCtxFPS`, `assignTransformDefaults` all already exist intra-package.

**Fixtures (require user JSX run)**:
- `test_data/项目/re_insert_layer_basic_{before,after}.aep`
- `test_data/项目/re_insert_layer_footage_{before,after}.aep`
- `test_data/项目/re_insert_layer_precomp_{before,after}.aep`

**Doc updates**:
- `flightdeck/flight-plans/coverage.md` — add InsertLayer row under structural mutation API.
- godoc / `internal/aep/doc.go` (if exists for InsertLayer category).

---

## 8. Open / deferred to Phase 5C.1+

Recorded so future-Phase-N can pick them up without re-RE:

- **Cross-Project InsertLayer** (`src.comp.proj != c.proj`) — needs Footage/Comp deep-clone strategy. Likely couples with `Project.DuplicateItem` (the next picklist candidate after 5C). Could land as either (a) refuse cross-Project entirely + ask user to call `DuplicateItem` first then InsertLayer same-Project, or (b) inline deep-clone of all reachable items. Lean (a) for clean separation.
- **Same-comp via InsertLayer alias** — if demand surfaces, internal redirect to `DuplicateLayer` with auto-derived `atIdx`. Trivial wrapper; not worth ship-gate retrigger.
- **Shape/Text/Camera/Light insert** — same RE prerequisite as DuplicateLayer's deferred list (`landed/specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md` §7). Lift the `Type != LayerTypeAV` refuse after those archetypes ship-gate.
- **String-level ref rewrite** — expressions referencing src's ID don't auto-rewrite. AE itself doesn't either; out of scope.
- **Auto-rename on name collision** — AE allows duplicate layer names; we mirror.
- **Multi-select batch insert** — `InsertLayers(srcs []*Layer, atIdx int)` wrapper. Trivial but call-order matters (insert-then-shift). Phase 5C.x if demand surfaces.
- **Layer-receiver convenience** — `(l *Layer) CopyTo(dest *Composition, atIdx int) (*Layer, error)` wrapping `dest.InsertLayer(l, atIdx)`. Mirrors `Layer.MoveAfter/MoveBefore` ergonomics added in Phase 4 follow-up. Add post-Stable if API ergonomics surface as a pain point.

---

## 9. Related

- `landed/specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md` — base same-comp clone strategy (this doc deltas it).
- `landed/specs/2026-05-28-v3-phase2-deletelayer-strategy.md` — adaptive splice pattern source of truth.
- `specs/2026-05-27-v3-deep-think.md` § 4 — picklist origin.
- `incident-reports/nextitemid-must-include-layer-ids.md` — `allocItemID` correctness invariant InsertLayer inherits via `c.proj.allocItemID()`.
- `incident-reports/ae-duplicatelayer-re.md` — F10 (byte-identical block) is the per-byte invariant InsertLayer extends with 3 additional zero-writes.
- `incident-reports/ae25-acceptance-gate.md` — ship-gate playbook InsertLayer follows verbatim.
- `incident-reports/concurrency-unsafe-shared-chunk-bytes.md` — why `deepCloneChunk` must use fresh Data slices.
- `CLAUDE.md` § 硬约束 #1 / #2 / #5 / #6 — length-preserving default (length-variable name path NOT exercised since Name is verbatim), public API discipline, opaque preservation (cloned followers preserved byte-identical), AE acceptance gate.
