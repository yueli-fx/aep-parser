// Acceptance verify for Twist Center (priority-3 shape remaining): the
// AE-default-elided `ADBE Vector Twist Center` Vec2 ([0,0] default),
// materialized from scratch by synthesis-insert. Reads
// mg_twist_center_args.json {input, done, resaved, png}. Proves AE accepts the
// from-scratch file, reads back Center ([150,0]) + Angle on the OFFSET card,
// resaves, and renders frame 0 so Go can assert on actual pixels that the offset
// pivot shifts the twist's centroid — distinct from the default centred twist.
// Red line 4: render the capability's surface; a Vec2 that only round-trips
// proves nothing about AE's render.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_twist_center_args.json");
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
            if (it instanceof CompItem && it.name === "TWCENTER") { comp = it; break; }
        }
        if (!comp) { fail("comp TWCENTER not found"); }
        else {
            var offset = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "OFFSET") offset = comp.layer(li);
            if (!offset) { fail("OFFSET layer missing"); }
            else {
                var tw = offset.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Twist");
                if (!tw) { fail("OFFSET has no Twist filter"); }
                else {
                    var ang = tw.property("ADBE Vector Twist Angle").value;
                    var ctr = tw.property("ADBE Vector Twist Center").value;
                    note("OFFSET twist Angle=" + ang + " Center=[" + ctr.join(",") + "]");
                    if (Math.abs(ctr[0] - 150) > 1 || Math.abs(ctr[1] - 0) > 1) fail("Center=[" + ctr.join(",") + "], want [150,0]");
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
