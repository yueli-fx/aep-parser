// Effect-library wave 7 RE fixture: probe sweep for the Cycore (CC) effect
// family + Lumetri Color + a few corrected ADBE names. canAddProperty filters
// real match-names; addProperty actually splices; OK/FAIL logged per name.
//
// Curation: parameter-only — excluded CC effects with a layer pickwhip
// (CC Composite / CC Mr. Mercury producer-from-layer / CC Glue Gun) and any
// that fail the post-extract all-tdpi audit (foreign layer id → dropped).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_lib7.aep");
    var log = [];

    var wanted = [
        // ADBE genuinely-missing / corrected
        "ADBE Lumetri",
        "ADBE Lightning",
        // CC Blur & Sharpen
        "CC Radial Fast Blur",
        "CC Radial Blur",
        "CC Vector Blur",
        "CC Cross Blur",
        // CC Distort
        "CC Bend It",
        "CC Bender",
        "CC Blobbylize",
        "CC Flo Motion",
        "CC Griddler",
        "CC Lens",
        "CC Page Turn",
        "CC Power Pin",
        "CC Ripple Pulse",
        "CC Slant",
        "CC Smear",
        "CC Split",
        "CC Split 2",
        "CC Tiler",
        "CC WarpoMatic",
        // CC Generate
        "CC Light Burst 2.5",
        "CC Light Rays",
        "CC Light Sweep",
        "CC Threads",
        // CC Perspective
        "CC Cylinder",
        "CC Sphere",
        "CC Spotlight",
        // CC Stylize
        "CC Glass",
        "CC HexTile",
        "CC Kaleida",
        "CC Mr. Smoothie",
        "CC Plastic",
        "CC RepeTile",
        "CC Threshold",
        "CC Threshold RGB",
        // CC Simulation (param-only ones)
        "CC Pixel Polly",
        "CC Scatterize",
        "CC Star Burst",
        // CC Time
        "CC Force Motion Blur",
        "CC Wide Time",
        // CC Color
        "CC Color Offset",
        "CC Toner",
        // CC Render
        "CC Burn Film",
        "CC Vignette",
        "CC Simple Wire Removal"
    ];

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("efxlib7", 1920, 1080, 1, 5, 30);
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

    var marker = new File(dir + "re_effect_lib7.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
