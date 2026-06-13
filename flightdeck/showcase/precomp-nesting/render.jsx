// flightdeck/showcase/precomp-nesting/render.jsx — open precomp_nesting.aep,
// render the parent comp frame 0 → precomp_nesting.png, showing the nested Badge
// comp reused 3×. Run via scripts/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/precomp-nesting/";
    var inAep = new File(dir + "precomp_nesting.aep");
    var png = new File(dir + "precomp_nesting.png");
    var done = new File(dir + "precomp_nesting.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "PrecompMain") { comp = it; break; }
        }
        if (!comp) throw new Error("comp PrecompMain not found");
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
