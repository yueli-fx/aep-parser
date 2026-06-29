// Generalization fixture #3: UNIFORM keyframe spacing but NON-1.0s (0.5s) to
// isolate whether the separate mapping (speed = spatTan×100, influence =
// 0.01/segDur) holds for uniform spacing at a different segment duration.
// Same values as fixture #1, keyframes at t=0/0.5/1.0.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var beforeFile = new File(dir + "re_sepdim_anim3_before.aep");
    var afterFile = new File(dir + "re_sepdim_anim3_after.aep");
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
    step("build", function () {
        comp = app.project.items.addComp("sepanim3", 1920, 1080, 1, 10, 24);
        solid = comp.layers.addSolid([0, 0, 1], "sepa3", 200, 200, 1);
        solid.threeDLayer = true;
        tg = solid.property("ADBE Transform Group");
        pos = tg.property("ADBE Position");
        pos.setValueAtTime(0.0, [100, 200, 50]);
        pos.setValueAtTime(0.5, [300, 400, 150]);
        pos.setValueAtTime(1.0, [500, 100, 250]);
    });
    step("save_before", function () { app.project.save(beforeFile); });
    step("separate", function () { pos.dimensionsSeparated = true; });
    step("save_after", function () { app.project.save(afterFile); });
    step("write_done_marker", function () {
        var marker = new File(dir + "re_sepdim_anim3.done");
        marker.open("w"); marker.write(log.join("\n")); marker.close();
    });
    step("quit", function () { app.quit(); });
})();
