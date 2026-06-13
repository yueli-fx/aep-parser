// flightdeck/showcase/effects/render.jsx — open effects.aep, dump each layer's
// effect + parameter match-names (so the gen.go param indices can be verified),
// then render frame 0 to effects.png for review. Run via scripts/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/effects/";
    var inAep = new File(dir + "effects.aep");
    var png = new File(dir + "effects.png");
    var done = new File(dir + "effects.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "EffectsShowcase") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp " + comp.width + "x" + comp.height + " layers=" + comp.numLayers);
        // Dump effects + parameter match-names per layer (for param-index verification).
        for (var li = 1; li <= comp.numLayers; li++) {
            var lyr = comp.layer(li);
            var fxGroup = lyr.property("ADBE Effect Parade");
            if (fxGroup && fxGroup.numProperties > 0) {
                for (var fi = 1; fi <= fxGroup.numProperties; fi++) {
                    var fx = fxGroup.property(fi);
                    var pnames = [];
                    for (var pi = 1; pi <= fx.numProperties; pi++) {
                        var pr = fx.property(pi);
                        pnames.push(pr.matchName + "(" + pr.name + ")");
                    }
                    log.push(lyr.name + " :: " + fx.matchName + " | " + pnames.join(", "));
                }
            }
        }
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
