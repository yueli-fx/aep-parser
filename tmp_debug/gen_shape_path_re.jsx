// tmp_debug/gen_shape_path_re.jsx — multi-path RE disambiguation fixture.
// One project, several ShapeLayers each with one Path of DISTINCT coords so
// every f32 in ldat is identifiable (the 100x100 square's repeated 0/1 values
// were ambiguous). Saved by AE → test_data/v2_2_shape_path_re.aep.
//
// Layers:
//   tri_open   — 3 distinct verts, zero tangents, open
//   tri_closed — same verts, closed
//   tan_path   — 2 verts with distinct non-zero in/out tangents (locate tangent
//                encoding + relative-vs-absolute)
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_path_re.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_shape_path_re.done");
    var log = [];
    var ok = false;
    function addPath(comp, name, verts, inT, outT, closed) {
        var s = comp.layers.addShape();
        s.name = name;
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var sc = sub.property("ADBE Vectors Group");
        var pg = sc.addProperty("ADBE Vector Shape - Group");
        var sh = new Shape();
        sh.vertices = verts;
        sh.inTangents = inT;
        sh.outTangents = outT;
        sh.closed = closed;
        pg.property("ADBE Vector Shape").setValue(sh);
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("PathRE", 1920, 1080, 1, 5, 30);
        addPath(c, "tri_open",
            [[10, 20], [70, 30], [40, 90]],
            [[0, 0], [0, 0], [0, 0]],
            [[0, 0], [0, 0], [0, 0]], false);
        addPath(c, "tri_closed",
            [[10, 20], [70, 30], [40, 90]],
            [[0, 0], [0, 0], [0, 0]],
            [[0, 0], [0, 0], [0, 0]], true);
        addPath(c, "tan_path",
            [[20, 20], [80, 60]],
            [[0, 0], [-7, -3]],
            [[5, 9], [0, 0]], false);
        app.project.save(outFile);
        log.push("saved " + outFile.fsName + " layers=" + c.numLayers);
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
