---
state: finished
shipped_on: 2026-05-28
ship_gate: AE 2025 PASS (single-mode, AE 23+ semantic so AE 2020 N/A per §5)
---

# V3 Phase 5B — DuplicateLayer: soften matte refuse (explicit AE 23+ case)

> **For agentic workers:** narrow follow-up to Phase 3 DuplicateLayer. Single refuse-condition split + 4 tests + 1 ge fixture mode. Pattern reference: Phase 3 DuplicateLayer shipping pattern + Phase 5A explicit-matte byte path reuse.

**Spec impact**: amends [`../specs/finish/2026-05-28-v3-phase3-duplicatelayer-strategy.md`](../specs/finish/2026-05-28-v3-phase3-duplicatelayer-strategy.md) § 5 refuse-table. Recorded inline; no new strategy spec doc.
**Predecessor**: V3 Phase 3 DuplicateLayer (commit `b2d3e18`) + V3 Phase 5A SetTrackMatteSource (uncommitted in working tree, 2026-05-28).
**Source preview**: Phase 5A plan §10.

**Goal**: relax the `source.TrackMatte != TrackMatteNone` refuse case so that **explicit AE 23+ matte** (`TrackMatteLayerID != 0`) is allowed. Implicit "layer-above" matte (`TrackMatteLayerID == 0`) stays refused per F2 quirk.

**Hypothesis (justifies skipping fresh RE)**: F2 quirk — AE relocating clone to preserve original's matte — applies ONLY to implicit matte (where matte source = "layer above" depends on position). AE 23+ explicit matte stores source ID in ldta @0xA0..0xA3, decoupled from layer order. Duplicating an explicit-matte layer should therefore:
1. Insert clone at source's old index (same as solo / dup_parent / dup_child — no F2 relocation)
2. Byte-copy @0xA0..0xA3 and @0x6B verbatim (F4 footage-share + F8 matte-byte-verbatim semantics already proven; deep-clone path handles this automatically — no new code)
3. Result: clone and source both matte-pointer to the same source layer; AE renders both correctly

This hypothesis is grounded in F2's stated reason ("AE preserves original L2's matte source = 'layer above original' = L1") and Phase 5A's existing AE 23+ acceptance proof (`re_trackmatte_ae24.aep` shows AE accepts explicit @0xA0 regardless of position). Ship-gate validates the hypothesis end-to-end.

**Architecture reuse**:
- Existing deep-clone path in `duplicate_layer.go` already byte-copies ldta @0xA0 + @0x6B verbatim (only @0x00..0x03 is overwritten with new ID). No new byte-write code.
- `re_trackmatte_ae24.aep` (existing, AE-saved) is reused as the Go-side fixture for explicit-matte tests — already has 3 layers with `setTrackMatte` applied, no new RE fixture needed.
- ge fixture producer follows Phase 3's `tmp_debug/ge_duplicate_layer/` pattern (Open baseline → mutate via Go → write).

**Non-goals**:
- AE ≤ 22 implicit matte auto-relocation (F2 special-case) — still refused; implementing AE's position-shift quirk is Phase 5B+ if demand surfaces.
- Cross-comp matte — DuplicateLayer is single-comp; cross-comp is Phase 5C `InsertLayer`.
- Shape/Text layer matte dup — V2.2.1 deferred (out of scope for Phase 5B).

---

## 1. Public API signature

No new API. Existing signature unchanged:

```go
func (c *Composition) DuplicateLayer(index int, name string) (*Layer, error)
```

What changes: refuse-table semantics inside the existing body. godoc is updated to reflect Phase 5B relaxation.

---

## 2. Algorithm change

**Single condition split in step 1 of existing algorithm** (`duplicate_layer.go:68-70`):

```diff
- if source.TrackMatte != TrackMatteNone {
-     return nil, fmt.Errorf("DuplicateLayer: refuse layer %q (idx=%d) with TrackMatte=%d set; AE relocates clone to preserve original's matte (F2 quirk), not yet supported in Phase 3", source.Name, index, source.TrackMatte)
- }
+ // F2 quirk applies only to implicit "layer-above" matte where matte source
+ // is positional. AE 23+ explicit matte (TrackMatteLayerID != 0) decouples
+ // matte from layer order, so duplicating is safe — the deep-clone path
+ // byte-copies @0xA0 + @0x6B verbatim. Implicit matte still refused.
+ if source.TrackMatte != TrackMatteNone && source.TrackMatteLayerID == 0 {
+     return nil, fmt.Errorf("DuplicateLayer: refuse layer %q (idx=%d) with implicit TrackMatte=%d (TrackMatteLayerID=0); AE relocates clone to preserve original's matte (F2 quirk), not yet supported", source.Name, index, source.TrackMatte)
+ }
```

That's the entire logic change. Steps 2-12 of the existing algorithm are unaffected — deep-clone already handles @0xA0 + @0x6B byte-copy via verbatim ldta.Data clone.

---

## 3. Refuse-cases (Phase 5B updated table)

| Case | Detection | Reason |
|---|---|---|
| `name == ""` | (unchanged) | Forces caller intent |
| Index OOR | (unchanged) | Standard bounds |
| Comp lacks back-ref | (unchanged) | No itemList to splice |
| Source not AV | (unchanged) | Camera/Light/Audio ldta not RE'd |
| **Implicit matte (NEW)** | `TrackMatte != None && TrackMatteLayerID == 0` | F2 quirk: AE relocates clone for implicit "layer-above" matte — not supported |
| ~~Source has TrackMatte != None (OLD)~~ | ~~always refuse~~ | **Replaced by implicit-only refuse above** |
| Backref corruption | (unchanged) | Defensive |

The relaxation: any source with **explicit** matte (`TrackMatteLayerID != 0`) now proceeds through the standard clone path.

---

## 4. Test plan

`internal/aep/duplicate_layer_test.go` — 4 new tests:

| Test | Fixture | Verifies |
|---|---|---|
| `ExplicitMatte_HappyPath` | `re_trackmatte_ae24.aep` | DuplicateLayer on a matted layer (e.g. mt_alpha_to_solidA, idx 0) succeeds; clone has expected new ID, name; layer count +1; itemList grew |
| `ExplicitMatte_VerbatimBytes` | `re_trackmatte_ae24.aep` | Clone's ldta @0xA0..0xA3 byte-equal to source; @0x6B byte-equal; clone.TrackMatteLayerID == source.TrackMatteLayerID; clone.TrackMatte == source.TrackMatte |
| `ImplicitMatte_StillRefused` | `re_delete_layer_baseline.aep` | Set `Layers[1].TrackMatte = Alpha` directly (TrackMatteLayerID stays 0); DuplicateLayer returns error mentioning "implicit" |
| `ExplicitMatte_RoundTrip` | `re_trackmatte_ae24.aep` | DuplicateLayer + WriteAEP + reparse → clone still has explicit matte; ldta bytes preserved |

Also update existing `TestDuplicateLayer_RefuseTrackMatte` to set both `TrackMatte = Alpha` AND `TrackMatteLayerID == 0` (i.e. assert implicit path still refused — error message updated to match new wording).

Acceptance: vet clean; PASS bumps by +4; `RefuseTrackMatte` migrates to `ImplicitMatte_StillRefused` semantics; FAIL = 0.

---

## 5. Ship-gate

**Required**: 1 new mode (`explicit_matte`) × AE 2025 only = +1 case. AE 2020 NOT applicable — explicit matte is AE 23+, the fixture uses @0xA0 ldta slot absent in AE 2020.

Justification for skipping AE 2020:
- Phase 5A precedent: `Layer.SetTrackMatteSource` shipped without AE 2020 ship-gate because explicit matte is AE 23+ only.
- Source fixture (`re_trackmatte_ae24.aep`) is itself AE 23+ — AE 2020 can't even open it.

Ship-gate steps:
1. `go run ./tmp_debug/ge_duplicate_layer_explicit_matte` → produces `test_data/ge_duplicate_layer_explicit_matte.aep`
2. User opens in AE 2025 — must not "file data missing" / silent-drop
3. User runs `test_data/verify_ge_duplicate_layer_explicit_matte.jsx` — must report PASS (clone exists; clone's `trackMatteLayer` is the expected source layer; original's `trackMatteLayer` still points to same source)

**Hanging task until user confirms ship-gate green**: Phase 5B stays godoc-alpha (no Stable promotion).

---

## 6. File map

Create:
- `workshop/plans/2026-05-28-v3-phase5b-duplicatelayer-explicit-matte-plan.md` — this doc
- `tmp_debug/ge_duplicate_layer_explicit_matte/main.go` — ge fixture producer
- `test_data/verify_ge_duplicate_layer_explicit_matte.jsx` — AE-side ship-gate verifier
- (after ship-gate green) commit: nothing new

Modify:
- `internal/aep/duplicate_layer.go` — refuse-condition split + godoc tweak (~5 LOC delta)
- `internal/aep/duplicate_layer_test.go` — +4 tests, 1 existing test updated (~120 LOC delta)
- `docs/layer.md` — `Composition.DuplicateLayer` matte note refreshed
- `workshop/plans/coverage-detail.md` — `trackMatteLayer` row: add "DuplicateLayer clones explicit matte verbatim"
- `workshop/board.md` — Phase 5B in flight → Recently finished after ship-gate
- `workshop/scars/ae-duplicatelayer-re.md` — F2 footnote: "Phase 5B confirmed explicit (AE 23+) matte is decoupled from F2 quirk; only implicit case still refused"

No changes to:
- `parse_layer.go` / `write_layer.go` / `types_core.go` — no new chunk semantics
- `delete_layer.go` / `move_layer.go` — unrelated
- existing `SetTrackMatte*` API — unchanged

---

## 7. Tasks

- [x] **Task 1**: this plan doc
- [ ] **Task 2**: implement refuse-split in `duplicate_layer.go` (~5 LOC) + godoc tweak
- [ ] **Task 3**: 4 new tests in `duplicate_layer_test.go` + migrate existing `RefuseTrackMatte`
- [ ] **Task 4**: `tmp_debug/ge_duplicate_layer_explicit_matte/main.go` + `test_data/verify_ge_duplicate_layer_explicit_matte.jsx`
- [ ] **Task 5**: docs sync (`docs/layer.md`, `coverage-detail.md`, scar F2 footnote)
- [ ] **Task 6**: `go run` ge producer to materialize `ge_duplicate_layer_explicit_matte.aep`
- [ ] **Task 7**: hand off ship-gate to user (AE 2025); board update; plan stays in `plans/` (not finish/) until green; promote Stable on green

---

## 8. Acceptance

1. `go vet ./...` clean
2. `go test ./internal/aep/ -count=1` PASS bumps by ≥4 vs Phase 5A baseline (307 → ≥311); FAIL = 0
3. Existing `TestDuplicateLayer_RefuseTrackMatte` semantics migrated to implicit-only refuse (still PASS)
4. `ge_duplicate_layer_explicit_matte.aep` exists, size > 0, no Go warnings during produce
5. Verify JSX exists with clear PASS/FAIL output for AE-side reading
6. godoc tagged Alpha (Phase 5B); promoted Stable + plan → finish/ + commit only after user confirms ship-gate green

---

## 9. Risk register

| Risk | Likelihood | Mitigation |
|---|---|---|
| AE 2025 rejects file ("data missing") despite hypothesis | Low | Ship-gate catches; if fail → bisect: was it explicit matte specifically (single-source dup) or something else (e.g. naming collision)? `tmp_debug/diff_blocks` to compare bytes |
| AE 2025 opens but renders both mattes wrong (e.g. silently re-points clone to clone-of-source) | Low-Medium | Verify JSX checks `clone.trackMatteLayer.name == expectedSource.name` AND `original.trackMatteLayer.name == expectedSource.name` |
| Clone's explicit @0xA0 points to a layer that doesn't exist (e.g. source was matted by a layer in another comp — shouldn't happen but defensive) | Very Low | `re_trackmatte_ae24.aep` has all sources in same comp; not a risk in our fixture. Future cross-comp dup (Phase 5C) will need a new refuse-case |
| Existing `TestDuplicateLayer_RefuseTrackMatte` test breaks unexpectedly | Low | Migrate test purposefully — change setup to explicitly set `TrackMatteLayerID = 0` so it tests the implicit path |

---

## 10. Open / deferred to future phases

- **AE ≤ 22 implicit matte position-shift duplicate** (F2 quirk implementation) — adds ~30 LOC special-case; Phase 5B+ if demand
- **Cross-comp matte duplicate** — `InsertLayer(src *Layer, atIdx int)` Phase 5C; refuse explicit matte when source layer in another comp
- **Shape/Text matte duplicate** — V2.2.1 ShapeLayer prerequisites must clear first

---

## Related

- Spec amended: [`../specs/finish/2026-05-28-v3-phase3-duplicatelayer-strategy.md`](../specs/finish/2026-05-28-v3-phase3-duplicatelayer-strategy.md) § 5 refuse-table
- F2 scar: [`../scars/ae-duplicatelayer-re.md`](../scars/ae-duplicatelayer-re.md) Finding 2 (implicit matte position-shift)
- Phase 5A predecessor: [`finish/2026-05-28-v3-phase5a-settrackmattesource-plan.md`](finish/2026-05-28-v3-phase5a-settrackmattesource-plan.md) §10 preview
- AE 23+ matte fixture: `test_data/re_trackmatte_ae24.aep` + `.jsx`
- AE acceptance gate: [`../scars/ae25-acceptance-gate.md`](../scars/ae25-acceptance-gate.md)
