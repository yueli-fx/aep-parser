// Ship-gate verify for batch-7 single-run text SetRun* setters. Reads
// text_run_args.json {input, done, resaved}. Opens the Go-written file, finds the
// text layer, reads back the whole-document textDocument properties AE ingested
// (a single-run text doc surfaces each SetRun0 attribute as a whole-doc value —
// readable on AE2020 too, no characterRange needed). Resaves for Go survival check.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/text_run_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL " + m); }
    function near(label, got, want, tol) {
        if (tol === undefined) tol = 0.02;
        if (got === undefined || got === null) { fail(label + " undefined"); return; }
        if (Math.abs(got - want) > tol) fail(label + "=" + got + " want " + want);
        else log.push("  " + label + "=" + got);
    }
    function eqBool(label, got, want) {
        if (got === undefined) { fail(label + " undefined"); return; }
        if (!!got !== !!want) fail(label + "=" + got + " want " + want);
        else log.push("  " + label + "=" + got);
    }
    function rgb(label, got, want) {
        if (!got) { fail(label + " undefined"); return; }
        for (var i = 0; i < 3; i++) if (Math.abs(got[i] - want[i]) > 0.02) { fail(label + "[" + i + "]=" + got[i] + " want " + want[i]); return; }
        log.push("  " + label + "=[" + got.join(",") + "]");
    }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { comp = it; break; }
        }
        if (!comp) { fail("no CompItem"); }
        else {
            var L = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                var cand = comp.layer(li);
                if (cand.property("Source Text") !== null) { L = cand; break; }
            }
            if (!L) { fail("no text layer"); }
            else {
                var td = L.property("Source Text").value;
                rgb("fillColor", td.fillColor, [0.2, 0.4, 0.8]);
                eqBool("applyStroke", td.applyStroke, true);
                rgb("strokeColor", td.strokeColor, [0.9, 0.1, 0.3]);
                near("strokeWidth", td.strokeWidth, 5);
                eqBool("strokeOverFill", td.strokeOverFill, false);
                eqBool("fauxBold", td.fauxBold, true);
                eqBool("fauxItalic", td.fauxItalic, true);
                near("baselineShift", td.baselineShift, 8, 0.1);
                eqBool("autoLeading", td.autoLeading, false);
                near("leading", td.leading, 80, 0.1);
                // Uncertain on AE2020 — logged, trimmed by gate if absent.
                try { near("horizontalScale", td.horizontalScale, 50, 0.5); } catch (e) { log.push("  horizontalScale N/A: " + e); }
                try { near("verticalScale", td.verticalScale, 150, 0.5); } catch (e) { log.push("  verticalScale N/A: " + e); }
                try { near("tsume", td.tsume, 0.5, 0.02); } catch (e) { log.push("  tsume N/A: " + e); }
            }
        }
        app.project.save(new File(args.resaved));
    } catch (e) {
        fail("EXC " + e.toString());
    }

    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
