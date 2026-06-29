// V2.2 Gradient-fill ship-gate — opens an .aep with NewShapeLayer + AddRect +
// AddGradientFill, asserts AE accepted the gradient fill (no silent-drop), then
// re-saves so the Go side decodes the re-saved Grad Colors stops XML.
//
// args.json fixed path: test_data/v2_2_gradient_gate_args.json — {"input","done","resaved"}
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/v2_2_gradient_gate_args.json");
    var args = {};
    var doneFile = null;
    var log = [];
    var ok = false;

    function check(name, cond) {
        if (!cond) { log.push("FAIL: " + name); return false; }
        log.push("OK:   " + name);
        return true;
    }
    function layerByName(comp, n) {
        for (var i = 1; i <= comp.layers.length; i++) {
            if (comp.layers[i].name === n) return comp.layers[i];
        }
        return null;
    }
    function findByMatchName(group, matchName) {
        if (!group || !group.numProperties) return null;
        for (var i = 1; i <= group.numProperties; i++) {
            var p = group.property(i);
            if (p.matchName === matchName) return p;
            if (p.numProperties && p.numProperties > 0) {
                var hit = findByMatchName(p, matchName);
                if (hit) return hit;
            }
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
        var c = app.project.items[1];
        log.push("opened comp: " + c.name + " layers=" + c.layers.length);

        var L = layerByName(c, "Grad_Static");
        var checks = [
            check("layer count 1", c.layers.length === 1),
            check("Grad_Static layer exists (not dropped)", L !== null)
        ];
        if (L) {
            var root = L.property("ADBE Root Vectors Group");
            var rect = findByMatchName(root, "ADBE Vector Shape - Rect");
            var gfill = findByMatchName(root, "ADBE Vector Graphic - G-Fill");
            checks.push(check("Rect present", rect !== null));
            checks.push(check("Gradient fill present", gfill !== null));
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
        if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_gradient_gate.done");
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}

    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (eClose) {}
    try { app.quit(); } catch (eQuit) {}
})();
