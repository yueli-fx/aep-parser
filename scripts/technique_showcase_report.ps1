[CmdletBinding()]
param(
    [string]$InputPath = "flightdeck\showcase",
    [string]$OutDir = "tmp\technique_showcase_report",
    [int]$Limit = 0,
    [switch]$Verify,
    [switch]$Open
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent
Push-Location $repoRoot
try {
    $argsList = @("run", "./cmd/aepselfhost", "technique-report", "-input", $InputPath, "-out", $OutDir)
    if ($Limit -gt 0) {
        $argsList += @("-limit", [string]$Limit)
    }
    if ($Verify) {
        $argsList += "-verify"
    }
    & go @argsList
    $exitCode = $LASTEXITCODE
    if ($exitCode -eq 0 -and $Open) {
        Invoke-Item (Join-Path $OutDir "report.html")
    }
    exit $exitCode
}
finally {
    Pop-Location
}
