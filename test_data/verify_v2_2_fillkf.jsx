// test_data/verify_v2_2_fillkf.jsx — V2.2.1 Fill-Color-keyframe ship-gate.
// Opens NewShapeLayer + AddRect + AddFill(Color animated, 2 kf), asserts AE
// accepts it, re-saves for Go-side keyframe decode. args: v2_2_fillkf_args.json.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_fillkf_args.json");
    var args = {}, doneFile = null, log = [], ok = false;
    function check(n, c) { if (!c) { log.push("FAIL: " + n); return false; } log.push("OK:   " + n); return true; }
    function layerByName(comp, n) { for (var i = 1; i <= comp.layers.length; i++) if (comp.layers[i].name === n) return comp.layers[i]; return null; }
    function findBy(g, mn) {
        if (!g || !g.numProperties) return null;
        for (var i = 1; i <= g.numProperties; i++) { var p = g.property(i); if (p.matchName === mn) return p; if (p.numProperties > 0) { var h = findBy(p, mn); if (h) return h; } }
        return null;
    }
    try {
        argsFile.open("r"); var s = argsFile.read(); argsFile.close(); args = eval("(" + s + ")"); doneFile = new File(args.done);
        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e0) {}
        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("comp " + c.name + " layers=" + c.layers.length);
        var L = layerByName(c, "FillKf");
        var checks = [check("1 layer", c.layers.length === 1), check("FillKf exists", L !== null)];
        if (L) {
            var fill = findBy(L.property("ADBE Root Vectors Group"), "ADBE Vector Graphic - Fill");
            checks.push(check("Fill present", fill !== null));
            if (fill) checks.push(check("Fill Color 2 keyframes", fill.property("ADBE Vector Fill Color").numKeys === 2));
        }
        ok = true; for (var k = 0; k < checks.length; k++) if (!checks[k]) { ok = false; break; }
        if (ok && args.resaved) { app.project.save(new File(args.resaved)); log.push("resaved"); }
    } catch (e) { ok = false; log.push("ERROR: " + e.toString()); }
    try { if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_fillkf_test.done"); doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
