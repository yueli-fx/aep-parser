// tmp_debug/gen_transform_kf_re.jsx — V2.2.1 子项⑥ RE fixture. One ShapeLayer
// whose Transform Anchor Point / Scale / Rotation / Opacity are each animated
// (2 linear keyframes) so AE writes their lhd3/ldat keyframe encodings. Dumped
// via tmp_debug/dump_kf to RE per-channel byte layout (spatial vs non-spatial,
// bpk, value offset).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_transform_kf_re.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_transform_kf_re.done");
    var log = [], ok = false;
    function linearKf(prop, times, values) {
        for (var i = 0; i < times.length; i++) prop.setValueAtTime(times[i], values[i]);
        for (var k = 1; k <= prop.numKeys; k++) prop.setInterpolationTypeAtKey(k, KeyframeInterpolationType.LINEAR, KeyframeInterpolationType.LINEAR);
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("XfKfRE", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "XfAnim";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var sc = sub.property("ADBE Vectors Group");
        var rect = sc.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
        var tg = s.property("ADBE Transform Group");
        linearKf(tg.property("ADBE Anchor Point"), [0, 2], [[10, 20], [50, 60]]);
        linearKf(tg.property("ADBE Scale"), [0, 2], [[100, 100], [150, 200]]);
        linearKf(tg.property("ADBE Rotate Z"), [0, 2], [0, 90]);
        linearKf(tg.property("ADBE Opacity"), [0, 2], [100, 50]);
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) { log.push("ERROR: " + e.toString()); }
    try { doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
