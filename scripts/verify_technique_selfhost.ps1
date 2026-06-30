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
    $argsList = @("run", "./cmd/aepselfhost", "verify", "-input", $InputPath, "-out-root", $OutRoot)
    if ($Limit -gt 0) {
        $argsList += @("-limit", [string]$Limit)
    }
    if ($Open) {
        $argsList += "-open"
    }
    & go @argsList
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
