# V3 Phase 5D DuplicateComposition — Design Strategy

**Status**: design draft — pending user review → plan + implementation
**Builds on**: `2026-05-29-v3-phase5c-insertlayer-design.md` (cross-comp clone + 3 ldta byte resets + atomic splice), `landed/specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md` (adaptive deep-clone block), `internal/aep/new_composition.go` (rootFold splice + parseComposition reparse loop).
**Picklist origin**: Phase 5 三选一 candidate #1 (`Project.DuplicateItem`). This doc scopes it to **comp duplication** — the only item kind with a scripting-RE'able path (`CompItem.duplicate()`); FootageItem/FolderItem have no `.duplicate()` ScriptingAPI method (verified `charts/after-effects-scripting-guide/docs/item/`).
**Unblocks**: Phase 5C.1 cross-Project InsertLayer (per 5C design §8, cross-Project should DuplicateItem the source into dest Project, then same-Project InsertLayer — not inline-clone).

---

## 0. RE findings (AE 2020 17.7x45, 2026-05-29)

Probe `test_data/re_duplicate_item.jsx` built `compA_main` (3 AV layers: A_precomp→compC, A_footage→F1, A_solid with `A_solid.parent = A_precomp`), then `compA.duplicate()`. Parsed result (`tmp_debug/dump_layers` + `list_items` on `re_duplicate_item_after.aep`):

```
items:  COMP id=16 compA_main   COMP id=32 "compA_main 2"   COMP id=1 compC_precomp
        FOOTAGE id=28 A_solid    FOOTAGE id=14 compC_filler   FOLDER id=13 Solids

compA_main (id=16):              compA_main 2 (id=32):  ← DUP
  [0] id=31 A_precomp src=1        [0] id=36 A_precomp src=1     ← source SHARED (compC id=1)
  [1] id=30 A_footage src=28       [1] id=35 A_footage src=28    ← source SHARED (footage id=28)
  [2] id=29 ""       parent=31     [2] id=34 ""       parent=36  ← parent REMAPPED 31→36 (dup's own A_precomp)
                     src=28                            src=28
```

**Four conclusions** (the byte contract DuplicateComposition must reproduce):

1. **New comp item ID** — dup gets a fresh item ID (16→32) from the head counter (`idta @0x10`).
2. **Every layer gets a fresh layer ID** — 31/30/29 → 36/35/34, monotonic from the same head counter (`ldta @0x00`). (Comp 32, then 34/35/36; AE's intermediate 33 is internal — exact values don't matter, uniqueness + monotonicity do.)
3. **Intra-comp parent/matte refs are REMAPPED** — source `[2].parent=31` (original A_precomp) becomes dup `[2].parent=36` (the **dup's own** A_precomp). This is the crux and the only genuinely new machinery vs InsertLayer: build a `srcLayerID → dupLayerID` map and rewrite every intra-comp `ParentID @0x84` + `TrackMatteLayerID @0xA0`.
4. **Sources are SHARED, not cloned** — dup layers keep `source=1`(compC) / `source=28`(footage) verbatim. No footage/precomp item duplication. `SourceID @0x28` is verbatim (same as InsertLayer).

---

## 1. Public API

```go
// DuplicateComposition deep-clones src (a comp in this Project) as a new
// sibling comp named name, appended to p.Compositions. The dup contains a
// fresh copy of every layer (new layer IDs), with intra-comp parent +
// track-matte refs remapped to the dup's own layers; layer SOURCES
// (footage / precomp items) are shared verbatim, not duplicated — matching
// AE ScriptingAPI's CompItem.duplicate(). Returns the new *Composition.
//
// Clone semantics (Phase 5D; same-Project comp only):
//   - new comp item ID = p.allocItemID()             (idta @0x10)
//   - per layer: new layer ID = p.allocItemID()      (ldta @0x00)
//   - intra-comp ParentID @0x84 / TrackMatteLayerID @0xA0 remapped via
//     srcLayerID→dupLayerID map (guarded by ldta length for @0xA0)
//   - SourceID @0x28 verbatim (shared Footage/Comp items)
//   - comp name = caller-supplied (length-variable Utf8 rewrite)
//
// Refuse-cases (§2): nil src; src not in this Project; src missing
// itemList back-ref; empty name; corruption (non-Layr/Ewst block).
//
// Atomic mutation (Inv-10/Inv-11): snapshot rootFold.Children +
// p.Compositions + p.nextItemID + len(p.Warnings); roll all back on any
// new parser warning during reparse.
//
// Alpha until AE 2020 + AE 2025 ship-gate green (CLAUDE.md #2/#6).
func (p *Project) DuplicateComposition(src *Composition, name string) (*Composition, error)
```

**Naming**: `DuplicateComposition` (not the cockpit's aspirational `DuplicateItem(item Item, name)`) — honest about scope. No `Item` interface exists in the codebase (concrete `Composition`/`Footage`/`Folder`), and footage/folder have no scripting-RE path. A future generic `DuplicateItem` umbrella can wrap this once footage/folder duplication is RE'd (likely UI-automation-only). Mirrors `DuplicateLayer`'s explicit-name convention.

**name param**: required non-empty (mirror `DuplicateLayer`); caller supplies (AE auto-suffixes " 2" but our callers decide). Empty → R-refuse.

---

## 2. Refuse-case matrix

| # | Detection | Reason |
|---|---|---|
| R1 | `src == nil` | nil guard |
| R2 | `p.back == nil \|\| p.back.rootFold == nil` | no rootFold to splice into |
| R3 | `src.back == nil \|\| src.back.itemList == nil` | src built outside parser; can't deep-clone its block |
| R4 | `src.proj != p` (or src not found in `p.Compositions`) | src must belong to this Project |
| R5 | `name == ""` | mirror DuplicateLayer non-empty requirement |
| R6 | structural: src's itemList not found in `rootFold.Children`, or its required trailing sibling chunks absent | corruption defense (mirror NewComposition's 8-sibling contract) |
| R7 | per-layer: a Layr block's ldta shorter than `0x88` (ParentID write) | corruption defense |

**Not refuse-cases**: layer with ParentID/matte pointing outside the comp (shouldn't happen intra-comp; if a ref isn't in the remap map, leave verbatim + emit a warning → atomic rollback catches it). Name collision with an existing comp (AE allows; we mirror).

---

## 3. Algorithm

```
algorithm: duplicateComposition(p, src, name) → (*Composition, error):

  1. Validate R1–R6.

  2. Locate src's itemList + trailing siblings in rootFold.Children.
     srcItemIdx := indexOfChunk(rootFold.Children, src.back.itemList)
     // siblings = the run of non-LIST/Item chunks AE keeps after each comp
     // Item (NewComposition appends lowerItemSiblings(nil) == 8 chunks).
     // Clone the Item LIST + that trailing run as one block.

  3. Snapshot for rollback:
     oldRootChildren := slices.Clone(rootFold.Children)
     oldComps        := slices.Clone(p.Compositions)
     oldNextItemID   := p.nextItemID
     oldWarningsLen  := len(p.Warnings)

  4. Deep-clone the comp block (fresh Data slices — concurrency scar):
     dupItemList := deepCloneChunk(src.back.itemList)
     dupSiblings := [deepCloneChunk(s) for s in trailing siblings]

  5. New comp ID + idta rewrite:
     newCompID := p.allocItemID()
     dupIdta := dupItemList.FindFirst(rifx.IDidta)
     PutUint32(dupIdta.Data[idtaItemID:+4], newCompID)   // @0x10

  6. ── REMAP PASS (the new machinery) ──
     // Pass A: walk dupItemList's Layr children in order; for each,
     //         read old layer ID @0x00, alloc new, record map.
     idMap := map[uint32]uint32{}
     for each Layr block L in dupItemList.Children:
        ldta := L.FindFirst(IDLdta)
        oldID := U32(ldta.Data[0x00:])
        newID := p.allocItemID()
        idMap[oldID] = newID
        PutUint32(ldta.Data[0x00:+4], newID)             // layer ID
     // Pass B: rewrite intra-comp refs via idMap.
     for each Layr block L:
        ldta := L.FindFirst(IDLdta)
        parent := U32(ldta.Data[0x84:])
        if parent != 0 && idMap has parent:
           PutUint32(ldta.Data[0x84:+4], idMap[parent])  // ParentID remap
        if len(ldta.Data) >= 0xA4:
           matte := U32(ldta.Data[0xA0:])
           if matte != 0 && idMap has matte:
              PutUint32(ldta.Data[0xA0:+4], idMap[matte]) // matte remap
        // SourceID @0x28 left verbatim (shared items).
        // refs NOT in idMap (cross-comp / corruption) → leave + the
        // reparse warning path will roll back if AE-invalid.

  7. Name rewrite: replace the Utf8 name chunk in dupItemList with `name`
     (length-variable path — reuse the same Utf8 splice WriteAEP already
     handles; see InsertLayer Name-verbatim note inverted).

  8. Splice dup block into rootFold after src's sibling run:
     insertAt := srcItemIdx + 1 + len(trailing siblings)
     rootFold.Children = splice(oldRootChildren, insertAt, [dupItemList]+dupSiblings)

  9. Reparse closed loop (mirror NewComposition):
     dupComp, err := parseComposition(dupItemList, newCompID, name, &p.Warnings)
     on err → rollback (restore 4 snapshots) → return err
     dupComp.proj = p   // wire back-ref

  10. p.Compositions = append(p.Compositions, dupComp)

  11. Warnings-as-failure: if len(p.Warnings) > oldWarningsLen → restore all
      4 snapshots (incl. nextItemID) → return error w/ new warnings.

  return dupComp, nil
```

**Per-byte diff vs InsertLayer**: InsertLayer clones ONE layer block and *zeroes* parent/matte (orphan in a foreign comp). DuplicateComposition clones the WHOLE comp and *remaps* parent/matte to the dup's own new layer IDs (refs stay meaningful — same comp, copied). SourceID verbatim in both. New: comp idta @0x10 write + the two-pass idMap.

---

## 4. Reference / propagation — only INSIDE the dup

No mutation of src, existing comps, or shared source items. The only rewrites are within the cloned block (comp ID, layer IDs, intra-comp parent/matte). Shared footage/precomp items are untouched (refs verbatim). Mirrors InsertLayer §4 / DuplicateLayer §4.

---

## 5. Atomic invariants

Standard V2.1 quadruple snapshot (rootFold.Children / p.Compositions / p.nextItemID / Warnings len) + warnings-as-failure rollback. Single-Project, single-sided write — same rollback simplicity as InsertLayer (§5 there). `p.nextItemID` bumps once per layer + once for the comp; rollback restores the pre-call value wholesale.

---

## 6. Ship-gate plan

Baselines: `re_duplicate_item_{before,after}.aep` via `re_duplicate_item.jsx` (already authored; `after` captured, `before` needs a warm AE retry — `before` cold-start hit the splash-screen grace timeout, not a data modal; per logbook 2026-05-28 warm 2nd attempt clears it).

- **ge emitter** `tmp_debug/ge_duplicate_composition/main.go`: open `re_duplicate_item_before.aep`, `p.DuplicateComposition(compA_main, "compA_main 2")`, WriteAEP → `ge_duplicate_composition.aep`. Assert (Go-side) dup layer IDs all fresh + `dup.Layers[2].ParentID == dup.Layers[0].ID` (remap) + sources verbatim.
- **verify JSX** `verify_ge_duplicate_composition.jsx` (per-version `.done` tag — re-fixture.md gotcha): AE opens, finds "compA_main 2", asserts numLayers==3, the parented layer's `.parent` is the dup's own A_precomp (NOT the original comp's), sources shared (`.source.id` equals original's source ids).
- **Matrix**: 1 base mode × 2 AE versions = 2 PASS minimum. Add an AE23+ explicit-matte variant (intra-comp matte ref remap @0xA0) as a 2nd mode if time permits → 4 PASS. 6/6 not required (single scenario covers the remap crux); justify mode count in the plan.

**Bisection candidates** (if AE rejects): (a) layer-ID alloc order — AE assigns dup IDs bottom-to-top ascending; if acceptance fails, try matching that order vs parse order; (b) the idta @0x10 is the only comp-ID location (parse reads @0x10 not @0x14 — see new_composition.go note); (c) trailing-sibling count/shape must match NewComposition's `lowerItemSiblings`.

---

## 7. File map

**Create**: `internal/aep/duplicate_composition.go` (+ no new exported helpers — reuse `deepCloneChunk`, `indexOfChunk`, `parseComposition`, `allocItemID`, `chunkIDString`, the Utf8 name splice). `internal/aep/duplicate_composition_test.go` (refuse matrix R1–R7 + happy-path: fresh IDs, parent remap, source-shared, name, nextItemID bump, round-trip, concurrent-mutate fresh-slice).

**Modify**: `flightdeck/flight-plans/coverage.md` (+row), godoc Alpha→Stable post-gate.

**Fixtures (gitignored, JSX-regenerated)**: `test_data/re_duplicate_item_{before,after}.aep`, `ge_duplicate_composition.aep`.

---

## 8. Deferred

- **Generic `Project.DuplicateItem(Item, name)`** — needs an `Item` interface + footage/folder duplication (no scripting API → UI-automation RE or skip). Umbrella over DuplicateComposition once those land.
- **Footage duplicate** (AE UI "Edit > Duplicate" on footage → 2nd footage item, same source) — no `FootageItem.duplicate()` method; RE would need GDI UI automation.
- **Folder duplicate** — same (no API).
- **Deep-clone sources** (duplicate a comp AND its precomp/footage tree into fresh items) — explicitly NOT what AE does; only if a "deep duplicate" use-case surfaces.
- **Cross-Project** — DuplicateComposition is the building block; cross-Project = clone into a *different* p.rootFold + remap source item IDs too. Phase 5C.1 territory.

---

## 9. Related

- `2026-05-29-v3-phase5c-insertlayer-design.md` — sibling clone strategy this deltas (zero-refs → remap-refs).
- `incident-reports/nextitemid-must-include-layer-ids.md` — `allocItemID` correctness DuplicateComposition inherits (explicitly anticipates this feature).
- `internal/aep/new_composition.go` — rootFold splice + parseComposition reparse loop reused verbatim.
- `incident-reports/concurrency-unsafe-shared-chunk-bytes.md` — `deepCloneChunk` fresh-Data-slice requirement.
- `incident-reports/ae25-acceptance-gate.md` + `checklists/re-fixture.md` — ship-gate playbook (incl. per-version `.done` tag gotcha).
- RE evidence: `test_data/re_duplicate_item_after.aep` + `re_duplicate_item.done` (this session).
