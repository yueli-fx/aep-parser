param(
    [string]$OutDir = "tmp\technique_showcase_report",
    [int]$MinProjects = 1
)

$ErrorActionPreference = "Stop"

function Require-File {
    param([string]$Path)
    if (-not (Test-Path -LiteralPath $Path)) {
        throw "missing report artifact: $Path"
    }
}

function Require-Text {
    param(
        [string]$Path,
        [string]$Pattern
    )
    if (-not (Select-String -LiteralPath $Path -Pattern $Pattern -Quiet)) {
        throw "missing pattern '$Pattern' in $Path"
    }
}

$summaryPath = Join-Path $OutDir "summary.json"
$corpusPath = Join-Path $OutDir "corpus.jsonl"
$digestPath = Join-Path $OutDir "digest.json"
$learningPath = Join-Path $OutDir "learning.md"
$projectsCsvPath = Join-Path $OutDir "projects.csv"
$projectPlaybooksCsvPath = Join-Path $OutDir "project_playbooks.csv"
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
$errorsCsvPath = Join-Path $OutDir "errors.csv"
$manifestPath = Join-Path $OutDir "manifest.json"
$reportPath = Join-Path $OutDir "report.md"
$htmlPath = Join-Path $OutDir "report.html"

foreach ($path in @($summaryPath, $corpusPath, $digestPath, $learningPath, $projectsCsvPath, $projectPlaybooksCsvPath, $patternsCsvPath, $studyQueueCsvPath, $studyTasksCsvPath, $recreationBlockersCsvPath, $signalLayersCsvPath, $effectStacksCsvPath, $shapeOperatorsCsvPath, $textAnimatorsCsvPath, $dependencyEdgesCsvPath, $learningActionsCsvPath, $mechanismsCsvPath, $mechanismExamplesCsvPath, $errorsCsvPath, $manifestPath, $reportPath, $htmlPath)) {
    Require-File -Path $path
}

$summary = Get-Content -Raw -LiteralPath $summaryPath | ConvertFrom-Json
$digest = Get-Content -Raw -LiteralPath $digestPath | ConvertFrom-Json
$manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
$corpusLines = @(Get-Content -LiteralPath $corpusPath | Where-Object { $_.Trim() -ne "" })
$corpusRecords = @($corpusLines | ForEach-Object { $_ | ConvertFrom-Json })
$projectRows = @(Import-Csv -LiteralPath $projectsCsvPath)
$projectPlaybookRows = @(Import-Csv -LiteralPath $projectPlaybooksCsvPath)
$patternRows = @(Import-Csv -LiteralPath $patternsCsvPath)
$studyQueueRows = @(Import-Csv -LiteralPath $studyQueueCsvPath)
$studyTaskRows = @(Import-Csv -LiteralPath $studyTasksCsvPath)
$recreationBlockerRows = @(Import-Csv -LiteralPath $recreationBlockersCsvPath)
$signalLayerRows = @(Import-Csv -LiteralPath $signalLayersCsvPath)
$effectStackRows = @(Import-Csv -LiteralPath $effectStacksCsvPath)
$shapeOperatorRows = @(Import-Csv -LiteralPath $shapeOperatorsCsvPath)
$textAnimatorRows = @(Import-Csv -LiteralPath $textAnimatorsCsvPath)
$dependencyEdgeRows = @(Import-Csv -LiteralPath $dependencyEdgesCsvPath)
$learningActionRows = @(Import-Csv -LiteralPath $learningActionsCsvPath)
$mechanismRows = @(Import-Csv -LiteralPath $mechanismsCsvPath)
$mechanismExampleRows = @(Import-Csv -LiteralPath $mechanismExamplesCsvPath)
$errorRows = @(Import-Csv -LiteralPath $errorsCsvPath)

if ([int]$summary.project_count -lt $MinProjects) {
    throw "project_count $($summary.project_count) is lower than MinProjects $MinProjects"
}
$expectedCorpusLines = [int]$summary.project_count
if ($null -ne $summary.error_count) {
    $expectedCorpusLines += [int]$summary.error_count
}
if ($corpusLines.Count -ne $expectedCorpusLines) {
    throw "corpus line count $($corpusLines.Count) does not match project_count + error_count $expectedCorpusLines"
}
if ([int]$digest.project_count -ne [int]$summary.project_count) {
    throw "digest project_count $($digest.project_count) does not match summary project_count $($summary.project_count)"
}
if ([int]$manifest.project_count -ne [int]$summary.project_count) {
    throw "manifest project_count $($manifest.project_count) does not match summary project_count $($summary.project_count)"
}
if ([int]$manifest.error_count -ne [int]$summary.error_count) {
    throw "manifest error_count $($manifest.error_count) does not match summary error_count $($summary.error_count)"
}
if ($null -eq $manifest.artifacts -or @($manifest.artifacts).Count -lt 8) {
    throw "manifest has no artifact inventory"
}
if ($null -eq $manifest.git -or [string]$manifest.git.commit -eq "") {
    throw "manifest has no git commit"
}
if ($null -eq $digest.patterns -or @($digest.patterns).Count -eq 0) {
    throw "digest has no pattern entries"
}
if ($projectRows.Count -ne [int]$summary.project_count) {
    throw "projects.csv row count $($projectRows.Count) does not match summary project_count $($summary.project_count)"
}
if ($projectPlaybookRows.Count -ne [int]$summary.project_count) {
    throw "project_playbooks.csv row count $($projectPlaybookRows.Count) does not match summary project_count $($summary.project_count)"
}
foreach ($row in $projectPlaybookRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.readiness -eq "" -or [string]$row.overview -eq "" -or [string]$row.ordered_steps -eq "") {
        throw "project_playbooks.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($studyQueueRows.Count -ne [int]$summary.project_count) {
    throw "study_queue.csv row count $($studyQueueRows.Count) does not match summary project_count $($summary.project_count)"
}
if ($studyTaskRows.Count -eq 0) {
    throw "study_tasks.csv has no rows"
}
foreach ($row in $studyTaskRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.focus -eq "" -or [string]$row.action -eq "" -or [int]$row.rank -le 0) {
        throw "study_tasks.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
foreach ($row in $recreationBlockerRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.blocker_type -eq "" -or [string]$row.blocker -eq "" -or [string]$row.action -eq "") {
        throw "recreation_blockers.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
$projectsWithBlockers = @($projectRows | Where-Object { [string]$_.readiness_blockers -ne "" })
if ($projectsWithBlockers.Count -gt 0 -and $recreationBlockerRows.Count -eq 0) {
    throw "recreation_blockers.csv has no rows but projects.csv reports readiness blockers"
}
if ($signalLayerRows.Count -eq 0) {
    throw "signal_layers.csv has no rows"
}
foreach ($row in $signalLayerRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.layer_name -eq "" -or [string]$row.role -eq "" -or [int]$row.score -le 0 -or [string]$row.signals -eq "") {
        throw "signal_layers.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($effectStackRows.Count -ne [int]$summary.totals.effect_count) {
    throw "effect_stacks.csv row count $($effectStackRows.Count) does not match summary totals.effect_count $($summary.totals.effect_count)"
}
foreach ($row in $effectStackRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.comp_name -eq "" -or [string]$row.match_name -eq "") {
        throw "effect_stacks.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($shapeOperatorRows.Count -ne [int]$summary.totals.shape_operator_count) {
    throw "shape_operators.csv row count $($shapeOperatorRows.Count) does not match summary totals.shape_operator_count $($summary.totals.shape_operator_count)"
}
foreach ($row in $shapeOperatorRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.comp_name -eq "" -or [string]$row.layer_name -eq "" -or [string]$row.family -eq "" -or [string]$row.source -eq "") {
        throw "shape_operators.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($textAnimatorRows.Count -ne [int]$summary.totals.text_animator_count) {
    throw "text_animators.csv row count $($textAnimatorRows.Count) does not match summary totals.text_animator_count $($summary.totals.text_animator_count)"
}
foreach ($row in $textAnimatorRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.comp_name -eq "" -or [string]$row.layer_name -eq "" -or [string]$row.property_kind -eq "" -or [string]$row.match_name -eq "") {
        throw "text_animators.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($dependencyEdgeRows.Count -ne [int]$summary.totals.dependency_count) {
    throw "dependency_edges.csv row count $($dependencyEdgeRows.Count) does not match summary totals.dependency_count $($summary.totals.dependency_count)"
}
foreach ($row in $dependencyEdgeRows) {
    if ([string]$row.project_path -eq "" -or [string]$row.comp_name -eq "" -or [string]$row.relation -eq "") {
        throw "dependency_edges.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($errorRows.Count -ne [int]$summary.error_count) {
    throw "errors.csv row count $($errorRows.Count) does not match summary error_count $($summary.error_count)"
}
if ($patternRows.Count -ne @($digest.patterns).Count) {
    throw "patterns.csv row count $($patternRows.Count) does not match digest pattern count $(@($digest.patterns).Count)"
}
if ($learningActionRows.Count -ne @($digest.patterns).Count) {
    throw "learning_actions.csv row count $($learningActionRows.Count) does not match digest pattern count $(@($digest.patterns).Count)"
}
foreach ($row in $learningActionRows) {
    if ([string]$row.pattern -eq "" -or [string]$row.action -eq "" -or [string]$row.representative_project -eq "") {
        throw "learning_actions.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($mechanismRows.Count -eq 0) {
    throw "mechanisms.csv has no rows"
}
foreach ($row in $mechanismRows) {
    if ([string]$row.category -eq "" -or [string]$row.name -eq "" -or [int]$row.count -le 0 -or [string]$row.action -eq "") {
        throw "mechanisms.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
if ($mechanismExampleRows.Count -eq 0) {
    throw "mechanism_examples.csv has no rows"
}
foreach ($row in $mechanismExampleRows) {
    if ([string]$row.category -eq "" -or [string]$row.name -eq "" -or [string]$row.project_path -eq "" -or [int]$row.project_count -le 0 -or [string]$row.action -eq "") {
        throw "mechanism_examples.csv contains incomplete row: $($row | ConvertTo-Json -Compress)"
    }
}
$recordsWithSteps = @($corpusRecords | Where-Object {
    $null -ne $_.explanation -and
    $null -ne $_.explanation.recreation_steps -and
    @($_.explanation.recreation_steps).Count -gt 0
})
if ($summary.project_count -gt 0 -and $recordsWithSteps.Count -eq 0) {
    throw "corpus has no explanation.recreation_steps entries"
}
$patternsWithSteps = @(@($digest.patterns) | Where-Object {
    $null -ne $_.recreation_steps -and
    @($_.recreation_steps).Count -gt 0
})
if (@($digest.patterns).Count -gt 0 -and $patternsWithSteps.Count -eq 0) {
    throw "digest patterns have no recreation_steps entries"
}

Require-Text -Path $learningPath -Pattern "^## Pattern Playbook$"
Require-Text -Path $learningPath -Pattern "^## Study Queue$"
Require-Text -Path $learningPath -Pattern "recreation steps"
Require-Text -Path $learningPath -Pattern "^## Plugin Risk Queue$"
Require-Text -Path $learningPath -Pattern "^## Readiness Queue$"
Require-Text -Path $reportPath -Pattern "^## Pattern Representatives$"
Require-Text -Path $reportPath -Pattern "^## Study Queue$"
Require-Text -Path $reportPath -Pattern "Step [0-9]+:"
Require-Text -Path $htmlPath -Pattern "Technique Corpus Report"
Require-Text -Path $htmlPath -Pattern "Step [0-9]+:"
Require-Text -Path $htmlPath -Pattern "manifest\.json"
Require-Text -Path $htmlPath -Pattern "learning\.md"
Require-Text -Path $htmlPath -Pattern "projects\.csv"
Require-Text -Path $htmlPath -Pattern "project_playbooks\.csv"
Require-Text -Path $htmlPath -Pattern "study_queue\.csv"
Require-Text -Path $htmlPath -Pattern "study_tasks\.csv"
Require-Text -Path $htmlPath -Pattern "recreation_blockers\.csv"
Require-Text -Path $htmlPath -Pattern "signal_layers\.csv"
Require-Text -Path $htmlPath -Pattern "effect_stacks\.csv"
Require-Text -Path $htmlPath -Pattern "shape_operators\.csv"
Require-Text -Path $htmlPath -Pattern "text_animators\.csv"
Require-Text -Path $htmlPath -Pattern "dependency_edges\.csv"
Require-Text -Path $htmlPath -Pattern "learning_actions\.csv"
Require-Text -Path $htmlPath -Pattern "mechanisms\.csv"
Require-Text -Path $htmlPath -Pattern "mechanism_examples\.csv"
Require-Text -Path $htmlPath -Pattern "errors\.csv"
Require-Text -Path $htmlPath -Pattern "Mechanism Explorer"
Require-Text -Path $htmlPath -Pattern "mechanismFilter"
Require-Text -Path $htmlPath -Pattern "Study Task Queue"
Require-Text -Path $htmlPath -Pattern "studyTaskFilter"
Require-Text -Path $htmlPath -Pattern "Project Playbooks"
Require-Text -Path $htmlPath -Pattern "Recreation Blockers"
Require-Text -Path $htmlPath -Pattern "Effect Stacks"
Require-Text -Path $htmlPath -Pattern "Shape Operators"
Require-Text -Path $htmlPath -Pattern "Text Animators"
Require-Text -Path $htmlPath -Pattern "Dependency Edges"
Require-Text -Path $htmlPath -Pattern "Representative Projects"
Require-Text -Path $htmlPath -Pattern "mechanism-representatives"

Write-Host "ok: $OutDir"
Write-Host "projects: $($summary.project_count)"
Write-Host "patterns: $(@($digest.patterns).Count)"
