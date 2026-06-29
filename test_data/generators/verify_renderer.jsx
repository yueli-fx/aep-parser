// test_data/generators/verify_renderer.jsx
//
// SetRenderer ship-gate verifier. Opens a Go-built .aep (renderer fixture with
// SetRenderer applied), proves AE ACCEPTS it (no data-loss / corrupt), and
// reports the readback: the first comp's CompItem.renderer plus the full
// available renderers list and numItems. Resaves for byte inspection.
//
// PASS = file opened, a CompItem was found, comp.renderer is non-empty.
// The exact renderer string is reported (not hard-asserted here) because the
// match-name -> engine mapping is AE-version-dependent; the Go side records it.
//
// args.json fixed path: test_data/renderer_args.json — {"input","done","resaved","expected"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/renderer_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    try {
        argsFile.open("r");
        var s = argsFile.read();
        argsFile.close();
        args = eval("(" + s + ")");
        doneFile = new File(args.done);

        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (ePre) {}

        app.open(new File(args.input));
        log.push("numItems=" + app.project.numItems);

        var c = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem) { c = it; break; }
        }
        if (c === null) {
            log.push("FAIL: no CompItem found");
        } else {
            log.push("comp=" + c.name);
            log.push("renderer=" + c.renderer);
            log.push("expected=" + (args.expected || "(none)"));
            try { log.push("renderers=" + c.renderers.join("|")); } catch (eR) { log.push("renderers=<err " + eR.toString() + ">"); }
            ok = (c.renderer !== null && c.renderer !== "" && c.renderer !== undefined);
            if (args.expected) {
                log.push(c.renderer === args.expected ? "MATCH expected" : "DIFF expected");
            }
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
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/verify_renderer.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
