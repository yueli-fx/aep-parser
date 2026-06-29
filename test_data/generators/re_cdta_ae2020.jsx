// AE 2020 baseline: simple comps with same settings as re_cdta_probe.jsx,
// to verify cdta layout stability across AE versions.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_cdta_ae2020.aep");
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

    function makeComp(name, configure) {
        var c = app.project.items.addComp(name, 1920, 1080, 1.0, 10, 29.97);
        try { configure(c); } catch (e) { log.push("    " + name + " configure: " + e.toString()); }
        return c;
    }

    step("comp_a_baseline", function () { makeComp("A_baseline", function () {}); });
    step("comp_b_resolution_half", function () {
        makeComp("B_resolution_half", function (c) { c.resolutionFactor = [2, 2]; });
    });
    step("comp_c_shutter_phase_neg90", function () {
        makeComp("C_shutter_phase_neg90", function (c) { c.shutterPhase = -90; });
    });

    step("probe_version", function () {
        log.push("  app.version: " + app.version);
        log.push("  app.buildName: " + app.buildName);
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_cdta_ae2020.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
