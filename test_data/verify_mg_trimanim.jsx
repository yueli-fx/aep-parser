// Acceptance verify for animated Trim Paths (line-draw reveal). Reads
// mg_trimanim_args.json {input, done, resaved, pngEarly, pngMid, pngFull}.
// Proves AE accepts the file, reads back the keyframed Trim End (2 keys: 0→100),
// resaves for the Go side, and renders three frames (t=0/2/4) so Go can assert
// on actual pixels that the stroked ring sweeps open over time (red line 4 —
// render the capability's visible surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_trimanim_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function trimEndProp(layer) {
        var vg = layer.property("ADBE Root Vectors Group").property(1); // Vector Group
        var contents = vg.property("ADBE Vectors Group");
        var trim = contents.property("ADBE Vector Filter - Trim");
        if (!trim) return null;
        return trim.property("ADBE Vector Trim End");
    }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MGTRIMANIM") { comp = it; break; }
        }
        if (!comp) { fail("comp MGTRIMANIM not found"); }
        else {
            note("comp=" + comp.name + " layers=" + comp.numLayers);
            if (comp.numLayers !== 2) fail("expected 2 layers, got " + comp.numLayers);

            var reveal = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "REVEAL") { reveal = comp.layer(li); break; }
            }
            if (!reveal) { fail("REVEAL missing"); }
            else {
                var ep = trimEndProp(reveal);
                if (ep === null) { fail("REVEAL has no Trim filter"); }
                else {
                    note("Trim End numKeys=" + ep.numKeys);
                    if (ep.numKeys !== 2) fail("Trim End numKeys=" + ep.numKeys + ", want 2");
                    else {
                        var k1 = ep.keyValue(1), k2 = ep.keyValue(2);
                        note("Trim End key1=" + k1 + " key2=" + k2);
                        if (Math.abs(k1 - 0) > 0.5) fail("Trim End key1=" + k1 + ", want 0");
                        if (Math.abs(k2 - 100) > 0.5) fail("Trim End key2=" + k2 + ", want 100");
                    }
                }
            }
        }

        app.project.save(new File(args.resaved));

        if (comp) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.pngEarly));
            comp.saveFrameToPng(2, new File(args.pngMid));
            comp.saveFrameToPng(4, new File(args.pngFull));
            $.sleep(2000);
            var names = ["pngEarly", "pngMid", "pngFull"];
            for (var n = 0; n < names.length; n++) {
                var pf = new File(args[names[n]]);
                if (pf.exists) { note(names[n] + " " + pf.length + " bytes"); }
                else { fail("saveFrameToPng wrote no file: " + names[n]); }
            }
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
