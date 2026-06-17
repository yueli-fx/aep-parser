// RE: on-disk keyframe layout of an ANIMATED effect COLOR / 2D-POINT / 3D-POINT
// param. AnimateScalarKeyframes ships 1D-scalar only; color/point need their own
// layout confirmed against an AE-native fixture (effect points may or may not be
// spatial like layer Position — that's the open question).
//
// NOTE: addProperty returns a live ref that goes STALE after a later addProperty
// on the same group (incidents/trim-paths-vector-filter-re.md) — so keyframe each
// effect IMMEDIATELY after adding it, and re-fetch by index in the probe.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_anim_effect_colorpoint.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString() + " line=" + e.line); }
    }

    var comp, solid, parade;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    step("build_and_keyframe", function () {
        comp = app.project.items.addComp("animfx", 1920, 1080, 1, 10, 30);
        solid = comp.layers.addSolid([1, 1, 1], "S", 400, 400, 1);
        parade = solid.property("ADBE Effect Parade");

        var cc = parade.addProperty("ADBE Color Control");
        cc.property("ADBE Color Control-0001").setValueAtTime(0, [1, 0, 0, 1]); // red
        cc.property("ADBE Color Control-0001").setValueAtTime(2, [0, 0, 1, 1]); // blue

        var pc = parade.addProperty("ADBE Point Control");
        pc.property("ADBE Point Control-0001").setValueAtTime(0, [100, 200]);
        pc.property("ADBE Point Control-0001").setValueAtTime(2, [800, 600]);

        var p3 = parade.addProperty("ADBE Point3D Control");
        p3.property("ADBE Point3D Control-0001").setValueAtTime(0, [100, 200, 50]);
        p3.property("ADBE Point3D Control-0001").setValueAtTime(2, [800, 600, 250]);
    });

    step("probe", function () {
        var defs = [
            ["color", 1, "ADBE Color Control-0001"],
            ["point2d", 2, "ADBE Point Control-0001"],
            ["point3d", 3, "ADBE Point3D Control-0001"],
        ];
        for (var i = 0; i < defs.length; i++) {
            var prop = parade.property(defs[i][1]).property(defs[i][2]);
            var spatial = "n/a";
            try { spatial = "" + prop.isSpatial; } catch (e) {}
            log.push("  " + defs[i][0] + " numKeys=" + prop.numKeys + " isSpatial=" + spatial);
            for (var k = 1; k <= prop.numKeys; k++) {
                log.push("    k" + k + " t=" + prop.keyTime(k) + " v=" + prop.keyValue(k).toString());
            }
        }
    });

    step("save", function () {
        app.project.save(outFile);
    });

    step("write_done", function () {
        var marker = new File(dir + "re_anim_effect_colorpoint.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
