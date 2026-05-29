// test_data/verify_v2_2_xfkf.jsx — V2.2.1 子项⑥ Layr Transform Anchor/Scale/
// Rotation/Opacity keyframe gate. Asserts AE accepts the animated transform
// channels and reads back user-facing values (AE units: scale %, opacity %,
// rotation degrees), then re-saves for Go-side keyframe decode.
// args: v2_2_xfkf_args.json.
(function () {
    var argsFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_xfkf_args.json");
    var args = {}, doneFile = null, log = [], ok = false;
    function check(n, c) { if (!c) { log.push("FAIL: " + n); return false; } log.push("OK:   " + n); return true; }
    function layerByName(comp, n) { for (var i = 1; i <= comp.layers.length; i++) if (comp.layers[i].name === n) return comp.layers[i]; return null; }
    try {
        argsFile.open("r"); var s = argsFile.read(); argsFile.close(); args = eval("(" + s + ")"); doneFile = new File(args.done);
        try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e0) {}
        app.open(new File(args.input));
        var c = app.project.items[1];
        log.push("comp " + c.name + " layers=" + c.layers.length);
        var L = layerByName(c, "XfAnim");
        var checks = [check("1 layer", c.layers.length === 1), check("XfAnim exists", L !== null)];
        if (L) {
            var tg = L.property("ADBE Transform Group");
            var anc = tg.property("ADBE Anchor Point");
            var scl = tg.property("ADBE Scale");
            var rot = tg.property("ADBE Rotate Z");
            var opa = tg.property("ADBE Opacity");
            checks.push(check("Anchor 2 kf", anc.numKeys === 2));
            checks.push(check("Scale 2 kf", scl.numKeys === 2));
            checks.push(check("Rotation 2 kf", rot.numKeys === 2));
            checks.push(check("Opacity 2 kf", opa.numKeys === 2));
            if (anc.numKeys === 2) {
                var a1 = anc.keyValue(2); log.push("anchor kf1=" + a1.toString());
                checks.push(check("anchor X~50", Math.abs(a1[0] - 50) < 0.5));
                checks.push(check("anchor Y~60", Math.abs(a1[1] - 60) < 0.5));
            }
            if (scl.numKeys === 2) {
                var s1 = scl.keyValue(2); log.push("scale kf1=" + s1.toString());
                checks.push(check("scale X~150", Math.abs(s1[0] - 150) < 0.5));
                checks.push(check("scale Y~200", Math.abs(s1[1] - 200) < 0.5));
            }
            if (rot.numKeys === 2) {
                var r1 = rot.keyValue(2); log.push("rot kf1=" + r1.toString());
                checks.push(check("rotation~90", Math.abs(r1 - 90) < 0.5));
            }
            if (opa.numKeys === 2) {
                var o1 = opa.keyValue(2); log.push("opa kf1=" + o1.toString());
                checks.push(check("opacity~50", Math.abs(o1 - 50) < 0.5));
            }
        }
        ok = true; for (var k = 0; k < checks.length; k++) if (!checks[k]) { ok = false; break; }
        if (ok && args.resaved) { app.project.save(new File(args.resaved)); log.push("resaved"); }
    } catch (e) { ok = false; log.push("ERROR: " + e.toString()); }
    try { if (!doneFile) doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_xfkf_test.done"); doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
