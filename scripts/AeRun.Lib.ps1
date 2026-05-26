# scripts/AeRun.Lib.ps1
# Pure-ish functions for ae_run.ps1; dot-sourced by main + tests.

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
