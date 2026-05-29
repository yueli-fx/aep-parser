// tmp_debug/gen_stroke_kf_re.jsx — V2.2.1 子项⑧ RE: Stroke Opacity + Width
// keyframe layout (the stroke body template already has both cdat slots; only
// need the keyframe encoding + whether Opacity is ÷100 like layer Opacity).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_stroke_kf_re.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_stroke_kf_re.done");
    var log = [], ok = false;
    function linearKf(prop, times, values) {
        for (var i = 0; i < times.length; i++) prop.setValueAtTime(times[i], values[i]);
        for (var k = 1; k <= prop.numKeys; k++) prop.setInterpolationTypeAtKey(k, KeyframeInterpolationType.LINEAR, KeyframeInterpolationType.LINEAR);
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("StrokeKfRE", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape(); s.name = "StrokeAnim";
        var sc = s.property("ADBE Root Vectors Group").addProperty("ADBE Vector Group").property("ADBE Vectors Group");
        sc.addProperty("ADBE Vector Shape - Rect").property("ADBE Vector Rect Size").setValue([200, 100]);
        var stroke = sc.addProperty("ADBE Vector Graphic - Stroke");
        linearKf(stroke.property("ADBE Vector Stroke Opacity"), [0, 2], [100, 50]);
        linearKf(stroke.property("ADBE Vector Stroke Width"), [0, 2], [5, 20]);
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) { log.push("ERROR: " + e.toString()); }
    try { doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
