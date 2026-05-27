// RE fixture for V3 Phase 2 — Composition.DeleteLayer.
//
// Goal: produce 4 AE-saved fixtures that answer RE-Q1..Q4 (see
// workshop/plans/2026-05-28-v3-phase2-deletelayer-plan.md):
//   Q1 itemList: splice (back-shift) vs gap (hold position)?
//   Q2 Layer.ParentID orphan: reset to 0 / lift to grandparent / refuse?
//   Q3 Layer.TrackMatteLayerID orphan: reset to 0 / re-pick neighbor / refuse?
//   Q4 Ewst sibling co-deleted with Layr?
//
// AE version compatibility:
//   - Uses ONLY APIs available in AE 2020 (project read floor per CLAUDE.md):
//     * comp.layers.addSolid / addNull — AE 2020
//     * layer.parent = otherLayer    — AE 2020
//     * layer.trackMatteType = TrackMatteType.ALPHA  — AE 2020 (implicit "layer above")
//     * layer.remove()                — AE 2020
//   - Explicit setTrackMatte(layer, type) (AE 23+) is NOT used — implicit
//     matte topology (source = index-1 above target) is portable.
//   - For "matte" mode the resulting AEP behaves differently across AE
//     versions: AE 23+ persists TrackMatteLayerID in ldta @0xA0, AE ≤22
//     stores only the trackMatteType byte. Run matte under AE 2025 to RE
//     the explicit-ID branch; run the rest under AE 2020 (project floor).
//
// Mode selection: $.getenv("RE_DELETE_MODE") in {baseline, middle, parent, matte}.
// Defaults to baseline. Each mode sets up ONLY the refs it tests so the
// fixture diff isolates one variable at a time.
//
// Outputs (per mode):
//   test_data/re_delete_layer_<mode>.aep
//   test_data/re_delete_layer_<mode>.done   (step log)
//
// Layout (constant across all modes):
//   index 1: L1_top (red solid)   — implicit matte source (layer-above) in "matte" mode
//   index 2: L2_mid (green solid) — matted target in "matte" mode; deleted in middle/parent
//   index 3: L3_bot (blue solid)  — child of L2 in "parent" mode
(function () {
    var mode = $.getenv("RE_DELETE_MODE");
    if (mode === null || mode === "") mode = "baseline";
    var validModes = "baseline|middle|parent|matte";
    if (validModes.indexOf(mode) === -1) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/re_delete_layer_unknown.done");
        errFile.open("w");
        errFile.write("ERR unknown mode: " + mode);
        errFile.close();
        return;
    }

    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_delete_layer_" + mode + ".aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_delete_layer_" + mode + ".done");
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

    var comp, l1, l2, l3;

    step("add_comp", function () {
        comp = app.project.items.addComp("DeleteTest", 1920, 1080, 1, 5, 24);
    });

    step("add_layers", function () {
        // addSolid puts the new layer at AE index 1 (top). Add in REVERSE
        // so the variable names match final indices: l1 = index 1 (top),
        // l2 = index 2 (mid), l3 = index 3 (bot).
        l3 = comp.layers.addSolid([0, 0, 1], "L3_bot", 100, 100, 1);
        l2 = comp.layers.addSolid([0, 1, 0], "L2_mid", 100, 100, 1);
        l1 = comp.layers.addSolid([1, 0, 0], "L1_top", 100, 100, 1);
    });

    // Per-mode setup: only attach refs needed for the mode's question.
    // baseline / middle keep the 3 layers ref-free.
    // parent : L3.parent = L2 → delete L2 → answers Q2 (orphan ParentID on L3)
    // matte  : L2.trackMatteType = ALPHA → implicit src = L1 (layer above) →
    //          delete L1 → answers Q3 (matte source orphan + how AE handles
    //          L2's trackMatteType byte / TrackMatteLayerID on AE 23+).
    step("setup_refs_for_mode", function () {
        if (mode === "parent") {
            l3.parent = l2;
        } else if (mode === "matte") {
            l2.trackMatteType = TrackMatteType.ALPHA;
        }
    });

    step("log_pre_delete_layers", function () {
        log.push("  pre-delete:");
        for (var i = 1; i <= comp.numLayers; i++) {
            var li = comp.layer(i);
            var info = "    [" + i + "] " + li.name;
            if (li.parent) info += " parent=" + li.parent.name;
            try {
                if (li.trackMatteLayer) info += " matteSrc=" + li.trackMatteLayer.name;
            } catch (e) { /* AE ≤22 — no trackMatteLayer accessor */ }
            if (li.trackMatteType && li.trackMatteType !== TrackMatteType.NO_TRACK_MATTE) {
                info += " trackMatteType=" + li.trackMatteType;
            }
            log.push(info);
        }
    });

    step("perform_delete", function () {
        if (mode === "baseline") {
            // no-op — baseline saves pre-delete state as ground truth
        } else if (mode === "middle") {
            l2.remove();
        } else if (mode === "parent") {
            l2.remove();
        } else if (mode === "matte") {
            l1.remove();
        }
    });

    step("log_post_delete_layers", function () {
        log.push("  post-delete:");
        for (var i = 1; i <= comp.numLayers; i++) {
            var li = comp.layer(i);
            var info = "    [" + i + "] " + li.name;
            if (li.parent) info += " parent=" + li.parent.name;
            try {
                if (li.trackMatteLayer) info += " matteSrc=" + li.trackMatteLayer.name;
            } catch (e) {}
            if (li.trackMatteType && li.trackMatteType !== TrackMatteType.NO_TRACK_MATTE) {
                info += " trackMatteType=" + li.trackMatteType;
            }
            log.push(info);
        }
    });

    step("save", function () { app.project.save(); });

    try {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    } catch (e) { /* best-effort */ }

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
