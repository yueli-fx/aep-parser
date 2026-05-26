# scripts/AeRun.Lib.ps1
# Pure-ish functions for ae_run.ps1; dot-sourced by main + tests.

function Test-FileStable {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Path,
        [int]$WindowMs = 500,
        [int]$PollMs = 100
    )
    if (-not (Test-Path -LiteralPath $Path)) { return $false }
    $deadline = (Get-Date).AddMilliseconds($WindowMs)
    $prev = Get-Item -LiteralPath $Path
    while ((Get-Date) -lt $deadline) {
        Start-Sleep -Milliseconds $PollMs
        if (-not (Test-Path -LiteralPath $Path)) { return $false }
        $cur = Get-Item -LiteralPath $Path
        if ($cur.Length -ne $prev.Length -or $cur.LastWriteTimeUtc -ne $prev.LastWriteTimeUtc) {
            return $false
        }
        $prev = $cur
    }
    return $true
}

function Parse-Rules {
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Path)

    $raw = Get-Content -LiteralPath $Path -Raw -ErrorAction Stop
    $parsed = $raw | ConvertFrom-Json -ErrorAction Stop

    foreach ($rule in $parsed) {
        if (-not $rule.name) { throw "rule missing 'name': $($rule | ConvertTo-Json -Compress)" }
        $wt = @($rule.windowTitle)
        $wc = @($rule.windowClass)
        $om = @($rule.ocrMatch)
        if ($wt.Count -eq 0 -and $wc.Count -eq 0 -and $om.Count -eq 0) {
            throw "rule '$($rule.name)': at least one of windowTitle/windowClass/ocrMatch must be non-empty"
        }
        if (-not $rule.action) { throw "rule '$($rule.name)': missing 'action'" }
        if (-not $rule.keys)   { throw "rule '$($rule.name)': missing 'keys'" }
        if ($null -eq $rule.cooldownMs) { throw "rule '$($rule.name)': missing 'cooldownMs'" }
    }
    return ,$parsed
}
