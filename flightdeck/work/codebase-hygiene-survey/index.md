# Index — codebase-hygiene-survey

## State

New investigation effort created from the user's 2026-07-03 request to make the growing codebase and knowledge base auditable before the next major work phase.

Initial scan and first cleanup batch completed:

- `internal/`: 1024 files, 717 Go files, 307 embedded templates/data files. Largest code surfaces are `aep_test` (258 Go tests), `aepmigrate` (150 Go files), `serializer` (95), `scene` (53), `registry` (32), and `selfhost` (27).
- `flightdeck/knowledge`: 141 local knowledge files after adding `architecture/internal-codebase-map.md`. The 44 recipe notes that lacked routing-complete `SUMMARY` / `READ WHEN` headers have been repaired; current missing header count is 0.
- `test_data/generators`: 239 JSX scripts. Prefix scan: `verify*` 147 total including 16 active `verify_v2_2_*`, `re*` 69, `probe*` 8, `build*` 6, `gen*` 4, `fdta*` 2, `ship_gate*` 2, `smoke*` 1. `verify_v2_2_*` is active and must not be treated as old AE 2022 code.
- `tmp`: started with 116 JSON files, about 4.0 MB, all ignored. The ledger marked 106 as `safe-delete`; those were removed. Current `tmp` JSON count is 10, all `regenerate-only` command/report paths.

## Next

Next batch:

1. Review `ledgers/generator-ledger.md` delete candidates against produced fixtures/templates/knowledge provenance before any JSX deletion.
2. Review the 2 `stale-codepath-review` knowledge entries and decide whether to update moved paths or preserve them as historical context.
3. Review `ledgers/internal-ledger.md` large-package watch areas (`serializer`, `scene`, `aepmigrate`, `aep_test`) and decide whether any need a dedicated split plan.
4. Keep the 10 remaining `tmp` JSON files unless their producer command references are changed or regenerated.

## Read now

- design.md
- ledgers/internal-ledger.md
- ledgers/knowledge-ledger.md
- ledgers/generator-ledger.md
- ledgers/tmp-json-ledger.md
- flightdeck/knowledge/architecture/internal-codebase-map.md
- flightdeck/knowledge/workflow/folder-usage-policy.md
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
  - `knowledge-ledger.md`: all local knowledge classified; current classes are `valid: 139`, `stale-codepath-review: 2`.
  - `generator-ledger.md`: 239 JSX scripts classified; current classes are `active-ship-gate: 136`, `active-re-fixture: 75`, `delete-candidate: 26`, `historical-re: 2`.
  - `tmp-json-ledger.md`: remaining 10 `tmp` JSON files are all `regenerate-only`.
- Repaired routing headers on 44 recipe knowledge files; verified `MISSING_HEADERS 0`.
- Updated README architecture section to include migration, recipe, governance, technique, and service layers.
- Removed 106 ignored `tmp/*.json` files classified as `safe-delete`; verified `TMP_JSON_COUNT 10`.
- Ran current registry/capindex sanity gates:
  - `go run ./cmd/aepregistry audit -root .` → pass, 0 errors, 0 warnings.
  - `go run ./cmd/aepregistry layout -root .` → 18 locations, 0 cleanup candidates, 0 blocked unowned.
  - `go run ./cmd/aepregistry ownership -root .` → 18 locations, 1572 files, 1274 owned, 298 unowned.
  - `go run ./cmd/capindex -check` → pass.

Current:

- Ready for generator provenance review and stale-codepath knowledge review.

## Open questions

- Should unreferenced generator candidates be archived to cold storage first, or deleted after a clean reference/gate pass?
