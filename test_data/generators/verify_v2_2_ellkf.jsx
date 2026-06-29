// test_data/generators/verify_v2_2_ellkf.jsx — V2.2.1 Ellipse Size+Position keyframe gate.
// Asserts AE accepts an Ellipse with animated Size + Position, re-saves for
// Go-side keyframe decode. args: v2_2_ellkf_args.json.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/generated/args/v2_2_ellkf_args.json");
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
        var L = layerByName(c, "EllAnim");
        var checks = [check("1 layer", c.layers.length === 1), check("EllAnim exists", L !== null)];
        if (L) {
            var ell = findBy(L.property("ADBE Root Vectors Group"), "ADBE Vector Shape - Ellipse");
            checks.push(check("Ellipse present", ell !== null));
            if (ell) {
                checks.push(check("Size 2 kf", ell.property("ADBE Vector Ellipse Size").numKeys === 2));
                checks.push(check("Position 2 kf", ell.property("ADBE Vector Ellipse Position").numKeys === 2));
            }
        }
        ok = true; for (var k = 0; k < checks.length; k++) if (!checks[k]) { ok = false; break; }
        if (ok && args.resaved) { app.project.save(new File(args.resaved)); log.push("resaved"); }
    } catch (e) { ok = false; log.push("ERROR: " + e.toString()); }
    try { if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_ellkf_test.done"); doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
