// test_data/re_shape_enums.jsx — RE the remaining shape-node enums in one run.
// Rect + Ellipse + Fill + Stroke on one ShapeLayer. For each, ENUMERATE every
// child (matchName + name + propertyValueType + value), then probe + set the
// enum candidates non-default so AE persists their cdat slots:
//   Fill/Stroke:  ADBE Vector Blend Mode, ADBE Vector Composite Order
//   Rect/Ellipse: shape Direction (matchName TBD — discovered via enumerate)
// AE 2020 — none of these are AE 24+ fields.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_shape_enums.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_shape_enums.done");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    function enumerate(label, grp) {
        if (!grp) { log.push("  " + label + ": <nil>"); return; }
        log.push("  " + label + ".numProperties = " + grp.numProperties);
        for (var i = 1; i <= grp.numProperties; i++) {
            var p = grp.property(i);
            var line = "    [" + i + "] mn=" + p.matchName + " pvt=" +
                       (p.propertyValueType != undefined ? p.propertyValueType : "(group)");
            try { if (p.value != undefined) line += " val=" + p.value.toString(); } catch (ev) {}
            log.push(line);
        }
    }

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    var sc = null;
    var rect = null, ell = null, fill = null, stroke = null;
    // Re-fetch from the Vectors Group by matchName — addProperty reindexes the
    // collection and invalidates handles obtained before later adds.
    function refetch() {
        rect = sc.property("ADBE Vector Shape - Rect");
        ell = sc.property("ADBE Vector Shape - Ellipse");
        fill = sc.property("ADBE Vector Graphic - Fill");
        stroke = sc.property("ADBE Vector Graphic - Stroke");
    }
    step("build_shape", function () {
        var c = app.project.items.addComp("ShapeEnums", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Enums";
        sc = s.property("ADBE Root Vectors Group").addProperty("ADBE Vector Group").property("ADBE Vectors Group");
        sc.addProperty("ADBE Vector Shape - Rect");
        sc.addProperty("ADBE Vector Shape - Ellipse");
        sc.addProperty("ADBE Vector Graphic - Fill");
        sc.addProperty("ADBE Vector Graphic - Stroke");
        refetch();
    });

    step("enumerate_all", function () {
        enumerate("Rect", rect);
        enumerate("Ellipse", ell);
        enumerate("Fill", fill);
        enumerate("Stroke", stroke);
    });

    // Blend Mode + Composite Order on Fill and Stroke. Try a recognizable
    // blend mode (Multiply is index 3 in AE's shape blend enum, typically).
    function trySet(label, grp, mn, v) {
        step("set " + label, function () {
            var p = grp.property(mn);
            if (p == null) { log.push("    " + label + " -> NULL matchName"); return; }
            p.setValue(v);
            log.push("    " + label + " set ok, readback=" + p.value);
        });
    }
    trySet("fill BlendMode=3", fill, "ADBE Vector Blend Mode", 3);
    trySet("fill CompositeOrder=2", fill, "ADBE Vector Composite Order", 2);
    trySet("stroke BlendMode=4", stroke, "ADBE Vector Blend Mode", 4);
    trySet("stroke CompositeOrder=2", stroke, "ADBE Vector Composite Order", 2);

    // Rect/Ellipse Direction — matchName guesses; enumerate output is truth.
    trySet("rect Direction=3", rect, "ADBE Vector Shape Direction", 3);
    trySet("ell Direction=3", ell, "ADBE Vector Shape Direction", 3);

    step("reenumerate_all", function () {
        enumerate("Rect", rect);
        enumerate("Ellipse", ell);
        enumerate("Fill", fill);
        enumerate("Stroke", stroke);
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
