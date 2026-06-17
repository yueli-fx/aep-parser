// Acceptance verify for MG roadmap S5: Wiggle Paths from scratch. Reads
// mg_wiggle_args.json {input, done, resaved, png}. Proves AE accepts the file,
// reads back the four modeled sub-streams (Size / Detail / Wiggles-Per-Second /
// Random Seed), resaves, and renders frame 0 so Go can assert on actual pixels
// that a sharp rectangle's straight edges were roughened into a noisy boundary —
// red line 4, render the capability's surface, not just the value.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_wiggle_args.json");
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
            if (it instanceof CompItem && it.name === "MGWG") { comp = it; break; }
        }
        if (!comp) { fail("comp MGWG not found"); }
        else {
            var card = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "CARD") card = comp.layer(li);
            if (!card) { fail("CARD layer missing"); }
            else {
                var wg = card.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Roughen");
                if (!wg) { fail("CARD has no Wiggle Paths filter"); }
                else {
                    var sz = wg.property("ADBE Vector Roughen Size").value;
                    var dt = wg.property("ADBE Vector Roughen Detail").value;
                    var fr = wg.property("ADBE Vector Temporal Freq").value;
                    var sd = wg.property("ADBE Vector Random Seed").value;
                    note("wiggle Size=" + sz + " Detail=" + dt + " Freq=" + fr + " Seed=" + sd);
                    if (Math.abs(sz - 60) > 1) fail("Size=" + sz + ", want 60");
                    if (Math.abs(dt - 30) > 1) fail("Detail=" + dt + ", want 30");
                    if (Math.abs(fr - 4) > 0.5) fail("Freq=" + fr + ", want 4");
                    if (Math.abs(sd - 9) > 0.5) fail("Seed=" + sd + ", want 9");
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
