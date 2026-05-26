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
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$AeExe,
    [Parameter(Mandatory)][string]$Jsx,
    [Parameter(Mandatory)][string]$Done,
    [int]$TimeoutSec  = 180,
    [string]$RulesPath = (Join-Path $PSScriptRoot 'ae_dialog_rules.json'),
    [int]$TickMs      = 500
)

. $PSScriptRoot/AeRun.Lib.ps1

$ErrorActionPreference = 'Stop'
$dumpDir = "$Done.fail"

Remove-Item -LiteralPath $Done -ErrorAction SilentlyContinue

try { $rules = Parse-Rules -Path $RulesPath }
catch { [Console]::Error.WriteLine("rules load failed: $_"); exit 5 }

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

        $handledOrCooling = $false
        $anyUnknown = $false

        foreach ($m in $modals) {
            # Layer B — title/class
            $info = @{ Title = $m.Title; Class = $m.Class; Ocr = '' }
            $match = Match-Rule -HwndInfo $info -Rules $rules
            $usedOcr = $false

            # Layer C — OCR if title/class miss
            if (-not $match) {
                $bmp = Capture-WindowBitmap -Rect $m.Rect
                try { $ocrText = Invoke-Ocr -Bitmap $bmp } finally { $bmp.Dispose() }
                $info.Ocr = $ocrText
                $match = Match-Rule -HwndInfo $info -Rules $rules
                $usedOcr = $true
            }

            if (-not $match) {
                $anyUnknown = $true
                Write-ActionLog -DumpDir $dumpDir -Event 'unknown-modal' -Data @{
                    hwnd = ('0x{0:X}' -f [int64]$m.Hwnd)
                    title = $m.Title; class = $m.Class
                    ocr = if ($usedOcr) { ($info.Ocr -replace "`n", ' / ') } else { '<not-attempted>' }
                }
                continue
            }

            # cooldown?
            if (Test-InCooldown -Cooldown $cooldown -Hwnd $m.Hwnd -Rule $match.rule.name) {
                $handledOrCooling = $true
                continue
            }

            Write-ActionLog -DumpDir $dumpDir -Event 'detect-modal' -Data @{
                hwnd  = ('0x{0:X}' -f [int64]$m.Hwnd)
                title = $m.Title; class = $m.Class
            }
            Write-ActionLog -DumpDir $dumpDir -Event "rule-match-$($match.layer)" -Data @{ name = $match.rule.name }

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
            $handledOrCooling = $true
            break
        }

        if ($anyUnknown -and -not $handledOrCooling) {
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
                $bmp = Capture-WindowBitmap -Rect $m.Rect
                $ocrTexts[('0x{0:X}' -f [int64]$m.Hwnd)] = Invoke-Ocr -Bitmap $bmp
                $bmp.Dispose()
            } catch {}
        }
        Write-ForensicsDump -DumpDir $dumpDir -ExitCode $exitCode -Reason $exitReason `
                            -Modals $finalModals `
                            -OcrTexts $ocrTexts `
                            -Meta @{ aeExe = $AeExe; jsx = $Jsx; tickMs = $TickMs }
    }

    if ($aeProc -and -not $aeProc.HasExited) {
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
