// test_data/build_renderer_ae2020.jsx
//
// AE-2020-native renderer fixture builder + probe. Creates a comp with a 3D
// layer, logs comp.renderers (the engines AE 2020 exposes), then saves four
// project files each with a distinct renderer set. Lets the Go side dump
// AE-2020-native prin/prda templates and compare against the AE 2025 ones.
//
// Run via ae_run.ps1 against AE 2020. Writes .done with the renderers probe.

(function () {
    var base = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var avail = [];
    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
    });

    step("make_comp_3d", function () {
        var c = app.project.items.addComp("Comp 1", 1920, 1080, 1.0, 5, 30);
        var s = c.layers.addSolid([1, 1, 1], "L", 1920, 1080, 1.0);
        s.threeDLayer = true; // renderer only meaningful with a 3D layer
        avail = c.renderers;
        log.push("  renderers=" + avail.join("|"));
        log.push("  default renderer=" + c.renderer);
    });

    // Save one file per available renderer so Go can dump native prin/prda.
    function saveWith(idx, tag) {
        step("save_" + tag, function () {
            var c = app.project.item(1);
            if (idx < avail.length) {
                c.renderer = avail[idx];
                log.push("  set[" + tag + "]=" + avail[idx] + " readback=" + c.renderer);
            } else {
                log.push("  skip[" + tag + "] (only " + avail.length + " renderers)");
            }
            app.project.save(new File(base + "renderer_ae2020_" + tag + ".aep"));
        });
    }
    saveWith(0, "r0");
    saveWith(1, "r1");
    saveWith(2, "r2");
    saveWith(3, "r3");

    step("write_done_marker", function () {
        var m = new File(base + "build_renderer_ae2020.done");
        m.open("w");
        m.write("PASS\n" + log.join("\n"));
        m.close();
    });

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
