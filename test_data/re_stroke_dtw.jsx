// test_data/re_stroke_dtw.jsx — RE Stroke Dashes / Taper / Wave nested groups.
// 1 ShapeLayer + Rect (geometry) + Stroke. Probes each group's children
// (matchName / name / propertyValueType / numProperties) AND sets them
// non-default so AE persists every cdat slot (defaults are elided), then saves
// test_data/re_stroke_dtw.aep for byte-level dump.
//
// AE 2020 (Stroke Taper/Wave shipped AE 2018.1 = 15.1, present in 2020).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_stroke_dtw.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_stroke_dtw.done");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    function dumpGroup(label, grp) {
        if (!grp) { log.push("  " + label + ": <nil>"); return; }
        log.push("  " + label + " numProperties=" + grp.numProperties +
                 " matchName=" + grp.matchName);
        for (var i = 1; i <= grp.numProperties; i++) {
            var p = grp.property(i);
            var vt = "";
            try { vt = p.propertyValueType; } catch (e) { vt = "?"; }
            var val = "";
            try { val = (p.numKeys === undefined ? "" : "" + p.value); } catch (e2) { val = "?"; }
            log.push("    [" + i + "] mn=" + p.matchName + " name=" + p.name +
                     " type=" + (p.propertyType) + " vt=" + vt + " val=" + val);
        }
    }

    var stroke = null, dashes = null, taper = null, wave = null;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    step("build_layer", function () {
        var c = app.project.items.addComp("StrokeDTW", 1920, 1080, 1, 5, 30);
        var s = c.layers.addShape();
        s.name = "Stroke_DTW";
        var root = s.property("ADBE Root Vectors Group");
        var sub = root.addProperty("ADBE Vector Group");
        var sc = sub.property("ADBE Vectors Group");
        var rect = sc.addProperty("ADBE Vector Shape - Rect");
        rect.property("ADBE Vector Rect Size").setValue([300, 150]);
        stroke = sc.addProperty("ADBE Vector Graphic - Stroke");
        stroke.property("ADBE Vector Stroke Color").setValue([1, 0, 0, 1]);
        stroke.property("ADBE Vector Stroke Width").setValue(12);
    });

    // --- DASHES ------------------------------------------------------------
    step("probe_dashes_empty", function () {
        dashes = stroke.property("ADBE Vector Stroke Dashes");
        dumpGroup("dashes(before)", dashes);
    });
    step("add_dashes", function () {
        // Add a Dash + Gap pair, then a second pair, set Offset.
        var d1 = dashes.addProperty("ADBE Vector Stroke Dash 1");
        d1.setValue(20);
        var g1 = dashes.addProperty("ADBE Vector Stroke Gap 1");
        g1.setValue(10);
        var d2 = dashes.addProperty("ADBE Vector Stroke Dash 2");
        d2.setValue(40);
        var g2 = dashes.addProperty("ADBE Vector Stroke Gap 2");
        g2.setValue(30);
        var off = dashes.property("ADBE Vector Stroke Offset");
        off.setValue(5);
    });
    step("probe_dashes_filled", function () { dumpGroup("dashes(after)", dashes); });

    // --- TAPER -------------------------------------------------------------
    step("probe_taper_empty", function () {
        taper = stroke.property("ADBE Vector Stroke Taper");
        dumpGroup("taper(before)", taper);
    });
    step("set_taper", function () {
        function trySet(mn, v) {
            try { taper.property(mn).setValue(v); log.push("    taper set " + mn + "=" + v); }
            catch (e) { log.push("    taper ERR " + mn + " -> " + e.toString()); }
        }
        trySet("ADBE Vector Taper Start Length", 25);
        trySet("ADBE Vector Taper End Length", 35);
        trySet("ADBE Vector Taper Start Width", 50);
        trySet("ADBE Vector Taper End Width", 60);
        trySet("ADBE Vector Taper Start Ease", 40);
        trySet("ADBE Vector Taper End Ease", 45);
        trySet("ADBE Vector Taper Length Units", 2);
    });
    step("probe_taper_filled", function () { dumpGroup("taper(after)", taper); });

    // --- WAVE --------------------------------------------------------------
    step("probe_wave_empty", function () {
        wave = stroke.property("ADBE Vector Stroke Wave");
        dumpGroup("wave(before)", wave);
    });
    step("set_wave", function () {
        function trySet(mn, v) {
            try { wave.property(mn).setValue(v); log.push("    wave set " + mn + "=" + v); }
            catch (e) { log.push("    wave ERR " + mn + " -> " + e.toString()); }
        }
        trySet("ADBE Vector Taper Wave Amount", 15);
        trySet("ADBE Vector Taper Wavelength", 80);
        trySet("ADBE Vector Taper Wave Units", 2);
        trySet("ADBE Vector Taper Wave Phase", 90);
    });
    step("probe_wave_filled", function () { dumpGroup("wave(after)", wave); });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
