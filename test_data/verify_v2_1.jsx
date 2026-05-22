// test_data/verify_v2_1.jsx
//
// AE-side ship gate for V2.1 (aep.NewProject + Project.NewComposition).
//
// Driven by Go test (TestV2_1_AEShipGate_*). Go end:
//   1. NewProject + NewComposition(Main, BG_loop) + WriteAEP → <input>.aep
//   2. Write args.json with absolute paths
//   3. AfterFX.exe -r verify_v2_1.jsx
//   4. Wait for <done>.txt
//   5. Assert first line == "PASS"
//
// args.json (fixed path: test_data/v2_1_args.json) contains:
//   {"input": "<abs path>", "done": "<abs path>", "resaved": "<abs path>"}
//
// Everything in try/catch to guarantee .done is written (防 Go 端 hang to timeout).

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_1_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        // ExtendScript may lack JSON.parse on old AE; eval is the cross-version fallback.
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        var inFile = new File(args.input);
        app.open(inFile);
        log.push("opened " + inFile.fsName);
        log.push("items.length=" + app.project.items.length);
        ok = (app.project.items.length === 2);

        if (ok) {
            // app.project.items is 1-indexed in AE Scripting
            // Find Main + BG_loop by name (order may vary)
            var main = null, bg = null;
            for (var i = 1; i <= app.project.items.length; i++) {
                var it = app.project.items[i];
                if (it.name === "Main") main = it;
                if (it.name === "BG_loop") bg = it;
            }
            if (!main || !bg) {
                ok = false;
                log.push("missing items: main=" + (main ? "ok" : "NIL") + " bg=" + (bg ? "ok" : "NIL"));
            } else {
                // AE save normalization checks
                // shutterAngle 不参与 gate: AE 25 ScriptingAPI 对 NTSC fps comp
                // 返回 stored × 1.2 (29.97: stored 180 → API 返回 216) —— 储存
                // 字节跟 AE 自存的 A_baseline 完全一致，AE UI 显示 180，仅 ScriptingAPI
                // bug。详 workshop/scars/ae25-acceptance-gate.md。
                // AE returns bgColor 0..1 normalized, NOT 0..255. 20/255≈0.0784 etc.
                ok = (main.width === 1920 && main.height === 1080
                      && Math.abs(main.frameRate - 29.97) < 1e-4
                      && Math.abs(main.duration - 10) < 0.05
                      && Math.abs(main.pixelAspect - 1.0) < 1e-5
                      && Math.abs(main.bgColor[0] - 20/255) < 1e-3
                      && Math.abs(main.bgColor[1] - 30/255) < 1e-3
                      && Math.abs(main.bgColor[2] - 40/255) < 1e-3);
                log.push("Main: " + main.width + "x" + main.height
                         + " fps=" + main.frameRate + " dur=" + main.duration
                         + " sa=" + main.shutterAngle + " par=" + main.pixelAspect
                         + " bg=[" + main.bgColor.join(",") + "]");
                log.push("BG_loop: " + bg.width + "x" + bg.height + " fps=" + bg.frameRate + " dur=" + bg.duration);
            }
        }

        // Reopen-save-resave-reopen tier:
        // AE 能开 ≠ AE 认结构合法。Resave + reopen 才暴露 silent rewrite/drop
        if (ok && args.resaved) {
            var resavedFile = new File(args.resaved);
            app.project.save(resavedFile);
            app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
            app.open(resavedFile);
            ok = (app.project.items.length === 2);
            var main2 = null;
            for (var j = 1; j <= app.project.items.length; j++) {
                if (app.project.items[j].name === "Main") main2 = app.project.items[j];
            }
            if (!main2) {
                ok = false;
                log.push("resaved: Main missing");
            } else {
                ok = ok && Math.abs(main2.frameRate - 29.97) < 1e-4
                     && Math.abs(main2.duration - 10) < 0.05
                     && Math.abs(main2.bgColor[0] - 20/255) < 1e-3;
                log.push("resaved Main: fps=" + main2.frameRate + " dur=" + main2.duration);
            }
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    // .done 必写出 (try/catch 防 Go 端 hang to timeout)
    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_1_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* 写不出 .done → Go 端 timeout */ }
})();
