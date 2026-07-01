param(
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
  [string]$SummaryPath = "tmp/migration_coverage_summary.json"
)

$ErrorActionPreference = "Stop"

function Assert-Equal {
  param(
    [string]$Label,
    $Expected,
    $Actual
  )

  if ("$Expected" -ne "$Actual") {
    throw "$Label mismatch: expected=$Expected actual=$Actual"
  }
}

function Assert-SameStringSet {
  param(
    [string]$Label,
    $Expected,
    $Actual
  )

  $expectedText = @($Expected | Sort-Object -Unique) -join ","
  $actualText = @($Actual | Sort-Object -Unique) -join ","
  if ($expectedText -ne $actualText) {
    throw "$Label mismatch: expected=[$expectedText] actual=[$actualText]"
  }
}

function New-EmptyTotals {
  return [pscustomobject]@{
    total = 0
    pass = 0
    blocked = 0
    failed = 0
    skipped = 0
  }
}

function Add-Totals {
  param(
    $Accumulator,
    $Totals
  )

  $Accumulator.total += [int]$Totals.total
  $Accumulator.pass += [int]$Totals.pass
  $Accumulator.blocked += [int]$Totals.blocked
  $Accumulator.failed += [int]$Totals.failed
  $Accumulator.skipped += [int]$Totals.skipped
}

function Get-RecipeNames {
  param($Record)

  if ($Record.PSObject.Properties.Name -contains "recipes") {
    return @($Record.recipes | Sort-Object -Unique)
  }

  if ($Record.PSObject.Properties.Name -contains "recipe_pattern") {
    $pattern = Join-Path "examples/recipes" $Record.recipe_pattern
    return @(Get-ChildItem $pattern | ForEach-Object { $_.BaseName } | Sort-Object -Unique)
  }

  return @()
}

function Assert-Totals {
  param(
    [string]$Label,
    $Expected,
    $Actual
  )

  Assert-Equal "$Label.total" ([int]$Expected.total) ([int]$Actual.total)
  Assert-Equal "$Label.pass" ([int]$Expected.pass) ([int]$Actual.pass)
  Assert-Equal "$Label.blocked" ([int]$Expected.blocked) ([int]$Actual.blocked)
  Assert-Equal "$Label.failed" ([int]$Expected.failed) ([int]$Actual.failed)
  Assert-Equal "$Label.skipped" ([int]$Expected.skipped) ([int]$Actual.skipped)
}

if (-not (Test-Path -LiteralPath $CoveragePath)) {
  throw "coverage not found: $CoveragePath"
}
if (-not (Test-Path -LiteralPath $SummaryPath)) {
  throw "summary not found: $SummaryPath"
}

$coverage = Get-Content -Raw $CoveragePath | ConvertFrom-Json
$summary = Get-Content -Raw $SummaryPath | ConvertFrom-Json

Assert-Equal "schema_version" 1 $summary.schema_version
Assert-Equal "generated_from" $CoveragePath $summary.generated_from
Assert-SameStringSet "axes.source_writers" $coverage.writer_axes.source_writers $summary.axes.source_writers
Assert-SameStringSet "axes.target_writers" $coverage.writer_axes.target_writers $summary.axes.target_writers
Assert-SameStringSet "axes.host_open_hosts" $coverage.host_open_axis.hosts $summary.axes.host_open_hosts

$writerTotals = New-EmptyTotals
$endpointHostTotals = New-EmptyTotals
$allRecipes = [System.Collections.Generic.HashSet[string]]::new()
$domainMap = @{}
$boundaryCount = 0

foreach ($record in @($coverage.coverage)) {
  $recipes = @(Get-RecipeNames $record)
  foreach ($recipe in $recipes) {
    [void]$allRecipes.Add($recipe)
  }

  Add-Totals $writerTotals $record.totals

  if (-not $domainMap.ContainsKey($record.domain)) {
    $domainMap[$record.domain] = [pscustomobject]@{
      coverage_ids = @()
      recipe_count = 0
      writer_totals = New-EmptyTotals
      writer_statuses = @()
      host_open_statuses = @()
      boundary_ids = @()
    }
  }

  $domain = $domainMap[$record.domain]
  $domain.coverage_ids = @($domain.coverage_ids + $record.id)
  $domain.recipe_count += $recipes.Count
  Add-Totals $domain.writer_totals $record.totals
  $domain.writer_statuses = @($domain.writer_statuses + $record.writer_status | Sort-Object -Unique)
  $domain.host_open_statuses = @($domain.host_open_statuses + $record.host_open_status | Sort-Object -Unique)

  if ($record.PSObject.Properties.Name -contains "boundary") {
    $boundaryCount++
    $domain.boundary_ids = @($domain.boundary_ids + $record.id | Sort-Object -Unique)
  }

  if ($record.PSObject.Properties.Name -contains "host_open_endpoint_evidence") {
    Add-Totals $endpointHostTotals $record.host_open_endpoint_evidence.totals
  }
}

Assert-Equal "totals.coverage_records" @($coverage.coverage).Count $summary.totals.coverage_records
Assert-Equal "totals.domains" $domainMap.Keys.Count $summary.totals.domains
Assert-Equal "totals.recipes" $allRecipes.Count $summary.totals.recipes
Assert-Totals "totals.writer_cases" $writerTotals $summary.totals.writer_cases
Assert-Totals "totals.endpoint_host_open_cases" $endpointHostTotals $summary.totals.endpoint_host_open_cases
Assert-Equal "totals.known_boundaries" $boundaryCount $summary.totals.known_boundaries
Assert-Equal "totals.open_items" @($coverage.open_items).Count $summary.totals.open_items

foreach ($domainName in $domainMap.Keys) {
  $expected = $domainMap[$domainName]
  $actual = @($summary.domain_rollup | Where-Object { $_.domain -eq $domainName })
  Assert-Equal "domain_rollup.$domainName.count" 1 $actual.Count
  Assert-SameStringSet "domain_rollup.$domainName.coverage_ids" $expected.coverage_ids $actual[0].coverage_ids
  Assert-Equal "domain_rollup.$domainName.recipe_count" $expected.recipe_count $actual[0].recipe_count
  Assert-Totals "domain_rollup.$domainName.writer_totals" $expected.writer_totals $actual[0].writer_totals
  Assert-SameStringSet "domain_rollup.$domainName.writer_statuses" $expected.writer_statuses $actual[0].writer_statuses
  Assert-SameStringSet "domain_rollup.$domainName.host_open_statuses" $expected.host_open_statuses $actual[0].host_open_statuses
  Assert-SameStringSet "domain_rollup.$domainName.boundary_ids" $expected.boundary_ids $actual[0].boundary_ids
}

foreach ($record in @($coverage.coverage)) {
  $recipes = @(Get-RecipeNames $record)
  foreach ($recipe in $recipes) {
    $actual = @($summary.recipe_index | Where-Object { $_.recipe -eq $recipe })
    Assert-Equal "recipe_index.$recipe.count" 1 $actual.Count
    Assert-Equal "recipe_index.$recipe.coverage_id" $record.id $actual[0].coverage_id
    Assert-Equal "recipe_index.$recipe.domain" $record.domain $actual[0].domain
    Assert-Equal "recipe_index.$recipe.writer_status" $record.writer_status $actual[0].writer_status
    Assert-SameStringSet "recipe_index.$recipe.writer_sources" $coverage.writer_axes.source_writers $actual[0].writer_sources
    Assert-SameStringSet "recipe_index.$recipe.writer_targets" $coverage.writer_axes.target_writers $actual[0].writer_targets
    Assert-Equal "recipe_index.$recipe.host_open_status" $record.host_open_status $actual[0].host_open_status

    $expectedBoundaryStatus = "none"
    if ($record.PSObject.Properties.Name -contains "boundary" -and @($record.boundary.blocked_recipe_ids) -contains $recipe) {
      $expectedBoundaryStatus = $record.boundary.status
    }
    Assert-Equal "recipe_index.$recipe.boundary_status" $expectedBoundaryStatus $actual[0].boundary_status
  }
}

Write-Output "coverage summary ok: checked $($summary.totals.coverage_records) records, $($summary.totals.domains) domains, and $($summary.totals.recipes) recipes"
