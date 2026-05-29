---
when_to_read: running AE ship-gates in an interactive session and hitting ae_run.ps1 exit 2 / unknown-modal; AE shows a dialog that OCR can't read; "崩溃修复选项" safe-mode dialog on launch
applies_to: [ship-gate, ae-automation, ae_run, ocr, debugging-tools]
last_updated: 2026-05-29
---

# AE ship-gate flakes in interactive sessions: occlusion + crash-recovery cascade

`ae_run.ps1` dispatches AE modals via Layer B (title/class) → Layer C (OCR). Two
failure modes recur when running ship-gates **while an interactive coding session
shares the desktop**, both surfacing as `exit 2` (unknown-modal):

## 1. OCR occlusion
`Capture-WindowBitmap` grabs the dialog **by screen rect** — if your editor/terminal
windows sit on top of where AE's dialog is, OCR reads YOUR screen, not the dialog,
so no rule matches → exit 2. The dialog IS there (window enumeration finds it); only
the pixel capture is wrong.

**Read the real dialog with `tmp_debug/capture_dialog.ps1`** — uses `PrintWindow`
(PW_RENDERFULLCONTENT) which renders the window's own content regardless of z-order
(occlusion-immune). Win32 `GetWindowText` on AE dialogs returns garbage ("O") because
AE uses owner-drawn controls — PrintWindow + read-the-image is the reliable path.

## 2. Crash-recovery cascade (self-inflicted)
A failed/timed-out ship-gate makes `ae_run.ps1` **force-kill** AE. The next AE launch
then shows **"崩溃修复选项 / Safe Mode"** (`ae-safe-mode-recovery` rule → ESC, but
OCR-occluded so it doesn't fire) → another exit 2. Repeated kills keep re-arming it.

**Break the cascade with `tmp_debug/clear_ae_crashstate.ps1`** before each gate:
launches AE with a quit-only jsx, foregrounds the dialog via Win32 + PostMessage
ENTER ("继续"), lets AE exit cleanly → clears the crash flag.

## Standard pre-gate ritual (interactive session)
```
Get-Process AfterFX* | Stop-Process -Force          # kill stragglers
pwsh clear_ae_crashstate.ps1 -AeExe <ae>            # clear safe-mode flag
$env:AE_SHIP_GATE=1; go test -run <Gate> ...        # then run; warm-retry on cold-start splash
```
Cold-start splash ("正在初始化 MediaCore") can exceed ae_run's 15 s unknown-grace →
exit 2; a warm retry passes. None of these are data problems — verify the actual AE
error via capture_dialog.ps1 before assuming the .aep is bad.

## Real data errors look different (don't confuse with the above)
- "项目文件似乎已损坏（跳过部分：N）" = structural/size bug → see ae2020-shape-ldta-164-corrupt.md
- "After Effects 已崩溃 (0::42)" on open = malformed chunk AE chokes on (e.g. from-scratch shape scaffolding)
- "文件数据缺失" = tdb4 says static but body has a keyframe LIST (or vice-versa)
