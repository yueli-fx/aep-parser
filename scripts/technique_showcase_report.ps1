param(
    [string]$InputPath = "flightdeck\showcase",
    [string]$OutDir = "tmp\technique_showcase_report",
    [int]$Limit = 0,
    [switch]$Open
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent
Push-Location $repoRoot
try {
    New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
    $summaryPath = Join-Path $OutDir "summary.json"
    $corpusPath = Join-Path $OutDir "corpus.jsonl"
    $reportPath = Join-Path $OutDir "report.md"
    $htmlPath = Join-Path $OutDir "report.html"

    $baseArgs = @(
        "run", "./cmd/aeptechnique",
        "-in", $InputPath,
        "-mode", "explain",
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

    function Get-CountRows {
        param(
            [object]$Counts,
            [int]$Max = 12
        )
        if ($null -eq $Counts) {
            return @()
        }
        return @($Counts.PSObject.Properties |
            Sort-Object @{ Expression = { [int]$_.Value }; Descending = $true }, Name |
            Select-Object -First $Max)
    }

    function Escape-Html {
        param([AllowNull()][object]$Value)
        if ($null -eq $Value) {
            return ""
        }
        return [System.Net.WebUtility]::HtmlEncode([string]$Value)
    }

    function Write-HtmlCountTable {
        param(
            [System.Text.StringBuilder]$Builder,
            [string]$Title,
            [object]$Counts,
            [int]$Max = 12
        )
        [void]$Builder.AppendLine("<section class=""panel"">")
        [void]$Builder.AppendLine("<h2>$(Escape-Html $Title)</h2>")
        $rows = Get-CountRows -Counts $Counts -Max $Max
        if ($rows.Count -eq 0) {
            [void]$Builder.AppendLine("<p class=""empty"">none</p>")
            [void]$Builder.AppendLine("</section>")
            return
        }
        [void]$Builder.AppendLine("<table><thead><tr><th>Name</th><th>Count</th></tr></thead><tbody>")
        foreach ($row in $rows) {
            [void]$Builder.AppendLine("<tr><td>$(Escape-Html $row.Name)</td><td>$($row.Value)</td></tr>")
        }
        [void]$Builder.AppendLine("</tbody></table>")
        [void]$Builder.AppendLine("</section>")
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

    $records = @()
    if (Test-Path $corpusPath) {
        $records = @(Get-Content -Path $corpusPath | Where-Object { $_.Trim() -ne "" } | ForEach-Object {
            $_ | ConvertFrom-Json
        })
    }

    [void]$b.AppendLine("## Project Explanations")
    [void]$b.AppendLine("")
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation) {
            continue
        }
        [void]$b.AppendLine("### ``$($record.path)``")
        [void]$b.AppendLine("")
        foreach ($line in @($explanation.overview | Select-Object -First 2)) {
            [void]$b.AppendLine("- $line")
        }
        foreach ($tech in @($explanation.techniques | Select-Object -First 4)) {
            [void]$b.AppendLine("- **$($tech.title)**: $($tech.summary)")
        }
        [void]$b.AppendLine("")
    }

    $b.ToString() | Set-Content -Path $reportPath -Encoding UTF8

    $h = [System.Text.StringBuilder]::new()
    [void]$h.AppendLine("<!doctype html>")
    [void]$h.AppendLine("<html lang=""en"">")
    [void]$h.AppendLine("<head>")
    [void]$h.AppendLine("<meta charset=""utf-8"">")
    [void]$h.AppendLine("<meta name=""viewport"" content=""width=device-width, initial-scale=1"">")
    [void]$h.AppendLine("<title>Technique Showcase Report</title>")
    [void]$h.AppendLine("<style>")
    [void]$h.AppendLine("body{font-family:Segoe UI,Arial,sans-serif;margin:0;background:#f5f7fa;color:#1f2937}main{max-width:1180px;margin:0 auto;padding:32px}h1{font-size:28px;margin:0 0 8px}h2{font-size:16px;margin:0 0 12px}.muted{color:#667085}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:12px;margin:22px 0}.metric,.panel,.project{background:white;border:1px solid #d8dee8;border-radius:8px;padding:16px}.metric .value{font-size:28px;font-weight:700;margin-top:6px}.tables{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:14px}.projects{display:grid;gap:14px;margin-top:18px}.project h3{margin:0 0 10px;font-size:17px}.chips{display:flex;flex-wrap:wrap;gap:6px;margin:8px 0}.chip{background:#eef2ff;color:#3730a3;border-radius:999px;padding:4px 9px;font-size:12px}.chip.warn{background:#fff7ed;color:#9a3412}.notes{display:grid;gap:8px;margin-top:10px}.note{border-left:3px solid #4f46e5;background:#f8fafc;padding:8px 10px}.note strong{display:block;margin-bottom:3px}.layers{font-size:12px;color:#475467;margin-top:8px}table{width:100%;border-collapse:collapse;font-size:13px}td,th{border-bottom:1px solid #e5e7eb;padding:7px 4px;text-align:left}th:last-child,td:last-child{text-align:right}.empty{color:#98a2b3}code{background:#eef2f7;padding:2px 5px;border-radius:4px}</style>")
    [void]$h.AppendLine("</head><body><main>")
    [void]$h.AppendLine("<h1>Technique Showcase Report</h1>")
    [void]$h.AppendLine("<p class=""muted"">input <code>$(Escape-Html $InputPath)</code></p>")
    [void]$h.AppendLine("<div class=""grid"">")
    foreach ($metric in @(
        @{ Label = "Projects"; Value = $summary.project_count },
        @{ Label = "Errors"; Value = $errorCount },
        @{ Label = "Comps"; Value = $summary.totals.comp_count },
        @{ Label = "Layers"; Value = $summary.totals.layer_count },
        @{ Label = "Effects"; Value = $summary.totals.effect_count },
        @{ Label = "Text Animators"; Value = $summary.totals.text_animator_count },
        @{ Label = "Shape Operators"; Value = $summary.totals.shape_operator_count },
        @{ Label = "Dependency Edges"; Value = $summary.totals.dependency_count }
    )) {
        [void]$h.AppendLine("<div class=""metric""><div class=""muted"">$(Escape-Html $metric.Label)</div><div class=""value"">$($metric.Value)</div></div>")
    }
    [void]$h.AppendLine("</div>")
    [void]$h.AppendLine("<div class=""tables"">")
    Write-HtmlCountTable -Builder $h -Title "Technique Hints" -Counts $summary.hint_counts
    Write-HtmlCountTable -Builder $h -Title "Effects" -Counts $summary.effect_counts
    Write-HtmlCountTable -Builder $h -Title "Shape Families" -Counts $summary.shape_families
    Write-HtmlCountTable -Builder $h -Title "Text Animators" -Counts $summary.text_animators
    Write-HtmlCountTable -Builder $h -Title "Layer Roles" -Counts $summary.layer_roles
    Write-HtmlCountTable -Builder $h -Title "Graph Edges" -Counts $summary.graph_edges
    [void]$h.AppendLine("</div>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Projects</h2>")
    [void]$h.AppendLine("<div class=""projects"">")
    foreach ($record in $records) {
        $explanation = $record.explanation
        $portrait = $record.portrait
        if (($null -eq $portrait) -and ($null -ne $explanation)) {
            $portrait = $explanation.portrait
        }
        [void]$h.AppendLine("<article class=""project"">")
        [void]$h.AppendLine("<h3>$(Escape-Html $record.path)</h3>")
        if ($record.error) {
            [void]$h.AppendLine("<div class=""chips""><span class=""chip warn"">$(Escape-Html $record.error)</span></div>")
        } elseif ($null -ne $portrait) {
            [void]$h.AppendLine("<p class=""muted"">$($portrait.fingerprint.comp_count) comps · $($portrait.fingerprint.layer_count) layers · $($portrait.fingerprint.effect_count) effects · $($portrait.graph.edge_count) edges</p>")
            if (($null -ne $explanation) -and ($null -ne $explanation.overview)) {
                foreach ($line in @($explanation.overview | Select-Object -First 2)) {
                    [void]$h.AppendLine("<p>$(Escape-Html $line)</p>")
                }
            }
            [void]$h.AppendLine("<div class=""chips"">")
            foreach ($hint in @($portrait.technique_hints | Select-Object -First 12)) {
                [void]$h.AppendLine("<span class=""chip"">$(Escape-Html $hint.id)</span>")
            }
            [void]$h.AppendLine("</div>")
            if (($null -ne $explanation) -and ($null -ne $explanation.techniques)) {
                [void]$h.AppendLine("<div class=""notes"">")
                foreach ($tech in @($explanation.techniques | Select-Object -First 4)) {
                    [void]$h.AppendLine("<div class=""note""><strong>$(Escape-Html $tech.title)</strong><span>$(Escape-Html $tech.summary)</span></div>")
                }
                [void]$h.AppendLine("</div>")
            }
            if (($null -ne $explanation) -and ($null -ne $explanation.top_signal_layers)) {
                $layers = @($explanation.top_signal_layers | Select-Object -First 3 | ForEach-Object { "$($_.layer_name) [$($_.role)] score=$($_.score)" })
                if ($layers.Count -gt 0) {
                    [void]$h.AppendLine("<div class=""layers"">Top signal layers: $(Escape-Html ($layers -join '; '))</div>")
                }
            }
        }
        [void]$h.AppendLine("</article>")
    }
    [void]$h.AppendLine("</div>")
    [void]$h.AppendLine("</main></body></html>")
    $h.ToString() | Set-Content -Path $htmlPath -Encoding UTF8

    Write-Host "summary: $summaryPath"
    Write-Host "corpus:  $corpusPath"
    Write-Host "report:  $reportPath"
    Write-Host "html:    $htmlPath"
    if ($Open) {
        Invoke-Item $htmlPath
    }
}
finally {
    Pop-Location
}
