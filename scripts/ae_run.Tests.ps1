BeforeAll {
    . $PSScriptRoot/AeRun.Lib.ps1
}

Describe 'Lib smoke' {
    It 'dot-source does not throw' {
        $true | Should -Be $true
    }
}

Describe 'Test-FileStable' {
    BeforeEach {
        $script:tmp = Join-Path $env:TEMP "filestable-$(New-Guid).bin"
    }
    AfterEach {
        Remove-Item $script:tmp -ErrorAction SilentlyContinue
    }

    It 'returns $false when file is being written' {
        Set-Content -Path $script:tmp -Value 'a' -NoNewline
        $job = Start-Job -ScriptBlock {
            param($p)
            1..10 | ForEach-Object {
                Add-Content -Path $p -Value $_ -NoNewline
                Start-Sleep -Milliseconds 80
            }
        } -ArgumentList $script:tmp

        Test-FileStable -Path $script:tmp -WindowMs 300 -PollMs 50 | Should -Be $false

        Wait-Job $job | Out-Null
        Remove-Job $job
    }

    It 'returns $true when file has not changed for the window' {
        Set-Content -Path $script:tmp -Value 'done' -NoNewline
        Test-FileStable -Path $script:tmp -WindowMs 300 -PollMs 50 | Should -Be $true
    }

    It 'returns $false when file does not exist' {
        Test-FileStable -Path "$script:tmp-nope" -WindowMs 100 -PollMs 50 | Should -Be $false
    }
}
