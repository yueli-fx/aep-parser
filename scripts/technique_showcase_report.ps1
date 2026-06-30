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
    $digestPath = Join-Path $OutDir "digest.json"
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

    function Get-RepresentativeProjects {
        param(
            [array]$Records,
            [string]$ArchetypeID,
            [int]$Max = 5
        )
        $rows = @()
        foreach ($record in $Records) {
            $explanation = $record.explanation
            if ($null -eq $explanation) {
                continue
            }
            foreach ($archetype in @($explanation.archetypes)) {
                if ($archetype.id -ne $ArchetypeID) {
                    continue
                }
                $overview = ""
                if ($null -ne $explanation.overview -and $explanation.overview.Count -gt 0) {
                    $overview = [string]$explanation.overview[0]
                }
                $readiness = ""
                if ($null -ne $explanation.recreation_readiness) {
                    $readiness = [string]$explanation.recreation_readiness.status
                }
                $rows += [pscustomobject]@{
                    path      = [string]$record.path
                    score     = [int]$archetype.score
                    label     = [string]$archetype.label
                    readiness = $readiness
                    summary   = [string]$archetype.summary
                    overview  = $overview
                }
            }
        }
        return @($rows | Sort-Object @{ Expression = { $_.score }; Descending = $true }, path | Select-Object -First $Max)
    }

    function Get-PatternExamples {
        param(
            [object]$Summary,
            [string]$PatternID,
            [int]$Max = 5
        )
        if ($null -eq $Summary -or $null -eq $Summary.pattern_examples) {
            return @()
        }
        foreach ($property in @($Summary.pattern_examples.PSObject.Properties)) {
            if ($property.Name -ne $PatternID) {
                continue
            }
            return @($property.Value |
                Sort-Object @{ Expression = { [int]$_.score }; Descending = $true }, path |
                Select-Object -First $Max)
        }
        return @()
    }

    function Get-PatternProfile {
        param(
            [object]$Summary,
            [string]$PatternID
        )
        if ($null -eq $Summary -or $null -eq $Summary.pattern_profiles) {
            return $null
        }
        foreach ($property in @($Summary.pattern_profiles.PSObject.Properties)) {
            if ($property.Name -eq $PatternID) {
                return $property.Value
            }
        }
        return $null
    }

    function Convert-CountRows {
        param([array]$Rows)
        return @($Rows | Where-Object { $null -ne $_ -and $_.Name } | ForEach-Object {
            [ordered]@{
                name  = [string]$_.Name
                count = [int]$_.Value
            }
        })
    }

    function Format-CountList {
        param([array]$Rows)
        $items = @($Rows |
            Where-Object { $null -ne $_ -and $_.name -and [int]$_.count -gt 0 } |
            Select-Object -First 5 |
            ForEach-Object { "$($_.name) ($($_.count))" })
        if ($items.Count -eq 0) {
            return ""
        }
        return $items -join ", "
    }

    function Get-ReadinessProjects {
        param(
            [array]$Records,
            [string]$Status,
            [int]$Max = 5
        )
        $rows = @()
        foreach ($record in $Records) {
            $explanation = $record.explanation
            if ($null -eq $explanation -or $null -eq $explanation.recreation_readiness) {
                continue
            }
            if ($explanation.recreation_readiness.status -ne $Status) {
                continue
            }
            $blockers = @()
            if ($null -ne $explanation.recreation_readiness.blockers) {
                $blockers = @($explanation.recreation_readiness.blockers)
            }
            $rows += [pscustomobject]@{
                path     = [string]$record.path
                status   = [string]$Status
                summary  = [string]$explanation.recreation_readiness.summary
                blockers = $blockers
            }
        }
        return @($rows | Sort-Object path | Select-Object -First $Max)
    }

    $b = [System.Text.StringBuilder]::new()
    [void]$b.AppendLine("# Technique Corpus Report")
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
    Write-CountTable -Builder $b -Title "Recreation Readiness" -Counts $summary.readiness_counts
    Write-CountTable -Builder $b -Title "Archetypes" -Counts $summary.archetype_counts
    Write-CountTable -Builder $b -Title "Pattern Catalog" -Counts $summary.pattern_counts
    Write-CountTable -Builder $b -Title "Technique Hints" -Counts $summary.hint_counts
    Write-CountTable -Builder $b -Title "Plugin Effects" -Counts $summary.plugin_effect_counts
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

    $archetypeRows = Get-CountRows -Counts $summary.archetype_counts -Max 12
    $patternRows = Get-CountRows -Counts $summary.pattern_counts -Max 12
    $readinessRows = Get-CountRows -Counts $summary.readiness_counts -Max 12
    $digestArchetypes = @()
    foreach ($row in $archetypeRows) {
        $digestArchetypes += [ordered]@{
            id              = [string]$row.Name
            count           = [int]$row.Value
            representatives = @(Get-RepresentativeProjects -Records $records -ArchetypeID $row.Name -Max 5)
        }
    }
    $digestPatterns = @()
    foreach ($row in $patternRows) {
        $profile = Get-PatternProfile -Summary $summary -PatternID $row.Name
        $digestPatterns += [ordered]@{
            id              = [string]$row.Name
            count           = [int]$row.Value
            representatives = @(Get-PatternExamples -Summary $summary -PatternID $row.Name -Max 5)
            readiness       = @(Convert-CountRows -Rows (Get-CountRows -Counts $profile.readiness_counts -Max 5))
            effects         = @(Convert-CountRows -Rows (Get-CountRows -Counts $profile.effect_counts -Max 8))
            plugin_effects  = @(Convert-CountRows -Rows (Get-CountRows -Counts $profile.plugin_effect_counts -Max 8))
            shape_families  = @(Convert-CountRows -Rows (Get-CountRows -Counts $profile.shape_families -Max 8))
            text_animators  = @(Convert-CountRows -Rows (Get-CountRows -Counts $profile.text_animators -Max 8))
        }
    }
    $digestReadiness = @()
    foreach ($row in $readinessRows) {
        $digestReadiness += [ordered]@{
            status = [string]$row.Name
            count  = [int]$row.Value
            projects = @(Get-ReadinessProjects -Records $records -Status $row.Name -Max 5)
        }
    }
    $digest = [ordered]@{
        input         = $InputPath
        project_count = [int]$summary.project_count
        error_count   = [int]$errorCount
        archetypes    = $digestArchetypes
        patterns      = $digestPatterns
        readiness     = $digestReadiness
    }
    $digest | ConvertTo-Json -Depth 10 | Set-Content -Path $digestPath -Encoding UTF8

    [void]$b.AppendLine("## Representative Projects")
    [void]$b.AppendLine("")
    foreach ($group in $digestArchetypes) {
        [void]$b.AppendLine("### $($group.id)")
        [void]$b.AppendLine("")
        foreach ($project in @($group.representatives)) {
            [void]$b.AppendLine("- ``$($project.path)`` score=$($project.score) readiness=$($project.readiness)")
            if ($project.overview) {
                [void]$b.AppendLine("  $($project.overview)")
            }
        }
        [void]$b.AppendLine("")
    }

    [void]$b.AppendLine("## Pattern Representatives")
    [void]$b.AppendLine("")
    foreach ($group in @($digest.patterns)) {
        [void]$b.AppendLine("### $($group.id)")
        [void]$b.AppendLine("")
        foreach ($line in @(
            @{ Label = "readiness"; Value = Format-CountList -Rows $group.readiness },
            @{ Label = "effects"; Value = Format-CountList -Rows $group.effects },
            @{ Label = "plugin effects"; Value = Format-CountList -Rows $group.plugin_effects },
            @{ Label = "shape families"; Value = Format-CountList -Rows $group.shape_families },
            @{ Label = "text animators"; Value = Format-CountList -Rows $group.text_animators }
        )) {
            if ($line.Value) {
                [void]$b.AppendLine("- $($line.Label): $($line.Value)")
            }
        }
        foreach ($project in @($group.representatives)) {
            [void]$b.AppendLine("- ``$($project.path)`` score=$($project.score) readiness=$($project.readiness)")
            if ($project.summary) {
                [void]$b.AppendLine("  $($project.summary)")
            }
        }
        [void]$b.AppendLine("")
    }

    [void]$b.AppendLine("## Readiness Drilldown")
    [void]$b.AppendLine("")
    foreach ($group in $digestReadiness) {
        [void]$b.AppendLine("### $($group.status)")
        [void]$b.AppendLine("")
        foreach ($project in @($group.projects)) {
            [void]$b.AppendLine("- ``$($project.path)``")
            if ($project.blockers.Count -gt 0) {
                [void]$b.AppendLine("  blockers: $($project.blockers -join '; ')")
            }
        }
        [void]$b.AppendLine("")
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
        if ($null -ne $explanation.recreation_readiness) {
            [void]$b.AppendLine("- Readiness: **$($explanation.recreation_readiness.status)** - $($explanation.recreation_readiness.summary)")
        }
        foreach ($archetype in @($explanation.archetypes | Select-Object -First 4)) {
            [void]$b.AppendLine("- Archetype: **$($archetype.label)** - $($archetype.summary)")
        }
        foreach ($pattern in @($explanation.patterns | Select-Object -First 3)) {
            [void]$b.AppendLine("- Pattern: **$($pattern.label)** - $($pattern.summary)")
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
    [void]$h.AppendLine("<title>Technique Corpus Report</title>")
    [void]$h.AppendLine("<style>")
    [void]$h.AppendLine("body{font-family:Segoe UI,Arial,sans-serif;margin:0;background:#f5f7fa;color:#1f2937}main{max-width:1180px;margin:0 auto;padding:32px}h1{font-size:28px;margin:0 0 8px}h2{font-size:16px;margin:0 0 12px}.muted{color:#667085}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:12px;margin:22px 0}.metric,.panel,.project,.representative{background:white;border:1px solid #d8dee8;border-radius:8px;padding:16px}.metric .value{font-size:28px;font-weight:700;margin-top:6px}.tables{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:14px}.projects,.representatives{display:grid;gap:14px;margin-top:18px}.representatives{grid-template-columns:repeat(auto-fit,minmax(320px,1fr))}.project h3,.representative h3{margin:0 0 10px;font-size:17px}.toolbar{display:flex;gap:12px;align-items:center;margin-top:10px}.toolbar input{width:min(520px,100%);border:1px solid #cfd7e3;border-radius:6px;padding:9px 11px;font:inherit}.chips{display:flex;flex-wrap:wrap;gap:6px;margin:8px 0}.chip{background:#eef2ff;color:#3730a3;border-radius:999px;padding:4px 9px;font-size:12px}.chip.warn{background:#fff7ed;color:#9a3412}.notes{display:grid;gap:8px;margin-top:10px}.note{border-left:3px solid #4f46e5;background:#f8fafc;padding:8px 10px}.note strong{display:block;margin-bottom:3px}.layers,.small{font-size:12px;color:#475467;margin-top:8px}ul.compact{margin:8px 0 0;padding-left:18px}ul.compact li{margin:5px 0}table{width:100%;border-collapse:collapse;font-size:13px}td,th{border-bottom:1px solid #e5e7eb;padding:7px 4px;text-align:left}th:last-child,td:last-child{text-align:right}.empty{color:#98a2b3}code{background:#eef2f7;padding:2px 5px;border-radius:4px}</style>")
    [void]$h.AppendLine("</head><body><main>")
    [void]$h.AppendLine("<h1>Technique Corpus Report</h1>")
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
    Write-HtmlCountTable -Builder $h -Title "Recreation Readiness" -Counts $summary.readiness_counts
    Write-HtmlCountTable -Builder $h -Title "Archetypes" -Counts $summary.archetype_counts
    Write-HtmlCountTable -Builder $h -Title "Pattern Catalog" -Counts $summary.pattern_counts
    Write-HtmlCountTable -Builder $h -Title "Plugin Effects" -Counts $summary.plugin_effect_counts
    Write-HtmlCountTable -Builder $h -Title "Effects" -Counts $summary.effect_counts
    Write-HtmlCountTable -Builder $h -Title "Shape Families" -Counts $summary.shape_families
    Write-HtmlCountTable -Builder $h -Title "Text Animators" -Counts $summary.text_animators
    Write-HtmlCountTable -Builder $h -Title "Layer Roles" -Counts $summary.layer_roles
    Write-HtmlCountTable -Builder $h -Title "Graph Edges" -Counts $summary.graph_edges
    [void]$h.AppendLine("</div>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Representative Projects</h2>")
    [void]$h.AppendLine("<div class=""representatives"">")
    foreach ($group in $digestArchetypes) {
        [void]$h.AppendLine("<section class=""representative"">")
        [void]$h.AppendLine("<h3>$(Escape-Html $group.id)</h3>")
        [void]$h.AppendLine("<p class=""muted"">$($group.count) projects</p>")
        [void]$h.AppendLine("<ul class=""compact"">")
        foreach ($project in @($group.representatives | Select-Object -First 5)) {
            [void]$h.AppendLine("<li><code>$(Escape-Html $project.path)</code><div class=""small"">score=$($project.score) · readiness=$(Escape-Html $project.readiness)</div></li>")
        }
        [void]$h.AppendLine("</ul>")
        [void]$h.AppendLine("</section>")
    }
    [void]$h.AppendLine("</div>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Pattern Representatives</h2>")
    [void]$h.AppendLine("<div class=""representatives"">")
    foreach ($group in @($digest.patterns)) {
        [void]$h.AppendLine("<section class=""representative"">")
        [void]$h.AppendLine("<h3>$(Escape-Html $group.id)</h3>")
        [void]$h.AppendLine("<p class=""muted"">$($group.count) projects</p>")
        foreach ($line in @(
            @{ Label = "readiness"; Value = Format-CountList -Rows $group.readiness },
            @{ Label = "effects"; Value = Format-CountList -Rows $group.effects },
            @{ Label = "plugin effects"; Value = Format-CountList -Rows $group.plugin_effects },
            @{ Label = "shape families"; Value = Format-CountList -Rows $group.shape_families },
            @{ Label = "text animators"; Value = Format-CountList -Rows $group.text_animators }
        )) {
            if ($line.Value) {
                [void]$h.AppendLine("<div class=""small""><strong>$(Escape-Html $line.Label)</strong>: $(Escape-Html $line.Value)</div>")
            }
        }
        [void]$h.AppendLine("<ul class=""compact"">")
        foreach ($project in @($group.representatives | Select-Object -First 5)) {
            [void]$h.AppendLine("<li><code>$(Escape-Html $project.path)</code><div class=""small"">score=$($project.score) · readiness=$(Escape-Html $project.readiness)</div></li>")
        }
        [void]$h.AppendLine("</ul>")
        [void]$h.AppendLine("</section>")
    }
    [void]$h.AppendLine("</div>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Readiness Drilldown</h2>")
    [void]$h.AppendLine("<div class=""representatives"">")
    foreach ($group in $digestReadiness) {
        [void]$h.AppendLine("<section class=""representative"">")
        [void]$h.AppendLine("<h3>$(Escape-Html $group.status)</h3>")
        [void]$h.AppendLine("<p class=""muted"">$($group.count) projects</p>")
        [void]$h.AppendLine("<ul class=""compact"">")
        foreach ($project in @($group.projects | Select-Object -First 5)) {
            [void]$h.AppendLine("<li><code>$(Escape-Html $project.path)</code>")
            if ($project.blockers.Count -gt 0) {
                [void]$h.AppendLine("<div class=""small"">$(Escape-Html ($project.blockers -join '; '))</div>")
            }
            [void]$h.AppendLine("</li>")
        }
        [void]$h.AppendLine("</ul>")
        [void]$h.AppendLine("</section>")
    }
    [void]$h.AppendLine("</div>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Projects</h2>")
    [void]$h.AppendLine("<div class=""toolbar""><input id=""projectFilter"" type=""search"" aria-label=""Filter projects"" placeholder=""Filter by path, readiness, pattern, effect, plugin""><span id=""projectCount"" class=""muted""></span></div>")
    [void]$h.AppendLine("<div class=""projects"">")
    foreach ($record in $records) {
        $explanation = $record.explanation
        $portrait = $record.portrait
        if (($null -eq $portrait) -and ($null -ne $explanation)) {
            $portrait = $explanation.portrait
        }
        $searchTerms = @([string]$record.path)
        if (($null -ne $explanation) -and ($null -ne $explanation.recreation_readiness)) {
            $searchTerms += [string]$explanation.recreation_readiness.status
        }
        if (($null -ne $explanation) -and ($null -ne $explanation.archetypes)) {
            foreach ($archetype in @($explanation.archetypes)) {
                $searchTerms += [string]$archetype.id
                $searchTerms += [string]$archetype.label
            }
        }
        if (($null -ne $explanation) -and ($null -ne $explanation.patterns)) {
            foreach ($pattern in @($explanation.patterns)) {
                $searchTerms += [string]$pattern.id
                $searchTerms += [string]$pattern.label
            }
        }
        if (($null -ne $portrait) -and ($null -ne $portrait.technique_hints)) {
            foreach ($hint in @($portrait.technique_hints)) {
                $searchTerms += [string]$hint.id
            }
        }
        if (($null -ne $portrait) -and ($null -ne $portrait.mechanisms)) {
            foreach ($row in @(Get-CountRows -Counts $portrait.mechanisms.effect_match_counts -Max 80)) {
                $searchTerms += [string]$row.Name
            }
            foreach ($row in @(Get-CountRows -Counts $portrait.mechanisms.third_party_effect_match_counts -Max 120)) {
                $searchTerms += [string]$row.Name
            }
            foreach ($row in @(Get-CountRows -Counts $portrait.mechanisms.shape_family_counts -Max 80)) {
                $searchTerms += [string]$row.Name
            }
            foreach ($row in @(Get-CountRows -Counts $portrait.mechanisms.text_animator_kind_counts -Max 80)) {
                $searchTerms += [string]$row.Name
            }
        }
        $searchText = (($searchTerms | Where-Object { $_ }) -join " ").ToLowerInvariant()
        [void]$h.AppendLine("<article class=""project"" data-search=""$(Escape-Html $searchText)"">")
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
            if (($null -ne $explanation) -and ($null -ne $explanation.recreation_readiness)) {
                $readinessClass = "chip"
                if ($explanation.recreation_readiness.status -ne "analysis_ready") {
                    $readinessClass = "chip warn"
                }
                [void]$h.AppendLine("<span class=""$readinessClass"">$(Escape-Html $explanation.recreation_readiness.status)</span>")
            }
            if (($null -ne $explanation) -and ($null -ne $explanation.archetypes)) {
                foreach ($archetype in @($explanation.archetypes | Select-Object -First 8)) {
                    [void]$h.AppendLine("<span class=""chip"">$(Escape-Html $archetype.label)</span>")
                }
            }
            if (($null -ne $explanation) -and ($null -ne $explanation.patterns)) {
                foreach ($pattern in @($explanation.patterns | Select-Object -First 6)) {
                    [void]$h.AppendLine("<span class=""chip"">$(Escape-Html $pattern.label)</span>")
                }
            }
            foreach ($hint in @($portrait.technique_hints | Select-Object -First 12)) {
                [void]$h.AppendLine("<span class=""chip"">$(Escape-Html $hint.id)</span>")
            }
            [void]$h.AppendLine("</div>")
            if (($null -ne $portrait) -and ($null -ne $portrait.mechanisms)) {
                $mechanismLines = @(
                    @{ Label = "effects"; Value = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.effect_match_counts -Max 5)) },
                    @{ Label = "plugin effects"; Value = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.third_party_effect_match_counts -Max 5)) },
                    @{ Label = "shape families"; Value = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.shape_family_counts -Max 5)) },
                    @{ Label = "text animators"; Value = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.text_animator_kind_counts -Max 5)) }
                )
                foreach ($line in $mechanismLines) {
                    if ($line.Value) {
                        [void]$h.AppendLine("<div class=""small""><strong>$(Escape-Html $line.Label)</strong>: $(Escape-Html $line.Value)</div>")
                    }
                }
            }
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
    [void]$h.AppendLine("<script>(function(){const input=document.getElementById('projectFilter');const count=document.getElementById('projectCount');const cards=[...document.querySelectorAll('.project')];function apply(){const q=(input.value||'').trim().toLowerCase();let shown=0;for(const card of cards){const ok=!q||card.dataset.search.includes(q);card.hidden=!ok;if(ok)shown++;}count.textContent=shown+' / '+cards.length+' projects';}input.addEventListener('input',apply);apply();})();</script>")
    [void]$h.AppendLine("</main></body></html>")
    $h.ToString() | Set-Content -Path $htmlPath -Encoding UTF8

    Write-Host "summary: $summaryPath"
    Write-Host "corpus:  $corpusPath"
    Write-Host "digest:  $digestPath"
    Write-Host "report:  $reportPath"
    Write-Host "html:    $htmlPath"
    if ($Open) {
        Invoke-Item $htmlPath
    }
}
finally {
    Pop-Location
}
