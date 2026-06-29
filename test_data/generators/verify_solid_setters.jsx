// test_data/generators/verify_solid_setters.jsx
//
// Solid setters ship-gate — opens an .aep where Go mutated a PARSED solid
// (Footage.SetSolidColor to [0.1, 0.6, 0.9] + SetSolidSize to 640x360 after a
// Reopen round-trip). Asserts AE reads the new color and dimensions through
// the scripting API, then re-saves for Go-side verification.
//
// args.json fixed path: test_data/solid_setters_args.json — {"input","done","resaved"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/solid_setters_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
    function near(a, b) { return Math.abs(a - b) < 0.005; }
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
            checks.push(check("layer count 1 (solid not dropped)", c.layers.length === 1));
            var L = c.layers.length >= 1 ? c.layers[1] : null;
            var src = L ? L.source : null;
            checks.push(check("layer has FootageItem source", src !== null && src instanceof FootageItem));
            if (src) {
                log.push("source dims=" + src.width + "x" + src.height);
                checks.push(check("solid width 640", src.width === 640));
                checks.push(check("solid height 360", src.height === 360));
                var ms = src.mainSource;
                checks.push(check("mainSource is SolidSource", ms !== null && ms instanceof SolidSource));
                if (ms instanceof SolidSource) {
                    var col = ms.color;
                    log.push("color=[" + col[0] + "," + col[1] + "," + col[2] + "]");
                    checks.push(check("color R 0.1", near(col[0], 0.1)));
                    checks.push(check("color G 0.6", near(col[1], 0.6)));
                    checks.push(check("color B 0.9", near(col[2], 0.9)));
                }
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
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/solid_setters_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
