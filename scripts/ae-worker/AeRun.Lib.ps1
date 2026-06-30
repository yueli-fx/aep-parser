# scripts/ae-worker/AeRun.Lib.ps1
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
        if ($rule.action -notin @('SendKeys', 'Ignore', 'Abort', 'Crash')) { throw "rule '$($rule.name)': unknown action '$($rule.action)'" }
        if ($rule.action -eq 'SendKeys' -and -not $rule.keys) { throw "rule '$($rule.name)': missing 'keys'" }
        if ($rule.action -eq 'Abort' -and -not $rule.message) { throw "rule '$($rule.name)': Abort rule missing 'message' (shown to the operator on stderr)" }
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

    # OCR-specific: Windows.Media.Ocr returns CJK glyphs space-separated
    # ("修 复 选 项"); patterns are typically written contiguous ("修复选项").
    # Retry match on whitespace-stripped versions of BOTH so English rules
    # ("file data is missing") still hit literal matches above, and CJK
    # rules hit the stripped fallback below.
    function _matchAnyOcr([string]$text, $needles) {
        if (_matchAny $text $needles) { return $true }
        if ([string]::IsNullOrEmpty($text)) { return $false }
        $stripped = ($text -replace '\s+', '')
        if ([string]::IsNullOrEmpty($stripped)) { return $false }
        foreach ($n in $needles) {
            if (-not $n) { continue }
            $needleStripped = ($n -replace '\s+', '')
            if ($needleStripped -and $stripped.IndexOf($needleStripped, [StringComparison]::OrdinalIgnoreCase) -ge 0) {
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
        if (_matchAnyOcr $HwndInfo.Ocr $rule.ocrMatch) {
            return @{ rule = $rule; layer = 'ocr' }
        }
    }
    return $null
}

function Test-RectContains {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]$Outer,
        [Parameter(Mandatory)]$Inner
    )

    if (-not $Outer -or -not $Inner) { return $false }
    return (
        $Outer.Left -le $Inner.Left -and
        $Outer.Top -le $Inner.Top -and
        $Outer.Right -ge $Inner.Right -and
        $Outer.Bottom -ge $Inner.Bottom
    )
}

function Test-StartupSplashWrapperModal {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]$Candidate,
        $KnownStartupSplashModals = @()
    )

    if (-not $Candidate) { return $false }
    if ($Candidate.Class -ne '#32770') { return $false }
    if (-not [string]::IsNullOrEmpty([string]$Candidate.Title)) { return $false }

    foreach ($splash in @($KnownStartupSplashModals)) {
        if (-not $splash) { continue }
        if ($Candidate.Pid -ne $splash.Pid) { continue }
        if ([int64]$Candidate.Hwnd -eq [int64]$splash.Hwnd) { continue }
        if (Test-RectContains -Outer $Candidate.Rect -Inner $splash.Rect) {
            return $true
        }
    }
    return $false
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

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool PrintWindow(System.IntPtr hWnd, System.IntPtr hdcBlt, uint nFlags);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool PostMessage(System.IntPtr hWnd, uint Msg, System.IntPtr wParam, System.IntPtr lParam);

        public const uint PW_RENDERFULLCONTENT = 2;
        public const uint GW_OWNER = 4;
        public const int GWL_EXSTYLE = -20;
        public const int WS_EX_DLGMODALFRAME = 0x00000001;
        public const uint WM_KEYDOWN = 0x0100;
        public const uint WM_KEYUP   = 0x0101;

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

$script:_ocrAvailable = $null
$script:_ocrHelperPath = Join-Path $PSScriptRoot 'ocr_helper.ps1'

function Initialize-Ocr {
    # PowerShell 7 (.NET 6+) dropped WinRT projection. Windows.Media.Ocr is still
    # reachable from Windows PowerShell 5.1, so we shell out to powershell.exe
    # with a tiny helper script that prints recognized text to stdout.
    # Caches availability after first check.
    if ($null -ne $script:_ocrAvailable) { return $script:_ocrAvailable }
    if (-not (Test-Path -LiteralPath $script:_ocrHelperPath)) {
        $script:_ocrAvailable = $false
        return $false
    }
    if (-not (Get-Command powershell.exe -ErrorAction SilentlyContinue)) {
        $script:_ocrAvailable = $false
        return $false
    }
    # ping the helper with a 1x1 white PNG to verify the engine instantiates
    Add-Type -AssemblyName System.Drawing
    $bmp = New-Object System.Drawing.Bitmap 50, 50
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $g.Clear([System.Drawing.Color]::White)
    $g.Dispose()
    $ping = Join-Path $env:TEMP "ocr-ping-$(New-Guid).png"
    $bmp.Save($ping, [System.Drawing.Imaging.ImageFormat]::Png)
    $bmp.Dispose()
    try {
        $null = & powershell.exe -NoProfile -File $script:_ocrHelperPath -ImagePath $ping 2>&1
        $script:_ocrAvailable = ($LASTEXITCODE -eq 0)
    } catch {
        $script:_ocrAvailable = $false
    } finally {
        Remove-Item -LiteralPath $ping -ErrorAction SilentlyContinue
    }
    return $script:_ocrAvailable
}

function Invoke-Ocr {
    [CmdletBinding()]
    param([Parameter(Mandatory)][System.Drawing.Bitmap]$Bitmap)

    if (-not (Initialize-Ocr)) {
        throw "OCR helper unavailable (scripts/ae-worker/ocr_helper.ps1 + Windows PowerShell 5.1 required)"
    }

    $tmp = Join-Path $env:TEMP "ocr-$(New-Guid).png"
    $out = Join-Path $env:TEMP "ocr-$(New-Guid).txt"
    $Bitmap.Save($tmp, [System.Drawing.Imaging.ImageFormat]::Png)
    try {
        # OCR helper writes UTF-8 text to -OutFile to avoid pipe encoding mojibake.
        & powershell.exe -NoProfile -File $script:_ocrHelperPath -ImagePath $tmp -OutFile $out 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $out)) {
            throw "ocr_helper.ps1 exit=$LASTEXITCODE"
        }
        return (Get-Content -LiteralPath $out -Raw -Encoding UTF8)
    } finally {
        Remove-Item -LiteralPath $tmp -ErrorAction SilentlyContinue
        Remove-Item -LiteralPath $out -ErrorAction SilentlyContinue
    }
}

function Capture-WindowBitmap {
    # Occlusion-immune capture: with -Hwnd, PrintWindow(PW_RENDERFULLCONTENT)
    # renders the window's OWN content regardless of z-order — an interactive
    # session's editor/terminal sitting on top of the AE dialog no longer
    # poisons OCR with its own pixels (incidents/ae-automation-occlusion-
    # crashstate.md §1; previously only tmp_debug/capture_dialog.ps1 had this).
    # Falls back to the legacy screen-rect copy when no Hwnd is given or
    # PrintWindow fails (some GPU-composited windows refuse it).
    [CmdletBinding()]
    param(
        [AeRunWin32+RECT]$Rect,
        [IntPtr]$Hwnd = [IntPtr]::Zero
    )

    Add-Type -AssemblyName System.Drawing
    Initialize-Win32

    if ($Hwnd -ne [IntPtr]::Zero) {
        $r = New-Object AeRunWin32+RECT
        if ([AeRunWin32]::GetWindowRect($Hwnd, [ref]$r)) {
            $w = $r.Right - $r.Left
            $h = $r.Bottom - $r.Top
            if ($w -gt 0 -and $h -gt 0) {
                $bmp = New-Object System.Drawing.Bitmap $w, $h
                $g = [System.Drawing.Graphics]::FromImage($bmp)
                $ok = $false
                try {
                    $hdc = $g.GetHdc()
                    try { $ok = [AeRunWin32]::PrintWindow($Hwnd, $hdc, [AeRunWin32]::PW_RENDERFULLCONTENT) }
                    finally { $g.ReleaseHdc($hdc) }
                } finally {
                    $g.Dispose()
                }
                if ($ok) { return $bmp }
                $bmp.Dispose()
            }
        }
    }

    if ($null -eq $Rect) { throw 'Capture-WindowBitmap: need -Hwnd or -Rect' }
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

function Invoke-SendKeysSafe {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][IntPtr]$Hwnd,
        [Parameter(Mandatory)][string]$Keys,
        [int]$DelayMs = 200
    )
    Initialize-Win32

    # Deliver keys via PostMessage straight to the dialog's own message queue —
    # focus-independent. The old SetForegroundWindow+SendKeys path failed whenever
    # another window held the foreground lock (the operator runs gates headless
    # while a game window owns the foreground), logging endless focus-mismatch and
    # never dismissing the modal (incidents/ae-automation-occlusion-crashstate.md
    # §2c). PostMessage needs no focus; the dialog's modal loop translates VK_RETURN
    # → default button / VK_ESCAPE → cancel / VK_TAB → next control, same as keys.
    $vk = @{ '{ENTER}' = 0x0D; '{ESC}' = 0x1B; '{TAB}' = 0x09 }
    $tokens = [regex]::Matches($Keys, '\{[^}]+\}|.') | ForEach-Object { $_.Value }
    $sentAny = $false
    foreach ($tok in $tokens) {
        $code = $null
        if ($vk.ContainsKey($tok)) { $code = $vk[$tok] }
        elseif ($tok.Length -eq 1) { $code = [int]([string]$tok).ToUpper()[0] }
        if ($null -eq $code) { continue }
        $wp = [IntPtr]$code
        # Sent tracks PostMessage delivery success — false for an invalid hwnd
        # (window gone / bogus), true once the dialog's queue accepts the key.
        $ok = [AeRunWin32]::PostMessage($Hwnd, [AeRunWin32]::WM_KEYDOWN, $wp, [IntPtr]0x00000001)
        Start-Sleep -Milliseconds 30
        [void][AeRunWin32]::PostMessage($Hwnd, [AeRunWin32]::WM_KEYUP, $wp, [IntPtr][int64]0xC0000001)
        Start-Sleep -Milliseconds $DelayMs
        if ($ok) { $sentAny = $true }
    }
    return [pscustomobject]@{ Sent = $sentAny; FocusActual = [AeRunWin32]::GetForegroundWindow() }
}

function Write-ActionLog {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$DumpDir,
        [Parameter(Mandatory)][string]$Event,
        [hashtable]$Data
    )
    if (-not (Test-Path -LiteralPath $DumpDir)) {
        New-Item -ItemType Directory -Path $DumpDir | Out-Null
    }
    $ts = (Get-Date).ToString('HH:mm:ss.fff')
    $kv = if ($Data) {
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

    if (-not (Test-Path -LiteralPath $DumpDir)) {
        New-Item -ItemType Directory -Path $DumpDir | Out-Null
    }

    Add-Type -AssemblyName System.Drawing
    Add-Type -AssemblyName System.Windows.Forms
    $bounds = [System.Windows.Forms.SystemInformation]::VirtualScreen
    $bmp = New-Object System.Drawing.Bitmap $bounds.Width, $bounds.Height
    $g   = [System.Drawing.Graphics]::FromImage($bmp)
    try {
        $g.CopyFromScreen($bounds.X, $bounds.Y, 0, 0, $bounds.Size)
    } finally {
        $g.Dispose()
    }
    $bmp.Save((Join-Path $DumpDir 'screenshot.png'), [System.Drawing.Imaging.ImageFormat]::Png)
    $bmp.Dispose()

    $winLines = @()
    foreach ($m in $Modals) {
        $winLines += "hwnd=0x$('{0:X}' -f [int64]$m.Hwnd) pid=$($m.Pid) title='$($m.Title)' class='$($m.Class)' rect=($($m.Rect.Left),$($m.Rect.Top))-($($m.Rect.Right),$($m.Rect.Bottom))"
    }
    Set-Content -LiteralPath (Join-Path $DumpDir 'windows.txt') -Value $winLines

    $ocrLines = @()
    foreach ($k in $OcrTexts.Keys) {
        $ocrLines += "=== hwnd=$k ==="
        $ocrLines += $OcrTexts[$k]
    }
    Set-Content -LiteralPath (Join-Path $DumpDir 'ocr.txt') -Value $ocrLines

    $Meta.exitCode  = $ExitCode
    $Meta.reason    = $Reason
    $Meta.timestamp = (Get-Date).ToString('o')
    $Meta | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $DumpDir 'meta.json')
}
