// showcase/transform-values/render.jsx — open transform_values.aep,
// dump each layer's Transform channel values (so the gen.go static values can be
// cross-checked against what AE actually read), then render frame 0 to
// transform_values.png for review. Run via scripts/ae-worker/ae_run.ps1.
(function () {
    var dir = (new File($.fileName).parent.fsName + "/");
    var inAep = new File(dir + "transform_values.aep");
    var png = new File(dir + "transform_values.png");
    var done = new File(dir + "transform_values.done");
    var log = [];
    var ok = false;

    function fmtVal(v) {
        if (v instanceof Array) {
            var parts = [];
            for (var i = 0; i < v.length; i++) parts.push(Math.round(v[i] * 100) / 100);
            return "[" + parts.join(",") + "]";
        }
        return Math.round(v * 100) / 100;
    }

    try {
        app.open(inAep);
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === "TransformValuesShowcase") { comp = it; break; }
        }
        if (!comp) throw new Error("comp not found");
        log.push("comp " + comp.width + "x" + comp.height + " layers=" + comp.numLayers);

        // Dump each layer's Transform channel static values (for value verification).
        for (var li = 1; li <= comp.numLayers; li++) {
            var lyr = comp.layer(li);
            var tg = lyr.property("ADBE Transform Group");
            if (!tg) { log.push(lyr.name + " :: (no transform group)"); continue; }
            var pos = tg.property("ADBE Position");
            var scl = tg.property("ADBE Scale");
            var rot = tg.property("ADBE Rotate Z");
            var opa = tg.property("ADBE Opacity");
            var anc = tg.property("ADBE Anchor Point");
            log.push(lyr.name +
                " :: pos=" + (pos ? fmtVal(pos.value) : "?") +
                " scale=" + (scl ? fmtVal(scl.value) : "?") +
                " rot=" + (rot ? fmtVal(rot.value) : "?") +
                " opacity=" + (opa ? fmtVal(opa.value) : "?") +
                " anchor=" + (anc ? fmtVal(anc.value) : "?"));
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
