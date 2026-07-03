# Index — codebase-hygiene-survey

## State

New investigation effort created from the user's 2026-07-03 request to make the growing codebase and knowledge base auditable before the next major work phase.

Initial scan and first cleanup batch completed:

- `internal/`: 1024 files, 717 Go files, 307 embedded templates/data files. Largest code surfaces are `aep_test` (258 Go tests), `aepmigrate` (150 Go files), `serializer` (95), `scene` (53), `registry` (32), and `selfhost` (27).
- `flightdeck/knowledge`: 142 local knowledge files after adding `architecture/internal-codebase-map.md` and `workflow/jsx-generator-cleanup-provenance.md`. The 44 recipe notes that lacked routing-complete `SUMMARY` / `READ WHEN` headers have been repaired; current missing header count is 0. The 2 stale-codepath rows have been reviewed and repaired/reclassified. A body-validity pass fixed stale current paths and marked generated evidence that is no longer present on disk.
- `test_data/generators`: 223 JSX scripts after cleanup. The `verify_v2_2_*` scripts are still active and must not be treated as old AE 2022 code.
- Generator provenance review of the 26 mechanical `delete-candidate` rows ended with `keep-active: 10`, `deleted-script: 16`, `deleted-tracked-fixture: 8`, and `stale-baseline-hashes-removed: 16`.
- `tmp`: started with 116 JSON files, about 4.0 MB, all ignored. The ledger marked 106 as `safe-delete`; those were removed. Current `tmp` JSON count is 10, all `regenerate-only` command/report paths.

## Next

Next batch:

1. Treat the broader `scripts/fixtures/regen_fixtures.ps1 -CheckOnly` missing generated-output inventory as a separate fixture regeneration package if those generated fixtures are needed again.
2. Keep `serializer`, `scene`, `aepmigrate`, and `aep_test` as watch areas only; no package split belongs in this survey without a dedicated plan.
3. Keep the 10 remaining `tmp` JSON files unless their producer command references are changed or regenerated.

## Read now

- design.md
- ledgers/internal-ledger.md
- ledgers/knowledge-ledger.md
- ledgers/knowledge-body-validity-review.md
- ledgers/generator-ledger.md
- ledgers/generator-provenance-review.md
- ledgers/tmp-json-ledger.md
- ledgers/internal-watch-review.md
- flightdeck/knowledge/architecture/internal-codebase-map.md
- flightdeck/knowledge/workflow/folder-usage-policy.md
- flightdeck/knowledge/workflow/jsx-generator-cleanup-provenance.md
- flightdeck/knowledge/effects/embed-template-architecture.md

## Read if

- flightdeck/knowledge/workflow/re-fixture.md — if changing or deleting `test_data/generators/*.jsx`
- flightdeck/knowledge/workflow/verify.md — if selecting verification commands for cleanup batches
- flightdeck/knowledge/docgen/docgen-alias-blindspot-false-green.md — if public API doc coverage changes during README/doc cleanup

## Progress

Done:

- Created the work package and initial investigation spec.
- Added a durable architecture knowledge note for `internal/`.
- Ran lightweight scans for `internal`, knowledge, JSX generators, and `tmp`.
- Generated four ledgers under `ledgers/`:
  - `internal-ledger.md`: package responsibility / size / first-pass class.
  - `knowledge-ledger.md`: all local knowledge classified; current classes are `valid: 142`.
  - `generator-ledger.md`: current JSX scripts classified; current classes are `active-ship-gate: 136`, `active-re-fixture: 85`, `historical-re: 2`.
  - `tmp-json-ledger.md`: remaining 10 `tmp` JSON files are all `regenerate-only`.
- Repaired routing headers on 44 recipe knowledge files; verified `MISSING_HEADERS 0`.
- Updated README architecture section to include migration, recipe, governance, technique, and service layers.
- Removed 106 ignored `tmp/*.json` files classified as `safe-delete`; verified `TMP_JSON_COUNT 10`.
- Repaired/reclassified the 2 stale-codepath knowledge entries:
  - `knowledge/workflow/ae25-acceptance-gate.md` now points at current `internal/serializer/templates/project` and `internal/aep_test` paths.
  - `knowledge/layer/v2-2-aelayer-structure.md` now explicitly marks old `internal/aep/*` mentions as historical context.
- Added `ledgers/knowledge-body-validity-review.md` and ran an evidence-first body validity pass:
  - Fixed stale test/template paths in `recipe-draft-3d-profile.md`, `pseudo-effect-continuation-handoff.md`, `ae-drops-unknown-chunks-on-resave.md`, `ae25-acceptance-gate.md`, and `re-fixture.md`.
  - Marked generated AEP evidence in effect, camera/light, and wiggle-modulation notes as historical or manifest-owned when it is not present in the current checkout.
  - Applied no simple similarity merges; high-similarity recipe notes remain separate unless a future family note preserves all per-field evidence.
- Added `ledgers/generator-provenance-review.md`; no generator was deleted because the review found active manifest rows and provenance/evidence questions among the mechanical delete candidates.
- Repaired 4 generator provenance gaps:
  - Added manifest ownership for `build_material_classic_2020.jsx`.
  - Repaired `build_re_comp_idta.jsx` output to `test_data/fixtures/re_comp_idta.aep` and added manifest ownership.
  - Repaired `build_re_layer_comment.jsx` output to `test_data/fixtures/re_layer_comment.aep` and added manifest ownership.
  - Added manifest ownership for `gen_text_range_adv_smoothness.jsx`.
- Verified `scripts/fixtures/fixtures_manifest.json` parses as JSON. `pwsh -NoProfile -File scripts/fixtures/regen_fixtures.ps1 -CheckOnly` ran and reported `79 manifest jobs, 47 with missing outputs (46 driveable, 1 manual-only), 0 ungoverned on-disk aep`; the repaired four tracked fixture entries were not reported missing, but broader generated fixture gaps remain.
- Ran current registry/capindex sanity gates after manifest/script repair:
  - `go run ./cmd/aepregistry audit -root .` -> pass, 0 errors, 0 warnings.
  - `go run ./cmd/aepregistry layout -root .` -> 18 locations, 0 cleanup candidates, 0 blocked unowned.
  - `go run ./cmd/aepregistry ownership -root .` -> 18 locations, 1574 files, 1274 owned, 300 unowned.
  - `go run ./cmd/capindex -check` -> pass.
  - `git diff --check` -> pass.
- Added `ledgers/internal-watch-review.md`; `serializer`, `scene`, `aepmigrate`, and `aep_test` remain watch areas, not in-survey split targets.
- Added `knowledge/workflow/jsx-generator-cleanup-provenance.md` as a reusable cleanup guard.
- Ran current registry/capindex sanity gates again after this batch:
  - `go run ./cmd/aepregistry audit -root .` -> pass, 0 errors, 0 warnings.
  - `go run ./cmd/aepregistry layout -root .` -> 18 locations, 0 cleanup candidates, 0 blocked unowned.
  - `go run ./cmd/aepregistry ownership -root .` -> 18 locations, 1572 files, 1274 owned, 298 unowned.
  - `go run ./cmd/capindex -check` -> pass.
  - `git diff --check` -> pass.
- Ran current registry/capindex sanity gates:
  - `go run ./cmd/aepregistry audit -root .` → pass, 0 errors, 0 warnings.
  - `go run ./cmd/aepregistry layout -root .` → 18 locations, 0 cleanup candidates, 0 blocked unowned.
  - `go run ./cmd/aepregistry ownership -root .` → 18 locations, 1572 files, 1274 owned, 298 unowned.
  - `go run ./cmd/capindex -check` → pass.
- Removed 16 stale generator scripts, 8 unreferenced tracked fixture AEPs, and 16 stale `test_data/generated/ship-gate/ge_*` split-roundtrip baseline hashes after provenance review.
- Removed stale `fixture_generators.smoke_helpers` registry atom because the deleted placeholder `smoke_ae_run.jsx` was the last `smoke_*.jsx` dependency and no current workflow owns that generator family.
- Post-cleanup fixture inventory:
  - `pwsh -NoProfile -File scripts/fixtures/regen_fixtures.ps1 -CheckOnly` → `79 manifest jobs, 47 with missing outputs (46 driveable, 1 manual-only), 0 ungoverned on-disk aep`.
  - The 47 missing outputs are the broader generated fixture inventory under `test_data/generated/fixtures`; they predate this deletion batch and are not ungoverned tracked AEPs.
- Post-cleanup registry/capindex sanity gates:
  - `go run ./cmd/aepregistry audit -root .` → pass, 0 errors, 0 warnings.
  - `go run ./cmd/aepregistry layout -root .` → 18 locations, 0 cleanup candidates, 0 blocked unowned.
  - `go run ./cmd/aepregistry ownership -root .` → 18 locations, 1550 files, 1250 owned, 300 unowned.
  - `go run ./cmd/capindex -check` → pass.

Current:

- Survey execution is complete for the requested internal map, knowledge classification/header repair, generator cleanup, and tmp cleanup scope.

## Open questions

- None for generator cleanup after verification; remaining generated fixture gaps are the broader `regen_fixtures.ps1 -CheckOnly` missing-output inventory.
