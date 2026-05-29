// tmp_debug/gen_ellipse_kf_re.jsx — animated Ellipse Size + Position fixture.
// RE whether Ellipse Position keyframes use the same non-spatial Vec2 layout
// as Size (→ injectAnimatedVec2 reuse). Saved to test_data/v2_2_ellipse_kf_re.aep.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_ellipse_kf_re.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_ellipse_kf_re.done");
    var log = [], ok = false;
    function linK(prop, times, values) {
        for (var i = 0; i < times.length; i++) prop.setValueAtTime(times[i], values[i]);
        for (var k = 1; k <= prop.numKeys; k++) prop.setInterpolationTypeAtKey(k, KeyframeInterpolationType.LINEAR, KeyframeInterpolationType.LINEAR);
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("EllKf", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "EllAnim";
        var sc = s.property("ADBE Root Vectors Group").addProperty("ADBE Vector Group").property("ADBE Vectors Group");
        var ell = sc.addProperty("ADBE Vector Shape - Ellipse");
        linK(ell.property("ADBE Vector Ellipse Size"), [0, 2], [[50, 50], [300, 200]]);
        linK(ell.property("ADBE Vector Ellipse Position"), [0, 2], [[10, 20], [70, 90]]);
        app.project.save(outFile);
        log.push("saved"); ok = true;
    } catch (e) { log.push("ERROR: " + e.toString()); }
    try { doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
