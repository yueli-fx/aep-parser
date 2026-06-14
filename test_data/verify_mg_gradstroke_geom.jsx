// Acceptance verify for gradient STROKE ramp geometry from scratch. Reads
// mg_gradstroke_geom_args.json {input, done, resaved, png}. Proves AE accepts
// the file, reads back GSDIR's G-Stroke as linear (Grad Type=1) and GSRAD's as
// radial (Grad Type=2), resaves, and renders frame 0 so Go can assert on actual
// pixels (red line 4) that the stroke ramp direction + type are honored.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_gradstroke_geom_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }

    function gstrokeOf(layer) {
        return layer.property("ADBE Root Vectors Group").property(1)
            .property("ADBE Vectors Group").property("ADBE Vector Graphic - G-Stroke");
    }
    function layerByName(comp, nm) {
        for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === nm) return comp.layer(li);
        return null;
    }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MGGSTROKE") { comp = it; break; }
        }
        if (!comp) { fail("comp MGGSTROKE not found"); }
        else {
            var dir = layerByName(comp, "GSDIR");
            var rad = layerByName(comp, "GSRAD");
            var hl = layerByName(comp, "GSHL");
            if (!dir) fail("GSDIR layer missing");
            if (!rad) fail("GSRAD layer missing");
            if (!hl) fail("GSHL layer missing");
            if (dir) {
                var gsd = gstrokeOf(dir);
                if (!gsd) fail("GSDIR has no G-Stroke");
                else {
                    var dt = gsd.property("ADBE Vector Grad Type").value;
                    note("GSDIR Grad Type=" + dt + " (want 1 linear)");
                    if (dt !== 1) fail("GSDIR Grad Type=" + dt + ", want 1");
                }
            }
            if (rad) {
                var gsr = gstrokeOf(rad);
                if (!gsr) fail("GSRAD has no G-Stroke");
                else {
                    var rt = gsr.property("ADBE Vector Grad Type").value;
                    note("GSRAD Grad Type=" + rt + " (want 2 radial)");
                    if (rt !== 2) fail("GSRAD Grad Type=" + rt + ", want 2");
                }
            }
            if (hl) {
                var gsh = gstrokeOf(hl);
                if (!gsh) fail("GSHL has no G-Stroke");
                else {
                    var ht = gsh.property("ADBE Vector Grad Type").value;
                    var hll = gsh.property("ADBE Vector Grad HiLite Length").value;
                    note("GSHL Grad Type=" + ht + " HiLite Length=" + hll + " (want 2 / 70)");
                    if (ht !== 2) fail("GSHL Grad Type=" + ht + ", want 2");
                    if (Math.abs(hll - 70) > 0.5) fail("GSHL HiLite Length=" + hll + ", want 70");
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
