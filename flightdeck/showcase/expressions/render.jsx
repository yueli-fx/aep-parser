// flightdeck/showcase/expressions/render.jsx — open expressions.aep, render t=1s
// to expressions.png so each time*N rotation expression has spun its dot to a
// distinct angle (90°/180°/270°). Run via scripts/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/expressions/";
    var inAep = new File(dir + "expressions.aep");
    var png = new File(dir + "expressions.png");
    var done = new File(dir + "expressions.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "Expressions") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp layers=" + comp.numLayers);
        app.project.bitsPerChannel = 8;
        comp.saveFrameToPng(1.0, png); // t=1s
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
