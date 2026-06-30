param(
    [string]$OutRoot = "tmp\technique_selfhost_gate",
    [string]$JsonPath = ""
)

$ErrorActionPreference = "Stop"

try {
    if ($JsonPath -eq "") {
        $JsonPath = Join-Path $OutRoot "latest_outcome.json"
    }
    if (-not (Test-Path -LiteralPath $JsonPath)) {
        throw "outcome json missing: $JsonPath"
    }

    $outcome = Get-Content -Raw -LiteralPath $JsonPath | ConvertFrom-Json
    if ($null -eq $outcome.outcome_summary -or [string]$outcome.outcome_summary.headline -eq "") {
        throw "outcome json missing outcome_summary.headline"
    }

    Write-Host "Self-Hosted Outcome"
    Write-Host "Headline: $($outcome.outcome_summary.headline)"
    Write-Host "Status:   $($outcome.outcome_status.status)"
    Write-Host "Corpus:   $($outcome.corpus.parsed_projects) projects, $($outcome.corpus.technique_patterns) patterns, $($outcome.corpus.parse_errors) parse errors"
    Write-Host "Smoke:    $($outcome.closed_loop.batch_passed)/$($outcome.closed_loop.batch_attempted)"
    Write-Host "Open:     $($outcome.stable_outputs.open_target)"
    Write-Host ""
    Write-Host "Next Actions"

    $actions = @($outcome.action_plan.next_actions)
    if ($actions.Count -eq 0) {
        Write-Host "- none"
    } else {
        foreach ($action in $actions) {
            Write-Host "- P$($action.priority): $($action.title) - $($action.detail)"
        }
    }
} catch {
    Write-Error $_
    exit 1
}
