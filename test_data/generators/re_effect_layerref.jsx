// RE fixture for the layer-reference effect family (Displacement Map / Compound
// Blur / CC Vector Blur). On a HOST layer, add each effect and point every
// LAYER_INDEX param at a MAP source layer so the layer-ref param MATERIALIZES
// with a tdpi (extract_effect_lib then pulls a template carrying it; AddEffect's
// retarget + SetEffectLayerParam repoint it like Set Matte). Logs each effect's
// param matchNames + which are layer-index params + their tdpi-bearing form.
(function () {
    var dir = "e:/projects/tools/aep-parser/test_data/";
    var outFile = new File(dir + "re_effect_layerref.aep");
    var log = [];

    var wanted = [
        "ADBE Displacement Map",
        "ADBE Compound Blur",
        "CC Vector Blur"
    ];

    try {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        var comp = app.project.items.addComp("efxlayerref", 1920, 1080, 1, 5, 30);
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
                log.push("OK   " + mn + " (canAdd=" + canAdd + ", stored=" + fx.matchName + ", nProps=" + fx.numProperties + ")");
                for (var j = 1; j <= fx.numProperties; j++) {
                    var pr = fx.property(j);
                    var isLayer = (pr.propertyValueType === PropertyValueType.LAYER_INDEX);
                    var note = "    [" + j + "] " + pr.matchName + " | " + pr.name + " | type=" + pr.propertyValueType + (isLayer ? " <<LAYER_INDEX>>" : "");
                    if (isLayer) {
                        try { pr.setValue(mapIdx); note += " set->" + mapIdx; }
                        catch (eSet) { note += " SETFAIL:" + eSet.toString(); }
                    }
                    log.push(note);
                }
            } catch (e2) {
                log.push("FAIL " + mn + " (canAdd=" + canAdd + ") -> " + e2.toString());
            }
        }
        app.project.save(outFile);
        log.push("saved " + outFile.fsName);
    } catch (e3) {
        log.push("EXC " + e3.toString() + " line=" + e3.line);
    }

    var marker = new File(dir + "re_effect_layerref.done");
    marker.open("w");
    marker.write(log.join("\n"));
    marker.close();

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
