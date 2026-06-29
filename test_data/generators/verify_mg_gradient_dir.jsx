// Acceptance verify for MG roadmap S5: gradient direction (Grad Start/End Pt)
// from scratch. Reads mg_gradient_dir_args.json {input, done, resaved, png}.
// Proves AE accepts the file, reads back the Grad Start Pt [-150,-150] + End Pt
// [150,150], resaves, and renders frame 0 so Go can assert on actual pixels that
// the red→blue ramp runs diagonally (red line 4 — render the capability's
// surface).
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/mg_gradient_dir_args.json");
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
            if (it instanceof CompItem && it.name === "MGGRAD") { comp = it; break; }
        }
        if (!comp) { fail("comp MGGRAD not found"); }
        else {
            var ramp = null;
            for (var li = 1; li <= comp.numLayers; li++) if (comp.layer(li).name === "RAMP") ramp = comp.layer(li);
            if (!ramp) { fail("RAMP layer missing"); }
            else {
                var gf = ramp.property("ADBE Root Vectors Group").property(1)
                    .property("ADBE Vectors Group").property("ADBE Vector Graphic - G-Fill");
                if (!gf) { fail("RAMP has no G-Fill"); }
                else {
                    var sp = gf.property("ADBE Vector Grad Start Pt").value;
                    var ep = gf.property("ADBE Vector Grad End Pt").value;
                    note("Grad StartPt=" + sp.join(",") + " EndPt=" + ep.join(","));
                    if (Math.abs(sp[0] + 150) > 1 || Math.abs(sp[1] + 150) > 1) fail("StartPt=" + sp.join(",") + ", want -150,-150");
                    if (Math.abs(ep[0] - 150) > 1 || Math.abs(ep[1] - 150) > 1) fail("EndPt=" + ep.join(",") + ", want 150,150");
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
