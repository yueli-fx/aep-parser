// Acceptance verify for Repeater Order (priority-3 shape remaining): the
// AE-default-elided `ADBE Vector Repeater Order` (Composite Below=1 default /
// Above=2) enum, materialized from scratch by synthesis-insert (spliced before
// the Repeater Transform group). Reads mg_repeater_order_args.json {input, done,
// resaved, png}. Proves AE accepts the file, reads back Order=2 on the ABOVE
// card, resaves, and renders frame 0. Note: Order is a compositing-order setting
// with NO visible effect on a single-color fill (compositing same-color layers
// is commutative) — the Go gate confirms BELOW and ABOVE render identical
// coverage, so this verify checks round-trip + acceptance, not a visible change.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_repeater_order_args.json");
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
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "REPORDER") { comp = it; break; }
        }
        if (!comp) { fail("comp REPORDER not found"); }
        else {
            var above = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "ABOVE") above = comp.layer(li);
            if (!above) { fail("ABOVE layer missing"); }
            else {
                var rep = above.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Repeater");
                if (!rep) { fail("ABOVE has no Repeater filter"); }
                else {
                    var ord = rep.property("ADBE Vector Repeater Order").value;
                    var cps = rep.property("ADBE Vector Repeater Copies").value;
                    note("ABOVE Repeater Copies=" + cps + " Order=" + ord);
                    if (Math.abs(ord - 2) > 0.01) fail("Order=" + ord + ", want 2 (Above)");
                    if (Math.abs(cps - 5) > 0.01) fail("Copies=" + cps + ", want 5");
                }
            }
        }

        app.project.save(new File(args.resaved));

        if (comp && args.png) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png));
            $.sleep(2000);
            var pngFile = new File(args.png);
            if (pngFile.exists) { note("frame png " + pngFile.length + " bytes"); }
            else { fail("saveFrameToPng wrote no file"); }
        }
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
