// Authors an AE2020-native classic-3D material carrier:
// re_material_classic_2020.aep. A single 3D solid "Mat3D" whose materialOption
// props are each set to a NON-DEFAULT value so they materialize (non-elided) in
// the saved bytes — giving the Go batch-5 gate a carrier it can mutate. Also
// dumps the materialOption tree AE2020 actually exposes (matchName + value) so
// we learn AE2020's classic material model definitively. Writes a .done log.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_material_classic_2020.aep");
    var log = [];
    function P(s) { log.push(s); }
    function step(name, fn) {
        try { fn(); P("OK  " + name); }
        catch (e) { P("ERR " + name + " -> " + e.toString()); }
    }

    step("fresh_project", function () {
        try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
        app.newProject();
        app.project.save(outFile);
    });

    var comp;
    step("make_comp", function () {
        comp = app.project.items.addComp("MAT2020", 1920, 1080, 1.0, 5, 30);
        P("  AE version " + app.version);
    });

    var solid;
    step("solid_3d", function () {
        solid = comp.layers.addSolid([0.3, 0.6, 0.3], "Mat3D", 800, 600, 1.0);
        solid.threeDLayer = true;
    });

    // Set every classic material prop to a distinctive NON-DEFAULT value so it
    // serializes (won't be elided). Try advanced props too; record miss/err.
    step("set_material", function () {
        var mo = solid.materialOption;
        var setters = [
            ["ADBE Casts Shadows", 1],
            ["ADBE Light Transmission", 0.3],
            ["ADBE Accepts Shadows", 0],
            ["ADBE Accepts Lights", 0],
            ["ADBE Shadow Color", [0.5, 0.0, 0.0]],
            ["ADBE Ambient Coefficient", 0.3],
            ["ADBE Diffuse Coefficient", 0.7],
            ["ADBE Specular Coefficient", 0.8],
            ["ADBE Shininess Coefficient", 25],
            ["ADBE Metal Coefficient", 0.4],
            // advanced / ray-traced (likely absent in AE2020 classic):
            ["ADBE Reflection Coefficient", 0.2],
            ["ADBE Glossiness Coefficient", 30],
            ["ADBE Fresnel Coefficient", 0.1],
            ["ADBE Transparency Coefficient", 0.05],
            ["ADBE Transp Rolloff", 0.5],
            ["ADBE Index of Refraction", 1.3],
            ["ADBE Appears in Reflections", 0]
        ];
        for (var j = 0; j < setters.length; j++) {
            try {
                var p = mo.property(setters[j][0]);
                if (p) { p.setValue(setters[j][1]); P("  set  " + setters[j][0] + " = " + setters[j][1]); }
                else { P("  MISS " + setters[j][0]); }
            } catch (e) { P("  ERR  " + setters[j][0] + " -> " + e.toString()); }
        }
    });

    step("dump_exposed", function () {
        var mo = solid.materialOption;
        P("  -- exposed materialOption (" + mo.numProperties + " props) --");
        for (var i = 1; i <= mo.numProperties; i++) {
            var pr = mo.property(i);
            var v;
            try {
                v = pr.value;
                if (v instanceof Array) { var a = []; for (var k = 0; k < v.length; k++) a.push("" + v[k]); v = "[" + a.join(",") + "]"; }
            } catch (e) { v = "ERR:" + e; }
            P("    [" + i + "] " + pr.matchName + " = " + v);
        }
    });

    step("save", function () { app.project.save(); });

    step("done", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/build_material_classic_2020.done");
        marker.open("w");
        marker.write("OK\n" + log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
