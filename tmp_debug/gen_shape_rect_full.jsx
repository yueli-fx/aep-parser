// tmp_debug/gen_shape_rect_full.jsx — V2.2.1 子项⑦ richer rect body template
// source. ONE shape layer, one Rect whose Size / Position / Roundness are all
// set to static non-default values so AE emits each as a flippable cdat (default
// Position/Roundness get elided — the old rect template had only Size).
// Extracted via extract_shape_bodies → templates/v2_2_shape_rect_body.bin.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_rect_full.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_rect_full.done");
    var log = [], ok = false;
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("RectFull", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape(); s.name = "Nested";
        var sc = s.property("ADBE Root Vectors Group").addProperty("ADBE Vector Group").property("ADBE Vectors Group");
        var rect = sc.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
        rect.property("ADBE Vector Rect Position").setValue([40, 50]);
        rect.property("ADBE Vector Rect Roundness").setValue(15);
        var fill = sc.addProperty("ADBE Vector Graphic - Fill");
        fill.property("ADBE Vector Fill Color").setValue([0.5, 0.5, 0.5, 1]);
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) { log.push("ERROR: " + e.toString()); }
    try { doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
