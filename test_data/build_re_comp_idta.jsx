// test_data/build_re_comp_idta.jsx
//
// RE fixture builder: AE-native comps each carrying ONE item-level metadata field
// (comment / label / draft3D) plus a BARE baseline, so the Go side can diff their
// idta / cdta chunks and locate the item has-comment flag (analog of layer
// ldta @0x3C) + the real draft3D bit. Saves re_comp_idta.aep.
//
// Run via ae_run.ps1 against AE 2020 (lower version = canonical baseline bytes).

(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString() + " line=" + e.line); }
    }

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    // Each comp identical template, ONE field changed, so idta/cdta diffs isolate it.
    step("comp_BARE", function () {
        app.project.items.addComp("BARE", 640, 360, 1.0, 2, 30);
    });
    step("comp_CMT", function () {
        var c = app.project.items.addComp("CMT", 640, 360, 1.0, 2, 30);
        c.comment = "IDTAPROBE_COMMENT";
        log.push("  comment=[" + c.comment + "]");
    });
    step("comp_LBL", function () {
        var c = app.project.items.addComp("LBL", 640, 360, 1.0, 2, 30);
        c.label = 9;
        log.push("  label=" + c.label);
    });
    step("comp_DRAFT", function () {
        var c = app.project.items.addComp("DRAFT", 640, 360, 1.0, 2, 30);
        c.draft3d = true;
        log.push("  draft3d=" + c.draft3d);
    });

    step("save", function () {
        app.project.save(new File(base + "re_comp_idta.aep"));
    });

    step("write_done_marker", function () {
        var m = new File(base + "build_re_comp_idta.done");
        m.open("w");
        m.write("PASS\n" + log.join("\n"));
        m.close();
    });

    try { app.quit(); } catch (eQuit) {}
})();
