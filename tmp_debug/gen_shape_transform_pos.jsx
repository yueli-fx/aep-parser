// tmp_debug/gen_shape_transform_pos.jsx
//
// V2.2.1 Layr Position keyframe (Path B) — produce a transform-group template
// source: 1 ShapeLayer (Rect + Fill, mirroring gen_shape_tolerance.jsx) whose
// COMBINED Transform Position is set to a static NON-default value so AE writes
// "ADBE Position" as a static cdat (dims NOT separated). The transform group is
// then extracted (tmp_debug/extract_transform_group, pointed at this file) into
// templates/ as the combined-Position template that injectAnimatedLayerPosition
// flips to a bpk-128 keyframe container.
//
// Run (AE 2025):
//   "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r \
//     "E:/projects/tools/aep-parser/tmp_debug/gen_shape_transform_pos.jsx"
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_transform_pos.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_transform_pos.done");
    var log = [];
    var ok = false;
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("TransformPos", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Nested";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var subContents = sub.property("ADBE Vectors Group");
        var rect = subContents.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
        var fill = subContents.addProperty("ADBE Vector Graphic - Fill");
        fill.property("ADBE Vector Fill Color").setValue([0.5, 0.5, 0.5, 1]);
        // Combined (non-separated) Position set to a static non-default value.
        s.property("ADBE Transform Group").property("ADBE Position").setValue([500, 300]);
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
