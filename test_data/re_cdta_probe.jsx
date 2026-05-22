// Probe: vary CompItem settings to map unknown cdta bytes.
//
// We make multiple comps, each setting one property to a non-default
// value, plus a baseline comp. After AE saves the project, dump_cdta
// will reveal which bytes changed.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_cdta_probe.aep");
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

    // 0. Baseline — all AE defaults (post-construction).
    step("comp_a_baseline", function () { makeComp("A_baseline", function () {}); });

    // 1. Resolution factor — half (downsample to 960x540 preview).
    step("comp_b_resolution_half", function () {
        makeComp("B_resolution_half", function (c) { c.resolutionFactor = [2, 2]; });
    });

    // 2. Resolution factor — quarter.
    step("comp_c_resolution_quarter", function () {
        makeComp("C_resolution_quarter", function (c) { c.resolutionFactor = [4, 4]; });
    });

    // 3. Resolution factor — custom non-square [3, 4].
    step("comp_d_resolution_3x4", function () {
        makeComp("D_resolution_3x4", function (c) { c.resolutionFactor = [3, 4]; });
    });

    // 4. Shutter phase at -90 (default is -180/0 depending on AE version).
    step("comp_e_shutter_phase_neg90", function () {
        makeComp("E_shutter_phase_neg90", function (c) { c.shutterPhase = -90; });
    });

    // 5. Shutter angle 360°.
    step("comp_f_shutter_angle_360", function () {
        makeComp("F_shutter_angle_360", function (c) { c.shutterAngle = 360; });
    });

    // 6. Probe to print every settable CompItem key for documentation.
    step("probe_keys", function () {
        var c = app.project.items.addComp("Z_probe", 1920, 1080, 1.0, 10, 29.97);
        var keys = [];
        for (var k in c) keys.push(k);
        log.push("  CompItem keys (" + keys.length + "): " + keys.join(","));
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_cdta_probe.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
