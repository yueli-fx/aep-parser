// Acceptance verify for shape Fill/Stroke Opacity (priority-3 shape remaining,
// secondary sub-properties). Reads mg_opacity_args.json {input, done, resaved,
// png}. Proves AE accepts the from-scratch file, reads back each card's opacity
// (100 / 50), resaves, and renders frame 0 so Go can assert on actual pixels that
// the 50% cards render at ~half luminance over the dark BG while the 100% cards
// render full white — red line 4: render the capability's surface; an opacity
// value that only round-trips proves nothing about AE's composite.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_opacity_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    var want = { "FILLFULL": 100, "FILLHALF": 50, "STRKFULL": 100, "STRKHALF": 50 };

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "OPACITY") { comp = it; break; }
        }
        if (!comp) { fail("comp OPACITY not found"); }
        else {
            for (var li = 1; li <= comp.numLayers; li++) {
                var lay = comp.layer(li);
                if (!(lay.name in want)) continue;
                var grp = lay.property("ADBE Root Vectors Group").property(1).property("ADBE Vectors Group");
                var op = null;
                try { op = grp.property("ADBE Vector Graphic - Fill").property("ADBE Vector Fill Opacity").value; }
                catch (e0) {
                    try { op = grp.property("ADBE Vector Graphic - Stroke").property("ADBE Vector Stroke Opacity").value; }
                    catch (e1) { op = null; }
                }
                if (op === null) { fail(lay.name + " has no opacity stream"); continue; }
                note(lay.name + " opacity=" + op);
                if (Math.abs(op - want[lay.name]) > 1) fail(lay.name + " opacity=" + op + ", want " + want[lay.name]);
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
