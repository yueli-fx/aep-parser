Describe 'verify_recipe_profiles.ps1' {
    It 'verifies a single recipe and reports covered profile paths as JSON' {
        $repoRoot = Split-Path $PSScriptRoot -Parent
        $scriptPath = Join-Path $PSScriptRoot 'verify_recipe_profiles.ps1'
        $recipePath = Join-Path $repoRoot 'examples/recipes/minimal-comp-object-profile.json'

        $output = & pwsh -NoProfile -File $scriptPath -Recipe $recipePath -Json
        $LASTEXITCODE | Should -Be 0

        $summary = ($output | Out-String) | ConvertFrom-Json
        $summary.total | Should -Be 1
        $summary.failed | Should -Be 0
        $summary.covered_profile_paths | Should -Contain 'expected_profile.background_color'
    }
}
