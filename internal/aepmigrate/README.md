# AEP Migration Package

`internal/aepmigrate` converts parsed AEP profiles and projects across supported
After Effects writer-version targets and reports migration coverage.

## Contents

- `convert_*.go` maps profiles, layers, effects, shapes, text, and properties
  into rebuildable project state.
- `matrix_*.go` builds and runs migration matrix cases.
- `verify*.go` checks converted output and reports acceptance state.
- `version*.go` and `version_capability_ledger.json` describe supported version
  capabilities.

## What Belongs Here

- Version-aware conversion logic and migration policy.
- Coverage ledgers and reports used by migration commands.
- Tests for rebuild fidelity, version boundaries, and matrix classification.

## What Does Not Belong Here

- Generic public API constructors or setters. Use `internal/aep` and
  `internal/scene`.
- Fixture JSX. Use `test_data/generators/`.
- Shell/PowerShell orchestration for batches. Use `scripts/migration/` or
  `cmd/aepmigrate`.
