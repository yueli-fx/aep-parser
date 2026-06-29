// test_data/generators/re_path_anim.jsx — Phase 0 RE for shape-path keyframes.
//
// Build ONE shape layer with a free-form path ("ADBE Vector Shape - Group"
// → "ADBE Vector Shape") and put 3 LINEAR keyframes on it via
// setValueAtTime(t, Shape). Each keyframe uses a DISTINCT vertex count and
// a DISTINCT bounding box so the Go-side dump can answer:
//   (a) does om-s→omks hold one shap geometry per frame (== animated mask)?
//   (b) what is the tdbs.kfl time/interp table layout (64B blocks)?
//   (c) is each frame's shph bbox independently normalized?
//   (d) what fields of lhd3 (52B) move with frame-count / vertex-count?
//
// Keyframes (all closed, axis-aligned where possible for easy byte reading):
//   t=0.0  square  bbox 0..40    (4 verts)
//   t=1.0  square  bbox 10..150  (4 verts)
//   t=2.0  triangle bbox 0..200  (3 verts)  ← different vert count + bbox
//
// AE 2020 (path keyframes are not an AE 23+ feature → run on the read floor;
// AE-2020-accepted bytes are the harder target our builder must match).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/generated/fixtures/re_path_anim.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_path_anim.done");
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

    var pathProp = null;
    step("build_path_layer", function () {
        var c = app.project.items.addComp("PathAnim", 200, 200, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "PathAnim";
        var vectorsGroup = s.property("ADBE Root Vectors Group")
                            .addProperty("ADBE Vector Group")
                            .property("ADBE Vectors Group");
        var pathGroup = vectorsGroup.addProperty("ADBE Vector Shape - Group");
        pathProp = pathGroup.property("ADBE Vector Shape");
        log.push("  pathProp.matchName=" + pathProp.matchName +
                 " pvt=" + pathProp.propertyValueType);
    });

    function mkShape(verts) {
        var sh = new Shape();
        sh.vertices = verts;
        var z = [];
        for (var i = 0; i < verts.length; i++) z.push([0, 0]);
        sh.inTangents = z;
        sh.outTangents = z;
        sh.closed = true;
        return sh;
    }

    step("set_keyframes", function () {
        pathProp.setValueAtTime(0.0, mkShape([[0, 0], [40, 0], [40, 40], [0, 40]]));
        pathProp.setValueAtTime(1.0, mkShape([[10, 10], [150, 10], [150, 150], [10, 150]]));
        pathProp.setValueAtTime(2.0, mkShape([[0, 0], [200, 0], [100, 200]]));
        log.push("  numKeys=" + pathProp.numKeys);
    });

    step("force_linear_interp", function () {
        for (var k = 1; k <= pathProp.numKeys; k++) {
            pathProp.setInterpolationTypeAtKey(
                k,
                KeyframeInterpolationType.LINEAR,
                KeyframeInterpolationType.LINEAR);
            log.push("  key " + k + " t=" + pathProp.keyTime(k) +
                     " inInterp=" + pathProp.keyInInterpolationType(k) +
                     " outInterp=" + pathProp.keyOutInterpolationType(k));
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
