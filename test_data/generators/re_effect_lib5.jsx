// Effect-library wave 5 RE fixture: add a big batch of MG-common, PARAMETER-ONLY
// built-in effects (distort / stylize / perspective / color-correction / blur /
// channel / generate / time / matte) to one solid so extract_effect_lib can pull
// each (tdmn, sspc) into a template. Tries each match-name and logs
// success/failure (a wrong match-name surfaces as FAIL instead of aborting).
//
// Curation rule (incidents/add-effect-splice-re.md): only effects WITHOUT a
// layer/path reference param (no pickwhip) — those carry a dangling tdpi a
// standalone splice can't satisfy. Excluded here: Camera Lens Blur (Blur Map),
// Texturize (Texture Layer), Displacement Map, Compound Blur, Calculations,
// Set Channels, Blend (all layer-ref → SetEffectLayerParam path, not AddEffect).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_lib5.aep");
    var log = [];

    var wanted = [
        // Distort
        "ADBE BULGE",
        "ADBE Twirl",
        "ADBE Polar Coordinates",
        "ADBE Spherize",
        "ADBE Magnify",
        "ADBE Ripple",
        "ADBE Optics Compensation",
        "ADBE Bezier Warp",
        // Stylize
        "ADBE Posterize",
        "ADBE Threshold2",
        "ADBE Find Edges",
        "ADBE Color Emboss",
        "ADBE Emboss",
        "ADBE Strobe",
        "ADBE Brush Strokes",
        // Perspective
        "ADBE Bevel Alpha",
        "ADBE Bevel Edges",
        // Color Correction
        "ADBE Photo Filter",
        "ADBE Vibrance",
        "ADBE Color Balance 2",
        "ADBE Color Balance (HLS)",
        "ADBE Black&White",
        "ADBE Channel Mixer",
        "ADBE Gamma/Pedestal/Gain2",
        // Blur & Sharpen
        "ADBE Channel Blur",
        "ADBE Bilateral",
        "ADBE Smart Blur",
        "ADBE Unsharp Mask2",
        // Channel
        "ADBE Shift Channels",
        "ADBE Solid Composite",
        "ADBE Minimax",
        "ADBE Arithmetic",
        // Generate
        "ADBE Circle",
        "ADBE Lens Flare",
        "ADBE Cell Pattern",
        "ADBE Lightning 2",
        "ADBE Laser",
        "ADBE Paint Bucket",
        // Time
        "ADBE Posterize Time",
        // Matte
        "ADBE Simple Choker",
        "ADBE Matte Choker"
    ];

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("efxlib5", 1920, 1080, 1, 5, 30);
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

    var marker = new File(dir + "re_effect_lib5.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
