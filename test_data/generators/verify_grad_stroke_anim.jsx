// Ship gate for ANIMATED gradient STROKE color stops. Reads
// grad_stroke_anim_args.json {input, done, resaved, png0, png1}. Proves AE
// accepts the Go-built animated gradient stroke (keyframe time-table +
// per-keyframe Utf8 stops), reads back the keyframe count on the stroke's Grad
// Colors stream, resaves, and renders t=0 and t=1s so Go can assert the stops
// swapped (red line 4 — the colour sweep along the stroke).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/grad_stroke_anim_args.json");
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
            if (it instanceof CompItem && it.name === "GRADSTROKEANIM") { comp = it; break; }
        }
        if (!comp) { fail("comp GRADSTROKEANIM not found"); }
        else {
            var lyr = comp.layer(1);
            var vg = lyr.property("ADBE Root Vectors Group").property(1);
            var contents = vg.property("ADBE Vectors Group");
            var gs = contents.property("ADBE Vector Graphic - G-Stroke");
            if (!gs) { fail("no G-Stroke"); }
            else {
                var colors = gs.property("ADBE Vector Grad Colors");
                note("Grad Colors numKeys=" + colors.numKeys);
                if (colors.numKeys !== 2) fail("Grad Colors numKeys=" + colors.numKeys + ", want 2");
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png0));
            comp.saveFrameToPng(1, new File(args.png1));
            $.sleep(1500);
            if (!new File(args.png0).exists) fail("png0 not written");
            if (!new File(args.png1).exists) fail("png1 not written");
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
