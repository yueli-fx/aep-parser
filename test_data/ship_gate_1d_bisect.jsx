// Bisect 8 variants — try opening each, log accept/reject.
(function () {
    var donePath = "e:/projects/tools/aep-parser/test_data/ship_gate_1d_bisect.done";
    var variants = ["none", "lnrb", "lnrp", "acer", "adfr", "dwga", "gpug", "exen"];
    var log = [];

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    for (var i = 0; i < variants.length; i++) {
        var v = variants[i];
        var path = "e:/projects/tools/aep-parser/test_data/ship_gate_1d_bisect_" + v + ".aep";
        (function (vName, p) {
            step("open_" + vName, function () {
                app.open(new File(p));
                // If we got here, file opened. Read one trivial property to
                // confirm project loaded.
                var n = app.project.numItems;
                log.push("    " + vName + ": opened, numItems=" + n);
                app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
            });
        })(v, path);
    }

    step("write_done", function () {
        var m = new File(donePath);
        m.open("w");
        m.write(log.join("\n"));
        m.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
