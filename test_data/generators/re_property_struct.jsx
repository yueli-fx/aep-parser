// RE fixture for P3 §3C — PropertyBase.Remove / Duplicate / MoveTo.
//
// Goal: establish AE's behavior contract for structural property ops on the
// canonical indexed group (ADBE Effect Parade). Questions:
//   Q1 Remove : does AE splice just the effect's tdmn+payload pair, or also
//               touch a count/index chunk in the parade?
//   Q2 Duplicate: what name/match-name does the clone get (suffix " 2"?),
//               and does AE assign a fresh internal effect-instance id?
//   Q3 MoveTo : pure child reorder, or is there an explicit order index?
//   Q4 Refuse : does .remove() on a fixed Transform leaf (ADBE Position)
//               throw? capture the error string + each node's propertyType
//               so we can derive the removability predicate.
//
// AE 2020 only (project read floor). Effects used are AE-2020 built-ins:
//   ADBE Gaussian Blur 2, ADBE Tint, ADBE Fill.
//
// Mode selection: $.getenv("RE_PROP_MODE") in {baseline, remove, duplicate, move}.
// Defaults to baseline. baseline also runs the Q4 probe (the failed remove()
// is caught and leaves the project untouched, so the saved baseline is clean).
//
// Outputs (per mode):
//   test_data/re_property_struct_<mode>.aep
//   test_data/re_property_struct_<mode>.done   (step log + probe output)
(function () {
    var mode = $.getenv("RE_PROP_MODE");
    if (mode === null || mode === "") mode = "baseline";
    var validModes = "baseline|remove|duplicate|move";
    if (validModes.indexOf(mode) === -1) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/re_property_struct_unknown.done");
        errFile.open("w");
        errFile.write("ERR unknown mode: " + mode);
        errFile.close();
        return;
    }

    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_property_struct_" + mode + ".aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_property_struct_" + mode + ".done");
    var log = ["mode=" + mode];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    step("fresh_project", function () {
        app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES);
        app.newProject();
        app.project.save(outFile);
    });

    var comp, layer, fx;

    step("add_comp_layer", function () {
        comp = app.project.items.addComp("PropStruct", 1920, 1080, 1, 5, 24);
        layer = comp.layers.addSolid([0.5, 0.5, 0.5], "FxLayer", 100, 100, 1);
    });

    step("add_effects", function () {
        fx = layer.property("ADBE Effect Parade");
        fx.addProperty("ADBE Gaussian Blur 2"); // index 1
        fx.addProperty("ADBE Tint");            // index 2
        fx.addProperty("ADBE Fill");            // index 3
    });

    function logParade(tag) {
        log.push("  parade " + tag + ": numProps=" + fx.numProperties);
        for (var i = 1; i <= fx.numProperties; i++) {
            var e = fx.property(i);
            log.push("    [" + i + "] name=" + e.name + " match=" + e.matchName +
                     " type=" + e.propertyType + " enabled=" + e.enabled);
        }
    }

    step("probe_parade_pre", function () { logParade("pre"); });

    // Q4 — propertyType of containers + refuse behavior on a fixed leaf.
    step("probe_types", function () {
        var tg = layer.property("ADBE Transform Group");
        log.push("  EffectParade.propertyType=" + fx.propertyType +
                 " (INDEXED_GROUP=" + PropertyType.INDEXED_GROUP + ")");
        log.push("  TransformGroup.propertyType=" + tg.propertyType +
                 " (NAMED_GROUP=" + PropertyType.NAMED_GROUP + ")");
        var pos = tg.property("ADBE Position");
        log.push("  Position.propertyType=" + pos.propertyType +
                 " (PROPERTY=" + PropertyType.PROPERTY + ")");
        try {
            pos.remove();
            log.push("  Position.remove() -> NO THROW (unexpected)");
        } catch (e) {
            log.push("  Position.remove() -> THROW: " + e.toString());
        }
        // canAddProperty distinguishes indexed (addable) from named (fixed)
        try { log.push("  EffectParade.canAddProperty('ADBE Fill')=" + fx.canAddProperty("ADBE Fill")); } catch (e) { log.push("  canAddProperty err " + e); }
        try { log.push("  TransformGroup.canAddProperty('ADBE Position')=" + tg.canAddProperty("ADBE Position")); } catch (e) { log.push("  canAddProperty err " + e); }
    });

    step("perform_mutation", function () {
        if (mode === "baseline") {
            // no-op: saved as shared input fixture
        } else if (mode === "remove") {
            fx.property(2).remove();          // remove middle (Tint)
        } else if (mode === "duplicate") {
            fx.property(1).duplicate();        // duplicate first (Gaussian Blur)
        } else if (mode === "move") {
            fx.property(3).moveTo(1);          // move last (Fill) to index 1
        }
    });

    step("probe_parade_post", function () { logParade("post"); });

    step("save", function () { app.project.save(); });

    try {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    } catch (e) { /* best-effort */ }

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
