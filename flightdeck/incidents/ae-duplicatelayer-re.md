---
status: active
when_to_read: implementing Composition.DuplicateLayer or any layer-clone mutation; understanding what AE's layer.duplicate() actually does at the chunk level; deciding whether to match AE's matte-preservation positioning quirk; tracking which fields get copied verbatim vs reset on clone
applies_to: [duplicate-layer, layer-clone, structural-mutation, itemList, Layr, Ewst, fvdv-followers, parent-clone, matte-preservation, id-allocation, ae2020, duplicatelayer-re]
last_updated: 2026-05-28
---

# DuplicateLayer RE — AE 行为契约

V3 Phase 3 DuplicateLayer 实现 reference。**数字来自 AE 2020 17.7x45 自己 save 的 4 个 fixture diff**。

Fixtures: `test_data/re_duplicate_layer_{solo,dup_parent,dup_child,dup_matted}.aep`（全 AE 2020）。JSX harness: `test_data/re_duplicate_layer.jsx`（4 mode via `$.getenv("RE_DUP_MODE")`）。

## Layout (constant across modes)

3 solid layers in single comp `DupTest`:

```
index 1 (top): L1_top  id=19   ← child of L2 in dup_child; implicit matte src in dup_matted
index 2 (mid): L2_mid  id=17   ← parent of L3 in dup_parent; matted target in dup_matted
index 3 (bot): L3_bot  id=15   ← child of L2 in dup_parent
```

Per-mode setup attaches one ref then calls `target.duplicate()` from JSX. Mode → target → expected post-state captured in `.done` log.

---

## Finding 1: New layer placement = source's OLD index (source pushed down by 1)

**3 of 4 modes** match: `solo`, `dup_parent`, `dup_child` all put the new layer at `oldIndexOf(source)`, pushing source + everything below down by one.

| Mode | Source | Source pre-idx | New layer post-idx | Source post-idx |
|---|---|---|---|---|
| solo | L2 | 2 | **2** | 3 |
| dup_parent | L2 | 2 | **2** | 3 |
| dup_child | L1 | 1 | **1** | 2 |
| dup_matted | L2 | 2 | **1** (!) | 3 |

The `dup_matted` exception is Finding 2.

**Impl rule (simple cases)**: insert new Layr at `oldIndexOf(source)` in `comp.Layers` (i.e. just before source). In itemList, locate source's Layr chunk → insert new `Layr + Ewst + 14 followers` block IMMEDIATELY BEFORE source's block.

---

## Finding 2: AE's matte-preservation quirk on dup_matted

`dup_matted` setup: `L2.trackMatteType = ALPHA` (implicit "layer above" = L1 at idx 1).

Pre-dup post-state: 4 layers `[new L2 @ idx 1, L1 @ idx 2, original L2 @ idx 3, L3 @ idx 4]`.

**AE jumped new L2 to idx 1 (above L1)** rather than inserting at idx 2 (source's old position). Why: AE preserves original L2's matte source = "layer above original" = L1. If new L2 had been placed at idx 2, then:
- new L2 at idx 2 — its matte source would be L1 at idx 1 ✓
- original L2 at idx 3 — its matte source would be **new L2** at idx 2 ✗ (broken; expected L1)

So AE moved new L2 above L1 to keep original's matte working: layout `[new L2, L1, original L2, L3]` → original L2's "layer above" stays L1 at idx 2.

New L2 itself: trackMatteType=5013 (copied), but now no layer above → matte intent inert. F6 (preserves intent byte) + breaks the actual matte → AE's compromise.

**Impl options for Phase 3**:

- **(a) Refuse**: reject DuplicateLayer when source has `Layer.TrackMatte != TrackMatteNone`. Force caller to clear matte intent first. **Recommended — simplest, easy to lift later.**
- **(b) Mirror AE quirk**: detect matte → place clone at "above matte source" position. Adds ~30 LOC of special-case logic for a niche scenario.
- **(c) Naive same-index insertion**: matches simple cases; breaks original's matte in this scenario silently (user discovers at render time). Worst UX.

Strategy spec § (TBD) picks (a) for Phase 3 MVP.

**Phase 5B amendment (2026-05-28)**: F2 quirk applies ONLY to **implicit** "layer-above" matte (positional source). For AE 23+ **explicit** matte (`TrackMatteLayerID != 0` in ldta `@0xA0`), the matte source is decoupled from layer order — duplicating is safe with verbatim byte-copy of `@0xA0` + `@0x6B`. DuplicateLayer refuse-case split: implicit (still refused per F2) vs explicit (allowed in Phase 5B). AE 2025 ship-gate via `ge_duplicate_layer_explicit_matte.aep` (uses `re_trackmatte_ae24.aep` as source). See [`plans/2026-05-28-v3-phase5b-duplicatelayer-explicit-matte-plan.md`](../archive/plans/2026-05-28-v3-phase5b-duplicatelayer-explicit-matte-plan.md).

---

## Finding 3: Layer ID allocation = head counter +1 (monotonic, Inv-9 holds)

Source IDs in baseline: L1=19, L2=17, L3=15. Source items (footage solids): 14 (L3), 16 (L2), 18 (L1). Max ID = 19.

**All 4 modes**: new layer's ID = 20 (= max + 1, head counter consumes next slot). Matches `NewShapeLayer`'s pattern via `proj.nextItemID` bump.

**Impl rule**: use `proj.allocItemID()` (or `proj.nextItemID++`) for the new layer's ID. Bump head counter. Same as NewShapeLayer.

---

## Finding 4: Source footage NOT cloned — clone shares Layer.SourceID

`solo` mode parse (post-dup):

```
[0] id=19  source=18    L1, source=L1's footage solid
[1] id=20  source=16    ← clone, source = SAME as original L2
[2] id=17  source=16    ← original L2, source=16
[3] id=15  source=14    L3, source=L3's footage solid
```

Both L2 instances point to **footage item id=16** (shared). AE doesn't duplicate the footage; the two layers reference the same source.

**Impl rule**: clone's `Layer.SourceID = source.SourceID`. Write byte-identical ldta @0x28..0x2B. Do NOT call any footage-duplication path.

---

## Finding 5: itemList growth = +16 chunks per duplicate (mirrors F4 from delete RE)

Baseline (3 layers) itemList: 237 children. All 4 dup fixtures: 253 children. Δ = **+16 per duplicate**, matching the "delete unit = 16 chunks" finding from DeleteLayer RE (`incidents/ae-deletelayer-re.md` F4).

The 16-chunk block = `Layr + Ewst + 14 followers (fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2 repeats)`.

**Impl rule**: cloning a layer = clone the entire 16-chunk block. Splice into itemList at the insertion position (Finding 1). Each chunk's `Data` slice MUST be a fresh `append([]byte(nil), src.Data...)` copy — sharing the slice violates `incidents/concurrency-unsafe-shared-chunk-bytes.md`.

**Open**: are the 14 followers byte-identical between source and clone, or does AE mutate some? Naive byte-clone assumption pending byte-diff verification in Task 2. Likely byte-identical (followers look like rendering knobs, not layer-bound IDs) — if not, Task 2 RE round-2 needed.

---

## Finding 6: Children's outgoing ParentID stays on ORIGINAL — clone is NOT promoted to parent

`dup_parent` mode: `L3.parent = L2` (L2 = id 17). Source = L2.

Post-dup parse:

```
[0] id=19  parent=0                     L1
[1] id=20  parent=0                     ← clone of L2 (NOT updated as new parent)
[2] id=17  parent=0                     original L2
[3] id=15  parent=17   matteSrc=0       L3 — parent still points to ORIGINAL L2 (id=17)
```

L3's `ParentID = 17` (original L2's ID). **AE did NOT update children to point at the clone.** Original L2 remains "the parent" from L3's perspective; clone is a sibling with the same shape/properties but no children attached.

**Impl rule**: when locating neighbors during DuplicateLayer, we do NOT need to update their refs. Source remains the canonical parent/matte-source for any incoming refs. Clone is a fresh shadow with zero incoming refs.

**Inverse case** (Finding 7) confirms this: clone's OUTGOING refs DO copy.

---

## Finding 7: Source's outgoing ParentID IS copied verbatim to clone

`dup_child` mode: `L1.parent = L2` (L2 = id 17). Source = L1.

Post-dup parse:

```
[0] id=20  parent=17    ← clone of L1, parent = L2's id (COPIED from source.ParentID)
[1] id=19  parent=17    original L1, parent=17
[2] id=17  parent=0     L2
[3] id=15  parent=0     L3
```

**Clone's ParentID = source's ParentID = 17**, written to ldta @0x84. Both L1 instances are children of L2.

**Impl rule**: clone's `Layer.ParentID = source.ParentID`. Write same bytes to ldta @0x84..0x87. Same for `Layer.TrackMatteLayerID` if present (ldta @0xA0..0xA3, AE 23+ — needs separate AE 2025 RE to verify; default assumption: copy verbatim like ParentID).

---

## Finding 8: trackMatteType byte (ldta @0x6B) copied verbatim

`dup_matted` JSX log (both pre and post-dup show L2-instances with trackMatteType=5013):

```
post-duplicate:
  [1] L2_mid trackMatteType=5013    ← new (clone)
  [2] L1_top
  [3] L2_mid trackMatteType=5013    ← original
```

**Impl rule**: trackMatte byte (ldta @0x6B) on the clone = source's value. Byte-copy. Matches DeleteLayer F3 (the type byte is "intent" — AE preserves it across delete/dup operations even when the implicit semantic breaks).

---

## Finding 9: No auto-suffix on layer name

ScriptingAPI's `layer.duplicate()` does NOT add " 2" / " (copy)" suffix. Both L2 instances are literally named "L2_mid" — no disambiguation. (The GUI Edit→Duplicate may differ; this is the JSX-driven scripting path.)

**Impl decision**: caller passes explicit `name string` parameter — cleaner than mimicking AE's name-collision quirk. Validate non-empty (matching NewShapeLayer signature). Refuse if name matches an existing layer in the comp? Out of scope — AE accepts duplicates, so we do too.

---

## Finding 10: 14 followers + ldta-body are byte-identical clone (verified)

`tmp_debug/diff_dup_blocks test_data/re_duplicate_layer_solo.aep 25 41` (clone block vs source block) — all 14 follower leaves (fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2) **BYTE-IDENTICAL**. AE does NOT mutate follower content.

`tmp_debug/probe_ldta` — ldta inside Layr LIST differs **ONLY at @0x00..0x03 (layer ID)** for the clone. @0x04 onward is verbatim copy (verified for solo + dup_child fixtures). In dup_child specifically, @0x84..0x87 (ParentID) reads `00 00 00 11` on BOTH original L1 and clone L1 (= L2's id 17). So clone's ParentID @0x84 is byte-copy of source's ParentID. Symmetric for @0x6B (TrackMatte byte) per Finding 8.

**Impl conclusion**: cloning a layer's 16-chunk block = byte-copy every chunk, then mutate ONLY ldta @0x00..0x03 to the new layer ID. Everything else (source ID @0x28, ParentID @0x84, TrackMatte @0x6B, all follower leaves) stays as source.

Per-chunk Data slice MUST be a fresh `append([]byte(nil), src.Data...)` copy (NOT slice-share) so subsequent mutations on clone don't touch source bytes. See `incidents/concurrency-unsafe-shared-chunk-bytes.md`.

---

## Open / not RE'd (deferred)

- **AE 23+ explicit TrackMatteLayerID (ldta @0xA0)**: needs an AE 2025 RE run with explicit setTrackMatte before duplicate. Phase 3 covers implicit (AE 2020) path via Finding 8.
- **Shape / text source clone**: non-AV layers refused in Phase 3 (per plan non-goals). Future phase.
- **Multi-select duplicate**: AE GUI allows Ctrl+D on multiple selected layers. Out of scope; `DuplicateLayers([]int)` wrapper is Phase 4+ if demand surfaces.
- **Effects / property animations**: cloned tdgp/property bytes presumed verbatim, but no per-property RE in this scope. Will need follow-up if dup'd animated layers misbehave in AE.

---

## Related

- Plan: [`../plans/2026-05-28-v3-phase3-duplicatelayer-plan.md`](../archive/plans/2026-05-28-v3-phase3-duplicatelayer-plan.md)
- Predecessor scar: [`ae-deletelayer-re.md`](ae-deletelayer-re.md) — F4 (16-chunk per-layer block) is what we clone
- AE acceptance gate: [`ae25-acceptance-gate.md`](ae25-acceptance-gate.md)
- Fixture JSX: `test_data/re_duplicate_layer.jsx`
- Dump tool: `tools/debug/dump_layers/main.go` (reused from Phase 2)
