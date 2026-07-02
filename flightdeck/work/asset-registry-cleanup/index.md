# Index — asset-registry-cleanup

## State

This package owns project data placement, registry ledgers, generated-output
boundaries, and cleanup policy across `data/`, `examples/`, `test_data/`,
`tmp/`, `registry/`, and package-local JSON state.

It exists because the project now has enough fixtures, generated artifacts, and
matrix outputs that cleanup must be governed by explicit ownership instead of
ad-hoc deletion.

## Next

When resumed, build a single machine-readable ownership ledger that answers:

- which files are source fixtures vs generated artifacts;
- which generated files are safe to delete or regenerate;
- which local-only files must remain outside git;
- which package owns each long-lived JSON truth source.

## Read now

- `../../briefing.md` — project conventions, especially `data/` being valid
  test data.
- `../../../registry/locations.json` — current location registry.
- `../../../registry/evidence.json` — current evidence artifact registry.
- `../versioned-aep-migration/mainline-spec.json` — current mainline and asset
  policy source until this package gets its own JSON spec.

## Read if

- `../versioned-aep-migration/versioned-aep-migration-current.json` — when
  checking generated migration outputs referenced by current state.
- `../versioned-aep-migration/versioned-aep-migration-coverage.json` — when
  checking long-lived coverage evidence artifacts.

## Progress

Done:

- Dedicated package created; registry and generated-output cleanup is no longer
  hidden inside the migration package.

Current:

- No cleanup execution is selected. Do not delete data from this package without
  first recording ownership and regeneration rules.

## Open questions

- Whether `mainline-spec.json` should be split so asset policy becomes a
  first-class JSON truth source owned here instead of a section inside the
  versioned migration package.
