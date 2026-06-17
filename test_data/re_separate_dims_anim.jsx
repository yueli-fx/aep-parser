// RE: how AE separates an ANIMATED 3D Position into per-axis keyframe streams.
// Builds a 3D layer with a 3-keyframe Position (default interp → auto-bezier
// spatial tangents), saves before, toggles dimensionsSeparated=true, saves
// after. Logs per-axis follower keyframe times/values for the byte-diff.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var beforeFile = new File(dir + "re_sepdim_anim_before.aep");
    var afterFile = new File(dir + "re_sepdim_anim_after.aep");
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

    step("probe_before", function () {
        log.push("  before dimensionsSeparated=" + pos.dimensionsSeparated);
        log.push("  before numKeys=" + pos.numKeys);
        for (var k = 1; k <= pos.numKeys; k++) {
            log.push("    leader k" + k + " t=" + pos.keyTime(k) + " v=" + pos.keyValue(k).toString()
                + " inInterp=" + pos.keyInTemporalEase(k)[0].influence
                + " outInterp=" + pos.keyOutTemporalEase(k)[0].influence);
        }
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
        var names = ["ADBE Position_0", "ADBE Position_1", "ADBE Position_2"];
        for (var i = 0; i < names.length; i++) {
            try {
                var f = tg.property(names[i]);
                log.push("  " + names[i] + " numKeys=" + f.numKeys);
                for (var k = 1; k <= f.numKeys; k++) {
                    log.push("    k" + k + " t=" + f.keyTime(k) + " v=" + f.keyValue(k)
                        + " inInf=" + f.keyInTemporalEase(k)[0].influence
                        + " inSpd=" + f.keyInTemporalEase(k)[0].speed
                        + " outInf=" + f.keyOutTemporalEase(k)[0].influence
                        + " outSpd=" + f.keyOutTemporalEase(k)[0].speed);
                }
            } catch (e) {
                log.push("  " + names[i] + " MISSING -> " + e.toString());
            }
        }
    });

    step("save_after", function () {
        app.project.save(afterFile);
    });

    step("write_done_marker", function () {
        var marker = new File(dir + "re_sepdim_anim.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
