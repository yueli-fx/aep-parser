// Probe for the wave-6 PARKED classic effects whose match-names were wrong
// (Bezier Warp/Mesh Warp/Channel Mixer/Reshape/Vegas/Warp + a few classics that
// may be missing). Old AE effect match-names are often ALL-CAPS with spaces and
// are case-sensitive, so wave 6's camelCase guesses failed. canAddProperty filters
// the real name; addProperty reads the stored match-name + numProperties (to flag
// any layer-ref). Probe only — not saved.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var log = [];
    var candidates = [
        "ADBE BEZMESH", "ADBE Bezier Warp", "ADBE BezierWarp",
        "ADBE MESH WARP", "ADBE Mesh Warp", "ADBE MeshWarp",
        "ADBE CHANNEL MIXER", "ADBE Channel Mixer", "ADBE ChannelMixer",
        "ADBE Reshape", "ADBE RESHAPE",
        "ADBE Vegas", "ADBE VEGAS",
        "ADBE Warp", "ADBE WarpRegion", "ADBE WARP",
        "ADBE Texturize", "ADBE Color Link", "ADBE Compound Arithmetic",
        "ADBE Set Channels", "ADBE Vector Paint", "ADBE PS Express",
        "ADBE Fill2", "ADBE Smear", "ADBE Wave World2"
    ];
    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("PROBE", 1920, 1080, 1, 5, 30);
        var solid = comp.layers.addSolid([0.5, 0.5, 0.5], "S", 1920, 1080, 1);
        var parade = solid.property("ADBE Effect Parade");
        for (var i = 0; i < candidates.length; i++) {
            var mn = candidates[i];
            var canAdd = false;
            try { canAdd = parade.canAddProperty(mn); } catch (e) { canAdd = "ERR:" + e.toString(); }
            if (canAdd === true) {
                try {
                    var fx = parade.addProperty(mn);
                    var nLayer = 0;
                    for (var j = 1; j <= fx.numProperties; j++) {
                        try { if (fx.property(j).propertyValueType === PropertyValueType.LAYER_INDEX) nLayer++; } catch (eP) {}
                    }
                    log.push("OK   " + mn + " -> stored=" + fx.matchName + " nProps=" + fx.numProperties + " nTopLayerRef=" + nLayer);
                } catch (e2) { log.push("ADDFAIL " + mn + " -> " + e2.toString()); }
            } else {
                log.push("no   " + mn + " (canAdd=" + canAdd + ")");
            }
        }
    } catch (e3) {
        log.push("EXC " + e3.toString() + " line=" + e3.line);
    }
    var marker = new File(dir + "re_effect_probe11.done");
    marker.open("w");
    marker.write("PASS\n" + log.join("\n"));
    marker.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
