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
$reportPath = Join-Path $OutDir "report.md"
$htmlPath = Join-Path $OutDir "report.html"

foreach ($path in @($summaryPath, $corpusPath, $digestPath, $learningPath, $reportPath, $htmlPath)) {
    Require-File -Path $path
}

$summary = Get-Content -Raw -LiteralPath $summaryPath | ConvertFrom-Json
$digest = Get-Content -Raw -LiteralPath $digestPath | ConvertFrom-Json
$corpusLines = @(Get-Content -LiteralPath $corpusPath | Where-Object { $_.Trim() -ne "" })

if ([int]$summary.project_count -lt $MinProjects) {
    throw "project_count $($summary.project_count) is lower than MinProjects $MinProjects"
}
if ($corpusLines.Count -ne [int]$summary.project_count) {
    throw "corpus line count $($corpusLines.Count) does not match summary project_count $($summary.project_count)"
}
if ([int]$digest.project_count -ne [int]$summary.project_count) {
    throw "digest project_count $($digest.project_count) does not match summary project_count $($summary.project_count)"
}
if ($null -eq $digest.patterns -or @($digest.patterns).Count -eq 0) {
    throw "digest has no pattern entries"
}

Require-Text -Path $learningPath -Pattern "^## Pattern Playbook$"
Require-Text -Path $learningPath -Pattern "^## Plugin Risk Queue$"
Require-Text -Path $learningPath -Pattern "^## Readiness Queue$"
Require-Text -Path $reportPath -Pattern "^## Pattern Representatives$"
Require-Text -Path $htmlPath -Pattern "Technique Corpus Report"
Require-Text -Path $htmlPath -Pattern "learning\.md"

Write-Host "ok: $OutDir"
Write-Host "projects: $($summary.project_count)"
Write-Host "patterns: $(@($digest.patterns).Count)"
