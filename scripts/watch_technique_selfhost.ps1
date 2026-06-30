param(
    [string]$OutRoot = "tmp\technique_selfhost_gate",
    [int]$DurationMinutes = 60,
    [int]$IntervalSeconds = 300,
    [int]$Iterations = 0,
    [int]$Limit = 0,
    [switch]$OpenFirst,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

function New-VerifyArgs {
    param(
        [string]$OutRoot,
        [int]$Limit,
        [bool]$Open
    )

    $args = @("-NoProfile", "-File", "scripts\verify_technique_selfhost.ps1", "-OutRoot", $OutRoot)
    if ($Limit -gt 0) {
        $args += @("-Limit", [string]$Limit)
    }
    if ($Open) {
        $args += "-Open"
    }
    return $args
}

function Write-WatchStatus {
    param(
        [string]$OutRoot,
        [object]$Status
    )

    New-Item -ItemType Directory -Force -Path $OutRoot | Out-Null
    $Status | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $OutRoot "watch_status.json") -Encoding UTF8
}

try {
    if ($DurationMinutes -lt 0) {
        throw "DurationMinutes must be >= 0"
    }
    if ($IntervalSeconds -lt 1) {
        throw "IntervalSeconds must be >= 1"
    }
    if ($Iterations -lt 0) {
        throw "Iterations must be >= 0"
    }
    if (-not $DryRun -and $Iterations -eq 0 -and $DurationMinutes -eq 0) {
        throw "Iterations and DurationMinutes cannot both be 0"
    }

    $startedAt = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    $deadline = (Get-Date).AddMinutes($DurationMinutes)
    $verifyArgs = New-VerifyArgs -OutRoot $OutRoot -Limit $Limit -Open ([bool]$OpenFirst)
    $showArgs = @("-NoProfile", "-File", "scripts\show_technique_outcome.ps1", "-OutRoot", $OutRoot)

    if ($DryRun) {
        Write-Host "DRY RUN technique selfhost watch"
        Write-Host "verify command: pwsh $($verifyArgs -join ' ')"
        Write-Host "show command:   pwsh $($showArgs -join ' ')"
        Write-WatchStatus -OutRoot $OutRoot -Status ([ordered]@{
            mode = "dry_run"
            started_at_utc = $startedAt
            planned_iterations = $Iterations
            duration_minutes = $DurationMinutes
            interval_seconds = $IntervalSeconds
            verify_command = "pwsh $($verifyArgs -join ' ')"
            show_command = "pwsh $($showArgs -join ' ')"
        })
        exit 0
    }

    $iteration = 0
    $lastExit = 0
    while ($true) {
        if ($Iterations -gt 0 -and $iteration -ge $Iterations) {
            break
        }
        if ($Iterations -eq 0 -and $DurationMinutes -gt 0 -and (Get-Date) -gt $deadline) {
            break
        }

        $iteration++
        $iterationStartedAt = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
        Write-Host "watch iteration $iteration started at $iterationStartedAt"

        $verifyArgs = New-VerifyArgs -OutRoot $OutRoot -Limit $Limit -Open ([bool]($OpenFirst -and $iteration -eq 1))
        & pwsh @verifyArgs
        $lastExit = $LASTEXITCODE
        if ($lastExit -eq 0) {
            & pwsh @showArgs
            $lastExit = $LASTEXITCODE
        }

        Write-WatchStatus -OutRoot $OutRoot -Status ([ordered]@{
            mode = "watch"
            started_at_utc = $startedAt
            last_iteration_at_utc = $iterationStartedAt
            completed_iterations = $iteration
            last_exit_code = $lastExit
            latest_outcome_json = Join-Path $OutRoot "latest_outcome.json"
            latest_outcome_html = Join-Path $OutRoot "latest_outcome.html"
        })

        if ($lastExit -ne 0) {
            exit $lastExit
        }
        if ($Iterations -gt 0 -and $iteration -ge $Iterations) {
            break
        }
        if ($Iterations -eq 0 -and $DurationMinutes -gt 0 -and (Get-Date).AddSeconds($IntervalSeconds) -gt $deadline) {
            break
        }
        Start-Sleep -Seconds $IntervalSeconds
    }

    exit $lastExit
} catch {
    Write-Error $_
    exit 1
}
