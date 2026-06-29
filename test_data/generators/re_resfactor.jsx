(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    // Save a comp with resolutionFactor set to a distinctive value. Read back
    // via NUMERIC INDEXING (rf[0], rf[1]) — never "" + rf, which throws
    // "数字结果无效（除以零？）" on array valueOf (effect-param incident finding 4).
    function makeComp(file, fx, fy) {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var c = app.project.items.addComp("R", 1920, 1080, 1, 5, 30);
        var out = new File(base + file);
        app.project.save(out);
        c.resolutionFactor = [fx, fy];
        var rf = c.resolutionFactor;
        log.push("  set [" + fx + "," + fy + "] readback=" + rf[0] + "x" + rf[1]);
        app.project.save();
    }

    step("res_half", function () { makeComp("re_resfactor_half.aep", 2, 2); });
    step("res_third", function () { makeComp("re_resfactor_third.aep", 3, 3); });
    step("res_full", function () { makeComp("re_resfactor_full.aep", 1, 1); });

    step("write_done_marker", function () {
        var marker = new File(base + "re_resfactor.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
