// Task 1D ship gate — open both baseline + modified .aep, read each
// ExtendScript-exposed Project setting, write results to a .done log.
// Workflow: workshop/playbooks/re-fixture.md.
(function () {
    var basePath = "e:/projects/tools/aep-parser/test_data/generated/fixtures/ship_gate_1d_baseline.aep";
    var modPath  = "e:/projects/tools/aep-parser/test_data/generated/fixtures/ship_gate_1d_modified.aep";
    var donePath = "e:/projects/tools/aep-parser/test_data/ship_gate_1d.done";
    var log = [];

    function readSettings(label) {
        log.push("=== " + label + " ===");
        // Each property wrapped in try/catch — some may not be exposed
        // by the ExtendScript Project object on this AE version.
        var keys = [
            "linearBlending",
            "linearizeWorkingSpace",
            "expressionEngine",
            "workingGamma",
            "bitsPerChannel",
            "timeDisplayType",
            "displayStartFrame",
            "framesCountType",
            "framesUseFeetFrames",
            "feetFramesFilmType",
            "footageTimecodeDisplayStartType",
            "transparencyGridThumbnails"
        ];
        for (var i = 0; i < keys.length; i++) {
            var k = keys[i];
            try {
                var v = app.project[k];
                log.push("  " + k + " = " + v + "  (" + typeof v + ")");
            } catch (e) {
                log.push("  " + k + " = <err: " + e.toString() + ">");
            }
        }
    }

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    // Fresh-project guard so -r reruns don't accumulate.
    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    step("open_baseline", function () {
        app.open(new File(basePath));
    });
    step("read_baseline_settings", function () {
        readSettings("BASELINE (re_cameralight.aep)");
    });

    step("close_baseline", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
    });

    step("open_modified", function () {
        app.open(new File(modPath));
    });
    step("read_modified_settings", function () {
        readSettings("MODIFIED (8 settings flipped Go-side)");
    });

    step("write_done", function () {
        var m = new File(donePath);
        m.open("w");
        m.write(log.join("\n"));
        m.close();
    });

    // Tear down — see re-fixture.md.
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
