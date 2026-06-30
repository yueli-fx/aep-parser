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
    $recipeDraftCompileDir = Join-Path $runRoot "recipe_draft_compile"
    $recipeDraftReparseDir = Join-Path $runRoot "recipe_draft_reparse"
    $recipeDraftBatchDir = Join-Path $runRoot "recipe_draft_batch"
    $acceptanceJsonPath = Join-Path $runRoot "acceptance.json"
    $acceptanceMdPath = Join-Path $runRoot "acceptance.md"
    $outcomeMdPath = Join-Path $runRoot "outcome.md"
    $outcomeHtmlPath = Join-Path $runRoot "outcome.html"
    $outcomeJsonPath = Join-Path $runRoot "outcome.json"
    $effectivenessMdPath = Join-Path $runRoot "effectiveness.md"
    $effectivenessJsonPath = Join-Path $runRoot "effectiveness.json"
    $latestRunPath = Join-Path $OutRoot "latest_run.txt"
    $latestAcceptancePath = Join-Path $OutRoot "latest_acceptance.md"
    $latestOutcomePath = Join-Path $OutRoot "latest_outcome.md"
    $latestOutcomeHtmlPath = Join-Path $OutRoot "latest_outcome.html"
    $latestOutcomeJsonPath = Join-Path $OutRoot "latest_outcome.json"
    $latestEffectivenessPath = Join-Path $OutRoot "latest_effectiveness.md"
    $latestEffectivenessJsonPath = Join-Path $OutRoot "latest_effectiveness.json"
    $historyJsonlPath = Join-Path $OutRoot "history.jsonl"
    $historyCsvPath = Join-Path $OutRoot "history.csv"
    $latestIndexPath = Join-Path $OutRoot "latest_index.html"
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
    Invoke-GateStep -Name "recipe draft compile smoke" -Body {
        New-Item -ItemType Directory -Force -Path $recipeDraftCompileDir | Out-Null
        $draftRecipePath = Join-Path $recipeDraftCompileDir "recipe_draft.json"
        $compiledAEP = Join-Path $recipeDraftCompileDir "recipe_draft.aep"
        $validateJsonPath = Join-Path $recipeDraftCompileDir "validate.json"
        $compileJsonPath = Join-Path $recipeDraftCompileDir "compile.json"
        $draftLine = Get-Content -LiteralPath (Join-Path $fullReportDir "recipe_drafts.jsonl") | Where-Object { $_.Trim() -ne "" } | Select-Object -First 1
        if ([string]::IsNullOrWhiteSpace($draftLine)) {
            Write-Error "recipe_drafts.jsonl has no rows"
            exit 1
        }
        $draft = $draftLine | ConvertFrom-Json
        $draft.recipe | ConvertTo-Json -Depth 12 | Set-Content -LiteralPath $draftRecipePath -Encoding UTF8
        $validateOutput = & go run ./cmd/aeprecipe validate -recipe $draftRecipePath -json 2>&1
        $validateExit = $LASTEXITCODE
        $validateOutput | Set-Content -LiteralPath $validateJsonPath -Encoding UTF8
        if ($validateExit -ne 0) {
            Write-Error "recipe draft validate failed with exit code $validateExit"
            exit $validateExit
        }
        $compileOutput = & go run ./cmd/aeprecipe compile -recipe $draftRecipePath -out $compiledAEP -json 2>&1
        $compileExit = $LASTEXITCODE
        $compileOutput | Set-Content -LiteralPath $compileJsonPath -Encoding UTF8
        if ($compileExit -ne 0) {
            Write-Error "recipe draft compile failed with exit code $compileExit"
            exit $compileExit
        }
        $compiledAEP = Join-Path $recipeDraftCompileDir "recipe_draft.aep"
        if (-not (Test-Path -LiteralPath $compiledAEP)) {
            Write-Error "missing compiled recipe draft: $compiledAEP"
            exit 1
        }
    }
    Invoke-GateStep -Name "compiled recipe draft reparse smoke" -Body {
        New-Item -ItemType Directory -Force -Path $recipeDraftReparseDir | Out-Null
        $reparseFacts = Join-Path $recipeDraftReparseDir "compiled_facts.json"
        $reparseSummaryPath = Join-Path $recipeDraftReparseDir "reparse_summary.json"
        $compiledAEP = Join-Path $recipeDraftCompileDir "recipe_draft.aep"
        $draftRecipePath = Join-Path $recipeDraftCompileDir "recipe_draft.json"
        & go run ./cmd/aeptechnique -in $compiledAEP -mode facts -out $reparseFacts
        $reparseExit = $LASTEXITCODE
        if ($reparseExit -ne 0) {
            Write-Error "compiled recipe draft reparse failed with exit code $reparseExit"
            exit $reparseExit
        }
        $facts = Get-Content -Raw -LiteralPath $reparseFacts | ConvertFrom-Json
        $draftRecipe = Get-Content -Raw -LiteralPath $draftRecipePath | ConvertFrom-Json
        $expectedCompCount = [int]$draftRecipe.expected_profile.comp_count
        $expectedLayerCount = 0
        if ($null -ne $draftRecipe.expected_profile.layer_count) {
            $expectedLayerCount = [int]$draftRecipe.expected_profile.layer_count
        }
        $actualCompCount = [int]$facts.summary.comp_count
        $actualLayerCount = [int]$facts.summary.layer_count
        $summary = [ordered]@{
            schema_version       = 1
            compiled_aep         = $compiledAEP
            facts_json           = $reparseFacts
            expected_comp_count  = $expectedCompCount
            actual_comp_count    = $actualCompCount
            expected_layer_count = $expectedLayerCount
            actual_layer_count   = $actualLayerCount
            passed               = ($expectedCompCount -eq $actualCompCount -and $expectedLayerCount -eq $actualLayerCount)
        }
        $summary | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $reparseSummaryPath -Encoding UTF8
        if (-not $summary.passed) {
            Write-Error "compiled recipe draft reparse mismatch: comps $actualCompCount/$expectedCompCount layers $actualLayerCount/$expectedLayerCount"
            exit 1
        }
    }
    Invoke-GateStep -Name "recipe draft batch smoke" -Body {
        New-Item -ItemType Directory -Force -Path $recipeDraftBatchDir | Out-Null
        $batchSummaryPath = Join-Path $recipeDraftBatchDir "summary.json"
        $draftLines = @(Get-Content -LiteralPath (Join-Path $fullReportDir "recipe_drafts.jsonl") | Where-Object { $_.Trim() -ne "" } | Select-Object -First 3)
        if ($draftLines.Count -eq 0) {
            Write-Error "recipe_drafts.jsonl has no rows for batch smoke"
            exit 1
        }
        $cases = [System.Collections.ArrayList]::new()
        $caseIndex = 0
        foreach ($draftLine in $draftLines) {
            $caseIndex++
            $caseID = "{0:d3}" -f $caseIndex
            $caseDir = Join-Path $recipeDraftBatchDir $caseID
            New-Item -ItemType Directory -Force -Path $caseDir | Out-Null
            $draftRecipePath = Join-Path $caseDir "recipe.json"
            $compiledAEP = Join-Path $caseDir "recipe.aep"
            $validateJsonPath = Join-Path $caseDir "validate.json"
            $compileJsonPath = Join-Path $caseDir "compile.json"
            $factsPath = Join-Path $caseDir "facts.json"
            $reparseSummaryPath = Join-Path $caseDir "reparse_summary.json"

            $draft = $draftLine | ConvertFrom-Json
            $draft.recipe | ConvertTo-Json -Depth 12 | Set-Content -LiteralPath $draftRecipePath -Encoding UTF8
            $validateOutput = & go run ./cmd/aeprecipe validate -recipe $draftRecipePath -json 2>&1
            $validateExit = $LASTEXITCODE
            $validateOutput | Set-Content -LiteralPath $validateJsonPath -Encoding UTF8
            if ($validateExit -ne 0) {
                Write-Error "batch recipe draft validate failed for case $caseID with exit code $validateExit"
                exit $validateExit
            }
            $compileOutput = & go run ./cmd/aeprecipe compile -recipe $draftRecipePath -out $compiledAEP -json 2>&1
            $compileExit = $LASTEXITCODE
            $compileOutput | Set-Content -LiteralPath $compileJsonPath -Encoding UTF8
            if ($compileExit -ne 0) {
                Write-Error "batch recipe draft compile failed for case $caseID with exit code $compileExit"
                exit $compileExit
            }
            if (-not (Test-Path -LiteralPath $compiledAEP)) {
                Write-Error "batch recipe draft output missing for case ${caseID}: $compiledAEP"
                exit 1
            }

            & go run ./cmd/aeptechnique -in $compiledAEP -mode facts -out $factsPath
            $reparseExit = $LASTEXITCODE
            if ($reparseExit -ne 0) {
                Write-Error "batch recipe draft reparse failed for case $caseID with exit code $reparseExit"
                exit $reparseExit
            }
            $facts = Get-Content -Raw -LiteralPath $factsPath | ConvertFrom-Json
            $expectedCompCount = [int]$draft.recipe.expected_profile.comp_count
            $expectedLayerCount = 0
            if ($null -ne $draft.recipe.expected_profile.layer_count) {
                $expectedLayerCount = [int]$draft.recipe.expected_profile.layer_count
            }
            $actualCompCount = [int]$facts.summary.comp_count
            $actualLayerCount = [int]$facts.summary.layer_count
            $passed = ($expectedCompCount -eq $actualCompCount -and $expectedLayerCount -eq $actualLayerCount)
            $caseSummary = [ordered]@{
                case_id = $caseID
                project_path = $draft.project_path
                dir = $caseDir
                recipe = $draftRecipePath
                output = $compiledAEP
                output_bytes = [int64](Get-Item -LiteralPath $compiledAEP).Length
                validate_json = $validateJsonPath
                compile_json = $compileJsonPath
                facts_json = $factsPath
                reparse_summary_json = $reparseSummaryPath
                expected_comp_count = $expectedCompCount
                actual_comp_count = $actualCompCount
                expected_layer_count = $expectedLayerCount
                actual_layer_count = $actualLayerCount
                passed = [bool]$passed
            }
            $caseSummary | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $reparseSummaryPath -Encoding UTF8
            [void]$cases.Add($caseSummary)
            if (-not $passed) {
                Write-Error "batch recipe draft reparse mismatch for case ${caseID}: comps $actualCompCount/$expectedCompCount layers $actualLayerCount/$expectedLayerCount"
                exit 1
            }
        }
        $passedCases = @($cases | Where-Object { $_.passed }).Count
        $summary = [ordered]@{
            schema_version = 1
            requested = 3
            attempted = [int]$cases.Count
            passed = [int]$passedCases
            cases = @($cases)
        }
        $summary | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $batchSummaryPath -Encoding UTF8
        if ([int]$summary.passed -ne [int]$summary.attempted) {
            Write-Error "batch recipe draft smoke passed $($summary.passed)/$($summary.attempted)"
            exit 1
        }
    }
    $recipeDraftBatchSummaryPath = Join-Path $recipeDraftBatchDir "summary.json"
    if (-not (Test-Path -LiteralPath $recipeDraftBatchSummaryPath)) {
        throw "recipe draft batch summary missing: $recipeDraftBatchSummaryPath"
    }
    $recipeDraftBatchSummary = Get-Content -Raw -LiteralPath $recipeDraftBatchSummaryPath | ConvertFrom-Json

    $fullManifest = Get-Content -Raw -LiteralPath (Join-Path $fullReportDir "manifest.json") | ConvertFrom-Json
    $partialManifest = Get-Content -Raw -LiteralPath (Join-Path $partialReportDir "manifest.json") | ConvertFrom-Json
    $compareSelf = Get-Content -Raw -LiteralPath (Join-Path $compareSelfDir "compare.json") | ConvertFrom-Json
    $comparePartial = Get-Content -Raw -LiteralPath (Join-Path $comparePartialDir "compare.json") | ConvertFrom-Json
    $recipeDraftCompileJson = Get-Content -Raw -LiteralPath (Join-Path $recipeDraftCompileDir "compile.json") | ConvertFrom-Json
    $compiledRecipeDraftAEP = Get-Item -LiteralPath (Join-Path $recipeDraftCompileDir "recipe_draft.aep")
    $recipeDraftReparseSummary = Get-Content -Raw -LiteralPath (Join-Path $recipeDraftReparseDir "reparse_summary.json") | ConvertFrom-Json
    $projectRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "projects.csv"))
    $allRecreationBlockerRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "recreation_blockers.csv"))
    $studyRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "study_queue.csv") | Select-Object -First 5)
    $projectPlaybookRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "project_playbooks.csv") | Select-Object -First 5)
    $compositionRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "compositions.csv") | Select-Object -First 5)
    $layerRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "layers.csv") | Select-Object -First 5)
    $recreationStepRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "recreation_steps.csv") | Select-Object -First 5)
    $patternRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "patterns.csv") | Select-Object -First 5)
    $recreationBlockerRows = @($allRecreationBlockerRows | Select-Object -First 5)
    $signalLayerRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "signal_layers.csv") | Select-Object -First 5)
    $effectStackRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "effect_stacks.csv") | Select-Object -First 5)
    $shapeOperatorRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "shape_operators.csv") | Select-Object -First 5)
    $textAnimatorRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "text_animators.csv") | Select-Object -First 5)
    $dependencyEdgeRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "dependency_edges.csv") | Select-Object -First 5)
    $learningActionRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "learning_actions.csv") | Select-Object -First 5)
    $mechanismRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "mechanisms.csv") | Select-Object -First 8)
    $coverageScorecardRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "coverage_scorecard.csv") | Select-Object -First 12)
    $reconstructionBlueprintRows = @(Get-Content -LiteralPath (Join-Path $fullReportDir "reconstruction_blueprints.jsonl") | Where-Object { $_.Trim() -ne "" } | Select-Object -First 5 | ForEach-Object { $_ | ConvertFrom-Json })
    $recipeDraftRows = @(Get-Content -LiteralPath (Join-Path $fullReportDir "recipe_drafts.jsonl") | Where-Object { $_.Trim() -ne "" } | Select-Object -First 5 | ForEach-Object { $_ | ConvertFrom-Json })
    $readinessSummaryRows = @($projectRows | Group-Object readiness | Sort-Object Name | ForEach-Object {
        [ordered]@{
            readiness = $_.Name
            count = [int]$_.Count
        }
    })
    $blockerSummaryRows = @($allRecreationBlockerRows | Group-Object blocker_type | Sort-Object Name | ForEach-Object {
        [ordered]@{
            blocker_type = $_.Name
            count = [int]$_.Count
        }
    })
    $pluginBlockerCounts = @{}
    foreach ($row in $allRecreationBlockerRows) {
        if ([string]$row.blocker_type -ne "plugin") {
            continue
        }
        $blockerText = ([string]$row.blocker) -replace '^third-party effects:\s*', ''
        foreach ($part in ($blockerText -split ',\s*')) {
            $entry = $part.Trim()
            if ($entry -eq "") {
                continue
            }
            $effectName = $entry
            $effectCount = 1
            if ($entry -match '^(?<name>.+)\s+\((?<count>\d+)\)$') {
                $effectName = $Matches.name.Trim()
                $effectCount = [int]$Matches.count
            }
            if (-not $pluginBlockerCounts.ContainsKey($effectName)) {
                $pluginBlockerCounts[$effectName] = 0
            }
            $pluginBlockerCounts[$effectName] = [int]$pluginBlockerCounts[$effectName] + $effectCount
        }
    }
    $pluginBlockerTopRows = @($pluginBlockerCounts.GetEnumerator() |
        Sort-Object @{ Expression = { $_.Value }; Descending = $true }, @{ Expression = { $_.Key }; Ascending = $true } |
        Select-Object -First 8 |
        ForEach-Object {
            [ordered]@{
                effect = $_.Key
                count = [int]$_.Value
            }
        })
    $nextActionRows = [System.Collections.ArrayList]::new()
    if ($pluginBlockerTopRows.Count -gt 0) {
        $topPlugin = $pluginBlockerTopRows[0]
        [void]$nextActionRows.Add([ordered]@{
            priority = 100
            category = "plugin_blocker"
            title = "Resolve top plugin blocker: $($topPlugin.effect)"
            detail = "Affects $($topPlugin.count) observed effect instance(s) across plugin-blocked projects."
        })
    }
    foreach ($row in @($learningActionRows | Select-Object -First 2)) {
        [void]$nextActionRows.Add([ordered]@{
            priority = [int]$row.priority
            category = "learning_pattern"
            title = $row.pattern
            detail = $row.action
        })
    }
    $analysisReadyProjects = 0
    $pluginBlockedProjects = 0
    foreach ($row in $readinessSummaryRows) {
        if ([string]$row.readiness -eq "analysis_ready") {
            $analysisReadyProjects = [int]$row.count
        } elseif ([string]$row.readiness -eq "needs_plugins") {
            $pluginBlockedProjects = [int]$row.count
        }
    }

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
        recipe_draft_compile = [ordered]@{
            path = $recipeDraftCompileDir
            recipe = Join-Path $recipeDraftCompileDir "recipe_draft.json"
            output = Join-Path $recipeDraftCompileDir "recipe_draft.aep"
            output_bytes = [int64]$compiledRecipeDraftAEP.Length
            validate_json = Join-Path $recipeDraftCompileDir "validate.json"
            compile_json = Join-Path $recipeDraftCompileDir "compile.json"
            valid = [bool]$recipeDraftCompileJson.valid
        }
        recipe_draft_reparse = [ordered]@{
            path = $recipeDraftReparseDir
            facts_json = Join-Path $recipeDraftReparseDir "compiled_facts.json"
            summary_json = Join-Path $recipeDraftReparseDir "reparse_summary.json"
            passed = [bool]$recipeDraftReparseSummary.passed
            expected_comp_count = [int]$recipeDraftReparseSummary.expected_comp_count
            actual_comp_count = [int]$recipeDraftReparseSummary.actual_comp_count
            expected_layer_count = [int]$recipeDraftReparseSummary.expected_layer_count
            actual_layer_count = [int]$recipeDraftReparseSummary.actual_layer_count
        }
        recipe_draft_batch = [ordered]@{
            path = $recipeDraftBatchDir
            summary_json = $recipeDraftBatchSummaryPath
            requested = [int]$recipeDraftBatchSummary.requested
            attempted = [int]$recipeDraftBatchSummary.attempted
            passed = [int]$recipeDraftBatchSummary.passed
        }
        steps = @($steps)
    }
    $acceptance | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $acceptanceJsonPath -Encoding UTF8

    $generatedAtUtc = [DateTime]::UtcNow.ToString("o")
    $historyEntry = [ordered]@{
        schema_version = 1
        generated_at_utc = $generatedAtUtc
        run_id = $runID
        run_root = $runRoot
        input_path = $InputPath
        parsed_projects = [int]$fullManifest.project_count
        parse_errors = [int]$fullManifest.error_count
        technique_patterns = [int]$fullManifest.pattern_count
        partial_error_gate_errors = [int]$partialManifest.error_count
        partial_to_full_count_diffs = [int]$partialCountDiffs
        recipe_draft_compile_valid = [bool]$recipeDraftCompileJson.valid
        compiled_draft_reparse_passed = [bool]$recipeDraftReparseSummary.passed
        batch_requested = [int]$recipeDraftBatchSummary.requested
        batch_attempted = [int]$recipeDraftBatchSummary.attempted
        batch_passed = [int]$recipeDraftBatchSummary.passed
        study_queue_top = [int]$studyRows.Count
        learning_actions_top = [int]$learningActionRows.Count
        coverage_summary = [int]$coverageScorecardRows.Count
    }
    $historyEntryObject = [pscustomobject]$historyEntry
    $existingHistoryRows = @()
    if (Test-Path -LiteralPath $historyCsvPath) {
        $existingHistoryRows = @(Import-Csv -LiteralPath $historyCsvPath)
    }
    $previousHistoryRow = $null
    if ($existingHistoryRows.Count -gt 0) {
        $previousHistoryRow = $existingHistoryRows | Select-Object -Last 1
    }
    $historyDelta = [ordered]@{
        has_previous = ($null -ne $previousHistoryRow)
        previous_run_id = $null
        parsed_projects_delta = 0
        technique_patterns_delta = 0
        batch_attempted_delta = 0
        batch_passed_delta = 0
        partial_to_full_count_diffs_delta = 0
    }
    if ($null -ne $previousHistoryRow) {
        $historyDelta.previous_run_id = [string]$previousHistoryRow.run_id
        $historyDelta.parsed_projects_delta = [int]$historyEntry.parsed_projects - [int]$previousHistoryRow.parsed_projects
        $historyDelta.technique_patterns_delta = [int]$historyEntry.technique_patterns - [int]$previousHistoryRow.technique_patterns
        $historyDelta.batch_attempted_delta = [int]$historyEntry.batch_attempted - [int]$previousHistoryRow.batch_attempted
        $historyDelta.batch_passed_delta = [int]$historyEntry.batch_passed - [int]$previousHistoryRow.batch_passed
        $historyDelta.partial_to_full_count_diffs_delta = [int]$historyEntry.partial_to_full_count_diffs - [int]$previousHistoryRow.partial_to_full_count_diffs
    }
    $outcomeReasons = [System.Collections.ArrayList]::new()
    $outcomeStatus = "pass"
    if ([int]$fullManifest.error_count -ne 0) {
        $outcomeStatus = "fail"
        [void]$outcomeReasons.Add("full report has $($fullManifest.error_count) parse error(s)")
    }
    if (-not [bool]$recipeDraftCompileJson.valid) {
        $outcomeStatus = "fail"
        [void]$outcomeReasons.Add("recipe draft compile validation is not valid")
    }
    if (-not [bool]$recipeDraftReparseSummary.passed) {
        $outcomeStatus = "fail"
        [void]$outcomeReasons.Add("compiled recipe draft reparse did not match expected counts")
    }
    if ([int]$recipeDraftBatchSummary.passed -ne [int]$recipeDraftBatchSummary.attempted) {
        $outcomeStatus = "fail"
        [void]$outcomeReasons.Add("recipe draft batch smoke passed $($recipeDraftBatchSummary.passed)/$($recipeDraftBatchSummary.attempted)")
    }
    if ($outcomeStatus -eq "pass" -and [bool]$historyDelta.has_previous) {
        if ([int]$historyDelta.parsed_projects_delta -lt 0 -or [int]$historyDelta.technique_patterns_delta -lt 0 -or [int]$historyDelta.batch_passed_delta -lt 0) {
            $outcomeStatus = "watch"
            [void]$outcomeReasons.Add("key metric decreased from previous run $($historyDelta.previous_run_id)")
        }
    }
    if ($outcomeReasons.Count -eq 0) {
        [void]$outcomeReasons.Add("self-hosted gate passed with no full-report parse errors and recipe smoke checks passing")
    }
    $outcomeStatusInfo = [ordered]@{
        status = $outcomeStatus
        has_regression = ($outcomeStatus -ne "pass")
        reasons = @($outcomeReasons)
    }
    $outcomeHeadline = "$outcomeStatus · $($fullManifest.project_count) projects · $analysisReadyProjects analysis-ready · $pluginBlockedProjects plugin-blocked · recipe smoke $($recipeDraftBatchSummary.passed)/$($recipeDraftBatchSummary.attempted)"
    $historyPreviewRows = @($existingHistoryRows)
    $historyPreviewRows += $historyEntryObject
    $historyPreviewRows = @($historyPreviewRows | Select-Object -Last 5)

    $effectiveness = [ordered]@{
        schema_version = 1
        generated_at_utc = $generatedAtUtc
        input_path = $InputPath
        run_id = $runID
        run_root = $runRoot
        latest_index = $latestIndexPath
        stable_outputs = [ordered]@{
            open_target = $latestOutcomeHtmlPath
            latest_run = $latestRunPath
            latest_acceptance = $latestAcceptancePath
            latest_outcome = $latestOutcomePath
            latest_outcome_html = $latestOutcomeHtmlPath
            latest_outcome_json = $latestOutcomeJsonPath
            latest_effectiveness = $latestEffectivenessPath
            latest_effectiveness_json = $latestEffectivenessJsonPath
            history_jsonl = $historyJsonlPath
            history_csv = $historyCsvPath
            latest_index = $latestIndexPath
        }
        corpus = [ordered]@{
            parsed_projects = [int]$fullManifest.project_count
            parse_errors = [int]$fullManifest.error_count
            technique_patterns = [int]$fullManifest.pattern_count
            partial_error_gate_errors = [int]$partialManifest.error_count
            partial_to_full_count_diffs = [int]$partialCountDiffs
        }
        closed_loop = [ordered]@{
            recipe_draft_compile_valid = [bool]$recipeDraftCompileJson.valid
            compiled_recipe_draft = Join-Path $recipeDraftCompileDir "recipe_draft.aep"
            compiled_aep_bytes = [int64]$compiledRecipeDraftAEP.Length
            compiled_draft_reparse_passed = [bool]$recipeDraftReparseSummary.passed
            expected_comp_count = [int]$recipeDraftReparseSummary.expected_comp_count
            actual_comp_count = [int]$recipeDraftReparseSummary.actual_comp_count
            expected_layer_count = [int]$recipeDraftReparseSummary.expected_layer_count
            actual_layer_count = [int]$recipeDraftReparseSummary.actual_layer_count
            batch_requested = [int]$recipeDraftBatchSummary.requested
            batch_attempted = [int]$recipeDraftBatchSummary.attempted
            batch_passed = [int]$recipeDraftBatchSummary.passed
        }
        primary_artifacts = [ordered]@{
            full_report_html = Join-Path $fullReportDir "report.html"
            coverage_scorecard = Join-Path $fullReportDir "coverage_scorecard.csv"
            reconstruction_blueprints = Join-Path $fullReportDir "reconstruction_blueprints.jsonl"
            recipe_drafts = Join-Path $fullReportDir "recipe_drafts.jsonl"
            recipe_draft_batch_summary = $recipeDraftBatchSummaryPath
            compiled_draft_facts = Join-Path $recipeDraftReparseDir "compiled_facts.json"
            reparse_summary = Join-Path $recipeDraftReparseDir "reparse_summary.json"
        }
        reconstruction_status = [ordered]@{
            readiness_summary = $readinessSummaryRows
            blocker_summary = $blockerSummaryRows
            plugin_blockers_top = $pluginBlockerTopRows
            blocker_count = [int]$allRecreationBlockerRows.Count
        }
        learning_signals = [ordered]@{
            study_queue_top = @($studyRows | ForEach-Object {
                [ordered]@{
                    rank = [int]$_.rank
                    project_path = $_.path
                    readiness = $_.readiness
                    study_score = [int]$_.study_score
                    patterns = $_.patterns
                    archetypes = $_.archetypes
                    recreation_steps = $_.recreation_steps
                    top_effects = $_.top_effects
                    top_plugin_effects = $_.top_plugin_effects
                    readiness_blockers = $_.readiness_blockers
                }
            })
            learning_actions_top = @($learningActionRows | ForEach-Object {
                [ordered]@{
                    priority = [int]$_.priority
                    pattern = $_.pattern
                    count = [int]$_.count
                    action = $_.action
                    representative_project = $_.representative_project
                    representative_readiness = $_.representative_readiness
                    recreation_steps = $_.recreation_steps
                    top_effects = $_.top_effects
                    top_plugin_effects = $_.top_plugin_effects
                    risk = $_.risk
                }
            })
            coverage_summary = @($coverageScorecardRows | ForEach-Object {
                [ordered]@{
                    artifact = $_.artifact
                    metric = $_.metric
                    expected_count = [int]$_.expected_count
                    actual_count = [int]$_.actual_count
                    status = $_.status
                    notes = $_.notes
                }
            })
        }
        action_plan = [ordered]@{
            next_actions = @($nextActionRows)
        }
        history_delta = $historyDelta
        history_recent = @($historyPreviewRows | ForEach-Object {
            [ordered]@{
                run_id = $_.run_id
                parsed_projects = [int]$_.parsed_projects
                technique_patterns = [int]$_.technique_patterns
                batch_passed = [int]$_.batch_passed
                batch_attempted = [int]$_.batch_attempted
            }
        })
        outcome_status = $outcomeStatusInfo
        outcome_summary = [ordered]@{
            headline = $outcomeHeadline
            analysis_ready_projects = [int]$analysisReadyProjects
            plugin_blocked_projects = [int]$pluginBlockedProjects
            recipe_smoke = "$($recipeDraftBatchSummary.passed)/$($recipeDraftBatchSummary.attempted)"
        }
        steps = @($steps)
    }
    $effectiveness | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $effectivenessJsonPath -Encoding UTF8

    $outcomeJson = [ordered]@{
        schema_version = 1
        generated_at_utc = $generatedAtUtc
        input_path = $InputPath
        run_id = $runID
        run_root = $runRoot
        stable_outputs = [ordered]@{
            open_target = $latestOutcomeHtmlPath
            latest_outcome = $latestOutcomePath
            latest_outcome_html = $latestOutcomeHtmlPath
            latest_outcome_json = $latestOutcomeJsonPath
            latest_effectiveness_json = $latestEffectivenessJsonPath
            latest_index = $latestIndexPath
        }
        outcome_status = $outcomeStatusInfo
        outcome_summary = [ordered]@{
            headline = $outcomeHeadline
            analysis_ready_projects = [int]$analysisReadyProjects
            plugin_blocked_projects = [int]$pluginBlockedProjects
            recipe_smoke = "$($recipeDraftBatchSummary.passed)/$($recipeDraftBatchSummary.attempted)"
        }
        action_plan = [ordered]@{
            next_actions = @($nextActionRows)
        }
        reconstruction_status = [ordered]@{
            readiness_summary = $readinessSummaryRows
            plugin_blockers_top = $pluginBlockerTopRows
            blocker_count = [int]$allRecreationBlockerRows.Count
        }
        corpus = [ordered]@{
            parsed_projects = [int]$fullManifest.project_count
            parse_errors = [int]$fullManifest.error_count
            technique_patterns = [int]$fullManifest.pattern_count
        }
        closed_loop = [ordered]@{
            batch_passed = [int]$recipeDraftBatchSummary.passed
            batch_attempted = [int]$recipeDraftBatchSummary.attempted
        }
        history_delta = $historyDelta
        history_recent = @($historyPreviewRows | ForEach-Object {
            [ordered]@{
                run_id = $_.run_id
                parsed_projects = [int]$_.parsed_projects
                technique_patterns = [int]$_.technique_patterns
                batch_passed = [int]$_.batch_passed
                batch_attempted = [int]$_.batch_attempted
            }
        })
    }
    $outcomeJson | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $outcomeJsonPath -Encoding UTF8

    function Escape-Html {
        param([AllowNull()][object]$Value)
        if ($null -eq $Value) {
            return ""
        }
        return [System.Net.WebUtility]::HtmlEncode([string]$Value)
    }

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
    [void]$b.AppendLine("- recipe draft compile: ``$recipeDraftCompileDir``")
    [void]$b.AppendLine("- compiled recipe draft bytes: $($compiledRecipeDraftAEP.Length)")
    [void]$b.AppendLine("- compiled recipe draft reparse: ``$recipeDraftReparseDir``")
    [void]$b.AppendLine("- reparse comp/layer counts: $($recipeDraftReparseSummary.actual_comp_count)/$($recipeDraftReparseSummary.expected_comp_count) comps, $($recipeDraftReparseSummary.actual_layer_count)/$($recipeDraftReparseSummary.expected_layer_count) layers")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("## Steps")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("| step | exit | seconds |")
    [void]$b.AppendLine("| --- | ---: | ---: |")
    foreach ($step in $steps) {
    [void]$b.AppendLine("| $($step.name) | $($step.exit) | $($step.seconds) |")
    }
    $b.ToString() | Set-Content -LiteralPath $acceptanceMdPath -Encoding UTF8

    $o = [System.Text.StringBuilder]::new()
    [void]$o.AppendLine("# Self-Hosted Outcome")
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Outcome Status")
    [void]$o.AppendLine("")
    [void]$o.AppendLine("- status: $($outcomeStatusInfo.status)")
    [void]$o.AppendLine("- regression: $($outcomeStatusInfo.has_regression)")
    foreach ($reason in $outcomeStatusInfo.reasons) {
        [void]$o.AppendLine("- reason: $reason")
    }
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Effectiveness Headline")
    [void]$o.AppendLine("")
    [void]$o.AppendLine($outcomeHeadline)
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Next Actions")
    [void]$o.AppendLine("")
    foreach ($row in @($nextActionRows | Select-Object -First 3)) {
        [void]$o.AppendLine("- P$($row.priority): $($row.title) - $($row.detail)")
    }
    if ($nextActionRows.Count -eq 0) {
        [void]$o.AppendLine("- No immediate actions reported.")
    }
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Reconstruction Readiness")
    [void]$o.AppendLine("")
    foreach ($row in $readinessSummaryRows) {
        [void]$o.AppendLine("- $($row.readiness): $($row.count)")
    }
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Recreation Blockers")
    [void]$o.AppendLine("")
    foreach ($row in @($recreationBlockerRows | Select-Object -First 3)) {
        [void]$o.AppendLine("- $($row.project_path): $($row.blocker_type) - $($row.blocker)")
    }
    if ($recreationBlockerRows.Count -eq 0) {
        [void]$o.AppendLine("- No blockers reported.")
    }
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Top Plugin Blockers")
    [void]$o.AppendLine("")
    foreach ($row in @($pluginBlockerTopRows | Select-Object -First 5)) {
        [void]$o.AppendLine("- $($row.effect): $($row.count)")
    }
    if ($pluginBlockerTopRows.Count -eq 0) {
        [void]$o.AppendLine("- No plugin blockers reported.")
    }
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Numbers")
    [void]$o.AppendLine("")
    [void]$o.AppendLine("- projects: $($fullManifest.project_count)")
    [void]$o.AppendLine("- patterns: $($fullManifest.pattern_count)")
    [void]$o.AppendLine("- full parse errors: $($fullManifest.error_count)")
    [void]$o.AppendLine("- recipe batch smoke: $($recipeDraftBatchSummary.passed)/$($recipeDraftBatchSummary.attempted)")
    [void]$o.AppendLine("- history delta: projects $($historyDelta.parsed_projects_delta), patterns $($historyDelta.technique_patterns_delta), batch passed $($historyDelta.batch_passed_delta)")
    [void]$o.AppendLine("")
    [void]$o.AppendLine("## Links")
    [void]$o.AppendLine("")
    [void]$o.AppendLine("- latest index: ``$latestIndexPath``")
    [void]$o.AppendLine("- effectiveness json: ``$latestEffectivenessJsonPath``")
    [void]$o.AppendLine("- run root: ``$runRoot``")
    $o.ToString() | Set-Content -LiteralPath $outcomeMdPath -Encoding UTF8

    $statusClass = "ok"
    if ($outcomeStatusInfo.status -eq "watch") {
        $statusClass = "watch"
    } elseif ($outcomeStatusInfo.status -ne "pass") {
        $statusClass = "fail"
    }
    $outcomeHtml = [System.Text.StringBuilder]::new()
    [void]$outcomeHtml.AppendLine("<!doctype html>")
    [void]$outcomeHtml.AppendLine("<html lang=""en""><head><meta charset=""utf-8""><meta name=""viewport"" content=""width=device-width, initial-scale=1"">")
    [void]$outcomeHtml.AppendLine("<title>Self-Hosted Outcome</title>")
    [void]$outcomeHtml.AppendLine("<style>body{font-family:Segoe UI,Arial,sans-serif;margin:0;background:#f6f8fb;color:#1f2937}main{max-width:760px;margin:0 auto;padding:32px}.hero,.panel{background:#fff;border:1px solid #d8dee8;border-radius:8px;padding:18px;margin-bottom:16px}h1{font-size:28px;margin:0 0 8px}.status{display:inline-block;border-radius:6px;padding:6px 10px;font-weight:700;text-transform:uppercase}.ok{background:#dcfce7;color:#166534}.watch{background:#fef3c7;color:#92400e}.fail{background:#fee2e2;color:#991b1b}.muted{color:#667085}table{width:100%;border-collapse:collapse;font-size:14px}td,th{border-bottom:1px solid #e5e7eb;padding:8px;text-align:left}td:last-child,th:last-child{text-align:right}.links{display:grid;gap:10px}.links a{display:block;border:1px solid #d8dee8;border-radius:8px;padding:10px 12px;color:#1d4ed8;text-decoration:none;background:#fff}</style>")
    [void]$outcomeHtml.AppendLine("</head><body><main>")
    [void]$outcomeHtml.AppendLine("<section class=""hero""><h1>Self-Hosted Outcome</h1><p class=""muted"">run <code>$runID</code> · input <code>$InputPath</code></p><h2>Outcome Status</h2><p><span class=""status $statusClass"">$(Escape-Html $outcomeStatusInfo.status)</span></p><p>$(Escape-Html ((@($outcomeStatusInfo.reasons)) -join "; "))</p></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>Effectiveness Headline</h2><p>$(Escape-Html $outcomeHeadline)</p></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>Numbers</h2><table><tbody>")
    [void]$outcomeHtml.AppendLine("<tr><td>Projects</td><td>$($fullManifest.project_count)</td></tr>")
    [void]$outcomeHtml.AppendLine("<tr><td>Patterns</td><td>$($fullManifest.pattern_count)</td></tr>")
    [void]$outcomeHtml.AppendLine("<tr><td>Full parse errors</td><td>$($fullManifest.error_count)</td></tr>")
    [void]$outcomeHtml.AppendLine("<tr><td>Recipe batch smoke</td><td>$($recipeDraftBatchSummary.passed) / $($recipeDraftBatchSummary.attempted)</td></tr>")
    [void]$outcomeHtml.AppendLine("</tbody></table></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>Next Actions</h2><table><thead><tr><th>Priority</th><th>Action</th><th>Detail</th></tr></thead><tbody>")
    foreach ($row in @($nextActionRows | Select-Object -First 3)) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.priority)</td><td>$(Escape-Html $row.title)</td><td>$(Escape-Html $row.detail)</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>Reconstruction Readiness</h2><table><thead><tr><th>Readiness</th><th>Projects</th></tr></thead><tbody>")
    foreach ($row in $readinessSummaryRows) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.readiness)</td><td>$(Escape-Html $row.count)</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table>")
    [void]$outcomeHtml.AppendLine("<h3>Recreation Blockers</h3><table><thead><tr><th>Project</th><th>Type</th><th>Blocker</th></tr></thead><tbody>")
    foreach ($row in @($recreationBlockerRows | Select-Object -First 3)) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.blocker_type)</td><td>$(Escape-Html $row.blocker)</td></tr>")
    }
    if ($recreationBlockerRows.Count -eq 0) {
        [void]$outcomeHtml.AppendLine("<tr><td colspan=""3"">No blockers reported.</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table>")
    [void]$outcomeHtml.AppendLine("<h3>Top Plugin Blockers</h3><table><thead><tr><th>Effect</th><th>Count</th></tr></thead><tbody>")
    foreach ($row in @($pluginBlockerTopRows | Select-Object -First 5)) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.effect)</td><td>$(Escape-Html $row.count)</td></tr>")
    }
    if ($pluginBlockerTopRows.Count -eq 0) {
        [void]$outcomeHtml.AppendLine("<tr><td colspan=""2"">No plugin blockers reported.</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>History Delta</h2><table><tbody>")
    [void]$outcomeHtml.AppendLine("<tr><td>Previous run</td><td>$(Escape-Html $historyDelta.previous_run_id)</td></tr>")
    [void]$outcomeHtml.AppendLine("<tr><td>Projects</td><td>$(Escape-Html $historyDelta.parsed_projects_delta)</td></tr>")
    [void]$outcomeHtml.AppendLine("<tr><td>Patterns</td><td>$(Escape-Html $historyDelta.technique_patterns_delta)</td></tr>")
    [void]$outcomeHtml.AppendLine("<tr><td>Batch passed</td><td>$(Escape-Html $historyDelta.batch_passed_delta)</td></tr>")
    [void]$outcomeHtml.AppendLine("</tbody></table></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>Recent Runs</h2><table><thead><tr><th>Run</th><th>Projects</th><th>Patterns</th><th>Batch</th></tr></thead><tbody>")
    foreach ($row in $historyPreviewRows) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.run_id)</td><td>$(Escape-Html $row.parsed_projects)</td><td>$(Escape-Html $row.technique_patterns)</td><td>$(Escape-Html $row.batch_passed) / $(Escape-Html $row.batch_attempted)</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>Learning Signals</h2>")
    [void]$outcomeHtml.AppendLine("<h3>Study Queue</h3><table><thead><tr><th>Rank</th><th>Project</th><th>Patterns</th></tr></thead><tbody>")
    foreach ($row in @($studyRows | Select-Object -First 3)) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.rank)</td><td>$(Escape-Html $row.path)</td><td>$(Escape-Html $row.patterns)</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table>")
    [void]$outcomeHtml.AppendLine("<h3>Learning Actions</h3><table><thead><tr><th>Pattern</th><th>Action</th><th>Risk</th></tr></thead><tbody>")
    foreach ($row in @($learningActionRows | Select-Object -First 3)) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.pattern)</td><td>$(Escape-Html $row.action)</td><td>$(Escape-Html $row.risk)</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table>")
    [void]$outcomeHtml.AppendLine("<h3>Coverage Summary</h3><table><thead><tr><th>Artifact</th><th>Count</th><th>Status</th></tr></thead><tbody>")
    foreach ($row in @($coverageScorecardRows | Select-Object -First 5)) {
        [void]$outcomeHtml.AppendLine("<tr><td>$(Escape-Html $row.artifact)</td><td>$(Escape-Html $row.actual_count)</td><td>$(Escape-Html $row.status)</td></tr>")
    }
    [void]$outcomeHtml.AppendLine("</tbody></table></section>")
    [void]$outcomeHtml.AppendLine("<section class=""panel""><h2>Links</h2><div class=""links"">")
    [void]$outcomeHtml.AppendLine("<a href=""latest_index.html"">Latest full index</a>")
    [void]$outcomeHtml.AppendLine("<a href=""latest_outcome.md"">Latest outcome markdown</a>")
    [void]$outcomeHtml.AppendLine("<a href=""latest_effectiveness.json"">Latest effectiveness JSON</a>")
    [void]$outcomeHtml.AppendLine("<a href=""$runID/outcome.md"">Run outcome markdown</a>")
    [void]$outcomeHtml.AppendLine("</div></section>")
    [void]$outcomeHtml.AppendLine("</main></body></html>")
    $outcomeHtml.ToString() | Set-Content -LiteralPath $outcomeHtmlPath -Encoding UTF8

    $e = [System.Text.StringBuilder]::new()
    [void]$e.AppendLine("# Technique Effectiveness Snapshot")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("- input: ``$InputPath``")
    [void]$e.AppendLine("- run root: ``$runRoot``")
    [void]$e.AppendLine("- latest index: ``$latestIndexPath``")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("## Outcome Status")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("- status: $($outcomeStatusInfo.status)")
    [void]$e.AppendLine("- regression: $($outcomeStatusInfo.has_regression)")
    foreach ($reason in $outcomeStatusInfo.reasons) {
        [void]$e.AppendLine("- reason: $reason")
    }
    [void]$e.AppendLine("")
    [void]$e.AppendLine("## Corpus")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("- parsed projects: $($fullManifest.project_count)")
    [void]$e.AppendLine("- parse errors: $($fullManifest.error_count)")
    [void]$e.AppendLine("- technique patterns: $($fullManifest.pattern_count)")
    [void]$e.AppendLine("- partial-error gate errors: $($partialManifest.error_count)")
    [void]$e.AppendLine("- partial-to-full count diffs: $partialCountDiffs")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("## Closed Loop")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("- recipe draft compile valid: $($recipeDraftCompileJson.valid)")
    [void]$e.AppendLine("- compiled recipe draft: ``$(Join-Path $recipeDraftCompileDir "recipe_draft.aep")``")
    [void]$e.AppendLine("- compiled AEP bytes: $($compiledRecipeDraftAEP.Length)")
    [void]$e.AppendLine("- compiled draft reparse passed: $($recipeDraftReparseSummary.passed)")
    [void]$e.AppendLine("- reparse comp count: $($recipeDraftReparseSummary.actual_comp_count)/$($recipeDraftReparseSummary.expected_comp_count)")
    [void]$e.AppendLine("- reparse layer count: $($recipeDraftReparseSummary.actual_layer_count)/$($recipeDraftReparseSummary.expected_layer_count)")
    [void]$e.AppendLine("- recipe draft batch smoke: $($recipeDraftBatchSummary.passed)/$($recipeDraftBatchSummary.attempted) passed")
    [void]$e.AppendLine("- history delta: projects $($historyDelta.parsed_projects_delta), patterns $($historyDelta.technique_patterns_delta), batch passed $($historyDelta.batch_passed_delta)")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("## Primary Artifacts")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("- full report HTML: ``$(Join-Path $fullReportDir "report.html")``")
    [void]$e.AppendLine("- coverage scorecard: ``$(Join-Path $fullReportDir "coverage_scorecard.csv")``")
    [void]$e.AppendLine("- reconstruction blueprints: ``$(Join-Path $fullReportDir "reconstruction_blueprints.jsonl")``")
    [void]$e.AppendLine("- recipe drafts: ``$(Join-Path $fullReportDir "recipe_drafts.jsonl")``")
    [void]$e.AppendLine("- recipe draft batch summary: ``$recipeDraftBatchSummaryPath``")
    [void]$e.AppendLine("- compiled draft facts: ``$(Join-Path $recipeDraftReparseDir "compiled_facts.json")``")
    [void]$e.AppendLine("- reparse summary: ``$(Join-Path $recipeDraftReparseDir "reparse_summary.json")``")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("## Gate Steps")
    [void]$e.AppendLine("")
    [void]$e.AppendLine("| step | exit | seconds |")
    [void]$e.AppendLine("| --- | ---: | ---: |")
    foreach ($step in $steps) {
        [void]$e.AppendLine("| $($step.name) | $($step.exit) | $($step.seconds) |")
    }
    $e.ToString() | Set-Content -LiteralPath $effectivenessMdPath -Encoding UTF8

    function Require-LatestIndexLink {
        param(
            [string]$Label,
            [string]$RelativePath
        )
        $target = Join-Path $OutRoot $RelativePath
        if (-not (Test-Path -LiteralPath $target)) {
            throw "latest index link target missing for ${Label}: $target"
        }
    }

    $runRel = $runID
    $index = [System.Text.StringBuilder]::new()
    [void]$index.AppendLine("<!doctype html>")
    [void]$index.AppendLine("<html lang=""en""><head><meta charset=""utf-8""><meta name=""viewport"" content=""width=device-width, initial-scale=1"">")
    [void]$index.AppendLine("<title>Technique Self-Hosted Acceptance</title>")
    [void]$index.AppendLine("<style>body{font-family:Segoe UI,Arial,sans-serif;margin:0;background:#f6f8fb;color:#1f2937}main{max-width:960px;margin:0 auto;padding:32px}h1{font-size:28px;margin:0 0 8px}.muted{color:#667085}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px;margin:22px 0}.metric,.panel{background:#fff;border:1px solid #d8dee8;border-radius:8px;padding:16px}.metric strong{display:block;font-size:26px;margin-top:6px}.links{display:grid;gap:10px;margin-top:16px}.links a{display:block;background:#fff;border:1px solid #d8dee8;border-radius:8px;padding:12px 14px;color:#1d4ed8;text-decoration:none}.links a:hover{text-decoration:underline}table{width:100%;border-collapse:collapse;font-size:14px}td,th{border-bottom:1px solid #e5e7eb;padding:8px;text-align:left}td:last-child,th:last-child{text-align:right}code{background:#eef2f7;padding:2px 5px;border-radius:4px}</style>")
    [void]$index.AppendLine("</head><body><main>")
    [void]$index.AppendLine("<h1>Technique Self-Hosted Acceptance</h1>")
    [void]$index.AppendLine("<p class=""muted"">run <code>$runRel</code> · input <code>$InputPath</code></p>")
    [void]$index.AppendLine("<div class=""grid"">")
    [void]$index.AppendLine("<div class=""metric""><span>Projects</span><strong>$($fullManifest.project_count)</strong></div>")
    [void]$index.AppendLine("<div class=""metric""><span>Errors</span><strong>$($fullManifest.error_count)</strong></div>")
    [void]$index.AppendLine("<div class=""metric""><span>Patterns</span><strong>$($fullManifest.pattern_count)</strong></div>")
    [void]$index.AppendLine("<div class=""metric""><span>Partial Errors</span><strong>$($partialManifest.error_count)</strong></div>")
    [void]$index.AppendLine("</div>")
    [void]$index.AppendLine("<section class=""panel""><h2>Outcome Status</h2><table><thead><tr><th>Signal</th><th>Value</th></tr></thead><tbody>")
    [void]$index.AppendLine("<tr><td>Status</td><td>$(Escape-Html $outcomeStatusInfo.status)</td></tr>")
    [void]$index.AppendLine("<tr><td>Regression</td><td>$(Escape-Html $outcomeStatusInfo.has_regression)</td></tr>")
    [void]$index.AppendLine("<tr><td>Reason</td><td>$(Escape-Html ((@($outcomeStatusInfo.reasons)) -join "; "))</td></tr>")
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel""><h2>Effectiveness Snapshot</h2><table><thead><tr><th>Signal</th><th>Value</th></tr></thead><tbody>")
    [void]$index.AppendLine("<tr><td>Parsed projects</td><td>$($fullManifest.project_count)</td></tr>")
    [void]$index.AppendLine("<tr><td>Recipe draft compile</td><td>$(Escape-Html $recipeDraftCompileJson.valid)</td></tr>")
    [void]$index.AppendLine("<tr><td>Compiled AEP bytes</td><td>$($compiledRecipeDraftAEP.Length)</td></tr>")
    [void]$index.AppendLine("<tr><td>Compiled draft reparse</td><td>$(Escape-Html $recipeDraftReparseSummary.passed)</td></tr>")
    [void]$index.AppendLine("<tr><td>Reparse comps</td><td>$($recipeDraftReparseSummary.actual_comp_count) / $($recipeDraftReparseSummary.expected_comp_count)</td></tr>")
    [void]$index.AppendLine("<tr><td>Reparse layers</td><td>$($recipeDraftReparseSummary.actual_layer_count) / $($recipeDraftReparseSummary.expected_layer_count)</td></tr>")
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>History Delta</h2><table><thead><tr><th>Metric</th><th>Delta</th></tr></thead><tbody>")
    [void]$index.AppendLine("<tr><td>Previous run</td><td>$(Escape-Html $historyDelta.previous_run_id)</td></tr>")
    [void]$index.AppendLine("<tr><td>Parsed projects</td><td>$(Escape-Html $historyDelta.parsed_projects_delta)</td></tr>")
    [void]$index.AppendLine("<tr><td>Technique patterns</td><td>$(Escape-Html $historyDelta.technique_patterns_delta)</td></tr>")
    [void]$index.AppendLine("<tr><td>Batch attempted</td><td>$(Escape-Html $historyDelta.batch_attempted_delta)</td></tr>")
    [void]$index.AppendLine("<tr><td>Batch passed</td><td>$(Escape-Html $historyDelta.batch_passed_delta)</td></tr>")
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Effectiveness History Preview</h2><table><thead><tr><th>Run</th><th>Projects</th><th>Patterns</th><th>Batch</th></tr></thead><tbody>")
    foreach ($row in $historyPreviewRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.run_id)</td><td>$(Escape-Html $row.parsed_projects)</td><td>$(Escape-Html $row.technique_patterns)</td><td>$(Escape-Html $row.batch_passed) / $(Escape-Html $row.batch_attempted)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel""><h2>Study Queue Preview</h2><table><thead><tr><th>Rank</th><th>Project</th><th>Score</th></tr></thead><tbody>")
    foreach ($row in $studyRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.rank)</td><td>$(Escape-Html $row.path)</td><td>$(Escape-Html $row.study_score)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Project Playbooks Preview</h2><table><thead><tr><th>Project</th><th>Readiness</th><th>Steps</th></tr></thead><tbody>")
    foreach ($row in $projectPlaybookRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.readiness)</td><td>$(Escape-Html $row.ordered_steps)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Compositions Preview</h2><table><thead><tr><th>Project</th><th>Name</th><th>Layers</th></tr></thead><tbody>")
    foreach ($row in $compositionRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.name)</td><td>$(Escape-Html $row.layer_count)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Layers Preview</h2><table><thead><tr><th>Project</th><th>Comp</th><th>Role</th></tr></thead><tbody>")
    foreach ($row in $layerRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.comp_name)</td><td>$(Escape-Html $row.role)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Recreation Steps Preview</h2><table><thead><tr><th>Project</th><th>Priority</th><th>Step</th></tr></thead><tbody>")
    foreach ($row in $recreationStepRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.priority)</td><td>$(Escape-Html $row.title)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Pattern Playbook Preview</h2><table><thead><tr><th>Pattern</th><th>Count</th><th>Steps</th></tr></thead><tbody>")
    foreach ($row in $patternRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.id)</td><td>$(Escape-Html $row.count)</td><td>$(Escape-Html $row.recreation_steps)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Learning Actions Preview</h2><table><thead><tr><th>Pattern</th><th>Action</th><th>Risk</th></tr></thead><tbody>")
    foreach ($row in $learningActionRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.pattern)</td><td>$(Escape-Html $row.action)</td><td>$(Escape-Html $row.risk)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Recreation Blockers Preview</h2><table><thead><tr><th>Project</th><th>Type</th><th>Blocker</th></tr></thead><tbody>")
    foreach ($row in $recreationBlockerRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.blocker_type)</td><td>$(Escape-Html $row.blocker)</td></tr>")
    }
    if ($recreationBlockerRows.Count -eq 0) {
        [void]$index.AppendLine("<tr><td colspan=""3"">No blockers reported.</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Signal Layers Preview</h2><table><thead><tr><th>Project</th><th>Layer</th><th>Score</th></tr></thead><tbody>")
    foreach ($row in $signalLayerRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.layer_name)</td><td>$(Escape-Html $row.score)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Effect Stacks Preview</h2><table><thead><tr><th>Project</th><th>Layer</th><th>Effect</th></tr></thead><tbody>")
    foreach ($row in $effectStackRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.layer_name)</td><td>$(Escape-Html $row.match_name)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Shape Operators Preview</h2><table><thead><tr><th>Project</th><th>Layer</th><th>Family</th></tr></thead><tbody>")
    foreach ($row in $shapeOperatorRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.layer_name)</td><td>$(Escape-Html $row.family)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Text Animators Preview</h2><table><thead><tr><th>Project</th><th>Layer</th><th>Kind</th></tr></thead><tbody>")
    foreach ($row in $textAnimatorRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.layer_name)</td><td>$(Escape-Html $row.property_kind)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Dependency Edges Preview</h2><table><thead><tr><th>Project</th><th>Relation</th><th>Target</th></tr></thead><tbody>")
    foreach ($row in $dependencyEdgeRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.relation)</td><td>$(Escape-Html $row.target_name)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Mechanism Catalog Preview</h2><table><thead><tr><th>Category</th><th>Name</th><th>Count</th></tr></thead><tbody>")
    foreach ($row in $mechanismRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.category)</td><td>$(Escape-Html $row.name)</td><td>$(Escape-Html $row.count)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Coverage Scorecard Preview</h2><table><thead><tr><th>Artifact</th><th>Expected</th><th>Actual</th><th>Status</th></tr></thead><tbody>")
    foreach ($row in $coverageScorecardRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.artifact)</td><td>$(Escape-Html $row.expected_count)</td><td>$(Escape-Html $row.actual_count)</td><td>$(Escape-Html $row.status)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Reconstruction Blueprints Preview</h2><table><thead><tr><th>Project</th><th>Readiness</th><th>Phases</th></tr></thead><tbody>")
    foreach ($row in $reconstructionBlueprintRows) {
        $phaseText = ((@($row.phases) | ForEach-Object { "$($_.id)=$($_.expected_count)" }) -join "; ")
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(Escape-Html $row.readiness)</td><td>$(Escape-Html $phaseText)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Recipe Drafts Preview</h2><table><thead><tr><th>Project</th><th>Comps</th><th>Layers</th><th>Gaps</th></tr></thead><tbody>")
    foreach ($row in $recipeDraftRows) {
        $gapText = ((@($row.gaps) | ForEach-Object { "$($_.id)=$($_.count)" }) -join "; ")
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.project_path)</td><td>$(@($row.recipe.comps).Count)</td><td>$(Escape-Html $row.counts.layers)</td><td>$(Escape-Html $gapText)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Recipe Draft Compile Smoke</h2><table><thead><tr><th>Artifact</th><th>Value</th></tr></thead><tbody>")
    [void]$index.AppendLine("<tr><td>compile valid</td><td>$(Escape-Html $recipeDraftCompileJson.valid)</td></tr>")
    [void]$index.AppendLine("<tr><td>compiled bytes</td><td>$($compiledRecipeDraftAEP.Length)</td></tr>")
    [void]$index.AppendLine("<tr><td>output</td><td>$(Escape-Html $compiledRecipeDraftAEP.FullName)</td></tr>")
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Recipe Draft Batch Smoke</h2><table><thead><tr><th>Check</th><th>Value</th></tr></thead><tbody>")
    [void]$index.AppendLine("<tr><td>requested</td><td>$($recipeDraftBatchSummary.requested)</td></tr>")
    [void]$index.AppendLine("<tr><td>attempted</td><td>$($recipeDraftBatchSummary.attempted)</td></tr>")
    [void]$index.AppendLine("<tr><td>passed</td><td>$($recipeDraftBatchSummary.passed)</td></tr>")
    [void]$index.AppendLine("<tr><td>summary</td><td>$(Escape-Html $recipeDraftBatchSummaryPath)</td></tr>")
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Compiled Draft Reparse Smoke</h2><table><thead><tr><th>Check</th><th>Expected</th><th>Actual</th><th>Status</th></tr></thead><tbody>")
    [void]$index.AppendLine("<tr><td>compositions</td><td>$($recipeDraftReparseSummary.expected_comp_count)</td><td>$($recipeDraftReparseSummary.actual_comp_count)</td><td>$(Escape-Html $recipeDraftReparseSummary.passed)</td></tr>")
    [void]$index.AppendLine("<tr><td>layers</td><td>$($recipeDraftReparseSummary.expected_layer_count)</td><td>$($recipeDraftReparseSummary.actual_layer_count)</td><td>$(Escape-Html $recipeDraftReparseSummary.passed)</td></tr>")
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel""><h2>Artifacts</h2><div class=""links"">")
    [void]$index.AppendLine("<a href=""$runRel/full_report/report.html"">Full report HTML</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/learning.md"">Learning index</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/project_playbooks.csv"">Project playbooks CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/compositions.csv"">Compositions CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/layers.csv"">Layers CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/recreation_steps.csv"">Recreation steps CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/study_queue.csv"">Study queue CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/study_tasks.csv"">Study tasks CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/recreation_blockers.csv"">Recreation blockers CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/signal_layers.csv"">Signal layers CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/effect_stacks.csv"">Effect stacks CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/shape_operators.csv"">Shape operators CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/text_animators.csv"">Text animators CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/dependency_edges.csv"">Dependency edges CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/learning_actions.csv"">Learning actions CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/mechanisms.csv"">Mechanisms CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/mechanism_examples.csv"">Mechanism examples CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/coverage_scorecard.csv"">Coverage scorecard CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/reconstruction_blueprints.jsonl"">Reconstruction blueprints JSONL</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/recipe_drafts.jsonl"">Recipe drafts JSONL</a>")
    [void]$index.AppendLine("<a href=""$runRel/recipe_draft_compile/recipe_draft.json"">Compiled recipe draft JSON</a>")
    [void]$index.AppendLine("<a href=""$runRel/recipe_draft_compile/recipe_draft.aep"">Compiled recipe draft AEP</a>")
    [void]$index.AppendLine("<a href=""$runRel/recipe_draft_compile/validate.json"">Recipe draft validate JSON</a>")
    [void]$index.AppendLine("<a href=""$runRel/recipe_draft_compile/compile.json"">Recipe draft compile JSON</a>")
    [void]$index.AppendLine("<a href=""$runRel/recipe_draft_batch/summary.json"">Recipe draft batch summary</a>")
    [void]$index.AppendLine("<a href=""$runRel/recipe_draft_reparse/compiled_facts.json"">Compiled draft facts JSON</a>")
    [void]$index.AppendLine("<a href=""$runRel/recipe_draft_reparse/reparse_summary.json"">Compiled draft reparse summary</a>")
    [void]$index.AppendLine("<a href=""$runRel/effectiveness.json"">Effectiveness JSON</a>")
    [void]$index.AppendLine("<a href=""latest_effectiveness.md"">Latest effectiveness snapshot</a>")
    [void]$index.AppendLine("<a href=""latest_outcome.md"">Latest outcome</a>")
    [void]$index.AppendLine("<a href=""latest_outcome.html"">Latest outcome HTML</a>")
    [void]$index.AppendLine("<a href=""latest_outcome.json"">Latest outcome JSON</a>")
    [void]$index.AppendLine("<a href=""latest_effectiveness.json"">Latest effectiveness JSON</a>")
    [void]$index.AppendLine("<a href=""history.jsonl"">Effectiveness history JSONL</a>")
    [void]$index.AppendLine("<a href=""history.csv"">Effectiveness history CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/projects.csv"">Projects CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/patterns.csv"">Patterns CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/errors.csv"">Errors CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/partial_report/report.html"">Partial-error report HTML</a>")
    [void]$index.AppendLine("<a href=""$runRel/compare_self/compare.md"">Self compare</a>")
    [void]$index.AppendLine("<a href=""$runRel/compare_partial_to_full/compare.md"">Partial-to-full compare</a>")
    [void]$index.AppendLine("<a href=""$runRel/acceptance.md"">Acceptance markdown</a>")
    [void]$index.AppendLine("<a href=""$runRel/outcome.md"">Outcome markdown</a>")
    [void]$index.AppendLine("<a href=""$runRel/outcome.html"">Outcome HTML</a>")
    [void]$index.AppendLine("<a href=""$runRel/outcome.json"">Outcome JSON</a>")
    [void]$index.AppendLine("<a href=""$runRel/acceptance.json"">Acceptance JSON</a>")
    [void]$index.AppendLine("</div></section>")
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Steps</h2><table><thead><tr><th>Step</th><th>Seconds</th></tr></thead><tbody>")
    foreach ($step in $steps) {
        [void]$index.AppendLine("<tr><td>$($step.name)</td><td>$($step.seconds)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("</main></body></html>")
    $index.ToString() | Set-Content -LiteralPath $latestIndexPath -Encoding UTF8
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Study Queue Preview" -Quiet)) {
        throw "latest index missing Study Queue Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Effectiveness Snapshot" -Quiet)) {
        throw "latest index missing Effectiveness Snapshot"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Effectiveness History Preview" -Quiet)) {
        throw "latest index missing Effectiveness History Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "History Delta" -Quiet)) {
        throw "latest index missing History Delta"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Outcome Status" -Quiet)) {
        throw "latest index missing Outcome Status"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Project Playbooks Preview" -Quiet)) {
        throw "latest index missing Project Playbooks Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Compositions Preview" -Quiet)) {
        throw "latest index missing Compositions Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Layers Preview" -Quiet)) {
        throw "latest index missing Layers Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Recreation Steps Preview" -Quiet)) {
        throw "latest index missing Recreation Steps Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Pattern Playbook Preview" -Quiet)) {
        throw "latest index missing Pattern Playbook Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Learning Actions Preview" -Quiet)) {
        throw "latest index missing Learning Actions Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Recreation Blockers Preview" -Quiet)) {
        throw "latest index missing Recreation Blockers Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Signal Layers Preview" -Quiet)) {
        throw "latest index missing Signal Layers Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Effect Stacks Preview" -Quiet)) {
        throw "latest index missing Effect Stacks Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Shape Operators Preview" -Quiet)) {
        throw "latest index missing Shape Operators Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Text Animators Preview" -Quiet)) {
        throw "latest index missing Text Animators Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Dependency Edges Preview" -Quiet)) {
        throw "latest index missing Dependency Edges Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Mechanism Catalog Preview" -Quiet)) {
        throw "latest index missing Mechanism Catalog Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Coverage Scorecard Preview" -Quiet)) {
        throw "latest index missing Coverage Scorecard Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Reconstruction Blueprints Preview" -Quiet)) {
        throw "latest index missing Reconstruction Blueprints Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Recipe Drafts Preview" -Quiet)) {
        throw "latest index missing Recipe Drafts Preview"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Recipe Draft Compile Smoke" -Quiet)) {
        throw "latest index missing Recipe Draft Compile Smoke"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Recipe Draft Batch Smoke" -Quiet)) {
        throw "latest index missing Recipe Draft Batch Smoke"
    }
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Compiled Draft Reparse Smoke" -Quiet)) {
        throw "latest index missing Compiled Draft Reparse Smoke"
    }
    Require-LatestIndexLink -Label "full report" -RelativePath "$runRel/full_report/report.html"
    Require-LatestIndexLink -Label "learning index" -RelativePath "$runRel/full_report/learning.md"
    Require-LatestIndexLink -Label "project playbooks" -RelativePath "$runRel/full_report/project_playbooks.csv"
    Require-LatestIndexLink -Label "compositions" -RelativePath "$runRel/full_report/compositions.csv"
    Require-LatestIndexLink -Label "layers" -RelativePath "$runRel/full_report/layers.csv"
    Require-LatestIndexLink -Label "recreation steps" -RelativePath "$runRel/full_report/recreation_steps.csv"
    Require-LatestIndexLink -Label "study queue" -RelativePath "$runRel/full_report/study_queue.csv"
    Require-LatestIndexLink -Label "study tasks" -RelativePath "$runRel/full_report/study_tasks.csv"
    Require-LatestIndexLink -Label "recreation blockers" -RelativePath "$runRel/full_report/recreation_blockers.csv"
    Require-LatestIndexLink -Label "signal layers" -RelativePath "$runRel/full_report/signal_layers.csv"
    Require-LatestIndexLink -Label "effect stacks" -RelativePath "$runRel/full_report/effect_stacks.csv"
    Require-LatestIndexLink -Label "shape operators" -RelativePath "$runRel/full_report/shape_operators.csv"
    Require-LatestIndexLink -Label "text animators" -RelativePath "$runRel/full_report/text_animators.csv"
    Require-LatestIndexLink -Label "dependency edges" -RelativePath "$runRel/full_report/dependency_edges.csv"
    Require-LatestIndexLink -Label "learning actions" -RelativePath "$runRel/full_report/learning_actions.csv"
    Require-LatestIndexLink -Label "mechanisms" -RelativePath "$runRel/full_report/mechanisms.csv"
    Require-LatestIndexLink -Label "mechanism examples" -RelativePath "$runRel/full_report/mechanism_examples.csv"
    Require-LatestIndexLink -Label "coverage scorecard" -RelativePath "$runRel/full_report/coverage_scorecard.csv"
    Require-LatestIndexLink -Label "reconstruction blueprints" -RelativePath "$runRel/full_report/reconstruction_blueprints.jsonl"
    Require-LatestIndexLink -Label "recipe drafts" -RelativePath "$runRel/full_report/recipe_drafts.jsonl"
    Require-LatestIndexLink -Label "compiled recipe draft json" -RelativePath "$runRel/recipe_draft_compile/recipe_draft.json"
    Require-LatestIndexLink -Label "compiled recipe draft aep" -RelativePath "$runRel/recipe_draft_compile/recipe_draft.aep"
    Require-LatestIndexLink -Label "recipe draft validate json" -RelativePath "$runRel/recipe_draft_compile/validate.json"
    Require-LatestIndexLink -Label "recipe draft compile json" -RelativePath "$runRel/recipe_draft_compile/compile.json"
    Require-LatestIndexLink -Label "recipe draft batch summary" -RelativePath "$runRel/recipe_draft_batch/summary.json"
    Require-LatestIndexLink -Label "compiled draft facts json" -RelativePath "$runRel/recipe_draft_reparse/compiled_facts.json"
    Require-LatestIndexLink -Label "compiled draft reparse summary" -RelativePath "$runRel/recipe_draft_reparse/reparse_summary.json"
    Require-LatestIndexLink -Label "effectiveness json" -RelativePath "$runRel/effectiveness.json"
    Require-LatestIndexLink -Label "partial report" -RelativePath "$runRel/partial_report/report.html"
    Require-LatestIndexLink -Label "partial compare" -RelativePath "$runRel/compare_partial_to_full/compare.md"
    Require-LatestIndexLink -Label "outcome markdown" -RelativePath "$runRel/outcome.md"
    Require-LatestIndexLink -Label "outcome html" -RelativePath "$runRel/outcome.html"
    Require-LatestIndexLink -Label "outcome json" -RelativePath "$runRel/outcome.json"

    $runRoot | Set-Content -LiteralPath $latestRunPath -Encoding UTF8
    Copy-Item -LiteralPath $acceptanceMdPath -Destination $latestAcceptancePath -Force
    Copy-Item -LiteralPath $outcomeMdPath -Destination $latestOutcomePath -Force
    Copy-Item -LiteralPath $outcomeHtmlPath -Destination $latestOutcomeHtmlPath -Force
    Copy-Item -LiteralPath $outcomeJsonPath -Destination $latestOutcomeJsonPath -Force
    Copy-Item -LiteralPath $effectivenessMdPath -Destination $latestEffectivenessPath -Force
    Copy-Item -LiteralPath $effectivenessJsonPath -Destination $latestEffectivenessJsonPath -Force
    if (-not (Test-Path -LiteralPath $latestEffectivenessPath)) {
        throw "latest effectiveness snapshot missing: $latestEffectivenessPath"
    }
    Require-LatestIndexLink -Label "latest effectiveness" -RelativePath "latest_effectiveness.md"
    if (-not (Test-Path -LiteralPath $latestOutcomePath)) {
        throw "latest outcome missing: $latestOutcomePath"
    }
    if (-not (Select-String -LiteralPath $latestOutcomePath -Pattern "Outcome Status" -Quiet)) {
        throw "latest outcome missing Outcome Status"
    }
    if (-not (Select-String -LiteralPath $latestOutcomePath -Pattern "Effectiveness Headline" -Quiet)) {
        throw "latest outcome missing Effectiveness Headline"
    }
    if (-not (Select-String -LiteralPath $latestOutcomePath -Pattern "Next Actions" -Quiet)) {
        throw "latest outcome missing Next Actions"
    }
    if (-not (Select-String -LiteralPath $latestOutcomePath -Pattern "Reconstruction Readiness" -Quiet)) {
        throw "latest outcome missing Reconstruction Readiness"
    }
    if (-not (Select-String -LiteralPath $latestOutcomePath -Pattern "Top Plugin Blockers" -Quiet)) {
        throw "latest outcome missing Top Plugin Blockers"
    }
    Require-LatestIndexLink -Label "latest outcome" -RelativePath "latest_outcome.md"
    if (-not (Test-Path -LiteralPath $latestOutcomeJsonPath)) {
        throw "latest outcome json missing: $latestOutcomeJsonPath"
    }
    Require-LatestIndexLink -Label "latest outcome json" -RelativePath "latest_outcome.json"
    $latestOutcomeJson = Get-Content -Raw -LiteralPath $latestOutcomeJsonPath | ConvertFrom-Json
    if ($null -eq $latestOutcomeJson.outcome_summary) {
        throw "latest outcome json missing outcome_summary"
    }
    if ($null -eq $latestOutcomeJson.outcome_summary.headline) {
        throw "latest outcome json missing outcome_summary.headline"
    }
    if ($null -eq $latestOutcomeJson.action_plan) {
        throw "latest outcome json missing action_plan"
    }
    if ($null -eq $latestOutcomeJson.action_plan.next_actions) {
        throw "latest outcome json missing action_plan.next_actions"
    }
    if ($null -eq $latestOutcomeJson.stable_outputs.open_target) {
        throw "latest outcome json missing stable_outputs.open_target"
    }
    $outcomeCliOutput = & pwsh -NoProfile -File scripts\show_technique_outcome.ps1 -OutRoot $OutRoot 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "show technique outcome failed: $outcomeCliOutput"
    }
    if (($outcomeCliOutput -join "`n") -notmatch [regex]::Escape([string]$latestOutcomeJson.outcome_summary.headline)) {
        throw "show technique outcome missing headline"
    }
    if (($outcomeCliOutput -join "`n") -notmatch "Next Actions") {
        throw "show technique outcome missing Next Actions"
    }
    $watchDryRunRoot = Join-Path $OutRoot "watch_dry_run"
    $watchDryRunOutput = & pwsh -NoProfile -File scripts\watch_technique_selfhost.ps1 -OutRoot $watchDryRunRoot -Iterations 1 -IntervalSeconds 1 -DryRun 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "watch technique selfhost dry run failed: $watchDryRunOutput"
    }
    $watchDryRunText = $watchDryRunOutput -join "`n"
    if ($watchDryRunText -notmatch "DRY RUN") {
        throw "watch technique selfhost dry run missing DRY RUN marker"
    }
    if ($watchDryRunText -notmatch "aepselfhost verify") {
        throw "watch technique selfhost dry run missing go verify command"
    }
    if ($watchDryRunText -notmatch "aepselfhost outcome") {
        throw "watch technique selfhost dry run missing go outcome command"
    }
    $watchStatusPath = Join-Path $watchDryRunRoot "watch_status.json"
    if (-not (Test-Path -LiteralPath $watchStatusPath)) {
        throw "watch technique selfhost dry run missing watch_status.json"
    }
    $watchStatus = Get-Content -Raw -LiteralPath $watchStatusPath | ConvertFrom-Json
    if ([string]$watchStatus.mode -ne "dry_run") {
        throw "watch technique selfhost dry run status mode mismatch"
    }
    $startDryRunRoot = Join-Path $OutRoot "start_dry_run"
    $startDryRunOutput = & pwsh -NoProfile -File scripts\start_technique_selfhost_watch.ps1 -OutRoot $startDryRunRoot -DurationMinutes 60 -IntervalSeconds 300 -DryRun 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "start technique selfhost watch dry run failed: $startDryRunOutput"
    }
    $startDryRunText = $startDryRunOutput -join "`n"
    if ($startDryRunText -notmatch "DRY RUN") {
        throw "start technique selfhost watch dry run missing DRY RUN marker"
    }
    if ($startDryRunText -notmatch "aepselfhost watch") {
        throw "start technique selfhost watch dry run missing go watch marker"
    }
    $startProcessPath = Join-Path $startDryRunRoot "watch_process.json"
    if (-not (Test-Path -LiteralPath $startProcessPath)) {
        throw "start technique selfhost watch dry run missing watch_process.json"
    }
    $startProcess = Get-Content -Raw -LiteralPath $startProcessPath | ConvertFrom-Json
    if ([string]$startProcess.mode -ne "dry_run") {
        throw "start technique selfhost watch dry run process mode mismatch"
    }
    if ($null -eq $startProcess.stdout_log -or $null -eq $startProcess.stderr_log) {
        throw "start technique selfhost watch dry run missing log paths"
    }
    $statusCliOutput = & pwsh -NoProfile -File scripts\show_technique_selfhost_status.ps1 -OutRoot $OutRoot 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "show technique selfhost status failed: $statusCliOutput"
    }
    $statusCliText = $statusCliOutput -join "`n"
    if ($statusCliText -notmatch "Self-Hosted Status") {
        throw "show technique selfhost status missing title"
    }
    if ($statusCliText -notmatch "Watch Process") {
        throw "show technique selfhost status missing Watch Process"
    }
    if ($statusCliText -notmatch "Latest Outcome") {
        throw "show technique selfhost status missing Latest Outcome"
    }
    if ($statusCliText -notmatch "Logs") {
        throw "show technique selfhost status missing Logs"
    }
    $goOutcomeOutput = & go run ./cmd/aepselfhost outcome -out-root $OutRoot 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "go aepselfhost outcome failed: $goOutcomeOutput"
    }
    $goOutcomeText = $goOutcomeOutput -join "`n"
    if ($goOutcomeText -notmatch [regex]::Escape([string]$latestOutcomeJson.outcome_summary.headline)) {
        throw "go aepselfhost outcome missing headline"
    }
    if ($goOutcomeText -notmatch "Next Actions") {
        throw "go aepselfhost outcome missing Next Actions"
    }
    $goStatusOutput = & go run ./cmd/aepselfhost status -out-root $OutRoot 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "go aepselfhost status failed: $goStatusOutput"
    }
    $goStatusText = $goStatusOutput -join "`n"
    if ($goStatusText -notmatch "Self-Hosted Status") {
        throw "go aepselfhost status missing title"
    }
    if ($goStatusText -notmatch "Latest Outcome") {
        throw "go aepselfhost status missing Latest Outcome"
    }
    if (-not (Test-Path -LiteralPath $latestOutcomeHtmlPath)) {
        throw "latest outcome html missing: $latestOutcomeHtmlPath"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Self-Hosted Outcome" -Quiet)) {
        throw "latest outcome html missing Self-Hosted Outcome"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Outcome Status" -Quiet)) {
        throw "latest outcome html missing Outcome Status"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Effectiveness Headline" -Quiet)) {
        throw "latest outcome html missing Effectiveness Headline"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Learning Signals" -Quiet)) {
        throw "latest outcome html missing Learning Signals"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Study Queue" -Quiet)) {
        throw "latest outcome html missing Study Queue"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Learning Actions" -Quiet)) {
        throw "latest outcome html missing Learning Actions"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Next Actions" -Quiet)) {
        throw "latest outcome html missing Next Actions"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Reconstruction Readiness" -Quiet)) {
        throw "latest outcome html missing Reconstruction Readiness"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Recreation Blockers" -Quiet)) {
        throw "latest outcome html missing Recreation Blockers"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Top Plugin Blockers" -Quiet)) {
        throw "latest outcome html missing Top Plugin Blockers"
    }
    if (-not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern "Recent Runs" -Quiet)) {
        throw "latest outcome html missing Recent Runs"
    }
    $topStudyPath = ""
    if ($studyRows.Count -gt 0) {
        $topStudyPath = [string]$studyRows[0].path
    }
    if ($topStudyPath -ne "" -and -not (Select-String -LiteralPath $latestOutcomeHtmlPath -Pattern ([regex]::Escape($topStudyPath)) -Quiet)) {
        throw "latest outcome html missing top study path: $topStudyPath"
    }
    Require-LatestIndexLink -Label "latest outcome html" -RelativePath "latest_outcome.html"
    if (-not (Test-Path -LiteralPath $latestEffectivenessJsonPath)) {
        throw "latest effectiveness json missing: $latestEffectivenessJsonPath"
    }
    Require-LatestIndexLink -Label "latest effectiveness json" -RelativePath "latest_effectiveness.json"
    $latestEffectivenessJson = Get-Content -Raw -LiteralPath $latestEffectivenessJsonPath | ConvertFrom-Json
    if ($null -eq $latestEffectivenessJson.learning_signals) {
        throw "latest effectiveness json missing learning_signals"
    }
    if ($null -eq $latestEffectivenessJson.learning_signals.study_queue_top) {
        throw "latest effectiveness json missing learning_signals.study_queue_top"
    }
    if ($null -eq $latestEffectivenessJson.learning_signals.learning_actions_top) {
        throw "latest effectiveness json missing learning_signals.learning_actions_top"
    }
    if ($null -eq $latestEffectivenessJson.learning_signals.coverage_summary) {
        throw "latest effectiveness json missing learning_signals.coverage_summary"
    }
    if ($null -eq $latestEffectivenessJson.reconstruction_status) {
        throw "latest effectiveness json missing reconstruction_status"
    }
    if ($null -eq $latestEffectivenessJson.reconstruction_status.readiness_summary) {
        throw "latest effectiveness json missing reconstruction_status.readiness_summary"
    }
    if ($null -eq $latestEffectivenessJson.reconstruction_status.blocker_summary) {
        throw "latest effectiveness json missing reconstruction_status.blocker_summary"
    }
    if ($null -eq $latestEffectivenessJson.reconstruction_status.plugin_blockers_top) {
        throw "latest effectiveness json missing reconstruction_status.plugin_blockers_top"
    }
    if ($null -eq $latestEffectivenessJson.action_plan) {
        throw "latest effectiveness json missing action_plan"
    }
    if ($null -eq $latestEffectivenessJson.action_plan.next_actions) {
        throw "latest effectiveness json missing action_plan.next_actions"
    }
    if ($null -eq $latestEffectivenessJson.history_delta) {
        throw "latest effectiveness json missing history_delta"
    }
    if ($null -eq $latestEffectivenessJson.history_delta.parsed_projects_delta) {
        throw "latest effectiveness json missing history_delta.parsed_projects_delta"
    }
    if ($null -eq $latestEffectivenessJson.history_recent) {
        throw "latest effectiveness json missing history_recent"
    }
    if ($null -eq $latestEffectivenessJson.stable_outputs.open_target) {
        throw "latest effectiveness json missing stable_outputs.open_target"
    }
    if ([string]$latestEffectivenessJson.stable_outputs.open_target -ne $latestOutcomeHtmlPath) {
        throw "latest effectiveness json open target = $($latestEffectivenessJson.stable_outputs.open_target), want $latestOutcomeHtmlPath"
    }
    if ($null -eq $latestEffectivenessJson.outcome_status) {
        throw "latest effectiveness json missing outcome_status"
    }
    if ($null -eq $latestEffectivenessJson.outcome_summary) {
        throw "latest effectiveness json missing outcome_summary"
    }
    if ($null -eq $latestEffectivenessJson.outcome_summary.headline) {
        throw "latest effectiveness json missing outcome_summary.headline"
    }
    if ($null -eq $latestEffectivenessJson.outcome_status.status) {
        throw "latest effectiveness json missing outcome_status.status"
    }
    if ($null -eq $latestEffectivenessJson.outcome_status.reasons) {
        throw "latest effectiveness json missing outcome_status.reasons"
    }
    if (-not (Select-String -LiteralPath $latestEffectivenessPath -Pattern "Outcome Status" -Quiet)) {
        throw "latest effectiveness markdown missing Outcome Status"
    }
    $historyEntryJson = $historyEntry | ConvertTo-Json -Depth 6 -Compress
    Add-Content -LiteralPath $historyJsonlPath -Value $historyEntryJson -Encoding UTF8
    if (Test-Path -LiteralPath $historyCsvPath) {
        $historyEntryObject | Export-Csv -LiteralPath $historyCsvPath -NoTypeInformation -Append -Encoding UTF8
    } else {
        $historyEntryObject | Export-Csv -LiteralPath $historyCsvPath -NoTypeInformation -Encoding UTF8
    }
    $latestHistoryLine = Get-Content -LiteralPath $historyJsonlPath | Select-Object -Last 1
    if ($latestHistoryLine -notmatch '"generated_at_utc":"\d{4}-\d{2}-\d{2}T') {
        throw "effectiveness history generated_at_utc is not ISO-8601: $latestHistoryLine"
    }
    if (-not (Test-Path -LiteralPath $historyJsonlPath)) {
        throw "effectiveness history jsonl missing: $historyJsonlPath"
    }
    if (-not (Test-Path -LiteralPath $historyCsvPath)) {
        throw "effectiveness history csv missing: $historyCsvPath"
    }
    Require-LatestIndexLink -Label "effectiveness history jsonl" -RelativePath "history.jsonl"
    Require-LatestIndexLink -Label "effectiveness history csv" -RelativePath "history.csv"

    Write-Host "acceptance json: $acceptanceJsonPath"
    Write-Host "acceptance md:   $acceptanceMdPath"
    Write-Host "effectiveness:   $effectivenessMdPath"
    Write-Host "effect json:     $effectivenessJsonPath"
    Write-Host "latest run:      $latestRunPath"
    Write-Host "latest summary:  $latestAcceptancePath"
    Write-Host "latest outcome:  $latestOutcomePath"
    Write-Host "latest outcome h: $latestOutcomeHtmlPath"
    Write-Host "latest outcome j: $latestOutcomeJsonPath"
    Write-Host "open target:     $latestOutcomeHtmlPath"
    Write-Host "latest effect:   $latestEffectivenessPath"
    Write-Host "latest effect j: $latestEffectivenessJsonPath"
    Write-Host "latest index:    $latestIndexPath"
    Write-Host "full report:     $fullReportDir"
    Write-Host "partial report:  $partialReportDir"
    if ($Open) {
        Invoke-Item $latestOutcomeHtmlPath
    }
}
finally {
    Pop-Location
}
