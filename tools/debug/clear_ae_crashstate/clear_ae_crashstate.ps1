<#
.SYNOPSIS
  Clear After Effects' "崩溃修复选项 / Safe Mode" crash-recovery flag.

.DESCRIPTION
  A force-killed AE (what ae_run.ps1 does on a failed/timed-out gate, and even a
  successful run whose post-done grace is too short) arms a crash flag, so the
  NEXT launch shows the "崩溃修复选项 / Safe Mode" recovery dialog and blocks
  startup. This tool launches AE with a quit-only script, dismisses that dialog
  by posting VK_RETURN (= 继续 / continue) straight to the dialog's message queue
  via PostMessage, then lets AE exit cleanly — which clears the flag.

  PostMessage is used on purpose, NOT SetForegroundWindow + SendKeys: the operator
  runs gates headless while a foreground game window holds a foreground-lock, so
  any focus-stealing approach fails (Sent=$false / focus-mismatch). PostMessage
  delivers to the hwnd's queue regardless of which window is foreground.

  Idempotent: if no crash flag is set, AE launches, runs the quit script, and
  exits with no dialog — a harmless no-op.

  Background: flightdeck/incidents/ae-automation-occlusion-crashstate.md §2 / §2c.

.PARAMETER AeExe
  Absolute path to AfterFX.exe. If omitted, probes the known installs
  (E:\adobe\...2025 then ...2020).

.PARAMETER TimeoutSec
  Max seconds to wait for AE to exit after launch (default 60).

.PARAMETER DialogWaitSec
  Max seconds to keep posting Enter to dialog-like AE windows while waiting for
  the safe-mode dialog to appear/dismiss (default 30).

.OUTPUTS
  Exit 0 — AE exited cleanly (flag cleared, or nothing to clear).
  Exit 1 — could not resolve AfterFX.exe, or AE did not exit within TimeoutSec.
#>
[CmdletBinding()]
param(
    [string]$AeExe,
    [int]$TimeoutSec = 60,
    [int]$DialogWaitSec = 30
)

$ErrorActionPreference = 'Stop'

if (-not $AeExe) {
    foreach ($cand in @(
        'E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe',
        'E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe')) {
        if (Test-Path -LiteralPath $cand) { $AeExe = $cand; break }
    }
}
if (-not $AeExe -or -not (Test-Path -LiteralPath $AeExe)) {
    [Console]::Error.WriteLine("clear_ae_crashstate: AfterFX.exe not found (pass -AeExe <path>)")
    exit 1
}

if (-not ('AeCrashClearWin32' -as [type])) {
    Add-Type -Namespace '' -Name 'AeCrashClearWin32' -MemberDefinition @"
        public delegate bool EnumWindowsProc(System.IntPtr hWnd, System.IntPtr lParam);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool EnumWindows(EnumWindowsProc lpEnumFunc, System.IntPtr lParam);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern uint GetWindowThreadProcessId(System.IntPtr hWnd, out uint lpdwProcessId);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool IsWindowVisible(System.IntPtr hWnd);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern System.IntPtr GetWindow(System.IntPtr hWnd, uint uCmd);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern int GetWindowLong(System.IntPtr hWnd, int nIndex);

        [System.Runtime.InteropServices.DllImport("user32.dll")]
        public static extern bool PostMessage(System.IntPtr hWnd, uint Msg, System.IntPtr wParam, System.IntPtr lParam);

        public const uint GW_OWNER = 4;
        public const int  GWL_EXSTYLE = -20;
        public const int  WS_EX_DLGMODALFRAME = 0x00000001;
        public const uint WM_KEYDOWN = 0x0100;
        public const uint WM_KEYUP   = 0x0101;
        public const int  VK_RETURN  = 0x0D;
"@
}

# Collect this AE pid's visible *dialog-like* top-level windows (owned OR with a
# modal dialog frame) — deliberately excludes the main app window, so we never
# post Enter into the editor.
function Get-DialogLikeHwnds([int]$AePid) {
    $found = New-Object System.Collections.Generic.List[IntPtr]
    $cb = [AeCrashClearWin32+EnumWindowsProc]{
        param([IntPtr]$hWnd, [IntPtr]$lParam)
        try {
            if (-not [AeCrashClearWin32]::IsWindowVisible($hWnd)) { return $true }
            $procId = 0
            [void][AeCrashClearWin32]::GetWindowThreadProcessId($hWnd, [ref]$procId)
            if ($procId -ne $AePid) { return $true }
            $owner   = [AeCrashClearWin32]::GetWindow($hWnd, [AeCrashClearWin32]::GW_OWNER)
            $exStyle = [AeCrashClearWin32]::GetWindowLong($hWnd, [AeCrashClearWin32]::GWL_EXSTYLE)
            $isDlgFrame = ($exStyle -band [AeCrashClearWin32]::WS_EX_DLGMODALFRAME) -ne 0
            if ($owner -ne [IntPtr]::Zero -or $isDlgFrame) { $found.Add($hWnd) | Out-Null }
        } catch {}
        return $true
    }
    [void][AeCrashClearWin32]::EnumWindows($cb, [IntPtr]::Zero)
    return $found
}

function Send-Enter([IntPtr]$hWnd) {
    $vk = [IntPtr][AeCrashClearWin32]::VK_RETURN
    [void][AeCrashClearWin32]::PostMessage($hWnd, [AeCrashClearWin32]::WM_KEYDOWN, $vk, [IntPtr]0x00000001)
    [void][AeCrashClearWin32]::PostMessage($hWnd, [AeCrashClearWin32]::WM_KEYUP,   $vk, [IntPtr][int64]0xC0000001)
}

# quit-only script: close the untitled project without saving (no save prompt),
# then quit — a clean exit is what actually clears the crash flag.
$jsx = Join-Path $env:TEMP "clear_ae_crashstate-$(New-Guid).jsx"
@'
try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
app.quit();
'@ | Set-Content -LiteralPath $jsx -Encoding UTF8

Write-Host "clear_ae_crashstate: launching $AeExe (quit-only) to clear safe-mode flag..."
$proc = Start-Process -FilePath $AeExe -ArgumentList '-r', "`"$jsx`"" -PassThru

$dismissed = 0
$deadline      = (Get-Date).AddSeconds($TimeoutSec)
$dialogDeadline = (Get-Date).AddSeconds($DialogWaitSec)
try {
    while (-not $proc.HasExited -and (Get-Date) -lt $deadline) {
        if ((Get-Date) -lt $dialogDeadline) {
            foreach ($h in Get-DialogLikeHwnds $proc.Id) {
                Send-Enter $h
                $dismissed++
            }
        }
        Start-Sleep -Milliseconds 500
    }
} finally {
    Remove-Item -LiteralPath $jsx -ErrorAction SilentlyContinue
}

if ($proc.HasExited) {
    Write-Host "clear_ae_crashstate: AE exited cleanly (posted Enter to dialog-like windows $dismissed time(s)). Crash flag cleared."
    exit 0
}

# Do NOT force-kill: that re-arms the very flag we are trying to clear.
[Console]::Error.WriteLine("clear_ae_crashstate: AE still running after ${TimeoutSec}s; left it alone (force-kill would re-arm the flag). Inspect manually.")
exit 1
