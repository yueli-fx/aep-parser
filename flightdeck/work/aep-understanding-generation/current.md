# Current — aep-understanding-generation

Versioned AEP migration state is now JSON-first.

Read:

- `versioned-aep-migration-current.json` — active state, batches, rules, frozen historical docs.
- `versioned-aep-migration-coverage.json` — machine-readable coverage ledger.

Use:

```powershell
go run ./cmd/aepregistry current -root . -current flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json -out tmp/registry_current.json
go run ./cmd/aepregistry migration-summary -root . -out tmp/migration_coverage_summary.json
go run ./cmd/aepregistry migration-summary -root . -out tmp/migration_coverage_summary.json -check
go run ./cmd/aepregistry migration-summary -root . -out tmp/migration_coverage_summary.json -totals
go run ./cmd/aepregistry host-open-gaps -root . -out tmp/host_open_gap_audit_go.json -ae-root E:/adobe
go run ./cmd/aepregistry coverage-batch -root . -current flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json -out tmp/registry_coverage_batches.json -list
go run ./cmd/aepregistry coverage-batch -root . -current flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json -coverage flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json -out tmp/registry_coverage_batch.json -batch-id all-current-writer-coverage -skip-run
go run ./cmd/aepregistry coverage -root . -coverage flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json -out tmp/registry_coverage.json -require-ledgers
pwsh -File scripts\migration\render_coverage_md.ps1
go run ./cmd/aepregistry recurring-matrix -root . -out tmp/registry_recurring_matrix.json -out-root tmp/migration_matrix_verify
go test ./...
go vet ./...
```

Rules:

- Update JSON truth sources, not long Markdown plans.
- Treat old matrix/blocker plans as frozen historical context.
- Batch matrix runs; do not churn cockpit, index, or history for each slice.
