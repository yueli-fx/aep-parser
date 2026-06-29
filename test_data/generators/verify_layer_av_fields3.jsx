// test_data/generators/verify_layer_av_fields3.jsx
//
// Layer AV-fields ship-gate (batch 3 of roundtrip→ae-accept backfill). Opens an
// .aep with two solids where Go renamed FG (length-variable), set its comment
// (length-variable), shifted startTime, and parented it to BG. Asserts AE reads
// each back through the DOM (incl. the parent layer ref), then re-saves.
//
// args.json fixed path: test_data/layer_av_fields3_args.json — {"input","done","resaved"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/layer_av_fields3_args.json");
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
    function layerByName(c, n) {
        for (var i = 1; i <= c.layers.length; i++) {
            if (c.layers[i].name === n) return c.layers[i];
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
            checks.push(check("layer count 2 (none dropped)", c.layers.length === 2));
            var bg = layerByName(c, "BG");
            var fg = layerByName(c, "FG_Renamed");
            checks.push(check("BG present", bg !== null));
            checks.push(check("FG_Renamed present (SetName)", fg !== null));
            if (fg && bg) {
                log.push("fg.name=" + fg.name + " comment=" + fg.comment +
                         " startTime=" + fg.startTime +
                         " parent=" + (fg.parent ? fg.parent.name : "null"));
                checks.push(check("comment batch3_note", fg.comment === "batch3_note"));
                checks.push(check("startTime 0.5", near(fg.startTime, 0.5)));
                checks.push(check("parent is BG", fg.parent !== null && fg.parent.name === "BG"));
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
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/layer_av_fields3_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
