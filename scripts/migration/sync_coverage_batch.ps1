param(
  [Parameter(Mandatory=$true)]
  [string]$BatchId,

  [string]$CurrentPath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json",
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json"
)

$ErrorActionPreference = "Stop"

$current = Get-Content -Raw $CurrentPath | ConvertFrom-Json
$batch = $current.coverage_batches | Where-Object { $_.id -eq $BatchId } | Select-Object -First 1
if (-not $batch) {
  $known = @($current.coverage_batches | ForEach-Object { $_.id }) -join ", "
  throw "Coverage batch '$BatchId' not found. Known batches: $known"
}

$updater = Join-Path (Resolve-Path ".").Path "scripts/migration/update_coverage.ps1"
foreach ($entry in $batch.entries) {
  if (-not (Test-Path $entry.matrix)) {
    throw "Matrix artifact not found for '$($entry.coverage_id)': $($entry.matrix)"
  }

  pwsh -File $updater `
    -CoveragePath $CoveragePath `
    -Id $entry.coverage_id `
    -MatrixPath $entry.matrix
  if ($LASTEXITCODE -ne 0) {
    throw "Coverage update failed for '$($entry.coverage_id)' with exit code $LASTEXITCODE"
  }
}

Write-Output "synced coverage batch '$BatchId'"
