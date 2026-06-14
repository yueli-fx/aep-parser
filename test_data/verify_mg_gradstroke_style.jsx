// Acceptance verify for gradient-stroke geometry (width / cap / join / miter)
// from scratch. Reads mg_gradstroke_style_args.json {input, done, resaved, png}.
// Proves AE accepts the file, reads back the G-Stroke's stroke geometry, resaves,
// and renders frame 0 so Go can pixel-assert the 60px band width (red line 4).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_gradstroke_style_args.json");
    argsFile.open("r");
    var raw = argsFile.read();
    argsFile.close();
    var args = eval("(" + raw + ")");

    var log = [];
    var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    function near(label, got, want) {
        if (Math.abs(got - want) > 0.5) fail(label + "=" + got + " want " + want);
        else note(label + "=" + got);
    }

    try {
        app.open(new File(args.input));

        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MGGSTYLE") { comp = it; break; }
        }
        if (!comp) { fail("comp MGGSTYLE not found"); }
        else {
            var layer = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "GSWIDTH") layer = comp.layer(li);
            if (!layer) { fail("GSWIDTH layer missing"); }
            else {
                var gs = layer.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Graphic - G-Stroke");
                if (!gs) { fail("GSWIDTH has no G-Stroke"); }
                else {
                    near("strokeWidth", gs.property("ADBE Vector Stroke Width").value, 60);
                    near("lineCap", gs.property("ADBE Vector Stroke Line Cap").value, 3);   // Projecting
                    near("lineJoin", gs.property("ADBE Vector Stroke Line Join").value, 3); // Bevel
                    near("miterLimit", gs.property("ADBE Vector Stroke Miter Limit").value, 10);
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
