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
