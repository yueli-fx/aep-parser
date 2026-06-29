// Touch-all fixture for the remaining effect-param control-type templates:
// angle / color / 2D point / 3D point / slider. One expression-control effect
// per control type, each single tunable param set to a non-default value so AE
// persists the (tdmn, LIST:tdbs) stream (value-keyed elision —
// incidents/effect-param-elision-synthesis-lite.md). Extraction:
// tmp_debug/extract_effect_params per effect match-name.
//
//   L1 "touched"  — Angle/Color/Point/Point3D/Slider Control, all touched
//   L2 "p3dlib"   — untouched ADBE Point3D Control instance (added LAST so it
//                   is the top layer = first parade in the chunk stream, where
//                   tmp_debug/extract_effect_lib looks for the AddEffect
//                   library template)
//
// Run via scripts/ae_run.ps1 against AE 2020.
// Output: test_data/generated/fixtures/re_effect_param_types.aep + .done
(function () {
    var outFile  = new File("e:/projects/tools/aep-parser/test_data/generated/fixtures/re_effect_param_types.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_effect_param_types.done");
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
        comp = app.project.items.addComp("ParamTypes", 1920, 1080, 1, 5, 24);
    });

    step("L1_touched_controls", function () {
        var l = comp.layers.addSolid([0.5, 0.5, 0.5], "touched", 100, 100, 1);
        var parade = l.property("ADBE Effect Parade");
        function touch(effectMN, value) {
            var fx = parade.addProperty(effectMN);
            var p = fx.property(effectMN + "-0001");
            log.push("  " + effectMN + "-0001 default=" + p.value.toString());
            p.setValue(value);
            log.push("  " + effectMN + "-0001 set=" + p.value.toString());
        }
        touch("ADBE Angle Control", 45);                  // angle (3)
        touch("ADBE Color Control", [0.25, 0.5, 0.75]);   // color (5)
        touch("ADBE Point Control", [123, 456]);          // 2D point (6)
        touch("ADBE Point3D Control", [12, 34, 56]);      // 3D point (18)
        touch("ADBE Slider Control", 42.5);               // slider (10)
    });

    step("L2_p3dlib_untouched", function () {
        var l = comp.layers.addSolid([0.5, 0.5, 0.5], "p3dlib", 100, 100, 1);
        l.property("ADBE Effect Parade").addProperty("ADBE Point3D Control");
    });

    step("save", function () { app.project.save(outFile); });

    doneFile.open("w");
    doneFile.write("PASS\n" + log.join("\n"));
    doneFile.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
