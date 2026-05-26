---
status: draft
created: 2026-05-27
last_updated: 2026-05-27
owner: claude + 月离
related:
  - workshop/playbooks/re-fixture.md
  - workshop/scars/ae25-acceptance-gate.md
  - workshop/scars/project-flag-chunks-lnrb-lnrp.md
---

# `scripts/ae_run.ps1` — Unattended AE ship-gate wrapper

## 1. 背景

`workshop/playbooks/re-fixture.md` §"AE 打开 .aep 的 3 种失败模式" 列出 ship-gate 失败的三类：
1. AE 完全崩溃
2. AE 打开但 GUI 弹对话框（convert / save changes / 其它 modal）
3. AE 打开但 JSX `app.open()` throw "文件数据丢失"

Mode 1 和 Mode 3 当前流程能区分（看 `.done` 是否生成 + 内容）。**Mode 2 是阻塞点**：JSX 跑不到 `app.open` 返回，`.done` 不生成或迟到，Go 端 ship-gate test 走到 timeout，要人 RDP 进去手点对话框。

playbook 当前权宜是**版本匹配规则**——fixture 用哪版 AE 写的就用哪版 AE 跑，避开 convert 对话框。能解决跨版本问题但要求：
- 维护 fixture↔AE-版本对照（隐式 cost）
- 不能跨版本 ship-gate（reduces test surface）
- 残留风险：上次 JSX 退出不干净留下的 "save changes?" 对话框，下次启动直接弹

**目标**：写 `scripts/ae_run.ps1` 当做 `AfterFX -r` 的 drop-in 替换，**OCR 检测对话框 + dispatch table 自动消化**，让 ship-gate 真正 unattended，任意 AE × 任意 .aep 都能跑。

## 2. 范围

**In scope**
- 替代版本匹配规则（任意 AE 版本跑任意 fixture，convert / save / data-loss 对话框自动消化）
- 与现有 `runV2_2ShipGate` / `runV2_1ShipGate` / 后续 ship-gate test 的 drop-in 集成
- Dispatch table 可外置 (JSON) + 易扩展（新对话框→加一行规则）
- 失败时收集 forensics（screenshot + OCR text + window enum）

**Out of scope**
- 不接管 JSX 端的错误处理（JSX 写 `.done` 的契约不变）
- 不做 AE 进程崩溃恢复（Mode 1 仍按现流程报错）
- 不实现 Claude in-loop 回调（unknown dialog → fail-loud + dump，下次手工/Claude session 加规则）
- 不做 GUI 交互记录/回放工具（场外开发者 add 规则靠看 dump 文件）

## 3. 架构

### 3.1 接口契约

Go 端 ship-gate test 改一行：

```go
// Before:
// cmd := exec.Command(aeExe, "-r", jsxPath)

// After:
cmd := exec.Command("pwsh", "-NoProfile", "-File",
    filepath.Join(repoRoot, "scripts", "ae_run.ps1"),
    "-AeExe", aeExe,
    "-Jsx", jsxPath,
    "-Done", doneFile,
    "-Timeout", "120")
```

ps1 内部承担：
1. 启动 AfterFX 子进程
2. **并发**轮询：dialog dismissal loop + `.done` 文件存在检查
3. `.done` 出现 → 等 1s（JSX 写完）→ 干净退出 AE → ps1 exit 0
4. timeout / unknown dialog → 收集 dump → kill AE → ps1 exit 非 0

`.done` 文件协议**不变**（JSX 写 `OK <step>` / `ERR <step> -> <msg>` 多行），Go 端读 `.done` 第一行 == "PASS" 的逻辑零修改。

### 3.2 进程拓扑

```
Go test process
   │
   └─ exec.Command pwsh -File ae_run.ps1
        │
        ├─ Start-Process AfterFX.exe -r foo.jsx  ← AE GUI 进程
        │       │
        │       └─ JSX 执行 → 写 .done → app.quit
        │
        └─ while loop (in ps1 主线程):
             ├─ Test-Path $Done?  → return success
             ├─ Get-ForegroundDialog (Win32 enum, scope=AE proc tree)
             │       └─ if found → Screenshot → OCR → Match-Rule → SendKeys
             └─ Start-Sleep 500ms; 累计 timeout 检查
```

**不用** ps1 background job——主线程轮询足够（500ms cadence vs AE 启动 60s，CPU 可忽略）。

### 3.3 Dialog detection 三层过滤

OCR 在最后一层。前两层是稳定 identifier（title/class），第三层 OCR 兜底。

每个 loop tick：

**Layer A — Win32 windows enum（快，~ms）**
- 用 `Add-Type` 加 `EnumWindows` + `GetWindow(GW_OWNER)` + `GetClassName` + `GetWindowLong(GWL_EXSTYLE)` 一组 P/Invoke
- 候选过滤（按顺序短路）：
  1. `IsWindowVisible(hwnd)` 真
  2. `pid` 属于 AE root 进程树（见 §3.9）
  3. **任一**命中即认为是 modal：
     - 窗口标题不是 AE 主窗口标题（"Adobe After Effects" + 项目名）
     - `GW_OWNER` 指向 AE 主窗口（典型 modal owner 模式）
     - `WS_EX_DLGMODALFRAME` 位被设置
- 拿到 `HWND` + window rect + window title + window class

**Layer B — Title / Class match（稳定，免 OCR）**
- 用 hwnd 的 title / class 去 dispatch table 的 `windowTitle` / `windowClass` 字段比对（substring 列表，case-insensitive）
- 命中 → 进入 §3.8 SendKeys 安全协议；不命中 → Layer C

**Layer C — OCR fallback（慢，~200ms，仅触发时）**
- `[System.Drawing]::CopyFromScreen` 截 window rect
- Bitmap → `Windows.Graphics.Imaging.SoftwareBitmap`
- `Windows.Media.Ocr.OcrEngine.TryCreateFromUserProfileLanguages()` → `RecognizeAsync` → text
- text 跟 dispatch table 的 `ocrMatch` 字段比对（substring, case-insensitive）

**为什么 OCR 是 fallback 而不是 primary**：OCR 受 DPI / locale / 字体抗锯齿 / 中英混排影响有 false-positive 风险；title / class 是 Win32 给的稳定 identifier。但有些 modal title 是空的或跟主窗口同名，这时 OCR 兜底。

**为什么不每 tick 都做 Layer C**：AE 主时间线 / 项目面板有任意用户文字，全屏 OCR 会噪。Layer A 先确认"有 modal 弹出"再 OCR 它的窗口区域，避免误判。

**Layer A 阈值的鲁棒性**：初版用 title heuristic；如果跑挂遇到某个 modal **三个条件都不命中**（极少见），dump 里看到陌生 hwnd 时再补 detection 规则。Layer A 失败 → Layer B / C 也跑不到，wrapper 走 timeout 路径（exit 1） + dump，不会静默吞掉。

### 3.4 Dispatch table — multi-signal

文件：`scripts/ae_dialog_rules.json`

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
    "comment": "AE 高版本打开低版本 .aep — convert 对话框 title 不稳定，靠 OCR"
  },
  {
    "name": "save-changes-on-quit",
    "windowTitle": ["Adobe After Effects"],
    "windowClass": [],
    "ocrMatch": ["save changes", "保存更改"],
    "action": "SendKeys",
    "keys": "{TAB}{TAB}{ENTER}",
    "cooldownMs": 3000,
    "comment": "JSX 未 quit 干净留下的 save-changes — 走到\"不保存\"按钮 (Tab Tab Enter，按钮顺序保存/不保存/取消)。注：locale/版本敏感，§7 有 UIA 升级路径"
  },
  {
    "name": "file-data-missing",
    "windowTitle": [],
    "windowClass": [],
    "ocrMatch": ["file data is missing", "文件数据丢失"],
    "action": "SendKeys",
    "keys": "{ENTER}",
    "cooldownMs": 2000,
    "comment": "Mode 3 — fixture 已坏，按 OK 让 JSX 走 catch 块写 .done(ERR)"
  }
]
```

**Schema 字段语义**：
- `windowTitle` / `windowClass`：substring 列表，任一命中算 match（**首选**，稳）
- `ocrMatch`：substring 列表，任一命中算 match（**fallback**，OCR）
- 一条 rule 三个数组**任一非空**就启用对应 layer；都为空 = 错误配置（启动时 reject）
- 优先级：title → class → ocr（按 §3.3 三层）
- `cooldownMs`：同一 hwnd + 同一 rule 在 cooldown 期内不重复触发（防止 dialog 消失前重复按键）

规则按数组顺序遍历，第一个命中即执行。

**初版只填上表三条** — 别预填没见过的对话框，避免错误规则误伤 fixture。后续跑挂遇到新对话框 → 看 dump 文件 → 加规则 + commit。

**Title / class 留空的原因**：convert 对话框跨 AE 版本 title 文字会变，class 也未必稳定。初版三条规则**都走 OCR fallback** 验证流程，等真实跑出来知道稳定的 title/class 再填进去（这是为什么 multi-signal schema 在初版就要立起来——后补不会改 schema 只改 data）。

### 3.5 OCR 后端

**Windows.Media.Ocr** (Windows 10/11 内置 UWP API)：
- 零安装、可离线、native 中英文支持（用 user profile 语言）
- PowerShell 通过 WinRT projection 调用（`Add-Type -AssemblyName System.Runtime.WindowsRuntime` + helper for `IAsyncOperation` → Task）

参考代码骨架（spec 不是 plan，不展开完整实现）：
```powershell
[Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime] | Out-Null
$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
$bmp = New-Object System.Drawing.Bitmap $w, $h
# ... CopyFromScreen → SoftwareBitmap → engine.RecognizeAsync ...
```

Fallback：如果 Win10 build < 17134（OCR API 缺失），ps1 立即 fail loud 提示装 Tesseract。**初版不实现 Tesseract fallback** — 等真碰到才加。

### 3.6 失败处理 & forensics

| 触发 | 行为 |
|---|---|
| timeout（默认 120s） | screenshot 全屏 + OCR + window enum dump → kill AE → exit 1 |
| Unknown dialog（OCR 命中但没规则） | 同上 + 多记 OCR 命中的 modal 窗口 ID → 按 ESC（保守，不破坏数据）→ exit 2 |
| OCR engine 初始化失败 | exit 3 + 提示 Windows 版本要求 |
| AE 启动失败（5s 内没进程） | exit 4 |

Dump 目录：`<doneFile>.fail/`（next to .done file，Go test cleanup 时一起删）：
- `screenshot.png` — 全屏 PNG
- `ocr.txt` — 当前所有可见窗口的 OCR 结果
- `windows.txt` — `EnumWindows` 输出（title + class + ex-style + owner + rect + pid）
- `actions.log` — wrapper 整个生命周期的 action 序列（时间戳 + 事件类型 + 参数），例：
  ```
  12:00:01.230  ae-start                 pid=12345
  12:00:08.140  detect-modal             hwnd=0x002A1B title="" class="#32770" owner=main
  12:00:08.140  rule-match-ocr           name=convert-old-project text="Convert this project from..."
  12:00:08.140  focus-protocol-ok        hwnd=0x002A1B latency=180ms
  12:00:08.420  sendkeys                 keys={ENTER}
  12:00:08.420  cooldown-set             rule=convert-old-project hwnd=0x002A1B until=12:00:10.420
  12:00:09.870  detect-modal-gone        hwnd=0x002A1B
  12:00:42.110  done-found
  12:00:42.612  done-stable              size=412
  12:00:43.330  ae-exited
  ```
  事后 debug 价值极高——焦点 race / 重复触发 / OCR 误判都能 timestamp 复盘。
- `meta.json` — `{exitCode, reason, timestamp, aeVersion, totalTicks, ocrInvocations, rulesMatched, ...}`

Go 端 ship-gate test 抓到 ps1 非 0 退出时：
- 打印 dump 路径
- `t.Logf` 把 `meta.json` 内容贴出来
- 不 auto-upload —— 留本地让人 / 下次 Claude session 看

### 3.7 退出 AE 的兜底

`.done` 出现后**不立即**收尾——可能 JSX 还在 flush 文件或调 `app.quit`。流程：

1. **File-stable poll** `.done`：连续 500ms 内 `(size, mtime)` 没变化 → 认为 JSX 写完。
   超时上限 5s（避免 JSX 卡住时无限等）。
2. **等 AE 优雅退出**：poll `Get-Process AfterFX`，每 500ms 检查一次，最多 5s。
3. 仍存活 → `Stop-Process -Force AfterFX`
4. 验证 `tasklist` 无 `AfterFX.exe`
5. exit 0

这是"双保险"——JSX 末尾的 `app.project.close + app.quit` 是首选，ps1 force-kill 是兜底。两者顺序不能反（先等优雅退出再 force kill）。**改用 file-stable 而不是固定 sleep** 是因为 .done 出现 ≠ 文件 flush 完，race 有过先例。

### 3.8 SendKeys 安全协议

GUI automation 最大风险**不是 OCR 错配，而是 SendKeys 焦点漂移**——AE splash / plugin window / RDP 焦点抢占都会把按键发给错窗口（发到 timeline / preview / project panel 而不是 dialog）。

每次 SendKeys 必须三步前置：

1. `SetForegroundWindow(hwnd)` —— 强制把对话框拉到前台
2. `Start-Sleep -Milliseconds 200` —— 等 OS 完成激活（实测 100~300ms 范围；初版 200，跑挂调）
3. `GetForegroundWindow() == hwnd` 校验 —— 不等 = SendKeys 取消，本 tick 走 fail path 不发键

**校验失败的处理**：不重试不发键，记 actions.log 一条 `focus-mismatch <ruleName> <expected_hwnd> <actual_hwnd>` 后跳过本 tick；下一 tick 重新走 detection。AE 焦点 race 通常几百 ms 就稳定，**不在 hot loop 里硬抢焦点**避免把别的窗口顶下去。

**Cooldown**：发完键之后，`(hwnd, ruleName)` 进 cooldown set，`cooldownMs` 过后才允许再次匹配。期间 detection 仍跑（hwnd 是否消失能感知），只是同 rule 不二次触发。

**Single-threaded 注**：detection / OCR / SendKeys 全在 ps1 主线程串行跑。500ms cadence vs OCR ~200ms vs SendKeys ~250ms（含 focus protocol），单 tick 总开销 < 600ms，AE 启动 60s 的 idle 窗口能轻松吞下；**不引入 PowerShell job / runspace** 避免 race。

### 3.9 进程树范围

§3.3 Layer A 第 2 条说"pid 属于 AE root 进程树"。明确定义：

- **AE root** = ps1 `Start-Process AfterFX.exe` 拿到的 `.Id`
- 候选 modal pid **必须等于** AE root（不收 child）
- AE 的 helper（`CEPHtmlEngine` / `dynamiclinkmanager` / 崩溃汇报）pid ≠ root，自动被过滤
- 不用 `Get-CimInstance Win32_Process` 爬 parent 链——AE root pid 我们启动时就知道，直接比对最便宜

## 4. 数据流

```
Go ship-gate test
  │
  │ exec pwsh -File ae_run.ps1 -AeExe ... -Jsx ... -Done ... -Timeout 120
  ▼
ps1 主循环:
  ├─ launch AfterFX (Start-Process) → $aeRootPid
  ├─ Load rules.json + init cooldown set + open actions.log
  └─ tick loop (every 500ms):
       1. Test-Path $Done ?
            yes → file-stable poll (§3.7) → cleanup AE → exit 0
       2. Layer A — EnumWindows filter (pid==$aeRootPid + visible + modal-heuristic) → $hwnd
            none → goto tick++
       3. (hwnd, *) in cooldown? → skip; goto tick++
       4. Layer B — match $hwnd.title / .class vs rules.windowTitle / .windowClass
            matched → goto step 6
       5. Layer C — Screenshot $hwnd → OCR → match vs rules.ocrMatch
            matched → goto step 6
            unmatched → dump-and-exit 2
       6. SendKeys safety protocol (§3.8):
            SetForegroundWindow → sleep 200ms → verify GetForegroundWindow
            fail → log focus-mismatch → goto tick++
            ok → SendKeys $rule.keys → set cooldown (hwnd, rule) → goto tick++
       7. elapsed > Timeout → dump-and-exit 1
  ▼
Go ship-gate test:
  ├─ ps1 exit 0 → read .done → assert PASS
  └─ ps1 exit ≠ 0 → log dump path + meta.json → t.Fatalf
```

## 5. 错误模式 & 测试

### 5.1 单元 (offline，不需要 AE)

ps1 模块化拆分让以下能在不启动 AE 时测：

| 函数 | 测试方法 |
|---|---|
| `Match-Rule $hwndInfo $rules` (Layer B/C 合一) | feed mock title/class/ocr → assert returned rule.name + matched-layer |
| `Parse-Rules $jsonPath` | feed malformed JSON / 三个 match 数组都空 → assert reject |
| `Test-IsModal $hwnd $aeRootPid` | mock with self pwsh window (own pid, no owner) → assert filtered out |
| `Invoke-Cooldown $set $key $ms` | add → check → wait → re-check expired |
| `Test-FileStable $path` | mock file 改 size → assert false；停止改 → assert true |

PowerShell Pester tests 放 `scripts/ae_run.Tests.ps1`，CI 跑（不需要 AE）。

### 5.2 集成（需要 AE）

新增 `internal/aep/ae_run_wrapper_test.go`，gated by `AE_SHIP_GATE=1`：

| 场景 | fixture | expected |
|---|---|---|
| Clean run（无对话框） | 写 `re_template.aep` + 用 AE 2025 跑 | exit 0, .done 内容 PASS |
| Convert dialog | 用 AE 2025 跑 `re_cameralight.aep` (AE 2020 fixture) | wrapper 自动按 Enter, exit 0 |
| Save-changes leftover | 故意 JSX 不写 quit → 跑两次 | 第二次启动 wrapper 自动消化, exit 0 |
| Unknown dialog | mock 一个改名后的对话框 | exit 2 + dump 文件齐 |
| Timeout | 给 5s timeout 跑 60s fixture | exit 1 + dump 文件齐 |

### 5.3 现有 ship-gate test 改造

`runV2_2ShipGate` 等已有函数改 `exec.Command` 那一行（drop-in）。**保留 fallback path**：环境变量 `AE_RUN_BARE=1` 时走原来的裸 `AfterFX -r`（调试 ps1 时用）。

## 6. 实现顺序（给 writing-plans 的种子）

1. ps1 骨架：参数解析、launch AE、wait .done、cleanup。**先不接 OCR**，验证 drop-in 流程跑通
2. 加 Win32 EnumWindows 过滤
3. 加 Windows.Media.Ocr 调用 + screenshot
4. 加 dispatch table JSON 加载 + match logic
5. 加失败 dump
6. Pester 单元测
7. 改一个 ship-gate test 验证 drop-in（建议 `shape_layer_shipgate_test.go`）
8. 把所有 ship-gate test 改完
9. 更新 `playbooks/re-fixture.md` § GDI 自动化（删 planned 标记 + 加用法）

## 7. Open questions（spec 不锁，给 plan / impl 阶段定）

- **Dispatch table 中文匹配语言**：用户 AE 是中文还是英文 UI？初版规则两边都填能 cover，但 OCR 引擎语言列表（`OcrEngine.AvailableRecognizerLanguages`）跟 user profile 走，要确认机器上 `zh-CN` + `en-US` 都装了。
- **SendKeys → UIAutomation 升级路径**：初版用 SendKeys + §3.8 focus protocol。Schema 已支持 `action` 字段，**未来增加 `"action": "UIAInvoke"` + `"buttonName": "Don't Save"` 不改 schema**，只加新 action handler。`save-changes-on-quit` 这条 `Tab Tab Enter` 跨 locale 最脆，是第一个候选迁移规则。
- **AE 中文 UI 的"不保存"按钮键序**：spec 写的是 `Tab Tab Enter`（保存 / 不保存 / 取消），但 AE 不同 locale 顺序可能不同；首次跑要人工确认。失败时 actions.log + screenshot 能快速诊断。
- **Layer A modal heuristic 失败的 fallback**：§3.3 三个条件（title-not-main / owner / WS_EX_DLGMODALFRAME）任一命中即认 modal。如果跑挂遇到一个都不命中的 modal，看 windows.txt 找它的 distinguishing feature 加进 detection 层（schema 已留 `windowClass` 字段，可能演变成"任意未知 visible window of AE root pid 都当 modal 候选"的更激进策略）。

## 8. 验收

| 项 | 通过条件 |
|---|---|
| Drop-in 替换 | `runV2_2ShipGate` 改 exec.Command 后 PASS（AE 2025, clean fixture） |
| 跨版本 ship-gate | `runV2_1ShipGate` 用 AE 2025 跑 AE 2020 fixture, wrapper 消化 convert dialog, PASS |
| Save-changes leftover 防御 | 故意制造 leftover, 下次 wrapper 自动消化, PASS |
| Unknown dialog 失败有据 | 跑 mock 未知对话框, dump 文件齐, Go test 看到清晰错误 |
| Pester 单元 | `scripts/ae_run.Tests.ps1` 全 PASS, 不需要 AE |
| 文档同步 | `playbooks/re-fixture.md` GDI 段落更新；`board.md` 归档 |
