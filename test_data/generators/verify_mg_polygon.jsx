// Acceptance verify for PolyStar POLYGON type (Star Type=2) from scratch. Reads
// mg_polygon_args.json {input, done, resaved, png}. Proves AE accepts the file,
// reads back Star Type=2 (Polygon) + Points=6, resaves, and renders frame 0 so
// Go can pixel-assert the convex hexagon (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_polygon_args.json");
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
            if (it instanceof CompItem && it.name === "MGPOLY") { comp = it; break; }
        }
        if (!comp) { fail("comp MGPOLY not found"); }
        else {
            var layer = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "POLY") layer = comp.layer(li);
            if (!layer) { fail("POLY layer missing"); }
            else {
                var star = layer.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Shape - Star");
                if (!star) { fail("POLY has no Star shape"); }
                else {
                    var ty = star.property("ADBE Vector Star Type").value;
                    var pts = star.property("ADBE Vector Star Points").value;
                    note("Star Type=" + ty + " Points=" + pts + " (want 2 / 6)");
                    if (ty !== 2) fail("Star Type=" + ty + ", want 2 (polygon)");
                    if (Math.abs(pts - 6) > 0.5) fail("Points=" + pts + ", want 6");
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
