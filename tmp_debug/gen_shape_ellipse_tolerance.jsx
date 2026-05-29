// tmp_debug/gen_shape_ellipse_tolerance.jsx
//
// V2.2.1 Ellipse embed-bytes — produce v2_2_shape_ellipse_tolerance.aep.
// Creates 1 ShapeLayer with a nested VectorGroup containing 1 Ellipse + 1 Fill.
// Ellipse Size AND Position are set to non-default values so AE persists both
// cdat slots (default-valued slots get elided → nothing to overwrite later).
// Source for extract_shape_bodies → internal/aep/templates/v2_2_shape_ellipse_body.bin.
//
// Run (AE 2025 recommended):
//   "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r \
//     "E:/projects/tools/aep-parser/tmp_debug/gen_shape_ellipse_tolerance.jsx"
//
// Wait for test_data/v2_2_shape_ellipse_tolerance.done — first line "PASS" or "FAIL".

(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_ellipse_tolerance.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_ellipse_tolerance.done");
    var log = [];
    var ok = false;

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();

        var c = app.project.items.addComp("EllipseTolerance", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "EllipseNested";

        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var subContents = sub.property("ADBE Vectors Group");
        var ell = subContents.addProperty("ADBE Vector Shape - Ellipse");
        ell.property("ADBE Vector Ellipse Size").setValue([200, 100]);
        ell.property("ADBE Vector Ellipse Position").setValue([50, 30]);
        var fill = subContents.addProperty("ADBE Vector Graphic - Fill");
        fill.property("ADBE Vector Fill Color").setValue([0.5, 0.5, 0.5, 1]);

        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) {
        log.push("ERROR: " + e.toString());
    }

    try {
        doneFile.open("w");
        doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n"));
        doneFile.close();
    } catch (e2) { /* swallow */ }

    // Tear down AE cleanly so it doesn't prompt about unsaved changes next launch.
    try {
        if (app.project) {
            app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        }
    } catch (eClose) { /* swallow */ }
    try {
        app.quit();
    } catch (eQuit) { /* swallow */ }
})();
