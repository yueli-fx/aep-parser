// flightdeck/showcase/layers/render.jsx — open layers.aep, render frame 0 →
// layers.png, and dump each layer's name + type flags so the non-rendering
// layers (Null / Adjustment) are provably present. Run via scripts/ae-worker/ae_run.ps1.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/layers/";
    var inAep = new File(dir + "layers.aep");
    var png = new File(dir + "layers.png");
    var done = new File(dir + "layers.done");
    var log = [];
    var ok = false;
    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "Layers") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp layers=" + comp.numLayers);
        for (var li = 1; li <= comp.numLayers; li++) {
            var L = comp.layer(li);
            var kind = "AV";
            try { if (L.nullLayer) kind = "Null"; else if (L.adjustmentLayer) kind = "Adjustment"; } catch (e0) {}
            log.push("  L" + li + " " + L.name + " [" + kind + "]");
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
