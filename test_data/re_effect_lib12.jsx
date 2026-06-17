// Wave 12: the wave-6 PARKED classic effects, now with correct ALL-CAPS
// match-names (probe11 confirmed). 5 parameter-only (BEZMESH/MESH WARP/
// CHANNEL MIXER/RESHAPE/Vector Paint) + 4 layer-ref (Texturize/Color Link/
// Compound Arithmetic/Set Channels). Materialize each LAYER_INDEX param at a MAP
// layer (recursive, like re_effect_layerref2.jsx) so the extracted template
// carries a tdpi; log the layer-ref param match-names for the Go consts.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_lib12.aep");
    var log = [];
    var ok = true;

    var wanted = [
        "ADBE BEZMESH", "ADBE MESH WARP", "ADBE CHANNEL MIXER", "ADBE RESHAPE", "ADBE Vector Paint",
        "ADBE Texturize", "ADBE Color Link", "ADBE Compound Arithmetic", "ADBE Set Channels"
    ];

    function walk(grp, mapIdx) {
        var found = 0;
        for (var j = 1; j <= grp.numProperties; j++) {
            var pr = grp.property(j);
            if (pr.propertyType === PropertyType.PROPERTY) {
                if (pr.propertyValueType === PropertyValueType.LAYER_INDEX) {
                    found++;
                    var note = "    LAYER_INDEX " + pr.matchName + " | " + pr.name;
                    try { pr.setValue(mapIdx); note += " set->" + mapIdx; }
                    catch (eSet) { note += " SETFAIL:" + eSet.toString(); ok = false; }
                    log.push(note);
                }
            } else {
                try { found += walk(pr, mapIdx); } catch (eW) {}
            }
        }
        return found;
    }

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("lib12", 1920, 1080, 1, 5, 30);
        var mapL = comp.layers.addSolid([1, 0, 0], "MAP", 1920, 1080, 1);
        var host = comp.layers.addSolid([0.5, 0.5, 0.5], "HOST", 1920, 1080, 1);
        var mapIdx = mapL.index;
        var parade = host.property("ADBE Effect Parade");
        for (var i = 0; i < wanted.length; i++) {
            var mn = wanted[i];
            try {
                var fx = parade.addProperty(mn);
                var nLayer = walk(fx, mapIdx);
                log.push("OK   " + mn + " stored=" + fx.matchName + " nProps=" + fx.numProperties + " nLayerRef=" + nLayer);
            } catch (e) {
                log.push("FAIL " + mn + " -> " + e.toString());
                ok = false;
            }
        }
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
    } catch (e3) {
        log.push("EXC " + e3.toString() + " line=" + e3.line);
        ok = false;
    }
    var marker = new File(dir + "re_effect_lib12.done");
    marker.open("w");
    marker.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    marker.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
