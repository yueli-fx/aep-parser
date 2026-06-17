// test_data/verify_layer_av_flags.jsx
//
// Layer AV-flags ship-gate (batch 1 of roundtrip→ae-accept backfill). Opens an
// .aep where Go set 7 flag/enum setters on a PARSED solid layer after a Reopen
// round-trip. Asserts AE reads each flag back through the DOM, then re-saves for
// Go-side verification.
//
// args.json fixed path: test_data/layer_av_flags_args.json — {"input","done","resaved"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/layer_av_flags_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
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
                log.push("enabled=" + L.enabled + " shy=" + L.shy + " solo=" + L.solo +
                         " locked=" + L.locked + " motionBlur=" + L.motionBlur);
                log.push("quality=" + L.quality + " blendingMode=" + L.blendingMode);
                checks.push(check("enabled false (SetVisible)", L.enabled === false));
                checks.push(check("shy true", L.shy === true));
                checks.push(check("solo true", L.solo === true));
                checks.push(check("locked true", L.locked === true));
                checks.push(check("motionBlur true", L.motionBlur === true));
                checks.push(check("quality DRAFT", L.quality === LayerQuality.DRAFT));
                checks.push(check("blendingMode MULTIPLY", L.blendingMode === BlendingMode.MULTIPLY));
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
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/layer_av_flags_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
