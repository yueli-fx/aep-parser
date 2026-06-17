// Acceptance verify for the Wiggle modulation ROUND-TRIP params — the noise
// phase / coherence knobs that select a different random instance with no
// categorically-correct pixel result (so they are roundtrip-verified, not
// pixel-gated; same class as RandomSeed). Reads mg_wiggle_modrt_args.json
// {input, done, resaved}. Proves AE ACCEPTS the from-scratch file (does not
// reject it as corrupt), reads back each spliced value from its specific filter,
// and resaves. Covers:
//   ROUGH  layer · ADBE Vector Filter - Roughen : Temporal Phase, Spatial Phase
//   WTRANS layer · ADBE Vector Filter - Wiggler : Correlation, Temporal Phase,
//                                                 Spatial Phase
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_wiggle_modrt_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function approx(a, b) { return Math.abs(a - b) < 0.01; }

    function filterOf(layer, mn) {
        return layer.property("ADBE Root Vectors Group").property(1)
            .property("ADBE Vectors Group").property(mn);
    }

    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "WGMODRT") { comp = it; break; }
        }
        if (!comp) { fail("comp WGMODRT not found"); }
        else {
            var rough = null, wtrans = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "ROUGH") rough = comp.layer(li);
                if (comp.layer(li).name === "WTRANS") wtrans = comp.layer(li);
            }
            if (!rough) fail("ROUGH layer missing");
            else {
                var rg = filterOf(rough, "ADBE Vector Filter - Roughen");
                if (!rg) fail("ROUGH has no Roughen filter");
                else {
                    var tp = rg.property("ADBE Vector Temporal Phase").value;
                    var sp = rg.property("ADBE Vector Spatial Phase").value;
                    note("ROUGH TemporalPhase=" + tp + " SpatialPhase=" + sp);
                    if (!approx(tp, 45)) fail("ROUGH TemporalPhase=" + tp + " want 45");
                    if (!approx(sp, 30)) fail("ROUGH SpatialPhase=" + sp + " want 30");
                }
            }
            if (!wtrans) fail("WTRANS layer missing");
            else {
                var wg = filterOf(wtrans, "ADBE Vector Filter - Wiggler");
                if (!wg) fail("WTRANS has no Wiggler filter");
                else {
                    var co = wg.property("ADBE Vector Correlation").value;
                    var tp2 = wg.property("ADBE Vector Temporal Phase").value;
                    var sp2 = wg.property("ADBE Vector Spatial Phase").value;
                    note("WTRANS Correlation=" + co + " TemporalPhase=" + tp2 + " SpatialPhase=" + sp2);
                    if (!approx(co, 20)) fail("WTRANS Correlation=" + co + " want 20");
                    if (!approx(tp2, 40)) fail("WTRANS TemporalPhase=" + tp2 + " want 40");
                    if (!approx(sp2, 50)) fail("WTRANS SpatialPhase=" + sp2 + " want 50");
                }
            }
        }
        app.project.save(new File(args.resaved));
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }

    var done = new File(args.done);
    done.open("w");
    done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
