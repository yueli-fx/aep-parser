// tmp_debug/gen_shape_path_tolerance.jsx — AE-native Path reference.
// 1 ShapeLayer + 1 Path (ADBE Vector Shape - Group) with a known 4-vertex
// closed square + 1 Fill. Saved by AE to test_data/v2_2_shape_path_tolerance.aep
// for structural RE + embed-body extraction (V2.2.1 Path embed bytes).
//
// Run: AfterFX.exe -r <this>. Wait for .done first line PASS/FAIL.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_path_tolerance.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_path_tolerance.done");
    var log = [];
    var ok = false;
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("PathTolerance", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Path_Static";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var subContents = sub.property("ADBE Vectors Group");
        var pathGroup = subContents.addProperty("ADBE Vector Shape - Group");
        var shapeProp = pathGroup.property("ADBE Vector Shape");
        var sh = new Shape();
        sh.vertices = [[0, 0], [100, 0], [100, 100], [0, 100]];
        sh.inTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.outTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.closed = true;
        shapeProp.setValue(sh);
        var fill = subContents.addProperty("ADBE Vector Graphic - Fill");
        fill.property("ADBE Vector Fill Color").setValue([0.5, 0.5, 0.5, 1]);
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
