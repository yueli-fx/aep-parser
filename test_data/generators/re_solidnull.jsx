(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_solidnull.aep");
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
        comp = app.project.items.addComp("TplComp", 1920, 1080, 1, 10, 30);
    });

    var solid;
    step("add_solid", function () {
        // Distinctive color (0.2, 0.4, 0.8) + non-comp size 640x360 so the
        // opti/sspc byte offsets are unambiguous in the Go-side dump.
        solid = comp.layers.addSolid([0.2, 0.4, 0.8], "TplSolid", 640, 360, 1, 10);
        solid.name = "TplSolid"; // explicit layer name (unset layers store an empty Utf8)
        log.push("  solid src color: " + solid.source.mainSource.color.toString());
        log.push("  solid src dims: " + solid.source.width + "x" + solid.source.height);
    });

    var nul;
    step("add_null", function () {
        nul = comp.layers.addNull(10);
        nul.name = "TplNull";
        log.push("  null src name: " + nul.source.name);
        log.push("  null src dims: " + nul.source.width + "x" + nul.source.height);
        log.push("  null flag: " + nul.nullLayer);
    });

    var adj;
    step("add_adjustment", function () {
        adj = comp.layers.addSolid([1, 1, 1], "TplAdj", 1920, 1080, 1, 10);
        adj.name = "TplAdj";
        adj.adjustmentLayer = true;
        log.push("  adj flag: " + adj.adjustmentLayer);
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_solidnull.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
