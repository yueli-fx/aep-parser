param(
    [string]$OutRoot = "tmp\technique_selfhost_gate",
    [string]$JsonPath = ""
)

$ErrorActionPreference = "Stop"

$argsList = @("run", "./cmd/aepselfhost", "outcome", "-out-root", $OutRoot)
if ($JsonPath -ne "") {
    $argsList += @("-json-path", $JsonPath)
}

& go @argsList
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
