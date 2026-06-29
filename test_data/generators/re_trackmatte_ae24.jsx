// RE fixture for AE 23+ AVLayer.trackMatteLayer (explicit matte-source
// pointer). AE <= 22 used implicit "layer immediately above" convention;
// AE 23 introduced an explicit ID slot so you can pick any layer as the
// matte source. setTrackMatte(layer, type) was also added in 23.0.
//
// Comp layout (RE_TRACKMATTE):
//   index 1: "mt_alpha_to_solidA"  — matted by solidA, ALPHA
//   index 2: "mt_luma_to_solidB"   — matted by solidB, LUMA
//   index 3: "mt_alphainv_to_solidC" — matted by solidC, ALPHA_INVERTED
//   index 4: "mt_baseline"         — no track matte
//   index 5: "solidA"              — used as matte source by layer 1
//   index 6: "solidB"              — used as matte source by layer 2
//   index 7: "solidC"              — used as matte source by layer 3
//
// All seven layers are 100×100 solids in their own colors. The variant
// layer names double as tags so the Go-side dumper can map each layer's
// ldta to the variant.
(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/fixtures/re_trackmatte_ae24.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("redirect_save", function () { app.project.save(outFile); });

    var comp;
    step("add_comp", function () {
        comp = app.project.items.addComp("RE_TRACKMATTE", 1920, 1080, 1, 5, 30);
    });

    // Add the 4 dependent layers (will get track mattes) + 3 source
    // layers. AE renders top-to-bottom so the source needs to be ABOVE
    // (lower index) the matted layer for the implicit "layer above"
    // convention to work — but with explicit trackMatteLayer the order
    // doesn't matter. We deliberately put sources BELOW their matted
    // layers to confirm the explicit ID is used (not the layer-above
    // fallback).
    var solidA, solidB, solidC;
    var mtAlpha, mtLuma, mtAlphaInv, mtBaseline;

    step("add_layers", function () {
        // top to bottom: matted layers first (indexes 1..4), then sources (5..7)
        mtAlpha    = comp.layers.addSolid([1, 0, 0], "mt_alpha_to_solidA",    100, 100, 1);
        mtLuma     = comp.layers.addSolid([0, 1, 0], "mt_luma_to_solidB",     100, 100, 1);
        mtAlphaInv = comp.layers.addSolid([0, 0, 1], "mt_alphainv_to_solidC", 100, 100, 1);
        mtBaseline = comp.layers.addSolid([0.5, 0.5, 0.5], "mt_baseline",     100, 100, 1);
        solidA     = comp.layers.addSolid([1, 1, 0], "solidA", 100, 100, 1);
        solidB     = comp.layers.addSolid([1, 0, 1], "solidB", 100, 100, 1);
        solidC     = comp.layers.addSolid([0, 1, 1], "solidC", 100, 100, 1);
    });

    step("set_trackmatte_alpha", function () {
        mtAlpha.setTrackMatte(solidA, TrackMatteType.ALPHA);
    });
    step("set_trackmatte_luma", function () {
        mtLuma.setTrackMatte(solidB, TrackMatteType.LUMA);
    });
    step("set_trackmatte_alpha_inv", function () {
        mtAlphaInv.setTrackMatte(solidC, TrackMatteType.ALPHA_INVERTED);
    });

    // log the layer IDs in script-visible form via layer.index — we use
    // names for fixture lookup, not script-visible IDs (AE script doesn't
    // expose internal ldta layer ID directly anyway).
    step("log_indices", function () {
        for (var i = 1; i <= comp.numLayers; i++) {
            log.push("  layer[" + i + "] = " + comp.layer(i).name);
        }
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_trackmatte_ae24.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
