// RE linearBlending — toggle it on via ScriptingAPI, see what AE writes.
(function () {
    var srcPath = "e:/projects/tools/aep-parser/test_data/fixtures/re_cameralight.aep";
    var outPath = "e:/projects/tools/aep-parser/test_data/fixtures/re_linear_blending_on.aep";
    var donePath = "e:/projects/tools/aep-parser/test_data/re_linear_blending.done";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }
    step("fresh", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });
    step("open_src", function () {
        app.open(new File(srcPath));
    });
    step("before", function () {
        log.push("  linearBlending before = " + app.project.linearBlending);
        log.push("  linearizeWorkingSpace before = " + app.project.linearizeWorkingSpace);
    });
    step("toggle_on", function () {
        app.project.linearBlending = true;
        log.push("  linearBlending after  = " + app.project.linearBlending);
    });
    step("save_as_modified", function () {
        var f = new File(outPath);
        app.project.save(f);
    });
    step("write_done", function () {
        var m = new File(donePath);
        m.open("w");
        m.write(log.join("\n"));
        m.close();
    });
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
