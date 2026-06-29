// Probe: vary CameraLayer.cameraOption.filmSize to confirm ldta @0x98 mapping.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_camera_filmsize.aep");
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

    function makeCam(name, configure) {
        var c = app.project.items.addComp(name, 1920, 1080, 1.0, 10, 29.97);
        var cam = c.layers.addCamera("MyCam", [960, 540]);
        try { configure(cam); } catch (e) { log.push("    " + name + " configure: " + e.toString()); }
        return c;
    }

    // Baseline — default film size 36 mm.
    step("a_default", function () { makeCam("A_default", function () {}); });

    // Probe what cameraOption properties exist.
    step("probe_camopt", function () {
        var c = app.project.items.addComp("Z_probe", 1920, 1080, 1.0, 10, 29.97);
        var cam = c.layers.addCamera("ZCam", [960, 540]);
        var co = cam.cameraOption;
        var keys = [];
        for (var k in co) keys.push(k);
        log.push("  cameraOption keys: " + keys.join(","));
    });

    // List all cameraOption sub-properties so we can find Film Size match name.
    step("probe_camopt_subprops", function () {
        var c = app.project.items.addComp("Z_subprops", 1920, 1080, 1.0, 10, 29.97);
        var cam = c.layers.addCamera("ZCam2", [960, 540]);
        var co = cam.cameraOption;
        for (var i = 1; i <= co.numProperties; i++) {
            var p = co.property(i);
            log.push("  cameraOption.prop[" + i + "] name=" + p.name + " matchName=" + p.matchName);
        }
    });

    // Try setting filmSize via match name.
    step("b_filmsize_50mm", function () {
        makeCam("B_filmsize_50mm", function (cam) {
            try {
                var fs = cam.cameraOption.property("ADBE Camera Film Size");
                fs.setValue(50);
                log.push("    OK ADBE Camera Film Size=50");
            } catch (e) { log.push("    ERR " + e); }
        });
    });

    step("c_filmsize_24mm", function () {
        makeCam("C_filmsize_24mm", function (cam) {
            try {
                var fs = cam.cameraOption.property("ADBE Camera Film Size");
                fs.setValue(24);
                log.push("    OK ADBE Camera Film Size=24");
            } catch (e) { log.push("    ERR " + e); }
        });
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_camera_filmsize.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
