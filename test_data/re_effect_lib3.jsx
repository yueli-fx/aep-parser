// Effect-library wave 4 RE fixture: add a batch of MG-common built-in effects
// (distort / generate / stylize / transition) to one solid so extract_effect_lib
// can pull each (tdmn, sspc) into a template. Tries each match-name and logs
// success/failure (so a wrong match-name surfaces instead of aborting).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_lib3.aep");
    var log = [];

    var wanted = [
        "ADBE Turbulent Displace",
        "ADBE Roughen Edges",
        "ADBE Echo",
        "ADBE Radial Blur",
        "ADBE 4ColorGradient",
        "ADBE Checkerboard",
        "ADBE Grid",
        "ADBE Stroke",
        "ADBE Corner Pin",
        "ADBE Venetian Blinds"
    ];

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("efxlib3", 1920, 1080, 1, 5, 30);
        var solid = comp.layers.addSolid([0.5, 0.5, 0.5], "S", 1920, 1080, 1);
        var parade = solid.property("ADBE Effect Parade");
        for (var i = 0; i < wanted.length; i++) {
            var mn = wanted[i];
            var canAdd = false;
            try { canAdd = parade.canAddProperty(mn); } catch (e) { canAdd = "?"; }
            try {
                parade.addProperty(mn);
                log.push("OK   " + mn + " (canAdd=" + canAdd + ")");
            } catch (e2) {
                log.push("FAIL " + mn + " (canAdd=" + canAdd + ") -> " + e2.toString());
            }
        }
        log.push("parade numProps=" + parade.numProperties);
        for (var j = 1; j <= parade.numProperties; j++) {
            log.push("  [" + j + "] " + parade.property(j).matchName);
        }
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
    } catch (e3) {
        log.push("EXC " + e3.toString() + " line=" + e3.line);
    }

    var marker = new File(dir + "re_effect_lib3.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
