// Acceptance verify for MG roadmap S3: Trim Paths from scratch. Reads
// mg_trim_args.json {input, done, resaved, png}. Proves AE accepts the file,
// reads back the trim End on the HALF layer (50) and FULL layer (100), resaves
// for the Go side, and renders frame 0 so Go can assert on actual pixels that
// the HALF ring is cut to an arc (right present / left absent) while FULL is a
// complete ring (red line 4 — render the capability's visible surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_trim_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function trimEnd(layer) {
        var vg = layer.property("ADBE Root Vectors Group").property(1); // Vector Group
        var contents = vg.property("ADBE Vectors Group");
        var trim = contents.property("ADBE Vector Filter - Trim");
        if (!trim) return null;
        return trim.property("ADBE Vector Trim End").value;
    }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MGTRIM") { comp = it; break; }
        }
        if (!comp) { fail("comp MGTRIM not found"); }
        else {
            note("comp=" + comp.name + " layers=" + comp.numLayers);
            if (comp.numLayers !== 3) fail("expected 3 layers, got " + comp.numLayers);

            var byName = {};
            for (var li = 1; li <= comp.numLayers; li++) byName[comp.layer(li).name] = comp.layer(li);

            var half = byName["HALF"];
            if (!half) { fail("HALF missing"); }
            else {
                var he = trimEnd(half);
                note("HALF trim End=" + he);
                if (he === null) fail("HALF has no Trim filter");
                else if (Math.abs(he - 50) > 0.5) fail("HALF trim End=" + he + ", want 50");
            }

            var full = byName["FULL"];
            if (!full) { fail("FULL missing"); }
            else {
                var fe = trimEnd(full);
                note("FULL trim End=" + fe);
                if (fe === null) fail("FULL has no Trim filter");
                else if (Math.abs(fe - 100) > 0.5) fail("FULL trim End=" + fe + ", want 100");
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
