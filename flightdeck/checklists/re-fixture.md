---
status: active
when_to_read: writing a new RE JSX fixture; debugging field locations via byte-diff against AE-saved baseline; setting up cross-version AE comparison; running a ship-gate against AE (modified .aep accepted/rejected); diagnosing why AE rejects a builder-written file; looking up "which AE version introduced field X"; invoking ae_run.ps1 wrapper for unattended ship-gate; deciding which AE version to use for a new fixture
applies_to: [jsx, re-workflow, fixture, ae-cli, byte-diff, baseline-strategy, ship-gate, ae-acceptance, version-mismatch, failure-modes, types-for-adobe, ae-version-introduced, ae-run-wrapper, gdi-automation, ocr-dispatch, agent-runs-ae, version-choice]
last_updated: 2026-05-28
---

# JSX RE 工作流 + ship-gate

## 总览

写 fixture 用 ExtendScript（JSX）→ 让 AE 跑一次 → 拿到 `.aep` → Go 端 parse_btdk diff 找字段。Ship-gate 反过来：Go 写改过的 `.aep` → AE 打开 → 读出值跟 Go 写入比对。

## Agent 自己调 AE — 不要 ask user

`scripts/ae_run.ps1` 是 **unattended wrapper**（OCR + multi-signal modal 自动消化，exit code 0/1/2/3/4/5）。Agent 应直接通过 Bash/PowerShell tool 调它，**不要**把 "请 user 跑 AE" 当默认。

调用模板（用 `$env:` 传 JSX 内 `$.getenv()` 读到的 mode/参数）：

```powershell
$env:MY_VAR = "value"   # 如果 JSX 用 $.getenv() 读 mode
Remove-Item -ErrorAction SilentlyContinue test_data/re_X.done
pwsh -NoProfile -File scripts/ae_run.ps1 `
    -AeExe "E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe" `
    -Jsx   "E:\projects\tools\aep-parser\test_data\re_X.jsx" `
    -Done  "E:\projects\tools\aep-parser\test_data\re_X.done" `
    -TimeoutSec 180
```

每次跑前先删旧 `.done`（wrapper 等的是新 `.done` 出现）。跑完读 `cat test_data/re_X.done` 看 JSX 步骤 log。

何时**应该** ask user：(a) AE 没装在标准路径需要他指路；(b) wrapper 反复 exit 2（未知 modal），rules 加完仍 fail；(c) 跨 monitor 输出 hang 等 GUI 异常。其他情况一律自己跑。

## 新 fixture 默认走 AE 2020

项目读取下限 = AE 2020（CLAUDE.md "读取下限 AE 2020"）。**新 RE fixture 默认用 AE 2020 跑**，理由：

1. AE 2020 = 项目语义基线，behavior 跟低版本读路径一致
2. AE 24+ 引入的字段（`fontCapsOption / strokeOverFill / autoHyphenate / trackMatteLayer / setTrackMatte` etc）在 AE 2020 跑会 throw 或 no-op，反而暴露不出来
3. AE 2020 不会"自动补"高版本字段（fresh save 干净），适合 byte-diff 找字段

**例外 — 升 AE 23+ 当且仅当**：被 RE 的字段是 AE 23+ 引入的。判断方法：diff `Types-for-Adobe/AfterEffects/22.0/` 跟 `Types-for-Adobe/AfterEffects/23.0/` TS 类型定义。引入版本对照：

| Field | 引入版本 | 旧版本回退 |
|---|---|---|
| `Layer.trackMatteLayer` / `Layer.setTrackMatte()` 显式 ID | AE 23 | implicit "layer above" (`l.trackMatteType = ALPHA`) |
| `TextDocument.fontCapsOption / strokeOverFill / fauxBold` etc | AE 24 | 不能 RE，必须 AE 24+ |
| `TextDocument.autoHyphenate / firstLineIndent` etc | AE 24 | 同上 |

混合 fixture 怎么办：跑大部分 mode 用 AE 2020，仅高版本字段那个 mode 单独切 AE 25。例：`re_delete_layer.jsx` baseline/middle/parent 用 AE 2020，matte mode 切 AE 2025（`TrackMatteLayerID` ldta @0xA0 是 AE 23+ 才有）。

## RE fixture 双轨

`test_data/` 下两类 fixture 文件名：

- `re_*.jsx` — AE 2020 ScriptingAPI 写的 fixture。**默认走这个**，覆盖 AE 2020 共有字段
- `re_*_ae24.jsx` — AE 24+ 才解封的字段（`fontCapsOption / strokeOverFill / autoHyphenate` 等），AE 2020 写不出来，必须 AE 24+ 跑

Test 代码 `t.Skipf("fixture missing", ...)` 缺文件跳过，不阻塞 CI。

## RE fixture 双轨

`test_data/` 下两类 fixture 文件名：

- `re_*.jsx` — AE 2020 ScriptingAPI 写的 fixture。**默认走这个**，覆盖 AE 2020 共有字段
- `re_*_ae24.jsx` — AE 24+ 才解封的字段（`fontCapsOption / strokeOverFill / autoHyphenate` 等），AE 2020 写不出来，必须 AE 24+ 跑

Test 代码 `t.Skipf("fixture missing", ...)` 缺文件跳过，不阻塞 CI。

## 参考资料常驻仓内

- `after-effects-scripting-guide/` — docsforadobe ScriptingAPI 镜像
- `Types-for-Adobe/AfterEffects/{8.0..26.0}/` — **判断"某字段在哪个 AE 版本引入"看这里最准**。按 AE 版本目录的 TS 类型定义，diff 两个版本目录就能看新字段引入点

## 何时启用 ship-gate

任何**结构性 / chunk-level 改动**入 main 前必须跑：

- 加新 chunk / 删 chunk / 改 chunk size（length-variable splice 含 Utf8 / expression / 字体名）
- Project-level setting chunk（lnrb/lnrp 等 toggle / acer/adfr/dwga 类单字段）
- NewProject / NewComposition / NewShapeLayer 等结构性创建
- 跨 AE 版本字段（AE 24+ 解封 / ldta 长度变化）

length-preserving 单字段（cdta 单 offset 改 / ldta flag bit 改）roundtrip Go test 够，不必走 ship-gate。

## Adobe 软件路径

- AE 2025 — `E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe`
- AE 2023 — `E:\adobe\Adobe After Effects 2023\Support Files\AfterFX.exe`
- AE 2022 — `E:\adobe\Adobe After Effects 2022\Support Files\AfterFX.exe`
- AE 2020 — `E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe`

跨版本对比时切对应 AE。Wave 1-3 字段默认用 AE 2025（24+ ScriptingAPI 解封）。

**版本匹配规则（建议，非强制）**：

- 用 **fixture 源版本** 的 AE 打开 (检查方式: `Application.Version()` 读 head 解出 e.g. `17.7x45` = AE 17.7 = AE 2020)
- 例：`re_cameralight.aep` 是 AE 2020 写的 → 优先用 AE 2020 跑（避无意义 convert 流程）
- 反向（高版本 fixture → 低版本 AE）经常崩，仍**避免**
- 跨版本测：`scripts/ae_run.ps1` wrapper 自动消化 convert 对话框，不再被 GUI 阻塞 unattended run（详 § GDI 自动化）

## AE 打开 .aep 的 3 种失败模式

调 AE-side ship-gate 不能盲信单一信号，先**辨别模式**：

| 模式 | 信号 | 处理 |
|---|---|---|
| **1. 完全崩溃** | `tasklist`/Get-Process 看不到 AfterFX.exe；没有 `.done`；exit code 非 0 | builder 写的字段触发 AE 内部 sanity-check fail (e.g. cdta timing 空)。看 `incidents/ae25-acceptance-gate.md` Stage 1 / 4 类 |
| **2. 打开但需转换** | GUI 弹 "Convert?" 对话框 → JSX 跑不到 `app.open` 返回，要么 catch 到 error，要么 hang。`.done` 含 ERR 信息（或根本写不出） | 用 `scripts/ae_run.ps1` wrapper 自动消化 convert 对话框；裸 `AfterFX -r` 仍需版本匹配 |
| **3. 打开但报数据损坏** | JSX 跑通；`app.open(...)` 在 try/catch 里 throw "After Effects 错误: 文件数据丢失" 类错误字符串 | builder chunk 写法 / 位置 / 大小破坏 AE 检查。这是最常见且最有 RE 价值的 — bisect 隔离哪个 setter 触发 |

诊断流程（按这个 order）：

1. AE process 是否仍 alive (`tasklist /v | grep -i afterfx`)
2. `.done` 是否生成（生成 = mode 3 / OK；不生成 = mode 1 / 2）
3. 看 `.done` 里 step error 信息：开头 fresh/open OK 但 open_modified ERR = mode 3；纯 fresh ERR = mode 1/2

**不要直接归 mode 3** 去 RE 字节，先排除 mode 1 / 2。

## GDI 自动化 — `scripts/ae_run.ps1`

Ship-gate 用 `scripts/ae_run.ps1` 替代裸 `AfterFX -r`：OCR + multi-signal dispatch 自动消化 convert / save / data-loss modal。任意 AE 版本 × 任意 fixture 都能 unattended 跑。**Cross-version smoke PASS**：AE 2020 fixture 用 AE 2025 跑，convert 对话框被自动消化（2026-05-27 验证）。

**调用契约**：

```powershell
pwsh -NoProfile -File scripts/ae_run.ps1 `
    -AeExe   $aeExe `
    -Jsx     $jsxPath `
    -Done    $doneFile `
    -TimeoutSec 180
```

退出码：

- `0` — `.done` 出现且 stable，按 PASS contract 走
- `1` — timeout
- `2` — 未知 modal（OCR 命中但没规则）
- `3` — OCR engine init 失败
- `4` — AE 进程启动失败

失败时 dump 落 `<doneFile>.fail/`：`screenshot.png` / `ocr.txt` / `windows.txt` / `actions.log` / `meta.json`。

**Go 端 ship-gate**用 `runAeRunShipGate(t, aeExe, jsxPath, doneFile, timeoutSec)` helper (`internal/aep/ship_gate_helpers_test.go`)，不要再直接 `exec.Command(aeExe, "-r", ...)`。

**⚠️ verify JSX `.done` 文件名必须 per-version 唯一**：ship-gate 同一 mode 跨 AE 2020/2025 跑两遍时，verify JSX 若把 `.done` 文件名只按 mode 命名，会(a)第二版覆盖第一版结果、(b)`ae_run.ps1 -Done` 等的是带 version-tag 的名 → 永远等不到 → 每次空等满 `TimeoutSec` 报 exit 1（实际验证早已跑通，假阴性）。修法：JSX 读一个 `$.getenv("..._TAG")`（如 `ae2020_basic`）拼进 `.done` 名，调用方 `-Done` 传同名。见 `verify_ge_insert_layer.jsx` + 2026-05-29 logbook。

**新对话框出现的流程**：

1. ship-gate FAIL, exit code 2
2. 看 `<doneFile>.fail/screenshot.png` + `ocr.txt`
3. 加规则到 `scripts/ae_dialog_rules.json`（substring 进 `ocrMatch`，或抄稳定 `windowTitle` / `windowClass`）
4. `Parse-Rules` 跑通 → re-run ship-gate

## 工作流

```
1. 写 test_data/re_<feature>.jsx
2. pwsh -File scripts/ae_run.ps1 -AeExe ... -Jsx ... -Done ...  (unattended)
3. JSX 末尾写 .done marker 标记完成
4. Go 端读 .aep → parse_btdk dump → 看新字段在哪
5. 写 test_data/re_<feature>_*_test.go 用 fixture 测试
```

裸 `AfterFX -r` 仍可手动用做 quick RE，但 ship-gate / 自动化场景一律走 wrapper。

## JSX 模板（关键防陷阱）

```js
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_<name>.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    // ⚠️ 必须！防止 -r 多轮 RE 累积 duplicate comps
    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    // ... 真正的 RE 步骤 ...

    step("save", function () { app.project.save(); });

    // .done marker 让 Go 端知道 RE 完了
    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_<name>.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    // ⚠️ 必须！tear down 否则 AE 留 "未保存改动" 状态，下次任何操作弹框
    // 烦死人 + 阻塞 unattended ship-gate run。顺序：先 close 再 quit。
    // 各裹 try/catch 防 quit 路径失败影响 .done 写入。
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
```

## AE 退出不弹框（关键陷阱）

`AfterFX.exe -r foo.jsx` 跑完 JSX 后 AE **不会自动退**。如果 JSX 用 `app.open(...)`
打开过文件 / 用 `app.newProject()` 起新 project，AE 会觉得 "用户打开了东西可能
要保存"，留在 GUI 等待。下次任何启动 / 用户切换/ 关 AE 都弹 "save changes?" 框。
对自动化是阻塞的（Go 端 ship-gate 没法 unattended 跑）；对人是骚扰（必须手点
"不保存"）。

**修法**：JSX 末尾（写完 .done 后）显式两步收尾，**顺序不能反**：

```js
try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
try { app.quit(); } catch (e) {}
```

- `app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES)` — 丢弃任何"未保存"状态，
  这样 `app.quit()` 不弹框
- 顺序反了 `app.quit()` 先发现 project 脏 → 弹框
- 两步都裹 try/catch 防异常吞掉 .done 写入（.done 必须在 close/quit 之前完成）
- pre-clean（开头那段 `app.project.close + app.newProject`）防 -r 多轮残留，
  end-cleanup 防遗害下次 launch — **两段都要**

验证（确认 AE 真退）:

```bash
AE_SHIP_GATE=1 go test ... # 或 RE 命令
sleep 3
tasklist 2>/dev/null | grep -i afterfx  # 应该返空
```

## 执行 + 等待 marker

```bash
rm -f test_data/re_<name>.done
"E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/test_data/re_<name>.jsx"

# 等 .done 出现（不要 sleep 死循环）
until [ -f test_data/re_<name>.done ]; do sleep 2; done
cat test_data/re_<name>.done
```

## 探针式 JSX（探未知 API）

不知道 AE ScriptingAPI 有什么字段时，先用探针：

```js
step("probe_keys", function () {
    var keys = [];
    for (var k in td) keys.push(k);
    log.push("  TextDocument keys: " + keys.join(","));
});

step("probe_app_fonts", function () {
    var keys = [];
    for (var k in app.fonts) keys.push(k);
    log.push("  app.fonts keys: " + keys.join(","));
});
```

输出在 `.done` 里。常见教训：ExtendScript 允许对任意 key 赋值不报错（即使 key 不存在），所以"不抛错 ≠ 真生效"，必须 dump btdk 字节确认。详见 `incidents/variable-fonts-write-noop.md`。

## 字节 diff

```bash
go run ./tmp_debug/parse_btdk test_data/re_<name>.aep <layer_filter> > /tmp/dump_a.txt
go run ./tmp_debug/parse_btdk test_data/re_<name>.aep <other_layer> > /tmp/dump_b.txt
diff /tmp/dump_a.txt /tmp/dump_b.txt
```

字节级 diff 是 ground truth：API 接受了赋值但 diff 为空 = runtime-only field。

## 验证后回归

新字段读 RE 完，加测试时复用现有 fixture：

```go
proj, err := aep.Open("../../test_data/re_<name>.aep")
if err != nil {
    t.Skipf("re_<name>.aep not present; run test_data/re_<name>.jsx in AE")
}
```

`t.Skipf` 让 fixture 缺失时跳过而不是失败，保持 CI 可移植。
