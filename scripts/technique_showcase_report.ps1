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
    $projectPlaybooksCsvPath = Join-Path $OutDir "project_playbooks.csv"
    $compositionsCsvPath = Join-Path $OutDir "compositions.csv"
    $layersCsvPath = Join-Path $OutDir "layers.csv"
    $recreationStepsCsvPath = Join-Path $OutDir "recreation_steps.csv"
    $patternsCsvPath = Join-Path $OutDir "patterns.csv"
    $studyQueueCsvPath = Join-Path $OutDir "study_queue.csv"
    $studyTasksCsvPath = Join-Path $OutDir "study_tasks.csv"
    $recreationBlockersCsvPath = Join-Path $OutDir "recreation_blockers.csv"
    $signalLayersCsvPath = Join-Path $OutDir "signal_layers.csv"
    $effectStacksCsvPath = Join-Path $OutDir "effect_stacks.csv"
    $shapeOperatorsCsvPath = Join-Path $OutDir "shape_operators.csv"
    $textAnimatorsCsvPath = Join-Path $OutDir "text_animators.csv"
    $dependencyEdgesCsvPath = Join-Path $OutDir "dependency_edges.csv"
    $learningActionsCsvPath = Join-Path $OutDir "learning_actions.csv"
    $mechanismsCsvPath = Join-Path $OutDir "mechanisms.csv"
    $mechanismExamplesCsvPath = Join-Path $OutDir "mechanism_examples.csv"
    $coverageScorecardCsvPath = Join-Path $OutDir "coverage_scorecard.csv"
    $reconstructionBlueprintsPath = Join-Path $OutDir "reconstruction_blueprints.jsonl"
    $recipeDraftsPath = Join-Path $OutDir "recipe_drafts.jsonl"
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

    function New-CoverageRow {
        param(
            [string]$Artifact,
            [string]$Metric,
            [int]$Expected,
            [int]$Actual,
            [string]$Notes
        )
        $status = "ok"
        if ($Expected -ne $Actual) {
            $status = "mismatch"
        }
        [pscustomobject]@{
            artifact       = $Artifact
            metric         = $Metric
            expected_count = [int]$Expected
            actual_count   = [int]$Actual
            status         = $status
            notes          = $Notes
        }
    }

    function Get-UniqueRecipeName {
        param(
            [AllowNull()][string]$Name,
            [hashtable]$Seen,
            [string]$Fallback
        )
        $baseName = $Name
        if ([string]::IsNullOrWhiteSpace($baseName)) {
            $baseName = $Fallback
        }
        if (-not $Seen.ContainsKey($baseName)) {
            $Seen[$baseName] = 1
            return $baseName
        }
        $Seen[$baseName] = [int]$Seen[$baseName] + 1
        return "$baseName #$($Seen[$baseName])"
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

    function Get-BlockerType {
        param([string]$Blocker)
        if ($Blocker -match "unknown") {
            return "unknown"
        }
        if ($Blocker -match "third-party|plugin") {
            return "plugin"
        }
        return "readiness"
    }

    function Get-BlockerAction {
        param([string]$BlockerType)
        switch ($BlockerType) {
            "unknown" { return "Inspect parser unknown facts before claiming exact recreation." }
            "plugin" { return "Verify plugin availability or document native fallback limits." }
            default { return "Review readiness blocker before 1:1 recreation work." }
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

    function Format-ProjectSteps {
        param([object]$Explanation)
        $items = @($Explanation.recreation_steps |
            Sort-Object @{ Expression = { [int]$_.priority }; Ascending = $true }, id |
            ForEach-Object {
                $risks = ""
                if ($null -ne $_.risks -and @($_.risks).Count -gt 0) {
                    $risks = " risks=" + ((@($_.risks) | ForEach-Object { [string]$_ }) -join ", ")
                }
                "$($_.priority): $($_.title) - $($_.summary)$risks"
            })
        return ($items -join " | ")
    }

    function Format-KeyLayers {
        param([object]$Explanation)
        $items = @($Explanation.top_signal_layers |
            Sort-Object @{ Expression = { [int]$_.score }; Descending = $true }, layer_name |
            Select-Object -First 5 |
            ForEach-Object {
                "$($_.layer_name) [$($_.role)] score=$($_.score)"
            })
        return ($items -join " | ")
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
    $studyTaskRank = 1
    $studyTaskRows = @($mechanismExampleRows |
        Sort-Object @{ Expression = { [int]$_.project_count }; Descending = $true }, @{ Expression = { [int]$_.study_score }; Descending = $true }, category, name, project_path |
        Select-Object -First 200 |
        ForEach-Object {
            $focus = "$($_.category):$($_.name)"
            $reasonParts = @("project count: $($_.project_count)")
            if ($_.patterns) {
                $reasonParts += "patterns: $($_.patterns)"
            }
            $row = [pscustomobject]@{
                rank          = $studyTaskRank
                project_path  = [string]$_.project_path
                focus         = $focus
                category      = [string]$_.category
                name          = [string]$_.name
                project_count = [int]$_.project_count
                readiness     = [string]$_.readiness
                patterns      = [string]$_.patterns
                study_score   = [int]$_.study_score
                action        = [string]$_.action
                reason        = ($reasonParts -join "; ")
            }
            $studyTaskRank++
            $row
        })
    $studyTaskRows | Export-Csv -LiteralPath $studyTasksCsvPath -NoTypeInformation -Encoding UTF8
    $recreationBlockerRows = @()
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation -or $null -eq $explanation.recreation_readiness) {
            continue
        }
        $readiness = [string]$explanation.recreation_readiness.status
        foreach ($blocker in @($explanation.recreation_readiness.blockers)) {
            $blockerText = [string]$blocker
            if ($blockerText -eq "") {
                continue
            }
            $blockerType = Get-BlockerType -Blocker $blockerText
            $recreationBlockerRows += [pscustomobject]@{
                project_path = [string]$record.path
                readiness    = $readiness
                blocker_type = $blockerType
                blocker      = $blockerText
                action       = Get-BlockerAction -BlockerType $blockerType
            }
        }
    }
    if ($recreationBlockerRows.Count -gt 0) {
        $recreationBlockerRows |
            Sort-Object blocker_type, project_path, blocker |
            Export-Csv -LiteralPath $recreationBlockersCsvPath -NoTypeInformation -Encoding UTF8
    } else {
        '"project_path","readiness","blocker_type","blocker","action"' | Set-Content -LiteralPath $recreationBlockersCsvPath -Encoding UTF8
    }
    $signalLayerRows = @()
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation -or $null -eq $explanation.top_signal_layers) {
            continue
        }
        $rank = 1
        foreach ($layer in @($explanation.top_signal_layers)) {
            $signalLayerRows += [pscustomobject]@{
                project_path = [string]$record.path
                rank         = [int]$rank
                comp_name    = [string]$layer.comp_name
                layer_name   = [string]$layer.layer_name
                role         = [string]$layer.role
                score        = [int]$layer.score
                signals      = ((@($layer.signals) | ForEach-Object { [string]$_ }) -join "; ")
            }
            $rank++
        }
    }
    if ($signalLayerRows.Count -gt 0) {
        $signalLayerRows |
            Sort-Object project_path, @{ Expression = { [int]$_.rank }; Ascending = $true } |
            Export-Csv -LiteralPath $signalLayersCsvPath -NoTypeInformation -Encoding UTF8
    } else {
        '"project_path","rank","comp_name","layer_name","role","score","signals"' | Set-Content -LiteralPath $signalLayersCsvPath -Encoding UTF8
    }
    $effectStackRows = @()
    foreach ($record in $records) {
        if ($null -eq $record.facts -or $null -eq $record.facts.effects) {
            continue
        }
        foreach ($effect in @($record.facts.effects)) {
            $effectStackRows += [pscustomobject]@{
                project_path        = [string]$record.path
                comp_name           = [string]$effect.comp_name
                layer_name          = [string]$effect.layer_name
                occurrence          = [int]$effect.occurrence
                match_name          = [string]$effect.match_name
                display_name        = [string]$effect.display_name
                dependency_class    = [string]$effect.dependency_class
                changed_param_count = [int]$effect.changed_param_count
                tuned_param_count   = [int]$effect.tuned_param_count
                unknown_param_count = [int]$effect.unknown_param_count
                has_expression      = [bool]$effect.has_expression
                has_keyframes       = [bool]$effect.has_keyframes
                has_layer_ref       = [bool]$effect.has_layer_ref
            }
        }
    }
    if ($effectStackRows.Count -gt 0) {
        $effectStackRows |
            Sort-Object project_path, comp_name, layer_name, occurrence, match_name |
            Export-Csv -LiteralPath $effectStacksCsvPath -NoTypeInformation -Encoding UTF8
    } else {
        '"project_path","comp_name","layer_name","occurrence","match_name","display_name","dependency_class","changed_param_count","tuned_param_count","unknown_param_count","has_expression","has_keyframes","has_layer_ref"' | Set-Content -LiteralPath $effectStacksCsvPath -Encoding UTF8
    }
    $shapeOperatorRows = @()
    foreach ($record in $records) {
        if ($null -eq $record.facts -or $null -eq $record.facts.shape_operators) {
            continue
        }
        foreach ($operator in @($record.facts.shape_operators)) {
            $shapeOperatorRows += [pscustomobject]@{
                project_path = [string]$record.path
                comp_name    = [string]$operator.comp_name
                layer_name   = [string]$operator.layer_name
                family       = [string]$operator.family
                source       = [string]$operator.source
                match_name   = [string]$operator.match_name
            }
        }
    }
    if ($shapeOperatorRows.Count -gt 0) {
        $shapeOperatorRows |
            Sort-Object project_path, comp_name, layer_name, family, source, match_name |
            Export-Csv -LiteralPath $shapeOperatorsCsvPath -NoTypeInformation -Encoding UTF8
    } else {
        '"project_path","comp_name","layer_name","family","source","match_name"' | Set-Content -LiteralPath $shapeOperatorsCsvPath -Encoding UTF8
    }
    $textAnimatorRows = @()
    foreach ($record in $records) {
        if ($null -eq $record.facts -or $null -eq $record.facts.text_animators) {
            continue
        }
        foreach ($animator in @($record.facts.text_animators)) {
            $textAnimatorRows += [pscustomobject]@{
                project_path     = [string]$record.path
                comp_name        = [string]$animator.comp_name
                layer_name       = [string]$animator.layer_name
                property_kind    = [string]$animator.property_kind
                property_name    = [string]$animator.property_name
                match_name       = [string]$animator.match_name
                has_static_value = [bool]$animator.has_static_value
                has_expression   = [bool]$animator.has_expression
                has_keyframes    = [bool]$animator.has_keyframes
            }
        }
    }
    if ($textAnimatorRows.Count -gt 0) {
        $textAnimatorRows |
            Sort-Object project_path, comp_name, layer_name, property_kind, match_name |
            Export-Csv -LiteralPath $textAnimatorsCsvPath -NoTypeInformation -Encoding UTF8
    } else {
        '"project_path","comp_name","layer_name","property_kind","property_name","match_name","has_static_value","has_expression","has_keyframes"' | Set-Content -LiteralPath $textAnimatorsCsvPath -Encoding UTF8
    }
    $dependencyEdgeRows = @()
    foreach ($record in $records) {
        if ($null -eq $record.facts -or $null -eq $record.facts.dependencies) {
            continue
        }
        foreach ($edge in @($record.facts.dependencies)) {
            $dependencyEdgeRows += [pscustomobject]@{
                project_path = [string]$record.path
                comp_name    = [string]$edge.comp_name
                relation     = [string]$edge.relation
                source_name  = [string]$edge.source_name
                source_id    = [uint32]$edge.source_id
                target_name  = [string]$edge.target_name
                target_id    = [uint32]$edge.target_id
                target_kind  = [string]$edge.target_kind
                property     = [string]$edge.property
            }
        }
    }
    if ($dependencyEdgeRows.Count -gt 0) {
        $dependencyEdgeRows |
            Sort-Object project_path, comp_name, relation, source_name, target_name, property |
            Export-Csv -LiteralPath $dependencyEdgesCsvPath -NoTypeInformation -Encoding UTF8
    } else {
        '"project_path","comp_name","relation","source_name","source_id","target_name","target_id","target_kind","property"' | Set-Content -LiteralPath $dependencyEdgesCsvPath -Encoding UTF8
    }
    $mechanismExampleIndex = @{}
    foreach ($example in $mechanismExampleRows) {
        $key = "$($example.category)`u{1f}$($example.name)"
        $values = @()
        if ($mechanismExampleIndex.ContainsKey($key)) {
            $values = @($mechanismExampleIndex[$key])
        }
        $values += [string]$example.project_path
        $mechanismExampleIndex[$key] = @($values | Select-Object -First 5)
    }

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

    $projectPlaybookRows = @()
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation) {
            continue
        }
        $portrait = $explanation.portrait
        $readiness = ""
        $readinessSummary = ""
        $blockers = ""
        if ($null -ne $explanation.recreation_readiness) {
            $readiness = [string]$explanation.recreation_readiness.status
            $readinessSummary = [string]$explanation.recreation_readiness.summary
            $blockers = ((@($explanation.recreation_readiness.blockers) | ForEach-Object { [string]$_ }) -join "; ")
        }
        $overview = ((@($explanation.overview) | ForEach-Object { [string]$_ }) -join " ")
        $projectPlaybookRows += [pscustomobject]@{
            project_path       = [string]$record.path
            readiness          = $readiness
            overview           = $overview
            readiness_summary  = $readinessSummary
            patterns           = ((@($explanation.patterns) | ForEach-Object { [string]$_.id }) -join "; ")
            archetypes         = ((@($explanation.archetypes) | ForEach-Object { [string]$_.id }) -join "; ")
            ordered_steps      = Format-ProjectSteps -Explanation $explanation
            key_layers         = Format-KeyLayers -Explanation $explanation
            blockers           = $blockers
            top_effects        = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.effect_match_counts -Max 5))
            top_plugin_effects = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.third_party_effect_match_counts -Max 5))
            top_shapes         = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.shape_family_counts -Max 5))
            study_score        = Get-StudyScore -Explanation $explanation
        }
    }
    $projectPlaybookRows |
        Sort-Object @{ Expression = { [int]$_.study_score }; Descending = $true }, project_path |
        Export-Csv -LiteralPath $projectPlaybooksCsvPath -NoTypeInformation -Encoding UTF8

    $compositionRows = @()
    $layerRowsForCsv = @()
    foreach ($record in $records) {
        if ($null -eq $record.facts) {
            continue
        }
        foreach ($comp in @($record.facts.comps)) {
            $compositionRows += [pscustomobject]@{
                project_path   = [string]$record.path
                id             = [uint32]$comp.id
                name           = [string]$comp.name
                width          = [int]$comp.width
                height         = [int]$comp.height
                frame_rate     = [double]$comp.frame_rate
                duration       = [double]$comp.duration_seconds
                layer_count    = [int]$comp.layer_count
                main_candidate = [bool]$comp.main_candidate
            }
        }
        foreach ($layer in @($record.facts.layers)) {
            $layerRowsForCsv += [pscustomobject]@{
                project_path = [string]$record.path
                comp_name    = [string]$layer.comp_name
                id           = [uint32]$layer.id
                index        = [int]$layer.index
                name         = [string]$layer.name
                type         = [string]$layer.type
                role         = [string]$layer.role
                confidence   = [string]$layer.confidence
            }
        }
    }
    $compositionRows |
        Sort-Object project_path, name, id |
        Export-Csv -LiteralPath $compositionsCsvPath -NoTypeInformation -Encoding UTF8
    $layerRowsForCsv |
        Sort-Object project_path, comp_name, index, name |
        Export-Csv -LiteralPath $layersCsvPath -NoTypeInformation -Encoding UTF8

    $recreationStepRows = @()
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation -or $null -eq $explanation.recreation_steps) {
            continue
        }
        $readiness = ""
        if ($null -ne $explanation.recreation_readiness) {
            $readiness = [string]$explanation.recreation_readiness.status
        }
        foreach ($step in @($explanation.recreation_steps)) {
            $recreationStepRows += [pscustomobject]@{
                project_path = [string]$record.path
                readiness    = $readiness
                step_id      = [string]$step.id
                priority     = [int]$step.priority
                title        = [string]$step.title
                summary      = [string]$step.summary
                inputs       = ((@($step.inputs) | ForEach-Object { [string]$_ }) -join "; ")
                risks        = ((@($step.risks) | ForEach-Object { [string]$_ }) -join "; ")
                evidence     = ((@($step.evidence) | ForEach-Object { [string]$_ }) -join "; ")
            }
        }
    }
    $recreationStepRows |
        Sort-Object project_path, @{ Expression = { [int]$_.priority }; Ascending = $true }, step_id |
        Export-Csv -LiteralPath $recreationStepsCsvPath -NoTypeInformation -Encoding UTF8

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

    $reconstructionBlueprintRows = @()
    foreach ($record in $records) {
        $explanation = $record.explanation
        if ($null -eq $explanation) {
            continue
        }
        $portrait = $explanation.portrait
        $readiness = ""
        $readinessSummary = ""
        $blockers = @()
        if ($null -ne $explanation.recreation_readiness) {
            $readiness = [string]$explanation.recreation_readiness.status
            $readinessSummary = [string]$explanation.recreation_readiness.summary
            $blockers = @($explanation.recreation_readiness.blockers | ForEach-Object { [string]$_ } | Where-Object { $_.Trim() -ne "" })
        }
        $steps = @($explanation.recreation_steps | ForEach-Object {
            [ordered]@{
                id       = [string]$_.id
                priority = [int]$_.priority
                title    = [string]$_.title
                summary  = [string]$_.summary
                risks    = @($_.risks | ForEach-Object { [string]$_ } | Where-Object { $_.Trim() -ne "" })
            }
        })
        $keyLayers = @($explanation.top_signal_layers | Select-Object -First 8 | ForEach-Object {
            [ordered]@{
                comp_name  = [string]$_.comp_name
                layer_name = [string]$_.layer_name
                role       = [string]$_.role
                score      = [int]$_.score
                signals    = @($_.signals | ForEach-Object { [string]$_ } | Where-Object { $_.Trim() -ne "" })
            }
        })
        $topMechanisms = [ordered]@{
            effects        = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.effect_match_counts -Max 8))
            plugin_effects = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.third_party_effect_match_counts -Max 8))
            shape_families = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.shape_family_counts -Max 8))
            text_animators = Format-CountList -Rows (Convert-CountRows -Rows (Get-CountRows -Counts $portrait.mechanisms.text_animator_kind_counts -Max 8))
        }
        $mechanismCount = [int]$portrait.fingerprint.effect_count + [int]$portrait.fingerprint.shape_operator_count + [int]$portrait.fingerprint.text_animator_count
        $reconstructionBlueprintRows += [ordered]@{
            schema_version = 1
            project_path   = [string]$record.path
            readiness      = $readiness
            summary        = $readinessSummary
            counts         = [ordered]@{
                compositions     = [int]$portrait.fingerprint.comp_count
                layers           = [int]$portrait.fingerprint.layer_count
                effects          = [int]$portrait.fingerprint.effect_count
                shape_operators  = [int]$portrait.fingerprint.shape_operator_count
                text_animators   = [int]$portrait.fingerprint.text_animator_count
                dependency_edges = [int]$portrait.fingerprint.dependency_count
                recreation_steps = [int]$steps.Count
                blockers         = [int]$blockers.Count
            }
            phases         = @(
                [ordered]@{ id = "create_compositions"; order = 1; expected_count = [int]$portrait.fingerprint.comp_count; artifact = "compositions.csv"; action = "Create comps with parsed size, frame rate, duration, and main-candidate flags." },
                [ordered]@{ id = "create_layers"; order = 2; expected_count = [int]$portrait.fingerprint.layer_count; artifact = "layers.csv"; action = "Create layers in comp/index order and assign parsed roles before mechanism application." },
                [ordered]@{ id = "apply_mechanisms"; order = 3; expected_count = $mechanismCount; artifact = "effect_stacks.csv, shape_operators.csv, text_animators.csv"; action = "Apply effect stacks, shape operators, and text animator properties from mechanism exports." },
                [ordered]@{ id = "wire_dependencies"; order = 4; expected_count = [int]$portrait.fingerprint.dependency_count; artifact = "dependency_edges.csv"; action = "Resolve parent, matte, source, and effect-parameter layer references after all layer IDs are available." },
                [ordered]@{ id = "verify_recreation"; order = 5; expected_count = [int]$steps.Count; artifact = "recreation_steps.csv, recreation_blockers.csv"; action = "Execute ordered recreation checks and review blockers before claiming exact reproduction." }
            )
            blockers       = $blockers
            key_layers     = $keyLayers
            mechanisms     = $topMechanisms
            steps          = $steps
        }
    }
    $reconstructionBlueprintRows |
        ForEach-Object { $_ | ConvertTo-Json -Depth 10 -Compress } |
        Set-Content -LiteralPath $reconstructionBlueprintsPath -Encoding UTF8

    $recipeDraftRows = @()
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
        $nameMap = @{}
        $recipeComps = @()
        $expectedComps = @()
        foreach ($comp in @($record.facts.comps)) {
            $recipeName = Get-UniqueRecipeName -Name ([string]$comp.name) -Seen $nameMap -Fallback ("Comp $($recipeComps.Count + 1)")
            $width = [int]$comp.width
            $height = [int]$comp.height
            $frameRate = [double]$comp.frame_rate
            $duration = [double]$comp.duration_seconds
            $recipeComps += [ordered]@{
                name       = $recipeName
                width      = $width
                height     = $height
                frame_rate = $frameRate
                duration   = $duration
                layers     = @()
            }
            $expectedComps += [ordered]@{
                name       = $recipeName
                width      = [double]$width
                height     = [double]$height
                frame_rate = $frameRate
                duration   = $duration
            }
        }
        $gaps = @()
        if ([int]$portrait.fingerprint.layer_count -gt 0) {
            $gaps += [ordered]@{ id = "layers_not_materialized"; count = [int]$portrait.fingerprint.layer_count; artifact = "layers.csv"; action = "Map parsed layer rows into recipe layer constructors after source, timing, and type support are selected." }
        }
        $mechanismCount = [int]$portrait.fingerprint.effect_count + [int]$portrait.fingerprint.shape_operator_count + [int]$portrait.fingerprint.text_animator_count
        if ($mechanismCount -gt 0) {
            $gaps += [ordered]@{ id = "mechanisms_not_materialized"; count = $mechanismCount; artifact = "effect_stacks.csv, shape_operators.csv, text_animators.csv"; action = "Translate parsed mechanisms into supported recipe effect, shape, and text animator specs." }
        }
        if ([int]$portrait.fingerprint.dependency_count -gt 0) {
            $gaps += [ordered]@{ id = "dependencies_not_materialized"; count = [int]$portrait.fingerprint.dependency_count; artifact = "dependency_edges.csv"; action = "Resolve parent, matte, source, and effect-parameter references after layer IDs are materialized." }
        }
        if ($null -ne $explanation.recreation_readiness -and @($explanation.recreation_readiness.blockers).Count -gt 0) {
            $gaps += [ordered]@{
                id       = "readiness_blockers"
                count    = @($explanation.recreation_readiness.blockers).Count
                artifact = "recreation_blockers.csv"
                action   = ((@($explanation.recreation_readiness.blockers) | ForEach-Object { [string]$_ } | Where-Object { $_.Trim() -ne "" }) -join "; ")
            }
        }
        $gaps += [ordered]@{ id = "verification_required"; count = @($explanation.recreation_steps).Count; artifact = "recreation_steps.csv"; action = "Run project-specific recreation checks before treating the draft as exact." }
        $recipeDraftRows += [ordered]@{
            schema_version = 1
            project_path   = [string]$record.path
            readiness      = $readiness
            recipe         = [ordered]@{
                schema_version   = 1
                project          = [ordered]@{
                    name           = [System.IO.Path]::GetFileNameWithoutExtension([string]$record.path)
                    target_version = "AE2020"
                }
                comps            = $recipeComps
                expected_profile = [ordered]@{
                    comp_count       = [int]$portrait.fingerprint.comp_count
                    layer_count      = [int]$portrait.fingerprint.layer_count
                    text_layer_count = [int]$portrait.fingerprint.text_layer_count
                    shape_layer_count = [int]$portrait.fingerprint.shape_layer_count
                    comps            = $expectedComps
                }
            }
            gaps           = $gaps
        }
    }
    $recipeDraftRows |
        ForEach-Object { $_ | ConvertTo-Json -Depth 10 -Compress } |
        Set-Content -LiteralPath $recipeDraftsPath -Encoding UTF8

    $coverageScorecardRows = @(
        New-CoverageRow -Artifact "projects.csv" -Metric "project rows" -Expected ([int]$summary.project_count) -Actual $projectRowsForCsv.Count -Notes "one explainable project row per parsed project"
        New-CoverageRow -Artifact "project_playbooks.csv" -Metric "project playbooks" -Expected ([int]$summary.project_count) -Actual $projectPlaybookRows.Count -Notes "one recreation playbook per parsed project"
        New-CoverageRow -Artifact "compositions.csv" -Metric "composition rows" -Expected ([int]$summary.totals.comp_count) -Actual $compositionRows.Count -Notes "one row per parsed composition"
        New-CoverageRow -Artifact "layers.csv" -Metric "layer rows" -Expected ([int]$summary.totals.layer_count) -Actual $layerRowsForCsv.Count -Notes "one row per parsed layer"
        New-CoverageRow -Artifact "recreation_steps.csv" -Metric "recreation steps" -Expected $recreationStepRows.Count -Actual $recreationStepRows.Count -Notes "one row per generated project recreation step"
        New-CoverageRow -Artifact "patterns.csv" -Metric "pattern rows" -Expected (@($digest.patterns).Count) -Actual $patternRowsForCsv.Count -Notes "one row per digest pattern"
        New-CoverageRow -Artifact "study_queue.csv" -Metric "study queue rows" -Expected ([int]$summary.project_count) -Actual $studyRows.Count -Notes "one ranked study target per parsed project"
        New-CoverageRow -Artifact "study_tasks.csv" -Metric "study tasks" -Expected $studyTaskRows.Count -Actual $studyTaskRows.Count -Notes "task count depends on observed mechanisms"
        New-CoverageRow -Artifact "recreation_blockers.csv" -Metric "blockers" -Expected $recreationBlockerRows.Count -Actual $recreationBlockerRows.Count -Notes "blocker count depends on readiness analysis"
        New-CoverageRow -Artifact "signal_layers.csv" -Metric "signal layers" -Expected $signalLayerRows.Count -Actual $signalLayerRows.Count -Notes "top explanatory layers selected from each project"
        New-CoverageRow -Artifact "effect_stacks.csv" -Metric "effect rows" -Expected ([int]$summary.totals.effect_count) -Actual $effectStackRows.Count -Notes "one row per parsed effect"
        New-CoverageRow -Artifact "shape_operators.csv" -Metric "shape operator rows" -Expected ([int]$summary.totals.shape_operator_count) -Actual $shapeOperatorRows.Count -Notes "one row per parsed shape operator"
        New-CoverageRow -Artifact "text_animators.csv" -Metric "text animator rows" -Expected ([int]$summary.totals.text_animator_count) -Actual $textAnimatorRows.Count -Notes "one row per parsed text animator property"
        New-CoverageRow -Artifact "dependency_edges.csv" -Metric "dependency edges" -Expected ([int]$summary.totals.dependency_count) -Actual $dependencyEdgeRows.Count -Notes "one row per parsed dependency relation"
        New-CoverageRow -Artifact "learning_actions.csv" -Metric "learning actions" -Expected (@($digest.patterns).Count) -Actual $learningActionRows.Count -Notes "one recommended action per digest pattern"
        New-CoverageRow -Artifact "mechanisms.csv" -Metric "mechanism rows" -Expected $mechanismRows.Count -Actual $mechanismRows.Count -Notes "mechanism count depends on corpus signals"
        New-CoverageRow -Artifact "mechanism_examples.csv" -Metric "mechanism examples" -Expected $mechanismExampleRows.Count -Actual $mechanismExampleRows.Count -Notes "example count depends on mechanism/project intersections"
        New-CoverageRow -Artifact "reconstruction_blueprints.jsonl" -Metric "project blueprints" -Expected ([int]$summary.project_count) -Actual $reconstructionBlueprintRows.Count -Notes "one deterministic reconstruction blueprint per parsed project"
        New-CoverageRow -Artifact "recipe_drafts.jsonl" -Metric "recipe drafts" -Expected ([int]$summary.project_count) -Actual $recipeDraftRows.Count -Notes "one safe recipe skeleton per parsed project"
        New-CoverageRow -Artifact "errors.csv" -Metric "parse errors" -Expected ([int]$errorCount) -Actual $errorRowsForCsv.Count -Notes "one row per per-file parse error"
    )
    $coverageScorecardRows | Export-Csv -LiteralPath $coverageScorecardCsvPath -NoTypeInformation -Encoding UTF8

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
    [void]$h.AppendLine("<p class=""muted"">artifacts <a href=""manifest.json"">manifest.json</a> · <a href=""learning.md"">learning.md</a> · <a href=""projects.csv"">projects.csv</a> · <a href=""project_playbooks.csv"">project_playbooks.csv</a> · <a href=""compositions.csv"">compositions.csv</a> · <a href=""layers.csv"">layers.csv</a> · <a href=""recreation_steps.csv"">recreation_steps.csv</a> · <a href=""patterns.csv"">patterns.csv</a> · <a href=""study_queue.csv"">study_queue.csv</a> · <a href=""study_tasks.csv"">study_tasks.csv</a> · <a href=""recreation_blockers.csv"">recreation_blockers.csv</a> · <a href=""signal_layers.csv"">signal_layers.csv</a> · <a href=""effect_stacks.csv"">effect_stacks.csv</a> · <a href=""shape_operators.csv"">shape_operators.csv</a> · <a href=""text_animators.csv"">text_animators.csv</a> · <a href=""dependency_edges.csv"">dependency_edges.csv</a> · <a href=""learning_actions.csv"">learning_actions.csv</a> · <a href=""mechanisms.csv"">mechanisms.csv</a> · <a href=""mechanism_examples.csv"">mechanism_examples.csv</a> · <a href=""coverage_scorecard.csv"">coverage_scorecard.csv</a> · <a href=""reconstruction_blueprints.jsonl"">reconstruction_blueprints.jsonl</a> · <a href=""recipe_drafts.jsonl"">recipe_drafts.jsonl</a> · <a href=""errors.csv"">errors.csv</a> · <a href=""digest.json"">digest.json</a> · <a href=""summary.json"">summary.json</a> · <a href=""corpus.jsonl"">corpus.jsonl</a> · <a href=""report.md"">report.md</a></p>")
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
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Coverage Scorecard</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Artifact</th><th>Metric</th><th>Expected</th><th>Actual</th><th>Status</th></tr></thead><tbody>")
    foreach ($row in @($coverageScorecardRows)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $row.artifact)</td><td>$(Escape-Html $row.metric)</td><td>$($row.expected_count)</td><td>$($row.actual_count)</td><td>$(Escape-Html $row.status)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Reconstruction Blueprints</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Readiness</th><th>Counts</th><th>Phases</th></tr></thead><tbody>")
    foreach ($blueprint in @($reconstructionBlueprintRows | Select-Object -First 80)) {
        $countText = "comps=$($blueprint.counts.compositions), layers=$($blueprint.counts.layers), mechanisms=$([int]$blueprint.counts.effects + [int]$blueprint.counts.shape_operators + [int]$blueprint.counts.text_animators), deps=$($blueprint.counts.dependency_edges)"
        $phaseText = ((@($blueprint.phases) | ForEach-Object { "$($_.order).$($_.id)=$($_.expected_count)" }) -join "; ")
        [void]$h.AppendLine("<tr><td>$(Escape-Html $blueprint.project_path)</td><td>$(Escape-Html $blueprint.readiness)</td><td>$(Escape-Html $countText)</td><td>$(Escape-Html $phaseText)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Recipe Drafts</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Comps</th><th>Layers</th><th>Gaps</th></tr></thead><tbody>")
    foreach ($draft in @($recipeDraftRows | Select-Object -First 80)) {
        $gapText = ((@($draft.gaps) | ForEach-Object { "$($_.id)=$($_.count)" }) -join "; ")
        [void]$h.AppendLine("<tr><td>$(Escape-Html $draft.project_path)</td><td>$(@($draft.recipe.comps).Count)</td><td>$($draft.recipe.expected_profile.layer_count)</td><td>$(Escape-Html $gapText)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Mechanism Explorer</h2>")
    [void]$h.AppendLine("<div class=""toolbar""><input id=""mechanismFilter"" type=""search"" aria-label=""Filter mechanisms"" placeholder=""Filter by category, name, risk, or action""><span id=""mechanismCount"" class=""muted""></span></div>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Category</th><th>Name</th><th>Risk</th><th>Action</th><th>Representative Projects</th><th>Count</th></tr></thead><tbody>")
    foreach ($mechanism in @($mechanismRows | Sort-Object @{ Expression = { [int]$_.count }; Descending = $true }, category, name | Select-Object -First 120)) {
        $exampleKey = "$($mechanism.category)`u{1f}$($mechanism.name)"
        $representatives = @()
        if ($mechanismExampleIndex.ContainsKey($exampleKey)) {
            $representatives = @($mechanismExampleIndex[$exampleKey])
        }
        $representativeText = ($representatives -join "; ")
        $searchText = ((@($mechanism.category, $mechanism.name, $mechanism.risk, $mechanism.action, $representativeText) | Where-Object { $_ }) -join " ").ToLowerInvariant()
        [void]$h.AppendLine("<tr class=""mechanism-row"" data-search=""$(Escape-Html $searchText)""><td>$(Escape-Html $mechanism.category)</td><td>$(Escape-Html $mechanism.name)</td><td>$(Escape-Html $mechanism.risk)</td><td>$(Escape-Html $mechanism.action)</td><td class=""mechanism-representatives"">$(Escape-Html $representativeText)</td><td>$($mechanism.count)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Study Task Queue</h2>")
    [void]$h.AppendLine("<div class=""toolbar""><input id=""studyTaskFilter"" type=""search"" aria-label=""Filter study tasks"" placeholder=""Filter by project, focus, readiness, pattern, or action""><span id=""studyTaskCount"" class=""muted""></span></div>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Rank</th><th>Project</th><th>Focus</th><th>Action</th><th>Reason</th></tr></thead><tbody>")
    foreach ($task in @($studyTaskRows | Select-Object -First 80)) {
        $searchText = ((@($task.project_path, $task.focus, $task.readiness, $task.patterns, $task.action, $task.reason) | Where-Object { $_ }) -join " ").ToLowerInvariant()
        [void]$h.AppendLine("<tr class=""study-task"" data-search=""$(Escape-Html $searchText)""><td>$($task.rank)</td><td>$(Escape-Html $task.project_path)</td><td>$(Escape-Html $task.focus)</td><td>$(Escape-Html $task.action)</td><td>$(Escape-Html $task.reason)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Project Playbooks</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Readiness</th><th>Key Layers</th><th>Ordered Steps</th></tr></thead><tbody>")
    foreach ($playbook in @($projectPlaybookRows | Select-Object -First 40)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $playbook.project_path)</td><td>$(Escape-Html $playbook.readiness)</td><td>$(Escape-Html $playbook.key_layers)</td><td>$(Escape-Html $playbook.ordered_steps)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Compositions</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Name</th><th>Size</th><th>Duration</th><th>Layers</th></tr></thead><tbody>")
    foreach ($comp in @($compositionRows | Select-Object -First 100)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $comp.project_path)</td><td>$(Escape-Html $comp.name)</td><td>$($comp.width)x$($comp.height)</td><td>$($comp.duration)</td><td>$($comp.layer_count)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Layers</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Comp</th><th>Index</th><th>Name</th><th>Role</th></tr></thead><tbody>")
    foreach ($layer in @($layerRowsForCsv | Select-Object -First 120)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $layer.project_path)</td><td>$(Escape-Html $layer.comp_name)</td><td>$($layer.index)</td><td>$(Escape-Html $layer.name)</td><td>$(Escape-Html $layer.role)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Recreation Steps</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Priority</th><th>Step</th><th>Summary</th><th>Risks</th></tr></thead><tbody>")
    foreach ($step in @($recreationStepRows | Select-Object -First 120)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $step.project_path)</td><td>$($step.priority)</td><td>$(Escape-Html $step.title)</td><td>$(Escape-Html $step.summary)</td><td>$(Escape-Html $step.risks)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Recreation Blockers</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Readiness</th><th>Type</th><th>Blocker</th><th>Action</th></tr></thead><tbody>")
    foreach ($blocker in @($recreationBlockerRows | Select-Object -First 80)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $blocker.project_path)</td><td>$(Escape-Html $blocker.readiness)</td><td>$(Escape-Html $blocker.blocker_type)</td><td>$(Escape-Html $blocker.blocker)</td><td>$(Escape-Html $blocker.action)</td></tr>")
    }
    if ($recreationBlockerRows.Count -eq 0) {
        [void]$h.AppendLine("<tr><td colspan=""5"" class=""empty"">No recreation blockers reported.</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Signal Layers</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Layer</th><th>Role</th><th>Score</th><th>Signals</th></tr></thead><tbody>")
    foreach ($layer in @($signalLayerRows | Sort-Object @{ Expression = { [int]$_.score }; Descending = $true }, project_path | Select-Object -First 80)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $layer.project_path)</td><td>$(Escape-Html $layer.layer_name)</td><td>$(Escape-Html $layer.role)</td><td>$($layer.score)</td><td>$(Escape-Html $layer.signals)</td></tr>")
    }
    if ($signalLayerRows.Count -eq 0) {
        [void]$h.AppendLine("<tr><td colspan=""5"" class=""empty"">No signal layers reported.</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Effect Stacks</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Layer</th><th>Effect</th><th>Params</th><th>Flags</th></tr></thead><tbody>")
    foreach ($effect in @($effectStackRows | Sort-Object @{ Expression = { [int]$_.tuned_param_count }; Descending = $true }, project_path | Select-Object -First 100)) {
        $flags = @()
        if ($effect.has_expression) { $flags += "expression" }
        if ($effect.has_keyframes) { $flags += "keyframes" }
        if ($effect.has_layer_ref) { $flags += "layer_ref" }
        [void]$h.AppendLine("<tr><td>$(Escape-Html $effect.project_path)</td><td>$(Escape-Html $effect.layer_name)</td><td>$(Escape-Html $effect.match_name)</td><td>changed=$($effect.changed_param_count), tuned=$($effect.tuned_param_count), unknown=$($effect.unknown_param_count)</td><td>$(Escape-Html ($flags -join ', '))</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Shape Operators</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Layer</th><th>Family</th><th>Source</th><th>Match</th></tr></thead><tbody>")
    foreach ($operator in @($shapeOperatorRows | Select-Object -First 100)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $operator.project_path)</td><td>$(Escape-Html $operator.layer_name)</td><td>$(Escape-Html $operator.family)</td><td>$(Escape-Html $operator.source)</td><td>$(Escape-Html $operator.match_name)</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Text Animators</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Layer</th><th>Kind</th><th>Match</th><th>Flags</th></tr></thead><tbody>")
    foreach ($animator in @($textAnimatorRows | Select-Object -First 100)) {
        $flags = @()
        if ($animator.has_static_value) { $flags += "static" }
        if ($animator.has_expression) { $flags += "expression" }
        if ($animator.has_keyframes) { $flags += "keyframes" }
        [void]$h.AppendLine("<tr><td>$(Escape-Html $animator.project_path)</td><td>$(Escape-Html $animator.layer_name)</td><td>$(Escape-Html $animator.property_kind)</td><td>$(Escape-Html $animator.match_name)</td><td>$(Escape-Html ($flags -join ', '))</td></tr>")
    }
    [void]$h.AppendLine("</tbody></table></section>")
    [void]$h.AppendLine("<h2 style=""margin-top:28px"">Dependency Edges</h2>")
    [void]$h.AppendLine("<section class=""panel"" style=""margin-top:14px""><table><thead><tr><th>Project</th><th>Relation</th><th>Source</th><th>Target</th><th>Property</th></tr></thead><tbody>")
    foreach ($edge in @($dependencyEdgeRows | Select-Object -First 120)) {
        [void]$h.AppendLine("<tr><td>$(Escape-Html $edge.project_path)</td><td>$(Escape-Html $edge.relation)</td><td>$(Escape-Html $edge.source_name)</td><td>$(Escape-Html $edge.target_name)</td><td>$(Escape-Html $edge.property)</td></tr>")
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
    [void]$h.AppendLine("<script>(function(){const input=document.getElementById('studyTaskFilter');const count=document.getElementById('studyTaskCount');const rows=[...document.querySelectorAll('.study-task')];function apply(){const q=(input.value||'').trim().toLowerCase();let shown=0;for(const row of rows){const ok=!q||row.dataset.search.includes(q);row.hidden=!ok;if(ok)shown++;}count.textContent=shown+' / '+rows.length+' tasks';}input.addEventListener('input',apply);apply();})();</script>")
    [void]$h.AppendLine("<script>(function(){const input=document.getElementById('projectFilter');const count=document.getElementById('projectCount');const cards=[...document.querySelectorAll('.project')];function apply(){const q=(input.value||'').trim().toLowerCase();let shown=0;for(const card of cards){const ok=!q||card.dataset.search.includes(q);card.hidden=!ok;if(ok)shown++;}count.textContent=shown+' / '+cards.length+' projects';}input.addEventListener('input',apply);apply();})();</script>")
    [void]$h.AppendLine("</main></body></html>")
    $h.ToString() | Set-Content -Path $htmlPath -Encoding UTF8

    $artifactPaths = @(
        $summaryPath,
        $corpusPath,
        $digestPath,
        $learningPath,
        $projectsCsvPath,
        $projectPlaybooksCsvPath,
        $compositionsCsvPath,
        $layersCsvPath,
        $recreationStepsCsvPath,
        $patternsCsvPath,
        $studyQueueCsvPath,
        $studyTasksCsvPath,
        $recreationBlockersCsvPath,
        $signalLayersCsvPath,
        $effectStacksCsvPath,
        $shapeOperatorsCsvPath,
        $textAnimatorsCsvPath,
        $dependencyEdgesCsvPath,
        $learningActionsCsvPath,
        $mechanismsCsvPath,
        $mechanismExamplesCsvPath,
        $coverageScorecardCsvPath,
        $reconstructionBlueprintsPath,
        $recipeDraftsPath,
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
    Write-Host "project playbooks csv: $projectPlaybooksCsvPath"
    Write-Host "compositions csv: $compositionsCsvPath"
    Write-Host "layers csv: $layersCsvPath"
    Write-Host "recreation steps csv: $recreationStepsCsvPath"
    Write-Host "patterns csv: $patternsCsvPath"
    Write-Host "study queue csv: $studyQueueCsvPath"
    Write-Host "study tasks csv: $studyTasksCsvPath"
    Write-Host "recreation blockers csv: $recreationBlockersCsvPath"
    Write-Host "signal layers csv: $signalLayersCsvPath"
    Write-Host "effect stacks csv: $effectStacksCsvPath"
    Write-Host "shape operators csv: $shapeOperatorsCsvPath"
    Write-Host "text animators csv: $textAnimatorsCsvPath"
    Write-Host "dependency edges csv: $dependencyEdgesCsvPath"
    Write-Host "learning actions csv: $learningActionsCsvPath"
    Write-Host "mechanisms csv: $mechanismsCsvPath"
    Write-Host "mechanism examples csv: $mechanismExamplesCsvPath"
    Write-Host "coverage scorecard csv: $coverageScorecardCsvPath"
    Write-Host "reconstruction blueprints: $reconstructionBlueprintsPath"
    Write-Host "recipe drafts: $recipeDraftsPath"
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
