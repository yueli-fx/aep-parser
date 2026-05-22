// Multi-RE fixture for AE 24+ Wave 2 leftovers + a few cdta extras:
//
//   RE_TIMEREMAP comp (30 fps, 5 s):
//     • baseline               — no time remap
//     • timeremap_on           — time remap enabled (default identity curve)
//   RE_LIGHTS comp (30 fps, 5 s):
//     • light_parallel         — LightType.PARALLEL
//     • light_spot             — LightType.SPOT
//     • light_point            — LightType.POINT
//     • light_ambient          — LightType.AMBIENT
//     • light_spot_shadows_on  — SPOT with castsShadows = true
//     • light_spot_shadows_off — SPOT with castsShadows = false
//   RE_CDTA_DSF comp (29.97 fps, 1 m, displayStart varies):
//     • baseline               — displayStartFrame default 0
//     • displayStart_120       — displayStartFrame = 120 (= 4 s @ 30)
//     • displayStart_dropframe — dropFrame = true (timecode display only)
//
// AE 25 doesn't write Utf8 layer name chunks for solids/lights/null —
// identify each layer by composition + Layer.Name fallback (AE script
// `layer.name = "foo"` does set it; the parser may need extra work to
// find that name in AE 25's chunk layout, but the script still sets
// the name attribute so the Source item name reflects it).

(function () {
    var outFile = new File("e:/projects/tools/aep-parser/test_data/re_wave2_ae24.aep");
    var log = [];
    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("redirect_save", function () { app.project.save(outFile); });

    // ───── time remap comp ─────
    // timeRemap only works on layers with video footage/precomp source.
    // Create a tiny nested comp, then use it as the source layer twice.
    var compTRInner, compTR;
    step("add_comp_timeremap_inner", function () {
        compTRInner = app.project.items.addComp("RE_TIMEREMAP_INNER", 320, 240, 1, 5, 30);
        compTRInner.layers.addSolid([1, 0.5, 0], "inner_solid", 320, 240, 1);
    });
    step("add_comp_timeremap", function () {
        compTR = app.project.items.addComp("RE_TIMEREMAP", 1920, 1080, 1, 5, 30);
    });
    step("add_tr_baseline", function () {
        var L = compTR.layers.add(compTRInner);
        L.name = "tr_baseline";
    });
    step("add_tr_remap_on", function () {
        var L = compTR.layers.add(compTRInner);
        L.name = "tr_remap_on";
        L.timeRemapEnabled = true;
    });

    // ───── lights comp ─────
    var compLi;
    step("add_comp_lights", function () {
        compLi = app.project.items.addComp("RE_LIGHTS", 1920, 1080, 1, 5, 30);
    });

    function addLight(name, type, castsShadows) {
        var L = compLi.layers.addLight(name, [960, 540]);
        L.name = name;
        L.lightType = type;
        // castsShadows is at property: layer.lightOption.castsShadows = bool
        // (AE scripting exposes as a Property, not a top-level layer attr)
        if (castsShadows !== undefined) {
            try {
                var p = L.property("Light Options").property("Casts Shadows");
                if (p) p.setValue(castsShadows ? 1 : 0);
            } catch (e) { log.push("  skip castsShadows on " + name + ": " + e.toString()); }
        }
    }

    step("add_light_parallel",     function () { addLight("light_parallel",     LightType.PARALLEL); });
    step("add_light_spot",         function () { addLight("light_spot",         LightType.SPOT); });
    step("add_light_point",        function () { addLight("light_point",        LightType.POINT); });
    step("add_light_ambient",      function () { addLight("light_ambient",      LightType.AMBIENT); });
    step("add_light_spot_shadow_on",  function () { addLight("light_spot_shadows_on",  LightType.SPOT, true); });
    step("add_light_spot_shadow_off", function () { addLight("light_spot_shadows_off", LightType.SPOT, false); });

    // ───── cdta displayStart / dropFrame comp ─────
    var compDS;
    step("add_comp_cdta", function () {
        compDS = app.project.items.addComp("RE_CDTA_DSF", 1920, 1080, 1, 60, 29.97);
    });
    step("set_cdta_baseline_solid", function () {
        var s = compDS.layers.addSolid([0.7, 0.7, 0.7], "cdta_solid", 100, 100, 1);
        s.name = "cdta_solid";
    });

    // For displayStartFrame variants we duplicate the comp and modify.
    var compDS_120, compDS_drop;
    step("dup_cdta_120", function () {
        compDS_120 = compDS.duplicate();
        compDS_120.name = "RE_CDTA_DSF_120";
        compDS_120.displayStartFrame = 120;
    });
    step("dup_cdta_drop", function () {
        compDS_drop = compDS.duplicate();
        compDS_drop.name = "RE_CDTA_DSF_DROP";
        compDS_drop.dropFrame = true;
    });

    step("save", function () { app.project.save(); });

    step("write_done_marker", function () {
        var marker = new File("e:/projects/tools/aep-parser/test_data/re_wave2_ae24.done");
        marker.open("w");
        marker.write(log.join("\n"));
        marker.close();
    });
})();
