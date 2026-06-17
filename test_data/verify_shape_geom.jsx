// Ship-gate verify for batch-8b shape geometry scalar setters. Reads
// shape_geom_args.json {input, done, resaved}. Confirms AE accepts the file and a
// shape layer exists, then resaves so the Go side can re-parse and assert each
// geometry cdat value survived AE's engine (acceptance-preservation, value-checked).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/shape_geom_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");
    var ok = true, log = [];
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp || comp.numLayers < 1) { ok = false; log.push("no comp/layer"); }
        else { log.push("layers=" + comp.numLayers); }
        app.project.save(new File(args.resaved));
    } catch (e) { ok = false; log.push("EXC " + e.toString()); }
    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
