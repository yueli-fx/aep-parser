param(
  [Parameter(Mandatory=$true)]
  [string]$Id,

  [Parameter(Mandatory=$true)]
  [string]$MatrixPath,

  [string]$CoveragePath = "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json",
  [string]$Domain = "",
  [string]$Scope = "",
  [string]$WriterStatus = "",
  [string]$HostOpenStatus = "",
  [string]$LedgerPath = ""
)

$ErrorActionPreference = "Stop"

function Convert-ToRepoPath {
  param([string]$Path)

  $fullRoot = (Resolve-Path ".").Path
  $fullPath = (Resolve-Path $Path).Path
  if ($fullPath.StartsWith($fullRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    return $fullPath.Substring($fullRoot.Length + 1).Replace("\", "/")
  }
  return $Path.Replace("\", "/")
}

function Get-StatusFromSummary {
  param($Summary)

  if ($Summary.failed -gt 0) {
    return "failed"
  }
  if ($Summary.blocked -gt 0 -or $Summary.skipped -gt 0) {
    return "boundary"
  }
  return "PD-6x6"
}

$coverage = Get-Content -Raw $CoveragePath | ConvertFrom-Json
$matrix = Get-Content -Raw $MatrixPath | ConvertFrom-Json

$recipeNames = @($matrix.cases | ForEach-Object { $_.recipe_name } | Sort-Object -Unique)
$sourceWriters = @($matrix.cases | ForEach-Object { $_.source_version } | Sort-Object -Unique)
$targetWriters = @($matrix.cases | ForEach-Object { $_.target_version } | Sort-Object -Unique)

if (-not $LedgerPath) {
  $matrixItem = Get-Item $MatrixPath
  $candidate = Join-Path $matrixItem.DirectoryName "ledger.md"
  if (Test-Path $candidate) {
    $LedgerPath = $candidate
  }
}

$summary = $matrix.summary
$totals = [ordered]@{
  total = [int]$summary.total
  pass = [int]$summary.passed
  blocked = [int]$summary.blocked
  failed = [int]$summary.failed
  skipped = [int]$summary.skipped
}

$record = $coverage.coverage | Where-Object { $_.id -eq $Id } | Select-Object -First 1
if (-not $record) {
  if (-not $Domain -or -not $Scope) {
    throw "Coverage id '$Id' does not exist. Provide -Domain and -Scope to create it."
  }

  $record = [pscustomobject][ordered]@{
    id = $Id
    domain = $Domain
    scope = $Scope
    writer_status = ""
    writer_coverage = ""
    host_open_status = ""
    artifact = ""
    ledger = ""
    totals = $null
  }
  $coverage.coverage += $record
}

if ($Domain) {
  $record.domain = $Domain
}
if ($Scope) {
  $record.scope = $Scope
}

$record.writer_status = if ($WriterStatus) { $WriterStatus } else { Get-StatusFromSummary $summary }
$record.writer_coverage = "$($sourceWriters -join ',') sources into $($targetWriters -join ',') targets"
if ($HostOpenStatus) {
  $record.host_open_status = $HostOpenStatus
} elseif (-not $record.host_open_status) {
  $record.host_open_status = "pending"
}

$record.artifact = Convert-ToRepoPath $MatrixPath
if ($LedgerPath) {
  $record.ledger = Convert-ToRepoPath $LedgerPath
}
$record.totals = [pscustomobject]$totals

if ($recipeNames.Count -eq 1) {
  if ($record.PSObject.Properties.Name -contains "recipes") {
    $record.recipes = @($recipeNames)
  } else {
    $record | Add-Member -NotePropertyName recipes -NotePropertyValue @($recipeNames)
  }
  if ($record.PSObject.Properties.Name -contains "recipe_count") {
    $record.recipe_count = 1
  }
} else {
  if ($record.PSObject.Properties.Name -contains "recipes") {
    $record.PSObject.Properties.Remove("recipes")
  }
  if ($record.PSObject.Properties.Name -contains "recipe_count") {
    $record.recipe_count = $recipeNames.Count
  } else {
    $record | Add-Member -NotePropertyName recipe_count -NotePropertyValue $recipeNames.Count
  }
}

$coverage.generated_at = (Get-Date -Format "yyyy-MM-dd")
$json = $coverage | ConvertTo-Json -Depth 20
Set-Content -Path $CoveragePath -Value $json -Encoding utf8

Write-Output "updated coverage id '$Id' from $MatrixPath"
Write-Output "totals: total=$($totals.total) pass=$($totals.pass) blocked=$($totals.blocked) failed=$($totals.failed) skipped=$($totals.skipped)"
