---
status: obsolete
since: 2026-05-27
last_updated: 2026-05-27
when_to_read: writing dispatch rules that ocrMatch Chinese / Japanese / Korean text via Windows.Media.Ocr; debugging "rule doesn't match even though OCR clearly captured the right phrase"; matching patterns against CJK OCR output
applies_to: [windows-media-ocr, cjk, ocr-matching, ae-run-wrapper, dispatch-rules]
---

# Windows.Media.Ocr separates CJK glyphs with spaces

## TL;DR

Windows.Media.Ocr returns `"修 复 选 项"` for what the dialog displays as `"修复选项"`. Substring match `"修复选项".IndexOf("修复选项")` fails. Dispatch rules in `scripts/ae_dialog_rules.json` keep CJK patterns **contiguous** (`"修复选项"`); `Match-Rule` in `scripts/AeRun.Lib.ps1` retries match on whitespace-stripped versions of both sides via `_matchAnyOcr`.

## How it bit

Task 14 smoke: AE 2025 popped a safe-mode recovery dialog after wrapper's prior force-kill. Added rule with `"ocrMatch": ["修复选项", "安全模式"]`. Wrapper still exit-2'd. Confusion: console output of OCR mojibake'd via cp936 looked vaguely like the rule pattern; assumed rule was matching but something else broken.

Dumping `ocr.txt` as Unicode codepoints revealed reality: OCR returned `U+5D29 U+0020 U+6E83 U+0020 U+4FEE U+0020 U+590D U+0020 ...` — every CJK glyph followed by a `U+0020` space.

Rule pattern `"修复选项"` is 4 contiguous CJK chars. OCR text contains `"修 复 选 项"` (7 chars including spaces). `String.IndexOf("修复选项")` against `"修 复 选 项"` returns -1.

## Why it happens

Windows.Media.Ocr's `OcrResult.Text` joins recognized words with spaces. For Latin scripts each "word" is a multi-char token (`"file data is missing"` → `"file data is missing"`). For CJK each glyph is its own "word" so they join with spaces (`"修复选项"` → `"修 复 选 项"`).

This is documented behaviour:
- `OcrResult.Lines[].Words[]` exposes individual word spans
- `OcrResult.Text` is `string.Join(" ", words)` per line
- For CJK, recognizer emits one glyph per Word — hence one space per glyph

## The fix in this repo

`Match-Rule` does layer-specific matching:
- **Title / Class** — direct `IndexOf` substring (no normalization; window titles are stable strings)
- **OCR** — `_matchAnyOcr` tries direct substring first (English rules: `"file data is missing"` still matches OCR `"file data is missing"`), falls back to substring on `($text -replace '\s+', '')` vs `($pattern -replace '\s+', '')` (CJK rules: `"修复选项"` matches OCR-stripped `"修复选项检测到..."`).

Rule authors write **patterns the way they appear in the UI** (`"修复选项"` not `"修 复 选 项"`); the matcher handles the OCR normalization invisibly.

## Test for this in Pester

`scripts/ae_run.Tests.ps1` has:
```powershell
It 'matches Chinese OCR with space-separated glyphs (Windows.Media.Ocr behaviour)' {
    $info = @{ Ocr = '错 误 : 文 件 数 据 丢 失 。' }
    Match-Rule -HwndInfo $info -Rules $script:rules
    # expects rule with ocrMatch=["文件数据丢失"] to hit
}
```

Don't delete this test — without it the CJK fallback regresses silently.

## Watch-out

- **Console output of OCR is mojibake** via Windows codepage 936/GBK trying to render UTF-8. The bytes are correct; the terminal lies. Read dumps with `Get-Content -Encoding UTF8` or `[System.IO.File]::ReadAllText(path, [Text.Encoding]::UTF8)`. Don't diagnose from console alone — dump to file and inspect codepoints if confused.
- **Don't strip whitespace at OCR time** (in Invoke-Ocr). Forensics dumps need the original spaced version to verify what OCR actually saw. The stripping is a match-time concern only.

## Related

- [[pwsh-7-no-winrt]] — paired OCR scar; why we sub-shell at all
- `scripts/ae_dialog_rules.json` — rule schema with `ocrMatch` field
- `specs/2026-05-27-ae-run-wrapper-design.md` §3.3 Layer C (OCR fallback)
