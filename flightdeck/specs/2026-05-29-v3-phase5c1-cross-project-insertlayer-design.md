# V3 Phase 5C.1 Cross-Project InsertLayer — Design Strategy

**Status**: design draft — pending user review → plan + implementation
**Builds on**: `landed/specs/2026-05-29-v3-phase5c-insertlayer-design.md` (same-Project sibling-comp clone, Stable post-gate — this doc lifts its R7 cross-Project refuse), `landed/specs/2026-05-29-v3-phase5d-duplicatecomposition-design.md` (intra-comp layer-ID remap two-pass — the *pattern* this reuses, the *file* it does NOT touch).
**Picklist origin**: 5C design §8 "Open / deferred to Phase 5C.1+" — cross-Project InsertLayer. Cockpit Next-session candidate #2 ("现已解锁").
**Unblocks**: a future generic cross-Project import / `Project.ImportComposition`; nothing depends on this yet.

This doc is the strategy matrix. Once approved → plan in `flight-plans/2026-05-29-v3-phase5c1-cross-project-insertlayer-plan.md`; implementation executes from this matrix without re-deriving decisions.

---

## 0. Key realization — this is NOT "DuplicateComposition into dest"

5C design §8 lean(a) sketched cross-Project as "先 DuplicateComposition 把源 comp 拷进 dest Project,再 same-Project InsertLayer". That sketch is **wrong / unbuildable** for two reasons, and this section supersedes it:

1. **Cross-Project has no AE scripting equivalent.** AE is single-project (`app.project` is singular; opening a project closes the current one). There is no API to copy a layer/comp between two open projects. Consequences:
   - `DuplicateComposition` is **same-Project** (clones a comp as a sibling in the *same* `p.rootFold`); it cannot place a comp into a *different* Project. §8 lean(a) step 1 does not exist.
   - **Ship-gate cannot byte-diff against an AE-produced baseline** (no AE op produces the "after"). The gate becomes **assert-based acceptance**: AE opens the Go-emitted file without silent-drop / missing-footage, and a verify JSX asserts the resulting semantics. See §6.

2. **The real new machinery = import the source's item closure into dest + remap `SourceID`.** Same-Project InsertLayer keeps `SourceID @0x28` **verbatim** (the referenced Footage/Comp item lives in the shared Project). Cross-Project, that item lives in the *source* Project; its ID is meaningless in dest. So we must deep-clone the layer's reachable item closure (footage / precomp, transitively) into dest's `rootFold` with fresh dest item IDs, then remap the inserted layer's `SourceID @0x28` (+ `AlternateSourceID` blsi) — and every imported precomp's internal layers' source refs — through a global `srcItemID → destItemID` map.

3. **"footage 无 scripting API" ≠ "我们搬不动它".** AE exposes no `FootageItem.duplicate()`, but we fully parse footage Item chunks; deep-cloning a footage Item LIST into dest `rootFold` with a fresh idta ID is pure file manipulation. The only open question is whether AE *accepts* the result — exactly what the ship-gate answers.

**Decisions locked in brainstorming** (user-approved 2026-05-29):
- **Full closure clone** — always bring the layer's whole reachable item closure (footage + precomp + nested), no "comp-source only" / "footage-source only" carve-out.
- **Footage cross-call dedup by file path** — a file-backed footage whose `Path` equals an existing dest footage's `Path` is **reused** (remap to the existing dest ID), not re-cloned. Comps are always cloned fresh. Solids / placeholders (no path) are always cloned fresh.
- **Lift R7, fold into existing `InsertLayer`** — no new public method; `c.InsertLayer(src, atIdx)` branches internally on `src.comp.proj != c.proj`. Closure import is an internal primitive (not public).
- **Trade-offs** (benefit-vs-complexity per user principle):
  - Do NOT refactor `duplicate_composition.go` (comp-clone reuse would need awkward parameterization + re-ship-gate of Stable 5D). Reuse its *pattern*, not its code.
  - DO extract a shared **layer-clone core** used by both same- and cross-Project paths (clean parameterization on SourceID handling; single source of truth for the byte-splice). Re-run same-Project InsertLayer ship-gate as a regression.
  - Folders are NOT imported; imported items land at dest root level.
  - Dangling source (a closure ref `AVItemByID` can't resolve in srcProj) → refuse, not silent-skip.

---

## 1. Public API — no signature change

`InsertLayer`'s signature, return type, and same-Project behavior are unchanged (it is Stable). The only change is **lifting the R7 cross-Project refuse** and adding a cross-Project branch:

```go
// InsertLayer deep-clones src into c.Layers at atIdx (0-based; == len appends).
//
// Same-Project (src.comp.proj == c.proj): unchanged from Phase 5C — clone one
// layer block, SourceID verbatim (shared item), ParentID/matte reset.
//
// Cross-Project (src.comp.proj != c.proj; Phase 5C.1): additionally imports the
// source's reachable ITEM CLOSURE (footage + precomp, transitively) into c's
// Project at root level, assigning fresh dest item IDs, then remaps the inserted
// layer's SourceID @0x28 + AlternateSourceID through the srcItemID→destItemID map.
// File-backed footage already present in dest (matched by Path) is reused, not
// re-cloned; comps and solids/placeholders are always cloned fresh. ParentID /
// track-matte still reset to 0 (cross-comp). Folders are not imported.
//
// Returns the inserted clone *Layer.
func (c *Composition) InsertLayer(src *Layer, atIdx int) (*Layer, error)
```

**No new public surface.** The closure importer is an unexported `c.insertLayerCrossProject` + helpers in a new file. A future `Project.ImportComposition(from *Project, src *Composition)` can be built on the same internals if demand surfaces (deferred §8).

**Stable discipline (CLAUDE.md #2)**: signature/type/JSON unchanged. The cross-Project path is a *new* structural write path → Alpha-tagged in godoc until its ship-gate is green (§6); same-Project path stays Stable. The shared layer-clone core extraction re-runs the same-Project ship-gate as regression (§6).

---

## 2. Refuse-case matrix

Same-Project refuses (R1–R6, R8–R11 from 5C §2) are unchanged. **R7 (cross-Project) is removed** and replaced by cross-Project-specific refuses:

| # | Detection | Reason |
|---|---|---|
| (R1–R6, R8–R11) | per 5C §2 | nil src / dest backref / atIdx range / src detached / same-comp redirect / non-AV / direct pre-comp loop / src backref / structural corruption — **apply to both paths** |
| X1 | `c.proj.back == nil \|\| c.proj.back.rootFold == nil` | dest Project built outside parser — no rootFold to import items into |
| X2 | `src.comp.proj == nil` | src layer's Project unknown — can't walk its rootFold for the closure |
| X3 | closure walk hits a nonzero `SourceID`/`AlternateSourceID` that `srcProj.AVItemByID` can't resolve | dangling source — refuse rather than import a broken ref AE would silent-drop |
| X4 | an imported item's Item LIST not locatable in `srcProj.rootFold` (by idta ID match) | corruption defense — can't clone what we can't find |
| X5 | a footage being cloned has no idta or idta too short for the ID write | corruption defense (mirror DuplicateComposition R-cases) |

**R9 (direct pre-comp loop `src.SourceID == c.ID`) is moot cross-Project** — `src.SourceID` names a srcProj item, `c.ID` a destProj item; different namespaces, can't collide. Keep R9 active for the same-Project path only (it's checked before the branch, harmless cross-Project since IDs won't match).

**Not refuse-cases** (deliberately):
- Indirect precomp cycles within the source closure (compA→compB→compA) — the `itemIDMap` visited-set + two-pass (import all, then remap) handles cycles correctly; AE itself catches genuinely invalid cycles at load (caught by the warnings-rollback gate).
- Footage path collision (dedup) — that's the intended reuse, not an error.
- Name collisions between imported items and existing dest items — AE allows; we mirror (no rename).

---

## 3. Algorithm

```
algorithm: InsertLayer(c, src, atIdx) → (*Layer, error):

  // ── Shared validation (R1–R6, R8–R11) runs first, before the branch. ──

  if src.comp.proj == c.proj:
     return sameProjectInsert(c, src, atIdx)   // 5C body, unchanged (via shared core)

  // ════════ CROSS-PROJECT BRANCH (Phase 5C.1) ════════
  return insertLayerCrossProject(c, src, atIdx)


function insertLayerCrossProject(c, src, atIdx):
  destProj := c.proj
  srcProj  := src.comp.proj

  1. Validate X1–X2.

  2. ── Snapshot dest for rollback (single-Project write; srcProj read-only) ──
     oldRootChildren := clone(destProj.back.rootFold.Children)
     oldComps        := clone(destProj.Compositions)
     oldFootage      := clone(destProj.Footage)
     oldDestItemList := clone(c.back.itemList.Children)
     oldDestLayers   := clone(c.Layers)
     oldNextItemID   := destProj.nextItemID
     oldWarningsLen  := len(destProj.Warnings)
     rollback() restores ALL seven on any failure below.

  3. ── PHASE 1: import the source item closure ──
     itemIDMap := map[uint32]uint32{}              // srcItemID → destItemID
     clonedCompLdtas := [][]*rifx.Chunk{}          // per imported comp: its layer ldtas (for Pass 2 source remap)
     worklist := nonzero({src.SourceID, src.AlternateSourceID})

     // Pass 1: import every reachable item; assign dest IDs; defer source remap.
     for len(worklist) > 0:
        srcID := pop(worklist)
        if srcID == 0 || srcID in itemIDMap:  continue
        item := srcProj.AVItemByID(srcID)
        if item == nil:  rollback(); return X3 dangling error

        switch item := item.(type):
        case *Footage:
           if isFileBacked(item) && item.Path != "":
              if existing := destFootageByPath(destProj, item.Path); existing != nil:
                 itemIDMap[srcID] = existing.ID        // DEDUP HIT — reuse, no clone
                 continue
           // clone footage Item block into dest rootFold
           destID := importItemBlock(destProj, srcProj, srcID)   // deepClone + idta@0x10 rewrite + splice + parseFootage
           itemIDMap[srcID] = destID
           // footage has no outgoing item refs → no recursion

        case *Composition:
           destID := destProj.allocItemID()
           block  := locateItemBlock(srcProj.rootFold, srcID)    // Item LIST + trailing non-Item run
           dup    := deepClone(block); rewrite dup idta@0x10 = destID
           // intra-comp layer-ID remap (DuplicateComposition pattern, reused not shared):
           ldtas  := remapLayerIDsIntraComp(dup, destProj)       // alloc fresh per layer; rewrite @0x00, ParentID@0x84, matte@0xA0
           clonedCompLdtas.append(ldtas)
           // collect this comp's layers' source refs for the closure
           for ldta in ldtas:
              worklist.push(U32(ldta[0x28]))                     // SourceID
              worklist.push(altSourceIDFromBlsi(ldta's Layr))    // AlternateSourceID, if any
           spliceItemBlockIntoRootFold(destProj, dup-with-siblings)
           itemIDMap[srcID] = destID

     // Pass 2: remap every imported comp's layer source refs through itemIDMap.
     for ldtas in clonedCompLdtas:
        for ldta in ldtas:
           if sid := U32(ldta[0x28]); sid != 0 && itemIDMap has sid:
              PutU32(ldta[0x28], itemIDMap[sid])
           if blsi present && altID != 0 && itemIDMap has altID:
              PutU32(blsi[0:4], itemIDMap[altID])
        // a source ref NOT in itemIDMap means it was unreachable from src's own
        // refs but present on an imported comp's layer — shouldn't happen (we
        // pushed all of them in Pass 1). If it does, leave verbatim → the reparse
        // warning path / AE acceptance gate catches it.

  4. ── Reparse imported items into dest model ──
     (importItemBlock already parseFootage'd footage; comps parseComposition'd here
      after Pass 2 so the parsed model sees final source IDs)
     For each imported comp: parseComposition(dup, destID, name, &destProj.Warnings);
        on err → rollback; append to destProj.Compositions; wire .proj = destProj.
     (Order note: parse comps AFTER Pass 2 so SourceComposition()/SourceFootage()
      resolve against final dest IDs.)

  5. ── PHASE 2: insert the layer via the shared layer-clone core ──
     sourceRemap := func(srcSourceID uint32) uint32 {
        if mapped, ok := itemIDMap[srcSourceID]; ok { return mapped }
        return srcSourceID   // 0 stays 0; unmapped shouldn't occur for src's own refs
     }
     cloneLayer, err := spliceLayerClone(c, src, atIdx, sourceRemap)
        // shared core: deepClone src layer block, alloc new layer ID @0x00,
        // reset ParentID@0x84 / matte mode@0x6B / explicit matte@0xA0,
        // REMAP SourceID@0x28 = sourceRemap(src.SourceID) and AlternateSourceID,
        // splice into c.back.itemList, parseLayer, insert into c.Layers.
     on err → rollback; return err

  6. ── Warnings-as-failure (Inv-11) ──
     if len(destProj.Warnings) > oldWarningsLen:
        rollback(); return error with the new warnings.

  return cloneLayer, nil
```

**Shared layer-clone core `spliceLayerClone(c, src, atIdx, sourceRemap)`**: extracted from the current 5C `InsertLayer` body. The same-Project path passes `sourceRemap = identity` (SourceID verbatim); the cross-Project path passes the itemIDMap lookup. This is the only behavioral knob; every other byte mutation (new layer ID, ParentID/matte reset) is identical. Extraction keeps the delicate adaptive-splice logic in one place.

**Per-byte diff vs same-Project InsertLayer**: identical except `SourceID @0x28` and `AlternateSourceID` (blsi) are remapped via itemIDMap instead of verbatim. Everything in Phase 1 (closure import) is net-new.

---

## 4. Reference / propagation — item-ID remap is the new pass

Three rewrite scopes, all confined to cloned bytes (src Project never mutated):
- **Inserted layer** (in dest comp `c`): SourceID/AltSourceID remapped via itemIDMap; ParentID/matte reset 0.
- **Imported comps' internal layers**: layer IDs remapped intra-comp (Pass 1); SourceID/AltSourceID remapped via itemIDMap (Pass 2); ParentID/matte **preserved+remapped** intra-comp (these refs stay meaningful — same comp, copied — exactly like DuplicateComposition, unlike the inserted layer which is orphaned into a foreign comp).
- **Imported footage**: idta item ID rewritten; no outgoing refs.

No existing dest comp/layer/item is rewritten. Dedup'd footage is referenced by its existing dest ID, untouched.

---

## 5. Atomic invariants

Single-Project (dest) write, wider snapshot than same-Project InsertLayer because Phase 1 mutates `rootFold.Children` + `Compositions` + `Footage` + bumps `nextItemID` once per imported item/layer. Seven-way snapshot (§3 step 2) + warnings-as-failure rollback. srcProj is read-only → no source-side snapshot. The `nextItemID` rollback restores the pre-call value wholesale (covers every closure + layer alloc).

Failure points that trigger rollback: dangling source (X3), unlocatable item block (X4), short footage idta (X5), any `parseComposition`/`parseLayer` error, any new parser warning.

---

## 6. Ship-gate plan — assert-based acceptance (NOT byte-diff)

No AE op produces a cross-Project "after" baseline (§0.1). The gate confirms **AE accepts the Go-emitted file and the semantics are correct**, via verify-JSX assertions.

**Fixtures** — TWO AE-native source projects per mode (no cross-project op in AE):
- `re_xproj_src_<mode>.aep` — source: a comp with the layer(s) to copy (+ its footage / precomp closure).
- `re_xproj_dest_<mode>.aep` — dest: a comp to receive the clone (+ for the dedup mode, a footage already sharing the source's file path).

**ge emitter** `tmp_debug/ge_cross_project_insert/main.go`: open both `re_*` files, `destComp.InsertLayer(srcLayer, 0)`, `WriteAEP` → `ge_cross_project_insert_<mode>.aep`. Go-side asserts before write: inserted `clone.SourceID` resolves in destProj; imported item count matches the expected closure size; for dedup mode the dest footage count did NOT grow.

**verify JSX** `verify_ge_cross_project_insert.jsx` (per-version `.done` tag — re-fixture.md gotcha): AE opens each `ge_*`, asserts: target comp has the new layer; `layer.source` is non-null and `.name` matches the expected imported item; `layer.source` has no `footageMissing`; `app.project.numItems` grew by the expected closure size; dedup mode asserts the shared footage was reused (item count grew by closure-minus-1).

**Matrix** (3 modes × 2 AE versions = **6 PASS** to promote Alpha→Stable):
| Mode | Closure | Assertion focus |
|---|---|---|
| `xproj_footage` | layer → 1 file footage | footage Item cloned into dest, SourceID remapped, source resolves |
| `xproj_precomp` | layer → precomp → nested footage | transitive closure imported, intra-comp parent remap survives, all sources resolve |
| `xproj_dedup` | layer → footage whose Path already exists in dest | footage reused (NOT duplicated), SourceID remapped to existing dest item |

**Regression** (shared layer-clone core extraction touches the Stable same-Project path): re-run the existing `re_insert_layer` 3-mode × 2-version gate; all 6 must stay PASS. Reuse `tmp_debug/ge_insert_layer` + `re_insert_layer.jsx` (already authored).

**Refuse-tests (Go-only)**: X1 (dest no rootFold), X2 (src no proj), X3 (dangling source — synthesize a src layer with a SourceID pointing at nothing), plus the inherited same-path refuses. No AE round-trip.

**Bisection candidates** (if AE rejects, per ship-gate scar): (a) imported footage Item LIST may need its trailing-sibling run / Fold placement AE expects — RE the first rejected mode's chunk shape vs an AE-imported footage; (b) layer-ID vs item-ID namespace — confirm `allocItemID` monotonicity holds across the mixed comp/footage/layer alloc sequence; (c) idta @0x10 is the only comp/footage ID location (parse reads @0x10, not @0x14 — new_composition.go note); (d) AlternateSourceID blsi remap — verify offset @0x00 of the blsi chunk and that absent blsi is handled.

---

## 7. File map

**Create**:
- `internal/aep/import_closure.go` — `insertLayerCrossProject` + closure helpers: `importItemBlock`, `locateItemBlock`(by idta-ID match in rootFold, returns Item LIST + trailing run), `remapLayerIDsIntraComp` (DuplicateComposition pattern, local copy), `destFootageByPath`, `isFileBacked`.
- `internal/aep/import_closure_test.go` — refuse matrix X1–X5 + happy-path × 3 modes (footage / precomp / dedup) Go-side assertions (fresh dest IDs, SourceID remap, intra-comp parent remap in imported precomp, dedup reuse, nextItemID bump, round-trip byte-stability, concurrent-mutate fresh-slice).

**Modify**:
- `internal/aep/insert_layer.go` — remove R7; add the `src.comp.proj == c.proj` branch; extract the layer-clone body into `spliceLayerClone(c, src, atIdx, sourceRemap)` (same-Project calls it with identity remap). Same-Project bytes must stay identical (regression gate).
- `flightdeck/flight-plans/coverage.md` — extend the InsertLayer row to note cross-Project; add closure-import capability.
- godoc on `InsertLayer` — document the cross-Project semantics; Alpha note on the cross-Project path until gate green.

**Do NOT modify**: `duplicate_composition.go` (5D Stable — pattern reused, code not shared).

**Fixtures (gitignored, JSX-regenerated)**: `test_data/re_xproj_src_{footage,precomp,dedup}.aep`, `test_data/re_xproj_dest_{footage,precomp,dedup}.aep`, `test_data/ge_cross_project_insert_*.aep`.

---

## 8. Deferred

- **`Project.ImportComposition(from *Project, src *Composition)`** — public comp-level cross-Project import, built on the same closure internals. Add if demand surfaces; the internals from this phase make it a thin wrapper + its own ship-gate.
- **Folder structure import** — bring the source's folder hierarchy + item placement instead of flattening to dest root. Needs Folder back-chunk refs + folder-membership RE (Folder is currently ID+Name only). Cosmetic; deferred.
- **Cross-Project for non-AV layers** — same RE prerequisite as same-Project (Shape/Text/Camera/Light deferred list).
- **Dedup beyond footage path** — comp dedup by name+structure / footage dedup for solids. Fragile identity; explicitly out of scope (user chose path-only footage dedup).
- **Batch cross-Project insert** — `InsertLayers([]*Layer, atIdx)` sharing one closure import (avoids re-importing shared sources per layer). Trivial wrapper if demand surfaces.

---

## 9. Related

- `landed/specs/2026-05-29-v3-phase5c-insertlayer-design.md` — same-Project base this lifts R7 from; §8 lean(a) superseded by §0 here.
- `landed/specs/2026-05-29-v3-phase5d-duplicatecomposition-design.md` — intra-comp layer-ID remap two-pass pattern reused (not shared).
- `incident-reports/nextitemid-must-include-layer-ids.md` — `allocItemID` correctness across the mixed comp/footage/layer alloc sequence.
- `incident-reports/concurrency-unsafe-shared-chunk-bytes.md` — `deepCloneChunk` fresh-Data-slice requirement (every cloned item/layer).
- `incident-reports/ae25-acceptance-gate.md` + `checklists/re-fixture.md` — ship-gate playbook; this phase uses its assert-based variant (per-version `.done` tag).
- `internal/aep/types_core.go` — `AVItemByID` / `CompositionByID` / `FootageByName` lookups + `Footage.Path` (dedup key) + `Layer.SourceID`/`AlternateSourceID`.
- `CLAUDE.md` § 硬约束 #1 / #2 / #5 / #6 — length-preserving (name path unused; SourceID remap is fixed-width @0x28), public API discipline (no signature change), opaque preservation (cloned items byte-identical except remapped IDs), AE acceptance gate.
</content>
</invoke>
