// Effect-library wave 9: BIG probe sweep toward full coverage. Tries a large
// authoritative master list of AE built-in match-names NOT yet in the library
// (keying / simulation / more Cycore CC / corrected old names / color / noise &
// grain / time / transition / utility / perspective / generate). canAddProperty
// filters real names; addProperty splices; OK/FAIL + stored match-name logged.
// Audio effects omitted (need an audio layer, not a solid).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_lib9.aep");
    var log = [];

    var wanted = [
        // --- corrected old names (failed earlier) ---
        "ADBE Add Grain", "ADBE AddGrain2", "ADBE Remove Grain", "ADBE Median",
        "ADBE Turbulent Noise", "ADBE Cartoon", "ADBE Ellipse", "ADBE Radio Waves",
        "ADBE Iris Wipe", "ADBE Pixel Motion Blur", "ADBE Shadow Highlight",
        "ADBE Shadow/Highlight", "ADBE Selective Color", "ADBE Liquify",
        "ADBE Noise HLS Auto2",
        // --- keying ---
        "ADBE Color Key", "ADBE Pro Linear Color Key2", "ADBE Color Range",
        "ADBE Extract", "ADBE Pro Spill2", "ADBE Advanced Spill Suppressor",
        "ADBE Key Cleaner", "ADBE Luma Key", "ADBE KeyCleaner",
        // --- color / utility ---
        "ADBE Broadcast Colors", "ADBE Color Stabilizer",
        "ADBE ProfileToProfile", "ADBE Cineon Converter2",
        "ADBE HDR Compander", "ADBE HDR Highlight Compression", "ADBE GROW BOUNDS",
        "ADBE Channel Combiner",
        // --- distort ---
        "ADBE SubspaceStabilizer", "ADBE Detailpreservingupscale", "ADBE Geometry",
        "ADBE Basic 3D",
        // --- perspective ---
        "ADBE 3D Glasses",
        // --- generate / text ---
        "ADBE Timecode", "ADBE Beam", "ADBE Lightning",
        // --- time ---
        "ADBE Timewarp", "ADBE Time Difference", "ADBE Time Displacement",
        // --- transition ---
        "ADBE Card Wipe", "ADBE Gradient Wipe", "ADBE Image Wipe",
        // --- simulation (ADBE) ---
        "ADBE Foam", "ADBE Wave World", "ADBE Caustics", "ADBE Shatter",
        "ADBE Particle Playground", "ADBE Rain", "ADBE Snow",
        // --- expression / data ---
        "ADBE Layer Control", "ADBE Dropdown Control",
        // --- CC simulation / particles ---
        "CC Ball Action", "CC Bubbles", "CC Drizzle", "CC Hair", "CC Mr. Mercury",
        "CC Particle Systems II", "CC Particle World", "CC Rainfall", "CC Snowfall",
        "CC Pixel Polly", "CC Scatterize", "CC Star Burst",
        // --- CC transition ---
        "CC Glass Wipe", "CC Grid Wipe", "CC Image Wipe", "CC Jaws",
        "CC Light Wipe", "CC Line Sweep", "CC Radial ScaleWipe", "CC Scale Wipe",
        "CC Twister", "CC WarpoMatic", "CC Block Load",
        // --- CC stylize / render / distort (any missed) ---
        "CC Environment", "CC Glue Gun", "CC Composite", "CC Mr. Smoothie",
        "CC Vignette", "CC Snow", "CC Rain", "CC Burn Film", "CC Plastic",
        "CC Smear", "CC Tiler",
        // --- CC color / blur extras ---
        "CC Color Neutralizer", "CC Kernel", "CC Noise", "CC Simple Wire Removal"
    ];

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("efxlib9", 1920, 1080, 1, 5, 30);
        var solid = comp.layers.addSolid([0.5, 0.5, 0.5], "S", 1920, 1080, 1);
        var parade = solid.property("ADBE Effect Parade");
        for (var i = 0; i < wanted.length; i++) {
            var mn = wanted[i];
            var canAdd = false;
            try { canAdd = parade.canAddProperty(mn); } catch (e) { canAdd = "?"; }
            try {
                var fx = parade.addProperty(mn);
                log.push("OK   " + mn + " -> stored=" + fx.matchName);
            } catch (e2) {
                log.push("FAIL " + mn + " (canAdd=" + canAdd + ")");
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

    var marker = new File(dir + "re_effect_lib9.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
