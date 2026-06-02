// RE for SetDimensionsSeparated extension: merge (separate->merged) on a 3D
// layer, and the 2D separate direction (2 followers, no Position_2).
// Produces four fixtures + readback log.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    function pos(layer) {
        return layer.property("ADBE Transform Group").property("ADBE Position");
    }

    // ---- merge pair: 3D separated -> merged ----
    step("fresh1", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });
    var mc, ms, mp;
    step("build_3d_separated", function () {
        mc = app.project.items.addComp("mergecomp", 1920, 1080, 1, 10, 24);
        ms = mc.layers.addSolid([1, 0, 0], "m", 200, 200, 1);
        ms.threeDLayer = true;
        mp = pos(ms);
        mp.setValue([100, 200, 50]);
        mp.dimensionsSeparated = true;
    });
    step("save_merge_before", function () { app.project.save(new File(dir + "re_sepdim_merge_before.aep")); });
    step("merge", function () { pos(ms).dimensionsSeparated = false; });
    step("probe_merge", function () {
        var p = pos(ms);
        log.push("  merged dimensionsSeparated=" + p.dimensionsSeparated);
        log.push("  merged value=" + p.value.toString());
    });
    step("save_merge_after", function () { app.project.save(new File(dir + "re_sepdim_merge_after.aep")); });

    // ---- 2D pair: merged -> separated ----
    step("fresh2", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });
    var dc, ds, dp;
    step("build_2d_merged", function () {
        dc = app.project.items.addComp("twodcomp", 1920, 1080, 1, 10, 24);
        ds = dc.layers.addSolid([0, 1, 0], "d", 200, 200, 1);
        // 2D layer (threeDLayer stays false)
        dp = pos(ds);
        dp.setValue([100, 200]);
        log.push("  2d merged value=" + dp.value.toString());
    });
    step("save_2d_before", function () { app.project.save(new File(dir + "re_sepdim_2d_before.aep")); });
    step("separate_2d", function () { pos(ds).dimensionsSeparated = true; });
    step("probe_2d", function () {
        var tg = ds.property("ADBE Transform Group");
        var p = tg.property("ADBE Position");
        log.push("  2d sep dimensionsSeparated=" + p.dimensionsSeparated);
        var names = ["ADBE Position_0", "ADBE Position_1", "ADBE Position_2"];
        for (var k = 0; k < names.length; k++) {
            try { log.push("  2d " + names[k] + "=" + tg.property(names[k]).value); }
            catch (e) { log.push("  2d " + names[k] + " MISSING"); }
        }
    });
    step("save_2d_after", function () { app.project.save(new File(dir + "re_sepdim_2d_after.aep")); });

    step("write_done", function () {
        var m = new File(dir + "re_separate_dims_ext.done");
        m.open("w"); m.write(log.join("\n")); m.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
