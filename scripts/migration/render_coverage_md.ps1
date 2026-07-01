param(
  [string]$CoveragePath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
  [string]$Out = "tmp/migration_coverage.md"
)

$ErrorActionPreference = "Stop"

function Format-Totals {
  param($Totals)
  return "$($Totals.pass)/$($Totals.total) pass, blocked=$($Totals.blocked), failed=$($Totals.failed), skipped=$($Totals.skipped)"
}

$coverage = Get-Content -Raw $CoveragePath | ConvertFrom-Json
$lines = @()

$lines += "# Versioned AEP Migration Coverage"
$lines += ""
$lines += "Generated from `$CoveragePath`." -replace '\$CoveragePath', $CoveragePath
$lines += ""
$lines += "## Host Open Policy"
$lines += ""
$lines += "- Matrix command: {0}" -f $coverage.host_open_policy.matrix_command_status
$lines += "- Default strategy: {0}" -f $coverage.host_open_policy.default_strategy
$lines += "- Broad fanout: {0}" -f $coverage.host_open_policy.broad_fanout_status
$lines += "- Endpoint inference: direct={0}; inferred={1}; label={2}" -f (($coverage.host_open_policy.endpoint_inference.direct_hosts) -join ","), (($coverage.host_open_policy.endpoint_inference.inferred_hosts) -join ","), $coverage.host_open_policy.endpoint_inference.label
$lines += ""
$lines += "## Recurring Gates"
$lines += ""
$lines += "| Gate | Status | Totals | Artifact |"
$lines += "| --- | --- | --- | --- |"
foreach ($gate in $coverage.recurring_gates) {
  $lines += '| `{0}` | {1} | {2} | `{3}` |' -f $gate.id, $gate.status, (Format-Totals $gate.totals), $gate.artifact
}

$lines += ""
$lines += "## Coverage"
$lines += ""
$lines += "| ID | Domain | Writer Status | Totals | Host Open | Artifact |"
$lines += "| --- | --- | --- | --- | --- | --- |"
foreach ($record in $coverage.coverage) {
  $lines += '| `{0}` | {1} | {2} | {3} | {4} | `{5}` |' -f $record.id, $record.domain, $record.writer_status, (Format-Totals $record.totals), $record.host_open_status, $record.artifact
}

$lines += ""
$lines += "## Open Items"
$lines += ""
foreach ($item in $coverage.open_items) {
  $lines += '- `{0}`: {1} - {2}' -f $item.id, $item.status, $item.scope
}

$outDir = Split-Path -Parent $Out
if ($outDir -and -not (Test-Path $outDir)) {
  New-Item -ItemType Directory -Path $outDir | Out-Null
}
Set-Content -Path $Out -Value ($lines -join "`n") -Encoding utf8
Write-Output "rendered coverage markdown: $Out"
