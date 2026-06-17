// Custom-chunk preservation probe. Reads chunk_probe_args.json {input, done,
// resaved}. Opens the Go-built file (which carries an injected unknown RIFX
// chunk) and resaves it — the Go side then checks whether the chunk survived
// AE's re-encode. PASS just means AE opened + resaved without error.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/chunk_probe_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    try {
        app.open(new File(args.input));
        note("numItems=" + app.project.numItems);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "PROBE") { comp = it; break; }
        }
        if (!comp) fail("comp PROBE not found");
        else note("comp PROBE layers=" + comp.numLayers);
        app.project.save(new File(args.resaved));
        note("resaved ok");
    } catch (e) {
        fail("EXC " + e.toString() + " line=" + e.line);
    }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
