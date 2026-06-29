// Build a render-queue ADD base: two comps (RQA, RQB) but only RQA queued.
// Go's RenderQueue.AddItem(RQB) then appends the second item; ship-gate checks
// AE accepts the cloned+remapped item and reads RQB as item 2.
(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    step("fresh", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });
    step("build", function () {
        var a = app.project.items.addComp("RQA", 1920, 1080, 1.0, 5, 30);
        app.project.items.addComp("RQB", 1280, 720, 1.0, 3, 24); // not queued
        app.project.renderQueue.items.add(a);
        log.push("  numItems=" + app.project.renderQueue.numItems);
    });
    step("save", function () { app.project.save(new File(base + "re_rq_add_before.aep")); });
    step("done", function () {
        var m = new File(base + "re_rq_add.done");
        m.open("w"); m.write("PASS\n" + log.join("\n")); m.close();
    });
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
