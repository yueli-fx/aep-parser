// test_data/build_rq_ae2020.jsx
//
// AE-2020-native render-queue base fixture builder. Creates a comp, adds it to
// the render queue, and saves WITHOUT a comment so the Go-side ship gate can
// exercise RenderQueueItem.SetComment's *insert* path (a fresh RCom chunk),
// then have AE read the comment back. AE 2020 opens it directly; AE 2025 opens
// it via the (auto-dismissed) convert dialog.
//
// Run via ae_run.ps1 against AE 2020. Writes rq_ae2020_base.aep + a .done log.

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

    step("make_comp_and_enqueue", function () {
        var c = app.project.items.addComp("TestComp", 1920, 1080, 1.0, 5, 30);
        var rqi = app.project.renderQueue.items.add(c);
        log.push("  numRQItems=" + app.project.renderQueue.numItems);
        log.push("  initial comment=[" + rqi.comment + "]");
    });

    step("save", function () {
        app.project.save(new File(base + "rq_ae2020_base.aep"));
    });

    step("write_done_marker", function () {
        var m = new File(base + "build_rq_ae2020.done");
        m.open("w");
        m.write("PASS\n" + log.join("\n"));
        m.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
