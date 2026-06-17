// Acceptance verify for MG roadmap S5: PolyStar from scratch. Reads
// mg_star_args.json {input, done, resaved, png}. Proves AE accepts the file,
// reads back the star Points (5) + Outer/Inner Radius (250/100), resaves, and
// renders frame 0 so Go can assert on actual pixels that a 5-pointed star was
// drawn (red line 4 — render the capability's surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/mg_star_args.json");
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
            if (it instanceof CompItem && it.name === "MGSTAR") { comp = it; break; }
        }
        if (!comp) { fail("comp MGSTAR not found"); }
        else {
            var lyr = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "STAR") lyr = comp.layer(li);
            if (!lyr) { fail("STAR layer missing"); }
            else {
                var star = lyr.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Shape - Star");
                if (!star) { fail("STAR has no Star shape"); }
                else {
                    var pts = star.property("ADBE Vector Star Points").value;
                    var outr = star.property("ADBE Vector Star Outer Radius").value;
                    var innr = star.property("ADBE Vector Star Inner Radius").value;
                    note("star Points=" + pts + " OuterR=" + outr + " InnerR=" + innr);
                    if (Math.abs(pts - 5) > 0.5) fail("Points=" + pts + ", want 5");
                    if (Math.abs(outr - 250) > 1) fail("OuterR=" + outr + ", want 250");
                    if (Math.abs(innr - 100) > 1) fail("InnerR=" + innr + ", want 100");
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
