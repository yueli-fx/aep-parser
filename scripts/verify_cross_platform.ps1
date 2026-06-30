$ErrorActionPreference = "Stop"

$oldGoos = $env:GOOS
$oldGoarch = $env:GOARCH
$oldCgo = $env:CGO_ENABLED

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64" },
    @{ GOOS = "darwin"; GOARCH = "arm64" },
    @{ GOOS = "linux"; GOARCH = "amd64" }
)

$packages = @(
    "./cmd/aepserver",
    "./cmd/aepdiff",
    "./cmd/aeprecipe",
    "./cmd/aepsearch",
    "./cmd/aeptechnique",
    "./cmd/aeoracle",
    "./cmd/aepselfhost",
    "./internal/aehost",
    "./internal/host",
    "./internal/profile",
    "./internal/profilediff",
    "./internal/projectindex",
    "./internal/recipe",
    "./internal/server",
    "./internal/technique"
)

try {
    foreach ($target in $targets) {
        $env:GOOS = $target.GOOS
        $env:GOARCH = $target.GOARCH
        $env:CGO_ENABLED = "0"
        Write-Host "building GOOS=$($target.GOOS) GOARCH=$($target.GOARCH)"
        go build $packages
    }
} finally {
    if ($null -eq $oldGoos) { Remove-Item Env:\GOOS -ErrorAction SilentlyContinue } else { $env:GOOS = $oldGoos }
    if ($null -eq $oldGoarch) { Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue } else { $env:GOARCH = $oldGoarch }
    if ($null -eq $oldCgo) { Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue } else { $env:CGO_ENABLED = $oldCgo }
}
