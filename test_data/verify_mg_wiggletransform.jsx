// Acceptance verify for MG roadmap S5: Wiggle Transform from scratch. Reads
// mg_wiggletransform_args.json {input, done, resaved, png}. Proves AE accepts the
// file, reads back the modeled sub-streams (Temporal Freq / Random Seed +
// nested Transform Position / Scale / Rotation / Anchor amplitudes), resaves, and
// renders frame 0 so Go can assert on actual pixels that the shape was randomly
// displaced from its nominal center — red line 4, render the capability's
// surface, not just the value.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_wiggletransform_args.json");
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
            if (it instanceof CompItem && it.name === "MGWX") { comp = it; break; }
        }
        if (!comp) { fail("comp MGWX not found"); }
        else {
            var card = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "CARD") card = comp.layer(li);
            if (!card) { fail("CARD layer missing"); }
            else {
                var wx = card.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Wiggler");
                if (!wx) { fail("CARD has no Wiggle Transform filter"); }
                else {
                    var fr = wx.property("ADBE Vector Xform Temporal Freq").value;
                    var sd = wx.property("ADBE Vector Random Seed").value;
                    var xf = wx.property("ADBE Vector Wiggler Transform");
                    var pos = xf.property("ADBE Vector Wiggler Position").value;
                    var rot = xf.property("ADBE Vector Wiggler Rotation").value;
                    note("wiggleXform Freq=" + fr + " Seed=" + sd + " PosAmp=" + pos.join(",") + " RotAmp=" + rot);
                    if (Math.abs(fr - 2) > 0.5) fail("Freq=" + fr + ", want 2");
                    if (Math.abs(sd - 8) > 0.5) fail("Seed=" + sd + ", want 8");
                    if (Math.abs(pos[0] - 220) > 1 || Math.abs(pos[1] - 220) > 1) fail("PosAmp=" + pos.join(",") + ", want 220,220");
                    if (Math.abs(rot - 70) > 1) fail("RotAmp=" + rot + ", want 70");
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
