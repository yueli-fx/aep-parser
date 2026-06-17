// Ship gate for an ANIMATED (keyframed) effect parameter. Reads
// anim_effect_args.json {input, done, resaved, png0, png1}. Proves AE accepts
// the Go-synthesized keyframe container on a materialized effect param (Gaussian
// Blur Blurriness, kf 0@0s → 100@2s): reads back numKeys + the values, resaves,
// renders t=0 and t=2.5s so Go can assert the blur grows (red line 4 — the
// effect param is keyframed, not just stored).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/anim_effect_args.json");
    argsFile.open("r"); var raw = argsFile.read(); argsFile.close();
    var args = eval("(" + raw + ")");
    var log = []; var ok = true;
    function fail(m) { ok = false; log.push("  FAIL: " + m); }
    function note(m) { log.push("  " + m); }
    try {
        app.open(new File(args.input));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "ANIMEFFECT") { comp = it; break; }
        }
        if (!comp) { fail("comp ANIMEFFECT not found"); }
        else {
            var sq = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "SQ") sq = comp.layer(li);
            if (!sq) { fail("SQ layer missing"); }
            else {
                var blur = sq.property("ADBE Effect Parade").property(1).property("ADBE Gaussian Blur 2-0001");
                note("blur numKeys=" + blur.numKeys);
                if (blur.numKeys !== 2) fail("blur numKeys=" + blur.numKeys + ", want 2");
                else note("kf1=" + blur.keyValue(1) + " kf2=" + blur.keyValue(2) + " | val@0=" + blur.valueAtTime(0, false) + " val@2.5=" + blur.valueAtTime(2.5, false));
            }
            app.project.save(new File(args.resaved));
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png0));    // blur 0  → sharp
            comp.saveFrameToPng(2.5, new File(args.png1));  // blur 100 → bleed
            $.sleep(2000);
            if (!new File(args.png0).exists) fail("png0 not written");
            if (!new File(args.png1).exists) fail("png1 not written");
        }
    } catch (e) { fail("EXC " + e.toString() + " line=" + e.line); }
    var done = new File(args.done);
    done.open("w"); done.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n")); done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
