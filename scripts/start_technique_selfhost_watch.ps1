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

function New-WatchArgumentList {
    param(
        [string]$OutRoot,
        [int]$DurationMinutes,
        [int]$IntervalSeconds,
        [int]$Iterations,
        [int]$Limit,
        [bool]$OpenFirst
    )

    $args = @("-NoProfile", "-File", "scripts\watch_technique_selfhost.ps1", "-OutRoot", $OutRoot, "-DurationMinutes", [string]$DurationMinutes, "-IntervalSeconds", [string]$IntervalSeconds)
    if ($Iterations -gt 0) {
        $args += @("-Iterations", [string]$Iterations)
    }
    if ($Limit -gt 0) {
        $args += @("-Limit", [string]$Limit)
    }
    if ($OpenFirst) {
        $args += "-OpenFirst"
    }
    return $args
}

function Write-ProcessStatus {
    param(
        [string]$OutRoot,
        [object]$Status
    )

    New-Item -ItemType Directory -Force -Path $OutRoot | Out-Null
    $Status | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $OutRoot "watch_process.json") -Encoding UTF8
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

    New-Item -ItemType Directory -Force -Path $OutRoot | Out-Null
    $logDir = Join-Path $OutRoot "watch_logs"
    New-Item -ItemType Directory -Force -Path $logDir | Out-Null
    $stamp = (Get-Date).ToUniversalTime().ToString("yyyyMMddTHHmmssZ")
    $stdoutLog = Join-Path $logDir "watch_$stamp.out.log"
    $stderrLog = Join-Path $logDir "watch_$stamp.err.log"
    $argumentList = New-WatchArgumentList -OutRoot $OutRoot -DurationMinutes $DurationMinutes -IntervalSeconds $IntervalSeconds -Iterations $Iterations -Limit $Limit -OpenFirst ([bool]$OpenFirst)
    $commandLine = "pwsh $($argumentList -join ' ')"
    $startedAt = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

    if ($DryRun) {
        Write-Host "DRY RUN technique selfhost background start"
        Write-Host "Start-Process pwsh $($argumentList -join ' ')"
        Write-Host "stdout: $stdoutLog"
        Write-Host "stderr: $stderrLog"
        Write-ProcessStatus -OutRoot $OutRoot -Status ([ordered]@{
            mode = "dry_run"
            started_at_utc = $startedAt
            command = $commandLine
            stdout_log = $stdoutLog
            stderr_log = $stderrLog
            out_root = $OutRoot
        })
        exit 0
    }

    $process = Start-Process -FilePath "pwsh" -ArgumentList $argumentList -WorkingDirectory (Get-Location).Path -RedirectStandardOutput $stdoutLog -RedirectStandardError $stderrLog -WindowStyle Hidden -PassThru
    Write-ProcessStatus -OutRoot $OutRoot -Status ([ordered]@{
        mode = "started"
        started_at_utc = $startedAt
        pid = $process.Id
        command = $commandLine
        stdout_log = $stdoutLog
        stderr_log = $stderrLog
        watch_status = Join-Path $OutRoot "watch_status.json"
        latest_outcome_html = Join-Path $OutRoot "latest_outcome.html"
        latest_outcome_json = Join-Path $OutRoot "latest_outcome.json"
    })
    Write-Host "started technique selfhost watch pid=$($process.Id)"
    Write-Host "stdout: $stdoutLog"
    Write-Host "stderr: $stderrLog"
    Write-Host "status: $(Join-Path $OutRoot 'watch_process.json')"
} catch {
    Write-Error $_
    exit 1
}
