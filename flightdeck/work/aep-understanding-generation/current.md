# Current — aep-understanding-generation

Versioned AEP migration state is now JSON-first.

Read:

- `versioned-aep-migration-current.json` — active state, batches, rules, frozen historical docs.
- `versioned-aep-migration-coverage.json` — machine-readable coverage ledger.

Use:

```powershell
pwsh -File scripts\migration\validate_current.ps1
pwsh -File scripts\migration\plan_host_open_gaps.ps1 -AERoot E:\adobe
pwsh -File scripts\migration\run_coverage_batch.ps1 -List
pwsh -File scripts\migration\run_coverage_batch.ps1 -BatchId all-current-writer-coverage -SkipRun
pwsh -File scripts\migration\validate_coverage.ps1
pwsh -File scripts\migration\render_coverage_md.ps1
pwsh -File scripts\migration\verify_matrix.ps1
go test ./...
go vet ./...
```

Rules:

- Update JSON truth sources, not long Markdown plans.
- Treat old matrix/blocker plans as frozen historical context.
- Batch matrix runs; do not churn cockpit, index, or history for each slice.
