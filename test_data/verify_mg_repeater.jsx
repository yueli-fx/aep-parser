// Acceptance verify for MG roadmap S5: Repeater from scratch. Reads
// mg_repeater_args.json {input, done, resaved, png}. Proves AE accepts the file,
// reads back the repeater Copies (5) + Transform Position (300,0), resaves, and
// renders frame 0 so Go can assert on actual pixels that the single dot was
// duplicated into a row of 5 (red line 4 — render the capability's surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_repeater_args.json");
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
            if (it instanceof CompItem && it.name === "MGREP") { comp = it; break; }
        }
        if (!comp) { fail("comp MGREP not found"); }
        else {
            var dots = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "DOTS") dots = comp.layer(li);
            if (!dots) { fail("DOTS layer missing"); }
            else {
                var rep = dots.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Repeater");
                if (!rep) { fail("DOTS has no Repeater filter"); }
                else {
                    var copies = rep.property("ADBE Vector Repeater Copies").value;
                    var pos = rep.property("ADBE Vector Repeater Transform")
                        .property("ADBE Vector Repeater Position").value;
                    note("repeater Copies=" + copies + " Position=" + pos.join(","));
                    if (Math.abs(copies - 5) > 0.5) fail("Copies=" + copies + ", want 5");
                    if (Math.abs(pos[0] - 300) > 1) fail("Position.x=" + pos[0] + ", want 300");
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
