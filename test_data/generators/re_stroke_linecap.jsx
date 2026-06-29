// test_data/generators/re_stroke_linecap.jsx — RE Line Cap / Line Join / Miter Limit.
// 1 ShapeLayer + Rect (geometry) + Stroke. First ENUMERATE the stroke group's
// children (matchName + name + propertyValueType + value) to discover the real
// matchNames, then SET Cap/Join/Miter to non-defaults so AE persists their
// slots (default values are elided). AE 2020 — these are not AE 24+ fields.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/generated/fixtures/re_stroke_linecap.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_stroke_linecap.done");
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

    var stroke = null;
    step("build_shape", function () {
        var c = app.project.items.addComp("StrokeCap", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Stroke_Cap";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var sc = sub.property("ADBE Vectors Group");
        var rect = sc.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([200, 100]);
        stroke = sc.addProperty("ADBE Vector Graphic - Stroke");
    });

    // Enumerate every child of the stroke group — definitive matchName source.
    step("enumerate_stroke_children", function () {
        log.push("  stroke.numProperties = " + stroke.numProperties);
        for (var i = 1; i <= stroke.numProperties; i++) {
            var p = stroke.property(i);
            var line = "  [" + i + "] mn=" + p.matchName + " name=" + p.name +
                       " pvt=" + (p.propertyValueType != undefined ? p.propertyValueType : "(group)");
            try { if (p.value != undefined) line += " val=" + p.value.toString(); } catch (ev) {}
            log.push(line);
        }
    });

    // Probe the documented matchNames directly — does .property() resolve them?
    var capMN = "ADBE Vector Stroke Line Cap";
    var joinMN = "ADBE Vector Stroke Line Join";
    var miterMN = "ADBE Vector Stroke Miter Limit";
    step("probe_documented_matchnames", function () {
        var probes = [capMN, joinMN, miterMN];
        for (var i = 0; i < probes.length; i++) {
            var pr = stroke.property(probes[i]);
            log.push("  probe " + probes[i] + " -> " +
                     (pr == null ? "NULL (not found)" : "FOUND name=" + pr.name + " val=" + pr.value));
        }
    });

    // Set non-defaults so AE persists the slots. Cap: 1=Butt 2=Round 3=Project.
    // Join: 1=Miter 2=Round 3=Bevel. Miter Limit: number (default 4).
    step("set_line_cap", function () { stroke.property(capMN).setValue(2); });   // Round
    step("set_line_join", function () { stroke.property(joinMN).setValue(3); }); // Bevel
    step("set_miter_limit", function () { stroke.property(miterMN).setValue(8); });

    // Re-enumerate post-set to confirm the slots now report the new values.
    step("reenumerate_after_set", function () {
        for (var i = 1; i <= stroke.numProperties; i++) {
            var p = stroke.property(i);
            var line = "  [" + i + "] mn=" + p.matchName + " name=" + p.name;
            try { if (p.value != undefined) line += " val=" + p.value.toString(); } catch (ev) {}
            log.push(line);
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
