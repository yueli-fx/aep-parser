param(
    [string]$OutRoot = "tmp\technique_selfhost_gate"
)

$ErrorActionPreference = "Stop"

function Read-JsonOrNull {
    param([string[]]$Paths)

    foreach ($path in $Paths) {
        if (Test-Path -LiteralPath $path) {
            return [ordered]@{
                path = $path
                value = Get-Content -Raw -LiteralPath $path | ConvertFrom-Json
            }
        }
    }
    return $null
}

function Format-Value {
    param([AllowNull()][object]$Value)
    if ($null -eq $Value -or [string]$Value -eq "") {
        return "n/a"
    }
    return [string]$Value
}

try {
    $processInfo = Read-JsonOrNull @(
        (Join-Path $OutRoot "watch_process.json"),
        (Join-Path $OutRoot "start_dry_run\watch_process.json")
    )
    $watchInfo = Read-JsonOrNull @(
        (Join-Path $OutRoot "watch_status.json"),
        (Join-Path $OutRoot "watch_dry_run\watch_status.json")
    )
    $outcomeInfo = Read-JsonOrNull @(
        (Join-Path $OutRoot "latest_outcome.json")
    )

    Write-Host "Self-Hosted Status"
    Write-Host "OutRoot: $OutRoot"
    Write-Host ""

    Write-Host "Watch Process"
    if ($null -eq $processInfo) {
        Write-Host "- process status: missing"
    } else {
        $process = $processInfo.value
        $state = "n/a"
        if ($null -ne $process.pid) {
            $running = $null -ne (Get-Process -Id ([int]$process.pid) -ErrorAction SilentlyContinue)
            if ($running) {
                $state = "running"
            } else {
                $state = "not-running"
            }
        } elseif ([string]$process.mode -eq "dry_run") {
            $state = "dry-run"
        }
        Write-Host "- file: $($processInfo.path)"
        Write-Host "- mode: $(Format-Value $process.mode)"
        Write-Host "- pid: $(Format-Value $process.pid)"
        Write-Host "- state: $state"
        Write-Host "- command: $(Format-Value $process.command)"
    }
    Write-Host ""

    Write-Host "Watch Status"
    if ($null -eq $watchInfo) {
        Write-Host "- watch status: missing"
    } else {
        $watch = $watchInfo.value
        Write-Host "- file: $($watchInfo.path)"
        Write-Host "- mode: $(Format-Value $watch.mode)"
        Write-Host "- completed iterations: $(Format-Value $watch.completed_iterations)"
        Write-Host "- last exit code: $(Format-Value $watch.last_exit_code)"
    }
    Write-Host ""

    Write-Host "Latest Outcome"
    if ($null -eq $outcomeInfo) {
        Write-Host "- latest outcome: missing"
    } else {
        $outcome = $outcomeInfo.value
        Write-Host "- file: $($outcomeInfo.path)"
        Write-Host "- headline: $(Format-Value $outcome.outcome_summary.headline)"
        Write-Host "- status: $(Format-Value $outcome.outcome_status.status)"
        Write-Host "- open: $(Format-Value $outcome.stable_outputs.open_target)"
    }
    Write-Host ""

    Write-Host "Logs"
    if ($null -eq $processInfo) {
        Write-Host "- stdout: n/a"
        Write-Host "- stderr: n/a"
    } else {
        $process = $processInfo.value
        Write-Host "- stdout: $(Format-Value $process.stdout_log)"
        Write-Host "- stderr: $(Format-Value $process.stderr_log)"
    }
} catch {
    Write-Error $_
    exit 1
}
