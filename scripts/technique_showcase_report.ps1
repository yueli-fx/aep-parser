param(
    [string]$InputPath = "flightdeck\showcase",
    [string]$OutDir = "tmp\technique_showcase_report",
    [int]$Limit = 0
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent
Push-Location $repoRoot
try {
    New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
    $summaryPath = Join-Path $OutDir "summary.json"
    $corpusPath = Join-Path $OutDir "corpus.jsonl"
    $reportPath = Join-Path $OutDir "report.md"

    $baseArgs = @(
        "run", "./cmd/aeptechnique",
        "-in", $InputPath,
        "-mode", "portrait",
        "-corpus",
        "-recursive"
    )
    if ($Limit -gt 0) {
        $baseArgs += @("-limit", "$Limit")
    }

    & go @baseArgs "-summary" "-out" $summaryPath
    if ($LASTEXITCODE -ne 0) {
        throw "aeptechnique summary failed with exit code $LASTEXITCODE"
    }

    & go @baseArgs "-out" $corpusPath
    if ($LASTEXITCODE -ne 0) {
        throw "aeptechnique corpus failed with exit code $LASTEXITCODE"
    }

    $summary = Get-Content -Raw -Path $summaryPath | ConvertFrom-Json
    $errorCount = 0
    if ($null -ne $summary.error_count) {
        $errorCount = $summary.error_count
    }

    function Write-CountTable {
        param(
            [System.Text.StringBuilder]$Builder,
            [string]$Title,
            [object]$Counts,
            [int]$Max = 12
        )
        [void]$Builder.AppendLine("## $Title")
        $rows = @()
        if ($null -ne $Counts) {
            $rows = $Counts.PSObject.Properties |
                Sort-Object @{ Expression = { [int]$_.Value }; Descending = $true }, Name |
                Select-Object -First $Max
        }
        if ($rows.Count -eq 0) {
            [void]$Builder.AppendLine("")
            [void]$Builder.AppendLine("_none_")
            [void]$Builder.AppendLine("")
            return
        }
        [void]$Builder.AppendLine("")
        [void]$Builder.AppendLine("| name | count |")
        [void]$Builder.AppendLine("| --- | ---: |")
        foreach ($row in $rows) {
            [void]$Builder.AppendLine("| $($row.Name) | $($row.Value) |")
        }
        [void]$Builder.AppendLine("")
    }

    $b = [System.Text.StringBuilder]::new()
    [void]$b.AppendLine("# Technique Showcase Report")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("- input: ``$InputPath``")
    [void]$b.AppendLine("- projects: $($summary.project_count)")
    [void]$b.AppendLine("- errors: $errorCount")
    [void]$b.AppendLine("- comps: $($summary.totals.comp_count)")
    [void]$b.AppendLine("- layers: $($summary.totals.layer_count)")
    [void]$b.AppendLine("- effects: $($summary.totals.effect_count)")
    [void]$b.AppendLine("- text animators: $($summary.totals.text_animator_count)")
    [void]$b.AppendLine("- shape operators: $($summary.totals.shape_operator_count)")
    [void]$b.AppendLine("- dependency edges: $($summary.totals.dependency_count)")
    [void]$b.AppendLine("")
    Write-CountTable -Builder $b -Title "Technique Hints" -Counts $summary.hint_counts
    Write-CountTable -Builder $b -Title "Effects" -Counts $summary.effect_counts
    Write-CountTable -Builder $b -Title "Shape Families" -Counts $summary.shape_families
    Write-CountTable -Builder $b -Title "Text Animators" -Counts $summary.text_animators
    Write-CountTable -Builder $b -Title "Layer Roles" -Counts $summary.layer_roles
    Write-CountTable -Builder $b -Title "Graph Edges" -Counts $summary.graph_edges

    $b.ToString() | Set-Content -Path $reportPath -Encoding UTF8

    Write-Host "summary: $summaryPath"
    Write-Host "corpus:  $corpusPath"
    Write-Host "report:  $reportPath"
}
finally {
    Pop-Location
}
