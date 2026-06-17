// RE: how AE MERGES a separated ANIMATED 3D Position back into a single
// spatial keyframe stream. Builds the same 3-keyframe 3D Position as
// re_separate_dims_anim.jsx, separates it, then merges (dimensionsSeparated=
// false) and saves — giving AE's ground-truth animated-merged form for the
// byte-diff (leader spatial stream layout, @0x10 arc-length, follower
// disposition: leader-only vs re-allocated zeroed Position_0/1).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var mergeFile = new File(dir + "re_sepdim_anim_merge_after.aep");
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

    step("build_layer_anim", function () {
        comp = app.project.items.addComp("sepanim", 1920, 1080, 1, 10, 24);
        solid = comp.layers.addSolid([1, 0, 0], "sepa", 200, 200, 1);
        solid.threeDLayer = true;
        tg = solid.property("ADBE Transform Group");
        pos = tg.property("ADBE Position");
        pos.setValueAtTime(0, [100, 200, 50]);
        pos.setValueAtTime(1, [300, 400, 150]);
        pos.setValueAtTime(2, [500, 100, 250]);
    });

    step("separate", function () {
        pos.dimensionsSeparated = true;
    });

    step("merge", function () {
        pos.dimensionsSeparated = false;
    });

    step("probe_merged", function () {
        log.push("  merged dimensionsSeparated=" + pos.dimensionsSeparated);
        log.push("  merged numKeys=" + pos.numKeys);
        log.push("  tg.numProperties=" + tg.numProperties);
        for (var k = 1; k <= pos.numKeys; k++) {
            log.push("    leader k" + k + " t=" + pos.keyTime(k) + " v=" + pos.keyValue(k).toString());
        }
        var names = ["ADBE Position_0", "ADBE Position_1", "ADBE Position_2"];
        for (var i = 0; i < names.length; i++) {
            try {
                var f = tg.property(names[i]);
                log.push("  " + names[i] + " present numKeys=" + (f ? f.numKeys : "nil"));
            } catch (e) {
                log.push("  " + names[i] + " MISSING");
            }
        }
    });

    step("save_merge", function () {
        app.project.save(mergeFile);
    });

    step("write_done_marker", function () {
        var marker = new File(dir + "re_sepdim_anim_merge.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    step("quit", function () {
        app.quit();
    });
})();
