# Index — arbitrary-aep-version-migration

## State

Parked future work. This package tracks the long-term goal of upgrading or
downgrading arbitrary After Effects projects across supported AE writer
versions.

The current codebase already has `internal/aepmigrate`, migration coverage
registry files, and migration scripts. This package is the planning and intake
ledger for the larger product goal: take an arbitrary `.aep`, understand the
source/target version boundary, preserve what can be preserved, and report what
must be degraded, dropped, rebuilt, or left untouched.

## Goal

- Upgrade lower-version projects into newer supported writer targets.
- Downgrade newer projects into older supported writer targets when the target
  AE version can represent the behavior.
- Detect high-version-only fields, chunks, effects, or semantics.
- Produce explicit compatibility findings instead of silently losing data.
- Feed confirmed version differences back into migration coverage and durable
  knowledge.

## Intake Rule

When work on another task discovers a version boundary, add it here if it
affects arbitrary project migration. Typical examples:

- A field exists only in AE 24+ or another high-version writer.
- The same visible behavior is encoded differently between AE versions.
- A chunk can be preserved but not interpreted by the current scene model.
- A downgrade needs a fallback, warning, or explicit unsupported marker.
- A fixture proves a version-specific open/save/render behavior.

Keep raw dumps in `tmp/` until the finding is promoted. Durable evidence belongs
in `registry/evidence/<topic>/`, `data/reference/<topic>/`, committed fixtures,
or a focused knowledge note.

## Related Code And Data

- `internal/aepmigrate/` — version-aware conversion and migration policy.
- `cmd/aepmigrate` — migration command entry point.
- `scripts/migration/` — batch coverage and matrix automation.
- `registry/version_boundaries.json` — known version boundary facts.
- `registry/versioned_aep_migration_current.json` — current migration state.
- `registry/versioned_aep_migration_coverage.json` — coverage state.

## Next

Do not start this as active implementation until selected. When resumed, first
define a version-boundary ledger format and decide how arbitrary-project
findings move between fixtures, registry evidence, and user-facing migration
reports.

## Open Intake

- Empty. Add future high-version-only fields, field layout differences, and
  downgrade fallback findings here before turning them into implementation
  slices.
