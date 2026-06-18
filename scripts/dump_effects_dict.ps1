<#
.SYNOPSIS
  Drive dump_effects_dict.jsx across every installed AE version, producing one
  version-tagged effects dictionary per version in data/effects-dict/.

  Each version: kill stragglers -> run via ae_run.ps1 -> on cold-start exit-2,
  warm-retry up to 2x (AE 2025 splash can outlast ae_run's unknown-grace).

.PARAMETER AeExes  Array of AfterFX.exe paths. Defaults to the E:\adobe installs.
#>
[CmdletBinding()]
param(
    [string[]]$AeExes = @(
        "E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe",
        "E:\adobe\Adobe After Effects 2022\Support Files\AfterFX.exe",
        "E:\adobe\Adobe After Effects 2023\Support Files\AfterFX.exe",
        "E:\adobe\Adobe After Effects 2024\Support Files\AfterFX.exe",
        "E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe"
    )
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path $PSScriptRoot -Parent
$jsx  = Join-Path $PSScriptRoot 'dump_effects_dict.jsx'
$done = Join-Path $repo 'data\effects-dict\effects_dict.done'
$aeRun = Join-Path $PSScriptRoot 'ae_run.ps1'

foreach ($ae in $AeExes) {
    if (-not (Test-Path $ae)) { Write-Host "SKIP (not found): $ae"; continue }
    $label = Split-Path (Split-Path $ae -Parent) -Parent | Split-Path -Leaf
    Write-Host "=== $label ===" -ForegroundColor Cyan
    $ok = $false
    for ($attempt = 1; $attempt -le 3; $attempt++) {
        Get-Process AfterFX* -ErrorAction SilentlyContinue | Stop-Process -Force
        Start-Sleep -Seconds 2
        Remove-Item $done -ErrorAction SilentlyContinue
        & pwsh -File $aeRun -AeExe $ae -Jsx $jsx -Done $done -TimeoutSec 300
        $code = $LASTEXITCODE
        if ($code -eq 0) {
            $ok = $true
            Write-Host (Get-Content $done -Raw)
            break
        }
        Write-Host "  attempt $attempt exit=$code (cold-start? retrying)" -ForegroundColor Yellow
    }
    if (-not $ok) { Write-Host "  FAILED after retries: $label" -ForegroundColor Red }
}

Get-Process AfterFX* -ErrorAction SilentlyContinue | Stop-Process -Force
Write-Host "`n=== dictionaries in data/effects-dict/ ===" -ForegroundColor Green
Get-ChildItem (Join-Path $repo 'data\effects-dict') -Filter 'effects_*.json' | ForEach-Object { "  {0}  ({1:N0} KB)" -f $_.Name, ($_.Length/1KB) }
