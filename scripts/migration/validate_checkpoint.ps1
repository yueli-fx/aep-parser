param(
  [string]$CurrentPath = "registry/versioned_aep_migration_current.json",
  [string]$CoveragePath = "registry/versioned_aep_migration_coverage.json",
  [string]$SummaryPath = "tmp/migration_coverage_summary.json",
  [string]$CoverageBatchId = "",
  [switch]$IncludeCoverageBatch,
  [switch]$RunCoverageBatchMatrices,
  [switch]$IncludeMatrix,
  [switch]$IncludeGo
)

$ErrorActionPreference = "Stop"

function Invoke-Step {
  param(
    [string]$Name,
    [scriptblock]$Command
  )

  Write-Output "==> $Name"
  & $Command
  if ($LASTEXITCODE -ne 0) {
    throw "$Name failed with exit code $LASTEXITCODE"
  }
}

$repoRoot = (Resolve-Path ".").Path
$current = Get-Content -Raw $CurrentPath | ConvertFrom-Json
$aeRoot = $current.current_state.ae_install_root
if (-not $CoverageBatchId) {
  $CoverageBatchId = $current.canonical_coverage_batch
}

Invoke-Step "validate current JSON" {
  go run ./cmd/aepregistry current -root . -current $CurrentPath -out tmp/registry_current.json
}

Invoke-Step "validate coverage JSON" {
  go run ./cmd/aepregistry coverage -root . -coverage $CoveragePath -out tmp/registry_coverage.json -require-ledgers
}

Invoke-Step "render coverage summary JSON" {
  go run ./cmd/aepregistry migration-summary -root . -coverage $CoveragePath -out $SummaryPath
}

Invoke-Step "validate coverage summary JSON" {
  go run ./cmd/aepregistry migration-summary -root . -coverage $CoveragePath -out $SummaryPath -check
}

if ($aeRoot) {
  Invoke-Step "plan host-open gaps" {
    go run ./cmd/aepregistry host-open-gaps -root . -coverage $CoveragePath -out tmp/host_open_gap_audit_go.json -ae-root $aeRoot
  }
}

if ($IncludeCoverageBatch) {
  if (-not $CoverageBatchId) {
    throw "CoverageBatchId is required because current JSON has no canonical_coverage_batch."
  }

  $tempCoverage = Join-Path $env:TEMP ("aep-coverage-candidate-" + [guid]::NewGuid() + ".json")
  try {
    if ($RunCoverageBatchMatrices) {
      $batchArgs = @(
        "-File", (Join-Path $repoRoot "scripts/migration/run_coverage_batch.ps1"),
        "-BatchId", $CoverageBatchId,
        "-CurrentPath", $CurrentPath,
        "-CoveragePath", $tempCoverage
      )

      Invoke-Step "replay coverage batch $CoverageBatchId" {
        pwsh @batchArgs
      }
    } else {
      Copy-Item -LiteralPath $CoveragePath -Destination $tempCoverage -Force
      Invoke-Step "replay coverage batch $CoverageBatchId" {
        go run ./cmd/aepregistry coverage-batch -root . -current $CurrentPath -coverage $tempCoverage -out tmp/registry_coverage_batch.json -batch-id $CoverageBatchId -skip-run -sync
      }
    }
  } finally {
    if (Test-Path -LiteralPath $tempCoverage) {
      Remove-Item -LiteralPath $tempCoverage -Force
    }
  }
}

if ($IncludeMatrix) {
  Invoke-Step "verify recurring matrix gates" {
    go run ./cmd/aepregistry recurring-matrix -root . -out tmp/registry_recurring_matrix.json -out-root registry/evidence/versioned-aep-migration/migration_matrix_verify
  }
}

if ($IncludeGo) {
  Invoke-Step "go test" {
    go test ./...
  }

  Invoke-Step "go vet" {
    go vet ./...
  }
}

Invoke-Step "git diff check" {
  git diff --check
}

Write-Output "checkpoint ok"
