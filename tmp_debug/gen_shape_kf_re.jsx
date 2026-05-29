// tmp_debug/gen_shape_kf_re.jsx — animated-property RE fixture. One ShapeLayer
// with Rect Size + Fill Color animated (2 linear keyframes each) and the layer
// Position animated (2 linear keyframes). Saved by AE → test_data/v2_2_shape_kf_re.aep
// to RE how AE stores animated shape/transform tdbs (lhd3/ldat keyframe encoding).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_kf_re.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_kf_re.done");
    var log = [];
    var ok = false;
    function linearKf(prop, times, values) {
        for (var i = 0; i < times.length; i++) prop.setValueAtTime(times[i], values[i]);
        for (var k = 1; k <= prop.numKeys; k++) {
            prop.setInterpolationTypeAtKey(k, KeyframeInterpolationType.LINEAR, KeyframeInterpolationType.LINEAR);
        }
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("KfRE", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Anim";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var sc = sub.property("ADBE Vectors Group");
        var rect = sc.addProperty("ADBE Vector Shape - Rect");
        linearKf(rect.property("ADBE Vector Rect Size"), [0, 2], [[50, 50], [300, 200]]);
        var fill = sc.addProperty("ADBE Vector Graphic - Fill");
        linearKf(fill.property("ADBE Vector Fill Color"), [0, 2], [[1, 0, 0, 1], [0, 0, 1, 1]]);
        linearKf(s.property("ADBE Transform Group").property("ADBE Position"), [0, 2], [[0, 0], [500, 300]]);
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) { log.push("ERROR: " + e.toString()); }
    try {
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
