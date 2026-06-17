// test_data/verify_layer_av_fields2.jsx
//
// Layer AV-fields ship-gate (batch 2 of roundtrip→ae-accept backfill). Opens an
// .aep where Go set 7 setters on a PARSED solid layer after a Reopen round-trip:
// in/out point, preserveTransparency, sampling, guide, adjustment, label.
// Asserts AE reads each back through the DOM, then re-saves for Go-side verify.
//
// args.json fixed path: test_data/layer_av_fields2_args.json — {"input","done","resaved"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/layer_av_fields2_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
    function near(a, b) { return Math.abs(a - b) < 0.01; }
    function compByName(n) {
        for (var i = 1; i <= app.project.items.length; i++) {
            var it = app.project.items[i];
            if (it instanceof CompItem && it.name === n) return it;
        }
        return null;
    }

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (ePre) {}

        app.open(new File(args.input));
        var c = compByName("Main");
        var checks = [check("comp Main exists", c !== null)];

        if (c) {
            checks.push(check("layer count 1 (not dropped)", c.layers.length === 1));
            var L = c.layers.length >= 1 ? c.layers[1] : null;
            if (L) {
                log.push("inPoint=" + L.inPoint + " outPoint=" + L.outPoint +
                         " preserveTransparency=" + L.preserveTransparency);
                log.push("samplingQuality=" + L.samplingQuality + " guideLayer=" + L.guideLayer +
                         " adjustmentLayer=" + L.adjustmentLayer + " label=" + L.label);
                checks.push(check("inPoint 1.0", near(L.inPoint, 1.0)));
                checks.push(check("outPoint 3.0", near(L.outPoint, 3.0)));
                checks.push(check("preserveTransparency true", L.preserveTransparency === true));
                checks.push(check("samplingQuality BICUBIC", L.samplingQuality === LayerSamplingQuality.BICUBIC));
                checks.push(check("guideLayer true", L.guideLayer === true));
                checks.push(check("adjustmentLayer true", L.adjustmentLayer === true));
                checks.push(check("label 9", L.label === 9));
            }
        }

        ok = true;
        for (var k = 0; k < checks.length; k++) {
            if (!checks[k]) { ok = false; break; }
        }

        if (ok && args.resaved) {
            app.project.save(new File(args.resaved));
            log.push("resaved to " + args.resaved);
        }
    } catch (e) {
        ok = false;
        log.push("ERROR: " + e.toString());
    }

    try {
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/layer_av_fields2_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
