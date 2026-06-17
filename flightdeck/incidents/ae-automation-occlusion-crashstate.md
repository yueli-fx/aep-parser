---
status: active
when_to_read: running AE ship-gates in an interactive session and hitting ae_run.ps1 exit 2 / unknown-modal; AE shows a dialog that OCR can't read; "崩溃修复选项" safe-mode dialog on launch (including on MANUAL launch after automated runs); writing PowerShell that touches WinRT/UWP (Windows.Media.Ocr) or "Unable to find type" from pwsh; ocrMatch rules failing on CJK text
applies_to: [ship-gate, ae-automation, ae_run, ocr, debugging-tools, exit-grace, crash-flag, powershell, pwsh, winrt, uwp, windows-media-ocr, cjk, ocr-matching, dispatch-rules, operator-context, cold-start-splash, dont-ask-user, no-resave]
last_updated: 2026-06-17
---

# AE ship-gate flakes in interactive sessions: occlusion + crash-recovery cascade

## Operator context（2026-06-17，用户立规）

**用户几乎不在这台机器上操作（在另一台机器玩游戏）。** 故本机前台可能挂着一个
游戏窗口(英雄联盟等)占满屏幕,但那不是用户在用——AE 自动化抢前台/OCR 失败时，
**这是 agent 自己要解的问题，不是环境问题，更不要 AskUserQuestion 让用户"让出
机器/关游戏"**。正确反应 = 跑下方 §"Standard pre-gate ritual"(清 crash flag +
warm-up + 跑已知-good 对照确认环境，再 warm-retry)。批22 footage gate 首撞:
cold-start splash「正在初始化本地化设置」+ 前台挂游戏 → exit 2,我误判为用户机器
占用并发问；用户纠正后跑 clear_ae_crashstate + comp_idta 对照(13s 绿)→ footage
gate 双版本即过。教训:CLAUDE.md「AE ship-gate = agent 自跑…别默认让用户手开 AE;
cold-start exit-2 先 warm-retry + 跑已知-good 对照」是铁律,exit-2 别第一反应甩给
用户。**另**:用户「经常看你卡在另存为界面」——任何 verify jsx 必须纯 DOM
readback、`close(DO_NOT_SAVE)+quit`,**禁 resave**(弹 Save 框 = OCR 无规则 → hang;
comp_idta/footage_idta 均已去 resave)。

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

**FIXED in the wrapper 2026-06-11**: `Capture-WindowBitmap` (AeRun.Lib.ps1) now
PrintWindow-captures by Hwnd first and only falls back to the legacy screen-rect
copy — gate OCR works with an editor sitting on top (AddMask gate triage proved
it: the dump's "unknown modal" OCR used to be the IDE's file tree). Same session
also added: `Crash` action (exit 8, AE crash dialog `ae-crashed` rule — dismiss
the corpse + fail fast; harness warm-retries 1/2/8), teardown `WaitForExit` after
force-kill (immediate warm retry used to trip the exit-6 concurrent-AE guard),
and failure forensics moved from t.TempDir (wiped on teardown!) to
`tmp_debug/gate_fails/`.

## 2. Crash-recovery cascade (self-inflicted)
A failed/timed-out ship-gate makes `ae_run.ps1` **force-kill** AE. The next AE launch
then shows **"崩溃修复选项 / Safe Mode"** (`ae-safe-mode-recovery` rule → ESC, but
OCR-occluded so it doesn't fire) → another exit 2. Repeated kills keep re-arming it.

**Break the cascade with `tmp_debug/clear_ae_crashstate.ps1`** before each gate:
launches AE with a quit-only jsx, foregrounds the dialog via Win32 + PostMessage
ENTER ("继续"), lets AE exit cleanly → clears the crash flag.

### 2b. 普通成功 run 也会埋雷：post-done 退出宽限太短（2026-06-10）

不止 failed/timed-out run——**成功的 gate run 也曾留 crash flag**。`ae_run.ps1`
拿到 `.done` 后只等 AE 自退 **5 秒** 就 `Stop-Process -Force`；AE 的干净退出
（写首选项 + session 收尾）在本机经常超过 5 秒（AE 2025 尤甚）。于是连续
命令行 gate run = 每次都把 AE 杀在收尾半路 → crash flag 置位 → 下次启动
（包括用户手动开 AE）弹「崩溃修复选项」（以安全模式启动/重置首选项/管理增效
工具/继续）。自动化 run 里该对话框被 `ae-safe-mode-recovery` 规则 ESC 消化，
但用户手动开 AE 时会直接看到——首发现场即用户手动启动（2026-06-10）。

**修复**：`ae_run.ps1` post-done 宽限 5s → 30s（teardown 注释里有 WHY）+
`ae_dialog_rules.json` 新增 `preferences-damaged` 规则（「首选项文件无效或
已损坏」对话框 Enter 接受重建；**必须排在 `project-corrupt-skip` 之前**——
首选项文案也含「已损坏」，first-match-wins，Pester 有 ordering guard 用例）。
配套既有教训：JSX 末尾必须 `app.project.close(DO_NOT_SAVE) + app.quit()`
（见 re-fixture checklist），wrapper 的宽限只兜 quit 之后的收尾时间。
若再看到「崩溃修复选项」：选「继续」即可（不要重置首选项）；自动化侧跑
`tmp_debug/clear_ae_crashstate.ps1` 清 flag。

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

---

## OCR 后端两坑（原 `pwsh-7-no-winrt` + `windows-media-ocr-cjk-glyph-spacing`,2026-06-16 折入）

Layer C 的 OCR 后端(`scripts/ocr_helper.ps1` + `AeRun.Lib.ps1`)有两个独立 scar:

### Case OCR-1 — pwsh 7 丢了 WinRT 投影 → 用 powershell.exe 5.1 子壳
`Add-Type ... ContentType=WindowsRuntime` / `[Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime]` 在 **pwsh 7+(.NET 6+)静默失败**(报 `Unable to find type` —— 是"类型不存在"不是"引擎创建失败";.NET 6 无 WinRT 投影层)。Windows PowerShell 5.1(`powershell.exe`)仍有。**修法**:`ae_run.ps1`(pwsh 7 主)shell out 到 `powershell.exe -NoProfile -File scripts/ocr_helper.ps1`(5.1,WinRT 可用)。成本 ~400ms/次冷启,便宜(OCR 仅 Layer B 命中 modal 时跑)。未来:`Microsoft.Windows.SDK.NET.Ref` 理论上让 pwsh 7 直接 WinRT(未用,子壳更稳);跨平台脚本**别 import 此模式**(Windows-only)。

### Case OCR-2 — Windows.Media.Ocr 给 CJK 每字插空格
OCR 把 `"修复选项"` 返回成 `"修 复 选 项"`(`OcrResult.Text` = 按 word 空格 join;CJK 每字一个 word)→ `IndexOf("修复选项")` 失败。**修法**:`ae_dialog_rules.json` 里 CJK pattern 写**连续**(`"修复选项"`);`AeRun.Lib.ps1::Match-Rule` 的 `_matchAnyOcr` 先直配(英文),再 fallback 比 `($text -replace '\s+','')` vs `($pattern -replace '\s+','')`(CJK)。规则作者按 UI 原样写,matcher 透明归一。**Pester 回归**:`ae_run.Tests.ps1` 有 CJK 空格用例(别删,否则 CJK fallback 静默退化)。**坑中坑**:OCR console 输出经 cp936 渲染 UTF-8 是 mojibake(字节对、终端骗你)——存疑时 dump 文件按 UTF-8 读 codepoint,别从 console 诊断;forensics dump 保留原始带空格版(stripping 仅 match-time)。

> 相关:`specs/2026-05-27-ae-run-wrapper-design.md` §3.3/§3.5 · `checklists/re-fixture.md` § GDI 自动化 · [[jsx-state-leak]](JSX fixture 工作流,正交主题,未并)。
