// test_data/build_re_layer_comment.jsx
//
// RE fixture builder: AE-native solid layer WITH a layer comment set, so we can
// observe where AE places the layer-level cmta chunk inside the Layr LIST
// (Layer.SetComment currently appends to the tail and AE ignores it —
// incidents/layer-setcomment-cmta-append-position.md). Saves re_layer_comment.aep.
//
// Run via ae_run.ps1 against AE 2020. Writes re_layer_comment.aep + a .done log.

(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    step("make_comp_solid_with_comment", function () {
        var c = app.project.items.addComp("Main", 1280, 720, 1.0, 5, 30);
        var s = c.layers.addSolid([1, 0, 0], "BG", 1280, 720, 1);
        s.comment = "REPROBE_COMMENT";
        log.push("  comment=[" + s.comment + "]");
    });

    step("save", function () {
        app.project.save(new File(base + "re_layer_comment.aep"));
    });

    step("write_done_marker", function () {
        var m = new File(base + "build_re_layer_comment.done");
        m.open("w");
        m.write("PASS\n" + log.join("\n"));
        m.close();
    });

    try { app.quit(); } catch (eQuit) {}
})();
