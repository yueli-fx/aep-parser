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
$patternsCsvPath = Join-Path $OutDir "patterns.csv"
$studyQueueCsvPath = Join-Path $OutDir "study_queue.csv"
$learningActionsCsvPath = Join-Path $OutDir "learning_actions.csv"
$mechanismsCsvPath = Join-Path $OutDir "mechanisms.csv"
$errorsCsvPath = Join-Path $OutDir "errors.csv"
$manifestPath = Join-Path $OutDir "manifest.json"
$reportPath = Join-Path $OutDir "report.md"
$htmlPath = Join-Path $OutDir "report.html"

foreach ($path in @($summaryPath, $corpusPath, $digestPath, $learningPath, $projectsCsvPath, $patternsCsvPath, $studyQueueCsvPath, $learningActionsCsvPath, $mechanismsCsvPath, $errorsCsvPath, $manifestPath, $reportPath, $htmlPath)) {
    Require-File -Path $path
}

$summary = Get-Content -Raw -LiteralPath $summaryPath | ConvertFrom-Json
$digest = Get-Content -Raw -LiteralPath $digestPath | ConvertFrom-Json
$manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
$corpusLines = @(Get-Content -LiteralPath $corpusPath | Where-Object { $_.Trim() -ne "" })
$corpusRecords = @($corpusLines | ForEach-Object { $_ | ConvertFrom-Json })
$projectRows = @(Import-Csv -LiteralPath $projectsCsvPath)
$patternRows = @(Import-Csv -LiteralPath $patternsCsvPath)
$studyQueueRows = @(Import-Csv -LiteralPath $studyQueueCsvPath)
$learningActionRows = @(Import-Csv -LiteralPath $learningActionsCsvPath)
$mechanismRows = @(Import-Csv -LiteralPath $mechanismsCsvPath)
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
if ($studyQueueRows.Count -ne [int]$summary.project_count) {
    throw "study_queue.csv row count $($studyQueueRows.Count) does not match summary project_count $($summary.project_count)"
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
Require-Text -Path $htmlPath -Pattern "study_queue\.csv"
Require-Text -Path $htmlPath -Pattern "learning_actions\.csv"
Require-Text -Path $htmlPath -Pattern "mechanisms\.csv"
Require-Text -Path $htmlPath -Pattern "errors\.csv"
Require-Text -Path $htmlPath -Pattern "Mechanism Explorer"
Require-Text -Path $htmlPath -Pattern "mechanismFilter"

Write-Host "ok: $OutDir"
Write-Host "projects: $($summary.project_count)"
Write-Host "patterns: $(@($digest.patterns).Count)"
