param(
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
  [string]$Out = "tmp/migration_coverage_summary.json"
)

$ErrorActionPreference = "Stop"

function New-EmptyTotals {
  return [ordered]@{
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

function Get-HostOpenSummary {
  param($Record)

  $summary = [ordered]@{
    status = $Record.host_open_status
  }

  if ($Record.PSObject.Properties.Name -contains "host_open_endpoint_evidence") {
    $evidence = $Record.host_open_endpoint_evidence
    $summary.evidence_level = "direct_endpoint_hosts"
    $summary.open_mode = $evidence.open_mode
    $summary.direct_hosts = @($evidence.direct_hosts)
    $summary.inferred_hosts = @($evidence.inferred_hosts)
    $summary.totals = $evidence.totals
    if ($evidence.PSObject.Properties.Name -contains "excluded_known_boundary_recipes") {
      $summary.excluded_known_boundary_recipes = @($evidence.excluded_known_boundary_recipes)
    }
    return $summary
  }

  if ($Record.PSObject.Properties.Name -contains "host_open_representatives") {
    $summary.evidence_level = "representative"
    $summary.representatives = @($Record.host_open_representatives)
    return $summary
  }

  $summary.evidence_level = "recorded_status_only"
  return $summary
}

$coverage = Get-Content -Raw $CoveragePath | ConvertFrom-Json
$writerTotals = New-EmptyTotals
$endpointHostTotals = New-EmptyTotals
$domainMap = [ordered]@{}
$records = @()
$recipeIndex = @()
$allRecipes = [System.Collections.Generic.HashSet[string]]::new()
$boundaryRecords = @()

foreach ($record in @($coverage.coverage)) {
  $recipes = @(Get-RecipeNames $record)
  foreach ($recipe in $recipes) {
    [void]$allRecipes.Add($recipe)
  }

  Add-Totals $writerTotals $record.totals

  if (-not $domainMap.Contains($record.domain)) {
    $domainMap[$record.domain] = [ordered]@{
      domain = $record.domain
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

  $recordSummary = [ordered]@{
    id = $record.id
    domain = $record.domain
    scope = $record.scope
    recipes = $recipes
    writer = [ordered]@{
      status = $record.writer_status
      sources = @($coverage.writer_axes.source_writers)
      targets = @($coverage.writer_axes.target_writers)
      coverage = $record.writer_coverage
      totals = $record.totals
      artifact = $record.artifact
    }
    host_open = Get-HostOpenSummary $record
  }

  if ($record.PSObject.Properties.Name -contains "boundary") {
    $boundary = [ordered]@{
      id = $record.id
      domain = $record.domain
      status = $record.boundary.status
      blocked_recipe_ids = @($record.boundary.blocked_recipe_ids)
      details = @($record.boundary.details)
    }
    $recordSummary.boundary = $boundary
    $boundaryRecords += $boundary
    $domain.boundary_ids = @($domain.boundary_ids + $record.id | Sort-Object -Unique)
  }

  foreach ($recipe in $recipes) {
    $boundaryStatus = "none"
    if ($record.PSObject.Properties.Name -contains "boundary" -and @($record.boundary.blocked_recipe_ids) -contains $recipe) {
      $boundaryStatus = $record.boundary.status
    }

    $recipeIndex += [pscustomobject][ordered]@{
      recipe = $recipe
      coverage_id = $record.id
      domain = $record.domain
      writer_status = $record.writer_status
      writer_sources = @($coverage.writer_axes.source_writers)
      writer_targets = @($coverage.writer_axes.target_writers)
      host_open_status = $record.host_open_status
      boundary_status = $boundaryStatus
    }
  }

  if ($record.host_open_endpoint_evidence) {
    Add-Totals $endpointHostTotals $record.host_open_endpoint_evidence.totals
  }

  $records += [pscustomobject]$recordSummary
}

$domainRollup = @(
  foreach ($key in $domainMap.Keys) {
    [pscustomobject]$domainMap[$key]
  }
)

$summary = [ordered]@{
  schema_version = 1
  generated_from = $CoveragePath
  coverage_generated_at = $coverage.generated_at
  axes = [ordered]@{
    source_writers = @($coverage.writer_axes.source_writers)
    target_writers = @($coverage.writer_axes.target_writers)
    host_open_hosts = @($coverage.host_open_axis.hosts)
  }
  totals = [ordered]@{
    coverage_records = @($coverage.coverage).Count
    domains = @($domainMap.Keys).Count
    recipes = $allRecipes.Count
    writer_cases = $writerTotals
    endpoint_host_open_cases = $endpointHostTotals
    known_boundaries = @($boundaryRecords).Count
    open_items = @($coverage.open_items).Count
  }
  host_open_policy = $coverage.host_open_policy
  recurring_gates = @($coverage.recurring_gates)
  domain_rollup = $domainRollup
  recipe_index = @($recipeIndex | Sort-Object recipe)
  coverage = $records
  boundaries = @($boundaryRecords)
  open_items = @($coverage.open_items)
}

$outDir = Split-Path -Parent $Out
if ($outDir -and -not (Test-Path $outDir)) {
  New-Item -ItemType Directory -Path $outDir | Out-Null
}

$summary | ConvertTo-Json -Depth 100 | Set-Content -Path $Out -Encoding utf8
Write-Output "rendered coverage summary json: $Out"
