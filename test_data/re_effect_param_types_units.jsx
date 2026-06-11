// Units probe for effect point-param cdat encoding: which denominator scales
// each axis (layer width / height / fixed 100), and whether source-less layers
// (shape) store raw pixels like mask paths do (cf.
// incidents/add-mask-create-re.md fraction-vs-pixel finding).
//
//   L1 solid 200x100 — Point Control [123,45], Point3D Control [123,45,67]
//   L2 shape (source-less) — same values
//
// Readback: go run ./tmp_debug/probe_effects test_data/re_effect_param_types_units.aep
// Run via scripts/ae_run.ps1 against AE 2020.
(function () {
    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_effect_param_types_units.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_effect_param_types_units.done");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    try { log.push("ae=" + app.version); } catch (e) {}

    var comp;
    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        comp = app.project.items.addComp("UnitsProbe", 1920, 1080, 1, 5, 24);
    });

    // NOTE: adding a second effect invalidates property references obtained
    // before the add — fetch, set, and log each one before the next addProperty.
    function touchPoints(l, tag) {
        var parade = l.property("ADBE Effect Parade");
        var p2 = parade.addProperty("ADBE Point Control").property("ADBE Point Control-0001");
        p2.setValue([123, 45]);
        log.push("  " + tag + " point=" + p2.value.toString());
        var p3 = parade.addProperty("ADBE Point3D Control").property("ADBE Point3D Control-0001");
        p3.setValue([123, 45, 67]);
        log.push("  " + tag + " point3d=" + p3.value.toString());
    }

    step("L1_solid_200x100", function () {
        touchPoints(comp.layers.addSolid([0.5, 0.5, 0.5], "solid200x100", 200, 100, 1), "solid");
    });

    step("L2_shape_sourceless", function () {
        touchPoints(comp.layers.addShape(), "shape");
    });

    step("save", function () { app.project.save(outFile); });

    doneFile.open("w");
    doneFile.write("PASS\n" + log.join("\n"));
    doneFile.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
