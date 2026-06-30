# JSX RE 工作流 + ship-gate — checklist

SUMMARY: JSX RE 工作流 + ship-gate
READ WHEN: writing a new RE JSX fixture; debugging field locations via byte-diff against AE-saved baseline; setting up cross-version AE comparison; running a ship-gate against AE (modified .aep accepted/rejected); diagnosing why AE rejects a builder-written file; looking up "which AE version introduced field X"; invoking ae_run.ps1 wrapper for unattended ship-gate; deciding which AE version to use for a new fixture

---

## 总览

写 fixture 用 ExtendScript（JSX）→ 让 AE 跑一次 → 拿到 `.aep` → Go 端 parse_btdk diff 找字段。Ship-gate 反过来：Go 写改过的 `.aep` → AE 打开 → 读出值跟 Go 写入比对。

## Agent 自己调 AE — 不要 ask user

`scripts/ae-worker/ae_run.ps1` 是 **unattended wrapper**（OCR + multi-signal modal 自动消化，exit code 0/1/2/3/4/5）。Agent 应直接通过 Bash/PowerShell tool 调它，**不要**把 "请 user 跑 AE" 当默认。

调用模板（用 `$env:` 传 JSX 内 `$.getenv()` 读到的 mode/参数）：

```powershell
$env:MY_VAR = "value"   # 如果 JSX 用 $.getenv() 读 mode
Remove-Item -ErrorAction SilentlyContinue test_data/re_X.done
pwsh -NoProfile -File scripts/ae-worker/ae_run.ps1 `
    -AeExe "E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe" `
    -Jsx   "E:\projects\tools\aep-parser\test_data\generators\re_X.jsx" `
    -Done  "E:\projects\tools\aep-parser\test_data\re_X.done" `
    -TimeoutSec 180
```

每次跑前先删旧 `.done`（wrapper 等的是新 `.done` 出现）。跑完读 `cat test_data/re_X.done` 看 JSX 步骤 log。

何时**应该** ask user：(a) AE 没装在标准路径需要他指路；(b) wrapper 反复 exit 2（未知 modal），rules 加完仍 fail；(c) 跨 monitor 输出 hang 等 GUI 异常。其他情况一律自己跑。

## 新 fixture 默认走 AE 2020

项目读取下限 = AE 2020。**新 RE fixture 默认用 AE 2020 跑**，理由：

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

**防 ID-巧合假绿（新结构性写路径必做）**：只在单一 canonical fixture 上 mutate 的 gate 可能**靠 ID 巧合通过**——embed 模板的宿主层/item ID 与 baseline fixture 恰好同号（同套 JSX → 同 ID 分配序列），AE open 校验时不暴露。AddEffect Phase-1 双版本 10/10 全绿却带潜伏 bug（效果模板 sspc 每参数 tdpi=来源宿主层 ID=15，baseline 也恰是 15）就栽在这，直到 parade-auto-create 用 100% Go-built 全新工程（层 ID≠15）才暴露。**对策**：① 除 fixture-mutate 场景外，**必加一个 from-scratch 全 Go-built 场景**（NewProject→…→WriteAEP，ID 空间天然错开提取 fixture）；② embed-AE-bytes 模板落库前用 `tools/debug/effect_id_scan`（全树扫 32-bit 值）查模板里有没有嵌着来源文件的 item/layer ID。

## Adobe 软件路径

- AE 2025 — `E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe`
- AE 2024 — `E:\adobe\Adobe After Effects 2024\Support Files\AfterFX.exe`（2026-06-12 解锁；AE 24-引入的字段/特性如 characterRange 多 run、fontCapsOption 可用它当「24+ 但非 25」的第二门禁版本——给那些 AE 2020 前向拒开的 24+ fixture 补双版本 gate，配 AE 2025 凑齐 24+×双版本）
- AE 2023 — `E:\adobe\Adobe After Effects 2023\Support Files\AfterFX.exe`
- AE 2022 — `E:\adobe\Adobe After Effects 2022\Support Files\AfterFX.exe`
- AE 2020 — `E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe`

跨版本对比时切对应 AE。Wave 1-3 字段默认用 AE 2025（24+ ScriptingAPI 解封）。

**版本匹配规则（建议，非强制）**：

- 用 **fixture 源版本** 的 AE 打开 (检查方式: `Application.Version()` 读 head 解出 e.g. `17.7x45` = AE 17.7 = AE 2020)
- 例：`re_cameralight.aep` 是 AE 2020 写的 → 优先用 AE 2020 跑（避无意义 convert 流程）
- 反向（高版本 fixture → 低版本 AE）经常崩，仍**避免**
- 跨版本测：`scripts/ae-worker/ae_run.ps1` wrapper 自动消化 convert 对话框，不再被 GUI 阻塞 unattended run（详 § GDI 自动化）

## AE 打开 .aep 的 3 种失败模式

调 AE-side ship-gate 不能盲信单一信号，先**辨别模式**：

| 模式 | 信号 | 处理 |
|---|---|---|
| **1. 完全崩溃** | `tasklist`/Get-Process 看不到 AfterFX.exe；没有 `.done`；exit code 非 0 | builder 写的字段触发 AE 内部 sanity-check fail (e.g. cdta timing 空)。先最小化生成工程，再逐字段二分定位触发字节。 |
| **2. 打开但需转换** | GUI 弹 "Convert?" 对话框 → JSX 跑不到 `app.open` 返回，要么 catch 到 error，要么 hang。`.done` 含 ERR 信息（或根本写不出） | 用 `scripts/ae-worker/ae_run.ps1` wrapper 自动消化 convert 对话框；裸 `AfterFX -r` 仍需版本匹配 |
| **3. 打开但报数据损坏** | JSX 跑通；`app.open(...)` 在 try/catch 里 throw "After Effects 错误: 文件数据丢失" 类错误字符串 | builder chunk 写法 / 位置 / 大小破坏 AE 检查。这是最常见且最有 RE 价值的 — bisect 隔离哪个 setter 触发 |

诊断流程（按这个 order）：

1. AE process 是否仍 alive (`tasklist /v | grep -i afterfx`)
2. `.done` 是否生成（生成 = mode 3 / OK；不生成 = mode 1 / 2）
3. 看 `.done` 里 step error 信息：开头 fresh/open OK 但 open_modified ERR = mode 3；纯 fresh ERR = mode 1/2

**不要直接归 mode 3** 去 RE 字节，先排除 mode 1 / 2。

## GDI 自动化 — `scripts/ae-worker/ae_run.ps1`

Ship-gate 用 `scripts/ae-worker/ae_run.ps1` 替代裸 `AfterFX -r`：OCR + multi-signal dispatch 自动消化 convert / save / data-loss modal。任意 AE 版本 × 任意 fixture 都能 unattended 跑。**Cross-version smoke PASS**：AE 2020 fixture 用 AE 2025 跑，convert 对话框被自动消化（2026-05-27 验证）。

**调用契约**：

```powershell
pwsh -NoProfile -File scripts/ae-worker/ae_run.ps1 `
    -AeExe   $aeExe `
    -Jsx     $jsxPath `
    -Done    $doneFile `
    -TimeoutSec 180
```

退出码：

- `0` — `.done` 出现且 stable，按 PASS contract 走
- `1` — timeout
- `2` — 未知 modal（OCR 命中但没规则）——**OCR 文本 + 规则骨架会直接打到 stderr**，不用翻 dump
- `3` — OCR engine init 失败
- `4` — AE 进程启动失败
- `6` — **已有 AfterFX 进程在跑**（屏幕矩形 OCR 会读到它的对话框）。先 `Get-Process AfterFX* | Stop-Process -Force`，确实要并行才传 `-IgnoreRunningAe`
- `7` — **Abort 规则命中 = 需用户介入的环境故障**（stderr 有 `USER INTERVENTION REQUIRED` + 修法）。已知案例：AE「允许脚本写入文件和访问网络」首选项未开（新装 / prefs 重建后默认关，**per-AE-version**），开 首选项>脚本和表达式 勾上再重跑——agent 自己修不了，必须提示用户

失败时 dump 落 `<doneFile>.fail/`：`screenshot.png` / `ocr.txt` / `windows.txt` / `actions.log` / `meta.json`。

**Go 端 ship-gate**用 `runAeRunShipGate(t, aeExe, jsxPath, doneFile, timeoutSec)` helper (`internal/aep/testutil_shipgate_test.go`)，不要再直接 `exec.Command(aeExe, "-r", ...)`。helper 自带两个机制：

- **test-cache 文件追踪**：helper 显式读 verify JSX / ae_run.ps1 / 规则表，把它们纳入 go test cache key——改 JSX 自动失效缓存，**不再需要靠 `-count=1` 纪律防「JSX 改了缓存还绿」**（gate 实跑时仍建议 `-count=1` 强制真跑）。
- **exit 1/2 自动 warm-retry 一次**：冷启动 flake 自愈；确定性 reject 重试照样红，不会被掩盖。首跑取证保留在 `<doneFile>.fail.1/`。

**⚠️ 非目标：不要并行化 AE gates**（双 AE 同屏会让矩形抓图/前台按键互踩；exit 6 守卫就是为此存在）。

## 全量 gate sweep + fixture 再生（批处理入口）

- **`scripts/fixtures/run_ship_gates.ps1`** — 全量 ship-gate 回归 sweep：预清理残留 AfterFX → `AE_SHIP_GATE=1 go test -run ShipGate -count=1 -v` → 写 PASS/FAIL/SKIP 台账（`test_data/gate_ledger.json` + `gate_sweep.log`）。跨切面改动（ID 分配 / write 路径 / wrapper）后跑一次；`-Run <pattern>` 可只跑子集；`-ClearCrashState` 在 force-kill 弄脏 crash flag 后用。
- **`scripts/fixtures/regen_fixtures.ps1`** — 按 `scripts/fixtures/fixtures_manifest.json`（44 个生成 JSX 的 AE 版本 / env-mode 矩阵 / 期望产物）无人值守重建 fixture。**默认 only-missing 不碰已有文件**（AE 保存非确定性，乱重生会 churn byte-diff 基线）；`-CheckOnly` 盘点缺失 + 无主 fixture；`-Force` 全重建；`-Only <substring>` 过滤。新增生成 JSX 时**必须**同步 manifest 加条目。

**⚠️ verify JSX `.done` 文件名必须 per-version 唯一**：ship-gate 同一 mode 跨 AE 2020/2025 跑两遍时，verify JSX 若把 `.done` 文件名只按 mode 命名，会(a)第二版覆盖第一版结果、(b)`ae_run.ps1 -Done` 等的是带 version-tag 的名 → 永远等不到 → 每次空等满 `TimeoutSec` 报 exit 1（实际验证早已跑通，假阴性）。修法：JSX 读一个 `$.getenv("..._TAG")`（如 `ae2020_basic`）拼进 `.done` 名，调用方 `-Done` 传同名。见 `verify_ge_insert_layer.jsx` + 2026-05-29 logbook。

**新对话框出现的流程**：

1. ship-gate FAIL, exit code 2
2. 看 `<doneFile>.fail/screenshot.png` + `ocr.txt`
3. 加规则到 `scripts/ae-worker/ae_dialog_rules.json`（substring 进 `ocrMatch`，或抄稳定 `windowTitle` / `windowClass`）
4. `Parse-Rules` 跑通 → re-run ship-gate

## 工作流

```
1. 写 test_data/re_<feature>.jsx
2. pwsh -File scripts/ae-worker/ae_run.ps1 -AeExe ... -Jsx ... -Done ...  (unattended)
3. JSX 末尾写 .done marker 标记完成
4. Go 端读 .aep → parse_btdk dump → 看新字段在哪
5. 写 test_data/re_<feature>_*_test.go 用 fixture 测试
```

裸 `AfterFX -r` 仍可手动用做 quick RE，但 ship-gate / 自动化场景一律走 wrapper。

## JSX 模板（关键防陷阱）

```js
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_<name>.aep");
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

## render-pixel gate：saveFrameToPng 磁盘缓存陷阱（关键陷阱）

写**多 comp 的 render-pixel ship-gate**（每个 knob/特征一个 comp，逐个 `comp.saveFrameToPng` 抓帧比像素）时，会撞 AE 的**持久磁盘帧缓存**：缓存键 ≈ `(comp.id, render-time)`，**跨进程不失效**（缓存在 `%LOCALAPPDATA%\Temp\Adobe\After Effects\<ver>\Disk Cache-*.noindex`）。

两个坑叠加 → **每张 PNG 内容都一样**（甚至跨独立冷启 AE 进程返回更早某次 run 的旧帧），且 gate 可能**假绿**（值/DOM 全对、像素被污染却恰好过断言）：

1. **comp.id 撞车**：纯 Go 从零 builder 给每个单 comp 工程的 comp 恒分配同一 `comp.id`（实测全 `=1`）。
2. **time 没传**：JSX 若硬编码 `saveFrameToPng(0)` 而忽略 per-comp time，所有帧同键。

**`app.purge(PurgeTarget.ALL_CACHES)` 救不了**——它清 RAM 缓存，不碰磁盘缓存（PurgeTarget 无 disk 项）。

**修法（两道防线，缺一不可）**：
- JSX 渲染每个 comp 用**唯一 time**（`comp.saveFrameToPng(job.time, png)`，Go 端逐 comp 给开 ≥0.5s 的不同 time）→ run 内每 comp 唯一缓存键。
- Go harness 渲染前**清磁盘缓存**：删 `Temp\Adobe\After Effects\*\Disk Cache*.noindex`（= AE「Empty Disk Cache」按钮，缓存会自动重建，无数据丢失）→ 消除跨会话/历史中毒帧。仅做唯一 time 不够：被历史 `t=0` run 污染的桶仍会喂旧帧。

参考实现：`internal/aep/mg_text_style_shipgate_test.go`（`clearAEDiskCache` helper）+ `test_data/generators/verify_mg_text_style.jsx`。**自验铁律**：gate 跑完逐张 `md5sum` 应**全不同**、并 `Read` 几张 PNG 目视确认渲的是各自 comp（红线4：值对 ≠ 渲染对）。如果多张 PNG 内容一样，优先怀疑 AE disk cache / stale-frame，而不是 setter 立即失败。

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
"E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r "E:/projects/tools/aep-parser/test_data/generators/re_<name>.jsx"

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

输出在 `.done` 里。常见教训：ExtendScript 允许对任意 key 赋值不报错（即使 key 不存在），所以"不抛错 ≠ 真生效"，必须 dump btdk 字节确认。

**⚠ 数组值日志拼接陷阱**：`"" + prop.value` 对 AE **color** 数组值会抛 `数字结果无效（除以零？）`（valueOf 走数值转换）——看起来像数据 reject，实为 JSX 日志行炸了。数组一律显式 `value.toString()` / `value.join(",")` 再拼。

## 字节 diff

```bash
go run ./tools/debug/parse_btdk test_data/fixtures/re_<name>.aep <layer_filter> > /tmp/dump_a.txt
go run ./tools/debug/parse_btdk test_data/fixtures/re_<name>.aep <other_layer> > /tmp/dump_b.txt
diff /tmp/dump_a.txt /tmp/dump_b.txt
```

字节级 diff 是 ground truth：API 接受了赋值但 diff 为空 = runtime-only field。

## 验证后回归

新字段读 RE 完，加测试时复用现有 fixture：

```go
proj, err := aep.Open("../../test_data/fixtures/re_<name>.aep")
if err != nil {
    t.Skipf("re_<name>.aep not present; run test_data/generators/re_<name>.jsx in AE")
}
```

`t.Skipf` 让 fixture 缺失时跳过而不是失败，保持 CI 可移植。
