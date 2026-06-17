// test_data/re_multi_shape_layer.jsx
//
// Baseline for "multiple ShapeLayers in one comp". AE builds 3 plain shape
// layers (each with a Rect+Fill) in a single comp and saves. Go side then
// byte-diffs each Layr's ldta against our NewShapeLayer output to find which
// field AE sets per-layer that we get wrong (causing AE to silent-drop
// layers 2+ from comp.layers).
// AE 2020 — nothing here is an AE 24+ field.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_multi_shape_layer.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_multi_shape_layer.done");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    step("build_three_shape_layers", function () {
        var c = app.project.items.addComp("MultiShape", 1920, 1080, 1, 5, 30);
        for (var i = 1; i <= 3; i++) {
            var s = c.layers.addShape();
            s.name = "L" + i;
            var vg = s.property("ADBE Root Vectors Group");
            var grp = vg.addProperty("ADBE Vector Group").property("ADBE Vectors Group");
            grp.addProperty("ADBE Vector Shape - Rect");
            grp.addProperty("ADBE Vector Graphic - Fill");
        }
        log.push("  comp layers.length = " + c.numLayers);
        for (var k = 1; k <= c.numLayers; k++) {
            log.push("  layer[" + k + "] name=" + c.layer(k).name + " index=" + c.layer(k).index);
        }
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
