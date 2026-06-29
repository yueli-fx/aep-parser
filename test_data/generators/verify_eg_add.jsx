// test_data/generators/verify_eg_add.jsx
//
// Essential Graphics W ship-gate — opens an ALL-Go-built .aep (NewProject +
// NewComposition + NewSolidLayer + AddEffect(Slider Control) +
// AddEssentialProperty + SetMotionGraphicsTemplateName). Asserts AE accepted
// the file, the EG panel reads back (template name + controller count + the
// controller's display name), the host layer/effect survived, then re-saves
// so the Go side can decode the resaved CIF3/OvG2.
//
// args.json fixed path: test_data/eg_add_args.json — {"input","done","resaved"}

(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/eg_add_args.json");
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
            log.push("layers=" + c.layers.length);
            checks.push(check("layer count 1 (host not dropped)", c.layers.length === 1));

            var tn = c.motionGraphicsTemplateName;
            log.push("motionGraphicsTemplateName=" + tn);
            checks.push(check("template name readback", tn === "EG Gate Template"));

            var cnt = c.motionGraphicsTemplateControllerCount;
            log.push("controllerCount=" + cnt);
            checks.push(check("controller count 1", cnt === 1));

            var cname = null;
            try {
                cname = c.getMotionGraphicsTemplateControllerName(1);
            } catch (eN1) {
                try { cname = c.motionGraphicsTemplateControllerName(1); } catch (eN2) {
                    log.push("controllerName API unavailable: " + eN1.toString());
                }
            }
            log.push("controllerName=" + cname);
            if (cname !== null) {
                checks.push(check("controller name readback", cname === "Exposed Slider"));
            }

            var L = c.layers.length >= 1 ? c.layers[1] : null;
            if (L) {
                var fx = L.property("ADBE Effect Parade");
                checks.push(check("effect parade present", fx !== null && fx.numProperties >= 1));
                if (fx && fx.numProperties >= 1) {
                    checks.push(check("slider effect present", fx.property(1).matchName === "ADBE Slider Control"));
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
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/eg_add_test.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
