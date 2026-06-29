// Acceptance verify for Merge Paths modes Add / Subtract / Intersect / Exclude
// (priority-3 shape remaining). Reads mg_merge_modes_args.json
// {input, done, resaved, png}. Proves AE accepts the from-scratch file, reads
// back each card's Merge Type (2/3/4/5), resaves, and renders frame 0 so Go can
// assert each boolean mode's distinct signature across three regions
// (left-only / overlap / right-only) — red line 4: render the capability's
// surface; the enum round-tripping proves nothing about AE's boolean combine.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_merge_modes_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    var want = { "ADD": 2, "SUB": 3, "INT": 4, "EXC": 5 };

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MERGEMODES") { comp = it; break; }
        }
        if (!comp) { fail("comp MERGEMODES not found"); }
        else {
            for (var li = 1; li <= comp.numLayers; li++) {
                var lay = comp.layer(li);
                if (!(lay.name in want)) continue;
                var mg = lay.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Merge");
                if (!mg) { fail(lay.name + " has no Merge filter"); continue; }
                var typ = mg.property("ADBE Vector Merge Type").value;
                note(lay.name + " MergeType=" + typ);
                if (Math.abs(typ - want[lay.name]) > 0.01) fail(lay.name + " Type=" + typ + ", want " + want[lay.name]);
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
