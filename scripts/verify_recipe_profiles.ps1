<#
.SYNOPSIS
  Validate and compile recipe JSON files, then summarize profile check coverage.

.PARAMETER Recipe
  One or more recipe JSON paths. Relative paths are resolved from the repo root.

.PARAMETER RecipeDir
  Directory scanned when -Recipe is not provided. Default: examples/recipes.

.PARAMETER Filter
  File filter used with -RecipeDir. Default: *.json.

.PARAMETER OutDir
  Directory for compiled .aep outputs and the temporary aeprecipe binary.
  Without this parameter, a temporary directory is created and removed.

.PARAMETER KeepOutputs
  Keep the work directory and compiled .aep files.

.PARAMETER Json
  Emit only a machine-readable JSON summary.

.OUTPUTS
  Exit 0 when every selected recipe validates, compiles, and all profile checks
  pass. Exit 1 for build, validation, compile, or profile check failures.
#>
[CmdletBinding()]
param(
    [string[]]$Recipe = @(),
    [string]$RecipeDir = 'examples/recipes',
    [string]$Filter = '*.json',
    [string]$OutDir = '',
    [switch]$KeepOutputs,
    [switch]$Json
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path $PSScriptRoot -Parent

function Resolve-RepoPath([string]$PathValue) {
    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return (Resolve-Path -LiteralPath $PathValue).Path
    }
    return (Resolve-Path -LiteralPath (Join-Path $repoRoot $PathValue)).Path
}

function Get-RelativePath([string]$PathValue) {
    try {
        $rootUri = [Uri]($repoRoot.TrimEnd('\') + '\')
        $pathUri = [Uri]$PathValue
        return [Uri]::UnescapeDataString($rootUri.MakeRelativeUri($pathUri).ToString()).Replace('/', '\')
    } catch {
        return $PathValue
    }
}

function ConvertTo-List($Value) {
    if ($null -eq $Value) { return @() }
    if ($Value -is [System.Array]) { return @($Value) }
    return @($Value)
}

function Invoke-JsonTool([string]$ToolPath, [string[]]$ArgumentList) {
    $output = & $ToolPath @ArgumentList 2>&1
    $exitCode = $LASTEXITCODE
    $text = ($output | Out-String).Trim()
    $parsed = $null
    $parseError = $null

    if ($text.Length -gt 0) {
        try {
            $parsed = $text | ConvertFrom-Json -ErrorAction Stop
        } catch {
            $parseError = $_.Exception.Message
        }
    }

    return [pscustomobject]@{
        ExitCode   = $exitCode
        Text       = $text
        Json       = $parsed
        ParseError = $parseError
    }
}

function New-EmptySummary([string]$FailureReason, [string]$Details) {
    return [ordered]@{
        schema_version         = 1
        total                  = 0
        passed                 = 0
        failed                 = 1
        validate_failed        = 0
        compile_failed         = 0
        invalid_reports        = 0
        profile_check_failures = 0
        covered_profile_paths  = @()
        recipes                = @()
        failure_reason         = $FailureReason
        details                = $Details
    }
}

function Write-Summary($Summary) {
    if ($Json) {
        $Summary | ConvertTo-Json -Depth 16
        return
    }

    Write-Output 'recipe profile verification'
    Write-Output ("recipes: {0}, passed: {1}, failed: {2}" -f $Summary.total, $Summary.passed, $Summary.failed)
    Write-Output ("validate_failed: {0}, compile_failed: {1}, invalid_reports: {2}, profile_check_failures: {3}" -f $Summary.validate_failed, $Summary.compile_failed, $Summary.invalid_reports, $Summary.profile_check_failures)
    Write-Output ("covered_profile_paths: {0}" -f $Summary.covered_profile_paths.Count)

    if ($Summary.PSObject.Properties['failure_reason'] -and $Summary.failure_reason) {
        Write-Output ("failure_reason: {0}" -f $Summary.failure_reason)
        if ($Summary.details) { Write-Output $Summary.details }
        return
    }

    foreach ($result in $Summary.recipes) {
        $status = if ($result.ok) { 'ok' } else { 'fail' }
        Write-Output ("[{0}] {1} validate={2} compile={3} checks={4} failed={5} caps={6}" -f $status, $result.recipe, $result.validate_ok, $result.compile_ok, $result.profile_check_count, $result.profile_check_failed_count, $result.capability_count)

        foreach ($check in $result.failed_checks) {
            Write-Output ("  check failed: {0} expected={1} actual={2}" -f $check.path, ($check.expected | ConvertTo-Json -Compress -Depth 8), ($check.actual | ConvertTo-Json -Compress -Depth 8))
        }

        if ($result.validate_error) { Write-Output ("  validate_error: {0}" -f $result.validate_error) }
        if ($result.compile_error) { Write-Output ("  compile_error: {0}" -f $result.compile_error) }
    }
}

$workRoot = $null
$ownsWorkRoot = $false

try {
    $recipePaths = @()
    if ($Recipe.Count -gt 0) {
        foreach ($path in $Recipe) {
            $recipePaths += Resolve-RepoPath $path
        }
    } else {
        $recipeDirPath = Resolve-RepoPath $RecipeDir
        $recipePaths = @(Get-ChildItem -LiteralPath $recipeDirPath -Filter $Filter -File | Sort-Object FullName | ForEach-Object { $_.FullName })
    }

    if ($recipePaths.Count -eq 0) {
        $summary = New-EmptySummary 'no_recipes' "No recipe files matched '$RecipeDir/$Filter'."
        Write-Summary $summary
        exit 1
    }

    if ([string]::IsNullOrWhiteSpace($OutDir)) {
        $workRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("aep-parser-recipe-profile-verify-{0:yyyyMMdd-HHmmss-fff}" -f (Get-Date))
        $ownsWorkRoot = $true
    } else {
        $workRoot = if ([System.IO.Path]::IsPathRooted($OutDir)) { $OutDir } else { Join-Path $repoRoot $OutDir }
    }
    New-Item -ItemType Directory -Force -Path $workRoot | Out-Null
    $workRoot = (Resolve-Path -LiteralPath $workRoot).Path

    $exeName = if ($env:OS -eq 'Windows_NT') { 'aeprecipe.exe' } else { 'aeprecipe' }
    $toolPath = Join-Path $workRoot $exeName

    Push-Location $repoRoot
    try {
        $buildOutput = & go build -o $toolPath ./cmd/aeprecipe 2>&1
        if ($LASTEXITCODE -ne 0) {
            $summary = New-EmptySummary 'tool_build_failed' (($buildOutput | Out-String).Trim())
            Write-Summary $summary
            exit 1
        }

        $results = @()
        $index = 0
        foreach ($recipePath in $recipePaths) {
            $index++
            $relativeRecipe = Get-RelativePath $recipePath
            $baseName = [System.IO.Path]::GetFileNameWithoutExtension($recipePath)
            $outputPath = Join-Path $workRoot ("{0:D3}-{1}.aep" -f $index, $baseName)

            $validate = Invoke-JsonTool $toolPath @('validate', '-recipe', $recipePath, '-json')
            $validateOk = $validate.ExitCode -eq 0 -and $null -ne $validate.Json -and $validate.Json.valid -eq $true
            $validateError = if ($validate.ParseError) { $validate.ParseError } elseif ($validate.ExitCode -ne 0) { $validate.Text } else { '' }

            $compile = Invoke-JsonTool $toolPath @('compile', '-recipe', $recipePath, '-out', $outputPath, '-json')
            $compileOk = $compile.ExitCode -eq 0 -and $null -ne $compile.Json -and $compile.Json.valid -eq $true
            $compileError = if ($compile.ParseError) { $compile.ParseError } elseif ($compile.ExitCode -ne 0) { $compile.Text } else { '' }

            $capabilities = @()
            if ($null -ne $compile.Json) {
                $capabilities = ConvertTo-List $compile.Json.capabilities
            } elseif ($null -ne $validate.Json) {
                $capabilities = ConvertTo-List $validate.Json.capabilities
            }

            $checks = @()
            if ($null -ne $compile.Json) {
                $checks = ConvertTo-List $compile.Json.profile_checks
            }
            $failedChecks = @($checks | Where-Object { $_.passed -ne $true })
            $checkPaths = @($checks | ForEach-Object { $_.path } | Where-Object { $_ } | Sort-Object -Unique)
            $validReport = (($null -ne $validate.Json -and $validate.Json.valid -eq $true) -and ($null -ne $compile.Json -and $compile.Json.valid -eq $true))
            $ok = $validateOk -and $compileOk -and $failedChecks.Count -eq 0

            $results += [pscustomobject]([ordered]@{
                recipe                     = $relativeRecipe
                ok                         = $ok
                validate_ok                = $validateOk
                compile_ok                 = $compileOk
                valid_report               = $validReport
                capability_count           = $capabilities.Count
                profile_check_count        = $checks.Count
                profile_check_failed_count = $failedChecks.Count
                profile_check_paths        = $checkPaths
                failed_checks              = $failedChecks
                output_path                = if ($KeepOutputs -or -not $ownsWorkRoot) { $outputPath } else { '' }
                validate_error             = $validateError
                compile_error              = $compileError
            })
        }

        $coveredPaths = @($results | ForEach-Object { $_.profile_check_paths } | Where-Object { $_ } | Sort-Object -Unique)
        $failedCount = @($results | Where-Object { -not $_.ok }).Count
        $profileCheckFailures = [int](($results | Measure-Object -Property profile_check_failed_count -Sum).Sum)
        $summary = [ordered]@{
            schema_version         = 1
            total                  = $results.Count
            passed                 = $results.Count - $failedCount
            failed                 = $failedCount
            validate_failed        = @($results | Where-Object { -not $_.validate_ok }).Count
            compile_failed         = @($results | Where-Object { -not $_.compile_ok }).Count
            invalid_reports        = @($results | Where-Object { -not $_.valid_report }).Count
            profile_check_failures = $profileCheckFailures
            covered_profile_paths  = $coveredPaths
            recipes                = $results
            work_dir               = if ($KeepOutputs -or -not $ownsWorkRoot) { $workRoot } else { '' }
        }

        Write-Summary $summary
        if ($summary.failed -gt 0) { exit 1 }
        exit 0
    } finally {
        Pop-Location
    }
} finally {
    if ($ownsWorkRoot -and -not $KeepOutputs -and $workRoot -and (Test-Path -LiteralPath $workRoot)) {
        $tempRoot = ([System.IO.Path]::GetTempPath()).TrimEnd('\')
        $resolvedWorkRoot = (Resolve-Path -LiteralPath $workRoot).Path
        if ($resolvedWorkRoot.StartsWith($tempRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
            Remove-Item -LiteralPath $resolvedWorkRoot -Recurse -Force
        }
    }
}
