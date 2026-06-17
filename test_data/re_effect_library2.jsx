// RE fixture generator for AddEffect's embedded effect-template library, wave 2.
//
// Same shape as re_effect_library.jsx: applies a curated set of parameter-only
// built-in effects (no layer/path reference params — "ADBE Layer Control" and
// friends stay excluded) to a single solid layer, then saves. A Go extractor
// (tmp_debug/extract_effect_lib) pulls each effect's (tdmn, LIST:sspc) pair
// into its own templates/effect_*.bin.
//
// Run via scripts/ae_run.ps1 against AE 2020 (read floor → maximally portable
// templates; wave-1 ship-gates proved AE 2025 accepts AE-2020 effect bytes).
//
// Output:
//   test_data/re_effect_library2.aep
//   test_data/re_effect_library2.done   (applied/skipped match-names)
(function () {
    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_effect_library2.aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_effect_library2.done");
    var log = [];
    try { log.push("ae=" + app.version); } catch (e) {}

    // Curated parameter-only built-ins present in AE 2020. If a match-name is
    // not installed, canAddProperty returns false and we skip it (logged).
    var wanted = [
        "ADBE Drop Shadow",        // Drop Shadow
        "ADBE Sharpen",            // Sharpen
        "ADBE Mosaic",             // Mosaic
        "ADBE Noise",              // Noise
        "ADBE Geometry2",          // Transform
        "ADBE Ramp",               // Gradient Ramp
        "ADBE Fractal Noise",      // Fractal Noise
        "ADBE Tile",               // Motion Tile
        "ADBE Motion Blur",        // Directional Blur
        "ADBE Linear Wipe",        // Linear Wipe
        "ADBE Wave Warp",          // Wave Warp
        "ADBE CurvesCustom",       // Curves
        "ADBE Slider Control",     // Slider Control (expression control)
        "ADBE Point Control",      // Point Control (expression control)
        "ADBE Color Control",      // Color Control (expression control)
        "ADBE Angle Control",      // Angle Control (expression control)
        "ADBE Checkbox Control"    // Checkbox Control (expression control)
    ];

    var applied = [];
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("FxLib2", 1920, 1080, 1, 5, 24);
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
