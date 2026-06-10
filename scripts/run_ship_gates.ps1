<#
.SYNOPSIS
  Full AE ship-gate sweep: pre-clean, run every gated test (AE_SHIP_GATE=1,
  -count=1), and write a PASS/FAIL/SKIP ledger. The regression backstop that
  per-feature gate runs don't provide — run after cross-cutting changes
  (ID allocation, write path, wrapper) or on a schedule.

.PARAMETER Run            go test -run pattern (default: every *ShipGate* test)
.PARAMETER GoTimeoutMin   go test -timeout, minutes (default 240 — 76 gates × ~40s worst + retries)
.PARAMETER Ledger         ledger JSON path (default test_data/gate_ledger.json, gitignored)
.PARAMETER SkipPreClean   don't kill straggler AfterFX processes first
.PARAMETER ClearCrashState  also run tmp_debug/clear_ae_crashstate.ps1 per AE exe before the sweep
                            (one extra AE launch each; only needed after a force-kill armed the flag)

.OUTPUTS
  Exit code = go test exit code (0 all green / skipped, non-zero on any FAIL).
  Ledger JSON: { generated, pattern, goExit, counts, results: [{name,status,seconds}] }
  Full go test -v output lands next to the ledger as gate_sweep.log.
#>
[CmdletBinding()]
param(
    [string]$Run = 'ShipGate',
    [int]$GoTimeoutMin = 240,
    [string]$Ledger = (Join-Path (Split-Path $PSScriptRoot -Parent) 'test_data/gate_ledger.json'),
    [switch]$SkipPreClean,
    [switch]$ClearCrashState
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path $PSScriptRoot -Parent
$logPath = Join-Path (Split-Path $Ledger -Parent) 'gate_sweep.log'

if (-not $SkipPreClean) {
    $straggler = @(Get-Process -Name 'AfterFX*' -ErrorAction SilentlyContinue)
    if ($straggler.Count -gt 0) {
        Write-Host "pre-clean: killing $($straggler.Count) straggler AfterFX process(es)"
        $straggler | Stop-Process -Force
        Start-Sleep -Seconds 2
    }
}
if ($ClearCrashState) {
    $clear = Join-Path $repoRoot 'tmp_debug/clear_ae_crashstate.ps1'
    foreach ($exe in @(
        'E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe',
        'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe')) {
        if (Test-Path $exe) {
            Write-Host "pre-clean: clearing crash state for $exe"
            pwsh -NoProfile -File $clear -AeExe $exe
        }
    }
}

Write-Host "sweep: go test -run '$Run' -count=1 (timeout ${GoTimeoutMin}m) — output → $logPath"
$sw = [System.Diagnostics.Stopwatch]::StartNew()
$env:AE_SHIP_GATE = '1'
& go test ./internal/aep/ -run $Run -count=1 -v -timeout "${GoTimeoutMin}m" 2>&1 | Tee-Object -FilePath $logPath | ForEach-Object {
    if ($_ -match '^(--- (PASS|FAIL|SKIP)|ok |FAIL)') { Write-Host $_ }
}
$goExit = $LASTEXITCODE
$sw.Stop()

$results = foreach ($line in Get-Content $logPath) {
    if ($line -match '^--- (PASS|FAIL|SKIP): (\S+) \(([\d.]+)s\)') {
        [pscustomobject]@{ name = $Matches[2]; status = $Matches[1]; seconds = [double]$Matches[3] }
    }
}
$counts = [ordered]@{
    PASS = @($results | Where-Object status -EQ 'PASS').Count
    FAIL = @($results | Where-Object status -EQ 'FAIL').Count
    SKIP = @($results | Where-Object status -EQ 'SKIP').Count
}
[ordered]@{
    generated = (Get-Date).ToString('yyyy-MM-dd HH:mm:ss')
    pattern   = $Run
    goExit    = $goExit
    elapsedMin = [math]::Round($sw.Elapsed.TotalMinutes, 1)
    counts    = $counts
    results   = $results | Sort-Object name
} | ConvertTo-Json -Depth 4 | Set-Content -Path $Ledger -Encoding UTF8

Write-Host ""
Write-Host ("=== gate sweep: {0} PASS / {1} FAIL / {2} SKIP in {3:N1} min — ledger: {4} ===" -f `
    $counts.PASS, $counts.FAIL, $counts.SKIP, $sw.Elapsed.TotalMinutes, $Ledger)
if ($counts.FAIL -gt 0) {
    $results | Where-Object status -EQ 'FAIL' | ForEach-Object { Write-Host "  FAIL $($_.name)" }
}
exit $goExit
