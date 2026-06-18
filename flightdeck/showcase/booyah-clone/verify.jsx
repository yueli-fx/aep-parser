// flightdeck/showcase/booyah-clone/verify.jsx — open booyah-clone.aep and dump the
// named comp's DOM to booyah-clone.verify.json, so the rebuilt structure can be
// reconciled against the original's parsed values (spec §3 "结构保真"). This is the
// AE-side confirmation that AE actually ingested the layers/effects (catches the
// false-green where Go reads bytes back fine but AE silently drops them).
//
// CAVEAT: same UTF-8 / Japanese-name matching caveat as render.jsx.
(function () {
    var dir = "e:/projects/tools/aep-parser/flightdeck/showcase/booyah-clone/";
    var COMP = "メインコンプ！"; // comp to dump
    var out = { comp: COMP, found: false, layers: [] };
    try {
        app.open(new File(dir + "booyah-clone.aep"));
        var comp = null;
        for (var i = 1; i <= app.project.numItems; i++) {
            var it = app.project.item(i);
            if (it instanceof CompItem && it.name === COMP) { comp = it; break; }
        }
        if (comp) {
            out.found = true;
            out.width = comp.width; out.height = comp.height;
            out.duration = comp.duration; out.frameRate = comp.frameRate;
            for (var li = 1; li <= comp.numLayers; li++) {
                var l = comp.layer(li);
                var effects = [];
                try {
                    var eg = l.property("ADBE Effect Parade");
                    if (eg) for (var e = 1; e <= eg.numProperties; e++) {
                        effects.push(eg.property(e).matchName);
                    }
                } catch (e) {}
                var numMasks = 0;
                try { numMasks = l.property("ADBE Mask Parade").numProperties; } catch (e) {}
                out.layers.push({
                    index: l.index, name: l.name,
                    blend: l.blendingMode, enabled: l.enabled,
                    effects: effects, numEffects: effects.length, numMasks: numMasks
                });
            }
        }
    } catch (e) { out.error = e.toString() + " line=" + e.line; }
    var f = new File(dir + "booyah-clone.verify.json");
    f.open("w"); f.write(JSON.stringify(out, null, 2)); f.close();
    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
