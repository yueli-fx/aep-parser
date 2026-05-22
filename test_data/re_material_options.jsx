// Probe AVLayer.materialOption sub-properties for 3D-enabled solid.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_material_options.aep");
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

    var comp;
    step("make_comp", function () {
        comp = app.project.items.addComp("MAT_OPT", 1920, 1080, 1.0, 5, 30);
    });

    step("baseline_solid_2d", function () {
        var s = comp.layers.addSolid([0.5, 0.5, 0.5], "Solid2D", 800, 600, 1.0);
    });

    step("solid_3d_default", function () {
        var s = comp.layers.addSolid([0.7, 0.3, 0.3], "Solid3D_default", 800, 600, 1.0);
        s.threeDLayer = true;
    });

    step("solid_3d_custom_material", function () {
        var s = comp.layers.addSolid([0.3, 0.7, 0.3], "Solid3D_custom", 800, 600, 1.0);
        s.threeDLayer = true;
        var mo = s.materialOption;
        // Enumerate so we can record match names.
        var keys = [];
        for (var i = 1; i <= mo.numProperties; i++) {
            var p = mo.property(i);
            keys.push("[" + i + "]" + p.matchName);
        }
        log.push("  materialOption sub-props: " + keys.join(", "));

        // Try setting a few via match name (these are AE standard match names).
        var setters = [
            ["ADBE Casts Shadows", 2],          // 0=Off, 1=On, 2=Only
            ["ADBE Light Transmission", 0.7],
            ["ADBE Accepts Shadows", 1],
            ["ADBE Accepts Lights", 0],
            ["ADBE Ambient Coefficient", 0.5],
            ["ADBE Diffuse Coefficient", 0.8],
            ["ADBE Specular Coefficient", 0.6],   // older name
            ["ADBE Glossiness Coefficient", 0.4], // newer name (specular intensity?)
            ["ADBE Specular Shininess", 25],
            ["ADBE Metal Coefficient", 0.9],
        ];
        for (var j = 0; j < setters.length; j++) {
            try {
                var p2 = mo.property(setters[j][0]);
                if (p2) {
                    p2.setValue(setters[j][1]);
                    log.push("  set " + setters[j][0] + " = " + setters[j][1]);
                } else {
                    log.push("  miss " + setters[j][0]);
                }
            } catch (e) {
                log.push("  err  " + setters[j][0] + " -> " + e.toString());
            }
        }
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_material_options.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
