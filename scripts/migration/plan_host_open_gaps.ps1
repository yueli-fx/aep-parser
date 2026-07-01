param(
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
  [string]$Out = "tmp/host_open_gap_audit.json",
  [string]$AERoot = "",
  [int]$MaxAEOpenCases = 24
)

$ErrorActionPreference = "Stop"

function Get-RecipeNames {
  param($Record)

  if ($Record.PSObject.Properties.Name -contains "recipes") {
    return @($Record.recipes)
  }

  if ($Record.PSObject.Properties.Name -contains "recipe_pattern") {
    $pattern = Join-Path "examples/recipes" $Record.recipe_pattern
    return @(Get-ChildItem $pattern | ForEach-Object { $_.BaseName } | Sort-Object)
  }

  return @()
}

function Split-IntoChunks {
  param(
    [string[]]$Items,
    [int]$ChunkSize
  )

  $chunks = @()
  for ($i = 0; $i -lt $Items.Count; $i += $ChunkSize) {
    $end = [Math]::Min($i + $ChunkSize - 1, $Items.Count - 1)
    $chunks += ,@($Items[$i..$end])
  }
  return $chunks
}

function Get-RecipePathArgs {
  param([string[]]$RecipeNames)

  return @($RecipeNames | ForEach-Object { "-recipe examples/recipes/$_.json" })
}

$coverage = Get-Content -Raw $CoveragePath | ConvertFrom-Json
$policy = $coverage.host_open_policy
if (-not $policy) {
  throw "host_open_policy missing from $CoveragePath"
}

$directHosts = @($policy.endpoint_inference.direct_hosts)
if ($directHosts.Count -eq 0) {
  throw "host_open_policy.endpoint_inference.direct_hosts is empty"
}

$sources = $directHosts
$targets = $directHosts
$aeOpenMode = "target_bound"
$casesPerRecipe = $sources.Count * $targets.Count
if ($casesPerRecipe -le 0) {
  throw "invalid endpoint case calculation"
}

$recipesPerChunk = [Math]::Max(1, [Math]::Floor($MaxAEOpenCases / $casesPerRecipe))
$gapRecords = @()

foreach ($record in $coverage.coverage) {
  $status = [string]$record.host_open_status
  $allRecipes = @(Get-RecipeNames $record)
  $representatives = @()
  if ($record.PSObject.Properties.Name -contains "host_open_representatives") {
    $representatives = @($record.host_open_representatives)
  }

  $knownBoundaryRecipes = @()
  if ($record.PSObject.Properties.Name -contains "boundary") {
    $knownBoundaryRecipes = @($record.boundary.blocked_recipe_ids)
  }

  $gapRecipes = @()
  $classification = "covered_or_not_required"
  $recommendation = "no_action"

  if ($status -eq "pending_per_capability") {
    $gapRecipes = $allRecipes
    $classification = "direct_endpoint_gap"
    $recommendation = "run_endpoint_direct_host_open"
  } elseif ($status -match "pending for others") {
    $gapRecipes = @($allRecipes | Where-Object { $representatives -notcontains $_ })
    $classification = "partial_representative_gap"
    $recommendation = "run_endpoint_direct_host_open_for_non_representatives"
  } elseif ($status -eq "representative checks only") {
    $gapRecipes = @($allRecipes | Where-Object { $knownBoundaryRecipes -notcontains $_ })
    $classification = "representative_only_gap"
    $recommendation = "run_endpoint_direct_host_open_for_non_boundary_recipes"
  }

  if ($gapRecipes.Count -eq 0) {
    continue
  }

  $chunks = @()
  $chunkIndex = 1
  foreach ($chunkRecipes in @(Split-IntoChunks -Items $gapRecipes -ChunkSize $recipesPerChunk)) {
    $outRoot = "tmp/host_open_gaps/$($record.id)/chunk-$chunkIndex"
    $recipeArgs = Get-RecipePathArgs -RecipeNames $chunkRecipes
    $commandParts = @(
      "go run ./cmd/aepmigrate matrix"
    ) + $recipeArgs

    if ($AERoot) {
      $commandParts += "-ae-root $AERoot"
    }

    $commandParts += @(
      "-sources $($sources -join ',')",
      "-targets $($targets -join ',')",
      "-ae-open",
      "-max-ae-open-cases $MaxAEOpenCases",
      "-out $outRoot",
      "-ledger-out $outRoot/ledger.md"
    )

    $chunks += [pscustomobject]@{
      id = "chunk-$chunkIndex"
      recipes = $chunkRecipes
      expected_ae_open_cases = $chunkRecipes.Count * $casesPerRecipe
      command = $commandParts -join " "
      matrix = "$outRoot/matrix.json"
      ledger = "$outRoot/ledger.md"
    }
    $chunkIndex++
  }

  $gapRecords += [pscustomobject]@{
    coverage_id = $record.id
    domain = $record.domain
    host_open_status = $status
    classification = $classification
    recommendation = $recommendation
    endpoint_direct_hosts = $directHosts
    inferred_hosts_after_endpoint_pass = @($policy.endpoint_inference.inferred_hosts)
    total_recipes = $allRecipes.Count
    representative_recipes = $representatives
    excluded_known_boundary_recipes = $knownBoundaryRecipes
    gap_recipes = $gapRecipes
    chunks = $chunks
  }
}

$audit = [pscustomobject]@{
  schema_version = 1
  generated_at = (Get-Date -Format "yyyy-MM-dd")
  coverage_source = $CoveragePath
  host_open_policy = [pscustomobject]@{
    matrix_command_status = $policy.matrix_command_status
    default_strategy = $policy.default_strategy
    broad_fanout_status = $policy.broad_fanout_status
    endpoint_label = $policy.endpoint_inference.label
    direct_hosts = $directHosts
    inferred_hosts = @($policy.endpoint_inference.inferred_hosts)
  }
  planner = [pscustomobject]@{
    max_ae_open_cases = $MaxAEOpenCases
    ae_root = $AERoot
    sources = $sources
    targets = $targets
    ae_open_mode = $aeOpenMode
    ae_open_hosts = $targets
    cases_per_recipe = $casesPerRecipe
    recipes_per_chunk = $recipesPerChunk
  }
  summary = [pscustomobject]@{
    gap_groups = $gapRecords.Count
    gap_recipes = ($gapRecords | ForEach-Object { $_.gap_recipes } | Measure-Object).Count
    chunks = ($gapRecords | ForEach-Object { $_.chunks } | Measure-Object).Count
  }
  gaps = $gapRecords
}

$outDir = Split-Path -Parent $Out
if ($outDir -and -not (Test-Path $outDir)) {
  New-Item -ItemType Directory -Path $outDir | Out-Null
}

$audit | ConvertTo-Json -Depth 20 | Set-Content -Path $Out -Encoding utf8
Write-Output "host-open gap audit written: $Out"
Write-Output "gap groups=$($audit.summary.gap_groups) gap recipes=$($audit.summary.gap_recipes) chunks=$($audit.summary.chunks)"
