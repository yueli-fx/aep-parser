// tmp_debug/gen_rect_subprops.jsx — V2.2.1 子项⑦ RE + template source for Rect
// Position + Roundness persistence. Two shape layers:
//   RectStatic: Size/Position/Roundness set static non-default → richer rect body
//               template source (AE emits each as a flippable cdat).
//   RectAnim:   Position + Roundness animated (2 linear kf) → keyframe layout RE.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_rect_subprops.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/v2_2_rect_subprops.done");
    var log = [], ok = false;
    function linearKf(prop, times, values) {
        for (var i = 0; i < times.length; i++) prop.setValueAtTime(times[i], values[i]);
        for (var k = 1; k <= prop.numKeys; k++) prop.setInterpolationTypeAtKey(k, KeyframeInterpolationType.LINEAR, KeyframeInterpolationType.LINEAR);
    }
    function addRect(comp, name) {
        var s = comp.layers.addShape(); s.name = name;
        var sc = s.property("ADBE Root Vectors Group").addProperty("ADBE Vector Group").property("ADBE Vectors Group");
        return sc.addProperty("ADBE Vector Shape - Rect");
    }
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("RectSub", 1920, 1080, 1, 5, 30);
        var rs = addRect(c, "RectStatic");
        rs.property("ADBE Vector Rect Size").setValue([200, 100]);
        rs.property("ADBE Vector Rect Position").setValue([40, 50]);
        rs.property("ADBE Vector Rect Roundness").setValue(15);
        var ra = addRect(c, "RectAnim");
        ra.property("ADBE Vector Rect Size").setValue([200, 100]);
        linearKf(ra.property("ADBE Vector Rect Position"), [0, 2], [[0, 0], [40, 50]]);
        linearKf(ra.property("ADBE Vector Rect Roundness"), [0, 2], [0, 20]);
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
        ok = true;
    } catch (e) { log.push("ERROR: " + e.toString()); }
    try { doneFile.open("w"); doneFile.write((ok ? "PASS\n" : "FAIL\n") + log.join("\n")); doneFile.close(); } catch (e2) {}
    try { if (app.project) app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e3) {}
    try { app.quit(); } catch (e4) {}
})();
