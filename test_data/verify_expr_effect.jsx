// Ship gate for an EXPRESSION driving an EFFECT parameter. Reads
// expr_effect_args.json {input, done, resaved, png0, png1}. Proves AE evaluates
// an expression on a materialized effect param (Gaussian Blur Blurriness =
// time*40): reads back expressionEnabled + the evaluated value at t=0 and t=2.5s,
// resaves, and renders both frames so Go can assert the blur actually grows
// (red line 4 — the effect is driven by the expression, not just stored).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/expr_effect_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "EXPREFFECT") { comp = it; break; }
        }
        if (!comp) { fail("comp EXPREFFECT not found"); }
        else {
            var sq = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "SQ") sq = comp.layer(li);
            if (!sq) { fail("SQ layer missing"); }
            else {
                var fx = sq.property("ADBE Effect Parade").property(1);
                if (!fx) { fail("no effect on SQ"); }
                else {
                    var blur = fx.property("ADBE Gaussian Blur 2-0001"); // Blurriness
                    note("blur exprEnabled=" + blur.expressionEnabled + " expr=" + blur.expression);
                    if (!blur.expressionEnabled) fail("blur expression not enabled");
                    note("blur valueAtTime(0)=" + blur.valueAtTime(0, false) + " valueAtTime(2.5)=" + blur.valueAtTime(2.5, false));
                }
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png0));    // blur 0  → sharp
            comp.saveFrameToPng(2.5, new File(args.png1));  // blur 100 → bleed
            $.sleep(2000);
            if (!new File(args.png0).exists) fail("png0 not written");
            if (!new File(args.png1).exists) fail("png1 not written");
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }
    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
