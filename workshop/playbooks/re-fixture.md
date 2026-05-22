# JSX RE 工作流

## 总览

写 fixture 用 ExtendScript（JSX）→ 让 AE 跑一次 → 拿到 `.aep` → Go 端 parse_btdk diff 找字段。

## Adobe 软件路径

- AE 2025 — `E:\adobe\Adobe After Effects 2025\Support Files\AfterFX.exe`
- AE 2022 — `E:\adobe\Adobe After Effects 2022\Support Files\AfterFX.exe`
- AE 2020 — `E:\adobe\Adobe After Effects 2020\Support Files\AfterFX.exe`

跨版本对比时切对应 AE。Wave 1-3 字段默认用 AE 2025（24+ ScriptingAPI 解封）。

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
})();
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
