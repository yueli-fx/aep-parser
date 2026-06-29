// RE: per-comp keyframe tick rate.
// Creates 7 comps at different fps, each with one solid layer that has
// Opacity keyframed at t=0 and t=2.0s. Save to a separate file so the
// user's original .aep on disk is untouched.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_tickrate.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); throw e; }
    }

    try {
        // Redirect future saves to outFile. AE's "saveAs": app.project.save(file).
        step("save_as_new", function () { app.project.save(outFile); });
        if (app.project.file == null || app.project.file.fsName.indexOf("re_tickrate") < 0) {
            throw new Error("save_as failed; project file is "
                + (app.project.file && app.project.file.fsName));
        }

        var fpsList = [
            ["RE_fps_24",     24],
            ["RE_fps_25",     25],
            ["RE_fps_29_97",  29.97],
            ["RE_fps_30",     30],
            ["RE_fps_50",     50],
            ["RE_fps_59_94",  59.94],
            ["RE_fps_60",     60]
        ];

        for (var i = 0; i < fpsList.length; i++) {
            (function (name, fps) {
                step("add_" + name, function () {
                    var comp = app.project.items.addComp(name, 100, 100, 1, 5, fps);
                    var solid = comp.layers.addSolid([0, 1, 0], "s", 100, 100, 1.0);
                    solid.opacity.setValueAtTime(0.0, 100);
                    solid.opacity.setValueAtTime(2.0, 50);
                });
            })(fpsList[i][0], fpsList[i][1]);
        }

        step("save", function () { app.project.save(); });
    } catch (e) {
        log.push("FATAL: " + e.toString());
    }

    alert(log.join("\n"));
})();
