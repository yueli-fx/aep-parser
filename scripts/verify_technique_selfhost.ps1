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
    $acceptanceJsonPath = Join-Path $runRoot "acceptance.json"
    $acceptanceMdPath = Join-Path $runRoot "acceptance.md"
    $latestRunPath = Join-Path $OutRoot "latest_run.txt"
    $latestAcceptancePath = Join-Path $OutRoot "latest_acceptance.md"
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

    $fullManifest = Get-Content -Raw -LiteralPath (Join-Path $fullReportDir "manifest.json") | ConvertFrom-Json
    $partialManifest = Get-Content -Raw -LiteralPath (Join-Path $partialReportDir "manifest.json") | ConvertFrom-Json
    $compareSelf = Get-Content -Raw -LiteralPath (Join-Path $compareSelfDir "compare.json") | ConvertFrom-Json
    $comparePartial = Get-Content -Raw -LiteralPath (Join-Path $comparePartialDir "compare.json") | ConvertFrom-Json
    $studyRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "study_queue.csv") | Select-Object -First 5)
    $projectPlaybookRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "project_playbooks.csv") | Select-Object -First 5)
    $patternRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "patterns.csv") | Select-Object -First 5)
    $recreationBlockerRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "recreation_blockers.csv") | Select-Object -First 5)
    $signalLayerRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "signal_layers.csv") | Select-Object -First 5)
    $effectStackRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "effect_stacks.csv") | Select-Object -First 5)
    $shapeOperatorRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "shape_operators.csv") | Select-Object -First 5)
    $textAnimatorRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "text_animators.csv") | Select-Object -First 5)
    $learningActionRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "learning_actions.csv") | Select-Object -First 5)
    $mechanismRows = @(Import-Csv -LiteralPath (Join-Path $fullReportDir "mechanisms.csv") | Select-Object -First 8)

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
        steps = @($steps)
    }
    $acceptance | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $acceptanceJsonPath -Encoding UTF8

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
    [void]$b.AppendLine("")
    [void]$b.AppendLine("## Steps")
    [void]$b.AppendLine("")
    [void]$b.AppendLine("| step | exit | seconds |")
    [void]$b.AppendLine("| --- | ---: | ---: |")
    foreach ($step in $steps) {
        [void]$b.AppendLine("| $($step.name) | $($step.exit) | $($step.seconds) |")
    }
    $b.ToString() | Set-Content -LiteralPath $acceptanceMdPath -Encoding UTF8

    function Escape-Html {
        param([AllowNull()][object]$Value)
        if ($null -eq $Value) {
            return ""
        }
        return [System.Net.WebUtility]::HtmlEncode([string]$Value)
    }

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
    [void]$index.AppendLine("<section class=""panel"" style=""margin-top:16px""><h2>Mechanism Catalog Preview</h2><table><thead><tr><th>Category</th><th>Name</th><th>Count</th></tr></thead><tbody>")
    foreach ($row in $mechanismRows) {
        [void]$index.AppendLine("<tr><td>$(Escape-Html $row.category)</td><td>$(Escape-Html $row.name)</td><td>$(Escape-Html $row.count)</td></tr>")
    }
    [void]$index.AppendLine("</tbody></table></section>")
    [void]$index.AppendLine("<section class=""panel""><h2>Artifacts</h2><div class=""links"">")
    [void]$index.AppendLine("<a href=""$runRel/full_report/report.html"">Full report HTML</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/learning.md"">Learning index</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/project_playbooks.csv"">Project playbooks CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/study_queue.csv"">Study queue CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/study_tasks.csv"">Study tasks CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/recreation_blockers.csv"">Recreation blockers CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/signal_layers.csv"">Signal layers CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/effect_stacks.csv"">Effect stacks CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/shape_operators.csv"">Shape operators CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/text_animators.csv"">Text animators CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/learning_actions.csv"">Learning actions CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/mechanisms.csv"">Mechanisms CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/mechanism_examples.csv"">Mechanism examples CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/projects.csv"">Projects CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/patterns.csv"">Patterns CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/full_report/errors.csv"">Errors CSV</a>")
    [void]$index.AppendLine("<a href=""$runRel/partial_report/report.html"">Partial-error report HTML</a>")
    [void]$index.AppendLine("<a href=""$runRel/compare_self/compare.md"">Self compare</a>")
    [void]$index.AppendLine("<a href=""$runRel/compare_partial_to_full/compare.md"">Partial-to-full compare</a>")
    [void]$index.AppendLine("<a href=""$runRel/acceptance.md"">Acceptance markdown</a>")
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
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Project Playbooks Preview" -Quiet)) {
        throw "latest index missing Project Playbooks Preview"
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
    if (-not (Select-String -LiteralPath $latestIndexPath -Pattern "Mechanism Catalog Preview" -Quiet)) {
        throw "latest index missing Mechanism Catalog Preview"
    }
    Require-LatestIndexLink -Label "full report" -RelativePath "$runRel/full_report/report.html"
    Require-LatestIndexLink -Label "learning index" -RelativePath "$runRel/full_report/learning.md"
    Require-LatestIndexLink -Label "project playbooks" -RelativePath "$runRel/full_report/project_playbooks.csv"
    Require-LatestIndexLink -Label "study queue" -RelativePath "$runRel/full_report/study_queue.csv"
    Require-LatestIndexLink -Label "study tasks" -RelativePath "$runRel/full_report/study_tasks.csv"
    Require-LatestIndexLink -Label "recreation blockers" -RelativePath "$runRel/full_report/recreation_blockers.csv"
    Require-LatestIndexLink -Label "signal layers" -RelativePath "$runRel/full_report/signal_layers.csv"
    Require-LatestIndexLink -Label "effect stacks" -RelativePath "$runRel/full_report/effect_stacks.csv"
    Require-LatestIndexLink -Label "shape operators" -RelativePath "$runRel/full_report/shape_operators.csv"
    Require-LatestIndexLink -Label "text animators" -RelativePath "$runRel/full_report/text_animators.csv"
    Require-LatestIndexLink -Label "learning actions" -RelativePath "$runRel/full_report/learning_actions.csv"
    Require-LatestIndexLink -Label "mechanisms" -RelativePath "$runRel/full_report/mechanisms.csv"
    Require-LatestIndexLink -Label "mechanism examples" -RelativePath "$runRel/full_report/mechanism_examples.csv"
    Require-LatestIndexLink -Label "partial report" -RelativePath "$runRel/partial_report/report.html"
    Require-LatestIndexLink -Label "partial compare" -RelativePath "$runRel/compare_partial_to_full/compare.md"

    $runRoot | Set-Content -LiteralPath $latestRunPath -Encoding UTF8
    Copy-Item -LiteralPath $acceptanceMdPath -Destination $latestAcceptancePath -Force

    Write-Host "acceptance json: $acceptanceJsonPath"
    Write-Host "acceptance md:   $acceptanceMdPath"
    Write-Host "latest run:      $latestRunPath"
    Write-Host "latest summary:  $latestAcceptancePath"
    Write-Host "latest index:    $latestIndexPath"
    Write-Host "full report:     $fullReportDir"
    Write-Host "partial report:  $partialReportDir"
    if ($Open) {
        Invoke-Item $latestIndexPath
    }
}
finally {
    Pop-Location
}
