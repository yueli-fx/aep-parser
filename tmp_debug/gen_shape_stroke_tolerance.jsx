// tmp_debug/gen_shape_stroke_tolerance.jsx — AE-native Stroke reference.
// 1 ShapeLayer + 1 Rect (so the stroke has geometry) + 1 Stroke with DISTINCT
// Color/Width/Opacity (forces AE to persist their cdat slots) → saved by AE to
// test_data/v2_2_shape_stroke_tolerance.aep for structural RE + embed-body
// extraction (V2.2.1 Stroke embed bytes).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_stroke_tolerance.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_stroke_tolerance.done");
    var log = [];
    var ok = false;
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("StrokeTolerance", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Stroke_Static";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var sc = sub.property("ADBE Vectors Group");
        var rect = sc.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
        var stroke = sc.addProperty("ADBE Vector Graphic - Stroke");
        stroke.property("ADBE Vector Stroke Color").setValue([0, 0, 1, 1]);
        stroke.property("ADBE Vector Stroke Width").setValue(7);
        stroke.property("ADBE Vector Stroke Opacity").setValue(80);
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
