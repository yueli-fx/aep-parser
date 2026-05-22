// Probe: identify which chunk/byte persists CompItem.renderer.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_renderer.aep");
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

    function makeRenderComp(name, renderer) {
        var c = app.project.items.addComp(name, 1920, 1080, 1.0, 5, 30);
        try {
            c.renderer = renderer;
            log.push("  " + name + " set renderer=" + renderer + " → actual=" + c.renderer);
        } catch (e) {
            log.push("  " + name + " renderer ERR: " + e);
        }
        // Probe available renderers
        if (name.indexOf("classic") === 0) {
            log.push("  available renderers: " + c.renderers.join(","));
        }
        return c;
    }

    step("comp_classic", function () { makeRenderComp("RDR_classic", "ADBE Standard 3d"); });
    step("comp_advanced", function () { makeRenderComp("RDR_advanced", "ADBE Picasso"); });
    step("comp_cinema4d", function () { makeRenderComp("RDR_cinema4d", "ADBE Ernst"); });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_renderer.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
