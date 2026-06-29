// RE fixture for V3 Phase 3 — Composition.DuplicateLayer.
//
// Goal: produce 4 AE-saved fixtures that answer RE-Q1..Q7 (see
// workshop/plans/2026-05-28-v3-phase3-duplicatelayer-plan.md):
//   Q1 New layer placement index — top / below source / bottom?
//   Q2 New layer ID — head counter +1 vs other?
//   Q3 14 follower chunks (fvdv/fiop/ftts/foac/fiac/fipc/fifl ×2)
//      verbatim clone or any field mutated?
//   Q4 Other layers' ParentID refs — follow the dup or stay on source?
//   Q5 Source's outgoing ParentID — copy to clone or zero?
//   Q6 Source's matte source — copy to clone or zero?
//   Q7 Auto-suffix " 2" on cloned layer name?
//
// AE 2020 compatible (no AE 23+ APIs). Implicit "layer above" matte.
//
// Mode via $.getenv("RE_DUP_MODE") in {solo, dup_parent, dup_child, dup_matted}.
// Per-mode setup ONLY attaches refs needed for that mode's question.
//
// Layout (constant):
//   index 1: L1_top — implicit matte source in dup_matted; child of L2 in dup_child
//   index 2: L2_mid — parent of L3 in dup_parent; matted target in dup_matted
//   index 3: L3_bot — child of L2 in dup_parent
//
// Outputs:
//   test_data/re_duplicate_layer_<mode>.aep
//   test_data/re_duplicate_layer_<mode>.done
(function () {
    var mode = $.getenv("RE_DUP_MODE");
    if (mode === null || mode === "") mode = "solo";
    var validModes = "solo|dup_parent|dup_child|dup_matted";
    if (validModes.indexOf(mode) === -1) {
        var errFile = new File("e:/projects/tools/aep-parser/test_data/re_duplicate_layer_unknown.done");
        errFile.open("w");
        errFile.write("ERR unknown mode: " + mode);
        errFile.close();
        return;
    }

    var outFile  = new File("e:/projects/tools/aep-parser/test_data/re_duplicate_layer_" + mode + ".aep");
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/re_duplicate_layer_" + mode + ".done");
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
        comp = app.project.items.addComp("DupTest", 1920, 1080, 1, 5, 24);
    });

    step("add_layers", function () {
        // Add in reverse so var names match AE indices: l1=1 (top), l2=2, l3=3.
        l3 = comp.layers.addSolid([0, 0, 1], "L3_bot", 100, 100, 1);
        l2 = comp.layers.addSolid([0, 1, 0], "L2_mid", 100, 100, 1);
        l1 = comp.layers.addSolid([1, 0, 0], "L1_top", 100, 100, 1);
    });

    // Per-mode setup.
    step("setup_refs_for_mode", function () {
        if (mode === "dup_parent") {
            l3.parent = l2; // L3 child of L2
        } else if (mode === "dup_child") {
            l1.parent = l2; // L1 child of L2 (L1 has outgoing parent ref)
        } else if (mode === "dup_matted") {
            // L2 matted by the layer above = L1 (implicit).
            l2.trackMatteType = TrackMatteType.ALPHA;
        }
    });

    step("log_pre_duplicate", function () {
        log.push("  pre-duplicate (" + comp.numLayers + " layers):");
        for (var i = 1; i <= comp.numLayers; i++) {
            var li = comp.layer(i);
            var info = "    [" + i + "] " + li.name;
            try { info += " id=" + li.id; } catch (e) {}
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

    // Resolve duplicate target per mode.
    var target = null;
    var newLayer = null;
    step("resolve_target", function () {
        if (mode === "solo")            target = l2; // middle, no refs
        else if (mode === "dup_parent") target = l2; // L2 is parent of L3
        else if (mode === "dup_child")  target = l1; // L1 has parent=L2
        else if (mode === "dup_matted") target = l2; // L2 matted by L1
        if (target === null) throw new Error("no target resolved for mode " + mode);
        log.push("  duplicating: " + target.name);
    });

    step("perform_duplicate", function () {
        newLayer = target.duplicate();
        if (!newLayer) throw new Error("duplicate() returned null");
        var info = "  new layer: name=" + newLayer.name;
        try { info += " id=" + newLayer.id; } catch (e) {}
        try { info += " index=" + newLayer.index; } catch (e) {}
        log.push(info);
    });

    step("log_post_duplicate", function () {
        log.push("  post-duplicate (" + comp.numLayers + " layers):");
        for (var i = 1; i <= comp.numLayers; i++) {
            var li = comp.layer(i);
            var info = "    [" + i + "] " + li.name;
            try { info += " id=" + li.id; } catch (e) {}
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
    } catch (e) {}

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
