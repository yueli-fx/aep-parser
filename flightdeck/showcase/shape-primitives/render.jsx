// flightdeck/showcase/shape-primitives/render.jsx — open shape_primitives.aep,
// verify AE accepts it, render frame 0 to shape_primitives.png for review.
// Run via scripts/ae_run.ps1 (writes the .done marker).
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/shape-primitives/";
    var inAep = new File(dir + "shape_primitives.aep");
    var png = new File(dir + "shape_primitives.png");
    var done = new File(dir + "shape_primitives.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "ShapePrimitives") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp layers=" + comp.numLayers);
        app.project.bitsPerChannel = 8;
        comp.saveFrameToPng(0, png);
        $.sleep(2500);
        log.push(png.exists ? ("png " + png.length + " bytes") : "PNG MISSING");
        ok = png.exists;
    } catch (e) { log.push("ERROR: " + e.toString() + " line=" + e.line); }
    done.open("w");
    done.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
    done.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
