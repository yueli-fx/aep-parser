// test_data/verify_v2_2_layrposkf.jsx — V2.2.1 Layr Transform Position keyframe
// gate (Path B: combined "ADBE Position", bpk-128 spatial dim-3). Asserts AE
// accepts a ShapeLayer with an animated Transform Position, re-saves for
// Go-side keyframe decode. args: v2_2_layrposkf_args.json.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_layrposkf_args.json");
    var args = {}, doneFile = null, log = [], ok = false;
    function check(n, c) { if (!c) { log.push("FAIL: " + n); return false; } log.push("OK:   " + n); return true; }
    function layerByName(comp, n) { for (var i = 1; i <= comp.layers.length; i++) if (comp.layers[i].name === n) return comp.layers[i]; return null; }
    try {
        argsFile.open("r"); var s = argsFile.read(); argsFile.close(); args = eval("(" + s + ")"); doneFile = new File(args.done);
        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e0) {}
        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("comp " + c.name + " layers=" + c.layers.length);
        var L = layerByName(c, "PosAnim");
        var checks = [check("1 layer", c.layers.length === 1), check("PosAnim exists", L !== null)];
        if (L) {
            var pos = L.property("ADBE Transform Group").property("ADBE Position");
            checks.push(check("Position present", pos !== null));
            if (pos) {
                checks.push(check("Position 2 kf", pos.numKeys === 2));
                if (pos.numKeys === 2) {
                    var v0 = pos.keyValue(1), v1 = pos.keyValue(2);
                    log.push("kf0=" + v0.toString() + " kf1=" + v1.toString());
                    checks.push(check("kf0 X~100", Math.abs(v0[0] - 100) < 0.5));
                    checks.push(check("kf0 Y~200", Math.abs(v0[1] - 200) < 0.5));
                    checks.push(check("kf1 X~700", Math.abs(v1[0] - 700) < 0.5));
                    checks.push(check("kf1 Y~400", Math.abs(v1[1] - 400) < 0.5));
                }
            }
        }
        ok = true; for (var k = 0; k < checks.length; k++) if (!checks[k]) { ok = false; break; }
        if (ok && args.resaved) { app.project.save(new File(args.resaved)); log.push("resaved"); }
    } catch (e) { ok = false; log.push("ERROR: " + e.toString()); }
    try { if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_layrposkf_test.done"); doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
