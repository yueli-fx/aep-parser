// tmp_debug/gen_shape_transform_full.jsx — V2.2.1 子项⑥ template source. A shape
// layer whose Transform Anchor / Position / Scale / Rotation / Opacity are all
// set to STATIC NON-default values so AE emits each as a cdat (default values
// get elided). Extracted (extract_transform_group) into the richer
// transform-group template that lowerLayerTransform overwrites / flips per
// channel. Also feeds static-cdat-encoding RE (dump per channel).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_transform_full.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_transform_full.done");
    var log = [], ok = false;
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("TransformFull", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Nested";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var sc = sub.property("ADBE Vectors Group");
        var rect = sc.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
        var fill = sc.addProperty("ADBE Vector Graphic - Fill");
        fill.property("ADBE Vector Fill Color").setValue([0.5, 0.5, 0.5, 1]);
        var tg = s.property("ADBE Transform Group");
        tg.property("ADBE Anchor Point").setValue([30, 40]);
        tg.property("ADBE Position").setValue([500, 300]);
        tg.property("ADBE Scale").setValue([120, 130]);
        tg.property("ADBE Rotate Z").setValue(25);
        tg.property("ADBE Opacity").setValue(80);
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) { log.push("ERROR: " + e.toString()); }
    try { doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
