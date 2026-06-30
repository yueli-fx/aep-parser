param(
    [string]$InputPath = "data\samples",
    [string]$OutRoot = "tmp\technique_selfhost_gate",
    [int]$Limit = 0
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
    $acceptanceJsonPath = Join-Path $runRoot "acceptance.json"
    $acceptanceMdPath = Join-Path $runRoot "acceptance.md"
    $latestRunPath = Join-Path $OutRoot "latest_run.txt"
    $latestAcceptancePath = Join-Path $OutRoot "latest_acceptance.md"
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null

    $steps = [System.Collections.ArrayList]::new()

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
        [void]$steps.Add([ordered]@{
            name    = $Name
            exit    = [int]$exit
            seconds = [Math]::Round($timer.Elapsed.TotalSeconds, 2)
        })
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
        & pwsh -NoProfile -File (Join-Path $PSScriptRoot "compare_technique_reports.ps1") -BaseDir $fullReportDir -NewDir $fullReportDir -OutDir $compareSelfDir
    }
    Invoke-GateStep -Name "partial-to-full compare" -Body {
        & pwsh -NoProfile -File (Join-Path $PSScriptRoot "compare_technique_reports.ps1") -BaseDir $partialReportDir -NewDir $fullReportDir -OutDir $comparePartialDir -Top 5
    }

    $fullManifest = Get-Content -Raw -LiteralPath (Join-Path $fullReportDir "manifest.json") | ConvertFrom-Json
    $partialManifest = Get-Content -Raw -LiteralPath (Join-Path $partialReportDir "manifest.json") | ConvertFrom-Json
    $compareSelf = Get-Content -Raw -LiteralPath (Join-Path $compareSelfDir "compare.json") | ConvertFrom-Json
    $comparePartial = Get-Content -Raw -LiteralPath (Join-Path $comparePartialDir "compare.json") | ConvertFrom-Json

    $selfCountDiffs = @($compareSelf.count_diffs).Count
    $partialCountDiffs = @($comparePartial.count_diffs).Count
    if ($selfCountDiffs -ne 0) {
        throw "self compare produced $selfCountDiffs count diff(s)"
    }
    if ($partialCountDiffs -eq 0) {
        throw "partial-to-full compare produced no count diffs"
    }
    if ([int]$partialManifest.error_count -lt 1) {
        throw "partial report did not record the intentional bad AEP"
    }

    $acceptance = [ordered]@{
        schema_version = 1
        generated_at_utc = [DateTime]::UtcNow.ToString("o")
        input_path = $InputPath
        run_root = $runRoot
        full_report = [ordered]@{
            path = $fullReportDir
            project_count = [int]$fullManifest.project_count
            error_count = [int]$fullManifest.error_count
            pattern_count = [int]$fullManifest.pattern_count
        }
        partial_report = [ordered]@{
            path = $partialReportDir
            project_count = [int]$partialManifest.project_count
            error_count = [int]$partialManifest.error_count
            pattern_count = [int]$partialManifest.pattern_count
        }
        compares = [ordered]@{
            self = $compareSelfDir
            partial_to_full = $comparePartialDir
            partial_to_full_count_diffs = $partialCountDiffs
        }
        steps = @($steps)
    }
    $acceptance | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $acceptanceJsonPath -Encoding UTF8

    $b = [System.Text.StringBuilder]::new()
    [void]$b.AppendLine("# Technique Self-Hosted Acceptance")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("- input: ``$InputPath``")
    [void]$b.AppendLine("- run root: ``$runRoot``")
    [void]$b.AppendLine("- full report: ``$fullReportDir``")
    [void]$b.AppendLine("- projects: $($fullManifest.project_count)")
    [void]$b.AppendLine("- errors: $($fullManifest.error_count)")
    [void]$b.AppendLine("- patterns: $($fullManifest.pattern_count)")
    [void]$b.AppendLine("- partial report errors: $($partialManifest.error_count)")
    [void]$b.AppendLine("- partial-to-full count diffs: $partialCountDiffs")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("## Steps")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("| step | exit | seconds |")
    [void]$b.AppendLine("| --- | ---: | ---: |")
    foreach ($step in $steps) {
        [void]$b.AppendLine("| $($step.name) | $($step.exit) | $($step.seconds) |")
    }
    $b.ToString() | Set-Content -LiteralPath $acceptanceMdPath -Encoding UTF8
    $runRoot | Set-Content -LiteralPath $latestRunPath -Encoding UTF8
    Copy-Item -LiteralPath $acceptanceMdPath -Destination $latestAcceptancePath -Force

    Write-Host "acceptance json: $acceptanceJsonPath"
    Write-Host "acceptance md:   $acceptanceMdPath"
    Write-Host "latest run:      $latestRunPath"
    Write-Host "latest summary:  $latestAcceptancePath"
    Write-Host "full report:     $fullReportDir"
    Write-Host "partial report:  $partialReportDir"
}
finally {
    Pop-Location
}
