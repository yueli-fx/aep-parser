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
$reportPath = Join-Path $OutDir "report.md"
$htmlPath = Join-Path $OutDir "report.html"

foreach ($path in @($summaryPath, $corpusPath, $digestPath, $learningPath, $projectsCsvPath, $patternsCsvPath, $studyQueueCsvPath, $reportPath, $htmlPath)) {
    Require-File -Path $path
}

$summary = Get-Content -Raw -LiteralPath $summaryPath | ConvertFrom-Json
$digest = Get-Content -Raw -LiteralPath $digestPath | ConvertFrom-Json
$corpusLines = @(Get-Content -LiteralPath $corpusPath | Where-Object { $_.Trim() -ne "" })
$corpusRecords = @($corpusLines | ForEach-Object { $_ | ConvertFrom-Json })
$projectRows = @(Import-Csv -LiteralPath $projectsCsvPath)
$patternRows = @(Import-Csv -LiteralPath $patternsCsvPath)
$studyQueueRows = @(Import-Csv -LiteralPath $studyQueueCsvPath)

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
if ($null -eq $digest.patterns -or @($digest.patterns).Count -eq 0) {
    throw "digest has no pattern entries"
}
if ($projectRows.Count -ne [int]$summary.project_count) {
    throw "projects.csv row count $($projectRows.Count) does not match summary project_count $($summary.project_count)"
}
if ($studyQueueRows.Count -ne [int]$summary.project_count) {
    throw "study_queue.csv row count $($studyQueueRows.Count) does not match summary project_count $($summary.project_count)"
}
if ($patternRows.Count -ne @($digest.patterns).Count) {
    throw "patterns.csv row count $($patternRows.Count) does not match digest pattern count $(@($digest.patterns).Count)"
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
Require-Text -Path $htmlPath -Pattern "learning\.md"
Require-Text -Path $htmlPath -Pattern "projects\.csv"
Require-Text -Path $htmlPath -Pattern "study_queue\.csv"

Write-Host "ok: $OutDir"
Write-Host "projects: $($summary.project_count)"
Write-Host "patterns: $(@($digest.patterns).Count)"
