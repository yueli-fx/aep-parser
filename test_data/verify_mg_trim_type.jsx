// Acceptance verify for Trim Paths Trim Type (priority-3 shape remaining): the
// AE-default-elided `ADBE Vector Trim Type` enum, materialized from scratch by
// synthesis-insert. Reads mg_trim_type_args.json {input, done, resaved, png}.
// Proves AE accepts the from-scratch file, reads back Trim Type (2 =
// Individually) + End (50), resaves, and renders frame 0 so Go can assert on
// actual pixels that Individually trims each path in sequence (left ellipse full,
// right ellipse empty) — distinct from the default Simultaneously (both ellipses
// half). Red line 4: render the capability's surface; an enum value that only
// round-trips proves nothing about AE's render.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_trim_type_args.json");
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
            if (it instanceof CompItem && it.name === "TRIMTYPE") { comp = it; break; }
        }
        if (!comp) { fail("comp TRIMTYPE not found"); }
        else {
            var card = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "CARD") card = comp.layer(li);
            if (!card) { fail("CARD layer missing"); }
            else {
                var trim = card.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Trim");
                if (!trim) { fail("CARD has no Trim filter"); }
                else {
                    var end = trim.property("ADBE Vector Trim End").value;
                    var typ = trim.property("ADBE Vector Trim Type").value;
                    note("trim End=" + end + " Type=" + typ);
                    if (Math.abs(end - 50) > 1) fail("End=" + end + ", want 50");
                    if (Math.abs(typ - 2) > 0.01) fail("Trim Type=" + typ + ", want 2 (Individually)");
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
