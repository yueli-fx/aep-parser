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

Describe 'Match-Rule' {
    BeforeAll {
        $script:rules = @(
            [pscustomobject]@{
                name        = 'by-title'
                windowTitle = @('Adobe After Effects')
                windowClass = @()
                ocrMatch    = @()
                action      = 'SendKeys'; keys = '{ENTER}'; cooldownMs = 1000
            },
            [pscustomobject]@{
                name        = 'by-class'
                windowTitle = @()
                windowClass = @('#32770')
                ocrMatch    = @()
                action      = 'SendKeys'; keys = '{ESC}'; cooldownMs = 1000
            },
            [pscustomobject]@{
                name        = 'by-ocr'
                windowTitle = @()
                windowClass = @()
                ocrMatch    = @('file data is missing', '文件数据丢失')
                action      = 'SendKeys'; keys = '{ENTER}'; cooldownMs = 1000
            }
        )
    }

    It 'matches by window title (Layer B, no OCR needed)' {
        $info = @{ Title = 'Adobe After Effects 2025'; Class = 'AEMainWindow'; Ocr = '' }
        $m = Match-Rule -HwndInfo $info -Rules $script:rules
        $m.rule.name | Should -Be 'by-title'
        $m.layer | Should -Be 'title'
    }

    It 'matches by window class when title misses' {
        $info = @{ Title = ''; Class = '#32770'; Ocr = '' }
        $m = Match-Rule -HwndInfo $info -Rules $script:rules
        $m.rule.name | Should -Be 'by-class'
        $m.layer | Should -Be 'class'
    }

    It 'falls back to OCR (case-insensitive)' {
        $info = @{ Title = ''; Class = ''; Ocr = 'After Effects: File Data Is Missing' }
        $m = Match-Rule -HwndInfo $info -Rules $script:rules
        $m.rule.name | Should -Be 'by-ocr'
        $m.layer | Should -Be 'ocr'
    }

    It 'matches Chinese OCR substring' {
        $info = @{ Title = ''; Class = ''; Ocr = '错误: 文件数据丢失。请检查项目。' }
        $m = Match-Rule -HwndInfo $info -Rules $script:rules
        $m.rule.name | Should -Be 'by-ocr'
    }

    It 'returns $null when nothing matches' {
        $info = @{ Title = 'unknown'; Class = 'unknown'; Ocr = 'unknown' }
        Match-Rule -HwndInfo $info -Rules $script:rules | Should -BeNullOrEmpty
    }

    It 'OCR-only rule does NOT match when Ocr field is empty/null' {
        $info = @{ Title = ''; Class = ''; Ocr = $null }
        Match-Rule -HwndInfo $info -Rules $script:rules | Should -BeNullOrEmpty
    }
}

Describe 'Cooldown set' {
    It 'starts empty' {
        $cd = New-Cooldown
        Test-InCooldown -Cooldown $cd -Hwnd 0x1234 -Rule 'foo' | Should -Be $false
    }

    It 'remembers a key for the given duration' {
        $cd = New-Cooldown
        Add-Cooldown -Cooldown $cd -Hwnd 0x1234 -Rule 'foo' -DurationMs 200
        Test-InCooldown -Cooldown $cd -Hwnd 0x1234 -Rule 'foo' | Should -Be $true
        Start-Sleep -Milliseconds 250
        Test-InCooldown -Cooldown $cd -Hwnd 0x1234 -Rule 'foo' | Should -Be $false
    }

    It 'distinguishes hwnd' {
        $cd = New-Cooldown
        Add-Cooldown -Cooldown $cd -Hwnd 0x1234 -Rule 'foo' -DurationMs 5000
        Test-InCooldown -Cooldown $cd -Hwnd 0x5678 -Rule 'foo' | Should -Be $false
    }

    It 'distinguishes rule' {
        $cd = New-Cooldown
        Add-Cooldown -Cooldown $cd -Hwnd 0x1234 -Rule 'foo' -DurationMs 5000
        Test-InCooldown -Cooldown $cd -Hwnd 0x1234 -Rule 'bar' | Should -Be $false
    }
}

Describe 'Initialize-Win32' {
    It 'creates the AeRunWin32 type and is idempotent' {
        Initialize-Win32
        ([AeRunWin32]) | Should -Not -BeNullOrEmpty
        { Initialize-Win32 } | Should -Not -Throw
    }

    It 'GetForegroundWindow returns a non-zero hwnd' {
        Initialize-Win32
        [AeRunWin32]::GetForegroundWindow() | Should -Not -Be ([IntPtr]::Zero)
    }
}
