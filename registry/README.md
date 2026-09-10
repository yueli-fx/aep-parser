# Registry

`registry/` is the machine-readable governance and truth layer for capabilities,
asset ownership, evidence, workflows, and version coverage.

## Contents

- `capability_atoms.json` defines capability atoms and their public coverage.
- `locations.json` tracks owned repo locations and cleanup responsibility.
- `evidence.json` indexes durable evidence references.
- `asset_policy.json` and `workflows.json` encode cleanup and process policy.
- `version_boundaries.json`, `versioned_aep_migration_current.json`, and
  `versioned_aep_migration_coverage.json` track migration support state.
- `evidence/` contains durable, topic-scoped evidence artifacts.

## What Belongs Here

- Structured facts that commands can audit, diff, or gate.
- Durable evidence that should survive beyond a local investigation.
- Ownership and cleanup policy for repo assets.
- Version/migration coverage state that must be reviewed by automation.

## What Does Not Belong Here

- Human-only planning drafts. Keep these in the maintainer’s work desk, outside the public documentation.
- Temporary dumps or scratch command output. Use `tmp/`.
- Public prose documentation generated from source comments. Use `docs/`.

Prefer `go run ./cmd/aepregistry ...` when updating or auditing this directory.
