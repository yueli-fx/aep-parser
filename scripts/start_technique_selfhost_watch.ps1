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

$argsList = @(
    "run", "./cmd/aepselfhost", "start-watch",
    "-out-root", $OutRoot,
    "-duration-minutes", [string]$DurationMinutes,
    "-interval-seconds", [string]$IntervalSeconds
)
if ($Iterations -gt 0) {
    $argsList += @("-iterations", [string]$Iterations)
}
if ($Limit -gt 0) {
    $argsList += @("-limit", [string]$Limit)
}
if ($OpenFirst) {
    $argsList += "-open-first"
}
if ($DryRun) {
    $argsList += "-dry-run"
}

& go @argsList
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
