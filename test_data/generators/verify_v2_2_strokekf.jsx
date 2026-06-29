// test_data/generators/verify_v2_2_strokekf.jsx — V2.2.1 子项⑧ Stroke Opacity + Width
// keyframe gate. Asserts AE accepts animated Stroke Opacity (raw %) + Width
// (raw px), both 1D non-spatial (bpk-48), reads back values, re-saves.
// args: v2_2_strokekf_args.json.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/v2_2_strokekf_args.json");
    var args = {}, doneFile = null, log = [], ok = false;
    function check(n, c) { if (!c) { log.push("FAIL: " + n); return false; } log.push("OK:   " + n); return true; }
    function layerByName(comp, n) { for (var i = 1; i <= comp.layers.length; i++) if (comp.layers[i].name === n) return comp.layers[i]; return null; }
    function findBy(g, mn) { if (!g || !g.numProperties) return null; for (var i = 1; i <= g.numProperties; i++) { var p = g.property(i); if (p.matchName === mn) return p; if (p.numProperties > 0) { var h = findBy(p, mn); if (h) return h; } } return null; }
    try {
        argsFile.open("r"); var s = argsFile.read(); argsFile.close(); args = eval("(" + s + ")"); doneFile = new File(args.done);
        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e0) {}
        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("comp " + c.name + " layers=" + c.layers.length);
        var L = layerByName(c, "StrokeAnim");
        var checks = [check("1 layer", c.layers.length === 1), check("StrokeAnim exists", L !== null)];
        if (L) {
            var stroke = findBy(L.property("ADBE Root Vectors Group"), "ADBE Vector Graphic - Stroke");
            checks.push(check("Stroke present", stroke !== null));
            if (stroke) {
                var opa = stroke.property("ADBE Vector Stroke Opacity");
                var wid = stroke.property("ADBE Vector Stroke Width");
                checks.push(check("Opacity 2 kf", opa.numKeys === 2));
                checks.push(check("Width 2 kf", wid.numKeys === 2));
                if (opa.numKeys === 2) { var o1 = opa.keyValue(2); log.push("opa kf1=" + o1); checks.push(check("opacity~50", Math.abs(o1 - 50) < 0.5)); }
                if (wid.numKeys === 2) { var w1 = wid.keyValue(2); log.push("wid kf1=" + w1); checks.push(check("width~20", Math.abs(w1 - 20) < 0.5)); }
            }
        }
        ok = true; for (var k = 0; k < checks.length; k++) if (!checks[k]) { ok = false; break; }
        if (ok && args.resaved) { app.project.save(new File(args.resaved)); log.push("resaved"); }
    } catch (e) { ok = false; log.push("ERROR: " + e.toString()); }
    try { if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_strokekf_test.done"); doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
