# V3 Phase 5A — Layer.SetTrackMatteSource Implementation Plan

> **For agentic workers:** trivial convenience-wrapper slice. No RE, no ship-gate, no separate strategy spec. Pattern reference: Layer.MoveAfter/MoveBefore wrappers (commit `8e3c17e`) — pure delegate over an existing Stable structural primitive.

**Spec**: extends [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 picklist (board recommendation).
**Predecessor**: V3 Phase 4 MoveLayer + Layer.Move* wrappers (2026-05-28, commits `5fd09c2` + `8e3c17e`). PASS 278 + 9 = 287.

**Goal**: `(l *Layer) SetTrackMatteSource(src *Layer, mode TrackMatteType) error` — high-level convenience wrapper mirroring AE ScriptingAPI 23+ `layer.setTrackMatte(srcLayer, type)`. Delegates to existing `Layer.SetTrackMatteLayer(sourceID uint32, mode TrackMatteType)` after `*Layer`-specific validation.

**Architecture reuse**:
- Existing `SetTrackMatteLayer` handles all byte writes (ldta @0xA0 + @0x6B). Already Stable + round-trip tested.
- `re_trackmatte_ae24.aep` proves AE 23+ accepts explicit @0xA0 reference regardless of layer order (no reorder needed).
- Mirrors `Layer.SetLightSource(target *Layer)` (codebase convention for `*Layer`-arg Set\*Source setters).

**Non-goals (Phase 5A)**:
- Soften DuplicateLayer matte refuse-case → Phase 5B (separate slice; requires new AE 2025 RE fixture `re_duplicatelayer_explicit_matte.aep`)
- AE ≤ 22 implicit-matte auto-reorder → would need MoveLayer integration + RE; not in V3 scope
- Cross-comp matte → AE refuses; we refuse early with clear error

---

## 1. Public API signature

```go
// SetTrackMatteSource designates `src` as this layer's explicit
// track-matte source and writes the matte mode. Mirrors AE ScriptingAPI
// 23+ layer.setTrackMatte(srcLayer, type). Requires AE 23+ ldta (the
// @0xA0 slot — re-save through AE 23+ first on AE 22 / 2020 files).
//
// On success, writes ldta @0xA0..0xA3 = src.ID (4 bytes BE) and
// @0x6B = mode (1 byte). length-preserving.
//
// Refuse-cases:
//   - src == nil
//   - layer or src lacks comp back-pointer (built outside parser)
//   - cross-comp matte (src.comp != l.comp)
//   - self-matte (src.ID == l.ID)
//   - ldta too short (caught by SetTrackMatteLayer; surfaced verbatim)
//
// To clear, use ClearTrackMatteLayer() — already exists.
func (l *Layer) SetTrackMatteSource(src *Layer, mode TrackMatteType) error
```

**Why no overload of existing `SetTrackMatte(t TrackMatteType)`**: Go has no overloading. Existing 1-arg `SetTrackMatte` is Stable (CLAUDE.md #2) and can't be renamed. `SetTrackMatteSource` is the unambiguous discoverable name next to `SetTrackMatteLayer` (id-arg variant).

---

## 2. Algorithm

```
algorithm: SetTrackMatteSource(l, src, mode) → error:
  // 1. Validation (early returns with clear error messages):
  if src == nil:                     return error "matte source nil"
  if l.comp == nil:                  return error "layer %q not in parsed comp"
  if src.comp == nil:                return error "matte source %q not in parsed comp"
  if l.comp != src.comp:             return error "cross-comp matte: src in %q, l in %q"
  if src.ID == l.ID:                 return error "self-matte (sourceID == own ID = %d) not allowed"

  // 2. Delegate to existing setter (handles ldta-length check + byte writes
  //    + sets Layer.TrackMatteLayerID + Layer.TrackMatte).
  return l.SetTrackMatteLayer(src.ID, mode)
```

No atomic invariants needed — `SetTrackMatteLayer` is itself length-preserving and side-effect-free on validation failure (it checks lengths before writing). Validation in this wrapper happens BEFORE delegation, so no partial state on refuse.

---

## 3. Refuse-cases

| Case | Detection | Reason |
|---|---|---|
| `src == nil` | direct nil check | Wrapper-specific |
| Layer not in parsed comp | `l.comp == nil` | Wrapper can't validate cross-comp without back-pointer |
| Source not in parsed comp | `src.comp == nil` | Same |
| Cross-comp matte | `l.comp != src.comp` | AE 23+ requires same-comp matte source |
| Self-matte | `src.ID == l.ID` | No-op in AE; usually a programmer error |
| ldta too short | (delegated) | AE 22 / 2020 file — re-save through AE 23+ first |
| Source ID not found | (delegated) | Defensive — shouldn't trigger when src belongs to l.comp |

---

## 4. Test plan

`internal/aep/layer_matte_test.go`:

| Test | Setup | Verifies |
|---|---|---|
| `HappyPath_Alpha` | open `re_trackmatte_ae24.aep`, mt_baseline + solidA | `SetTrackMatteSource(solidA, Alpha)` succeeds; TrackMatteLayerID == solidA.ID; TrackMatte == Alpha |
| `HappyPath_AlphaInverse` | same | mode=AlphaInverse |
| `HappyPath_Luma` | same | mode=Luma |
| `HappyPath_LumaInverse` | same | mode=LumaInverse |
| `IntentWithoutMode` | mt_baseline + solidA + mode=None | source pointer stored; matte channel inert (AE-legal preserve-state) |
| `RefuseNilSource` | mt_baseline | `SetTrackMatteSource(nil, Alpha)` returns error |
| `RefuseSelf` | mt_baseline | `SetTrackMatteSource(mt_baseline, Alpha)` returns error |
| `RefuseCrossComp` | mt_baseline + a layer from a 2nd comp | returns "cross-comp" error |
| `RefuseStub` | hand-built `&Layer{Name:"stub"}` calling on parsed mt_baseline as src | returns "not in parsed comp" error (l.comp == nil) |
| `RoundTrip` | mt_baseline + solidA | Set → WriteAEP → re-parse → both fields preserved |
| `ParityWithSetTrackMatteLayer` | two cloned projects | wrapper(src, mode) produces byte-identical WriteAEP output vs SetTrackMatteLayer(src.ID, mode) |

Acceptance: PASS bumps by ≥11 (pre-slice count read from CI); FAIL = 0; vet clean.

---

## 5. Ship-gate

**None.** Justification:
- Wrapper writes zero new byte patterns. All byte writes pass through `SetTrackMatteLayer`, which is Stable and round-trip green.
- Underlying acceptance pattern is already verified by `re_trackmatte_ae24.aep` (RE fixture explicitly places sources BELOW matted layers, proving AE 23+ explicit `@0xA0` works regardless of order).
- `TestSetTrackMatteLayerRoundtrip` covers the WriteAEP → re-parse path.
- `ParityWithSetTrackMatteLayer` test confirms the wrapper produces byte-identical output, transitively guaranteeing AE acceptance.

Same justification chain as Layer.Move\* wrappers shipped without separate ship-gate (board entry: 2026-05-28 V3 Phase 4 follow-up).

---

## 6. File map

Create:
- `internal/aep/layer_matte.go` — ~50 LOC (method + godoc)
- `internal/aep/layer_matte_test.go` — ~250 LOC (~11 tests + helpers)

Modify:
- `docs/layer.md` — add `Layer.SetTrackMatteSource` section near existing `SetTrackMatteLayer` (~25 lines)
- `workshop/plans/coverage-detail.md` — `AVLayer.trackMatteLayer` row: append "+ `SetTrackMatteSource(src, mode)` parity wrapper"
- `README.md` § Layer 字节字段 — add 1-line note for parity wrapper
- `workshop/board.md` — move plan to finish/, drop SetTrackMatte from Phase 5 picklist, add recently-finished entry

No changes to:
- existing `SetTrackMatte(t)` / `SetTrackMatteLayer(id, mode)` / `ClearTrackMatteLayer()` (all stay Stable, intact)
- types_core.go / parse_layer.go / write_layer.go (delegation only)

---

## 7. Tasks

- [ ] **Task 1**: this plan doc (you're reading it)
- [ ] **Task 2**: implement `layer_matte.go` + `layer_matte_test.go` (~11 tests)
- [ ] **Task 3**: run `go vet ./...` + `go test ./internal/aep/ -run 'TestSetTrackMatteSource|TestSetTrackMatteSourceRoundTrip|TestSetTrackMatteSourceParity'` → PASS
- [ ] **Task 4**: docs sync — `docs/layer.md` + `coverage-detail.md` + `README.md`
- [ ] **Task 5**: Stable promotion (godoc from start) + plan → finish/ + board update + commit

---

## 8. Acceptance

1. `go vet ./...` clean + tests PASS bumps by ≥11 vs pre-slice baseline, FAIL = 0
2. Public API: `Layer.SetTrackMatteSource(*Layer, TrackMatteType) error` godoc complete + Stable from start (no Alpha gate — pure delegate; underlying setter already ship-gate green via existing AE 23+ fixture)
3. `ParityWithSetTrackMatteLayer` test confirms byte-identical WriteAEP output vs underlying setter
4. Plan migrated to finish/, board.md updated to reflect Phase 5A ship + Phase 5B (DuplicateLayer matte soften) deferred
5. docs sync complete

---

## 9. Risk register

| Risk | Likelihood | Mitigation |
|---|---|---|
| Wrapper diverges from underlying byte writes over time (silent drift) | Low | `ParityWithSetTrackMatteLayer` test enforces equivalence; CI catches drift |
| `SetTrackMatteLayer` itself is wrong on edge case (e.g. AE ≤ 22 path) | Already known: errors clearly on short ldta | Wrapper inherits same behavior; tested via `RefuseStub` |
| API name conflict / discoverability issue | None | `SetTrackMatteSource` distinct from existing 3 methods; `*Source` suffix matches `SetLightSource` precedent |
| User confuses `SetTrackMatte(mode)` vs `SetTrackMatteSource(*Layer, mode)` | Medium docs UX issue | godoc cross-references all three matte setters; docs/layer.md gains a "三个 setter 怎么选" mini-table |

---

## 10. Phase 5B preview (deferred — separate slice)

Soften DuplicateLayer matte refuse-case when source has **explicit** AE 23+ matte (TrackMatteLayerID != 0):
- Spec impact: DuplicateLayer §5 refuse-table — split "TrackMatte != None" into "implicit (TrackMatteLayerID==0)" (still refuse) vs "explicit (TrackMatteLayerID!=0)" (allow, byte-copy F4+F8)
- RE need: AE 2025 fixture `re_duplicatelayer_explicit_matte.aep` — set @0xA0 BEFORE duplicate; verify clone gets verbatim @0xA0 and @0x6B; verify original's matte still works
- Ship-gate: 1 new mode × 2 AE versions = +2 cases on top of Phase 3's 6/6
- Estimated work: 1 session

Not part of Phase 5A. Recorded here so context isn't lost.

---

## Related

- Spec: [`../specs/2026-05-27-v3-deep-think.md`](../specs/2026-05-27-v3-deep-think.md) § 4 Phase 5 picklist
- Underlying setter: `internal/aep/write_layer.go::SetTrackMatteLayer` (Stable, already round-trip tested)
- RE fixture (existing, unchanged): `test_data/re_trackmatte_ae24.aep` + `.jsx`
- Pattern reference: Layer.Move\* wrappers shipping pattern — commit `8e3c17e`, no ship-gate, no strategy spec
- DuplicateLayer F2 scar (Phase 5B context): [`../scars/ae-duplicatelayer-re.md`](../scars/ae-duplicatelayer-re.md) Finding 2
