// V3 Phase 5D — DuplicateComposition ship-gate verify JSX. Opens the
// Go-emitted ge_duplicate_composition.aep in AE and checks:
//   (a) AE doesn't throw a corruption signal on open,
//   (b) the dup comp "compA_dup" exists with 3 layers,
//   (c) intra-comp parent ref is REMAPPED to the dup's OWN layer — the
//       parented layer's .parent.containingComp === compA_dup (not the
//       original compA_main),
//   (d) sources are SHARED — the dup's A_precomp layer source is
//       compC_precomp; the A_footage layer has a source item.
//
// Inputs:
//   $.getenv("GE_DUP_TAG")   optional .done suffix so the same scenario on
//                            two AE versions writes distinct files
//                            (e.g. "ae2020"); defaults to "default".
//
// Outputs:
//   test_data/generated/ship-gate/verify_ge_duplicate_composition_<tag>.done   (LAST line PASS/FAIL)
(function () {
    var tag = $.getenv("GE_DUP_TAG");
    if (tag === null || tag === "") tag = "default";

    var aepPath  = "e:/projects/tools/aep-parser/test_data/generated/ship-gate/ge_duplicate_composition.aep";
    var doneFile = new File("e:/projects/tools/aep-parser/test_data/generated/ship-gate/verify_ge_duplicate_composition_" + tag + ".done");
    var log = ["tag=" + tag];
    try { log.push("ae=" + app.version); } catch (e) {}

    function step(name, fn) {
        try { fn(); log.push("OK  " + name); }
        catch (e) { log.push("ERR " + name + " -> " + e.toString()); }
    }

    var passed = false;
    var failReason = "";

    step("close_existing_project", function () {
        try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    });

    var openErr = null;
    step("open_ge_file", function () {
        var f = new File(aepPath);
        if (!f.exists) throw new Error("file does not exist: " + aepPath);
        try { app.open(f); } catch (e) { openErr = e.toString(); throw e; }
    });

    if (openErr !== null) {
        failReason = "AE rejected file: " + openErr;
    } else {
        step("inspect_dup", function () {
            if (!app.project) throw new Error("app.project is null after open");
            var dup = null, orig = null;
            for (var i = 1; i <= app.project.numItems; i++) {
                var it = app.project.items[i];
                if (it instanceof CompItem) {
                    if (it.name === "compA_dup") dup = it;
                    else if (it.name === "compA_main") orig = it;
                }
            }
            if (dup === null) throw new Error("comp 'compA_dup' not found");
            if (orig === null) throw new Error("comp 'compA_main' not found");

            log.push("  compA_dup.numLayers=" + dup.numLayers + " (want 3)");
            for (var j = 1; j <= dup.numLayers; j++) {
                var li = dup.layer(j);
                var info = "    [" + j + "] " + li.name;
                if (li.parent) info += " parent=" + li.parent.name + "@" + li.parent.containingComp.name;
                try { if (li.source) info += " source=" + li.source.name; } catch (e) {}
                log.push(info);
            }
            if (dup.numLayers !== 3) throw new Error("dup layer count = " + dup.numLayers + ", want 3");

            // (c) find the layer with a parent; its parent must live in compA_dup.
            var parented = null;
            for (var k = 1; k <= dup.numLayers; k++) {
                if (dup.layer(k).parent) { parented = dup.layer(k); break; }
            }
            if (parented === null) throw new Error("no parented layer found in dup (remap check needs one)");
            var pc = parented.parent.containingComp;
            if (pc !== dup) {
                throw new Error("parent ref NOT remapped: parented layer's parent is in '" + pc.name + "', want 'compA_dup' (must point at dup's OWN layer)");
            }
            if (parented.parent.name !== "A_precomp") {
                throw new Error("remapped parent name = '" + parented.parent.name + "', want 'A_precomp'");
            }

            // (d) sources shared — locate the A_precomp layer in the dup, its
            // source must be the compC_precomp item (same precomp, not cloned).
            var precompLayer = null, footageLayer = null;
            for (var m = 1; m <= dup.numLayers; m++) {
                var lm = dup.layer(m);
                if (lm.name === "A_precomp") precompLayer = lm;
                if (lm.name === "A_footage") footageLayer = lm;
            }
            if (precompLayer === null) throw new Error("dup missing 'A_precomp' layer");
            if (!precompLayer.source || precompLayer.source.name !== "compC_precomp") {
                throw new Error("dup A_precomp source = '" + (precompLayer.source ? precompLayer.source.name : "null") + "', want 'compC_precomp' (shared)");
            }
            if (footageLayer === null || !footageLayer.source) {
                throw new Error("dup A_footage layer missing or has no source");
            }

            passed = true;
        });
    }

    log.push(passed ? "PASS" : ("FAIL " + (failReason || "see ERR lines above")));

    try {
        doneFile.open("w");
        doneFile.write(log.join("\n"));
        doneFile.close();
    } catch (e) {}

    try { app.project.close(CloseOptions.DO_NOT_SAVE_CHANGES); } catch (e) {}
    try { app.quit(); } catch (e) {}
})();
