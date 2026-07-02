param(
  [string]$BatchId = "",

  [string]$CurrentPath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json",
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
  [switch]$SkipRun,
  [switch]$List
)

$ErrorActionPreference = "Stop"

function Get-RecipeArgs {
  param($Entry)

  $recipes = @()
  if ($Entry.recipe_paths) {
    foreach ($recipePath in @($Entry.recipe_paths)) {
      $recipes += Get-Item $recipePath
    }
  } elseif ($Entry.recipe_glob) {
    $recipes = @(Get-ChildItem $Entry.recipe_glob | Sort-Object FullName)
  } else {
    throw "Batch entry '$($Entry.coverage_id)' lacks recipe_glob or recipe_paths."
  }
  if ($recipes.Count -eq 0) {
    throw "No recipes matched batch entry '$($Entry.coverage_id)'"
  }

  $args = @()
  foreach ($recipe in $recipes) {
    $args += @("-recipe", $recipe.FullName)
  }
  return $args
}

function Test-AllowedExit {
  param($Entry, [int]$ExitCode)

  if ($ExitCode -eq 0) {
    return $true
  }
  if ($null -eq $Entry.allow_matrix_exit) {
    return $false
  }
  return @($Entry.allow_matrix_exit) -contains $ExitCode
}

$current = Get-Content -Raw $CurrentPath | ConvertFrom-Json
if ($List) {
  foreach ($knownBatch in $current.coverage_batches) {
    $entryCount = @($knownBatch.entries).Count
    Write-Output "$($knownBatch.id): $entryCount entries"
    if ($knownBatch.description) {
      Write-Output "  $($knownBatch.description)"
    }
  }
  exit 0
}

if (-not $BatchId) {
  throw "BatchId is required unless -List is used."
}

$batch = $current.coverage_batches | Where-Object { $_.id -eq $BatchId } | Select-Object -First 1
if (-not $batch) {
  $known = @($current.coverage_batches | ForEach-Object { $_.id }) -join ", "
  throw "Coverage batch '$BatchId' not found. Known batches: $known"
}

if ($SkipRun) {
  Write-Output "skip-run: using existing matrix artifacts for batch '$BatchId'"
} else {
  foreach ($entry in $batch.entries) {
    if (-not $entry.out -or -not $entry.ledger) {
      throw "Batch entry '$($entry.coverage_id)' lacks out/ledger."
    }

    $recipeArgs = Get-RecipeArgs $entry
    $matrixArgs = @("./cmd/aepmigrate", "matrix") +
      $recipeArgs +
      @("-sources", "all", "-targets", "all", "-out", $entry.out, "-ledger-out", $entry.ledger)

    & go run @matrixArgs
    $exitCode = $LASTEXITCODE
    if (-not (Test-AllowedExit $entry $exitCode)) {
      throw "Matrix run failed for '$($entry.coverage_id)' with exit code $exitCode"
    }
    if ($exitCode -ne 0) {
      Write-Output "matrix for '$($entry.coverage_id)' exited $exitCode; allowed: $($entry.allowed_nonpass_reason)"
    }
  }
}

if (-not (Test-Path -LiteralPath $CoveragePath)) {
  $sourceCoverage = $current.truth_sources.coverage
  if (-not $sourceCoverage) {
    throw "CoveragePath '$CoveragePath' does not exist and current JSON has no truth_sources.coverage"
  }
  if (-not (Test-Path -LiteralPath $sourceCoverage)) {
    throw "CoveragePath '$CoveragePath' does not exist and source coverage not found: $sourceCoverage"
  }
  $coverageDir = Split-Path -Parent $CoveragePath
  if ($coverageDir -and -not (Test-Path -LiteralPath $coverageDir)) {
    New-Item -ItemType Directory -Path $coverageDir | Out-Null
  }
  Copy-Item -LiteralPath $sourceCoverage -Destination $CoveragePath
  Write-Output "seeded coverage candidate from '$sourceCoverage': $CoveragePath"
}

$sync = Join-Path (Resolve-Path ".").Path "scripts/migration/sync_coverage_batch.ps1"
pwsh -File $sync -BatchId $BatchId -CurrentPath $CurrentPath -CoveragePath $CoveragePath
if ($LASTEXITCODE -ne 0) {
  throw "Coverage sync failed for batch '$BatchId' with exit code $LASTEXITCODE"
}

$validatorArgs = @(
  "run", "./cmd/aepregistry", "coverage",
  "-root", ".",
  "-coverage", $CoveragePath,
  "-out", "tmp/registry_coverage.json",
  "-require-ledgers"
)
go @validatorArgs
if ($LASTEXITCODE -ne 0) {
  throw "Coverage validation failed for batch '$BatchId' with exit code $LASTEXITCODE"
}

Write-Output "coverage batch '$BatchId' complete"
