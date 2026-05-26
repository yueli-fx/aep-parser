BeforeAll {
    . $PSScriptRoot/AeRun.Lib.ps1
}

Describe 'Lib smoke' {
    It 'dot-source does not throw' {
        $true | Should -Be $true
    }
}
