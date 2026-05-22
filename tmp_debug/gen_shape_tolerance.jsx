// tmp_debug/gen_shape_tolerance.jsx
//
// Phase 5 Task 5.3 — produce v2_2_shape_tolerance.aep fixture.
// Creates 1 ShapeLayer with a nested VectorGroup containing 1 Rect + 1 Fill.
// Saves to test_data/v2_2_shape_tolerance.aep — feeds Tier 3 preservation
// tests (TestV2_2_NestedGroup_Preservation, TestV2_2_OpaquePreservation_*).
//
// Run (AE 2025 recommended):
//   "E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe" -r \
//     "E:/projects/tools/aep-parser/tmp_debug/gen_shape_tolerance.jsx"
//
// Wait for test_data/v2_2_shape_tolerance.done — first line "PASS" or "FAIL".

(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_tolerance.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_tolerance.done");
    var log = [];
    var ok = false;

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();

        var c = app.project.items.addComp("ToleranceMain", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Nested";

        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var subContents = sub.property("ADBE Vectors Group");
        var rect = subContents.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
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

    // Tear down AE cleanly so it doesn't prompt user about unsaved
    // changes on the next launch. Close project (no save — we already
    // saved via project.save above) then quit.
    try {
        if (app.project) {
            app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        }
    } catch (eClose) { /* swallow */ }
    try {
        app.quit();
    } catch (eQuit) { /* swallow */ }
})();
