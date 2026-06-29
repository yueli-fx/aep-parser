// Acceptance verify for MG roadmap S5: Merge Paths from scratch. Reads
// mg_merge_args.json {input, done, resaved, png}. Proves AE accepts the file,
// reads back the Merge Type (3=Subtract), resaves, and renders frame 0 so Go can
// assert on actual pixels that the ellipse was subtracted from the rectangle —
// a square with a round hole (red line 4 — render the capability's surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_merge_args.json");
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
            if (it instanceof CompItem && it.name === "MGMERGE") { comp = it; break; }
        }
        if (!comp) { fail("comp MGMERGE not found"); }
        else {
            var cut = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "CUT") cut = comp.layer(li);
            if (!cut) { fail("CUT layer missing"); }
            else {
                var mg = cut.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Merge");
                if (!mg) { fail("CUT has no Merge Paths filter"); }
                else {
                    var type = mg.property("ADBE Vector Merge Type").value;
                    note("merge Type=" + type);
                    if (Math.abs(type - 3) > 0.5) fail("Type=" + type + ", want 3 (Subtract)");
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
