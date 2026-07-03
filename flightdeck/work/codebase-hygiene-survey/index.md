# Index — codebase-hygiene-survey

## State

New investigation effort created from the user's 2026-07-03 request to make the growing codebase and knowledge base auditable before the next major work phase.

Initial scan and first cleanup batch completed:

- `internal/`: 1024 files, 717 Go files, 307 embedded templates/data files. Largest code surfaces are `aep_test` (258 Go tests), `aepmigrate` (150 Go files), `serializer` (95), `scene` (53), `registry` (32), and `selfhost` (27).
- `flightdeck/knowledge`: 142 local knowledge files after adding `architecture/internal-codebase-map.md` and `workflow/jsx-generator-cleanup-provenance.md`. The 44 recipe notes that lacked routing-complete `SUMMARY` / `READ WHEN` headers have been repaired; current missing header count is 0. The 2 stale-codepath rows have been reviewed and repaired/reclassified.
- `test_data/generators`: 239 JSX scripts. Prefix scan: `verify*` 147 total including 16 active `verify_v2_2_*`, `re*` 69, `probe*` 8, `build*` 6, `gen*` 4, `fdta*` 2, `ship_gate*` 2, `smoke*` 1. `verify_v2_2_*` is active and must not be treated as old AE 2022 code.
- Generator provenance review of the 26 mechanical `delete-candidate` rows found `keep-active: 10`, `historical-evidence-review: 1`, `archive-or-delete-candidate: 15`, `direct-delete-now: 0`; no JSX was deleted in this batch.
- `tmp`: started with 116 JSON files, about 4.0 MB, all ignored. The ledger marked 106 as `safe-delete`; those were removed. Current `tmp` JSON count is 10, all `regenerate-only` command/report paths.

## Next

Next batch:

1. Decide historical evidence policy for `verify_ge_duplicate_layer_explicit_matte.jsx`.
2. Decide archive/delete policy for the 15 generator candidates listed in `ledgers/generator-provenance-review.md`, including whether their tracked fixture outputs and `tools/debug/split_roundtrip_baseline/baseline.txt` entries still matter.
3. Keep `serializer`, `scene`, `aepmigrate`, and `aep_test` as watch areas only; no package split belongs in this survey without a dedicated plan.
4. Keep the 10 remaining `tmp` JSON files unless their producer command references are changed or regenerated.

## Read now

- design.md
- ledgers/internal-ledger.md
- ledgers/knowledge-ledger.md
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
  - `generator-ledger.md`: 239 JSX scripts classified; current classes are `active-ship-gate: 136`, `active-re-fixture: 75`, `delete-candidate: 26`, `historical-re: 2`.
  - `tmp-json-ledger.md`: remaining 10 `tmp` JSON files are all `regenerate-only`.
- Repaired routing headers on 44 recipe knowledge files; verified `MISSING_HEADERS 0`.
- Updated README architecture section to include migration, recipe, governance, technique, and service layers.
- Removed 106 ignored `tmp/*.json` files classified as `safe-delete`; verified `TMP_JSON_COUNT 10`.
- Repaired/reclassified the 2 stale-codepath knowledge entries:
  - `knowledge/workflow/ae25-acceptance-gate.md` now points at current `internal/serializer/templates/project` and `internal/aep_test` paths.
  - `knowledge/layer/v2-2-aelayer-structure.md` now explicitly marks old `internal/aep/*` mentions as historical context.
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

Current:

- Ready for historical evidence policy on `verify_ge_duplicate_layer_explicit_matte.jsx` and archive/delete policy decisions for the 15 generator candidates.

## Open questions

- For 15 generator archive/delete candidates, should their tracked fixture outputs and split-roundtrip baseline hashes be deleted together, migrated to durable evidence, or preserved as historical baselines?
