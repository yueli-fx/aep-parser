param(
  [string]$CurrentPath = "flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json"
)

$ErrorActionPreference = "Stop"

function Assert-PathExists {
  param(
    [string]$Label,
    [string]$Path
  )

  if (-not $Path) {
    throw "$Label is empty"
  }
  if (-not (Test-Path $Path)) {
    throw "$Label not found: $Path"
  }
}

function Assert-Equal {
  param(
    [string]$Label,
    [int]$Expected,
    [int]$Actual
  )

  if ($Expected -ne $Actual) {
    throw "$Label mismatch: current=$Expected actual=$Actual"
  }
}

function Assert-MatrixGate {
  param(
    [string]$Label,
    $Gate
  )

  if (-not $Gate) {
    return
  }

  Assert-PathExists "$Label.artifact" $Gate.artifact
  $matrix = Get-Content -Raw $Gate.artifact | ConvertFrom-Json
  Assert-Equal "$Label.totals.total" ([int]$Gate.totals.total) ([int]$matrix.summary.total)
  Assert-Equal "$Label.totals.pass" ([int]$Gate.totals.pass) ([int]$matrix.summary.passed)
  Assert-Equal "$Label.totals.blocked" ([int]$Gate.totals.blocked) ([int]$matrix.summary.blocked)
  Assert-Equal "$Label.totals.failed" ([int]$Gate.totals.failed) ([int]$matrix.summary.failed)
  Assert-Equal "$Label.totals.skipped" ([int]$Gate.totals.skipped) ([int]$matrix.summary.skipped)
}

$current = Get-Content -Raw $CurrentPath | ConvertFrom-Json

Assert-PathExists "truth_sources.current" $current.truth_sources.current
Assert-PathExists "truth_sources.coverage" $current.truth_sources.coverage
if ($current.truth_sources.capability_ledger) {
  Assert-PathExists "truth_sources.capability_ledger" $current.truth_sources.capability_ledger
}

$coverage = Get-Content -Raw $current.truth_sources.coverage | ConvertFrom-Json
$coverageIds = @($coverage.coverage | ForEach-Object { $_.id })

Assert-MatrixGate "current_state.latest_recurring_gate" $current.current_state.latest_recurring_gate
Assert-MatrixGate "current_state.latest_explicit_matte_gate" $current.current_state.latest_explicit_matte_gate

foreach ($doc in $current.frozen_markdown) {
  Assert-PathExists "frozen_markdown.path" $doc.path
}

if ($current.host_open_gap_audit) {
  Assert-PathExists "host_open_gap_audit.script" $current.host_open_gap_audit.script
  foreach ($group in @($current.host_open_gap_audit.completed_groups)) {
    if ($group.artifact) {
      Assert-PathExists "host_open_gap_audit.completed_groups.$($group.coverage_id).artifact" $group.artifact
    }
    if ($group.PSObject.Properties.Name -contains "artifacts") {
      foreach ($artifact in @($group.artifacts)) {
        Assert-PathExists "host_open_gap_audit.completed_groups.$($group.coverage_id).artifacts" $artifact
      }
    }
  }

  if ($current.host_open_gap_audit.artifact -and (Test-Path $current.host_open_gap_audit.artifact)) {
    $audit = Get-Content -Raw $current.host_open_gap_audit.artifact | ConvertFrom-Json
    Assert-Equal "host_open_gap_audit.summary.gap_groups" ([int]$current.host_open_gap_audit.summary.gap_groups) ([int]$audit.summary.gap_groups)
    Assert-Equal "host_open_gap_audit.summary.gap_recipes" ([int]$current.host_open_gap_audit.summary.gap_recipes) ([int]$audit.summary.gap_recipes)
    Assert-Equal "host_open_gap_audit.summary.chunks" ([int]$current.host_open_gap_audit.summary.chunks) ([int]$audit.summary.chunks)
  }
}

if ($current.current_state.latest_standalone_verify_gate) {
  $gate = $current.current_state.latest_standalone_verify_gate
  Assert-PathExists "current_state.latest_standalone_verify_gate.artifact" $gate.artifact
  Assert-PathExists "current_state.latest_standalone_verify_gate.source" $gate.source
  Assert-PathExists "current_state.latest_standalone_verify_gate.target" $gate.target
  Assert-PathExists "current_state.latest_standalone_verify_gate.migration_report" $gate.migration_report

  $verify = Get-Content -Raw $gate.artifact | ConvertFrom-Json
  if ($gate.status -ne $verify.summary.status) {
    throw "current_state.latest_standalone_verify_gate.status mismatch: current=$($gate.status) actual=$($verify.summary.status)"
  }
  if ($gate.profile_diff_status -ne $verify.verification.profile_diff_status) {
    throw "current_state.latest_standalone_verify_gate.profile_diff_status mismatch: current=$($gate.profile_diff_status) actual=$($verify.verification.profile_diff_status)"
  }
  Assert-Equal "current_state.latest_standalone_verify_gate.profile_diff_count" ([int]$gate.profile_diff_count) ([int]$verify.verification.profile_diff_count)
}

foreach ($tool in $current.tooling) {
  if (-not $tool.id) {
    throw "tooling entry missing id"
  }
  Assert-PathExists "tooling.$($tool.id).script" $tool.script
}

foreach ($item in $current.next_batches) {
  if ($item.script) {
    Assert-PathExists "next_batches.$($item.id).script" $item.script
  }
  if ($item.candidate_script -and (Test-Path $item.candidate_script)) {
    continue
  }
}

foreach ($batch in $current.coverage_batches) {
  if (-not $batch.id) {
    throw "coverage batch missing id"
  }
  $entries = @($batch.entries)
  if ($entries.Count -eq 0) {
    throw "coverage batch '$($batch.id)' has no entries"
  }

  foreach ($entry in $entries) {
    if (-not $entry.coverage_id) {
      throw "coverage batch '$($batch.id)' has entry without coverage_id"
    }
    if ($coverageIds -notcontains $entry.coverage_id) {
      throw "coverage batch '$($batch.id)' references unknown coverage_id '$($entry.coverage_id)'"
    }
    Assert-PathExists "coverage batch '$($batch.id)' matrix" $entry.matrix
    if ($entry.ledger) {
      Assert-PathExists "coverage batch '$($batch.id)' ledger" $entry.ledger
    }
    if ($entry.recipe_glob) {
      $recipes = @(Get-ChildItem $entry.recipe_glob)
      if ($recipes.Count -eq 0) {
        throw "coverage batch '$($batch.id)' recipe_glob matched no files: $($entry.recipe_glob)"
      }
    }
    if ($entry.recipe_paths) {
      foreach ($recipePath in @($entry.recipe_paths)) {
        Assert-PathExists "coverage batch '$($batch.id)' recipe_path" $recipePath
      }
    }
  }
}

if ($current.canonical_coverage_batch) {
  $canonical = $current.coverage_batches | Where-Object { $_.id -eq $current.canonical_coverage_batch } | Select-Object -First 1
  if (-not $canonical) {
    throw "canonical_coverage_batch not found: $($current.canonical_coverage_batch)"
  }

  $artifactCoverageIds = @($coverage.coverage | Where-Object { $_.artifact } | ForEach-Object { $_.id } | Sort-Object -Unique)
  $canonicalIds = @($canonical.entries | ForEach-Object { $_.coverage_id } | Sort-Object -Unique)
  $missingFromCanonical = @($artifactCoverageIds | Where-Object { $canonicalIds -notcontains $_ })
  $extraInCanonical = @($canonicalIds | Where-Object { $artifactCoverageIds -notcontains $_ })

  if ($missingFromCanonical.Count -gt 0 -or $extraInCanonical.Count -gt 0) {
    throw "canonical coverage batch '$($canonical.id)' mismatch. missing=[$($missingFromCanonical -join ',')] extra=[$($extraInCanonical -join ',')]"
  }
}

Write-Output "current ok: checked $(@($current.coverage_batches).Count) coverage batches and $(@($current.tooling).Count) tooling entries"
