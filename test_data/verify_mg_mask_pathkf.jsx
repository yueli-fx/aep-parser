// Acceptance verify for SetMaskPathKeyframes (priority-4 mask): an ANIMATED mask
// outline (left-half rect at t=0 → right-half rect at t=2s). Reads
// mg_mask_pathkf_args.json {input, done, resaved, png0, png1}. Proves AE accepts
// the animated mask (ADBE Mask Shape numKeys==6, 2 lhd3 capacity pages), resaves it, and renders the two
// keyframe times so Go can assert the revealed region MOVED (left→right) across
// time — red line 4: a static mask cannot reveal opposite halves at two frames, so
// the render proves AE honored the time table, not just a value round-trip.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_mask_pathkf_args.json");
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
            if (it instanceof CompItem && it.name === "MASKPATHKF") { comp = it; break; }
        }
        if (!comp) { fail("comp MASKPATHKF not found"); }
        else {
            var card = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "CARD") card = comp.layer(li);
            if (!card) { fail("CARD layer missing"); }
            else {
                var mk = card.property("ADBE Mask Parade").property(1);
                if (!mk) { fail("CARD has no mask"); }
                else {
                    var shapeProp = mk.property("ADBE Mask Shape");
                    note("mask shape numKeys=" + shapeProp.numKeys);
                    if (shapeProp.numKeys !== 6) fail("mask shape numKeys=" + shapeProp.numKeys + ", want 6 (animated, 2 capacity pages)");
                    else {
                        note("kf1 verts=" + shapeProp.keyValue(1).vertices.length +
                             " kf6 verts=" + shapeProp.keyValue(6).vertices.length);
                        note("kf1 t=" + shapeProp.keyTime(1) + " kf6 t=" + shapeProp.keyTime(6));
                    }
                }
            }
        }

        app.project.save(new File(args.resaved));

        if (comp && args.png0 && args.png1) {
            app.project.bitsPerChannel = 8;
            comp.saveFrameToPng(0, new File(args.png0));   // t=0  → left half revealed
            comp.saveFrameToPng(2, new File(args.png1));   // t=2s → right half revealed
            $.sleep(2000);
            var p0 = new File(args.png0), p1 = new File(args.png1);
            if (p0.exists) { note("png0 " + p0.length + " bytes"); } else { fail("saveFrameToPng wrote no png0"); }
            if (p1.exists) { note("png1 " + p1.length + " bytes"); } else { fail("saveFrameToPng wrote no png1"); }
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
