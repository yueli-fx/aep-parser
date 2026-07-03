# Migration Scripts

`scripts/migration/` contains shell and PowerShell automation around AEP version
migration coverage.

## What Belongs Here

- Batch runners for migration coverage and matrix checks.
- Helpers that render, validate, or summarize migration coverage artifacts.
- Scripts that coordinate `cmd/aepmigrate`, registry coverage files, and AE host
  checks.

## What Does Not Belong Here

- Core conversion logic. Use `internal/aepmigrate`.
- Stable user-facing command entry points. Use `cmd/aepmigrate`.
- One-off migration probes or temporary reports. Use `tmp/` until the result is
  promoted.

Keep scripts explicit about their expected input and output paths so coverage
artifacts can be audited later.
