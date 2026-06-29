// RE fixture: AE-native masks across layer kinds + open/closed — ground truth
// for AddMask's from-scratch encoding (shph open flag bytes; per-layer-kind
// coordinate divisor).
//   L1 solid 200x200: OPEN 3-vertex polyline mask, known px coords
//   L2 shape layer:   CLOSED rect mask, known px coords (divisor probe)
//   L3 solid 200x200: CLOSED rect mask (control, matches re_template's)
// Writes .done when finished (ae_run contract).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_mask_open.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_mask_open.done");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("save_as_new_file", function () { app.project.save(outFile); });

    step("comp", function () {
        app.testComp = app.project.items.addComp("RE_MASK", 400, 300, 1, 5, 30);
    });

    step("solid_open_mask", function () {
        var solid = app.testComp.layers.addSolid([1, 0, 0], "SolidOpen", 200, 200, 1.0);
        var mask = solid.property("ADBE Mask Parade").addProperty("ADBE Mask Atom");
        mask.name = "OpenLine";
        var sh = new Shape();
        sh.vertices = [[20, 30], [100, 150], [180, 30]];
        sh.inTangents = [[0, 0], [0, 0], [0, 0]];
        sh.outTangents = [[0, 0], [0, 0], [0, 0]];
        sh.closed = false;
        mask.property("ADBE Mask Shape").setValue(sh);
    });

    step("shape_layer_mask", function () {
        var sl = app.testComp.layers.addShape();
        sl.name = "ShapeMasked";
        var grp = sl.property("ADBE Root Vectors Group").addProperty("ADBE Vector Group");
        var rect = grp.property("ADBE Vectors Group").addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([120, 80]);
        var mask = sl.property("ADBE Mask Parade").addProperty("ADBE Mask Atom");
        mask.name = "ShapeRect";
        var sh = new Shape();
        sh.vertices = [[40, 50], [360, 50], [360, 250], [40, 250]];
        sh.inTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.outTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.closed = true;
        mask.property("ADBE Mask Shape").setValue(sh);
    });

    step("solid_closed_mask", function () {
        var solid = app.testComp.layers.addSolid([0, 1, 0], "SolidClosed", 200, 200, 1.0);
        var mask = solid.property("ADBE Mask Parade").addProperty("ADBE Mask Atom");
        mask.name = "ClosedRect";
        var sh = new Shape();
        sh.vertices = [[10, 10], [190, 10], [190, 190], [10, 190]];
        sh.inTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.outTangents = [[0, 0], [0, 0], [0, 0], [0, 0]];
        sh.closed = true;
        mask.property("ADBE Mask Shape").setValue(sh);
    });

    step("save_again", function () { app.project.save(); });

    doneFile.open("w");
    doneFile.write("PASS\n" + log.join("\n"));
    doneFile.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
