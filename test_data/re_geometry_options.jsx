// Probe AVLayer.geometryOption for 3D solid (Advanced 3D renderer required).
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_geometry_options.aep");
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
    step("make_comp_advanced3d", function () {
        comp = app.project.items.addComp("GEO_OPT", 1920, 1080, 1.0, 5, 30);
        // Try Advanced 3D renderer (matchName "ADBE Picasso").
        try {
            comp.renderer = "ADBE Picasso";
            log.push("  renderer set to ADBE Picasso (Advanced 3D)");
        } catch (e) {
            log.push("  renderer set FAILED: " + e);
        }
        log.push("  current renderer: " + comp.renderer);
    });

    step("solid_3d_default", function () {
        var s = comp.layers.addSolid([0.5, 0.5, 0.5], "Solid3D_geo_default", 800, 600, 1.0);
        s.threeDLayer = true;
    });

    step("solid_3d_geo_custom", function () {
        var s = comp.layers.addSolid([0.7, 0.3, 0.3], "Solid3D_geo_custom", 800, 600, 1.0);
        s.threeDLayer = true;
        // geometryOption is a separate PropertyGroup (AE 24+).
        var go = null;
        try {
            go = s.geometryOption;
            log.push("  geometryOption present");
        } catch (e) {
            log.push("  geometryOption ERR: " + e);
        }
        if (go) {
            var keys = [];
            for (var i = 1; i <= go.numProperties; i++) {
                var p = go.property(i);
                keys.push("[" + i + "]" + p.matchName);
            }
            log.push("  geo subprops: " + keys.join(", "));

            // Try setting common ones.
            var setters = [
                ["ADBE Plane Curvature", 0.5],
                ["ADBE Plane Subdivision", 8],
                ["ADBE Bevel Style", 2],
                ["ADBE Bevel Direction", 0],
                ["ADBE Bevel Depth", 5],
                ["ADBE Hole Bevel Depth", 3],
                ["ADBE Extrusion Depth", 10],
            ];
            for (var j = 0; j < setters.length; j++) {
                try {
                    var p2 = go.property(setters[j][0]);
                    if (p2) {
                        p2.setValue(setters[j][1]);
                        log.push("  set " + setters[j][0] + " = " + setters[j][1]);
                    } else {
                        log.push("  miss " + setters[j][0]);
                    }
                } catch (e) {
                    log.push("  err " + setters[j][0] + " -> " + e);
                }
            }
        }
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_geometry_options.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
