[CmdletBinding()]
param(
    [string]$InputPath = "flightdeck\showcase",
    [string]$OutDir = "tmp\technique_showcase_report",
    [int]$Limit = 0,
    [switch]$Verify,
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
    $learningPath = Join-Path $OutDir "learning.md"
    $projectsCsvPath = Join-Path $OutDir "projects.csv"
    $patternsCsvPath = Join-Path $OutDir "patterns.csv"
    $studyQueueCsvPath = Join-Path $OutDir "study_queue.csv"
    $learningActionsCsvPath = Join-Path $OutDir "learning_actions.csv"
    $mechanismsCsvPath = Join-Path $OutDir "mechanisms.csv"
    $mechanismExamplesCsvPath = Join-Path $OutDir "mechanism_examples.csv"
    $errorsCsvPath = Join-Path $OutDir "errors.csv"
    $manifestPath = Join-Path $OutDir "manifest.json"
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

    $scanTimer = [System.Diagnostics.Stopwatch]::StartNew()
    & go @baseArgs "-out" $corpusPath "-summary-out" $summaryPath
    $goExitCode = $LASTEXITCODE
    $scanTimer.Stop()
    $scanSeconds = [Math]::Round($scanTimer.Elapsed.TotalSeconds, 2)

    $summary = Get-Content -Raw -Path $summaryPath | ConvertFrom-Json
    $errorCount = 0
    if ($null -ne $summary.error_count) {
        $errorCount = $summary.error_count
    }
    if ($goExitCode -ne 0) {
        if ($goExitCode -eq 1 -and $errorCount -gt 0) {
            Write-Warning "aeptechnique reported $errorCount per-file error(s); continuing with partial corpus report."
        } else {
            throw "aeptechnique corpus failed with exit code $goExitCode"
        }
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

    function Get-GitValue {
        param([string[]]$GitArgs)
        $value = @(& git @GitArgs 2>$null)
        if ($LASTEXITCODE -ne 0) {
            return ""
        }
        if ($value.Count -eq 0) {
            return ""
        }
        return (($value -join "`n").Trim())
    }

    function Get-ArtifactRows {
        param([array]$Paths)
        $rows = @()
        foreach ($path in $Paths) {
            if (-not (Test-Path -LiteralPath $path)) {
                continue
            }
            $item = Get-Item -LiteralPath $path
            $rows += [ordered]@{
                name  = [string]$item.Name
                path  = [string]$path
                bytes = [int64]$item.Length
            }
        }
        return $rows
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

    function Get-LearningAction {
        param([string]$PatternID)
        switch ($PatternID) {
            "effect_controlled_shape_system" { return "Study controller references first, then rebuild the shape stack and tuned effect chain." }
            "plugin_dependent_effect_stack" { return "Inventory required plugins, then compare native fallback options against the representative project." }
            "precomp_effect_pipeline" { return "Map the comp nesting order before studying the effect stack on each assembly stage." }
            "kinetic_text_system" { return "Inspect text animator properties and timing before copying layer-level transforms or effects." }
            default { return "Study the representative project and extract the repeated recreation steps for this pattern." }
        }
    }

    function Get-LearningRisk {
        param([object]$Pattern)
        if ((Format-CountList -Rows $Pattern.plugin_effects) -ne "") {
            return "plugins"
        }
        $readiness = Format-CountList -Rows $Pattern.readiness
        if ($readiness -match "needs_reverse_engineering") {
            return "unknowns"
        }
        if ($readiness -match "needs_plugins") {
            return "plugins"
        }
        return "low"
    }

    function Get-MechanismAction {
        param(
            [string]$Category,
            [string]$Name
        )
        switch ($Category) {
            "plugin_effect" { return "Verify plugin availability and collect fallback candidates before claiming recreation." }
            "effect" { return "Inspect representative projects for tuned params, expressions, keyframes, and ordering." }
            "shape_family" { return "Study the vector operator family and map it to recipe/profile coverage." }
            "text_animator" { return "Inspect selector timing and animator value channels in representative text layers." }
            "layer_role" { return "Use this role to prioritize layer-stack reconstruction and signal-layer review." }
            "graph_edge" { return "Trace this dependency type before recreating downstream properties." }
            "hint" { return "Use this technique hint as a corpus filter for deeper manual review." }
            default { return "Review this mechanism in representative projects." }
        }
    }

    function Convert-MechanismRows {
        param(
            [string]$Category,
            [object]$Counts,
            [string]$Risk = ""
        )
        $rows = Get-CountRows -Counts $Counts -Max 100000
        foreach ($row in $rows) {
            [pscustomobject]@{
                category = $Category
                name     = [string]$row.Name
                count    = [int]$row.Value
                risk     = $Risk
                action   = Get-MechanismAction -Category $Category -Name $row.Name
            }
        }
    }

    function Add-MechanismExamples {
        param(
            [hashtable]$Examples,
            [string]$Category,
            [object]$Counts,
            [object]$Record,
            [object]$Explanation
        )
        if ($null -eq $Counts -or $null -eq $Record -or $null -eq $Explanation) {
            return
        }
        $readiness = ""
        if ($null -ne $Explanation.recreation_readiness) {
            $readiness = [string]$Explanation.recreation_readiness.status
        }
        $patterns = ((@($Explanation.patterns) | ForEach-Object { [string]$_.id }) -join "; ")
        $score = Get-StudyScore -Explanation $Explanation
        foreach ($property in @($Counts.PSObject.Properties)) {
            if ([string]$property.Name -eq "" -or [int]$property.Value -le 0) {
                continue
            }
            $key = "$Category`u{1f}$($property.Name)"
            $rows = @()
            if ($Examples.ContainsKey($key)) {
                $rows = @($Examples[$key])
            }
            $rows += [pscustomobject]@{
                category      = $Category
                name          = [string]$property.Name
                project_path  = [string]$Record.path
                project_count = [int]$property.Value
                readiness     = $readiness
                patterns      = $patterns
                study_score   = [int]$score
                action        = Get-MechanismAction -Category $Category -Name $property.Name
            }
            $Examples[$key] = @($rows | Sort-Object @{ Expression = { [int]$_.project_count }; Descending = $true }, @{ Expression = { [int]$_.study_score }; Descending = $true }, project_path | Select-Object -First 5)
        }
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

    function Get-ReadinessRank {
        param([string]$Status)
        switch ($Status) {
            "analysis_ready" { return 0 }
            "needs_reverse_engineering" { return 1 }
            "needs_plugins" { return 2 }
            default { return 3 }
        }
    }

    function Get-StudyScore {
        param([object]$Explanation)
        if ($null -eq $Explanation) {
            return 0
        }
        $fp = $Explanation.portrait.fingerprint
        return (
            (@($Explanation.patterns).Count * 100) +
            (@($Explanation.archetypes).Count * 40) +
            [Math]::Min([int]$fp.effect_count, 120) +
            [Math]::Min([int]$fp.text_animator_count * 2, 120) +
            [Math]::Min([int]([Math]::Floor([int]$fp.shape_operator_count / 10)), 120) +
            [Math]::Min([int]([Math]::Floor([int]$fp.dependency_count / 5)), 120)
        )
    }

    $b = [System.Text.StringBuilder]::new()
    [void]$b.AppendLine("# Technique Corpus Report")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("- input: ``$InputPath``")
    [void]$b.AppendLine("- projects: $($summary.project_count)")
    [void]$b.AppendLine("- errors: $errorCount")
    [void]$b.AppendLine("- scan seconds: $scanSeconds")
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
            recreation_steps = @(Convert-CountRows -Rows (Get-CountRows -Counts $profile.recreation_step_counts -Max 8))
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

    $patternRowsForCsv = @($digest.patterns | ForEach-Object {
        [pscustomobject]@{
            id                      = [string]$_.id
            count                   = [int]$_.count
            readiness               = Format-CountList -Rows $_.readiness
            top_effects             = Format-CountList -Rows $_.effects
            top_plugin_effects      = Format-CountList -Rows $_.plugin_effects
            top_shape_families      = Format-CountList -Rows $_.shape_families
            top_text_animators      = Format-CountList -Rows $_.text_animators
            recreation_steps        = Format-CountList -Rows $_.recreation_steps
            representative_projects = ((@($_.representatives) | Select-Object -First 5 | ForEach-Object { [string]$_.path }) -join "; ")
        }
    })
    $patternRowsForCsv | Export-Csv -LiteralPath $patternsCsvPath -NoTypeInformation -Encoding UTF8

    $learningActionRows = @($digest.patterns | ForEach-Object {
        $representativeProject = ""
        $representativeReadiness = ""
        if (@($_.representatives).Count -gt 0) {
            $representativeProject = [string]$_.representatives[0].path
            $representativeReadiness = [string]$_.representatives[0].readiness
        }
        [pscustomobject]@{
            priority               = [int]$_.count
            pattern                = [string]$_.id
            count                  = [int]$_.count
            action                 = Get-LearningAction -PatternID $_.id
            representative_project = $representativeProject
            representative_readiness = $representativeReadiness
            recreation_steps       = Format-CountList -Rows $_.recreation_steps
            top_effects            = Format-CountList -Rows $_.effects
            top_plugin_effects     = Format-CountList -Rows $_.plugin_effects
            top_shape_families     = Format-CountList -Rows $_.shape_families
            top_text_animators     = Format-CountList -Rows $_.text_animators
            risk                   = Get-LearningRisk -Pattern $_
        }
    } | Sort-Object @{ Expression = { [int]$_.priority }; Descending = $true }, pattern)
    $learningActionRows | Export-Csv -LiteralPath $learningActionsCsvPath -NoTypeInformation -Encoding UTF8

    $mechanismRows = @(
        Convert-MechanismRows -Category "plugin_effect" -Counts $summary.plugin_effect_counts -Risk "plugins"
        Convert-MechanismRows -Category "effect" -Counts $summary.effect_counts
        Convert-MechanismRows -Category "shape_family" -Counts $summary.shape_families
        Convert-MechanismRows -Category "text_animator" -Counts $summary.text_animators
        Convert-MechanismRows -Category "layer_role" -Counts $summary.layer_roles
        Convert-MechanismRows -Category "graph_edge" -Counts $summary.graph_edges
        Convert-MechanismRows -Category "hint" -Counts $summary.hint_counts
    ) | Sort-Object category, @{ Expression = { [int]$_.count }; Descending = $true }, name
    $mechanismRows | Export-Csv -LiteralPath $mechanismsCsvPath -NoTypeInformation -Encoding UTF8

    $mechanismExamples = @{}
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation -or $null -eq $explanation.portrait -or $null -eq $explanation.portrait.mechanisms) {
            continue
        }
        $portrait = $explanation.portrait
        Add-MechanismExamples -Examples $mechanismExamples -Category "effect" -Counts $portrait.mechanisms.effect_match_counts -Record $record -Explanation $explanation
        Add-MechanismExamples -Examples $mechanismExamples -Category "plugin_effect" -Counts $portrait.mechanisms.third_party_effect_match_counts -Record $record -Explanation $explanation
        Add-MechanismExamples -Examples $mechanismExamples -Category "shape_family" -Counts $portrait.mechanisms.shape_family_counts -Record $record -Explanation $explanation
        Add-MechanismExamples -Examples $mechanismExamples -Category "text_animator" -Counts $portrait.mechanisms.text_animator_kind_counts -Record $record -Explanation $explanation
        Add-MechanismExamples -Examples $mechanismExamples -Category "layer_role" -Counts $portrait.fingerprint.layer_role_counts -Record $record -Explanation $explanation
        Add-MechanismExamples -Examples $mechanismExamples -Category "graph_edge" -Counts $portrait.graph.relation_counts -Record $record -Explanation $explanation
    }
    $mechanismExampleRows = @($mechanismExamples.Values | ForEach-Object { $_ } | Sort-Object category, name, @{ Expression = { [int]$_.project_count }; Descending = $true }, project_path)
    $mechanismExampleRows | Export-Csv -LiteralPath $mechanismExamplesCsvPath -NoTypeInformation -Encoding UTF8

    $projectRowsForCsv = @()
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation) {
            continue
        }
        $portrait = $explanation.portrait
        $readiness = ""
        if ($null -ne $explanation.recreation_readiness) {
            $readiness = [string]$explanation.recreation_readiness.status
        }
        $projectRowsForCsv += [pscustomobject]@{
            path                 = [string]$record.path
            readiness            = $readiness
            comps                = [int]$portrait.fingerprint.comp_count
            layers               = [int]$portrait.fingerprint.layer_count
            effects              = [int]$portrait.fingerprint.effect_count
            text_animators       = [int]$portrait.fingerprint.text_animator_count
            shape_operators      = [int]$portrait.fingerprint.shape_operator_count
            dependency_edges     = [int]$portrait.fingerprint.dependency_count
            archetypes           = ((@($explanation.archetypes) | ForEach-Object { [string]$_.id }) -join "; ")
            patterns             = ((@($explanation.patterns) | ForEach-Object { [string]$_.id }) -join "; ")
            top_effects          = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.effect_match_counts -Max 5))
            top_plugin_effects   = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.third_party_effect_match_counts -Max 5))
            top_shape_families   = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.shape_family_counts -Max 5))
            top_text_animators   = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.text_animator_kind_counts -Max 5))
            readiness_blockers   = ((@($explanation.recreation_readiness.blockers) | ForEach-Object { [string]$_ }) -join "; ")
            recreation_steps     = ((@($explanation.recreation_steps) | ForEach-Object { [string]$_.id }) -join "; ")
        }
    }
    $projectRowsForCsv | Export-Csv -LiteralPath $projectsCsvPath -NoTypeInformation -Encoding UTF8

    $errorRowsForCsv = @($records | Where-Object { $_.error } | ForEach-Object {
        [pscustomobject]@{
            path  = [string]$_.path
            mode  = [string]$_.mode
            error = [string]$_.error
        }
    })
    if ($errorRowsForCsv.Count -gt 0) {
        $errorRowsForCsv | Export-Csv -LiteralPath $errorsCsvPath -NoTypeInformation -Encoding UTF8
    } else {
        '"path","mode","error"' | Set-Content -LiteralPath $errorsCsvPath -Encoding UTF8
    }

    $studyRows = @()
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation) {
            continue
        }
        $portrait = $explanation.portrait
        $readiness = ""
        if ($null -ne $explanation.recreation_readiness) {
            $readiness = [string]$explanation.recreation_readiness.status
        }
        $studyRows += [pscustomobject]@{
            rank                = 0
            path                = [string]$record.path
            readiness           = $readiness
            study_score         = Get-StudyScore -Explanation $explanation
            patterns            = ((@($explanation.patterns) | ForEach-Object { [string]$_.id }) -join "; ")
            archetypes          = ((@($explanation.archetypes) | ForEach-Object { [string]$_.id }) -join "; ")
            recreation_steps    = ((@($explanation.recreation_steps) | ForEach-Object { [string]$_.id }) -join "; ")
            top_effects         = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.effect_match_counts -Max 5))
            top_plugin_effects  = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.third_party_effect_match_counts -Max 5))
            top_shape_families  = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.shape_family_counts -Max 5))
            top_text_animators  = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.text_animator_kind_counts -Max 5))
            readiness_blockers  = ((@($explanation.recreation_readiness.blockers) | ForEach-Object { [string]$_ }) -join "; ")
            _readiness_rank     = Get-ReadinessRank -Status $readiness
        }
    }
    $rank = 1
    $studyRows = @($studyRows | Sort-Object @{ Expression = { $_._readiness_rank }; Ascending = $true }, @{ Expression = { [int]$_.study_score }; Descending = $true }, path | ForEach-Object {
        $_.rank = $rank
        $rank++
        $_ | Select-Object rank,path,readiness,study_score,patterns,archetypes,recreation_steps,top_effects,top_plugin_effects,top_shape_families,top_text_animators,readiness_blockers
    })
    $studyRows | Export-Csv -LiteralPath $studyQueueCsvPath -NoTypeInformation -Encoding UTF8

    $learn = [System.Text.StringBuilder]::new()
    [void]$learn.AppendLine("# Technique Learning Index")
    [void]$learn.AppendLine("")
    [void]$learn.AppendLine("- input: ``$InputPath``")
    [void]$learn.AppendLine("- projects: $($summary.project_count)")
    [void]$learn.AppendLine("- errors: $errorCount")
    [void]$learn.AppendLine("- scan seconds: $scanSeconds")
    [void]$learn.AppendLine("")
    [void]$learn.AppendLine("## Study Queue")
    [void]$learn.AppendLine("")
    foreach ($project in @($studyRows | Select-Object -First 10)) {
        [void]$learn.AppendLine("- #$($project.rank) ``$($project.path)`` score=$($project.study_score) readiness=$($project.readiness)")
        if ($project.patterns) {
            [void]$learn.AppendLine("  patterns: $($project.patterns)")
        }
        if ($project.readiness_blockers) {
            [void]$learn.AppendLine("  blockers: $($project.readiness_blockers)")
        }
    }
    [void]$learn.AppendLine("")
    [void]$learn.AppendLine("## Learning Actions")
    [void]$learn.AppendLine("")
    foreach ($action in @($learningActionRows | Select-Object -First 12)) {
        [void]$learn.AppendLine("- [$($action.pattern)] $($action.action)")
        [void]$learn.AppendLine("  representative: ``$($action.representative_project)``")
        if ($action.recreation_steps) {
            [void]$learn.AppendLine("  recreation steps: $($action.recreation_steps)")
        }
        if ($action.risk -ne "low") {
            [void]$learn.AppendLine("  risk: $($action.risk)")
        }
    }
    [void]$learn.AppendLine("")
    [void]$learn.AppendLine("## Pattern Playbook")
    [void]$learn.AppendLine("")
    foreach ($group in @($digest.patterns)) {
        [void]$learn.AppendLine("### $($group.id) ($($group.count))")
        [void]$learn.AppendLine("")
        foreach ($line in @(
            @{ Label = "readiness"; Value = Format-CountList -Rows $group.readiness },
            @{ Label = "effects"; Value = Format-CountList -Rows $group.effects },
            @{ Label = "plugin effects"; Value = Format-CountList -Rows $group.plugin_effects },
            @{ Label = "shape families"; Value = Format-CountList -Rows $group.shape_families },
            @{ Label = "text animators"; Value = Format-CountList -Rows $group.text_animators },
            @{ Label = "recreation steps"; Value = Format-CountList -Rows $group.recreation_steps }
        )) {
            if ($line.Value) {
                [void]$learn.AppendLine("- $($line.Label): $($line.Value)")
            }
        }
        [void]$learn.AppendLine("- representative projects:")
        foreach ($project in @($group.representatives | Select-Object -First 5)) {
            [void]$learn.AppendLine("  - ``$($project.path)`` score=$($project.score) readiness=$($project.readiness)")
        }
        [void]$learn.AppendLine("")
    }
    [void]$learn.AppendLine("## Plugin Risk Queue")
    [void]$learn.AppendLine("")
    foreach ($row in @(Get-CountRows -Counts $summary.plugin_effect_counts -Max 20)) {
        [void]$learn.AppendLine("- $($row.Name): $($row.Value)")
    }
    [void]$learn.AppendLine("")
    [void]$learn.AppendLine("## Readiness Queue")
    [void]$learn.AppendLine("")
    foreach ($group in $digestReadiness) {
        [void]$learn.AppendLine("### $($group.status) ($($group.count))")
        [void]$learn.AppendLine("")
        foreach ($project in @($group.projects | Select-Object -First 5)) {
            [void]$learn.AppendLine("- ``$($project.path)``")
            if ($project.blockers.Count -gt 0) {
                [void]$learn.AppendLine("  blockers: $($project.blockers -join '; ')")
            }
        }
        [void]$learn.AppendLine("")
    }
    $learn.ToString() | Set-Content -Path $learningPath -Encoding UTF8

    [void]$b.AppendLine("## Study Queue")
    [void]$b.AppendLine("")
    foreach ($project in @($studyRows | Select-Object -First 10)) {
        [void]$b.AppendLine("- #$($project.rank) ``$($project.path)`` score=$($project.study_score) readiness=$($project.readiness)")
        if ($project.patterns) {
            [void]$b.AppendLine("  patterns: $($project.patterns)")
        }
        if ($project.readiness_blockers) {
            [void]$b.AppendLine("  blockers: $($project.readiness_blockers)")
        }
    }
    [void]$b.AppendLine("")

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
            @{ Label = "text animators"; Value = Format-CountList -Rows $group.text_animators },
            @{ Label = "recreation steps"; Value = Format-CountList -Rows $group.recreation_steps }
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
        foreach ($step in @($explanation.recreation_steps | Select-Object -First 6)) {
            $stepLine = "- Step $($step.priority): **$($step.title)** - $($step.summary)"
            if ($step.risks.Count -gt 0) {
                $stepLine += " Risks: $($step.risks -join '; ')"
            }
            [void]$b.AppendLine($stepLine)
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
    [void]$h.AppendLine("<p class=""muted"">artifacts <a href=""manifest.json"">manifest.json</a> · <a href=""learning.md"">learning.md</a> · <a href=""projects.csv"">projects.csv</a> · <a href=""patterns.csv"">patterns.csv</a> · <a href=""study_queue.csv"">study_queue.csv</a> · <a href=""learning_actions.csv"">learning_actions.csv</a> · <a href=""mechanisms.csv"">mechanisms.csv</a> · <a href=""mechanism_examples.csv"">mechanism_examples.csv</a> · <a href=""errors.csv"">errors.csv</a> · <a href=""digest.json"">digest.json</a> · <a href=""summary.json"">summary.json</a> · <a href=""corpus.jsonl"">corpus.jsonl</a> · <a href=""report.md"">report.md</a></p>")
    [void]$h.AppendLine("<div class=""grid"">")
    foreach ($metric in @(
        @{ Label = "Projects"; Value = $summary.project_count },
        @{ Label = "Errors"; Value = $errorCount },
        @{ Label = "Scan Seconds"; Value = $scanSeconds },
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
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Mechanism Explorer</h2>")
    [void]$h.AppendLine("<div class=""toolbar""><input id=""mechanismFilter"" type=""search"" aria-label=""Filter mechanisms"" placeholder=""Filter by category, name, risk, or action""><span id=""mechanismCount"" class=""muted""></span></div>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Category</th><th>Name</th><th>Risk</th><th>Action</th><th>Count</th></tr></thead><tbody>")
    foreach ($mechanism in @($mechanismRows | Sort-Object @{ Expression = { [int]$_.count }; Descending = $true }, category, name | Select-Object -First 120)) {
        $searchText = ((@($mechanism.category, $mechanism.name, $mechanism.risk, $mechanism.action) | Where-Object { $_ }) -join " ").ToLowerInvariant()
        [void]$h.AppendLine("<tr class=""mechanism-row"" data-search=""$(Escape-Html $searchText)""><td>$(Escape-Html $mechanism.category)</td><td>$(Escape-Html $mechanism.name)</td><td>$(Escape-Html $mechanism.risk)</td><td>$(Escape-Html $mechanism.action)</td><td>$($mechanism.count)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Study Queue</h2>")
    [void]$h.AppendLine("<div class=""representatives"">")
    foreach ($project in @($studyRows | Select-Object -First 6)) {
        [void]$h.AppendLine("<section class=""representative"">")
        [void]$h.AppendLine("<h3>#$($project.rank) $(Escape-Html $project.readiness)</h3>")
        [void]$h.AppendLine("<p><code>$(Escape-Html $project.path)</code></p>")
        [void]$h.AppendLine("<div class=""small"">score=$($project.study_score)</div>")
        if ($project.patterns) {
            [void]$h.AppendLine("<div class=""small""><strong>patterns</strong>: $(Escape-Html $project.patterns)</div>")
        }
        if ($project.readiness_blockers) {
            [void]$h.AppendLine("<div class=""small""><strong>blockers</strong>: $(Escape-Html $project.readiness_blockers)</div>")
        }
        [void]$h.AppendLine("</section>")
    }
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
            @{ Label = "text animators"; Value = Format-CountList -Rows $group.text_animators },
            @{ Label = "recreation steps"; Value = Format-CountList -Rows $group.recreation_steps }
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
            if (($null -ne $explanation) -and (($null -ne $explanation.recreation_steps) -or ($null -ne $explanation.techniques))) {
                [void]$h.AppendLine("<div class=""notes"">")
                foreach ($step in @($explanation.recreation_steps | Select-Object -First 4)) {
                    $riskText = ""
                    if ($step.risks.Count -gt 0) {
                        $riskText = " Risks: " + ($step.risks -join "; ")
                    }
                    [void]$h.AppendLine("<div class=""note""><strong>Step $($step.priority): $(Escape-Html $step.title)</strong><span>$(Escape-Html ($step.summary + $riskText))</span></div>")
                }
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
    [void]$h.AppendLine("<script>(function(){const input=document.getElementById('mechanismFilter');const count=document.getElementById('mechanismCount');const rows=[...document.querySelectorAll('.mechanism-row')];function apply(){const q=(input.value||'').trim().toLowerCase();let shown=0;for(const row of rows){const ok=!q||row.dataset.search.includes(q);row.hidden=!ok;if(ok)shown++;}count.textContent=shown+' / '+rows.length+' mechanisms';}input.addEventListener('input',apply);apply();})();</script>")
    [void]$h.AppendLine("<script>(function(){const input=document.getElementById('projectFilter');const count=document.getElementById('projectCount');const cards=[...document.querySelectorAll('.project')];function apply(){const q=(input.value||'').trim().toLowerCase();let shown=0;for(const card of cards){const ok=!q||card.dataset.search.includes(q);card.hidden=!ok;if(ok)shown++;}count.textContent=shown+' / '+cards.length+' projects';}input.addEventListener('input',apply);apply();})();</script>")
    [void]$h.AppendLine("</main></body></html>")
    $h.ToString() | Set-Content -Path $htmlPath -Encoding UTF8

    $artifactPaths = @(
        $summaryPath,
        $corpusPath,
        $digestPath,
        $learningPath,
        $projectsCsvPath,
        $patternsCsvPath,
        $studyQueueCsvPath,
        $learningActionsCsvPath,
        $mechanismsCsvPath,
        $mechanismExamplesCsvPath,
        $errorsCsvPath,
        $reportPath,
        $htmlPath
    )
    $gitStatus = Get-GitValue -GitArgs @("status", "--short")
    $commandParts = @(
        "pwsh",
        "-NoProfile",
        "-File",
        "scripts\technique_showcase_report.ps1",
        "-InputPath",
        $InputPath,
        "-OutDir",
        $OutDir
    )
    if ($Limit -gt 0) {
        $commandParts += @("-Limit", "$Limit")
    }
    if ($Verify) {
        $commandParts += "-Verify"
    }
    if ($Open) {
        $commandParts += "-Open"
    }

    $manifest = [ordered]@{
        schema_version   = 1
        generated_at_utc = [DateTime]::UtcNow.ToString("o")
        input_path       = $InputPath
        out_dir          = $OutDir
        mode             = "explain"
        recursive        = $true
        limit            = [int]$Limit
        command          = ($commandParts -join " ")
        go_exit_code     = [int]$goExitCode
        scan_seconds     = [double]$scanSeconds
        project_count    = [int]$summary.project_count
        error_count      = [int]$errorCount
        pattern_count    = [int]@($digest.patterns).Count
        git              = [ordered]@{
            commit         = Get-GitValue -GitArgs @("rev-parse", "--short", "HEAD")
            branch         = Get-GitValue -GitArgs @("branch", "--show-current")
            worktree_dirty = ($gitStatus -ne "")
        }
        artifacts        = @(Get-ArtifactRows -Paths $artifactPaths)
    }
    $manifest | ConvertTo-Json -Depth 6 | Set-Content -Path $manifestPath -Encoding UTF8

    Write-Host "summary: $summaryPath"
    Write-Host "corpus:  $corpusPath"
    Write-Host "digest:  $digestPath"
    Write-Host "learn:   $learningPath"
    Write-Host "projects csv: $projectsCsvPath"
    Write-Host "patterns csv: $patternsCsvPath"
    Write-Host "study queue csv: $studyQueueCsvPath"
    Write-Host "learning actions csv: $learningActionsCsvPath"
    Write-Host "mechanisms csv: $mechanismsCsvPath"
    Write-Host "mechanism examples csv: $mechanismExamplesCsvPath"
    Write-Host "errors csv: $errorsCsvPath"
    Write-Host "manifest: $manifestPath"
    Write-Host "report:  $reportPath"
    Write-Host "html:    $htmlPath"
    if ($Verify) {
        & pwsh -NoProfile -File (Join-Path $PSScriptRoot "verify_technique_report.ps1") -OutDir $OutDir
        if ($LASTEXITCODE -ne 0) {
            throw "technique report verification failed with exit code $LASTEXITCODE"
        }
    }
    if ($Open) {
        Invoke-Item $htmlPath
    }
}
finally {
    Pop-Location
}
