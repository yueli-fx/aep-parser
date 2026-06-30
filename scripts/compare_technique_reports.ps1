param(
    [Parameter(Mandatory = $true)]
    [string]$BaseDir,
    [Parameter(Mandatory = $true)]
    [string]$NewDir,
    [string]$OutDir = "",
    [int]$Top = 20
)

$ErrorActionPreference = "Stop"

function Read-JsonFile {
    param([string]$Path)
    if (-not (Test-Path -LiteralPath $Path)) {
        throw "missing JSON file: $Path"
    }
    return Get-Content -Raw -LiteralPath $Path | ConvertFrom-Json
}

function Get-CountMap {
    param([object]$Counts)
    $map = @{}
    if ($null -eq $Counts) {
        return $map
    }
    foreach ($property in @($Counts.PSObject.Properties)) {
        $map[[string]$property.Name] = [int]$property.Value
    }
    return $map
}

function Compare-CountMap {
    param(
        [string]$Name,
        [object]$BaseCounts,
        [object]$NewCounts,
        [int]$Max = 20
    )
    $base = Get-CountMap -Counts $BaseCounts
    $new = Get-CountMap -Counts $NewCounts
    $keys = @($base.Keys + $new.Keys | Sort-Object -Unique)
    $rows = @()
    foreach ($key in $keys) {
        $baseValue = 0
        $newValue = 0
        if ($base.ContainsKey($key)) {
            $baseValue = [int]$base[$key]
        }
        if ($new.ContainsKey($key)) {
            $newValue = [int]$new[$key]
        }
        $delta = $newValue - $baseValue
        if ($delta -eq 0 -and $baseValue -eq $newValue) {
            continue
        }
        $rows += [pscustomobject]@{
            group = $Name
            name  = [string]$key
            base  = $baseValue
            new   = $newValue
            delta = $delta
        }
    }
    return @($rows | Sort-Object @{ Expression = { [Math]::Abs([int]$_.delta) }; Descending = $true }, name | Select-Object -First $Max)
}

function Add-ScalarDiff {
    param(
        [System.Collections.ArrayList]$Rows,
        [string]$Name,
        [int]$BaseValue,
        [int]$NewValue
    )
    [void]$Rows.Add([pscustomobject]@{
        name  = $Name
        base  = $BaseValue
        new   = $NewValue
        delta = $NewValue - $BaseValue
    })
}

if ($Top -le 0) {
    $Top = 20
}
if ($OutDir -eq "") {
    $OutDir = Join-Path $NewDir "compare"
}
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$baseSummary = Read-JsonFile -Path (Join-Path $BaseDir "summary.json")
$newSummary = Read-JsonFile -Path (Join-Path $NewDir "summary.json")
$baseDigest = Read-JsonFile -Path (Join-Path $BaseDir "digest.json")
$newDigest = Read-JsonFile -Path (Join-Path $NewDir "digest.json")

$scalarRows = [System.Collections.ArrayList]::new()
Add-ScalarDiff -Rows $scalarRows -Name "projects" -BaseValue ([int]$baseSummary.project_count) -NewValue ([int]$newSummary.project_count)
Add-ScalarDiff -Rows $scalarRows -Name "errors" -BaseValue ([int]$baseSummary.error_count) -NewValue ([int]$newSummary.error_count)
Add-ScalarDiff -Rows $scalarRows -Name "comps" -BaseValue ([int]$baseSummary.totals.comp_count) -NewValue ([int]$newSummary.totals.comp_count)
Add-ScalarDiff -Rows $scalarRows -Name "layers" -BaseValue ([int]$baseSummary.totals.layer_count) -NewValue ([int]$newSummary.totals.layer_count)
Add-ScalarDiff -Rows $scalarRows -Name "effects" -BaseValue ([int]$baseSummary.totals.effect_count) -NewValue ([int]$newSummary.totals.effect_count)
Add-ScalarDiff -Rows $scalarRows -Name "text_animators" -BaseValue ([int]$baseSummary.totals.text_animator_count) -NewValue ([int]$newSummary.totals.text_animator_count)
Add-ScalarDiff -Rows $scalarRows -Name "shape_operators" -BaseValue ([int]$baseSummary.totals.shape_operator_count) -NewValue ([int]$newSummary.totals.shape_operator_count)
Add-ScalarDiff -Rows $scalarRows -Name "dependency_edges" -BaseValue ([int]$baseSummary.totals.dependency_count) -NewValue ([int]$newSummary.totals.dependency_count)
Add-ScalarDiff -Rows $scalarRows -Name "patterns" -BaseValue (@($baseDigest.patterns).Count) -NewValue (@($newDigest.patterns).Count)

$countGroups = @(
    @{ Name = "readiness"; Base = $baseSummary.readiness_counts; New = $newSummary.readiness_counts },
    @{ Name = "patterns"; Base = $baseSummary.pattern_counts; New = $newSummary.pattern_counts },
    @{ Name = "archetypes"; Base = $baseSummary.archetype_counts; New = $newSummary.archetype_counts },
    @{ Name = "technique_hints"; Base = $baseSummary.hint_counts; New = $newSummary.hint_counts },
    @{ Name = "plugin_effects"; Base = $baseSummary.plugin_effect_counts; New = $newSummary.plugin_effect_counts },
    @{ Name = "effects"; Base = $baseSummary.effect_counts; New = $newSummary.effect_counts },
    @{ Name = "shape_families"; Base = $baseSummary.shape_families; New = $newSummary.shape_families },
    @{ Name = "text_animators"; Base = $baseSummary.text_animators; New = $newSummary.text_animators },
    @{ Name = "layer_roles"; Base = $baseSummary.layer_roles; New = $newSummary.layer_roles },
    @{ Name = "graph_edges"; Base = $baseSummary.graph_edges; New = $newSummary.graph_edges }
)

$countDiffs = @()
foreach ($group in $countGroups) {
    $countDiffs += Compare-CountMap -Name $group.Name -BaseCounts $group.Base -NewCounts $group.New -Max $Top
}

$jsonPath = Join-Path $OutDir "compare.json"
$mdPath = Join-Path $OutDir "compare.md"
$result = [ordered]@{
    schema_version = 1
    base_dir       = $BaseDir
    new_dir        = $NewDir
    scalar_diffs   = @($scalarRows)
    count_diffs    = @($countDiffs)
}
$result | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $jsonPath -Encoding UTF8

$b = [System.Text.StringBuilder]::new()
[void]$b.AppendLine("# Technique Report Compare")
[void]$b.AppendLine("")
[void]$b.AppendLine("- base: ``$BaseDir``")
[void]$b.AppendLine("- new: ``$NewDir``")
[void]$b.AppendLine("")
[void]$b.AppendLine("## Scalar Diffs")
[void]$b.AppendLine("")
[void]$b.AppendLine("| metric | base | new | delta |")
[void]$b.AppendLine("| --- | ---: | ---: | ---: |")
foreach ($row in $scalarRows) {
    [void]$b.AppendLine("| $($row.name) | $($row.base) | $($row.new) | $($row.delta) |")
}
[void]$b.AppendLine("")
[void]$b.AppendLine("## Count Diffs")
[void]$b.AppendLine("")
if ($countDiffs.Count -eq 0) {
    [void]$b.AppendLine("_no count changes_")
} else {
    [void]$b.AppendLine("| group | name | base | new | delta |")
    [void]$b.AppendLine("| --- | --- | ---: | ---: | ---: |")
    foreach ($row in $countDiffs) {
        [void]$b.AppendLine("| $($row.group) | $($row.name) | $($row.base) | $($row.new) | $($row.delta) |")
    }
}
$b.ToString() | Set-Content -LiteralPath $mdPath -Encoding UTF8

Write-Host "compare json: $jsonPath"
Write-Host "compare md:   $mdPath"
