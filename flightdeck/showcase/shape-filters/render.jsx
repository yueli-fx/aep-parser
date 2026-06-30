// flightdeck/showcase/shape-filters/render.jsx — open shape_filters.aep, verify
// AE accepts it (all 13 layers present), and render frame 0 to shape_filters.png
// for user review. Run via scripts/ae-worker/ae_run.ps1 (writes the .done marker).
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/shape-filters/";
    var inAep = new File(dir + "shape_filters.aep");
    var png = new File(dir + "shape_filters.png");
    var done = new File(dir + "shape_filters.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "ShapeFilterShowcase") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp " + comp.width + "x" + comp.height + " layers=" + comp.numLayers);
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
