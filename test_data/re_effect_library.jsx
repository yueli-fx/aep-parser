// RE fixture generator for AddEffect's embedded effect-template library.
//
// Applies a curated set of parameter-only built-in effects (no layer/path
// reference params, which would not survive a standalone splice) to a single
// solid layer, then saves. A Go extractor (TestZZExtractEffectLibrary) pulls
// each effect's (tdmn, LIST:sspc) pair into its own templates/effect_*.bin.
//
// Run via scripts/ae_run.ps1 against AE 2020 (read floor → maximally portable
// templates; AE-2020 effect bytes are accepted by AE 2025, mirroring the
// gradient version-portability finding).
//
// Output:
//   test_data/re_effect_library.aep
//   test_data/re_effect_library.done   (applied/skipped match-names)
(function () {
    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_effect_library.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_effect_library.done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}

    // Curated parameter-only built-ins present in AE 2020. If a match-name is
    // not installed, canAddProperty returns false and we skip it (logged).
    var wanted = [
        "ADBE Gaussian Blur 2",      // Gaussian Blur
        "ADBE Fill",                 // Fill
        "ADBE Tint",                 // Tint
        "ADBE Brightness & Contrast 2", // Brightness & Contrast
        "ADBE Tritone",              // Tritone
        "ADBE Easy Levels2",         // Levels
        "ADBE Pro Levels2",          // Levels (Individual Controls)
        "ADBE HUE SATURATION",       // Hue/Saturation
        "ADBE Box Blur",             // Fast Box Blur (ADBE Box Blur in 2020)
        "ADBE Glo2",                 // Glow
        "ADBE Invert",               // Invert
        "ADBE Exposure2"             // Exposure
    ];

    var applied = [];
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("FxLib", 1920, 1080, 1, 5, 24);
        var layer = comp.layers.addSolid([0.5, 0.5, 0.5], "FxLayer", 100, 100, 1);
        var parade = layer.property("ADBE Effect Parade");

        for (var i = 0; i < wanted.length; i++) {
            var mn = wanted[i];
            try {
                if (parade.canAddProperty(mn)) {
                    parade.addProperty(mn);
                    applied.push(mn);
                    log.push("OK  " + mn);
                } else {
                    log.push("SKIP(cannot) " + mn);
                }
            } catch (e) {
                log.push("ERR " + mn + " -> " + e.toString());
            }
        }
        app.project.save(outFile);
        log.push("applied=" + applied.length + " [" + applied.join(", ") + "]");
    } catch (e) {
        log.push("EXC " + e.toString());
    }

    doneFile.open("w");
    doneFile.write((applied.length > 0 ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    doneFile.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
