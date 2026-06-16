---
status: active
since: 2026-05-27
last_updated: 2026-05-27
when_to_read: writing PowerShell that touches UWP / WinRT APIs (Windows.Media.Ocr, Windows.Graphics.*, Windows.Storage.*); debugging "Unable to find type" errors when loading WinRT classes from pwsh; choosing PowerShell version for Windows automation in this repo
applies_to: [powershell, pwsh, winrt, uwp, windows-media-ocr, ae-run-wrapper, gdi-automation]
---

# pwsh 7 dropped WinRT projection — use powershell.exe 5.1 sub-shell

## TL;DR

`Add-Type ... ContentType=WindowsRuntime` and `[Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime]` **silently fail in pwsh 7+** (.NET 6+). Windows PowerShell 5.1 (`powershell.exe`) still has the WinRT projection. In this repo, `scripts/ae_run.ps1` is pwsh 7 but shells out to `powershell.exe -File scripts/ocr_helper.ps1 ...` for OCR.

## The trap

You write what looks like canonical WinRT-in-PowerShell:

```powershell
[void][Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime]
$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
```

In pwsh 7:
```
Unable to find type [Windows.Media.Ocr.OcrEngine,Windows.Foundation, ContentType=WindowsRuntime].
Unable to find type [System.WindowsRuntimeSystemExtensions].
```

In `powershell.exe` (5.1) with `Add-Type -AssemblyName System.Runtime.WindowsRuntime`:
```
✓ ocrEngine type: OK
✓ AsTask MI: OK
✓ engine: OK lang=zh-Hans-CN
```

The PS 7 error is **not "engine creation failed"** — it's "type doesn't exist." pwsh 7's .NET 6 runtime doesn't have the WinRT projection layer that .NET Framework 4.5+ had.

## Why it bit

`scripts/ae_run.ps1` Initialize-Ocr originally tried `[void][Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime]` directly. Silently caught + returned `$null` from the try/catch, leading to "OCR backend unavailable" with no clear pointer. Wasted ~30 min checking language packs / Windows build before realizing the projection layer was missing entirely.

## The fix in this repo

Two-file split:
- `scripts/ae_run.ps1` — pwsh 7 main (script syntax modernity, ternary, etc.)
- `scripts/ocr_helper.ps1` — runs under powershell.exe 5.1 (WinRT works), takes `-ImagePath` + `-OutFile`, writes UTF-8 OCR text

Main wrapper:
```powershell
& powershell.exe -NoProfile -File $script:_ocrHelperPath -ImagePath $tmp -OutFile $out 2>&1 | Out-Null
```

Cost: ~400ms per OCR call (powershell.exe cold start + WinRT init). Cheap because OCR only runs when Layer A enum finds a modal (i.e. rarely on the happy path).

## Future-proofing notes

- `Microsoft.Windows.SDK.NET.Ref` NuGet provides projected types for .NET 6+ in theory; would let pwsh 7 do WinRT directly. Not used here because the sub-shell pattern is simpler + battle-tested.
- WinRT-out-of-process (PS 7.4+ might add it) would obsolete the sub-shell. Check `$PSVersionTable.PSVersion` if upgrading; redo the OCR Tests.ps1 smoke to confirm.
- If shipping cross-platform PowerShell scripts in this repo, **don't import this scar's pattern** — the sub-shell is Windows-only. Only the GDI automation wrapper needs it.

## Related

- `checklists/re-fixture.md` § "GDI 自动化 — scripts/ae_run.ps1" (Task 18 pending)
- `specs/2026-05-27-ae-run-wrapper-design.md` §3.5 "OCR 后端"
- [[windows-media-ocr-cjk-glyph-spacing]] — second OCR scar; CJK chars come back space-separated
