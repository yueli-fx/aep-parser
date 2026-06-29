// RE linearizeWorkingSpace — enable it via ScriptingAPI, save, see chunks.
(function () {
    var srcPath = "e:/projects/tools/aep-parser/test_data/fixtures/re_cameralight.aep";
    var outPath = "e:/projects/tools/aep-parser/test_data/fixtures/re_linearize_workspace_on.aep";
    var donePath = "e:/projects/tools/aep-parser/test_data/re_linearize_workspace.done";
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
        log.push("  linearizeWorkingSpace before = " + app.project.linearizeWorkingSpace);
    });
    step("toggle_on", function () {
        app.project.linearizeWorkingSpace = true;
        log.push("  linearizeWorkingSpace after  = " + app.project.linearizeWorkingSpace);
    });
    step("save_as", function () {
        var f = new File(outPath);
        app.project.save(f);
    });
    step("done", function () {
        var m = new File(donePath);
        m.open("w");
        m.write(log.join("\n"));
        m.close();
    });
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
