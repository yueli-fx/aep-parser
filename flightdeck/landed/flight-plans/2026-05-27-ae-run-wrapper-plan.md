# ae_run.ps1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `scripts/ae_run.ps1` as drop-in replacement for `AfterFX -r` in Go ship-gate tests; OCR + multi-signal dispatch automatically消化 convert / save-changes / data-loss modals so ship-gate is unattended across AE versions.

**Architecture:** PowerShell 7+ main script orchestrates AE launch + tick loop; pure functions live in dot-sourced `scripts/AeRun.Lib.ps1` for Pester TDD; rules in external JSON; failure forensics dumped next to .done file. Multi-signal dispatch (windowTitle → windowClass → ocrMatch); SendKeys gated by `SetForegroundWindow + verify` focus protocol with per-(hwnd,rule) cooldown.

**Tech Stack:** PowerShell 7+ / Pester 5 / Add-Type P/Invoke (user32 / System.Drawing) / Windows.Media.Ocr (WinRT) / Go test driver (existing `runV2_*ShipGate`).

**Spec:** [`workshop/specs/2026-05-27-ae-run-wrapper-design.md`](../specs/2026-05-27-ae-run-wrapper-design.md)

---

## File Structure

**Create:**
- `scripts/ae_run.ps1` — main orchestration (param parse, AE launch, tick loop, cleanup)
- `scripts/AeRun.Lib.ps1` — testable pure-ish functions, dot-sourced by main + tests
- `scripts/ae_dialog_rules.json` — initial 3-rule dispatch table
- `scripts/ae_run.Tests.ps1` — Pester 5 unit tests

**Modify:**
- `internal/aep/shape_layer_shipgate_test.go:94` — `exec.Command(aeExe, "-r", jsxPath)` → ps1 invocation
- `internal/aep/new_composition_test.go:276` — same
- `workshop/playbooks/re-fixture.md` — § "GDI / 屏幕截图自动化" 改 planned → shipped
- `workshop/board.md` — Last updated / archive entry

**Rule files imported / used (no edits, just referenced):**
- `test_data/re_*.jsx` — JSX fixtures unchanged; `.done` contract unchanged

---

## Phase 1 — Setup

### Task 1: Scaffold scripts/ + verify Pester 5

**Files:**
- Create: `scripts/AeRun.Lib.ps1` (empty stub)
- Create: `scripts/ae_run.Tests.ps1` (smoke test only)

- [ ] **Step 1: Create empty lib + smoke test files**

```powershell
# scripts/AeRun.Lib.ps1
# Pure-ish functions for ae_run.ps1; dot-sourced by main + tests.
```

```powershell
# scripts/ae_run.Tests.ps1
BeforeAll {
    . $PSScriptRoot/AeRun.Lib.ps1
}

Describe 'Lib smoke' {
    It 'dot-source does not throw' {
        $true | Should -Be $true
    }
}
```

- [ ] **Step 2: Verify Pester 5 (install if missing — repo has 3.4 only)**

```powershell
pwsh -NoProfile -c "Get-Module -ListAvailable Pester | Select-Object Name,Version"
```

If max version < 5, install:

```powershell
pwsh -NoProfile -c "Install-Module Pester -MinimumVersion 5.5.0 -Force -SkipPublisherCheck -Scope CurrentUser"
```

Expected: Pester ≥ 5.5 listed.

- [ ] **Step 3: Run smoke test, verify Pester 5 picks it up**

Run:
```powershell
pwsh -NoProfile -c "Invoke-Pester -Path scripts/ae_run.Tests.ps1 -Output Detailed"
```

Expected: `Tests Passed: 1, Failed: 0`. If "BeforeAll outside of Describe" parse error, Pester 3.4 is being picked up — explicit import: `Import-Module Pester -MinimumVersion 5.5.0`.

- [ ] **Step 4: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "scaffold: scripts/ pwsh + Pester 5 smoke for ae_run wrapper"
```

---

## Phase 2 — Pure-ish functions (TDD)

### Task 2: `Test-FileStable` — file-stable polling

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

Replaces spec §3.7 "wait 1s" with `(size, mtime)` constant-over-window check.

- [ ] **Step 1: Write failing tests**

Append to `scripts/ae_run.Tests.ps1`:

```powershell
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
```

- [ ] **Step 2: Run, verify fail**

```powershell
pwsh -NoProfile -c "Invoke-Pester -Path scripts/ae_run.Tests.ps1 -Output Detailed"
```

Expected: 3 fails, "Test-FileStable is not recognized".

- [ ] **Step 3: Implement in AeRun.Lib.ps1**

```powershell
function Test-FileStable {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$Path,
        [int]$WindowMs = 500,
        [int]$PollMs = 100
    )
    if (-not (Test-Path -LiteralPath $Path)) { return $false }
    $deadline = (Get-Date).AddMilliseconds($WindowMs)
    $prev = Get-Item -LiteralPath $Path
    while ((Get-Date) -lt $deadline) {
        Start-Sleep -Milliseconds $PollMs
        if (-not (Test-Path -LiteralPath $Path)) { return $false }
        $cur = Get-Item -LiteralPath $Path
        if ($cur.Length -ne $prev.Length -or $cur.LastWriteTimeUtc -ne $prev.LastWriteTimeUtc) {
            return $false
        }
        $prev = $cur
    }
    return $true
}
```

- [ ] **Step 4: Run, verify pass**

Expected: 4 tests passed (3 + smoke).

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Test-FileStable for ae_run .done flush race"
```

---

### Task 3: `Parse-Rules` — JSON loader + schema validation

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

Per spec §3.4: rules JSON, schema = `{name, windowTitle[], windowClass[], ocrMatch[], action, keys, cooldownMs, comment?}`; reject if all three match-arrays empty.

- [ ] **Step 1: Write failing tests**

```powershell
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
```

- [ ] **Step 2: Run, verify fail**

Expected: 4 fails.

- [ ] **Step 3: Implement**

```powershell
function Parse-Rules {
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$Path)

    $raw = Get-Content -LiteralPath $Path -Raw -ErrorAction Stop
    $parsed = $raw | ConvertFrom-Json -ErrorAction Stop

    foreach ($rule in $parsed) {
        if (-not $rule.name) { throw "rule missing 'name': $($rule | ConvertTo-Json -Compress)" }
        $wt = @($rule.windowTitle)
        $wc = @($rule.windowClass)
        $om = @($rule.ocrMatch)
        if ($wt.Count -eq 0 -and $wc.Count -eq 0 -and $om.Count -eq 0) {
            throw "rule '$($rule.name)': at least one of windowTitle/windowClass/ocrMatch must be non-empty"
        }
        if (-not $rule.action) { throw "rule '$($rule.name)': missing 'action'" }
        if (-not $rule.keys)   { throw "rule '$($rule.name)': missing 'keys'" }
        if ($null -eq $rule.cooldownMs) { throw "rule '$($rule.name)': missing 'cooldownMs'" }
    }
    return ,$parsed
}
```

- [ ] **Step 4: Run, verify pass**

Expected: All previous + 4 new = 8 passed.

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Parse-Rules JSON loader with multi-signal schema validation"
```

---

### Task 4: `Match-Rule` — three-layer dispatch match

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

Per spec §3.3: windowTitle → windowClass → ocrMatch; return `{rule, layer}` or `$null`.

- [ ] **Step 1: Write failing tests**

```powershell
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
```

- [ ] **Step 2: Run, verify fail**

Expected: 6 fails.

- [ ] **Step 3: Implement**

```powershell
function Match-Rule {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][hashtable]$HwndInfo,
        [Parameter(Mandatory)]$Rules
    )

    function _matchAny([string]$text, $needles) {
        if ([string]::IsNullOrEmpty($text)) { return $false }
        foreach ($n in $needles) {
            if (-not $n) { continue }
            if ($text.IndexOf($n, [StringComparison]::OrdinalIgnoreCase) -ge 0) {
                return $true
            }
        }
        return $false
    }

    foreach ($rule in $Rules) {
        if (_matchAny $HwndInfo.Title $rule.windowTitle) {
            return @{ rule = $rule; layer = 'title' }
        }
        if (_matchAny $HwndInfo.Class $rule.windowClass) {
            return @{ rule = $rule; layer = 'class' }
        }
        if (_matchAny $HwndInfo.Ocr $rule.ocrMatch) {
            return @{ rule = $rule; layer = 'ocr' }
        }
    }
    return $null
}
```

- [ ] **Step 4: Run, verify pass**

Expected: 6 new pass, total 14.

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Match-Rule three-layer dispatch (title→class→ocr)"
```

---

### Task 5: `Cooldown` set helpers — `New-Cooldown / Test-InCooldown / Add-Cooldown`

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

Per spec §3.8: per-(hwnd, ruleName) cooldown set; reject re-match within window.

- [ ] **Step 1: Write failing tests**

```powershell
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
```

- [ ] **Step 2: Run, verify fail**

- [ ] **Step 3: Implement**

```powershell
function New-Cooldown {
    return @{}
}

function Add-Cooldown {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][hashtable]$Cooldown,
        [Parameter(Mandatory)]$Hwnd,
        [Parameter(Mandatory)][string]$Rule,
        [Parameter(Mandatory)][int]$DurationMs
    )
    $key = "$Hwnd|$Rule"
    $Cooldown[$key] = (Get-Date).AddMilliseconds($DurationMs)
}

function Test-InCooldown {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][hashtable]$Cooldown,
        [Parameter(Mandatory)]$Hwnd,
        [Parameter(Mandatory)][string]$Rule
    )
    $key = "$Hwnd|$Rule"
    if (-not $Cooldown.ContainsKey($key)) { return $false }
    if ((Get-Date) -ge $Cooldown[$key]) {
        $Cooldown.Remove($key) | Out-Null
        return $false
    }
    return $true
}
```

- [ ] **Step 4: Run, verify pass**

Expected: 4 new pass.

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Cooldown set (per hwnd+rule TTL) for SendKeys re-trigger guard"
```

---

### Task 6: `Initialize-Win32` — Add-Type P/Invoke surface

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

P/Invoke surface for `EnumWindows / GetWindowThreadProcessId / GetWindow / GetWindowTextW / GetClassNameW / GetWindowLong / IsWindowVisible / GetWindowRect / SetForegroundWindow / GetForegroundWindow`. Wrap in a single `Initialize-Win32` (idempotent — repeated calls no-op).

- [ ] **Step 1: Write failing test**

```powershell
Describe 'Initialize-Win32' {
    It 'creates the AeRunWin32 type and is idempotent' {
        Initialize-Win32
        ([AeRunWin32]) | Should -Not -BeNullOrEmpty
        # second call must not throw "type already exists"
        { Initialize-Win32 } | Should -Not -Throw
    }

    It 'GetForegroundWindow returns a non-zero hwnd' {
        Initialize-Win32
        [AeRunWin32]::GetForegroundWindow() | Should -Not -Be ([IntPtr]::Zero)
    }
}
```

- [ ] **Step 2: Run, verify fail**

Expected: `Unable to find type [AeRunWin32]`.

- [ ] **Step 3: Implement**

```powershell
function Initialize-Win32 {
    if ('AeRunWin32' -as [type]) { return }
    Add-Type -Namespace '' -Name 'AeRunWin32' -MemberDefinition @"
        public delegate bool EnumWindowsProc(System.IntPtr hWnd, System.IntPtr lParam);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, System.IntPtr lParam);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern uint GetWindowThreadProcessId(System.IntPtr hWnd, out uint lpdwProcessId);

        [System.Runtime.InteropServices.DllImport("user32.dll", CharSet=System.Runtime.InteropServices.CharSet.Unicode)]
        public static extern int GetWindowTextW(System.IntPtr hWnd, System.Text.StringBuilder lpString, int nMaxCount);

        [System.Runtime.InteropServices.DllImport("user32.dll", CharSet=System.Runtime.InteropServices.CharSet.Unicode)]
        public static extern int GetClassNameW(System.IntPtr hWnd, System.Text.StringBuilder lpClassName, int nMaxCount);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool IsWindowVisible(System.IntPtr hWnd);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern System.IntPtr GetWindow(System.IntPtr hWnd, uint uCmd);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern int GetWindowLong(System.IntPtr hWnd, int nIndex);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        [return: System.Runtime.InteropServices.MarshalAs(System.Runtime.InteropServices.UnmanagedType.Bool)]
        public static extern bool GetWindowRect(System.IntPtr hWnd, out RECT lpRect);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool SetForegroundWindow(System.IntPtr hWnd);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern System.IntPtr GetForegroundWindow();

        public const uint GW_OWNER = 4;
        public const int GWL_EXSTYLE = -20;
        public const int WS_EX_DLGMODALFRAME = 0x00000001;

        public struct RECT {
            public int Left; public int Top; public int Right; public int Bottom;
        }
"@
}
```

- [ ] **Step 4: Run, verify pass**

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Initialize-Win32 P/Invoke surface for window enum + focus"
```

---

### Task 7: `Get-AeModals` — enumerate modal windows of AE root pid

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

Per spec §3.3 Layer A: candidate filter = visible + pid==AeRoot + (title!=main OR has owner OR `WS_EX_DLGMODALFRAME`). Return list of `@{Hwnd; Title; Class; Rect; Pid}`.

- [ ] **Step 1: Write failing tests**

```powershell
Describe 'Get-AeModals' {
    BeforeAll {
        Initialize-Win32
    }

    It 'returns empty list when no AE-pid windows exist' {
        Get-AeModals -AeRootPid 999999 -MainTitleHints @('Adobe After Effects') | Should -BeNullOrEmpty
    }

    It 'returns the current pwsh console window when called with own pid and a non-matching main title' {
        # current process has at least the console window
        $list = Get-AeModals -AeRootPid $PID -MainTitleHints @('___NOPE___')
        $list.Count | Should -BeGreaterThan 0
        $list[0].Pid | Should -Be $PID
    }

    It 'filters out own window when its title matches MainTitleHints' {
        # use current process title — should be excluded
        $self = Get-Process -Id $PID
        $title = $self.MainWindowTitle
        if ([string]::IsNullOrEmpty($title)) {
            Set-ItResult -Skipped -Because 'no console title to test with'
            return
        }
        $list = Get-AeModals -AeRootPid $PID -MainTitleHints @($title)
        # main-titled window must be excluded; if any others remain they'd be popups
        ($list | Where-Object Title -eq $title).Count | Should -Be 0
    }
}
```

- [ ] **Step 2: Run, verify fail**

- [ ] **Step 3: Implement**

```powershell
function Get-AeModals {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][int]$AeRootPid,
        [string[]]$MainTitleHints = @()
    )
    Initialize-Win32
    $results = New-Object System.Collections.Generic.List[hashtable]

    $cb = [AeRunWin32+EnumWindowsProc]{
        param([IntPtr]$hWnd, [IntPtr]$lParam)
        try {
            if (-not [AeRunWin32]::IsWindowVisible($hWnd)) { return $true }

            $procId = 0
            [void][AeRunWin32]::GetWindowThreadProcessId($hWnd, [ref]$procId)
            if ($procId -ne $AeRootPid) { return $true }

            $sb = New-Object System.Text.StringBuilder 512
            [void][AeRunWin32]::GetWindowTextW($hWnd, $sb, 512)
            $title = $sb.ToString()

            $sb.Clear() | Out-Null
            [void][AeRunWin32]::GetClassNameW($hWnd, $sb, 512)
            $cls = $sb.ToString()

            $owner   = [AeRunWin32]::GetWindow($hWnd, [AeRunWin32]::GW_OWNER)
            $exStyle = [AeRunWin32]::GetWindowLong($hWnd, [AeRunWin32]::GWL_EXSTYLE)
            $isDlgFrame = ($exStyle -band [AeRunWin32]::WS_EX_DLGMODALFRAME) -ne 0

            $titleIsMain = $false
            foreach ($hint in $MainTitleHints) {
                if ($title -and $title.IndexOf($hint, [StringComparison]::OrdinalIgnoreCase) -ge 0) {
                    $titleIsMain = $true; break
                }
            }

            $isModalCandidate = (-not $titleIsMain) -or ($owner -ne [IntPtr]::Zero) -or $isDlgFrame
            if (-not $isModalCandidate) { return $true }

            $rect = New-Object AeRunWin32+RECT
            [void][AeRunWin32]::GetWindowRect($hWnd, [ref]$rect)

            $results.Add(@{
                Hwnd  = $hWnd
                Title = $title
                Class = $cls
                Rect  = $rect
                Pid   = $procId
            }) | Out-Null
        } catch {}
        return $true
    }

    [void][AeRunWin32]::EnumWindows($cb, [IntPtr]::Zero)
    return $results
}
```

- [ ] **Step 4: Run, verify pass**

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Get-AeModals — Win32 enum filter for AE modal candidates"
```

---

## Phase 3 — Integration helpers (OCR / SendKeys / forensics)

### Task 8: `Invoke-Ocr` — Windows.Media.Ocr wrapper

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

WinRT async → sync via `AsTask`. Two helpers: `Initialize-Ocr` (engine + AsTask helper, idempotent) + `Invoke-Ocr -Bitmap` (returns plain text). Test calls Initialize and feeds a generated bitmap with known text — verify substring.

- [ ] **Step 1: Write failing tests**

```powershell
Describe 'Invoke-Ocr' {
    It 'Initialize-Ocr returns an OCR engine for user-profile language' {
        $engine = Initialize-Ocr
        $engine | Should -Not -BeNullOrEmpty
        # idempotent
        $engine2 = Initialize-Ocr
        $engine2 | Should -Be $engine
    }

    It 'reads "HELLO" from a programmatically drawn bitmap' {
        Initialize-Ocr | Out-Null
        Add-Type -AssemblyName System.Drawing
        $bmp = New-Object System.Drawing.Bitmap 400, 100
        $g = [System.Drawing.Graphics]::FromImage($bmp)
        $g.Clear([System.Drawing.Color]::White)
        $font = New-Object System.Drawing.Font 'Arial', 36, ([System.Drawing.FontStyle]::Bold)
        $g.DrawString('HELLO', $font, [System.Drawing.Brushes]::Black, 20, 20)
        $g.Dispose()

        $text = Invoke-Ocr -Bitmap $bmp
        $bmp.Dispose()

        $text | Should -Match 'HELLO'
    }
}
```

- [ ] **Step 2: Run, verify fail**

- [ ] **Step 3: Implement**

```powershell
$script:_ocrEngine = $null
$script:_asTaskMI  = $null

function Initialize-Ocr {
    if ($script:_ocrEngine) { return $script:_ocrEngine }

    # load WinRT types
    [void][Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime]
    [void][Windows.Graphics.Imaging.SoftwareBitmap, Windows.Foundation, ContentType=WindowsRuntime]
    [void][Windows.Graphics.Imaging.BitmapDecoder, Windows.Foundation, ContentType=WindowsRuntime]
    [void][Windows.Storage.Streams.InMemoryRandomAccessStream, Windows.Foundation, ContentType=WindowsRuntime]
    [void][Windows.Storage.Streams.RandomAccessStream, Windows.Foundation, ContentType=WindowsRuntime]

    # WindowsRuntimeSystemExtensions.AsTask<T>(IAsyncOperation<T>) — reflection grab
    $script:_asTaskMI = [System.WindowsRuntimeSystemExtensions].GetMethods() |
        Where-Object { $_.Name -eq 'AsTask' -and $_.IsGenericMethod -and $_.GetParameters().Count -eq 1 } |
        Select-Object -First 1

    $engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
    if (-not $engine) {
        throw "OCR engine unavailable — install zh-CN / en-US language pack via Settings > Time & Language > Language"
    }
    $script:_ocrEngine = $engine
    return $engine
}

function _AwaitWinRT($asyncOp, [type]$resultType) {
    $generic = $script:_asTaskMI.MakeGenericMethod($resultType)
    $task    = $generic.Invoke($null, @($asyncOp))
    $task.Wait()
    return $task.Result
}

function Invoke-Ocr {
    [CmdletBinding()]
    param([Parameter(Mandatory)][System.Drawing.Bitmap]$Bitmap)

    Initialize-Ocr | Out-Null

    # Bitmap → PNG byte stream → SoftwareBitmap (via BitmapDecoder)
    $ms = New-Object System.IO.MemoryStream
    $Bitmap.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
    $bytes = $ms.ToArray()
    $ms.Dispose()

    $rastream = New-Object Windows.Storage.Streams.InMemoryRandomAccessStream
    $writer   = New-Object Windows.Storage.Streams.DataWriter $rastream
    $writer.WriteBytes($bytes)
    _AwaitWinRT $writer.StoreAsync()           ([uint32])     | Out-Null
    _AwaitWinRT $writer.FlushAsync()           ([bool])       | Out-Null
    $writer.DetachStream() | Out-Null
    $rastream.Seek(0) | Out-Null

    $decoder = _AwaitWinRT ([Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($rastream)) ([Windows.Graphics.Imaging.BitmapDecoder])
    $sb      = _AwaitWinRT $decoder.GetSoftwareBitmapAsync() ([Windows.Graphics.Imaging.SoftwareBitmap])
    $result  = _AwaitWinRT ($script:_ocrEngine.RecognizeAsync($sb)) ([Windows.Media.Ocr.OcrResult])

    return $result.Text
}
```

> **Note** — the `AsTask` reflection approach is the standard WinRT-in-PowerShell pattern; if Step 4 throws on `MakeGenericMethod`, the fallback is to use `[Runspaces].GetAwaiter().GetResult()` which works for `Task<T>` but not directly for `IAsyncOperation`. If reflection fails, log the error and skip this task — leave Invoke-Ocr unimplemented, mark §3.5 Tesseract fallback as required and re-plan.

- [ ] **Step 4: Run, verify pass**

If second test ("HELLO") fails because OCR confidence is low for a 100×400 thin bitmap, increase font size to 60 and bitmap to 600×150. OCR readability on small text is acceptable; the test exists to prove the pipeline works.

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Invoke-Ocr — Windows.Media.Ocr WinRT wrapper for dialog text"
```

---

### Task 9: `Capture-WindowBitmap` — screenshot a hwnd rect

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

`CopyFromScreen` of the given RECT to a System.Drawing.Bitmap.

- [ ] **Step 1: Write failing test**

```powershell
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
```

- [ ] **Step 2: Run, verify fail**

- [ ] **Step 3: Implement**

```powershell
function Capture-WindowBitmap {
    [CmdletBinding()]
    param([Parameter(Mandatory)][AeRunWin32+RECT]$Rect)

    Add-Type -AssemblyName System.Drawing
    $w = [Math]::Max(1, $Rect.Right - $Rect.Left)
    $h = [Math]::Max(1, $Rect.Bottom - $Rect.Top)
    $bmp = New-Object System.Drawing.Bitmap $w, $h
    $g   = [System.Drawing.Graphics]::FromImage($bmp)
    try {
        $g.CopyFromScreen($Rect.Left, $Rect.Top, 0, 0, (New-Object System.Drawing.Size $w, $h))
    } finally {
        $g.Dispose()
    }
    return $bmp
}
```

- [ ] **Step 4: Run, verify pass**

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Capture-WindowBitmap — CopyFromScreen of hwnd rect"
```

---

### Task 10: `Invoke-SendKeysSafe` — focus protocol

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

Per spec §3.8: SetForegroundWindow → sleep → verify → SendKeys. Returns `@{Sent=$bool; FocusActual=$hwnd}`.

- [ ] **Step 1: Write failing tests**

```powershell
Describe 'Invoke-SendKeysSafe' {
    BeforeAll {
        Initialize-Win32
        Add-Type -AssemblyName System.Windows.Forms
    }

    It 'returns Sent=$false when SetForegroundWindow fails / focus mismatch' {
        # invalid hwnd → SetForegroundWindow returns false; FocusActual != target
        $r = Invoke-SendKeysSafe -Hwnd ([IntPtr]0xDEADBEEF) -Keys '{ENTER}' -DelayMs 50
        $r.Sent | Should -Be $false
    }

    It 'sets sent flag when focus matches own console (best-effort smoke)' {
        # own foreground hwnd — focus protocol trivially passes
        $self = [AeRunWin32]::GetForegroundWindow()
        if ($self -eq [IntPtr]::Zero) {
            Set-ItResult -Skipped -Because 'no foreground window in this session'
            return
        }
        # use harmless key
        $r = Invoke-SendKeysSafe -Hwnd $self -Keys '+(z)' -DelayMs 50  # Shift+z, harmless echo
        $r.Sent | Should -Be $true
    }
}
```

- [ ] **Step 2: Run, verify fail**

- [ ] **Step 3: Implement**

```powershell
function Invoke-SendKeysSafe {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][IntPtr]$Hwnd,
        [Parameter(Mandatory)][string]$Keys,
        [int]$DelayMs = 200
    )
    Initialize-Win32
    Add-Type -AssemblyName System.Windows.Forms

    $brought = [AeRunWin32]::SetForegroundWindow($Hwnd)
    Start-Sleep -Milliseconds $DelayMs
    $actual = [AeRunWin32]::GetForegroundWindow()

    if (-not $brought -or $actual -ne $Hwnd) {
        return @{ Sent = $false; FocusActual = $actual }
    }

    [System.Windows.Forms.SendKeys]::SendWait($Keys)
    return @{ Sent = $true; FocusActual = $actual }
}
```

- [ ] **Step 4: Run, verify pass**

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Invoke-SendKeysSafe — SetForegroundWindow + verify protocol"
```

---

### Task 11: `Write-ForensicsDump` + `Write-ActionLog`

**Files:**
- Modify: `scripts/AeRun.Lib.ps1`
- Modify: `scripts/ae_run.Tests.ps1`

Per spec §3.6. `Write-ActionLog` appends timestamped lines to `<dump>/actions.log`; `Write-ForensicsDump` collects screenshot.png + ocr.txt + windows.txt + meta.json on failure.

- [ ] **Step 1: Write failing tests**

```powershell
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
```

- [ ] **Step 2: Run, verify fail**

- [ ] **Step 3: Implement**

```powershell
function Write-ActionLog {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$DumpDir,
        [Parameter(Mandatory)][string]$Event,
        [hashtable]$Data
    )
    if (-not (Test-Path $DumpDir)) { New-Item -ItemType Directory -Path $DumpDir | Out-Null }
    $ts   = (Get-Date).ToString('HH:mm:ss.fff')
    $kv   = if ($Data) {
        ($Data.GetEnumerator() | ForEach-Object { "$($_.Key)=$($_.Value)" }) -join ' '
    } else { '' }
    $line = "$ts  $Event  $kv".TrimEnd()
    Add-Content -LiteralPath (Join-Path $DumpDir 'actions.log') -Value $line
}

function Write-ForensicsDump {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$DumpDir,
        [Parameter(Mandatory)][int]$ExitCode,
        [Parameter(Mandatory)][string]$Reason,
        $Modals      = @(),
        [hashtable]$OcrTexts = @{},
        [hashtable]$Meta     = @{}
    )

    if (-not (Test-Path $DumpDir)) { New-Item -ItemType Directory -Path $DumpDir | Out-Null }

    # screenshot.png — full virtual screen
    Add-Type -AssemblyName System.Drawing
    Add-Type -AssemblyName System.Windows.Forms
    $bounds = [System.Windows.Forms.SystemInformation]::VirtualScreen
    $bmp = New-Object System.Drawing.Bitmap $bounds.Width, $bounds.Height
    $g   = [System.Drawing.Graphics]::FromImage($bmp)
    $g.CopyFromScreen($bounds.X, $bounds.Y, 0, 0, $bounds.Size)
    $g.Dispose()
    $bmp.Save((Join-Path $DumpDir 'screenshot.png'), [System.Drawing.Imaging.ImageFormat]::Png)
    $bmp.Dispose()

    # windows.txt
    $winLines = @()
    foreach ($m in $Modals) {
        $winLines += "hwnd=0x$('{0:X}' -f [int64]$m.Hwnd) pid=$($m.Pid) title='$($m.Title)' class='$($m.Class)' rect=($($m.Rect.Left),$($m.Rect.Top))-($($m.Rect.Right),$($m.Rect.Bottom))"
    }
    Set-Content -LiteralPath (Join-Path $DumpDir 'windows.txt') -Value $winLines

    # ocr.txt
    $ocrLines = @()
    foreach ($k in $OcrTexts.Keys) {
        $ocrLines += "=== hwnd=$k ==="
        $ocrLines += $OcrTexts[$k]
    }
    Set-Content -LiteralPath (Join-Path $DumpDir 'ocr.txt') -Value $ocrLines

    # meta.json
    $Meta.exitCode  = $ExitCode
    $Meta.reason    = $Reason
    $Meta.timestamp = (Get-Date).ToString('o')
    $Meta | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $DumpDir 'meta.json')
}
```

- [ ] **Step 4: Run, verify pass**

- [ ] **Step 5: Commit**

```bash
git add scripts/AeRun.Lib.ps1 scripts/ae_run.Tests.ps1
git commit -m "feat(scripts): Write-ActionLog + Write-ForensicsDump for failure forensics"
```

---

## Phase 4 — Main script + rules

### Task 12: `scripts/ae_dialog_rules.json` — initial 3-rule table

**Files:**
- Create: `scripts/ae_dialog_rules.json`

Per spec §3.4 — three OCR-only rules; title/class arrays empty (filled later when real dialogs profiled).

- [ ] **Step 1: Write file**

```json
[
  {
    "name": "convert-old-project",
    "windowTitle": [],
    "windowClass": [],
    "ocrMatch": ["convert", "转换", "升级项目"],
    "action": "SendKeys",
    "keys": "{ENTER}",
    "cooldownMs": 2000,
    "comment": "AE high-version opening low-version .aep — accept convert with Enter"
  },
  {
    "name": "save-changes-on-quit",
    "windowTitle": ["Adobe After Effects"],
    "windowClass": [],
    "ocrMatch": ["save changes", "保存更改"],
    "action": "SendKeys",
    "keys": "{TAB}{TAB}{ENTER}",
    "cooldownMs": 3000,
    "comment": "Leftover from prev JSX missing app.quit — walk to 'Don't Save' button (locale-sensitive)"
  },
  {
    "name": "file-data-missing",
    "windowTitle": [],
    "windowClass": [],
    "ocrMatch": ["file data is missing", "文件数据丢失"],
    "action": "SendKeys",
    "keys": "{ENTER}",
    "cooldownMs": 2000,
    "comment": "Mode 3 — fixture broken; press OK to let JSX catch + write .done(ERR)"
  }
]
```

- [ ] **Step 2: Verify it parses**

```powershell
pwsh -NoProfile -c ". scripts/AeRun.Lib.ps1; Parse-Rules -Path scripts/ae_dialog_rules.json | Format-Table name,ocrMatch"
```

Expected: 3 rows, no error.

- [ ] **Step 3: Commit**

```bash
git add scripts/ae_dialog_rules.json
git commit -m "feat(scripts): initial 3-rule dispatch table (convert / save / file-data-missing)"
```

---

### Task 13: `scripts/ae_run.ps1` — main orchestration

**Files:**
- Create: `scripts/ae_run.ps1`

Wires Phase 2-3 helpers; param-parsed CLI entry. ~120 lines.

- [ ] **Step 1: Write main script**

```powershell
#requires -Version 7.0

<#
.SYNOPSIS
  Drop-in replacement for `AfterFX -r <jsx>` in Go ship-gate tests.
  OCR + multi-signal dispatch dismisses convert / save / data-loss modals
  automatically; on unknown dialog or timeout, dumps forensics + non-zero exit.

.PARAMETER AeExe   AfterFX.exe absolute path
.PARAMETER Jsx     verify_*.jsx absolute path
.PARAMETER Done    .done marker absolute path (JSX writes this on completion)
.PARAMETER TimeoutSec   max wait for .done (default 120)
.PARAMETER RulesPath    dispatch table JSON (default scripts/ae_dialog_rules.json)

.OUTPUTS exit codes
   0 — .done found, PASS contract met (Go side reads .done first line)
   1 — timeout waiting for .done
   2 — unknown modal (OCR didn't match any rule)
   3 — OCR engine init failed
   4 — AE process failed to start
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$AeExe,
    [Parameter(Mandatory)][string]$Jsx,
    [Parameter(Mandatory)][string]$Done,
    [int]$TimeoutSec  = 120,
    [string]$RulesPath = (Join-Path $PSScriptRoot 'ae_dialog_rules.json'),
    [int]$TickMs      = 500
)

. $PSScriptRoot/AeRun.Lib.ps1

$ErrorActionPreference = 'Stop'
$dumpDir = "$Done.fail"

# pre-clean any leftover .done from prior run
Remove-Item -LiteralPath $Done -ErrorAction SilentlyContinue

try { $rules = Parse-Rules -Path $RulesPath }
catch { Write-Error "rules load failed: $_"; exit 5 }

try {
    Initialize-Win32
    Initialize-Ocr | Out-Null
} catch {
    Write-Error "OCR init failed: $_"
    Write-ActionLog -DumpDir $dumpDir -Event 'ocr-init-fail' -Data @{ err = "$_" }
    exit 3
}

$aeProc = Start-Process -FilePath $AeExe -ArgumentList @('-r', $Jsx) -PassThru -ErrorAction Stop
if (-not $aeProc -or $aeProc.HasExited) {
    Write-Error "AE failed to start"
    exit 4
}
$aeRootPid = $aeProc.Id
Write-ActionLog -DumpDir $dumpDir -Event 'ae-start' -Data @{ pid = $aeRootPid; exe = $AeExe }

$cooldown   = New-Cooldown
$deadline   = (Get-Date).AddSeconds($TimeoutSec)
$exitCode   = 0
$exitReason = 'ok'

try {
    while ($true) {
        # 1. .done?
        if (Test-Path -LiteralPath $Done) {
            Write-ActionLog -DumpDir $dumpDir -Event 'done-found'
            if (Test-FileStable -Path $Done -WindowMs 500 -PollMs 100) {
                Write-ActionLog -DumpDir $dumpDir -Event 'done-stable' -Data @{ size = (Get-Item $Done).Length }
                break
            }
        }

        # 7. timeout?
        if ((Get-Date) -ge $deadline) {
            $exitCode = 1; $exitReason = 'timeout'; break
        }

        # 2. enum modals
        $modals = Get-AeModals -AeRootPid $aeRootPid -MainTitleHints @('Adobe After Effects')
        if ($modals.Count -eq 0) {
            Start-Sleep -Milliseconds $TickMs
            continue
        }

        $handled = $false
        foreach ($m in $modals) {
            # 3. cooldown?
            # 4. Layer B — title/class
            $info = @{ Title = $m.Title; Class = $m.Class; Ocr = '' }
            $match = Match-Rule -HwndInfo $info -Rules $rules

            # 5. Layer C — OCR if title/class miss
            if (-not $match) {
                $bmp = Capture-WindowBitmap -Rect $m.Rect
                try { $ocrText = Invoke-Ocr -Bitmap $bmp } finally { $bmp.Dispose() }
                $info.Ocr = $ocrText
                $match = Match-Rule -HwndInfo $info -Rules $rules
            }

            if (-not $match) { continue }

            # cooldown check
            if (Test-InCooldown -Cooldown $cooldown -Hwnd $m.Hwnd -Rule $match.rule.name) {
                continue
            }

            Write-ActionLog -DumpDir $dumpDir -Event 'detect-modal' -Data @{
                hwnd  = ('0x{0:X}' -f [int64]$m.Hwnd)
                title = $m.Title; class = $m.Class
            }
            Write-ActionLog -DumpDir $dumpDir -Event "rule-match-$($match.layer)" -Data @{ name = $match.rule.name }

            # 6. SendKeys safety
            $r = Invoke-SendKeysSafe -Hwnd $m.Hwnd -Keys $match.rule.keys -DelayMs 200
            if (-not $r.Sent) {
                Write-ActionLog -DumpDir $dumpDir -Event 'focus-mismatch' -Data @{
                    rule = $match.rule.name
                    expected = ('0x{0:X}' -f [int64]$m.Hwnd)
                    actual   = ('0x{0:X}' -f [int64]$r.FocusActual)
                }
                continue
            }
            Write-ActionLog -DumpDir $dumpDir -Event 'sendkeys' -Data @{ keys = $match.rule.keys }
            Add-Cooldown -Cooldown $cooldown -Hwnd $m.Hwnd -Rule $match.rule.name -DurationMs $match.rule.cooldownMs
            $handled = $true
            break
        }

        if (-not $handled -and $modals.Count -gt 0) {
            # all visible modals were either: in cooldown, or no rule matched
            $anyUnknown = $false
            foreach ($m in $modals) {
                $info = @{ Title=$m.Title; Class=$m.Class; Ocr='' }
                if (Match-Rule -HwndInfo $info -Rules $rules) { continue }   # waiting on cooldown
                $bmp = Capture-WindowBitmap -Rect $m.Rect
                try { $info.Ocr = Invoke-Ocr -Bitmap $bmp } finally { $bmp.Dispose() }
                if (Match-Rule -HwndInfo $info -Rules $rules) { continue }
                $anyUnknown = $true; break
            }
            if ($anyUnknown) {
                $exitCode = 2; $exitReason = 'unknown-modal'; break
            }
        }

        Start-Sleep -Milliseconds $TickMs
    }
} finally {
    # cleanup AE
    if ($exitCode -ne 0) {
        $ocrTexts = @{}
        foreach ($m in (Get-AeModals -AeRootPid $aeRootPid -MainTitleHints @('Adobe After Effects'))) {
            try {
                $bmp = Capture-WindowBitmap -Rect $m.Rect
                $ocrTexts[('0x{0:X}' -f [int64]$m.Hwnd)] = Invoke-Ocr -Bitmap $bmp
                $bmp.Dispose()
            } catch {}
        }
        Write-ForensicsDump -DumpDir $dumpDir -ExitCode $exitCode -Reason $exitReason `
                            -Modals (Get-AeModals -AeRootPid $aeRootPid -MainTitleHints @('Adobe After Effects')) `
                            -OcrTexts $ocrTexts `
                            -Meta @{ aeExe = $AeExe; jsx = $Jsx; tickMs = $TickMs }
    }
    # graceful → force quit AE
    if (-not $aeProc.HasExited) {
        $waitDeadline = (Get-Date).AddSeconds(5)
        while (-not $aeProc.HasExited -and (Get-Date) -lt $waitDeadline) {
            Start-Sleep -Milliseconds 250
        }
        if (-not $aeProc.HasExited) {
            Stop-Process -Id $aeRootPid -Force -ErrorAction SilentlyContinue
            Write-ActionLog -DumpDir $dumpDir -Event 'ae-force-killed'
        } else {
            Write-ActionLog -DumpDir $dumpDir -Event 'ae-exited'
        }
    } else {
        Write-ActionLog -DumpDir $dumpDir -Event 'ae-exited'
    }
}

exit $exitCode
```

- [ ] **Step 2: Smoke test param parsing (without AE)**

```powershell
pwsh -NoProfile -File scripts/ae_run.ps1 -AeExe "C:/notfound.exe" -Jsx "C:/nope.jsx" -Done "C:/temp/x.done"
```

Expected: AE start fails → exit 4.

```powershell
echo $LASTEXITCODE
```

Expected: `4`.

- [ ] **Step 3: Commit**

```bash
git add scripts/ae_run.ps1
git commit -m "feat(scripts): ae_run.ps1 main — orchestrate AE launch + dispatch loop + forensics"
```

---

## Phase 5 — Smoke + Go drop-in

### Task 14: Manual smoke — clean run (no dialogs)

**Files:**
- (none new; uses existing `test_data/re_template.jsx` or similar)

Goal: verify drop-in works end-to-end on a fixture that should run cleanly (no convert prompt).

- [ ] **Step 1: Find / write a no-conflict fixture**

`test_data/re_template.jsx` exists. Verify it has the close+quit at the end (per playbook §"AE 退出不弹框"); if not, edit to add.

- [ ] **Step 2: Run wrapper directly**

```powershell
$done = "$env:TEMP/ae_run_smoke.done"
Remove-Item $done -ErrorAction SilentlyContinue
pwsh -NoProfile -File scripts/ae_run.ps1 `
    -AeExe "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" `
    -Jsx   "E:/projects/tools/aep-parser/test_data/re_template.jsx" `
    -Done  $done `
    -TimeoutSec 180
echo "exit=$LASTEXITCODE done-exists=$(Test-Path $done)"
```

Expected: `exit=0  done-exists=True`. If `exit=1` (timeout), check `$done.fail/` for dump and inspect screenshot — likely save-changes leftover; fix JSX to call `app.project.close + app.quit`.

- [ ] **Step 3: Inspect actions.log for sanity**

```powershell
Get-Content "$done.fail/actions.log" -ErrorAction SilentlyContinue
```

Expected on PASS: only `ae-start`, `done-found`, `done-stable`, `ae-exited`. On FAIL: dump dir contents tell the story.

- [ ] **Step 4: No commit** (smoke only)

---

### Task 15: Drop-in `shape_layer_shipgate_test.go`

**Files:**
- Modify: `internal/aep/shape_layer_shipgate_test.go:94-105`

Switch `exec.Command(aeExe, "-r", jsxPath)` to ps1 invocation; remove deadline loop (ps1 owns timeout).

- [ ] **Step 1: Find repo root from test**

Add helper or reuse: the file already lives in `internal/aep/`; from there the wrapper is `../../scripts/ae_run.ps1`. Use `runtime.Caller(0)` once.

- [ ] **Step 2: Edit ship-gate driver**

Replace lines 92-122 (the "3. AfterFX -r" + "4. wait for .done" block) with:

```go
// 3. ae_run.ps1 (drop-in for AfterFX -r — handles convert / save-changes / etc.)
repoRoot := filepath.Join(filepath.Dir(currentFile()), "..", "..")
script   := filepath.Join(repoRoot, "scripts", "ae_run.ps1")
cmd := exec.Command("pwsh", "-NoProfile", "-File", script,
    "-AeExe", aeExe,
    "-Jsx", jsxPath,
    "-Done", doneFile,
    "-TimeoutSec", "180")
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
if err := cmd.Run(); err != nil {
    t.Fatalf("ae_run.ps1 failed: %v (check %s.fail/ for dump)", err, doneFile)
}

// .done is guaranteed to exist after ps1 exit 0 — no polling needed
```

And add (above `runV2_2ShipGate` or in a shared test helper file `ae_test_helpers.go`):

```go
func currentFile() string {
    _, file, _, _ := runtime.Caller(1)
    return file
}
```

with `import "runtime"`.

- [ ] **Step 3: Run ship-gate**

```bash
AE_SHIP_GATE=1 go test ./internal/aep/ -run TestV2_2_AEShipGate_AE2025 -v -timeout 5m
```

Expected: PASS. Failure → check `<doneFile>.fail/` (test creates inputAEP in `t.TempDir()`, so dump lives there — `cp -r` it out before the test cleanup if needed).

- [ ] **Step 4: Commit**

```bash
git add internal/aep/shape_layer_shipgate_test.go
git commit -m "refactor(test): V2.2 ship-gate uses ae_run.ps1 drop-in"
```

---

### Task 16: Drop-in `new_composition_test.go` (V2.1 ship-gate)

**Files:**
- Modify: `internal/aep/new_composition_test.go:274-291`

Same change as Task 15.

- [ ] **Step 1: Apply same diff** at lines 274-291. Replace:

```go
// 3. AfterFX -r
jsxPath := `E:/projects/tools/aep-parser/test_data/verify_v2_1.jsx`
cmd := exec.Command(aeExe, "-r", jsxPath)
if err := cmd.Start(); err != nil {
    t.Fatalf("start AE: %v", err)
}

// 4. wait for .done (90s timeout — AE cold start may take ~60s)
deadline := time.Now().Add(90 * time.Second)
for {
    if _, err := os.Stat(doneFile); err == nil {
        break
    }
    if time.Now().After(deadline) {
        t.Fatalf("timeout waiting for %s (90s)", doneFile)
    }
    time.Sleep(2 * time.Second)
}
```

with the same ps1-invocation block from Task 15 step 2 (substituting `verify_v2_1.jsx` for jsxPath).

Also remove now-unused `time` import if it's the only use, or leave it.

- [ ] **Step 2: Run**

```bash
AE_SHIP_GATE=1 go test ./internal/aep/ -run TestV2_1_AEShipGate_AE2025 -v -timeout 5m
AE_SHIP_GATE=1 go test ./internal/aep/ -run TestV2_1_AEShipGate_AE2020 -v -timeout 5m
```

Expected: both PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/aep/new_composition_test.go
git commit -m "refactor(test): V2.1 ship-gate uses ae_run.ps1 drop-in (AE 2020 + 2025)"
```

---

### Task 17: Cross-version smoke — convert dialog dismissal

**Files:**
- (no edits — smoke validates Task 12 rule 1 in real environment)

Goal: run the **AE 2020 fixture** through **AE 2025** to trigger the convert dialog; wrapper must dismiss it and PASS.

- [ ] **Step 1: Run V2.1 AE 2020 fixture through AE 2025**

Temporarily override `aeExe` in `TestV2_1_AEShipGate_AE2020` by env var:

```bash
AE_SHIP_GATE=1 AE2020_EXE="E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" \
    go test ./internal/aep/ -run TestV2_1_AEShipGate_AE2020 -v -timeout 5m
```

Expected: PASS. Wrapper detects convert dialog via OCR ("convert" or "升级项目"), dismisses with Enter, JSX runs, ship-gate passes.

If FAIL: `<doneFile>.fail/screenshot.png` + `ocr.txt` reveal what OCR saw. Common causes:
- OCR didn't see "convert" — actual dialog text different; update `ae_dialog_rules.json` rule 1 `ocrMatch` with the seen text + commit.
- Focus race — `actions.log` shows `focus-mismatch`. Bump `Invoke-SendKeysSafe -DelayMs 300` in main script.

- [ ] **Step 2: If new OCR text needed, update rule + commit**

```bash
# edit scripts/ae_dialog_rules.json — append observed text to rule 1 ocrMatch
git add scripts/ae_dialog_rules.json
git commit -m "fix(scripts): widen convert-old-project ocrMatch with observed AE 2025 text"
```

Then re-run.

- [ ] **Step 3: No commit if rules unchanged** — Cross-version PASS proves Task 12 rule.

---

## Phase 6 — Docs + close

### Task 18: Update playbook + board

**Files:**
- Modify: `workshop/playbooks/re-fixture.md` (§"GDI / 屏幕截图自动化")
- Modify: `workshop/board.md` (Last updated, Recently finished)

- [ ] **Step 1: Update playbook GDI section**

Replace the existing `## GDI / 屏幕截图自动化 (planned, 未实现)` section with:

````markdown
## GDI 自动化 — `scripts/ae_run.ps1`

Ship-gate 用 `scripts/ae_run.ps1` 替代裸 `AfterFX -r`：OCR + multi-signal dispatch 自动消化 convert / save / data-loss modal。任意 AE 版本 × 任意 fixture 都能 unattended 跑。

**调用契约**：
```powershell
pwsh -NoProfile -File scripts/ae_run.ps1 `
    -AeExe   $aeExe `
    -Jsx     $jsxPath `
    -Done    $doneFile `
    -TimeoutSec 180
```

退出码：
- `0` — `.done` 出现且 stable，按 PASS contract 走
- `1` — timeout
- `2` — 未知 modal（OCR 命中但没规则）
- `3` — OCR engine init 失败
- `4` — AE 进程启动失败

失败时 dump 落 `<doneFile>.fail/`：`screenshot.png` / `ocr.txt` / `windows.txt` / `actions.log` / `meta.json`。

**版本匹配规则降级**：仍**建议**用 fixture 源版本 AE 跑（避免无意义的 convert 流程），但**不再 mandatory**——wrapper 能消化。跨版本 ship-gate 现在可行（参考 Task 17 cross-version smoke）。

**新对话框出现的流程**：
1. ship-gate FAIL, exit code 2
2. 看 `<doneFile>.fail/screenshot.png` + `ocr.txt`
3. 加规则到 `scripts/ae_dialog_rules.json`（substring 进 `ocrMatch`，或抄稳定 `windowTitle` / `windowClass`）
4. `Parse-Rules` 跑通 → re-run ship-gate
````

- [ ] **Step 2: Update playbook frontmatter `last_updated`**

```diff
- last_updated: 2026-05-26
+ last_updated: 2026-05-27
```

- [ ] **Step 3: Update board.md**

Bump `Last updated`, add to "Recently finished" top:

```markdown
- **2026-05-27 ae_run.ps1 unattended ship-gate wrapper** — `scripts/ae_run.ps1` drop-in replaces `AfterFX -r` in V2.1 / V2.2 ship-gate; OCR (Windows.Media.Ocr) + multi-signal dispatch (windowTitle/Class/ocrMatch) + SendKeys focus protocol + per-(hwnd,rule) cooldown + forensics dump. Cross-version smoke PASS (AE 2020 fixture through AE 2025 with convert dialog auto-dismissed). 18 Pester unit tests; 2 Go ship-gate tests refactored. PASS 202 不变（无新 Go API）.
```

Update `Active focus`:

```markdown
**Active focus**: ae_run.ps1 闭环 — ship-gate unattended，可起 Phase 2 plan / nnhd 展开
```

Remove `(并行候选) scripts/ae_run.ps1 起 GDI 自动化 wrapper` from `## Next session`.

- [ ] **Step 4: Final ship-gate sanity**

```bash
AE_SHIP_GATE=1 go test ./internal/aep/ -run "TestV2_1_AEShipGate|TestV2_2_AEShipGate" -v -timeout 10m
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add workshop/playbooks/re-fixture.md workshop/board.md
git commit -m "docs(workshop): ae_run.ps1 GDI 自动化 playbook + board archive 闭环"
```

---

## Self-review checklist

| Spec section | Plan coverage |
|---|---|
| §3.1 接口契约 (drop-in) | Task 13 (script CLI) + Task 15/16 (Go drop-in) |
| §3.2 进程拓扑 | Task 13 main script structure |
| §3.3 三层过滤 (Layer A/B/C) | Task 6 (Win32) + Task 7 (Get-AeModals) + Task 4 (Match-Rule) + Task 8 (OCR) |
| §3.4 Multi-signal dispatch schema | Task 3 (Parse-Rules schema validation) + Task 12 (initial rules) |
| §3.5 Windows.Media.Ocr | Task 8 (Invoke-Ocr) |
| §3.6 Failure forensics | Task 11 (Write-ForensicsDump + Write-ActionLog) |
| §3.7 .done file-stable poll | Task 2 (Test-FileStable) + Task 13 main loop step 1 |
| §3.8 SendKeys safety protocol | Task 10 (Invoke-SendKeysSafe) + Task 5 (Cooldown set) |
| §3.9 Process tree (AE root pid only) | Task 7 enforces `procId -ne $AeRootPid → skip` |
| §4 Data flow | Task 13 main loop mirrors §4 step-by-step |
| §5.1 Unit tests | Tasks 2-11 are TDD Pester |
| §5.2 Integration tests | Task 14 (clean run smoke), Task 17 (cross-version) |
| §5.3 现有 ship-gate test 改造 | Task 15 + Task 16 |
| §6 Implementation order | Tasks 1-18 follow §6 expanded |
| §8 验收 | Task 14/15/17/18 cover all rows |

**Type / naming consistency check** — `Match-Rule` / `Test-FileStable` / `Get-AeModals` / `Invoke-Ocr` / `Capture-WindowBitmap` / `Invoke-SendKeysSafe` / `Write-ActionLog` / `Write-ForensicsDump` / `Initialize-Win32` / `Initialize-Ocr` / `Parse-Rules` / `New-Cooldown` / `Add-Cooldown` / `Test-InCooldown` — each defined in exactly one Task, used consistently downstream. ✅

**No placeholders** — all functions show full implementation; all tests show full assertion code; all bash/powershell commands explicit. ✅
