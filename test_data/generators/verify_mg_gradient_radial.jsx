// Acceptance verify for gradient RADIAL type (Grad Type=2) from scratch.
// Reads mg_gradient_radial_args.json {input, done, resaved, png}. Proves AE
// accepts the file, reads back Grad Type=2 (Radial), resaves, and renders
// frame 0 so Go can assert on actual pixels that the red→blue fill is
// concentric (rotationally symmetric), not a linear band (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_gradient_radial_args.json");
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
            if (it instanceof CompItem && it.name === "MGRADIAL") { comp = it; break; }
        }
        if (!comp) { fail("comp MGRADIAL not found"); }
        else {
            var ramp = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "RADIAL") ramp = comp.layer(li);
            if (!ramp) { fail("RADIAL layer missing"); }
            else {
                var gf = ramp.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Graphic - G-Fill");
                if (!gf) { fail("RADIAL has no G-Fill"); }
                else {
                    var gt = gf.property("ADBE Vector Grad Type").value;
                    note("Grad Type=" + gt + " (1=linear 2=radial)");
                    if (gt !== 2) fail("Grad Type=" + gt + ", want 2 (radial)");
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
