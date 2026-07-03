# Index — codebase-hygiene-survey

## State

New investigation effort created from the user's 2026-07-03 request to make the growing codebase and knowledge base auditable before the next major work phase.

Initial scan completed:

- `internal/`: 1024 files, 717 Go files, 307 embedded templates/data files. Largest code surfaces are `aep_test` (258 Go tests), `aepmigrate` (150 Go files), `serializer` (95), `scene` (53), `registry` (32), and `selfhost` (27).
- `flightdeck/knowledge`: 140 local knowledge files. 44 recipe notes lack routing-complete `SUMMARY` / `READ WHEN` headers; no duplicate titles found in the quick scan.
- `test_data/generators`: 239 JSX scripts. Prefix scan: `verify*` 147 total including 16 active `verify_v2_2_*`, `re*` 69, `probe*` 8, `build*` 6, `gen*` 4, `fdta*` 2, `ship_gate*` 2, `smoke*` 1. `verify_v2_2_*` is active and must not be treated as old AE 2022 code.
- `tmp`: 116 JSON files, about 4.0 MB, all under ignored `tmp/`. Registry gates currently pass: `aepregistry audit`, `layout`, and `ownership`; ownership still reports 403 unowned files, which needs triage context rather than blind cleanup.

## Next

Read `design.md`, then execute the survey in four ledgers:

1. Internal package map and ownership boundaries.
2. Knowledge freshness / merge / header-repair audit.
3. JSX generator reference and cleanup audit.
4. `tmp` JSON dependency and cleanup audit.

Do not delete generators, knowledge, or `tmp` files until the relevant ledger marks them safe and the registry/doc/test gates still pass.

## Read now

- design.md
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
- Ran current registry/capindex sanity gates:
  - `go run ./cmd/aepregistry audit -root .` → pass, 0 errors, 0 warnings.
  - `go run ./cmd/aepregistry layout -root .` → 18 locations, 0 cleanup candidates, 0 blocked unowned.
  - `go run ./cmd/aepregistry ownership -root .` → 18 locations, 1677 files, 1274 owned, 403 unowned.
  - `go run ./cmd/capindex -check` → pass.

Current:

- Ready to turn the spec into a staged execution plan after user review.

## Open questions

- Should the next phase update `README.md` architecture prose as a first-class deliverable, or keep the source of truth in Flightdeck knowledge until the cleanup is complete?
- Should unreferenced generator candidates be archived to cold storage first, or deleted after a clean reference/gate pass?
