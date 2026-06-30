<#
.SYNOPSIS
  Regenerate test_data fixtures from scripts/fixtures/fixtures_manifest.json by
  driving each generator JSX through ae_run.ps1 (unattended). Default mode regenerates
  ONLY entries with missing outputs — existing fixtures are never overwritten
  unless -Force (AE saves are nondeterministic: GUIDs/timestamps shift, which
  would churn byte-diff RE baselines for no reason).

  Granularity note: a job reruns its WHOLE generator JSX — sibling outputs of
  the same JSX (e.g. a before/after pair) are rewritten together even if only
  one was missing. Deliberate: pairs are only byte-consistent when produced in
  one AE session; mixing an old "before" with a fresh "after" would corrupt
  byte-diff RE comparisons.

.PARAMETER CheckOnly   inventory only: report missing outputs + ungoverned .aep
                       (present on disk but no manifest entry and not git-tracked)
.PARAMETER Force       regenerate ALL driveable entries, even if outputs exist
.PARAMETER Only        substring filter on the jsx name (e.g. 'delete_layer')
.PARAMETER TimeoutSec  per-AE-run timeout passed to ae_run.ps1 (default 180)

.OUTPUTS exit 0 = nothing missing / all regens succeeded; 1 = failures or
  (in -CheckOnly) missing driveable outputs.
#>
[CmdletBinding()]
param(
    [switch]$CheckOnly,
    [switch]$Force,
    [string]$Only = '',
    [int]$TimeoutSec = 180
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$manifest = Get-Content (Join-Path $PSScriptRoot 'fixtures_manifest.json') -Raw -Encoding UTF8 | ConvertFrom-Json

# Expand an entry's env (var -> value list) into the cartesian product of
# combinations. No env -> one empty combination.
function Get-EnvCombos([pscustomobject]$envSpec) {
    $combos = @(@{})
    if ($null -eq $envSpec) { return ,$combos }
    foreach ($p in $envSpec.PSObject.Properties) {
        $next = @()
        foreach ($combo in $combos) {
            foreach ($v in $p.Value) {
                $c = $combo.Clone(); $c[$p.Name] = $v; $next += ,$c
            }
        }
        $combos = $next
    }
    return ,$combos
}

function Expand-Placeholders([string]$template, [hashtable]$combo) {
    $s = $template
    foreach ($k in $combo.Keys) { $s = $s.Replace("{$k}", $combo[$k]) }
    return $s
}

function Resolve-AeExe([pscustomobject]$entry, [hashtable]$combo) {
    $ae = $entry.ae
    if ($ae -is [string]) { return $manifest.ae_exes.$ae }
    # per-mode map keyed by the first env var's value, 'default' fallback
    $firstVal = if ($combo.Count -gt 0) { $combo[@($combo.Keys)[0]] } else { '' }
    $key = if ($ae.PSObject.Properties[$firstVal]) { $firstVal } else { 'default' }
    return $manifest.ae_exes.($ae.$key)
}

$jobs = @()      # one job per (entry, combo): everything needed to run + verify
foreach ($entry in $manifest.fixtures) {
    if ($Only -and ($entry.jsx -notlike "*$Only*")) { continue }
    foreach ($combo in (Get-EnvCombos $entry.env)) {
        $outputs = @($entry.outputs | ForEach-Object { Expand-Placeholders $_ $combo })
        $missing = @($outputs | Where-Object { -not (Test-Path (Join-Path $repoRoot $_)) })
        $jobs += [pscustomobject]@{
            Entry = $entry; Combo = $combo; Outputs = $outputs; Missing = $missing
            Label = "$(Split-Path $entry.jsx -Leaf)$(if ($combo.Count) { ' [' + (@($combo.Keys | ForEach-Object { "$_=$($combo[$_])" }) -join ' ') + ']' })"
        }
    }
}

# Ungoverned inventory: on-disk .aep with no manifest entry and not git-tracked.
# ge_* are AE-side ground-truth byproducts of ship-gate verify runs (regenerated
# by the gates themselves). Anything left has NO known generator — regen
# impossible if lost.
$governed = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
foreach ($j in $jobs) { foreach ($o in $j.Outputs) { [void]$governed.Add($o.Replace('\', '/')) } }
$trackedPaths = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
git -C $repoRoot ls-files test_data | ForEach-Object { [void]$trackedPaths.Add($_.Replace('\', '/')) }
$unknown = Get-ChildItem (Join-Path $repoRoot 'test_data') -Filter *.aep -File -Recurse |
    ForEach-Object { $_.FullName.Substring($repoRoot.Length + 1).Replace('\', '/') } |
    Where-Object { -not $governed.Contains($_) -and -not $trackedPaths.Contains($_) }
$gateByproducts = @($unknown | Where-Object { (Split-Path $_ -Leaf) -like 'ge_*' })
$ungoverned = @($unknown | Where-Object { $_ -notin $gateByproducts })

$missingJobs = @($jobs | Where-Object { $_.Missing.Count -gt 0 -and -not $_.Entry.manual })
$manualMissing = @($jobs | Where-Object { $_.Missing.Count -gt 0 -and $_.Entry.manual })

Write-Host ("inventory: {0} manifest jobs, {1} with missing outputs ({2} driveable, {3} manual-only), {4} ungoverned on-disk aep" -f `
    $jobs.Count, ($missingJobs.Count + $manualMissing.Count), $missingJobs.Count, $manualMissing.Count, $ungoverned.Count)
foreach ($j in $missingJobs) { Write-Host "  missing: $($j.Label) -> $($j.Missing -join ', ')" }
foreach ($j in $manualMissing) { Write-Host "  missing (MANUAL entry, not driveable): $($j.Label)" }
if ($gateByproducts.Count -gt 0) {
    Write-Host "  gate byproducts (ge_*, regenerated by ship-gate verify runs): $($gateByproducts.Count)"
}
if ($ungoverned.Count -gt 0) {
    Write-Host "  ungoverned (no generator known, not git-tracked — regen impossible if lost):"
    $ungoverned | ForEach-Object { Write-Host "    $_" }
}

if ($CheckOnly) { exit ([int]($missingJobs.Count -gt 0)) }

$toRun = if ($Force) { @($jobs | Where-Object { -not $_.Entry.manual }) } else { $missingJobs }
if ($toRun.Count -eq 0) { Write-Host 'nothing to regenerate.'; exit 0 }

$aeRun = Join-Path (Split-Path $PSScriptRoot -Parent) 'ae-worker\ae_run.ps1'
$failures = @()
foreach ($j in $toRun) {
    $donePath = Join-Path $repoRoot (Expand-Placeholders $j.Entry.done $j.Combo)
    $jsxPath = Join-Path $repoRoot $j.Entry.jsx
    $aeExe = Resolve-AeExe $j.Entry $j.Combo
    Write-Host "regen: $($j.Label)  (AE: $(Split-Path (Split-Path (Split-Path $aeExe -Parent) -Parent) -Leaf))"
    foreach ($k in $j.Combo.Keys) { Set-Item -Path "env:$k" -Value $j.Combo[$k] }
    try {
        & pwsh -NoProfile -File $aeRun -AeExe $aeExe -Jsx $jsxPath -Done $donePath -TimeoutSec $TimeoutSec
        $code = $LASTEXITCODE
    } finally {
        foreach ($k in $j.Combo.Keys) { Remove-Item -Path "env:$k" -ErrorAction SilentlyContinue }
    }
    $stillMissing = @($j.Outputs | Where-Object { -not (Test-Path (Join-Path $repoRoot $_)) })
    if ($code -ne 0 -or $stillMissing.Count -gt 0) {
        $failures += "$($j.Label): ae_run exit $code, missing after run: $($stillMissing -join ', ')"
    }
}

Write-Host ""
if ($failures.Count -gt 0) {
    Write-Host "=== regen FAILURES ($($failures.Count)/$($toRun.Count)) ==="
    $failures | ForEach-Object { Write-Host "  $_" }
    exit 1
}
Write-Host "=== regen complete: $($toRun.Count)/$($toRun.Count) OK ==="
exit 0
