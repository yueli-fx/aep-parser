// Acceptance verify for the Wiggle modulation render gate (Wiggle Paths Points +
// Correlation — the shape vector-filter family's last visually-gatable knobs).
// Reads mg_wiggle_mod_args.json {input, done, resaved, png}. Proves AE accepts
// the from-scratch file, reads back Points / Correlation from each card's
// `ADBE Vector Filter - Roughen`, resaves, and renders frame 0 so Go can assert
// on actual pixels (red line 4: a value that only round-trips proves nothing
// about AE's render). Fixed Seed → frame 0 is deterministic + cross-version
// reproducible.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_wiggle_mod_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function roughenOf(layer) {
        return layer.property("ADBE Root Vectors Group").property(1)
            .property("ADBE Vectors Group").property("ADBE Vector Filter - Roughen");
    }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "WGMOD") { comp = it; break; }
        }
        if (!comp) { fail("comp WGMOD not found"); }
        else {
            var want = { CORNER: { pts: 1 }, SMOOTH: { pts: 2 }, CORRLOW: { corr: 0 }, CORRHIGH: { corr: 100 } };
            for (var li = 1; li <= comp.numLayers; li++) {
                var ly = comp.layer(li);
                if (!want[ly.name]) continue;
                var rg = roughenOf(ly);
                if (!rg) { fail(ly.name + " has no Roughen filter"); continue; }
                var pts = rg.property("ADBE Vector Roughen Points");
                var corr = rg.property("ADBE Vector Correlation");
                note(ly.name + " Points=" + (pts ? pts.value : "?") + " Correlation=" + (corr ? corr.value : "?"));
                if (want[ly.name].pts !== undefined && pts && Math.abs(pts.value - want[ly.name].pts) > 0.01)
                    fail(ly.name + " Points=" + pts.value + " want " + want[ly.name].pts);
                if (want[ly.name].corr !== undefined && corr && Math.abs(corr.value - want[ly.name].corr) > 0.01)
                    fail(ly.name + " Correlation=" + corr.value + " want " + want[ly.name].corr);
            }
        }

        app.project.save(new File(args.resaved));

        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png));
            $.sleep(2000);
            var pngFile = new File(args.png);
            if (pngFile.exists) note("frame png " + pngFile.length + " bytes");
            else fail("saveFrameToPng wrote no file");
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
