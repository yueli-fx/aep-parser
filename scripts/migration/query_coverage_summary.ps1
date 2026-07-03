param(
  [string]$CoveragePath = "registry/versioned_aep_migration_coverage.json",
  [string]$SummaryPath = "tmp/migration_coverage_summary.json",
  [string]$Domain = "",
  [string]$CoverageId = "",
  [string]$Recipe = "",
  [string]$EvidenceLevel = "",
  [switch]$Totals,
  [switch]$Refresh,
  [switch]$SkipValidate
)

$ErrorActionPreference = "Stop"

if ($Refresh -or -not (Test-Path -LiteralPath $SummaryPath)) {
  $renderer = Join-Path (Resolve-Path ".").Path "scripts/migration/render_coverage_summary_json.ps1"
  pwsh -File $renderer -CoveragePath $CoveragePath -Out $SummaryPath | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "coverage summary render failed with exit code $LASTEXITCODE"
  }
}

if (-not $SkipValidate) {
  $validator = Join-Path (Resolve-Path ".").Path "scripts/migration/validate_coverage_summary_json.ps1"
  pwsh -File $validator -CoveragePath $CoveragePath -SummaryPath $SummaryPath | Out-Null
  if ($LASTEXITCODE -ne 0) {
    throw "coverage summary validation failed with exit code $LASTEXITCODE"
  }
}

$summary = Get-Content -Raw $SummaryPath | ConvertFrom-Json

if ($Totals) {
  $summary.totals | ConvertTo-Json -Depth 100
  exit 0
}

if ($Recipe) {
  $matches = @($summary.recipe_index | Where-Object { $_.recipe -eq $Recipe })
  if ($matches.Count -eq 0) {
    throw "recipe not found in coverage summary: $Recipe"
  }
  $matches | ConvertTo-Json -Depth 100
  exit 0
}

if ($Domain) {
  $matches = @($summary.domain_rollup | Where-Object { $_.domain -eq $Domain })
  if ($matches.Count -eq 0) {
    throw "domain not found in coverage summary: $Domain"
  }
  $matches | ConvertTo-Json -Depth 100
  exit 0
}

if ($CoverageId) {
  $matches = @($summary.coverage | Where-Object { $_.id -eq $CoverageId })
  if ($matches.Count -eq 0) {
    throw "coverage id not found in coverage summary: $CoverageId"
  }
  $matches | ConvertTo-Json -Depth 100
  exit 0
}

if ($EvidenceLevel) {
  $declared = @($summary.host_open_policy.evidence_levels)
  if ($declared -notcontains $EvidenceLevel) {
    throw "host-open evidence level is not declared by coverage policy: $EvidenceLevel"
  }

  $matches = @($summary.recipe_index | Where-Object { $_.host_open_evidence.evidence_level -eq $EvidenceLevel })
  [pscustomobject]@{
    evidence_level = $EvidenceLevel
    count = $matches.Count
    recipes = @($matches | Select-Object recipe, coverage_id, domain, host_open_evidence)
  } | ConvertTo-Json -Depth 100
  exit 0
}

[pscustomobject]@{
  totals = $summary.totals
  domains = @($summary.domain_rollup | Select-Object domain, recipe_count, writer_totals, host_open_evidence_levels, boundary_ids)
  boundaries = $summary.boundaries
  open_items = $summary.open_items
} | ConvertTo-Json -Depth 100
