param(
  [string]$CurrentPath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json",
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
  [string]$SummaryPath = "tmp/migration_coverage_summary.json",
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

Invoke-Step "validate current JSON" {
  pwsh -File (Join-Path $repoRoot "scripts/migration/validate_current.ps1") -CurrentPath $CurrentPath
}

Invoke-Step "validate coverage JSON" {
  pwsh -File (Join-Path $repoRoot "scripts/migration/validate_coverage.ps1") -CoveragePath $CoveragePath
}

Invoke-Step "render coverage summary JSON" {
  pwsh -File (Join-Path $repoRoot "scripts/migration/render_coverage_summary_json.ps1") -CoveragePath $CoveragePath -Out $SummaryPath
}

Invoke-Step "validate coverage summary JSON" {
  pwsh -File (Join-Path $repoRoot "scripts/migration/validate_coverage_summary_json.ps1") -CoveragePath $CoveragePath -SummaryPath $SummaryPath
}

if ($aeRoot) {
  Invoke-Step "plan host-open gaps" {
    pwsh -File (Join-Path $repoRoot "scripts/migration/plan_host_open_gaps.ps1") -AERoot $aeRoot
  }
}

if ($IncludeMatrix) {
  Invoke-Step "verify recurring matrix gates" {
    pwsh -File (Join-Path $repoRoot "scripts/migration/verify_matrix.ps1")
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
