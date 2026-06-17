// Effect-library wave 6 RE fixture: corrects 3 wave-5 misses (Bulge / Bezier
// Warp / Channel Mixer) and adds a further batch of MG-common, PARAMETER-ONLY
// built-ins (distort / generate / color-correction / perspective / channel /
// time / matte / noise-grain / transition / stylize). Tries each match-name,
// logs OK/FAIL (a wrong name surfaces instead of aborting).
//
// Curation: parameter-only only — excluded layer/path-reference effects
// (Channel Combiner / Color Link / Match Grain / Set Channels / Blend /
// Calculations / Compound Arithmetic / Gradient Wipe / Card Wipe / Time
// Difference / Time Displacement / Camera Lens Blur / Compound Blur /
// Texturize → SetEffectLayerParam path, not AddEffect).
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_lib6.aep");
    var log = [];

    var wanted = [
        // wave-5 corrections
        "ADBE Bulge",
        "ADBE BezMesh",
        "ADBE ChannelMixer",
        // Distort
        "ADBE Warp",
        "ADBE Mesh Warp",
        "ADBE Liquify",
        "ADBE Reshape",
        "ADBE Offset",
        "ADBE Mirror",
        "ADBE Smear",
        // Generate
        "ADBE Vegas",
        "ADBE Radio Waves",
        "ADBE Fractal",
        "ADBE Ellipse",
        "ADBE Write-on",
        "ADBE Scribble Fill",
        "ADBE Eyedropper Fill",
        "ADBE AudSpect",
        "ADBE AudWave",
        // Color Correction
        "ADBE AutoLevels",
        "ADBE AutoColor",
        "ADBE AutoContrast",
        "ADBE Equalize",
        "ADBE Leave Color",
        "ADBE Change To Color",
        "ADBE Change Color",
        "ADBE Shadow/Highlight",
        "ADBE Pro Selective Color",
        // Perspective
        "ADBE Radial Shadow",
        // Channel
        "ADBE Remove Color Matting",
        // Time
        "ADBE Pixel Motion Blur",
        // Matte
        "ADBE Refine Matte",
        // Noise & Grain
        "ADBE AddGrain",
        "ADBE Dust & Scratches",
        "ADBE Median Blur",
        "ADBE RemoveGrain",
        "ADBE Turbulent Noise2",
        "ADBE Noise Alpha2",
        "ADBE Noise HLS2",
        // Transition
        "ADBE Radial Wipe",
        "ADBE Iris Wipe",
        "ADBE Block Dissolve",
        // Stylize
        "ADBE Cartoonizer"
    ];

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("efxlib6", 1920, 1080, 1, 5, 30);
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

    var marker = new File(dir + "re_effect_lib6.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
