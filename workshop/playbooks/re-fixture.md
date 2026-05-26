---
when_to_read: writing a new RE JSX fixture; debugging field locations via byte-diff against AE-saved baseline; setting up cross-version AE comparison; running a ship-gate against AE (modified .aep accepted/rejected); diagnosing why AE rejects a builder-written file
applies_to: [jsx, re-workflow, fixture, ae-cli, byte-diff, baseline-strategy, ship-gate, ae-acceptance, version-mismatch, failure-modes]
last_updated: 2026-05-26
---

# JSX RE 工作流 + ship-gate

## 总览

写 fixture 用 ExtendScript（JSX）→ 让 AE 跑一次 → 拿到 `.aep` → Go 端 parse_btdk diff 找字段。Ship-gate 反过来：Go 写改过的 `.aep` → AE 打开 → 读出值跟 Go 写入比对。

## 何时启用 ship-gate

任何**结构性 / chunk-level 改动**入 main 前必须跑：
- 加新 chunk / 删 chunk / 改 chunk size（length-variable splice 含 Utf8 / expression / 字体名）
- Project-level setting chunk（lnrb/lnrp 等 toggle / acer/adfr/dwga 类单字段）
- NewProject / NewComposition / NewShapeLayer 等结构性创建
- 跨 AE 版本字段（AE 24+ 解封 / ldta 长度变化）

length-preserving 单字段（cdta 单 offset 改 / ldta flag bit 改）roundtrip Go test 够，不必走 ship-gate。

## Adobe 软件路径

- AE 2025 — `E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe`
- AE 2022 — `E:\adobe\Adobe After Effects 2022\Support Files\AfterFX.exe`
- AE 2020 — `E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe`

跨版本对比时切对应 AE。Wave 1-3 字段默认用 AE 2025（24+ ScriptingAPI 解封）。

**版本匹配规则（避免 conversion prompt）**：
- 用 **fixture 源版本** 的 AE 打开 (检查方式: `Application.Version()` 读 head 解出 e.g. `17.7x45` = AE 17.7 = AE 2020)
- 例：`re_cameralight.aep` 是 AE 2020 写的 → 用 AE 2020 跑 ship-gate / RE，**不要用 AE 2025**（触发 17.1→25 转换提示，可能弹 GUI 阻塞 unattended run）
- 跨版本测必须时，用 AE 高版本测低版本文件（向后兼容更稳）；反向高版本→低版本经常崩

## AE 打开 .aep 的 3 种失败模式

调 AE-side ship-gate 不能盲信单一信号，先**辨别模式**：

| 模式 | 信号 | 处理 |
|---|---|---|
| **1. 完全崩溃** | `tasklist`/Get-Process 看不到 AfterFX.exe；没有 `.done`；exit code 非 0 | builder 写的字段触发 AE 内部 sanity-check fail (e.g. cdta timing 空)。看 `scars/ae25-acceptance-gate.md` Stage 1 / 4 类 |
| **2. 打开但需转换** | GUI 弹 "Convert?" 对话框 → JSX 跑不到 `app.open` 返回，要么 catch 到 error，要么 hang。`.done` 含 ERR 信息（或根本写不出） | 版本不匹配。换匹配版本 AE 跑，或后期上 GDI 自动化点 "确定" |
| **3. 打开但报数据损坏** | JSX 跑通；`app.open(...)` 在 try/catch 里 throw "After Effects 错误: 文件数据丢失" 类错误字符串 | builder chunk 写法 / 位置 / 大小破坏 AE 检查。这是最常见且最有 RE 价值的 — bisect 隔离哪个 setter 触发 |

诊断流程（按这个 order）：
1. AE process 是否仍 alive (`tasklist /v | grep -i afterfx`)
2. `.done` 是否生成（生成 = mode 3 / OK；不生成 = mode 1 / 2）
3. 看 `.done` 里 step error 信息：开头 fresh/open OK 但 open_modified ERR = mode 3；纯 fresh ERR = mode 1/2

**不要直接归 mode 3** 去 RE 字节，先排除 mode 1 / 2。

## GDI / 屏幕截图自动化 (planned, 未实现)

目前 ship-gate 还遇 GUI 拦截需要 fall back 到匹配版本绕开。**计划**：`scripts/ae_run.ps1` 包 AfterFX -r，用 PowerShell `[System.Drawing]::CopyFromScreen` 截图 + `[System.Windows.Forms.SendKeys]` 模拟键盘，自动点掉常见对话框 (convert / save changes / OK)。Claude 读截图认对话框文本 → 决定按哪个键。**目标**：unattended 跑任意 AE 版本，不被弹框阻塞。

当前权宜：版本匹配 + .done timeout 失败时人工排查。

## 工作流

```
1. 写 test_data/re_<feature>.jsx
2. 让用户开 AE
3. AfterFX.exe -r path.jsx 执行（不需要交互）
4. JSX 末尾写 .done marker 标记完成
5. Go 端读 .aep → parse_btdk dump → 看新字段在哪
6. 写 test_data/re_<feature>_*_test.go 用 fixture 测试
```

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

输出在 `.done` 里。常见教训：ExtendScript 允许对任意 key 赋值不报错（即使 key 不存在），所以"不抛错 ≠ 真生效"，必须 dump btdk 字节确认。详见 `scars/variable-fonts-write-noop.md`。

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
