(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var beforeFile = new File(dir + "re_separate_dims_before.aep");
    var afterFile = new File(dir + "re_separate_dims_after.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var comp, solid, tg, pos;

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    step("build_layer", function () {
        comp = app.project.items.addComp("sepcomp", 1920, 1080, 1, 10, 24);
        solid = comp.layers.addSolid([1, 0, 0], "sep", 200, 200, 1);
        solid.threeDLayer = true;
        tg = solid.property("ADBE Transform Group");
        pos = tg.property("ADBE Position");
        // move off-default so AE persists the property (avoid default omission)
        pos.setValue([100, 200, 50]);
    });

    step("probe_before", function () {
        log.push("  before pos.value=" + pos.value.toString());
        log.push("  before dimensionsSeparated=" + pos.dimensionsSeparated);
        log.push("  before tg.numProperties=" + tg.numProperties);
    });

    step("save_before", function () {
        app.project.save(beforeFile);
    });

    step("separate", function () {
        pos.dimensionsSeparated = true;
    });

    step("probe_after", function () {
        log.push("  after dimensionsSeparated=" + pos.dimensionsSeparated);
        log.push("  after tg.numProperties=" + tg.numProperties);
        // re-fetch leader + followers
        var leader = tg.property("ADBE Position");
        log.push("  leader.matchName=" + leader.matchName + " sep=" + leader.dimensionsSeparated);
        var names = ["ADBE Position_0", "ADBE Position_1", "ADBE Position_2"];
        for (var i = 0; i < names.length; i++) {
            try {
                var f = tg.property(names[i]);
                log.push("  follower " + names[i] + " value=" + f.value);
            } catch (e) {
                log.push("  follower " + names[i] + " MISSING -> " + e.toString());
            }
        }
    });

    step("save_after", function () {
        app.project.save(afterFile);
    });

    step("write_done_marker", function () {
        var marker = new File(dir + "re_separate_dims.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
