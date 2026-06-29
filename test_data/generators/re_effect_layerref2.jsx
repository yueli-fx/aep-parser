// RE fixture for layer-ref effect WAVE 2: 3D Glasses / Warp Stabilizer
// (ADBE SubspaceStabilizer) / Timewarp / CC Particle World. Same materialize
// flow as wave 1 (re_effect_layerref.jsx): add each effect on a HOST solid,
// recursively find every LAYER_INDEX param (these complex effects nest their
// layer pickwhip inside sub-groups, unlike Displacement Map's top-level param),
// point it at a MAP source layer so the param MATERIALIZES with a tdpi, then
// save so extract_effect_lib pulls a tdpi-bearing template.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_layerref2.aep");
    var log = [];
    var ok = true;

    var wanted = [
        "ADBE SubspaceStabilizer",
        "ADBE 3D Glasses",
        "ADBE Timewarp",
        "CC Particle World"
    ];

    // recursively walk an effect (or group), recording + materializing every
    // LAYER_INDEX leaf param. Returns count of layer-index params found.
    function walk(grp, path, mapIdx) {
        var found = 0;
        for (var j = 1; j <= grp.numProperties; j++) {
            var pr = grp.property(j);
            if (pr.propertyType === PropertyType.PROPERTY) {
                if (pr.propertyValueType === PropertyValueType.LAYER_INDEX) {
                    found++;
                    var note = "    LAYER_INDEX " + pr.matchName + " | " + pr.name + " | " + path;
                    try { pr.setValue(mapIdx); note += " set->" + mapIdx; }
                    catch (eSet) { note += " SETFAIL:" + eSet.toString(); ok = false; }
                    log.push(note);
                }
            } else {
                // PropertyGroup / INDEXED_GROUP — recurse
                try { found += walk(pr, path + " > " + pr.matchName, mapIdx); }
                catch (eW) { log.push("    walk-err " + pr.matchName + " -> " + eW.toString()); }
            }
        }
        return found;
    }

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("efxlayerref2", 1920, 1080, 1, 5, 30);
        var mapL = comp.layers.addSolid([1, 0, 0], "MAP", 1920, 1080, 1);
        var host = comp.layers.addSolid([0.5, 0.5, 0.5], "HOST", 1920, 1080, 1);
        var mapIdx = mapL.index;
        log.push("MAP layer index=" + mapIdx);
        var parade = host.property("ADBE Effect Parade");
        for (var i = 0; i < wanted.length; i++) {
            var mn = wanted[i];
            var canAdd = false;
            try { canAdd = parade.canAddProperty(mn); } catch (e) { canAdd = "?"; }
            try {
                var fx = parade.addProperty(mn);
                var nLayer = walk(fx, "(root)", mapIdx);
                log.push("OK   " + mn + " (canAdd=" + canAdd + ", stored=" + fx.matchName +
                    ", nProps=" + fx.numProperties + ", nLayerRef=" + nLayer + ")");
                if (nLayer === 0) { log.push("    !! NO LAYER_INDEX param found"); }
            } catch (e2) {
                log.push("FAIL " + mn + " (canAdd=" + canAdd + ") -> " + e2.toString());
                ok = false;
            }
        }
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
    } catch (e3) {
        log.push("EXC " + e3.toString() + " line=" + e3.line);
        ok = false;
    }

    var marker = new File(dir + "re_effect_layerref2.done");
    marker.open("w");
    marker.write((ok ? "PASS" : "FAIL") + "\n" + log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
