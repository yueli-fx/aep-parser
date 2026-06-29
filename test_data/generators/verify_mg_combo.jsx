// Acceptance verify for MG roadmap S6: the end-to-end combination gate. Reads
// mg_combo_args.json {input, done, resaved, png}. One all-Go-built project that
// combines precomp nesting + Trim + ease keyframes + Repeater. Proves AE accepts
// it, reads back the structural facts (parent has 4 layers; BADGE is a precomp
// of MGX_Badge; MOVER has 2 eased position keyframes; DOTS has a Repeater), then
// renders the mid-frame so Go can pixel-assert all four features coexist
// (red line 4 — the composition is an independent deliverable).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_combo_args.json");
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

        var scene = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "MGX_Scene") { scene = it; break; }
        }
        if (!scene) { fail("comp MGX_Scene not found"); }
        else {
            note("scene layers=" + scene.numLayers);
            if (scene.numLayers !== 4) fail("expected 4 layers, got " + scene.numLayers);

            var byName = {};
            for (var li = 1; li <= scene.numLayers; li++) byName[scene.layer(li).name] = scene.layer(li);

            var badge = byName["BADGE"];
            if (!badge) fail("BADGE missing");
            else {
                var src = badge.source;
                note("BADGE source=" + (src ? src.name : "?") + " isComp=" + (src instanceof CompItem));
                if (!(src instanceof CompItem) || src.name !== "MGX_Badge") fail("BADGE source not MGX_Badge precomp");
            }

            var mover = byName["MOVER"];
            if (!mover) fail("MOVER missing");
            else {
                var pos = mover.transform.position;
                note("MOVER numKeys=" + pos.numKeys);
                if (pos.numKeys !== 2) fail("MOVER numKeys=" + pos.numKeys + ", want 2");
                else {
                    var inf = pos.keyOutTemporalEase(1)[0].influence;
                    note("MOVER kf1 out influence=" + inf);
                    if (inf < 80) fail("MOVER ease lost (influence=" + inf + ")");
                }
            }

            var dots = byName["DOTS"];
            if (!dots) fail("DOTS missing");
            else {
                var rep = dots.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Repeater");
                if (!rep) fail("DOTS has no Repeater");
                else note("DOTS repeater Copies=" + rep.property("ADBE Vector Repeater Copies").value);
            }
        }

        app.project.save(new File(args.resaved));

        if (scene && args.png) {
            app.project.bitsPerChannel = 8;
            scene.saveFrameToPng(2, new File(args.png));
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
