param(
    [Parameter(Mandatory = $true)]
    [string]$BaseDir,
    [Parameter(Mandatory = $true)]
    [string]$NewDir,
    [string]$OutDir = "",
    [int]$Top = 20
)

$ErrorActionPreference = "Stop"

$argsList = @("run", "./cmd/aepselfhost", "compare-reports", "-base", $BaseDir, "-new", $NewDir, "-top", [string]$Top)
if ($OutDir -ne "") {
    $argsList += @("-out", $OutDir)
}

& go @argsList
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
