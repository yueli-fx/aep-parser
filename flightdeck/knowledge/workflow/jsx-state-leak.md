# ⚠ JSX -r 多轮累积 duplicate comps

JSX -r 多轮累积 duplicate comps

## 症状

第 N 次跑 `AfterFX.exe -r re_varfont_ae24.jsx` 后，Go 端测试找到的是第 1 轮的 layer（带历史 mutation 的 `Bahnschrift:wght=700`），不是第 N 轮新写的（resolved PostScript=Bahnschrift-Bold）。

## 根因

AE 保持 project 在内存中跨 `-r` 调用。每次 JSX `app.project.items.addComp("RE_VARFONT", ...)` 都是 *新加* 一个同名 comp（AE 允许重名），不是替换。多轮后 .aep 累积成 N 个 "RE_VARFONT" comp。

Go 测试 `for _, c := range proj.Compositions { if c.Name == "..." }` 命中第一个匹配的，是历史污染版。

## 教训

JSX 顶端必须 reset:

```js
step("fresh_project", function () {
    app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
    app.newProject();
    app.project.save(outFile);
});
```

写新 fixture 时固定做这一步：每次 rerun 前删除旧 comp 或新建工程，避免历史 comp 留在项目里污染 Go test。

Bug-fix 类 RE 改 JSX 后重跑前最好 `rm -f test_data/<name>.aep` 强制从零生成。
