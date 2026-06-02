// RE for RenderQueueItem.remove: build 2 comps, enqueue both (2 RQ items),
// save before; remove the 2nd item, save after. Diff LRdr subtree.
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
    step("build_two_items", function () {
        var c1 = app.project.items.addComp("RQA", 1920, 1080, 1.0, 5, 30);
        var c2 = app.project.items.addComp("RQB", 1280, 720, 1.0, 3, 24);
        app.project.renderQueue.items.add(c1);
        app.project.renderQueue.items.add(c2);
        log.push("  numItems=" + app.project.renderQueue.numItems);
    });
    step("save_before", function () { app.project.save(new File(base + "re_rq_delete_before.aep")); });
    step("remove_item2", function () {
        app.project.renderQueue.item(2).remove();
        log.push("  after remove numItems=" + app.project.renderQueue.numItems);
        log.push("  remaining item1 comp=" + app.project.renderQueue.item(1).comp.name);
    });
    step("save_after", function () { app.project.save(new File(base + "re_rq_delete_after.aep")); });

    step("done", function () {
        var m = new File(base + "re_rq_delete.done");
        m.open("w"); m.write("PASS\n" + log.join("\n")); m.close();
    });
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
