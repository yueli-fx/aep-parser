[CmdletBinding()]
param(
    [string]$InputPath = "data\samples",
    [string]$OutRoot = "tmp\technique_selfhost_gate",
    [int]$Limit = 0,
    [switch]$Open
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent
Push-Location $repoRoot
try {
    New-Item -ItemType Directory -Force -Path $OutRoot | Out-Null

    $runID = [DateTime]::UtcNow.ToString("yyyyMMddTHHmmssZ")
    $runRoot = Join-Path $OutRoot $runID
    $fullReportDir = Join-Path $runRoot "full_report"
    $partialInputDir = Join-Path $runRoot "partial_input"
    $partialReportDir = Join-Path $runRoot "partial_report"
    $compareSelfDir = Join-Path $runRoot "compare_self"
    $comparePartialDir = Join-Path $runRoot "compare_partial_to_full"
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null

    function Invoke-GateStep {
        param(
            [string]$Name,
            [scriptblock]$Body
        )
        $timer = [System.Diagnostics.Stopwatch]::StartNew()
        & $Body
        $exit = $LASTEXITCODE
        $timer.Stop()
        if ($null -eq $exit) {
            $exit = 0
        }
        Write-Host ("{0}: exit={1} seconds={2}" -f $Name, [int]$exit, [Math]::Round($timer.Elapsed.TotalSeconds, 2))
        if ($exit -ne 0) {
            throw "$Name failed with exit code $exit"
        }
    }

    $reportArgs = @(
        "-NoProfile",
        "-File",
        (Join-Path $PSScriptRoot "technique_showcase_report.ps1"),
        "-InputPath",
        $InputPath,
        "-OutDir",
        $fullReportDir,
        "-Verify"
    )
    if ($Limit -gt 0) {
        $reportArgs += @("-Limit", "$Limit")
    }

    Invoke-GateStep -Name "go technique tests" -Body {
        & go test ./cmd/aeptechnique ./internal/technique -count=1
    }
    Invoke-GateStep -Name "full technique report" -Body {
        & pwsh @reportArgs
    }

    New-Item -ItemType Directory -Force -Path $partialInputDir | Out-Null
    Copy-Item -LiteralPath "flightdeck\showcase\text\text.aep" -Destination (Join-Path $partialInputDir "good.aep")
    Set-Content -LiteralPath (Join-Path $partialInputDir "bad.aep") -Value "not an aep" -Encoding ASCII

    Invoke-GateStep -Name "partial-error technique report" -Body {
        & pwsh -NoProfile -File (Join-Path $PSScriptRoot "technique_showcase_report.ps1") -InputPath $partialInputDir -OutDir $partialReportDir -Verify
    }
    Invoke-GateStep -Name "self compare" -Body {
        & go run ./cmd/aepselfhost compare-reports -base $fullReportDir -new $fullReportDir -out $compareSelfDir
    }
    Invoke-GateStep -Name "partial-to-full compare" -Body {
        & go run ./cmd/aepselfhost compare-reports -base $partialReportDir -new $fullReportDir -out $comparePartialDir -top 5
    }
    Invoke-GateStep -Name "recipe draft smoke" -Body {
        & go run ./cmd/aepselfhost recipe-smoke -full-report $fullReportDir -run-root $runRoot -batch-limit 3
    }
    Invoke-GateStep -Name "finalize selfhost run" -Body {
        & go run ./cmd/aepselfhost finalize-run -out-root $OutRoot -run-root $runRoot -run-id $runID -input $InputPath
    }

    Write-Host "latest run:      $(Join-Path $OutRoot "latest_run.txt")"
    Write-Host "latest outcome:  $(Join-Path $OutRoot "latest_outcome.md")"
    Write-Host "latest outcome h: $(Join-Path $OutRoot "latest_outcome.html")"
    Write-Host "latest outcome j: $(Join-Path $OutRoot "latest_outcome.json")"
    Write-Host "latest effect:   $(Join-Path $OutRoot "latest_effectiveness.md")"
    Write-Host "latest effect j: $(Join-Path $OutRoot "latest_effectiveness.json")"
    Write-Host "latest index:    $(Join-Path $OutRoot "latest_index.html")"
    Write-Host "full report:     $fullReportDir"
    Write-Host "partial report:  $partialReportDir"
    if ($Open) {
        Invoke-Item (Join-Path $OutRoot "latest_outcome.html")
    }
}
finally {
    Pop-Location
}
