param(
    [string]$OutDir = "tmp\technique_showcase_report",
    [int]$MinProjects = 1
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path $PSScriptRoot -Parent
Push-Location $repoRoot
try {
    & go run ./cmd/aepselfhost verify-report -out-dir $OutDir -min-projects $MinProjects
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
