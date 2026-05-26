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

Describe 'Get-AeModals' {
    BeforeAll {
        Initialize-Win32
    }

    It 'returns empty list when no AE-pid windows exist' {
        Get-AeModals -AeRootPid 999999 -MainTitleHints @('Adobe After Effects') | Should -BeNullOrEmpty
    }

    It 'returns at least one window for own pid with non-matching main title' {
        $self = Get-Process -Id $PID
        if ([string]::IsNullOrEmpty($self.MainWindowTitle)) {
            Set-ItResult -Skipped -Because 'headless pwsh has no own visible window to enumerate'
            return
        }
        $list = Get-AeModals -AeRootPid $PID -MainTitleHints @('___NOPE___')
        $list.Count | Should -BeGreaterThan 0
        $list[0].Pid | Should -Be $PID
    }

    It 'filters out own window when its title matches MainTitleHints' {
        $self = Get-Process -Id $PID
        $title = $self.MainWindowTitle
        if ([string]::IsNullOrEmpty($title)) {
            Set-ItResult -Skipped -Because 'no console title to test with'
            return
        }
        $list = Get-AeModals -AeRootPid $PID -MainTitleHints @($title)
        ($list | Where-Object { $_.Title -eq $title }).Count | Should -Be 0
    }
}

Describe 'Invoke-Ocr (sub-shell to powershell.exe + ocr_helper.ps1)' {
    It 'Initialize-Ocr returns $true when helper + powershell.exe + OCR engine all work' {
        $ok = Initialize-Ocr
        if (-not $ok) {
            Set-ItResult -Skipped -Because 'OCR backend not available (helper missing or no language pack)'
            return
        }
        $ok | Should -Be $true
        # idempotent: cached
        Initialize-Ocr | Should -Be $true
    }

    It 'reads "HELLO OCR" from a programmatically drawn bitmap' {
        if (-not (Initialize-Ocr)) {
            Set-ItResult -Skipped -Because 'OCR backend unavailable'
            return
        }
        Add-Type -AssemblyName System.Drawing
        $bmp = New-Object System.Drawing.Bitmap 600, 150
        $g = [System.Drawing.Graphics]::FromImage($bmp)
        $g.Clear([System.Drawing.Color]::White)
        $font = New-Object System.Drawing.Font 'Arial', 60, ([System.Drawing.FontStyle]::Bold)
        $g.DrawString('HELLO OCR', $font, [System.Drawing.Brushes]::Black, 20, 20)
        $g.Dispose()

        $text = Invoke-Ocr -Bitmap $bmp
        $bmp.Dispose()

        $text | Should -Match 'HELLO'
    }
}

Describe 'Capture-WindowBitmap' {
    It 'returns a Bitmap of the given rect dimensions' {
        Initialize-Win32
        $rect = New-Object AeRunWin32+RECT -Property @{ Left=100; Top=100; Right=300; Bottom=200 }
        $bmp = Capture-WindowBitmap -Rect $rect
        try {
            $bmp.Width  | Should -Be 200
            $bmp.Height | Should -Be 100
        } finally {
            $bmp.Dispose()
        }
    }
}

Describe 'Invoke-SendKeysSafe' {
    BeforeAll {
        Initialize-Win32
        Add-Type -AssemblyName System.Windows.Forms
    }

    It 'returns Sent=$false when SetForegroundWindow fails / focus mismatch' {
        $r = Invoke-SendKeysSafe -Hwnd ([IntPtr]0xDEADBEEF) -Keys '{ENTER}' -DelayMs 50
        $r.Sent | Should -Be $false
    }

    It 'returns FocusActual reflecting current foreground when target hwnd cannot be made foreground' {
        $r = Invoke-SendKeysSafe -Hwnd ([IntPtr]0xDEADBEEF) -Keys '{ENTER}' -DelayMs 50
        # may be IntPtr::Zero in headless env; key invariant is Sent=$false
        $r.PSObject.Properties['FocusActual'] | Should -Not -BeNullOrEmpty
    }
}

Describe 'Write-ActionLog' {
    BeforeEach {
        $script:dump = Join-Path $env:TEMP "dump-$(New-Guid)"
        New-Item -ItemType Directory -Path $script:dump | Out-Null
    }
    AfterEach {
        Remove-Item -Recurse -Force $script:dump -ErrorAction SilentlyContinue
    }

    It 'creates actions.log and appends timestamped lines' {
        Write-ActionLog -DumpDir $script:dump -Event 'ae-start' -Data @{ pid = 12345 }
        Write-ActionLog -DumpDir $script:dump -Event 'sendkeys' -Data @{ keys = '{ENTER}' }
        $log = Get-Content (Join-Path $script:dump 'actions.log')
        $log.Count | Should -Be 2
        $log[0] | Should -Match '^\d{2}:\d{2}:\d{2}\.\d{3}\s+ae-start\s+pid=12345$'
        $log[1] | Should -Match 'keys=\{ENTER\}'
    }
}

Describe 'Write-ForensicsDump' {
    BeforeEach {
        $script:dump = Join-Path $env:TEMP "fdump-$(New-Guid)"
    }
    AfterEach {
        Remove-Item -Recurse -Force $script:dump -ErrorAction SilentlyContinue
    }

    It 'creates dump dir with all 4 artifacts on call' {
        Initialize-Win32
        $modals = Get-AeModals -AeRootPid $PID -MainTitleHints @('___nope___')
        Write-ForensicsDump -DumpDir $script:dump -ExitCode 1 -Reason 'timeout' `
                            -Modals $modals -OcrTexts @{} -Meta @{ aeVersion='2025' }

        (Test-Path (Join-Path $script:dump 'screenshot.png')) | Should -Be $true
        (Test-Path (Join-Path $script:dump 'windows.txt'))    | Should -Be $true
        (Test-Path (Join-Path $script:dump 'ocr.txt'))        | Should -Be $true
        (Test-Path (Join-Path $script:dump 'meta.json'))      | Should -Be $true

        $meta = Get-Content (Join-Path $script:dump 'meta.json') | ConvertFrom-Json
        $meta.exitCode | Should -Be 1
        $meta.reason   | Should -Be 'timeout'
    }
}
