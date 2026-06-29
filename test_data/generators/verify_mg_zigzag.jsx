// Acceptance verify for MG roadmap S5: ZigZag from scratch. Reads
// mg_zigzag_args.json {input, done, resaved, png}. Proves AE accepts the file,
// reads back the ZigZag Size (40) + Detail (8), resaves, and renders frame 0 so
// Go can assert on actual pixels that the rectangle's straight edges became
// jagged (red line 4 — render the capability's surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_zigzag_args.json");
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
            if (it instanceof CompItem && it.name === "MGZIG") { comp = it; break; }
        }
        if (!comp) { fail("comp MGZIG not found"); }
        else {
            var wave = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "WAVE") wave = comp.layer(li);
            if (!wave) { fail("WAVE layer missing"); }
            else {
                var zz = wave.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Filter - Zigzag");
                if (!zz) { fail("WAVE has no ZigZag filter"); }
                else {
                    var size = zz.property("ADBE Vector Zigzag Size").value;
                    var detail = zz.property("ADBE Vector Zigzag Detail").value;
                    note("zigzag Size=" + size + " Detail=" + detail);
                    if (Math.abs(size - 40) > 1) fail("Size=" + size + ", want 40");
                    if (Math.abs(detail - 8) > 1) fail("Detail=" + detail + ", want 8");
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
