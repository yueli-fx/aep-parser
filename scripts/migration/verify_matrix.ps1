[CmdletBinding()]
param(
    [string]$OutRoot = "registry\evidence\versioned-aep-migration\migration_matrix_verify"
)

$ErrorActionPreference = "Stop"

function Invoke-Checked {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$Args
    )

    & $Args[0] $Args[1..($Args.Length - 1)]
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: $($Args -join ' ')"
    }
}

function Assert-MatrixSummary {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [string]$MatrixPath,
        [Parameter(Mandatory = $true)]
        [int]$Total,
        [Parameter(Mandatory = $true)]
        [int]$Passed,
        [Parameter(Mandatory = $true)]
        [int]$Blocked,
        [Parameter(Mandatory = $true)]
        [int]$Failed,
        [Parameter(Mandatory = $true)]
        [int]$Skipped
    )

    if (-not (Test-Path -LiteralPath $MatrixPath)) {
        throw "${Name}: missing matrix JSON at ${MatrixPath}"
    }

    $matrix = Get-Content -LiteralPath $MatrixPath -Raw | ConvertFrom-Json
    $summary = $matrix.summary
    $actual = @(
        [int]$summary.total,
        [int]$summary.passed,
        [int]$summary.blocked,
        [int]$summary.failed,
        [int]$summary.skipped
    )
    $expected = @($Total, $Passed, $Blocked, $Failed, $Skipped)

    for ($i = 0; $i -lt $expected.Length; $i++) {
        if ($actual[$i] -ne $expected[$i]) {
            throw "${Name}: summary drifted; got total=$($summary.total) pass=$($summary.passed) blocked=$($summary.blocked) failed=$($summary.failed) skipped=$($summary.skipped)"
        }
    }

    Write-Host "PASS ${Name}: total=$Total pass=$Passed blocked=$Blocked failed=$Failed skipped=$Skipped"
}

function Assert-VerifyReport {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [string]$ReportPath
    )

    if (-not (Test-Path -LiteralPath $ReportPath)) {
        throw "${Name}: missing verify JSON at ${ReportPath}"
    }

    $report = Get-Content -LiteralPath $ReportPath -Raw | ConvertFrom-Json
    if ($report.summary.status -ne "pass" -or $report.verification.profile_diff_status -ne "pass" -or [int]$report.verification.profile_diff_count -ne 0) {
        throw "${Name}: verify drifted; status=$($report.summary.status) profile_diff_status=$($report.verification.profile_diff_status) profile_diff_count=$($report.verification.profile_diff_count)"
    }

    Write-Host "PASS ${Name}: status=pass profile_diff_status=pass profile_diff_count=0"
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Push-Location $repoRoot
try {
    $fullOut = Join-Path $OutRoot "smoke_all"
    $explicitMatteOut = Join-Path $OutRoot "explicit_matte_ae2025"

    Invoke-Checked @(
        "go", "run", "./cmd/aepmigrate", "matrix",
        "-recipes", "examples\recipes",
        "-sources", "AE2020",
        "-targets", "all",
        "-out", $fullOut,
        "-ledger-out", (Join-Path $fullOut "ledger.md")
    )
    Assert-MatrixSummary `
        -Name "full W2020 no-AE matrix" `
        -MatrixPath (Join-Path $fullOut "matrix.json") `
        -Total 906 `
        -Passed 900 `
        -Blocked 0 `
        -Failed 0 `
        -Skipped 6

    $verifyCase = Join-Path $fullOut "minimal-adjustment-layer\AE2020_to_AE2025"
    $verifyReport = Join-Path $verifyCase "verify_report.json"
    Invoke-Checked @(
        "go", "run", "./cmd/aepmigrate", "verify",
        "-source", (Join-Path $verifyCase "source.aep"),
        "-target", (Join-Path $verifyCase "target.aep"),
        "-target-version", "AE2025",
        "-report", (Join-Path $verifyCase "convert_report.json"),
        "-out", $verifyReport
    )
    Assert-VerifyReport `
        -Name "standalone verify CLI smoke" `
        -ReportPath $verifyReport

    Invoke-Checked @(
        "go", "run", "./cmd/aepmigrate", "matrix",
        "-recipe", "examples\recipes\minimal-layer-explicit-matte.json",
        "-sources", "AE2025",
        "-targets", "AE2025",
        "-out", $explicitMatteOut,
        "-ledger-out", (Join-Path $explicitMatteOut "ledger.md")
    )
    Assert-MatrixSummary `
        -Name "explicit matte AE2025 contract" `
        -MatrixPath (Join-Path $explicitMatteOut "matrix.json") `
        -Total 1 `
        -Passed 1 `
        -Blocked 0 `
        -Failed 0 `
        -Skipped 0
}
finally {
    Pop-Location
}
