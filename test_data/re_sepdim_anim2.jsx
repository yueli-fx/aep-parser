// RE/generalization fixture #2 for animated Separate Dimensions: a DIFFERENT
// animated 3D Position (4 keyframes, non-uniform timing, different values) so
// the byte-diff confirms the 100×/0.01-influence + arc-length mapping is not
// specific to the first fixture's spacing/values. Builds, saves before, toggles
// dimensionsSeparated=true, saves after.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var beforeFile = new File(dir + "re_sepdim_anim2_before.aep");
    var afterFile = new File(dir + "re_sepdim_anim2_after.aep");
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

    step("build_layer_anim2", function () {
        comp = app.project.items.addComp("sepanim2", 1280, 720, 1, 10, 30);
        solid = comp.layers.addSolid([0, 1, 0], "sepa2", 200, 200, 1);
        solid.threeDLayer = true;
        tg = solid.property("ADBE Transform Group");
        pos = tg.property("ADBE Position");
        pos.setValueAtTime(0.0, [0, 0, 0]);
        pos.setValueAtTime(0.5, [200, -100, 300]);
        pos.setValueAtTime(1.5, [-50, 400, 100]);
        pos.setValueAtTime(3.0, [600, 50, -200]);
    });

    step("save_before", function () { app.project.save(beforeFile); });
    step("separate", function () { pos.dimensionsSeparated = true; });
    step("save_after", function () { app.project.save(afterFile); });

    step("write_done_marker", function () {
        var marker = new File(dir + "re_sepdim_anim2.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    step("quit", function () { app.quit(); });
})();
