// Acceptance verify for ZigZag Points (priority-3 shape remaining): the
// AE-default-elided `ADBE Vector Zigzag Points` enum (Corner=1 default /
// Smooth=2), materialized from scratch by synthesis-insert. Reads
// mg_zigzag_points_args.json {input, done, resaved, png}. Proves AE accepts the
// from-scratch file, reads back Points (2 = Smooth) + Size/Detail, resaves, and
// renders frame 0 so Go can assert on actual pixels that the ridges are SMOOTH
// (rounded scalloped wave) — distinct from the default CORNER (sharp sawtooth).
// Red line 4: render the capability's surface; an enum value that only
// round-trips proves nothing about AE's render.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_zigzag_points_args.json");
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
            if (it instanceof CompItem && it.name === "ZZPOINTS") { comp = it; break; }
        }
        if (!comp) { fail("comp ZZPOINTS not found"); }
        else {
            var smooth = null, corner = null;
            for (var li = 1; li <= comp.numLayers; li++) {
                if (comp.layer(li).name === "SMOOTH") smooth = comp.layer(li);
                if (comp.layer(li).name === "CORNER") corner = comp.layer(li);
            }
            if (!smooth) { fail("SMOOTH layer missing"); }
            else {
                var zz = smooth.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Zigzag");
                if (!zz) { fail("SMOOTH has no ZigZag filter"); }
                else {
                    var size = zz.property("ADBE Vector Zigzag Size").value;
                    var detail = zz.property("ADBE Vector Zigzag Detail").value;
                    var pts = zz.property("ADBE Vector Zigzag Points").value;
                    note("SMOOTH zigzag Size=" + size + " Detail=" + detail + " Points=" + pts);
                    if (Math.abs(pts - 2) > 0.01) fail("SMOOTH Points=" + pts + ", want 2 (Smooth)");
                }
            }
            if (corner) {
                var czz = corner.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Zigzag");
                if (czz) note("CORNER zigzag Points=" + czz.property("ADBE Vector Zigzag Points").value + " (default 1)");
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
