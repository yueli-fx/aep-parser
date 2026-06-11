<#
.SYNOPSIS
  Drop-in replacement for `AfterFX -r <jsx>` in Go ship-gate tests.
  OCR + multi-signal dispatch dismisses convert / save / data-loss modals
  automatically; on unknown dialog or timeout, dumps forensics + non-zero exit.

  Requires pwsh 7+ for itself; sub-shells to powershell.exe 5.1 for WinRT OCR.

.PARAMETER AeExe   AfterFX.exe absolute path
.PARAMETER Jsx     verify_*.jsx absolute path
.PARAMETER Done    .done marker absolute path (JSX writes this on completion)
.PARAMETER TimeoutSec   max wait for .done (default 180)
.PARAMETER RulesPath    dispatch table JSON (default scripts/ae_dialog_rules.json)

.OUTPUTS exit codes
   0 — .done found, PASS contract met
   1 — timeout waiting for .done
   2 — unknown modal (OCR didn't match any rule)
   3 — OCR engine init failed
   4 — AE process failed to start
   5 — rules file load failed
   6 — another AfterFX process already running (screen-rect OCR would read its
       dialogs; kill stragglers first, or pass -IgnoreRunningAe to proceed)
   7 — Abort rule fired: environment failure needing USER INTERVENTION (e.g.
       AE scripting write-access preference disabled); message on stderr
   8 — Crash rule fired: AE itself crashed ("After Effects 已崩溃 (0 :: 42)");
       dialog dismissed, .done can never appear. Cold-start crashes are
       transient — caller should warm-retry once; identical failure = real
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$AeExe,
    [Parameter(Mandatory)][string]$Jsx,
    [Parameter(Mandatory)][string]$Done,
    [int]$TimeoutSec  = 180,
    [string]$RulesPath = (Join-Path $PSScriptRoot 'ae_dialog_rules.json'),
    [int]$TickMs      = 500,
    [switch]$IgnoreRunningAe
)

. $PSScriptRoot/AeRun.Lib.ps1

$ErrorActionPreference = 'Stop'
$dumpDir = "$Done.fail"

Remove-Item -LiteralPath $Done -ErrorAction SilentlyContinue

try { $rules = Parse-Rules -Path $RulesPath }
catch { [Console]::Error.WriteLine("rules load failed: $_"); exit 5 }

# Concurrent-AE guard. Modal capture is by SCREEN RECT — a straggler AfterFX
# (or the user's own interactive session) puts ITS dialogs where we OCR, and
# Get-AeModals can enumerate the wrong process's windows. Fail fast instead of
# producing a confusing exit-2 cascade (incidents/ae-automation-occlusion-crashstate.md).
if (-not $IgnoreRunningAe) {
    $straggler = @(Get-Process -Name 'AfterFX*' -ErrorAction SilentlyContinue)
    if ($straggler.Count -gt 0) {
        $pids = ($straggler | ForEach-Object { $_.Id }) -join ', '
        [Console]::Error.WriteLine("AfterFX already running (pid $pids) — kill stragglers (Get-Process AfterFX* | Stop-Process -Force) or pass -IgnoreRunningAe")
        Write-ActionLog -DumpDir $dumpDir -Event 'ae-already-running' -Data @{ pids = $pids }
        exit 6
    }
}

try {
    Initialize-Win32
    if (-not (Initialize-Ocr)) { throw "OCR helper unavailable" }
} catch {
    [Console]::Error.WriteLine("init failed: $_")
    Write-ActionLog -DumpDir $dumpDir -Event 'init-fail' -Data @{ err = "$_" }
    exit 3
}

$aeProc = $null
try {
    $aeProc = Start-Process -FilePath $AeExe -ArgumentList @('-r', $Jsx) -PassThru -ErrorAction Stop
} catch {
    [Console]::Error.WriteLine("AE start failed: $_")
    exit 4
}
if (-not $aeProc -or $aeProc.HasExited) {
    [Console]::Error.WriteLine("AE failed to start")
    exit 4
}
$aeRootPid = $aeProc.Id
Write-ActionLog -DumpDir $dumpDir -Event 'ae-start' -Data @{ pid = $aeRootPid; exe = $AeExe }

$cooldown          = New-Cooldown
$unknownFirstSeen  = @{}    # hwnd-hex → DateTime first seen as unknown modal
$unknownGraceSec   = 30     # how long an unknown modal can persist before exit-2 (AE 2020 cold-start splash can linger ~20s on this machine)
$deadline          = (Get-Date).AddSeconds($TimeoutSec)
$exitCode          = 0
$exitReason        = 'ok'

try {
    while ($true) {
        # 1. .done?
        if (Test-Path -LiteralPath $Done) {
            Write-ActionLog -DumpDir $dumpDir -Event 'done-found'
            if (Test-FileStable -Path $Done -WindowMs 500 -PollMs 100) {
                Write-ActionLog -DumpDir $dumpDir -Event 'done-stable' -Data @{ size = (Get-Item -LiteralPath $Done).Length }
                break
            }
        }

        # timeout?
        if ((Get-Date) -ge $deadline) {
            $exitCode = 1; $exitReason = 'timeout'; break
        }

        # 2. enum modals
        $modals = Get-AeModals -AeRootPid $aeRootPid -MainTitleHints @('Adobe After Effects')
        if (-not $modals -or $modals.Count -eq 0) {
            Start-Sleep -Milliseconds $TickMs
            continue
        }

        # Prune unknownFirstSeen entries whose hwnd no longer appears
        $currentHwnds = @($modals | ForEach-Object { ('0x{0:X}' -f [int64]$_.Hwnd) })
        $stale = @($unknownFirstSeen.Keys | Where-Object { $_ -notin $currentHwnds })
        foreach ($k in $stale) { $unknownFirstSeen.Remove($k) | Out-Null }

        $persistentUnknownHwnd = $null
        $persistentUnknownInfo = $null

        foreach ($m in $modals) {
            # Layer B — title/class
            $info = @{ Title = $m.Title; Class = $m.Class; Ocr = '' }
            $match = Match-Rule -HwndInfo $info -Rules $rules
            $usedOcr = $false

            # Layer C — OCR if title/class miss
            if (-not $match) {
                $bmp = Capture-WindowBitmap -Hwnd $m.Hwnd -Rect $m.Rect
                try { $ocrText = Invoke-Ocr -Bitmap $bmp } finally { $bmp.Dispose() }
                $info.Ocr = $ocrText
                $match = Match-Rule -HwndInfo $info -Rules $rules
                $usedOcr = $true
            }

            if (-not $match) {
                # Track first-seen for grace-period escalation. Splash + transient
                # AE startup windows naturally dismiss well within $unknownGraceSec.
                $hwndKey = '0x{0:X}' -f [int64]$m.Hwnd
                if (-not $unknownFirstSeen.ContainsKey($hwndKey)) {
                    $unknownFirstSeen[$hwndKey] = Get-Date
                    Write-ActionLog -DumpDir $dumpDir -Event 'unknown-modal-seen' -Data @{
                        hwnd  = $hwndKey
                        title = $m.Title; class = $m.Class
                        ocr   = if ($usedOcr) { ($info.Ocr -replace "`n", ' / ') } else { '<not-attempted>' }
                    }
                }
                $age = ((Get-Date) - $unknownFirstSeen[$hwndKey]).TotalSeconds
                if ($age -ge $unknownGraceSec) {
                    $persistentUnknownHwnd = $hwndKey
                    $persistentUnknownInfo = $info
                }
                continue
            }

            # cooldown?
            if (Test-InCooldown -Cooldown $cooldown -Hwnd $m.Hwnd -Rule $match.rule.name) {
                continue
            }

            Write-ActionLog -DumpDir $dumpDir -Event 'detect-modal' -Data @{
                hwnd  = ('0x{0:X}' -f [int64]$m.Hwnd)
                title = $m.Title; class = $m.Class
            }
            Write-ActionLog -DumpDir $dumpDir -Event "rule-match-$($match.layer)" -Data @{ name = $match.rule.name }

            # Ignore action: known-benign window (e.g. startup splash). Don't
            # send keys, don't escalate — let it clear on its own (bounded by
            # TimeoutSec). Cooldown suppresses per-tick log spam.
            if ($match.rule.action -eq 'Ignore') {
                Add-Cooldown -Cooldown $cooldown -Hwnd $m.Hwnd -Rule $match.rule.name -DurationMs $match.rule.cooldownMs
                continue
            }

            # Abort action: hard environment failure that no keystroke can fix
            # (e.g. AE scripting write-access preference disabled) — the user
            # must intervene. Fail fast with exit 7 instead of dismissing and
            # burning TimeoutSec on a .done that can never appear.
            if ($match.rule.action -eq 'Abort') {
                Write-ActionLog -DumpDir $dumpDir -Event 'abort-rule' -Data @{ name = $match.rule.name }
                [Console]::Error.WriteLine("USER INTERVENTION REQUIRED [$($match.rule.name)]: $($match.rule.message)")
                $exitCode = 7; $exitReason = "abort:$($match.rule.name)"
                break
            }

            # Crash action: AE itself died — the JSX can never write .done.
            # Dismiss the dialog (unblocks the dead process's shutdown so the
            # next launch isn't haunted by a zombie modal), then fail fast with
            # exit 8 instead of burning TimeoutSec. Cold-start crashes are
            # transient and heal on the caller's warm retry; a deterministic
            # crash fails the retry identically — same flake-vs-real split as
            # exit 2.
            if ($match.rule.action -eq 'Crash') {
                Write-ActionLog -DumpDir $dumpDir -Event 'crash-rule' -Data @{ name = $match.rule.name }
                if ($match.rule.keys) {
                    $r = Invoke-SendKeysSafe -Hwnd $m.Hwnd -Keys $match.rule.keys -DelayMs 200
                    Write-ActionLog -DumpDir $dumpDir -Event 'crash-dismiss' -Data @{ keys = $match.rule.keys; sent = $r.Sent }
                }
                [Console]::Error.WriteLine("AE CRASHED [$($match.rule.name)] — dialog dismissed; warm-retry once, identical failure = real crash")
                $exitCode = 8; $exitReason = "crash:$($match.rule.name)"
                break
            }

            $r = Invoke-SendKeysSafe -Hwnd $m.Hwnd -Keys $match.rule.keys -DelayMs 200
            if (-not $r.Sent) {
                Write-ActionLog -DumpDir $dumpDir -Event 'focus-mismatch' -Data @{
                    rule     = $match.rule.name
                    expected = ('0x{0:X}' -f [int64]$m.Hwnd)
                    actual   = ('0x{0:X}' -f [int64]$r.FocusActual)
                }
                continue
            }
            Write-ActionLog -DumpDir $dumpDir -Event 'sendkeys' -Data @{ keys = $match.rule.keys }
            Add-Cooldown -Cooldown $cooldown -Hwnd $m.Hwnd -Rule $match.rule.name -DurationMs $match.rule.cooldownMs
            break
        }

        if ($exitCode -ne 0) { break }   # Abort rule fired inside the foreach

        if ($persistentUnknownHwnd) {
            Write-ActionLog -DumpDir $dumpDir -Event 'unknown-modal-persistent' -Data @{
                hwnd = $persistentUnknownHwnd
                graceSec = $unknownGraceSec
            }
            $exitCode = 2; $exitReason = 'unknown-modal'; break
        }

        Start-Sleep -Milliseconds $TickMs
    }
} finally {
    if ($exitCode -ne 0) {
        $ocrTexts = @{}
        $finalModals = Get-AeModals -AeRootPid $aeRootPid -MainTitleHints @('Adobe After Effects')
        foreach ($m in $finalModals) {
            try {
                $bmp = Capture-WindowBitmap -Hwnd $m.Hwnd -Rect $m.Rect
                $ocrTexts[('0x{0:X}' -f [int64]$m.Hwnd)] = Invoke-Ocr -Bitmap $bmp
                $bmp.Dispose()
            } catch {}
        }
        Write-ForensicsDump -DumpDir $dumpDir -ExitCode $exitCode -Reason $exitReason `
                            -Modals $finalModals `
                            -OcrTexts $ocrTexts `
                            -Meta @{ aeExe = $AeExe; jsx = $Jsx; tickMs = $TickMs }

        # Unknown-modal triage straight to stderr: show what OCR saw + a rule
        # skeleton, so adding a rule doesn't require digging through the dump.
        if ($exitCode -eq 2) {
            [Console]::Error.WriteLine("=== unknown modal — OCR capture(s) ===")
            foreach ($k in $ocrTexts.Keys) {
                $txt = ($ocrTexts[$k] -replace "`r?`n", ' / ')
                [Console]::Error.WriteLine("  $k : $txt")
            }
            [Console]::Error.WriteLine("Add a rule to scripts/ae_dialog_rules.json (CJK patterns CONTIGUOUS — matcher strips OCR spaces):")
            [Console]::Error.WriteLine('  { "name": "<describe>", "windowTitle": [], "windowClass": [], "ocrMatch": ["<distinctive substring>"], "action": "SendKeys", "keys": "{ENTER}", "cooldownMs": 2000, "comment": "<why + which button this presses>" }')
            [Console]::Error.WriteLine("Full dump: $dumpDir")
        }
    }

    if ($aeProc -and -not $aeProc.HasExited) {
        # Post-run exit grace. AE's clean shutdown (prefs write + session
        # bookkeeping) routinely exceeds 5s on this machine; killing it
        # mid-shutdown sets the crash flag, so the NEXT launch shows the
        # 崩溃修复选项 (crash-repair) dialog — back-to-back gate runs then
        # cascade. Give app.quit() a generous window before force-kill.
        $waitDeadline = (Get-Date).AddSeconds(30)
        while (-not $aeProc.HasExited -and (Get-Date) -lt $waitDeadline) {
            Start-Sleep -Milliseconds 250
        }
        if (-not $aeProc.HasExited) {
            Stop-Process -Id $aeRootPid -Force -ErrorAction SilentlyContinue
            # Block until the kill lands (bounded): exiting while the corpse is
            # still unwinding makes the caller's immediate warm retry trip the
            # concurrent-AE guard (exit 6) — observed on the exit-8 crash path.
            try { [void]$aeProc.WaitForExit(15000) } catch {}
            Write-ActionLog -DumpDir $dumpDir -Event 'ae-force-killed'
        } else {
            Write-ActionLog -DumpDir $dumpDir -Event 'ae-exited'
        }
    } else {
        Write-ActionLog -DumpDir $dumpDir -Event 'ae-exited'
    }
}

exit $exitCode
