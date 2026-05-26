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

Describe 'Parse-Rules' {
    BeforeEach {
        $script:rulesPath = Join-Path $env:TEMP "rules-$(New-Guid).json"
    }
    AfterEach {
        Remove-Item $script:rulesPath -ErrorAction SilentlyContinue
    }

    It 'parses a minimal valid rule with windowTitle only' {
        @"
[
  {
    "name": "test-rule",
    "windowTitle": ["foo"],
    "windowClass": [],
    "ocrMatch": [],
    "action": "SendKeys",
    "keys": "{ENTER}",
    "cooldownMs": 1000
  }
]
"@ | Set-Content -Path $script:rulesPath
        $rules = Parse-Rules -Path $script:rulesPath
        $rules.Count | Should -Be 1
        $rules[0].name | Should -Be 'test-rule'
        $rules[0].windowTitle | Should -Be @('foo')
    }

    It 'rejects rule with all three match arrays empty' {
        @"
[
  {
    "name": "bad",
    "windowTitle": [],
    "windowClass": [],
    "ocrMatch": [],
    "action": "SendKeys",
    "keys": "{ENTER}",
    "cooldownMs": 1000
  }
]
"@ | Set-Content -Path $script:rulesPath
        { Parse-Rules -Path $script:rulesPath } | Should -Throw "*at least one of windowTitle/windowClass/ocrMatch*"
    }

    It 'rejects rule missing required name' {
        '[{"windowTitle":["x"],"action":"SendKeys","keys":"{ENTER}","cooldownMs":1000}]' |
            Set-Content -Path $script:rulesPath
        { Parse-Rules -Path $script:rulesPath } | Should -Throw "*missing 'name'*"
    }

    It 'rejects malformed JSON' {
        'not json {' | Set-Content -Path $script:rulesPath
        { Parse-Rules -Path $script:rulesPath } | Should -Throw
    }
}
