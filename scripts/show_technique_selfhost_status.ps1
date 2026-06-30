param(
    [string]$OutRoot = "tmp\technique_selfhost_gate"
)

$ErrorActionPreference = "Stop"

& go run ./cmd/aepselfhost status -out-root $OutRoot
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
