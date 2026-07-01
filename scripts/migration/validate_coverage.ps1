param(
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
  [switch]$RequireLedgers
)

$ErrorActionPreference = "Stop"

function Get-MatrixTotals {
  param($Matrix)

  return [pscustomobject]@{
    total = [int]$Matrix.summary.total
    pass = [int]$Matrix.summary.passed
    blocked = [int]$Matrix.summary.blocked
    failed = [int]$Matrix.summary.failed
    skipped = [int]$Matrix.summary.skipped
  }
}

function Assert-Equal {
  param(
    [string]$Label,
    [int]$Expected,
    [int]$Actual
  )

  if ($Expected -ne $Actual) {
    throw "$Label mismatch: coverage=$Expected matrix=$Actual"
  }
}

function Assert-ContainsString {
  param(
    [string]$Label,
    $Values,
    [string]$Expected
  )

  if (@($Values) -notcontains $Expected) {
    throw "$Label missing '$Expected'"
  }
}

$coverage = Get-Content -Raw $CoveragePath | ConvertFrom-Json
$checked = 0
$checkedGates = 0
$missing = @()

if (-not $coverage.host_open_policy) {
  throw "host_open_policy is required"
}

$policy = $coverage.host_open_policy
Assert-ContainsString "host_open_policy.available_flags" $policy.available_flags "-ae-open"
Assert-ContainsString "host_open_policy.available_flags" $policy.available_flags "-ae-versions"
Assert-ContainsString "host_open_policy.available_flags" $policy.available_flags "-max-ae-open-cases"
Assert-ContainsString "host_open_policy.evidence_levels" $policy.evidence_levels "direct_all_hosts"
Assert-ContainsString "host_open_policy.evidence_levels" $policy.evidence_levels "inferred_by_endpoint"
Assert-ContainsString "host_open_policy.evidence_levels" $policy.evidence_levels "pending_per_capability"

if (-not $policy.endpoint_inference.enabled) {
  throw "host_open_policy.endpoint_inference.enabled must be true"
}
if ($policy.endpoint_inference.open_mode -ne "target_bound") {
  throw "host_open_policy.endpoint_inference.open_mode must be target_bound"
}

$knownHosts = @($coverage.host_open_axis.hosts)
foreach ($hostLabel in @($policy.endpoint_inference.direct_hosts + $policy.endpoint_inference.inferred_hosts)) {
  Assert-ContainsString "host_open_policy.endpoint_inference hosts" $knownHosts $hostLabel
}

foreach ($gate in $coverage.recurring_gates) {
  if (-not $gate.artifact) {
    continue
  }
  if (-not (Test-Path $gate.artifact)) {
    $missing += "$($gate.id): $($gate.artifact)"
    continue
  }
  if ($RequireLedgers -and $gate.ledger -and -not (Test-Path $gate.ledger)) {
    $missing += "$($gate.id): $($gate.ledger)"
    continue
  }

  $matrix = Get-Content -Raw $gate.artifact | ConvertFrom-Json
  $matrixTotals = Get-MatrixTotals $matrix

  Assert-Equal "$($gate.id).total" ([int]$gate.totals.total) $matrixTotals.total
  Assert-Equal "$($gate.id).pass" ([int]$gate.totals.pass) $matrixTotals.pass
  Assert-Equal "$($gate.id).blocked" ([int]$gate.totals.blocked) $matrixTotals.blocked
  Assert-Equal "$($gate.id).failed" ([int]$gate.totals.failed) $matrixTotals.failed
  Assert-Equal "$($gate.id).skipped" ([int]$gate.totals.skipped) $matrixTotals.skipped

  $checkedGates++
}

foreach ($record in $coverage.coverage) {
  if (-not $record.artifact) {
    continue
  }
  if (-not (Test-Path $record.artifact)) {
    $missing += "$($record.id): $($record.artifact)"
    continue
  }

  if ($RequireLedgers -and $record.ledger -and -not (Test-Path $record.ledger)) {
    $missing += "$($record.id): $($record.ledger)"
    continue
  }

  $matrix = Get-Content -Raw $record.artifact | ConvertFrom-Json
  $matrixTotals = Get-MatrixTotals $matrix
  $matrixRecipes = @($matrix.cases | ForEach-Object { $_.recipe_name } | Sort-Object -Unique)

  Assert-Equal "$($record.id).total" ([int]$record.totals.total) $matrixTotals.total
  Assert-Equal "$($record.id).pass" ([int]$record.totals.pass) $matrixTotals.pass
  Assert-Equal "$($record.id).blocked" ([int]$record.totals.blocked) $matrixTotals.blocked
  Assert-Equal "$($record.id).failed" ([int]$record.totals.failed) $matrixTotals.failed
  Assert-Equal "$($record.id).skipped" ([int]$record.totals.skipped) $matrixTotals.skipped

  if ($record.PSObject.Properties.Name -contains "recipe_count") {
    Assert-Equal "$($record.id).recipe_count" ([int]$record.recipe_count) $matrixRecipes.Count
  }
  if ($record.PSObject.Properties.Name -contains "recipe_pattern") {
    $pattern = Join-Path "examples/recipes" $record.recipe_pattern
    $repoRecipes = @(Get-ChildItem $pattern | ForEach-Object { $_.BaseName } | Sort-Object -Unique)
    if ($repoRecipes.Count -eq 0) {
      throw "$($record.id).recipe_pattern matched no files: $($record.recipe_pattern)"
    }
    Assert-Equal "$($record.id).recipe_pattern_count" $repoRecipes.Count $matrixRecipes.Count
  }
  if ($record.PSObject.Properties.Name -contains "recipes") {
    $coverageRecipes = @($record.recipes | Sort-Object -Unique)
    $coverageRecipeText = $coverageRecipes -join ","
    $matrixRecipeText = $matrixRecipes -join ","
    if ($coverageRecipeText -ne $matrixRecipeText) {
      throw "$($record.id).recipes mismatch: coverage=$coverageRecipeText matrix=$matrixRecipeText"
    }
  }

  if ($record.PSObject.Properties.Name -contains "host_open_endpoint_evidence") {
    $evidence = $record.host_open_endpoint_evidence
    if ($evidence.open_mode -ne "target_bound") {
      throw "$($record.id).host_open_endpoint_evidence.open_mode must be target_bound"
    }

    $evidenceChunks = @()
    if ($evidence.PSObject.Properties.Name -contains "chunks") {
      $evidenceChunks = @($evidence.chunks)
    } else {
      $evidenceChunks = @([pscustomobject]@{
        artifact = $evidence.artifact
        ledger = $evidence.ledger
        totals = $evidence.totals
      })
    }

    $aggregate = [pscustomobject]@{ total = 0; pass = 0; blocked = 0; failed = 0; skipped = 0 }
    $hostRecipeValues = @()
    foreach ($chunk in $evidenceChunks) {
      if (-not (Test-Path $chunk.artifact)) {
        throw "$($record.id).host_open_endpoint_evidence.artifact not found: $($chunk.artifact)"
      }
      if ($RequireLedgers -and $chunk.ledger -and -not (Test-Path $chunk.ledger)) {
        throw "$($record.id).host_open_endpoint_evidence.ledger not found: $($chunk.ledger)"
      }

      $hostMatrix = Get-Content -Raw $chunk.artifact | ConvertFrom-Json
      $hostTotals = Get-MatrixTotals $hostMatrix
      Assert-Equal "$($record.id).host_open_endpoint_evidence.$($chunk.id).total" ([int]$chunk.totals.total) $hostTotals.total
      Assert-Equal "$($record.id).host_open_endpoint_evidence.$($chunk.id).pass" ([int]$chunk.totals.pass) $hostTotals.pass
      Assert-Equal "$($record.id).host_open_endpoint_evidence.$($chunk.id).blocked" ([int]$chunk.totals.blocked) $hostTotals.blocked
      Assert-Equal "$($record.id).host_open_endpoint_evidence.$($chunk.id).failed" ([int]$chunk.totals.failed) $hostTotals.failed
      Assert-Equal "$($record.id).host_open_endpoint_evidence.$($chunk.id).skipped" ([int]$chunk.totals.skipped) $hostTotals.skipped

      $aggregate.total += $hostTotals.total
      $aggregate.pass += $hostTotals.pass
      $aggregate.blocked += $hostTotals.blocked
      $aggregate.failed += $hostTotals.failed
      $aggregate.skipped += $hostTotals.skipped
      $hostRecipeValues += @($hostMatrix.cases | ForEach-Object { $_.recipe_name })
    }

    $hostRecipes = @($hostRecipeValues | Sort-Object -Unique)
    $evidenceRecipes = @($evidence.recipes | Sort-Object -Unique)

    Assert-Equal "$($record.id).host_open_endpoint_evidence.total" ([int]$evidence.totals.total) $aggregate.total
    Assert-Equal "$($record.id).host_open_endpoint_evidence.pass" ([int]$evidence.totals.pass) $aggregate.pass
    Assert-Equal "$($record.id).host_open_endpoint_evidence.blocked" ([int]$evidence.totals.blocked) $aggregate.blocked
    Assert-Equal "$($record.id).host_open_endpoint_evidence.failed" ([int]$evidence.totals.failed) $aggregate.failed
    Assert-Equal "$($record.id).host_open_endpoint_evidence.skipped" ([int]$evidence.totals.skipped) $aggregate.skipped

    if (($evidenceRecipes -join ",") -ne ($hostRecipes -join ",")) {
      throw "$($record.id).host_open_endpoint_evidence.recipes mismatch: evidence=$($evidenceRecipes -join ',') matrix=$($hostRecipes -join ',')"
    }
  }

  $checked++
}

if ($missing.Count -gt 0) {
  throw "Missing coverage artifacts:`n$($missing -join "`n")"
}

Write-Output "coverage ok: checked $checked matrix artifacts, $checkedGates recurring gates, and host-open policy"
